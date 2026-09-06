#!/bin/bash
# Installs the production backup LaunchAgents outside macOS-protected folders.
# Versioned templates and source stay in Git; the installed runtime and its
# private operator environment remain local to the Mac user.
set -euo pipefail

script_directory=$(cd -- "$(dirname -- "$0")" && pwd)
repository_root=$(cd -- "$script_directory/../../.." && pwd)
operator_environment="$repository_root/infra/k3s/.env"
runtime_root="${FASTTOURNEY_K3S_RUNTIME_ROOT:-$HOME/Library/Application Support/FastTourney/k3s}"
agent_directory="$HOME/Library/LaunchAgents"
log_directory="${FASTTOURNEY_LOG_DIRECTORY:-$HOME/Library/Logs/FastTourney}"
user_domain="gui/$(id -u)"

if [[ ! -r "$operator_environment" ]]; then
  echo "missing operator environment: $operator_environment" >&2
  exit 78
fi

for key in K3S_SSH_HOST K3S_SSH_USER K3S_PGBACKREST_REPLICA_DESTINATION; do
  if ! grep -q "^${key}=." "$operator_environment"; then
    echo "missing $key in $operator_environment" >&2
    exit 78
  fi
done

mkdir -p "$runtime_root" "$agent_directory" "$log_directory"
chmod 700 "$runtime_root" "$log_directory"
mkdir -p "$runtime_root/staging"
chmod 700 "$runtime_root/staging"

publisher_source="$repository_root/infra/k3s/macos-backup-publisher/BackupPublisher.swift"
publisher_entitlements="$repository_root/infra/k3s/macos-backup-publisher/BackupPublisher.entitlements"
publisher_bundle="$runtime_root/BackupPublisher.app"
publisher_temporary="$publisher_bundle/Contents/MacOS/.BackupPublisher.unsigned"
publisher="$publisher_bundle/Contents/MacOS/BackupPublisher"
mkdir -p "$publisher_bundle/Contents/MacOS"
install -m 644 "$repository_root/infra/k3s/macos-backup-publisher/Info.plist" "$publisher_bundle/Contents/Info.plist"
swiftc -framework AppKit "$publisher_source" -o "$publisher_temporary"
codesign --force --sign - --identifier com.fasttourney.backup-publisher \
  --entitlements "$publisher_entitlements" "$publisher_temporary"
install -m 700 "$publisher_temporary" "$publisher"
rm -f "$publisher_temporary"
codesign --force --sign - "$publisher_bundle"

install -m 700 "$repository_root/infra/k3s/scripts/backup-prod.sh" "$runtime_root/backup-prod.sh"
{
  grep -v -E '^K3S_PGBACKREST_(STAGE_ROOT|PUBLISHER)=' "$operator_environment"
  printf 'K3S_PGBACKREST_STAGE_ROOT=%q\n' "$runtime_root/staging"
  printf 'K3S_PGBACKREST_PUBLISHER=%q\n' "$publisher"
} > "$runtime_root/k3s.env"
chmod 600 "$runtime_root/k3s.env"

for schedule in full incremental; do
  label="com.fasttourney.prod-postgresql-backup-$schedule"
  template="$repository_root/infra/home/launchd/$label.plist.template"
  plist="$agent_directory/$label.plist"
  temporary_plist=$(mktemp "$agent_directory/.${label}.XXXXXX")

  sed "s|__K3S_RUNTIME_ROOT__|$runtime_root|g; s|__LOG_DIRECTORY__|$log_directory|g" \
    "$template" > "$temporary_plist"
  plutil -lint "$temporary_plist" >/dev/null

  launchctl bootout "$user_domain/$label" 2>/dev/null || true
  install -m 600 "$temporary_plist" "$plist"
  rm -f "$temporary_plist"
  launchctl bootstrap "$user_domain" "$plist"
  echo "installed and loaded $label"
done
