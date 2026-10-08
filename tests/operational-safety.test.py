"""Safety boundaries for disposable telemetry; no real Docker or host writes."""
import json
import os
from pathlib import Path
import runpy
import subprocess
import tempfile
import unittest
from unittest.mock import patch

ROOT = Path(__file__).resolve().parents[1]


class TelemetryCleanupSafety(unittest.TestCase):
    def invoke(self, scenario, profile="local"):
        with tempfile.TemporaryDirectory() as directory:
            temporary = Path(directory)
            log = temporary / "calls.jsonl"
            docker = temporary / "docker"
            docker.write_text('''#!/usr/bin/env python3
import json, os, sys
args = sys.argv[1:]
with open(os.environ["TEST_DOCKER_LOG"], "a") as stream:
    stream.write(json.dumps(args) + "\\n")
scenario = os.environ["TEST_DOCKER_SCENARIO"]
if args[0] == "ps" and scenario == "attached":
    print("existing-container")
if args[0] == "run" and scenario in ("recent", "inspection_failure"):
    sys.exit(1)
if args[:2] == ["volume", "rm"] and scenario == "attachment_race":
    sys.exit(1)
''')
            docker.chmod(0o755)
            env = dict(os.environ, PATH=str(temporary) + os.pathsep + os.environ["PATH"],
                       TEST_DOCKER_LOG=str(log), TEST_DOCKER_SCENARIO=scenario)
            result = subprocess.run(["sh", str(ROOT / "infra/observability/clean-development-data.sh"), profile],
                                    env=env, text=True, capture_output=True)
            calls = [json.loads(line) for line in log.read_text().splitlines()] if log.exists() else []
            return result, calls

    def test_recent_data_or_failed_inspection_preserves_every_volume(self):
        for scenario in ("recent", "inspection_failure", "attached"):
            with self.subTest(scenario=scenario):
                result, calls = self.invoke(scenario)
                self.assertEqual(result.returncode, 0)
                self.assertFalse(any(call[:2] == ["volume", "rm"] for call in calls))
                if scenario == "attached":
                    self.assertFalse(any(call[0] == "run" for call in calls))

    def test_expired_volumes_are_scoped_and_inspected_read_only(self):
        for profile in ("local", "dev"):
            with self.subTest(profile=profile):
                result, calls = self.invoke("expired", profile)
                self.assertEqual(result.returncode, 0)
                removed = [call[2] for call in calls if call[:2] == ["volume", "rm"]]
                self.assertEqual(removed, [f"tournaments-manager-{profile}_{service}-data"
                                          for service in ("loki", "tempo", "prometheus", "promtail")])
                for call in (call for call in calls if call[0] == "run"):
                    self.assertIn("none", call)
                    self.assertIn("never", call)
                    self.assertTrue(call[call.index("--mount") + 1].endswith(",readonly"))
                self.assertFalse(any("prune" in call or "--force" in call for call in calls))

    def test_attachment_race_does_not_force_removal(self):
        result, calls = self.invoke("attachment_race")
        self.assertNotEqual(result.returncode, 0)
        self.assertEqual(sum(call[:2] == ["volume", "rm"] for call in calls), 1)
        self.assertFalse(any("--force" in call or "-f" in call for call in calls))

    def test_unknown_profile_cannot_touch_docker(self):
        result, calls = self.invoke("expired", "prod")
        self.assertNotEqual(result.returncode, 0)
        self.assertEqual(calls, [])


class DiskMeasurementSafety(unittest.TestCase):
    def test_ambiguous_or_missing_volume_fails_without_partial_sample(self):
        collect = runpy.run_path(str(ROOT / "infra/k3s/host/observability-disk-metrics.py"))["collect"]
        for paths in ([], ["volume-one", "volume-two"]):
            with self.subTest(paths=paths), patch("glob.glob", return_value=paths), patch("subprocess.check_output") as du:
                with self.assertRaises(RuntimeError):
                    collect()
                du.assert_not_called()

    def test_failed_disk_read_cannot_be_reported_as_zero_usage(self):
        collect = runpy.run_path(str(ROOT / "infra/k3s/host/observability-disk-metrics.py"))["collect"]
        with patch("glob.glob", return_value=["synthetic-volume"]), patch("subprocess.check_output", side_effect=subprocess.TimeoutExpired("du", 60)):
            with self.assertRaises(subprocess.TimeoutExpired):
                collect()


class OnDemandDevelopmentSafety(unittest.TestCase):
    def invoke_stack(self, profile="dev", action="up", running=True):
        with tempfile.TemporaryDirectory() as directory:
            temporary = Path(directory)
            log = temporary / "calls.jsonl"
            docker = temporary / "docker"
            docker.write_text('''#!/usr/bin/env python3
import json, os, sys
args = sys.argv[1:]
with open(os.environ["TEST_DOCKER_LOG"], "a") as stream:
    stream.write(json.dumps({"args": args, "endpoint": os.environ.get("DEV_OTEL_TRACES_ENDPOINT"), "image": os.environ.get("DEV_API_IMAGE"), "profiles": os.environ.get("COMPOSE_PROFILES")}) + "\\n")
if "ps" in args and os.environ["TEST_API_RUNNING"] == "1": print("active-api")
if args[0] == "inspect": print("retained-api:git-original")
''')
            docker.chmod(0o755)
            env = dict(os.environ, PATH=str(temporary) + os.pathsep + os.environ["PATH"],
                       TEST_DOCKER_LOG=str(log), TEST_API_RUNNING="1" if running else "0",
                       COMPOSE_PROFILES="observability", DEV_OTEL_TRACES_ENDPOINT="https://wrong.example")
            result = subprocess.run(["sh", str(ROOT / "infra/observability/development-stack.sh"), profile, action],
                                    env=env, text=True, capture_output=True)
            calls = [json.loads(line) for line in log.read_text().splitlines()] if log.exists() else []
            return result, calls

    def test_opt_in_requires_running_application_and_never_starts_it_implicitly(self):
        result, calls = self.invoke_stack(profile="local", running=False)
        self.assertNotEqual(result.returncode, 0)
        self.assertFalse(any("up" in call["args"] for call in calls))

    def test_toggle_off_preserves_public_runtime_image_and_disables_exporter(self):
        result, calls = self.invoke_stack(action="down")
        self.assertEqual(result.returncode, 0, result.stderr)
        up = next(call for call in calls if "up" in call["args"])
        self.assertEqual(up["image"], "retained-api:git-original")
        self.assertEqual(up["endpoint"], "")
        self.assertEqual(up["profiles"], "")
        self.assertIn("--no-deps", up["args"])
        self.assertIn("--no-build", up["args"])
        self.assertEqual(up["args"][-1], "api")
        self.assertEqual(calls[-1]["args"][-7:], ["stop", "prometheus", "alertmanager", "loki", "promtail", "tempo", "grafana"])

    def test_opt_in_starts_technical_stack_then_enables_internal_exporter(self):
        result, calls = self.invoke_stack(profile="local")
        self.assertEqual(result.returncode, 0, result.stderr)
        ups = [call for call in calls if "up" in call["args"]]
        self.assertEqual(len(ups), 2)
        self.assertEqual(ups[0]["endpoint"], "")
        self.assertEqual(ups[1]["endpoint"], "http://tempo:4318/v1/traces")
        self.assertEqual(ups[1]["args"][-1], "api")
        self.assertTrue(all("--no-build" in call["args"] for call in ups))

    def test_idle_shutdown_and_invalid_target_cannot_start_production_or_application(self):
        result, calls = self.invoke_stack(action="down", running=False)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertFalse(any("up" in call["args"] for call in calls))
        result, calls = self.invoke_stack(profile="prod")
        self.assertNotEqual(result.returncode, 0)
        self.assertEqual(calls, [])


class DevelopmentLaunchAgentSafety(unittest.TestCase):
    def test_shutdown_only_disables_known_dev_agents_and_preserves_plists(self):
        with tempfile.TemporaryDirectory() as directory:
            temporary = Path(directory)
            log = temporary / "calls.jsonl"
            launchctl = temporary / "launchctl"
            launchctl.write_text('''#!/usr/bin/env python3
import json, os, sys
with open(os.environ["TEST_LAUNCHCTL_LOG"], "a") as stream:
    stream.write(json.dumps(sys.argv[1:]) + "\\n")
''')
            launchctl.chmod(0o755)
            uname = temporary / "uname"
            uname.write_text("#!/bin/sh\n echo Darwin\n")
            uname.chmod(0o755)
            agents = temporary / "agents"
            agents.mkdir()
            plist = agents / "com.fasttourney.dev-legal-audit-backup.plist"
            plist.write_text("preserved-installed-plist")
            env = dict(os.environ, PATH=str(temporary) + os.pathsep + os.environ["PATH"],
                       TEST_LAUNCHCTL_LOG=str(log), FASTTOURNEY_DEV_LAUNCH_AGENTS_DIR=str(agents))
            result = subprocess.run(["sh", str(ROOT / "infra/home/dev-launch-agents.sh"), "down"],
                                    env=env, text=True, capture_output=True)
            self.assertEqual(result.returncode, 0, result.stderr)
            calls = [json.loads(line) for line in log.read_text().splitlines()]
            mutations = [call for call in calls if call[0] in ("disable", "bootout")]
            self.assertEqual(len(mutations), 12)
            self.assertTrue(all("/com.fasttourney.dev-" in call[1] for call in mutations))
            self.assertFalse(any("prod" in str(call) for call in calls))
            self.assertEqual(plist.read_text(), "preserved-installed-plist")
            log.write_text("")
            result = subprocess.run(["sh", str(ROOT / "infra/home/dev-launch-agents.sh"), "up"],
                                    env=env, text=True, capture_output=True)
            self.assertEqual(result.returncode, 0, result.stderr)
            calls = [json.loads(line) for line in log.read_text().splitlines()]
            enabled = [call for call in calls if call[0] == "enable"]
            self.assertEqual(len(enabled), 1)
            self.assertTrue(enabled[0][1].endswith("/com.fasttourney.dev-legal-audit-backup"))


if __name__ == "__main__":
    unittest.main()
