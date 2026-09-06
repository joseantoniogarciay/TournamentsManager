# Administración remota de la VM K3s

> Estado: aplicado y verificado el 2026-09-05.

## Límite

Este procedimiento crea la identidad administrativa SSH dedicada de ADR-0117. Permite gestionar por remoto el host, K3s y `kube-system`; no publica SSH, Cloudflare, Caddy ni los hostnames de producción.

## Fundamento

El Llavero de macOS protege la clave privada y `ssh-agent` la presenta al servidor. No proporciona ni almacena la contraseña de Ubuntu. La cuenta `fasttourney-operator` elimina esa contraseña de las operaciones posteriores mediante una regla `sudoers` explícita, manteniendo la clave dedicada como frontera de administración.

## Crear y cargar la clave en el Mac

Ejecutar una vez, fuera del repositorio:

```sh
ssh-keygen -t ed25519 -f ~/.ssh/fasttourney_k3s_operator -C fasttourney-k3s-operator
ssh-add --apple-use-keychain ~/.ssh/fasttourney_k3s_operator
```

Copiar el bloque de [`fasttourney-operator.ssh-config.template`](../../infra/k3s/host/fasttourney-operator.ssh-config.template) a `~/.ssh/config`. La clave privada y su passphrase no se muestran, copian ni guardan en el proyecto.

## Bootstrap único

`infra/k3s/.env` debe conservar temporalmente la cuenta Ubuntu actual en `K3S_SSH_USER`, porque es quien puede introducir su contraseña de `sudo` una vez. Desde la raíz:

```sh
bash infra/k3s/scripts/bootstrap-remote-operator.sh \
  ~/.ssh/fasttourney_k3s_operator.pub
```

El script transmite solo la clave pública, el bootstrap y la plantilla `sudoers` a un directorio temporal de la VM. Abre un TTY exclusivamente para el prompt de `sudo`, valida `sudoers` con `visudo -cf` y borra ese directorio al terminar. No modifica la configuración de `sshd`.

Después, cambiar localmente el valor ignorado por Git:

```sh
K3S_SSH_USER=fasttourney-operator
```

## Verificación

```sh
ssh -o BatchMode=yes fasttourney-k3s 'sudo -n true'
ssh -o BatchMode=yes fasttourney-k3s \
  'sudo /usr/local/bin/k3s kubectl get nodes && sudo /usr/local/bin/k3s kubectl -n kube-system get pods'
```

Ambos comandos deben terminar sin pedir contraseña. Confirmar también que root y autenticación por contraseña siguen rechazados desde SSH. No ejecutar consultas de Secret salvo cuando una recuperación concreta lo exija, y nunca imprimir su contenido.

**Evidencia 2026-09-05:** `fasttourney-operator` autenticó por su clave dedicada,
`sudo -n true` terminó correctamente y K3s informó el nodo
`fasttourney-k3s-lab` en estado `Ready`. La consulta de `kube-system` confirmó
CoreDNS, Traefik, ServiceLB, metrics-server y local-path-provisioner activos.

## Revocación

Ante pérdida de la clave o del Mac, desde una sesión administrativa segura:

```sh
sudo usermod --lock fasttourney-operator
sudo rm /etc/sudoers.d/fasttourney-operator
sudo rm -f /home/fasttourney-operator/.ssh/authorized_keys
```

La cuenta bloqueada conserva datos mínimos para auditoría; borrarla requiere confirmar primero que no posee procesos, ficheros o tareas necesarias. Crear una clave nueva y repetir el bootstrap restaura el acceso.
