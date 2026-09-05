#!/usr/bin/env bash

# Aplica exclusivamente las integraciones de producción de la API ya desplegada.
# Se ejecuta en el Mac; no construye ni importa imágenes y no imprime secretos.
set -euo pipefail

script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
repo_root=$(cd -- "$script_dir/../../.." && pwd)
work_dir=/tmp/tournaments-manager-k3s

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

ssh "$K3S_SSH_USER@$K3S_SSH_HOST" "mkdir -p '$work_dir'"
scp "$repo_root/infra/k3s/core/api-config.yaml" \
  "$repo_root/infra/k3s/core/api.yaml" \
  "$K3S_SSH_USER@$K3S_SSH_HOST:$work_dir/"

# La transferencia crea el fichero remoto con modo 0600 desde el inicio; scp
# podría aplicar el umask remoto y dejar una ventana con permisos más amplios.
ssh "$K3S_SSH_USER@$K3S_SSH_HOST" \
  'umask 077; cat > /tmp/tournaments-manager-api-integrations.env' < "$integrations_file"

ssh "$K3S_SSH_USER@$K3S_SSH_HOST" \
  'cat > /tmp/tournaments-manager-apply-api-integrations.sh && chmod 700 /tmp/tournaments-manager-apply-api-integrations.sh' <<'REMOTE'
set -euo pipefail

work_dir=/tmp/tournaments-manager-k3s
integrations_file=/tmp/tournaments-manager-api-integrations.env
trap 'rm -f "$integrations_file"' EXIT

sudo -v
sudo /usr/local/bin/k3s kubectl create secret generic api-integrations \
  --namespace prod \
  --from-env-file="$integrations_file" \
  --dry-run=client -o yaml |
  sudo /usr/local/bin/k3s kubectl apply -f -

sudo /usr/local/bin/k3s kubectl apply --dry-run=server \
  -f "$work_dir/api-config.yaml" \
  -f "$work_dir/api.yaml"
sudo /usr/local/bin/k3s kubectl apply \
  -f "$work_dir/api-config.yaml" \
  -f "$work_dir/api.yaml"
sudo /usr/local/bin/k3s kubectl -n prod rollout status deployment/api --timeout=120s
sudo /usr/local/bin/k3s kubectl -n prod get pods -l app.kubernetes.io/name=api
REMOTE

# El TTY queda reservado para la contraseña sudo de la VM. Ningún secreto pasa
# por ese canal ni queda persistido allí.
ssh -tt "$K3S_SSH_USER@$K3S_SSH_HOST" \
  'bash --noprofile --norc /tmp/tournaments-manager-apply-api-integrations.sh; status=$?; rm -f /tmp/tournaments-manager-apply-api-integrations.sh; exit "$status"'
