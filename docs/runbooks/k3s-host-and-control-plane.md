# Host y control plane K3s

> Estado: completado el 2026-08-24.

Este runbook cubre el primer módulo de la Fase 4: comprobar el host Ubuntu ARM64 y después verificar que K3s mantiene un único nodo sano. No crea el namespace `prod`, no despliega workloads, no cambia Caddy ni retira los `503` de producción.

## Fundamento

En Compose, Docker Desktop crea y supervisa los contenedores del proyecto. En K3s, el host Linux ejecuta el servicio `k3s`; este incluye control plane, runtime y kubelet para el nodo único. Antes de aplicar un `Deployment`, el host debe tener espacio en su raíz para imágenes y volúmenes y el control plane debe ser recuperable tras reiniciar la VM.

La alternativa de instalar directamente manifests o Helm ahora ocultaría este límite y adelantaría dependencias. La decisión aceptada en ADR-0112 exige manifiestos propios primero; Helm llegará solo con observabilidad de terceros.

## Prerrequisitos

- UTM ejecuta `fasttourney-k3s-lab` con Ubuntu Server 24.04 ARM64, 4 vCPU, 6 GB RAM y disco dinámico de 30 GB (ADR-0110).
- El operador usa la cuenta no root `joseantoniogarciay`. Su clave privada
  `~/.ssh/id_ed25519`, creada para esta VM, permanece fuera del repositorio y
  está disponible mediante `ssh-agent`; el host conserva su clave pública
  autorizada. La contraseña que `sudo` solicita es la de esa misma cuenta
  Ubuntu y no se guarda en el repositorio ni en un fichero local.
- La raíz de Ubuntu se ha extendido con el espacio libre del grupo LVM antes de instalar imágenes y volúmenes.

## Verificación previa

En el Mac, `ssh-agent` debe tener disponible `~/.ssh/id_ed25519`; si no lo
tiene, el operador ejecuta `ssh-add ~/.ssh/id_ed25519` y escribe la passphrase
localmente. No se pega la clave ni la passphrase en la conversación. El fichero
local `infra/k3s/.env`, ignorado por Git, conserva solo host y usuario.

Después, desde la raíz del repositorio:

```sh
set -a
. infra/k3s/.env
set +a
bash infra/k3s/scripts/verify-remote-host.sh
```

El wrapper transmite solo el verificador al host; no sube la clave ni escribe en
la VM. Como primera conexión, OpenSSH guarda la huella del host para detectar
cambios posteriores. Si ya se trabaja dentro de la VM, se puede ejecutar
directamente:

```sh
bash infra/k3s/scripts/verify-host.sh
```

Debe informar ARM64, Ubuntu 24.04, al menos 15 GiB libres en `/`, systemd y el
servicio K3s. El kubeconfig se restringe a `root`, lo cual es esperado. Sin una
terminal interactiva, el estado del nodo queda pendiente porque `sudo` requiere
la contraseña del operador.

## Instalación y verificación posterior

K3s ya está instalado y activo en la VM. Antes de continuar con namespaces o
workloads se documentará su versión efectiva y se completará la verificación
administrativa; cualquier actualización posterior dejará la versión visible en
Git y un rollback explícito.

**Evidencia no invasiva, 2026-08-24:** host `aarch64`, Ubuntu 24.04, 21 GiB
libres en `/`, servicio activo y `k3s v1.36.3+k3s1` (Go 1.26.5). El kubeconfig
existe y queda protegido para `root`, como corresponde.

**Prueba de recuperación, 2026-08-24:** se reinició la VM desde UTM. Al volver,
SSH respondió, el servicio K3s estaba activo y la comprobación administrativa
confirmó un único nodo `Ready` y la `StorageClass` predeterminada `local-path`.
Su provisionador es `rancher.io/local-path`, su política de recuperación es
`Delete`, espera al consumidor antes de crear el volumen y no admite expansión
en caliente. La prueba no creó ni eliminó PVC, Pods ni datos.

Para la comprobación administrativa completa, abrir una terminal SSH interactiva
en el Mac, cargar la misma configuración y ejecutar:

```sh
ssh -t "$K3S_SSH_USER@$K3S_SSH_HOST" \
  'sudo /usr/local/bin/k3s kubectl get nodes -o wide && sudo /usr/local/bin/k3s kubectl get storageclass'
```

El criterio de éxito es un nodo `Ready` y el almacenamiento local que K3s anuncia para el nodo. Guardar la salida operativa fuera de Git si contiene metadatos de la máquina.

## Fallo controlado y rollback

Para probar recuperación, reiniciar la VM de forma controlada desde UTM y repetir la verificación posterior. No se prueba aún borrando el servicio ni el directorio de datos de K3s: eso sería destructivo y no aporta una recuperación de datos de producción. Si K3s no vuelve a estar activo, conservar `journalctl -u k3s` y detener el avance; restaurar el snapshot de UTM tomado antes de la instalación solo si el diagnóstico confirma que el host no se puede reparar con seguridad.

## Cierre del módulo

El módulo quedó cerrado el 2026-08-24: la VM superó ambas verificaciones y el
reinicio controlado. El siguiente módulo puede crear el aislamiento con
namespace y RBAC mínimos.
