#!/usr/bin/env bash

# Ejecuta verify-host.sh en la VM mediante SSH con la identidad de ssh-agent.
set -euo pipefail

: "${K3S_SSH_HOST:?define K3S_SSH_HOST}"
: "${K3S_SSH_USER:?define K3S_SSH_USER}"

if [[ ${1:-} == "--require-k3s" ]]; then
  verify_argument=--require-k3s
elif [[ $# -eq 0 ]]; then
  verify_argument=
else
  echo "uso: $0 [--require-k3s]" >&2
  exit 2
fi

script_directory=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
ssh \
  -o BatchMode=yes \
  -o StrictHostKeyChecking=accept-new \
  "$K3S_SSH_USER@$K3S_SSH_HOST" \
  "bash -s -- $verify_argument" < "$script_directory/verify-host.sh"
