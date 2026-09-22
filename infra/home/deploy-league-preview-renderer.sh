#!/bin/sh
set -eu

# Promueve el renderer del mismo SHA que la API, reinicia solo su entorno y
# verifica que el proceso nuevo sirve el shell activo. Si readiness falla,
# restaura el binario anterior y vuelve a arrancarlo.
repository_root=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
environment=${1:?"Uso: deploy-league-preview-renderer.sh <dev|prod> <SHA-completo>"}
release_sha=${2:?"Uso: deploy-league-preview-renderer.sh <dev|prod> <SHA-completo>"}

case "$environment" in
  dev)
    port=8092
    public_host=dev.fasttourney.com
    service=com.fasttourney.dev-league-preview-renderer
    ;;
  prod)
    port=8091
    public_host=fasttourney.com
    service=com.fasttourney.prod-league-preview-renderer
    ;;
  *) echo "El entorno debe ser dev o prod." >&2; exit 1 ;;
esac

target=/opt/homebrew/var/www/fasttourney/bin/$environment/league-preview-renderer
target_directory=$(dirname -- "$target")
backup=
restored=false
if [ -f "$target" ]; then
  backup=$(mktemp "$target_directory/.league-preview-renderer.previous.XXXXXX")
  cp -p "$target" "$backup"
fi

restore_previous() {
  if [ "$restored" = true ]; then return; fi
  restored=true
  if [ -n "$backup" ] && [ -f "$backup" ]; then
    mv "$backup" "$target"
    launchctl kickstart -k "gui/$(id -u)/$service" >/dev/null 2>&1 || true
  elif [ -f "$target" ]; then
    rm -f -- "$target"
  fi
}
trap 'restore_previous' EXIT
trap 'exit 1' HUP INT TERM

"$repository_root/infra/home/build-league-preview-renderer.sh" "$environment" "$release_sha"
launchctl kickstart -k "gui/$(id -u)/$service"

attempt=0
while [ "$attempt" -lt 20 ]; do
  readiness=$(curl --silent --show-error --fail -H "Host: $public_host" "http://127.0.0.1:$port/-/ready" 2>/dev/null || true)
  if [ "$(printf '%s' "$readiness" | jq -r '.revision // empty' 2>/dev/null)" = "$release_sha" ]; then
    if [ -n "$backup" ]; then rm -f "$backup"; fi
    trap - EXIT HUP INT TERM
    printf 'Renderer %s desplegado y verificado: %s\n' "$environment" "$release_sha"
    exit 0
  fi
  attempt=$((attempt + 1))
  sleep 1
done

echo "El renderer $environment no publicó readiness para $release_sha; se restaura el anterior." >&2
exit 1
