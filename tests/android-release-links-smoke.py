"""Opt-in smoke test for an installed local Android Release and existing QA data.

Does not start services, install APKs, create fixtures, or certify screen readers.
Cases JSON: [{"name": "final", "url": "fasttourney-local://...",
              "text": ["Final"], "description": ["Ganador"]}].
"""

import argparse
import json
from pathlib import Path
import re
import shlex
import subprocess
import time
import xml.etree.ElementTree as ET


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("cases", type=Path)
    parser.add_argument("--adb", default="adb")
    parser.add_argument("--serial", required=True)
    parser.add_argument("--evidence", type=Path, required=True)
    parser.add_argument("--timeout", type=float, default=35,
                        help="Seconds allowed for visible content per case (default: 35).")
    args = parser.parse_args()
    if not args.serial.startswith("emulator-"):
        parser.error("Use an emulator serial; this QA does not target physical devices.")
    if not 0 < args.timeout <= 120:
        parser.error("Use a content timeout greater than 0 and at most 120 seconds.")
    cases = json.loads(args.cases.read_text())
    if not cases:
        parser.error("Supply at least one case.")
    for case in cases:
        if not re.fullmatch(r"[a-z0-9-]+", case["name"]):
            parser.error("Case names must contain lowercase letters, digits or hyphens.")
        if not case["url"].startswith("fasttourney-local://tournament/"):
            parser.error("Only local tournament links are allowed.")
        if not case.get("text") and not case.get("description"):
            parser.error("Each case must declare expected visible content.")

    def adb(*command, check=True):
        return subprocess.run(
            [args.adb, "-s", args.serial, *command],
            capture_output=True, check=check, timeout=45,
        )

    def shell(*command, check=True):
        return adb("shell", shlex.join(command), check=check)

    # UIAutomator inspection can interfere with accessibility service focus.
    services = shell("settings", "get", "secure", "enabled_accessibility_services")
    if services.stdout.strip() not in (b"", b"null"):
        parser.error("Disable screen readers before running this UI inspection smoke test.")
    input_state = shell("dumpsys", "input").stdout
    input_mode = re.search(rb"Display: 0 TouchMode: (\S+)", input_state)
    package = "com.fasttourney.app.local"
    debug_probe = shell("run-as", package, "true", check=False)
    if debug_probe.returncode == 0 or b"not debuggable" not in debug_probe.stderr:
        parser.error("Install a non-debuggable local Release before running this test.")
    sizes = re.findall(rb"(\d+)x(\d+)", shell("wm", "size").stdout)
    if not sizes:
        parser.error("Unable to determine the emulator screen size.")
    width, height = map(int, sizes[-1])
    args.evidence.mkdir(parents=True, exist_ok=True)
    (args.evidence / "input-mode.txt").write_bytes(
        b"TouchMode: " + (input_mode.group(1) if input_mode else b"unknown") + b"\n"
    )
    failures = []
    for case in cases:
        shell("am", "force-stop", package)
        launch = shell("am", "start", "-W", "-a", "android.intent.action.VIEW",
                       "-d", case["url"], package)
        (args.evidence / f'{case["name"]}-launch.txt').write_bytes(launch.stdout)
        cold = b"LaunchState: COLD" in launch.stdout
        deadline = time.monotonic() + args.timeout
        matched = False
        while time.monotonic() < deadline:
            device_xml = "/sdcard/tm-qa-release-links.xml"
            shell("uiautomator", "dump", device_xml)
            xml = shell("cat", device_xml).stdout
            (args.evidence / f'{case["name"]}.xml').write_bytes(xml)
            root = ET.fromstring(xml)
            values = {key: set() for key in ("text", "content-desc")}
            for node in root.iter("node"):
                bounds = [int(n) for n in re.findall(r"\d+", node.get("bounds", ""))]
                # Dumped nodes can include off-screen content; require screen intersection.
                if (len(bounds) != 4 or bounds[2] <= bounds[0] or bounds[3] <= bounds[1]
                        or bounds[0] >= width or bounds[1] >= height
                        or bounds[2] <= 0 or bounds[3] <= 0):
                    continue
                for key in values:
                    values[key].add(node.get(key, ""))
            matched = all(t in values["text"] for t in case.get("text", [])) and all(
                d in values["content-desc"] for d in case.get("description", [])
            )
            if matched:
                break
        screenshot = adb("exec-out", "screencap", "-p")
        (args.evidence / f'{case["name"]}.png').write_bytes(screenshot.stdout)
        passed = cold and matched
        print(f'{"PASS" if passed else "FAIL"}: {case["name"]} (cold={cold}, content={matched})',
              flush=True)
        if not passed:
            failures.append(case["name"])
    raise SystemExit(1 if failures else 0)


if __name__ == "__main__":
    main()
