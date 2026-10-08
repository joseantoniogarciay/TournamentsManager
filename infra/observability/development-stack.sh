#!/bin/sh
# Explicit diagnostic session; no production targets, builds, or data deletion.
set -eu
repository_root=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
cd "$repository_root"
case "${1:-}" in
    local) directory=infra/local; configuration=compose.dev.yaml ;;
    dev) directory=infra/dev; configuration=compose.yaml ;;
    *) echo 'Usage: development-stack.sh local|dev up|down' >&2; exit 1 ;;
esac
case "${2:-}" in up|down) action=$2 ;; *) echo 'Usage: development-stack.sh local|dev up|down' >&2; exit 1 ;; esac
export COMPOSE_PROFILES=
export DEV_OTEL_TRACES_ENDPOINT=
compose() {
    docker compose --env-file "$directory/.env" -f "$directory/$configuration" --profile observability "$@"
}
api_container=$(compose ps --status running --quiet api)
if [ "$action" = up ] && [ -z "$api_container" ]; then
    echo 'Primero inicia dev para pruebas; la observabilidad no arranca la aplicación.' >&2
    exit 1
fi
# Public dev may run a retained git-SHA image rather than :latest. Never promote
# another image merely to toggle the exporter.
if [ -n "$api_container" ] && [ "$1" = dev ]; then
    DEV_API_IMAGE=$(docker inspect --format '{{.Config.Image}}' "$api_container")
    export DEV_API_IMAGE
fi
if [ "$action" = up ]; then
    if [ "$1" = dev ] && ! awk 'NR != 1 || $0 !~ /^re_[A-Za-z0-9_-]+$/ { bad = 1 } END { exit (bad || NR != 1) }' infra/dev/alertmanager.smtp-password; then
        echo 'Falta la clave Resend exclusiva y válida de Alertmanager dev.' >&2
        exit 1
    fi
    compose up --no-build --detach --wait prometheus alertmanager loki promtail tempo grafana
    DEV_OTEL_TRACES_ENDPOINT=http://tempo:4318/v1/traces
    export DEV_OTEL_TRACES_ENDPOINT
    compose up --no-build --no-deps --detach --wait api
else
    if [ -n "$api_container" ]; then
        compose up --no-build --no-deps --detach --wait api
    fi
    compose stop prometheus alertmanager loki promtail tempo grafana
fi
