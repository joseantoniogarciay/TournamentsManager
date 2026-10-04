#!/bin/sh
# Local Compose uses its administrator; public environments keep their roles.
set -eu
cd /app
: "${DATABASE_URL:?DATABASE_URL is required}"
migrations=$(mktemp -d)
trap 'rm -rf "$migrations"' EXIT HUP INT TERM
for source in db/migrations/*.sql; do
    sed \
        -e '/^SET ROLE tournaments_manager_dev_schema_owner;$/d' \
        -e '/^GRANT .* TO tournaments_manager_dev_app;$/d' \
        "$source" > "$migrations/$(basename "$source")"
done
GOOSE_DRIVER=postgres GOOSE_DBSTRING="$DATABASE_URL" \
    go tool -modfile=go.tool.mod goose -dir "$migrations" up
