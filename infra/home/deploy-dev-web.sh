#!/bin/sh
set -eu

repository_root=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
deploy_directory=/opt/homebrew/var/www/fasttourney/dev
staging_directory=$(mktemp -d "${TMPDIR:-/tmp}/fasttourney-dev-web.XXXXXX")
trap 'rm -rf "$staging_directory"' EXIT

cd "$repository_root"
EXPO_PUBLIC_API_BASE_URL=https://dev-api.fasttourney.com/v1 \
	pnpm --filter @tournaments-manager/client exec expo export --platform web --output-dir "$staging_directory"

install -d -m 755 "$deploy_directory"
rsync -a --delete "$staging_directory/" "$deploy_directory/"
