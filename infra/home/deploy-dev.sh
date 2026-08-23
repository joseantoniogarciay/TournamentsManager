#!/bin/sh
set -eu

repository_root=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
release_root=/opt/homebrew/var/www/fasttourney/dev/releases
current_link=/opt/homebrew/var/www/fasttourney/dev/current
api_repository=tournaments-manager-dev-api

cd "$repository_root"

diagnose_compose_start_failure() {
  echo "El arranque de Compose no completó la comprobación de salud." >&2
  echo "Estado de los servicios de dev:" >&2
  docker compose --env-file infra/dev/.env -f infra/dev/compose.yaml ps >&2 || true
  echo "Últimos logs de API y PostgreSQL:" >&2
  docker compose --env-file infra/dev/.env -f infra/dev/compose.yaml logs --tail=200 api postgres >&2 || true
}

# La API debe recibir únicamente la identidad PostgreSQL restringida. Se
# comprueba antes de construir o conmutar artefactos y sin imprimir secretos.
set -a
. infra/dev/.env
. infra/dev/api.docker.env
set +a

if [ ! -s infra/dev/alertmanager.smtp-password ]; then
  echo "Falta infra/dev/alertmanager.smtp-password con la clave SMTP exclusiva de Alertmanager." >&2
  exit 1
fi

if grep -Fqx 'replace-with-resend-alerts-sending-access-key' infra/dev/alertmanager.smtp-password; then
  echo "infra/dev/alertmanager.smtp-password conserva el valor de ejemplo." >&2
  exit 1
fi

if ! awk '
  NR != 1 || $0 !~ /^re_[A-Za-z0-9_-]+$/ { valid = 0; next }
  { valid = 1 }
  END { exit !(valid && NR == 1) }
' infra/dev/alertmanager.smtp-password; then
  echo "infra/dev/alertmanager.smtp-password debe contener solo una clave Resend re_... en una unica linea." >&2
  exit 1
fi

expected_database_url="postgres://${POSTGRES_APP_USER}:${POSTGRES_APP_PASSWORD}@postgres:5432/${POSTGRES_DB}?sslmode=disable"
if [ "${DATABASE_URL:-}" != "$expected_database_url" ]; then
  echo "DATABASE_URL de dev debe usar exactamente POSTGRES_APP_USER y POSTGRES_APP_PASSWORD." >&2
  exit 1
fi

if [ "${OTEL_TRACES_ENDPOINT:-}" != "http://tempo:4318/v1/traces" ]; then
  echo "OTEL_TRACES_ENDPOINT de dev debe dirigir las trazas a Tempo interno." >&2
  exit 1
fi

branch=$(git branch --show-current)
if [ "$branch" != "develop" ]; then
  echo "El despliegue dev solo se ejecuta desde develop; rama actual: $branch." >&2
  exit 1
fi

if [ -n "$(git status --porcelain)" ]; then
  echo "El árbol Git debe estar limpio antes de desplegar dev." >&2
  exit 1
fi

release_sha=$(git rev-parse HEAD)
remote_sha=$(git rev-parse origin/develop)
if [ "$release_sha" != "$remote_sha" ]; then
  echo "HEAD debe coincidir con origin/develop antes de desplegar." >&2
  exit 1
fi

api_image="$api_repository:git-$release_sha"
if docker image inspect "$api_image" >/dev/null 2>&1; then
  echo "La imagen $api_image ya existe; el SHA no se reutiliza." >&2
  exit 1
fi

docker build --target runtime --tag "$api_image" apps/backend

# Las migraciones son forward-only y ocurren antes de sustituir la API. El
# despliegue no puede llegar a ejecutar código que espere un esquema inexistente.
make dev-public-migrate

./infra/home/deploy-dev-web.sh "$release_sha"

if ! DEV_API_IMAGE="$api_image" \
  docker compose --env-file infra/dev/.env -f infra/dev/compose.yaml up --detach --wait --remove-orphans; then
  diagnose_compose_start_failure
  exit 1
fi

next_link="$current_link.next"
ln -s "releases/$release_sha" "$next_link"
mv -f -h "$next_link" "$current_link"

deployed_at=$(date -u +%Y-%m-%dT%H:%M:%SZ)
printf '{\n  "commit": "%s",\n  "apiImage": "%s",\n  "deployedAt": "%s"\n}\n' \
  "$release_sha" "$api_image" "$deployed_at" >"$release_root/$release_sha/deployment.json"

for old_release in $(ls -1dt "$release_root"/* 2>/dev/null | sed -n '3,$p'); do
  old_sha=$(basename "$old_release")
  rm -r -- "$old_release"
  docker image rm "$api_repository:git-$old_sha" >/dev/null 2>&1 || true
done

printf 'Dev desplegado: %s\n' "$release_sha"
