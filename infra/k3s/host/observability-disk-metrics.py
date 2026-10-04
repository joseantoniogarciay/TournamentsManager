#!/usr/bin/env python3
"""Export only aggregate production storage measurements; never delete data."""
import glob
import os
from pathlib import Path
import subprocess
import time

OUTPUT = Path("/var/lib/fasttourney/observability-metrics/disk.prom")
PATTERNS = {
    "loki": "*_prod_storage-observability-loki-0",
    "tempo": "*_prod_storage-observability-tempo-0",
    "prometheus": "*_prod_observability-prometheus-server",
}


def collect():
    root = os.statvfs("/")
    lines = [
        "# TYPE fasttourney_host_disk_size_bytes gauge",
        f"fasttourney_host_disk_size_bytes {root.f_blocks * root.f_frsize}",
        "# TYPE fasttourney_host_disk_available_bytes gauge",
        f"fasttourney_host_disk_available_bytes {root.f_bavail * root.f_frsize}",
        "# TYPE fasttourney_telemetry_storage_bytes gauge",
    ]
    for service, pattern in PATTERNS.items():
        matches = glob.glob("/var/lib/rancher/k3s/storage/" + pattern)
        if len(matches) != 1:
            raise RuntimeError("Expected exactly one telemetry volume for " + service)
        result = subprocess.check_output(
            ["du", "--block-size=1", "-s", matches[0]], text=True, timeout=60
        )
        size = int(result.split()[0])
        lines.append(f'fasttourney_telemetry_storage_bytes{{service="{service}"}} {size}')
    lines.extend([
        "# TYPE fasttourney_disk_check_timestamp_seconds gauge",
        f"fasttourney_disk_check_timestamp_seconds {int(time.time())}",
    ])
    return "\n".join(lines) + "\n"


if __name__ == "__main__":
    content = collect()  # A failure preserves the previous sample and timestamp.
    OUTPUT.parent.mkdir(parents=True, exist_ok=True)
    temporary = OUTPUT.with_suffix(".tmp")
    with temporary.open("w") as stream:
        stream.write(content)
    temporary.chmod(0o644)
    os.replace(temporary, OUTPUT)
