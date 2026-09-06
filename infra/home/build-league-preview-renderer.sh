#!/bin/sh
set -eu

# Compila de forma atómica el pequeño proceso local que da metadatos sociales a
# /league/{uuid}. No publica tráfico: Caddy y launchd se actualizan aparte.
repository_root=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
target=${FASTTOURNEY_LEAGUE_PREVIEW_BINARY:-/opt/homebrew/var/www/fasttourney/bin/league-preview-renderer}
target_directory=$(dirname -- "$target")

install -d -m 755 "$target_directory"
staging_binary=$(mktemp "$target_directory/.league-preview-renderer.XXXXXX")
trap 'rm -f "$staging_binary"' EXIT

go -C "$repository_root/apps/backend" build -o "$staging_binary" ./cmd/league-preview-renderer
chmod 755 "$staging_binary"
mv "$staging_binary" "$target"
trap - EXIT
printf 'Renderer de previews preparado en %s\n' "$target"
