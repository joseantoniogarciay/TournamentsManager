#!/bin/sh
set -eu

# Construye un release estático de prod sin cambiar el enlace `current` ni
# publicar tráfico. La activación es un paso deliberado separado.
repository_root=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
release_sha=${1:?"Uso: stage-prod-web.sh <SHA-completo>"}
release_root=/opt/homebrew/var/www/fasttourney/prod/releases
release_directory="$release_root/$release_sha"
config_file=${FASTTOURNEY_PROD_WEB_CONFIG:-$repository_root/infra/home/secrets/production-web.env}
app_links_directory=${FASTTOURNEY_PROD_APP_LINKS_DIR:-$repository_root/infra/home/secrets/app-links}

case "$release_sha" in
  *[!0123456789abcdef]* | "")
    echo "El SHA debe tener 40 caracteres hexadecimales." >&2
    exit 1
    ;;
esac
if [ "${#release_sha}" -ne 40 ]; then
  echo "El SHA debe tener 40 caracteres hexadecimales." >&2
  exit 1
fi

if [ ! -r "$config_file" ]; then
  echo "Falta la configuración privada de exportación: $config_file." >&2
  exit 1
fi
# shellcheck disable=SC1090
set -a
. "$config_file"
set +a
for required_variable in EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID; do
  eval "required_value=\${$required_variable:-}"
  if [ -z "$required_value" ] || printf '%s' "$required_value" | grep -q 'replace-with-'; then
    echo "$required_variable debe estar definido con el ID de producción." >&2
    exit 1
  fi
done

has_app_links=false
if [ -e "$app_links_directory/apple-app-site-association" ] || \
  [ -e "$app_links_directory/assetlinks.json" ]; then
  if [ ! -r "$app_links_directory/apple-app-site-association" ] || \
    [ ! -r "$app_links_directory/assetlinks.json" ]; then
    echo "Las asociaciones móviles deben aportar ambos ficheros en $app_links_directory." >&2
    exit 1
  fi
  if ! jq -e '.applinks.details[0].appIDs[0] | test("^[A-Z0-9]+\\.com\\.fasttourney\\.app$")' \
    "$app_links_directory/apple-app-site-association" >/dev/null; then
    echo "apple-app-site-association no contiene el Team ID real de producción." >&2
    exit 1
  fi
  if ! jq -e '.[]?.target | select(.package_name == "com.fasttourney.app") | .sha256_cert_fingerprints[] | test("^[A-F0-9:]{95}$")' \
    "$app_links_directory/assetlinks.json" >/dev/null; then
    echo "assetlinks.json no contiene una huella SHA-256 real de producción." >&2
    exit 1
  fi
  has_app_links=true
fi

cd "$repository_root"
if [ -n "$(git status --porcelain)" ]; then
  echo "El árbol Git debe estar limpio antes de preparar prod." >&2
  exit 1
fi
if [ "$(git rev-parse HEAD)" != "$release_sha" ]; then
  echo "El SHA indicado debe coincidir con HEAD para que el artefacto sea trazable." >&2
  exit 1
fi
if [ -e "$release_directory" ]; then
  echo "El release web de $release_sha ya existe; no se sobrescribe." >&2
  exit 1
fi

install -d -m 755 "$release_root"
staging_directory=$(mktemp -d "$release_root/.staging.XXXXXX")
trap 'rm -rf "$staging_directory"' EXIT

EXPO_NO_DOTENV=1 \
  APP_ENV=production \
  EXPO_PUBLIC_API_BASE_URL=https://api.fasttourney.com/v1 \
  EXPO_PUBLIC_APP_LINK_URL=https://fasttourney.com \
  pnpm --filter @tournaments-manager/client exec expo export --platform web --clear --output-dir "$staging_directory"

if [ "$has_app_links" = true ]; then
  install -d -m 755 "$staging_directory/.well-known"
  install -m 644 "$app_links_directory/apple-app-site-association" \
    "$staging_directory/.well-known/apple-app-site-association"
  install -m 644 "$app_links_directory/assetlinks.json" \
    "$staging_directory/.well-known/assetlinks.json"
fi

deployed_at=$(date -u +%Y-%m-%dT%H:%M:%SZ)
printf '{\n  "commit": "%s",\n  "builtAt": "%s"\n}\n' \
  "$release_sha" "$deployed_at" > "$staging_directory/deployment.json"

mv "$staging_directory" "$release_directory"
trap - EXIT
if [ "$has_app_links" = true ]; then
  printf 'Release web de prod preparado con asociaciones móviles, sin activar: %s\n' "$release_sha"
else
  printf 'Release web de prod preparado sin asociaciones móviles, sin activar: %s\n' "$release_sha"
fi
