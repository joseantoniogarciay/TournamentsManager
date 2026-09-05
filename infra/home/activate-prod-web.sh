#!/bin/sh
set -eu

# Conmuta únicamente la web estática ya preparada. No despliega API ni altera
# K3s, PostgreSQL, Secrets o Caddy.
release_sha=${1:?"Uso: activate-prod-web.sh <SHA-completo>"}
release_root=/opt/homebrew/var/www/fasttourney/prod/releases
release_directory="$release_root/$release_sha"
current_link=/opt/homebrew/var/www/fasttourney/prod/current

case "$release_sha" in
  *[!0123456789abcdef]* | "")
    echo "El SHA debe tener 40 caracteres hexadecimales." >&2
    exit 1
    ;;
esac
if [ "${#release_sha}" -ne 40 ] || [ ! -f "$release_directory/deployment.json" ]; then
  echo "No existe un release web de prod preparado para $release_sha." >&2
  exit 1
fi

next_link="$current_link.next"
ln -s "releases/$release_sha" "$next_link"
mv -f -h "$next_link" "$current_link"
printf 'Web de prod activada: %s\n' "$release_sha"
