#!/bin/sh
# Only entirely expired, unattached telemetry is disposable.
set -eu
case "${1:-}" in
    local) project=tournaments-manager-local ;;
    dev) project=tournaments-manager-dev ;;
    *) echo 'Usage: clean-development-data.sh local|dev' >&2; exit 1 ;;
esac

for service in loki tempo prometheus promtail; do
    containers=$(docker ps --all --quiet \
        --filter "label=com.docker.compose.project=$project" \
        --filter "label=com.docker.compose.service=$service")
    if [ -n "$containers" ]; then
        echo "$project/$service: container present; automatic retention handles its data"
        continue
    fi
    volume="${project}_${service}-data"
    if docker volume inspect "$volume" >/dev/null 2>&1; then
        # Read-only inspection. Any recent file or inspection failure preserves
        # the complete volume; a running writer is excluded above. Never force
        # removal if a container attaches the volume during this check.
        if docker run --rm --pull never --network none --entrypoint sh \
            --mount "type=volume,source=$volume,target=/data,readonly" \
            grafana/promtail:3.5.0 -ec \
            'recent=$(find /data -type f -mmin -1440 -print -quit); test -z "$recent"'; then
            docker volume rm "$volume"
        else
            echo "$volume: preserved; recent data or age could not be verified"
        fi
    fi
done
