#!/bin/sh
# Pause only known dev jobs; preserve their installed plist and all data.
set -eu
case "${1:-}" in up|down) action=$1 ;; *) echo 'Usage: dev-launch-agents.sh up|down' >&2; exit 1 ;; esac
[ "$(uname -s)" = Darwin ] || exit 0
domain="gui/$(id -u)"
agent_directory=${FASTTOURNEY_DEV_LAUNCH_AGENTS_DIR:-$HOME/Library/LaunchAgents}
for name in account-purge purge-expired-accounts legal-audit-backup postgresql-backup-full postgresql-backup-incremental league-preview-renderer; do
    label="com.fasttourney.dev-$name"
    plist="$agent_directory/$label.plist"
    if [ "$action" = down ]; then
        if [ -f "$plist" ] || launchctl print "$domain/$label" >/dev/null 2>&1; then
            launchctl disable "$domain/$label"
            if launchctl print "$domain/$label" >/dev/null 2>&1; then
                launchctl bootout "$domain/$label"
            fi
        fi
    elif [ -f "$plist" ]; then
        launchctl enable "$domain/$label"
        if ! launchctl print "$domain/$label" >/dev/null 2>&1; then
            launchctl bootstrap "$domain" "$plist"
        fi
    fi
done
