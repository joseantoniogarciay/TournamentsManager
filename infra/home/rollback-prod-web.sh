#!/bin/sh
set -eu

# Recupera un artefacto web existente. No toca la API ni los datos.
exec "$(dirname -- "$0")/activate-prod-web.sh" "$@"
