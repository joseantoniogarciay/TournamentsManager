#!/usr/bin/env bash

# Verifica los prerrequisitos y, opcionalmente, el control plane del único nodo
# K3s de producción doméstica. No instala paquetes ni modifica el host.
set -euo pipefail

require_k3s=false
if [[ ${1:-} == "--require-k3s" ]]; then
  require_k3s=true
elif [[ $# -ne 0 ]]; then
  echo "uso: $0 [--require-k3s]" >&2
  exit 2
fi

failures=0
pass() { printf 'PASS  %s\n' "$1"; }
fail() { printf 'FAIL  %s\n' "$1" >&2; failures=$((failures + 1)); }
info() { printf 'INFO  %s\n' "$1"; }

architecture=$(uname -m)
case "$architecture" in
  aarch64|arm64) pass "arquitectura ARM64 ($architecture)" ;;
  *) fail "se esperaba ARM64 y se recibió $architecture" ;;
esac

if [[ -r /etc/os-release ]]; then
  # shellcheck disable=SC1091
  . /etc/os-release
  if [[ ${ID:-} == "ubuntu" && ${VERSION_ID:-} == "24.04" ]]; then
    pass "Ubuntu ${VERSION_ID}"
  else
    fail "se esperaba Ubuntu 24.04; se detectó ${PRETTY_NAME:-desconocido}"
  fi
else
  fail "/etc/os-release no está disponible"
fi

root_available_kb=$(df -Pk / | awk 'NR == 2 { print $4 }')
root_available_gib=$((root_available_kb / 1024 / 1024))
if (( root_available_gib >= 15 )); then
  pass "${root_available_gib} GiB libres en /"
else
  fail "solo ${root_available_gib} GiB libres en /; ampliar LVM antes de instalar workloads"
fi

if command -v systemctl >/dev/null 2>&1; then
  pass "systemd disponible"
else
  fail "systemd no está disponible"
fi

if [[ -x /usr/local/bin/k3s ]]; then
  pass "binario k3s instalado"
  if systemctl is-active --quiet k3s; then pass "servicio k3s activo"; else fail "servicio k3s inactivo"; fi
  if [[ -e /etc/rancher/k3s/k3s.yaml ]]; then
    pass "kubeconfig de administración presente"
    if [[ ! -r /etc/rancher/k3s/k3s.yaml ]]; then info "kubeconfig restringido a root (esperado)"; fi
  else
    fail "kubeconfig de administración ausente"
  fi
  if sudo -n true 2>/dev/null; then
    node_count=$(sudo /usr/local/bin/k3s kubectl get nodes --no-headers 2>/dev/null | wc -l | tr -d ' ')
    if [[ $node_count == "1" ]]; then pass "un único nodo registrado"; else fail "se esperaba un nodo y se encontraron $node_count"; fi
  elif $require_k3s; then
    fail "la comprobación completa requiere sudo; ejecútala en una terminal SSH interactiva"
  else
    info "nodo no comprobado: sudo requiere la contraseña del operador"
  fi
elif $require_k3s; then
  fail "k3s no está instalado"
else
  info "k3s aún no está instalado; los prerrequisitos del host se han comprobado"
fi

if (( failures > 0 )); then exit 1; fi
