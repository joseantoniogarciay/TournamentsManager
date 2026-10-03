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


if __name__ == "__main__":
    unittest.main()
