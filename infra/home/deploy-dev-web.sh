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
	EXPO_PUBLIC_API_BASE_URL=https://dev-api.fasttourney.com/v1 \
	EXPO_PUBLIC_APP_LINK_URL=https://dev.fasttourney.com \
	EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID=267995166257-2favmuuhnu14p9na8le1rlmtpbgcb56g.apps.googleusercontent.com \
	EXPO_PUBLIC_GOOGLE_IOS_CLIENT_ID=267995166257-75hpab46tjjbcho5av9d6qfo2s3md80b.apps.googleusercontent.com \
	EXPO_PUBLIC_GOOGLE_ANDROID_CLIENT_ID=267995166257-rlqmqi79b16fddta929ah9o8f6792afc.apps.googleusercontent.com \
	pnpm --filter @tournaments-manager/client exec expo export --platform web --clear --output-dir "$staging_directory"

api_docs_directory="$staging_directory/api-docs"
install -d -m 755 "$api_docs_directory"
install -m 644 infra/home/api-docs/index.html "$api_docs_directory/index.html"
install -m 644 contracts/openapi/v1/openapi.yaml "$api_docs_directory/openapi.yaml"

mv "$staging_directory" "$release_directory"
trap - EXIT
