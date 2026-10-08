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

configured_tip_links=0
for amount in 2 5 10; do
  variable_name="EXPO_PUBLIC_TIP_PAYMENT_LINK_${amount}_EUR"
  eval "variable_value=\${$variable_name:-}"
  if [ -n "$variable_value" ]; then
    case "$variable_value" in
      https://buy.stripe.com/test_*)
        echo "$variable_name no puede contener un Payment Link de prueba en producción." >&2
        exit 1
        ;;
      https://buy.stripe.com/*) configured_tip_links=$((configured_tip_links + 1)) ;;
      *)
        echo "$variable_name debe contener un Payment Link live de Stripe." >&2
        exit 1
        ;;
    esac
  fi
done
if [ "$configured_tip_links" -ne 0 ] && [ "$configured_tip_links" -ne 3 ]; then
  echo "Los Payment Links live deben configurarse juntos para 2 €, 5 € y 10 €." >&2
  exit 1
fi

node "$repository_root/infra/app-links/prepare.mjs" production "$app_links_directory"

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
  pnpm --filter @tournaments-manager/client exec expo export --platform web --source-maps --clear --output-dir "$staging_directory"

node "$repository_root/infra/home/prepare-web-telemetry.mjs" production "$staging_directory" "$release_sha"

node "$repository_root/infra/app-links/prepare.mjs" production "$app_links_directory" "$staging_directory"

deployed_at=$(date -u +%Y-%m-%dT%H:%M:%SZ)
printf '{\n  "commit": "%s",\n  "builtAt": "%s"\n}\n' \
  "$release_sha" "$deployed_at" > "$staging_directory/deployment.json"

mv "$staging_directory" "$release_directory"
trap - EXIT
printf 'Release web de prod preparado, sin activar: %s\n' "$release_sha"
