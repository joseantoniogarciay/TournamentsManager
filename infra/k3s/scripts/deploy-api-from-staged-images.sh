#!/usr/bin/env bash

# Importa en la VM K3s las imágenes API y migrator ya copiadas a /tmp, crea el
# Secret mínimo de la API desde la contraseña runtime de PostgreSQL y despliega
# la API. Se ejecuta en el Mac, no dentro de la VM.
set -euo pipefail

script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
repo_root=$(cd -- "$script_dir/../../.." && pwd)

# shellcheck disable=SC1091
source "$repo_root/infra/k3s/.env"

ssh "$K3S_SSH_USER@$K3S_SSH_HOST" \
  'cat > /tmp/tournaments-manager-deploy-api.sh && chmod 700 /tmp/tournaments-manager-deploy-api.sh' <<'REMOTE'
set -euo pipefail

work_dir=/tmp/tournaments-manager-k3s
secret_file=/tmp/api-runtime.env
trap 'rm -f "$secret_file"' EXIT

printf '%s\n' 'Validando sudo en la VM...'
sudo -v

printf '%s\n' 'Importando imagen runtime de la API...'
sudo /usr/local/bin/k3s ctr images import "$work_dir/tournaments-manager-api.tar"
printf '%s\n' 'Importando imagen migrator...'
sudo /usr/local/bin/k3s ctr images import "$work_dir/tournaments-manager-migrator.tar"
sudo /usr/local/bin/k3s ctr images ls | grep tournaments-manager

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
