#!/bin/bash
# Verifies a restore from the replica received on the Mac, never from the PVC.
set -euo pipefail

script_directory=$(cd -- "$(dirname -- "$0")" && pwd)
repository_root=$(cd -- "$script_directory/../../.." && pwd)
operator_environment="$repository_root/infra/k3s/.env"

if [[ ! -r "$operator_environment" ]]; then
  echo "missing operator environment: $operator_environment" >&2
  exit 78
fi

# shellcheck disable=SC1090
source "$operator_environment"
: "${K3S_PGBACKREST_REPLICA_DESTINATION:?K3S_PGBACKREST_REPLICA_DESTINATION is required}"
: "${PGBACKREST_REPO1_CIPHER_PASS:?set PGBACKREST_REPO1_CIPHER_PASS interactively}"

if [[ ! -f "$K3S_PGBACKREST_REPLICA_DESTINATION/backup/fasttourney-prod/backup.info" ]]; then
  echo "missing pgBackRest backup metadata in replica" >&2
  exit 66
fi

if ! printf '%s\n' "$PGBACKREST_REPO1_CIPHER_PASS" | \
  ssh "$K3S_SSH_USER@$K3S_SSH_HOST" \
  'sudo /usr/local/sbin/fasttourney-prod-pgbackrest-verify-cipher'; then
  echo "the supplied cipher pass does not match the running PostgreSQL Secret" >&2
  exit 65
fi

restore_volume="tournaments-manager-prod-postgresql-restore-verify-$$"
docker volume create "$restore_volume" >/dev/null
cleanup() {
  docker volume rm -f "$restore_volume" >/dev/null 2>&1 || true
}
trap cleanup EXIT

docker run --rm \
  --name tournaments-manager-prod-postgresql-restore-verify \
  --user root \
  --env PGBACKREST_REPO1_CIPHER_PASS \
  --env POSTGRES_USER=postgres \
  --env POSTGRES_DB=fasttourney_prod \
  --volume "$K3S_PGBACKREST_REPLICA_DESTINATION:/var/lib/pgbackrest:ro" \
  --mount "type=volume,src=$restore_volume,dst=/restore" \
  tournaments-manager-postgresql:18.4-pgbackrest-2.59.1 \
  sh -ec '
    chown 999:999 /restore
    gosu postgres pgbackrest --stanza=fasttourney-prod --pg1-path=/restore restore
    gosu postgres postgres -D /restore -c archive_mode=off -c listen_addresses="" -c unix_socket_directories=/tmp -c port=55432 &
    postgres_pid=$!
    trap "kill $postgres_pid 2>/dev/null || true; wait $postgres_pid 2>/dev/null || true" EXIT
    attempt=0
    until gosu postgres pg_isready -h /tmp -p 55432 -U "$POSTGRES_USER" -d "$POSTGRES_DB"; do
      attempt=$((attempt + 1))
      test "$attempt" -lt 60 || exit 1
      sleep 1
    done
    gosu postgres psql -h /tmp -p 55432 -U "$POSTGRES_USER" -d "$POSTGRES_DB" -Atc "select current_database(), pg_is_in_recovery();"
  '
