#!/bin/sh
set -eu

repository_root=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
release_sha=${1:?"Uso: deploy-dev-web.sh <SHA-completo>"}
release_root=/opt/homebrew/var/www/fasttourney/dev/releases
release_directory="$release_root/$release_sha"
config_file=${FASTTOURNEY_DEV_WEB_CONFIG:-$repository_root/infra/home/secrets/development-web.env}

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

if [ -r "$config_file" ]; then
	# La configuración de desarrollo es opcional: sin enlaces de prueba, la
	# superficie de propinas permanece oculta en dev.
	set -a
	# shellcheck disable=SC1090
	. "$config_file"
	set +a
fi

configured_tip_links=0
for amount in 2 5 10; do
	variable_name="EXPO_PUBLIC_TIP_PAYMENT_LINK_${amount}_EUR"
	eval "variable_value=\${$variable_name:-}"
	if [ -n "$variable_value" ]; then
		case "$variable_value" in
			https://buy.stripe.com/test_*) configured_tip_links=$((configured_tip_links + 1)) ;;
			*)
				echo "$variable_name debe contener un Payment Link de prueba de Stripe." >&2
				exit 1
				;;
		esac
	fi
done
if [ "$configured_tip_links" -ne 0 ] && [ "$configured_tip_links" -ne 3 ]; then
	echo "Los Payment Links de prueba deben configurarse juntos para 2 €, 5 € y 10 €." >&2
	exit 1
fi

cd "$repository_root"
EXPO_NO_DOTENV=1 \
	APP_ENV=development \
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
