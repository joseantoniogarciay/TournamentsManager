# ADR-0117: Usar una identidad SSH dedicada para administrar la VM K3s

- **Estado:** Aceptado
- **Fecha:** 2026-09-05
- **Decisor:** Usuario, mediante decisión explícita
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

La administración remota de la VM K3s se detiene cuando `sudo` solicita una
contraseña interactiva. El proyecto necesita operar K3s, sus componentes en
`kube-system` y el host desde el Mac sin guardar la contraseña de Ubuntu ni
requerir intervención por cada operación.

## Contexto y restricciones

- La VM es privada en la red compartida de UTM; SSH no se publica en LAN/WAN.
- SSH ya usa claves; root, contraseña e interacción continúan deshabilitados en
  `sshd` conforme a ADR-0110.
- Administrar completamente K3s incluye `kube-system`, Traefik, ServiceLB,
  CoreDNS, upgrades, nodos y secretos del clúster.
- El repositorio es público: claves privadas, contraseñas y kubeconfigs no
  entran en Git ni se pegan en conversaciones o logs.

## Alternativas

### A — Cuenta SSH dedicada con `sudo` no interactivo

Crear `fasttourney-operator`, permitirle solo autenticación por clave desde la
red privada y otorgarle `sudo` sin contraseña para la administración de la VM.

- **Ventajas:** una frontera de identidad separada, administración completa y
  simple, y ningún secreto de contraseña en el Mac o en automatizaciones.
- **Inconvenientes:** quien controle la clave privada puede controlar la VM y
  el clúster completo.
- **Coste de adopción:** bajo; un bootstrap interactivo único.
- **Coste de mantenimiento:** bajo; revocar la clave o desactivar la cuenta
  corta el acceso.
- **Riesgos:** es una credencial de administrador; se mitiga con clave dedicada,
  Llavero/`ssh-agent`, permisos locales y red privada.

### B — Convertir la cuenta SSH personal en administradora no interactiva

- **Ventajas:** no crea otra cuenta ni otra clave.
- **Inconvenientes:** mezcla identidad humana, backups y automatización; una
  revocación rompe todas esas funciones a la vez.
- **Coste de mantenimiento:** bajo, con peor aislamiento.

### C — Wrappers `sudoers` limitados por comando

- **Ventajas:** menor privilegio por operación.
- **Inconvenientes:** no cumple la administración completa de K3s y host; una
  lista que cubra upgrades, diagnóstico y recuperación terminaría siendo frágil
  o una vía indirecta de ejecución arbitraria.
- **Coste de mantenimiento:** medio o alto.

### No cambiar

Seguir pidiendo la contraseña `sudo` en cada sesión.

- **Consecuencias:** conserva una barrera interactiva, pero bloquea la operación
  remota autónoma aceptada para la VM doméstica.

## Recomendación

**Recomendación:** A. La VM es de un único operador, privada y de bajo tráfico;
una identidad separada hace explícito el privilegio sin fingir que un conjunto
de wrappers puede administrar todo el sistema.

## Decisión del usuario

**Aceptada el 2026-09-05:** usar `fasttourney-operator`, una cuenta no root
autenticada exclusivamente por una clave SSH dedicada. Su clave privada vive en
el Llavero/`ssh-agent` del Mac, fuera del repositorio. La cuenta recibe `sudo`
sin contraseña para administrar la VM y K3s, incluido `kube-system`.

El bootstrap requiere una sola sesión con la contraseña de la cuenta Ubuntu
actual. Después, `infra/k3s/.env` usará esta identidad. No se configura
`NOPASSWD` para la cuenta SSH personal ni se guarda la contraseña de Ubuntu.

## Consecuencias

- La automatización puede diagnosticar y operar host y clúster remotamente.
- La clave dedicada es una credencial administrativa: su pérdida exige revocarla
  de `authorized_keys` o bloquear la cuenta antes de crear otra.
- Los secretos vistos durante una operación siguen sin imprimirse ni guardarse
  en el repositorio.

## Validación

1. La clave de operador se carga en `ssh-agent` y en el Llavero de macOS.
2. La cuenta entra por SSH sin contraseña y `sudo -n true` termina con éxito.
3. `sudo /usr/local/bin/k3s kubectl get nodes` y una consulta de `kube-system`
   funcionan, mientras root y contraseña continúan rechazados por SSH.
4. Al bloquear la cuenta o retirar su clave, SSH vuelve a rechazarse.

## Disparadores de revisión

- más de un operador humano o automatizado;
- exposición de SSH fuera de la red privada;
- pérdida o sospecha de compromiso de la clave;
- una necesidad de separar privilegios de host y clúster.
