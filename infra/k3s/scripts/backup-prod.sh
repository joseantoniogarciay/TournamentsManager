#!/bin/bash
# Creates one pgBackRest backup in K3s and publishes a complete encrypted
# repository replica on the Mac only after transfer and layout checks succeed.
set -euo pipefail

if [[ $# -ne 1 || ( "$1" != full && "$1" != incremental ) ]]; then
  echo "usage: $0 {full|incremental}" >&2
  exit 64
fi

script_directory=$(cd -- "$(dirname -- "$0")" && pwd)
repository_root=$(cd -- "$script_directory/../../.." && pwd)
# The versioned source runs with infra/k3s/.env. The LaunchAgent runs an
# installed copy outside macOS-protected folders and supplies its private copy
# through K3S_OPERATOR_ENVIRONMENT.
operator_environment=${K3S_OPERATOR_ENVIRONMENT:-"$repository_root/infra/k3s/.env"}

if [[ ! -r "$operator_environment" ]]; then
  echo "missing operator environment: $operator_environment" >&2
  exit 78
fi

# shellcheck disable=SC1090
source "$operator_environment"
: "${K3S_SSH_HOST:?K3S_SSH_HOST is required}"
: "${K3S_SSH_USER:?K3S_SSH_USER is required}"
: "${K3S_PGBACKREST_REPLICA_DESTINATION:?K3S_PGBACKREST_REPLICA_DESTINATION is required}"

destination=${K3S_PGBACKREST_REPLICA_DESTINATION%/}
parent_directory=$(dirname -- "$destination")
destination_name=$(basename -- "$destination")
stage_root=${K3S_PGBACKREST_STAGE_ROOT:-"$parent_directory"}

if [[ "$destination" != /* || "$destination_name" != prod || ! -d "$parent_directory" || ! -d "$stage_root" ]]; then
  echo "replica destination must be an existing absolute .../prod directory: $destination" >&2
  exit 64
fi

case "$1" in
  full) remote_backup=/usr/local/sbin/fasttourney-prod-pgbackrest-backup-full ;;
  incremental) remote_backup=/usr/local/sbin/fasttourney-prod-pgbackrest-backup-incremental ;;
esac

staging_directory=$(mktemp -d "$stage_root/.${destination_name}.staging.XXXXXX")
previous_directory="$parent_directory/.${destination_name}.previous"
cleanup() {
  rm -rf -- "$staging_directory"
}
trap cleanup EXIT

ssh "$K3S_SSH_USER@$K3S_SSH_HOST" "sudo $remote_backup"
mkdir -p "$staging_directory/repository"
ssh "$K3S_SSH_USER@$K3S_SSH_HOST" \
  'sudo /usr/local/sbin/fasttourney-prod-pgbackrest-export' | \
  tar -xf - -C "$staging_directory/repository"

test -f "$staging_directory/repository/backup/fasttourney-prod/backup.info"
test -f "$staging_directory/repository/archive/fasttourney-prod/archive.info"

if [[ -n ${K3S_PGBACKREST_PUBLISHER:-} ]]; then
  "$K3S_PGBACKREST_PUBLISHER" publish "$staging_directory/repository"
else
  # The interactive fallback preserves the original workflow. LaunchAgents set
  # K3S_PGBACKREST_PUBLISHER so a sandboxed helper owns iCloud access.
  rm -rf -- "$previous_directory"
  if [[ -e "$destination" ]]; then
    mv -- "$destination" "$previous_directory"
  fi
  mv -- "$staging_directory/repository" "$destination"
  rm -rf -- "$previous_directory"
fi

echo "published pgBackRest $1 replica at $destination"
