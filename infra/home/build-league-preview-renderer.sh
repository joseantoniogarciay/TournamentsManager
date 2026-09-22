#!/bin/sh
set -eu

# Compila de forma atómica el renderer propio de un entorno y le incrusta el SHA
# que debe exponer en readiness. No reinicia launchd; esa promoción pertenece a
# deploy-league-preview-renderer.sh.
repository_root=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
environment=${1:?"Uso: build-league-preview-renderer.sh <dev|prod> <SHA-completo>"}
release_sha=${2:?"Uso: build-league-preview-renderer.sh <dev|prod> <SHA-completo>"}

case "$environment" in
  dev | prod) ;;
  *) echo "El entorno debe ser dev o prod." >&2; exit 1 ;;
esac
case "$release_sha" in
  *[!0123456789abcdef]* | "") echo "El SHA debe tener 40 caracteres hexadecimales." >&2; exit 1 ;;
esac
if [ "${#release_sha}" -ne 40 ] || [ "$(git -C "$repository_root" rev-parse HEAD)" != "$release_sha" ]; then
  echo "El SHA debe coincidir con HEAD para que el renderer sea trazable." >&2
  exit 1
fi
if [ -n "$(git -C "$repository_root" status --porcelain)" ]; then
  echo "El árbol Git debe estar limpio antes de compilar el renderer." >&2
  exit 1
fi

target=${FASTTOURNEY_LEAGUE_PREVIEW_BINARY:-/opt/homebrew/var/www/fasttourney/bin/$environment/league-preview-renderer}
target_directory=$(dirname -- "$target")

install -d -m 755 "$target_directory"
staging_binary=$(mktemp "$target_directory/.league-preview-renderer.XXXXXX")
trap 'rm -f "$staging_binary"' EXIT

go -C "$repository_root/apps/backend" build -ldflags "-X main.revision=$release_sha" -o "$staging_binary" ./cmd/league-preview-renderer
chmod 755 "$staging_binary"
mv "$staging_binary" "$target"
trap - EXIT
printf 'Renderer %s de %s preparado en %s\n' "$environment" "$release_sha" "$target"
