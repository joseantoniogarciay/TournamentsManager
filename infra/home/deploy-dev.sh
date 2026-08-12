#!/bin/sh
set -eu

repository_root=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
release_root=/opt/homebrew/var/www/fasttourney/dev/releases
current_link=/opt/homebrew/var/www/fasttourney/dev/current
api_repository=tournaments-manager-dev-api

cd "$repository_root"

branch=$(git branch --show-current)
if [ "$branch" != "develop" ]; then
  echo "El despliegue dev solo se ejecuta desde develop; rama actual: $branch." >&2
  exit 1
fi

if ! git diff --quiet || ! git diff --cached --quiet; then
  echo "El árbol Git debe estar limpio antes de desplegar dev." >&2
  exit 1
fi

release_sha=$(git rev-parse HEAD)
remote_sha=$(git rev-parse origin/develop)
if [ "$release_sha" != "$remote_sha" ]; then
  echo "HEAD debe coincidir con origin/develop antes de desplegar." >&2
  exit 1
fi

api_image="$api_repository:git-$release_sha"
if docker image inspect "$api_image" >/dev/null 2>&1; then
  echo "La imagen $api_image ya existe; el SHA no se reutiliza." >&2
  exit 1
fi

docker build --target runtime --tag "$api_image" apps/backend
./infra/home/deploy-dev-web.sh "$release_sha"

DEV_API_IMAGE="$api_image" \
  docker compose --env-file infra/dev/.env -f infra/dev/compose.yaml up --detach --wait --remove-orphans

next_link="$current_link.next"
ln -s "releases/$release_sha" "$next_link"
mv -f "$next_link" "$current_link"

deployed_at=$(date -u +%Y-%m-%dT%H:%M:%SZ)
printf '{\n  "commit": "%s",\n  "apiImage": "%s",\n  "deployedAt": "%s"\n}\n' \
  "$release_sha" "$api_image" "$deployed_at" >"$release_root/$release_sha/deployment.json"

for old_release in $(ls -1dt "$release_root"/* 2>/dev/null | sed -n '3,$p'); do
  old_sha=$(basename "$old_release")
  rm -r -- "$old_release"
  docker image rm "$api_repository:git-$old_sha" >/dev/null 2>&1 || true
done

printf 'Dev desplegado: %s\n' "$release_sha"
