#!/usr/bin/env bash

# Importa en la VM K3s las imágenes API y migrator ya copiadas a /tmp, crea el
# Secret mínimo de la API desde la contraseña runtime de PostgreSQL y despliega
# la API. Se ejecuta en el Mac, no dentro de la VM.
set -euo pipefail

script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
repo_root=$(cd -- "$script_dir/../../.." && pwd)

# shellcheck disable=SC1091
source "$repo_root/infra/k3s/.env"

integrations_file=${FASTTOURNEY_PROD_API_INTEGRATIONS_FILE:-$repo_root/infra/k3s/secrets/api-integrations.env}
if [ ! -r "$integrations_file" ]; then
  echo "Falta $integrations_file con las integraciones de producción de la API." >&2
  exit 1
fi

# shellcheck disable=SC1090
set -a
. "$integrations_file"
set +a
for required_variable in SMTP_USERNAME SMTP_PASSWORD GOOGLE_CLIENT_IDS; do
  eval "required_value=\${$required_variable:-}"
  if [ -z "$required_value" ] || printf '%s' "$required_value" | grep -q 'replace-with-'; then
    echo "$required_variable debe estar definido en $integrations_file." >&2
    exit 1
  fi
done
unset SMTP_USERNAME SMTP_PASSWORD GOOGLE_CLIENT_IDS required_value

image_revision=$(sed -n 's/^[[:space:]]*image: tournaments-manager-api:git-\([[:xdigit:]]\{7,40\}\)$/\1/p' "$repo_root/infra/k3s/core/api.yaml")
if [ -z "$image_revision" ] || [ "$(printf '%s\n' "$image_revision" | wc -l | tr -d ' ')" -ne 1 ]; then
  echo "infra/k3s/core/api.yaml debe declarar exactamente una imagen API con SHA Git." >&2
  exit 1
fi
release_sha=$(git -C "$repo_root" rev-parse "$image_revision^{commit}")
if [ -n "$(git -C "$repo_root" status --porcelain)" ] || [ "$(git -C "$repo_root" rev-parse HEAD)" != "$release_sha" ]; then
  echo "HEAD limpio debe coincidir con el SHA de la imagen API antes de desplegar producción." >&2
  exit 1
fi

# La transferencia crea el fichero remoto con modo 0600 desde el inicio; scp
# podría aplicar el umask remoto y dejar una ventana con permisos más amplios.
ssh "$K3S_SSH_USER@$K3S_SSH_HOST" \
  'umask 077; cat > /tmp/tournaments-manager-api-integrations.env' < "$integrations_file"

ssh "$K3S_SSH_USER@$K3S_SSH_HOST" \
  'cat > /tmp/tournaments-manager-deploy-api.sh && chmod 700 /tmp/tournaments-manager-deploy-api.sh' <<'REMOTE'
set -euo pipefail

work_dir=/tmp/tournaments-manager-k3s
secret_file=/tmp/api-runtime.env
migration_secret_file=/tmp/api-migrations-runtime.env
integrations_file=/tmp/tournaments-manager-api-integrations.env
trap 'rm -f "$secret_file" "$migration_secret_file" "$integrations_file"; sudo /usr/local/bin/k3s kubectl -n prod delete secret api-migrations-runtime --ignore-not-found >/dev/null 2>&1 || true' EXIT

printf '%s\n' 'Validando sudo en la VM...'
sudo -v

printf '%s\n' 'Importando imagen runtime de la API...'
sudo /usr/local/bin/k3s ctr images import "$work_dir/tournaments-manager-api.tar"
printf '%s\n' 'Importando imagen migrator...'
sudo /usr/local/bin/k3s ctr images import "$work_dir/tournaments-manager-migrator.tar"
sudo /usr/local/bin/k3s ctr images ls | grep tournaments-manager

# Goose se ejecuta antes de la API con la identidad migradora, nunca con la
# credencial runtime. El Secret es efímero: se elimina al terminar este script.
api_image="$(sed -n 's/^[[:space:]]*image: \(tournaments-manager-api:git-[[:alnum:]]*\)$/\1/p' "$work_dir/api.yaml")"
if [ -z "$api_image" ] || [ "$(printf '%s\n' "$api_image" | wc -l | tr -d ' ')" -ne 1 ]; then
  echo 'api.yaml debe declarar exactamente una imagen tournaments-manager-api:git-<SHA>.' >&2
  exit 1
fi
migrator_image="${api_image/tournaments-manager-api:/tournaments-manager-migrator:}"

printf '%s\n' 'Creando el Secret efímero de migración...'
umask 077
migrator_password="$(
  sudo /usr/local/bin/k3s kubectl -n prod get secret postgresql-runtime \
    -o jsonpath='{.data.POSTGRES_MIGRATOR_PASSWORD}' | base64 --decode
)"
export migrator_password
encoded_migrator_password="$(python3 -c 'import os, urllib.parse; print(urllib.parse.quote(os.environ["migrator_password"], safe=""))')"
unset migrator_password
printf '%s\n' \
  "DATABASE_URL=postgres://tournaments_manager_prod_migrator:${encoded_migrator_password}@postgresql.prod.svc.cluster.local:5432/fasttourney_prod?sslmode=disable&search_path=public&options=-c%20role%3Dtournaments_manager_prod_schema_owner" \
  > "$migration_secret_file"
unset encoded_migrator_password

sudo /usr/local/bin/k3s kubectl create secret generic api-migrations-runtime \
  --namespace prod \
  --from-env-file="$migration_secret_file" \
  --dry-run=client -o yaml |
  sudo /usr/local/bin/k3s kubectl apply -f -

printf '%s\n' 'Ejecutando migraciones Goose (máximo 180 segundos)...'
sudo /usr/local/bin/k3s kubectl -n prod delete job api-migrations --ignore-not-found
cat <<EOF | sudo /usr/local/bin/k3s kubectl apply -f -
apiVersion: batch/v1
kind: Job
metadata:
  name: api-migrations
  namespace: prod
  labels:
    app.kubernetes.io/name: api-migrations
    app.kubernetes.io/part-of: tournaments-manager
spec:
  backoffLimit: 0
  ttlSecondsAfterFinished: 600
  template:
    metadata:
      labels:
        app.kubernetes.io/name: api-migrations
        app.kubernetes.io/part-of: tournaments-manager
    spec:
      automountServiceAccountToken: false
      restartPolicy: Never
      containers:
        - name: migrator
          image: $migrator_image
          imagePullPolicy: Never
          args: ["-dir", "/migrations", "up"]
          env:
            - name: GOOSE_DRIVER
              value: postgres
            - name: GOOSE_DBSTRING
              valueFrom:
                secretKeyRef:
                  name: api-migrations-runtime
                  key: DATABASE_URL
          securityContext:
            allowPrivilegeEscalation: false
            capabilities:
              drop:
                - ALL
            readOnlyRootFilesystem: true
            runAsGroup: 65532
            runAsNonRoot: true
            runAsUser: 65532
            seccompProfile:
              type: RuntimeDefault
          resources:
            requests:
              cpu: 100m
              memory: 64Mi
            limits:
              cpu: 500m
              memory: 128Mi
EOF
sudo /usr/local/bin/k3s kubectl -n prod wait --for=condition=complete job/api-migrations --timeout=180s
sudo /usr/local/bin/k3s kubectl -n prod logs job/api-migrations

printf '%s\n' 'Creando el Secret runtime mínimo de la API...'
umask 077
app_password="$(
  sudo /usr/local/bin/k3s kubectl -n prod get secret postgresql-runtime \
    -o jsonpath='{.data.POSTGRES_APP_PASSWORD}' | base64 --decode
)"
export app_password
encoded_password="$(python3 -c 'import os, urllib.parse; print(urllib.parse.quote(os.environ["app_password"], safe=""))')"
unset app_password

printf '%s\n' \
  "DATABASE_URL=postgres://tournaments_manager_prod_app:${encoded_password}@postgresql.prod.svc.cluster.local:5432/fasttourney_prod?sslmode=disable" \
  > "$secret_file"
unset encoded_password

sudo /usr/local/bin/k3s kubectl create secret generic api-runtime \
  --namespace prod \
  --from-env-file="$secret_file" \
  --dry-run=client -o yaml |
  sudo /usr/local/bin/k3s kubectl apply -f -

sudo /usr/local/bin/k3s kubectl create secret generic api-integrations \
  --namespace prod \
  --from-env-file="$integrations_file" \
  --dry-run=client -o yaml |
  sudo /usr/local/bin/k3s kubectl apply -f -

printf '%s\n' 'Validando manifests contra el API server...'
sudo /usr/local/bin/k3s kubectl apply --dry-run=server \
  -f "$work_dir/api-config.yaml" \
  -f "$work_dir/api.yaml"
sudo /usr/local/bin/k3s kubectl apply \
  -f "$work_dir/api-config.yaml" \
  -f "$work_dir/api.yaml"

printf '%s\n' 'Esperando el rollout de la API (máximo 120 segundos)...'
sudo /usr/local/bin/k3s kubectl -n prod rollout status deployment/api --timeout=120s
sudo /usr/local/bin/k3s kubectl -n prod get pods -l app.kubernetes.io/name=api
sudo /usr/local/bin/k3s kubectl -n prod get service api

rm -f "$work_dir/tournaments-manager-api.tar" \
  "$work_dir/tournaments-manager-migrator.tar"
REMOTE

# La VM recibe un script antes de abrir el TTY. De ese modo Bash ejecuta un
# fichero no interactivo y sudo puede leer su contraseña desde el terminal sin
# que el perfil interactivo de zsh/Warp altere el heredoc.
ssh -tt "$K3S_SSH_USER@$K3S_SSH_HOST" \
  'bash --noprofile --norc /tmp/tournaments-manager-deploy-api.sh; status=$?; rm -f /tmp/tournaments-manager-deploy-api.sh; exit "$status"'

# El renderer vive en el borde del Mac pero consume el contrato de la API que
# acaba de promocionarse. Se despliega después del rollout y con el mismo SHA.
"$repo_root/infra/home/deploy-league-preview-renderer.sh" prod "$release_sha"
