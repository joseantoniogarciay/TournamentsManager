#!/usr/bin/env bash

# Crea una cuenta administrativa SSH por clave en la VM. El operador actual
# introduce una única vez su contraseña sudo en un TTY; la nueva clave privada
# nunca abandona el Mac ni se escribe en el repositorio.
set -euo pipefail

repository_root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../../.." && pwd)
operator_environment=${K3S_OPERATOR_ENVIRONMENT:-"$repository_root/infra/k3s/.env"}
operator_user=${K3S_REMOTE_OPERATOR_USER:-fasttourney-operator}
operator_public_key=${1:-}

if [[ ! -r "$operator_environment" ]]; then
  echo "missing operator environment: $operator_environment" >&2
  exit 1
fi

if [[ ! -r "$operator_public_key" ]]; then
  echo "usage: $0 /absolute/path/to/fasttourney_k3s_operator.pub" >&2
  exit 2
fi

if [[ ! "$operator_user" =~ ^[a-z_][a-z0-9_-]{0,31}$ ]]; then
  echo "invalid operator user: $operator_user" >&2
  exit 2
fi

if ! ssh-keygen -lf "$operator_public_key" >/dev/null; then
  echo "invalid SSH public key: $operator_public_key" >&2
  exit 2
fi

# shellcheck disable=SC1090
source "$operator_environment"
: "${K3S_SSH_HOST:?K3S_SSH_HOST is required}"
: "${K3S_SSH_USER:?K3S_SSH_USER is required}"

remote_directory=/tmp/fasttourney-operator-bootstrap
script_source="$repository_root/infra/k3s/scripts/bootstrap-remote-operator-vm.sh"
sudoers_template="$repository_root/infra/k3s/host/fasttourney-operator.sudoers.template"

ssh "$K3S_SSH_USER@$K3S_SSH_HOST" "rm -rf '$remote_directory' && mkdir -m 700 '$remote_directory'"
scp "$script_source" "$operator_public_key" "$sudoers_template" "$K3S_SSH_USER@$K3S_SSH_HOST:$remote_directory/"

ssh -tt "$K3S_SSH_USER@$K3S_SSH_HOST" \
  "bash '$remote_directory/bootstrap-remote-operator-vm.sh' '$operator_user' '$remote_directory/$(basename "$operator_public_key")' '$remote_directory/$(basename "$sudoers_template")'; status=\$?; rm -rf '$remote_directory'; exit \$status"

echo "Bootstrap completed. Set K3S_SSH_USER=$operator_user in $operator_environment and run the validation from the remote-administration runbook."
