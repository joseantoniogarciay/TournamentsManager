#!/bin/sh
set -eu

release_sha=${1:?"Uso: rollback-dev.sh <SHA-completo>"}
release_root=/opt/homebrew/var/www/fasttourney/dev/releases
release_directory="$release_root/$release_sha"
current_link=/opt/homebrew/var/www/fasttourney/dev/current
api_repository=tournaments-manager-dev-api
api_image="$api_repository:git-$release_sha"

case "$release_sha" in
  *[!0123456789abcdef]*)
    echo "El SHA debe tener 40 caracteres hexadecimales." >&2
    exit 1
    ;;
esac

if [ "${#release_sha}" -ne 40 ]; then
  echo "El SHA debe tener 40 caracteres hexadecimales." >&2
  exit 1
fi

if [ ! -f "$release_directory/deployment.json" ]; then
  echo "No existe un despliegue dev conservado para $release_sha." >&2
  exit 1
fi

if ! docker image inspect "$api_image" >/dev/null 2>&1; then
  echo "No se conserva la imagen $api_image; no es posible este rollback." >&2
  exit 1
fi

repository_root=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
cd "$repository_root"
DEV_API_IMAGE="$api_image" \
  docker compose --env-file infra/dev/.env -f infra/dev/compose.yaml up --detach --wait --remove-orphans

next_link="$current_link.next"
ln -s "releases/$release_sha" "$next_link"
mv -f "$next_link" "$current_link"

printf 'Dev restaurado: %s\n' "$release_sha"
