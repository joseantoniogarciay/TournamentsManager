#!/usr/bin/env bash

set -euo pipefail

operator_user=${1:?operator user is required}
public_key_file=${2:?public key file is required}
sudoers_template=${3:?sudoers template is required}

if [[ ! "$operator_user" =~ ^[a-z_][a-z0-9_-]{0,31}$ ]]; then
  echo "invalid operator user" >&2
  exit 2
fi

if [[ ! -r "$public_key_file" || ! -r "$sudoers_template" ]]; then
  echo "missing bootstrap input" >&2
  exit 2
fi

sudo -v

if ! id "$operator_user" >/dev/null 2>&1; then
  sudo useradd --create-home --shell /bin/bash "$operator_user"
fi

operator_home=$(getent passwd "$operator_user" | cut -d: -f6)
sudo install -d -m 700 -o "$operator_user" -g "$operator_user" "$operator_home/.ssh"
sudo install -m 600 -o "$operator_user" -g "$operator_user" "$public_key_file" "$operator_home/.ssh/authorized_keys"

sudoers_file=/etc/sudoers.d/fasttourney-operator
sed "s/__K3S_OPERATOR__/$operator_user/g" "$sudoers_template" | sudo tee "$sudoers_file" >/dev/null
sudo chown root:root "$sudoers_file"
sudo chmod 440 "$sudoers_file"
sudo visudo -cf "$sudoers_file"

sudo -u "$operator_user" -H ssh-keygen -lf "$operator_home/.ssh/authorized_keys" >/dev/null
echo "Created or updated $operator_user. SSH root and password settings were not changed."
