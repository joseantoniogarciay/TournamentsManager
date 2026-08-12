#!/bin/sh
set -eu

repository_root=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
release_sha=${1:?"Uso: deploy-dev-web.sh <SHA-completo>"}
release_root=/opt/homebrew/var/www/fasttourney/dev/releases
release_directory="$release_root/$release_sha"

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

if [ -e "$release_directory" ]; then
  echo "La web de $release_sha ya existe; no se sobrescribe." >&2
  exit 1
fi

install -d -m 755 "$release_root"
staging_directory=$(mktemp -d "$release_root/.staging.XXXXXX")
trap 'rm -rf "$staging_directory"' EXIT

cd "$repository_root"
EXPO_NO_DOTENV=1 \
	EXPO_UNSTABLE_WEB_MODAL=1 \
	EXPO_PUBLIC_API_BASE_URL=https://dev-api.fasttourney.com/v1 \
	EXPO_PUBLIC_APP_LINK_URL=https://dev.fasttourney.com \
	pnpm --filter @tournaments-manager/client exec expo export --platform web --clear --output-dir "$staging_directory"

mv "$staging_directory" "$release_directory"
trap - EXIT
