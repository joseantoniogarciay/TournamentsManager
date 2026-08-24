# ADR-0110: Usar UTM y Ubuntu Server para la VM local de K3s

- **Estado:** Aceptado
- **Fecha:** 2026-08-23
- **Decisor:** Usuario, mediante decisión explícita
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

ADR-0101 acepta una VM Linux ligera de un nodo con K3s, pero deja sin concretar
el virtualizador, la distribución y sus recursos. El Mac Apple Silicon debe
ejecutar una VM ARM64 reproducible que permita aprender la operación de un host
Linux y conserve una ruta comprensible hacia EKS.

## Contexto y restricciones

- El host es un Mac Apple Silicon con 16 GB de memoria; no había virtualizador
  instalado.
- K3s soporta nodos `arm64/aarch64`, Ubuntu y un servidor de un nodo funcional.
- EKS admite AMIs optimizadas de Ubuntu para grupos de nodos; EKS Auto Mode usa
  una variante administrada e inmutable de Bottlerocket, que se aprenderá en el
  laboratorio AWS y no se imita prematuramente en la VM.
- El disco debe crecer bajo demanda y ser ampliable sin reservar de inmediato su
  límite completo en el SSD del Mac.

## Criterios de decisión

1. virtualización nativa ARM64, sin emulación x86;
2. coste cero y mantenimiento proporcionado a un laboratorio de un nodo;
3. base Linux convencional con transferencia razonable hacia EKS;
4. recursos acotados que permitan mantener Compose y herramientas del host;
5. disco dinámico y recuperable.

## Alternativas

### Alternativa A — UTM + Ubuntu Server 24.04 LTS ARM64

- **Ventajas:** UTM es gratuito, soporta virtualización ARM64 en macOS; Ubuntu
  es una base de servidor conocida y EKS ofrece AMIs optimizadas de Ubuntu.
- **Inconvenientes:** no reproduce el sistema operativo inmutable de EKS Auto
  Mode ni sus integraciones AWS.
- **Coste de adopción:** bajo; descarga, ISO ARM64 y una VM local.
- **Coste de mantenimiento:** bajo; actualizaciones del host Linux y snapshots
  antes de cambios relevantes.
- **Riesgos:** consumir recursos del Mac si se sobredimensionan workloads o
  retenciones.

### Alternativa B — UTM + Debian 13 ARM64

- **Ventajas:** base mínima y estable con ciclo de soporte largo.
- **Inconvenientes:** no tiene la misma continuidad directa con las AMIs Ubuntu
  optimizadas de EKS.
- **Coste de adopción y mantenimiento:** bajo.
- **Riesgos:** añadir una diferencia de distribución sin aportar una capacidad
  necesaria al objetivo cloud.

### Alternativa C — Fedora CoreOS o Bottlerocket local

- **Ventajas:** aproxima el patrón inmutable de algunos nodos EKS.
- **Inconvenientes:** cambia el laboratorio de operación Linux convencional por
  un sistema especializado y no reproduce EKS Auto Mode, que sigue siendo AWS
  gestionado.
- **Coste de adopción y mantenimiento:** medio.
- **Riesgos:** sobreingeniería y confusión entre K3s local y EKS administrado.

## Comparación

La alternativa A conserva una base generalista sobre la que practicar systemd,
red, almacenamiento y K3s, sin emulación ni coste. Aporta más continuidad con
los grupos de nodos EKS que Debian, mientras que C adelanta una complejidad que
pertenece al laboratorio cloud.

## Recomendación

**Opinión/recomendación:** alternativa A. Es la solución mínima suficiente para
un host K3s real y una transición pedagógica hacia EKS.

## Decisión del usuario

**Aceptada el 2026-08-23:** usar UTM con Ubuntu Server 24.04 LTS ARM64 para la
VM local de K3s. La VM tendrá 4 vCPU, 6 GB de RAM y un disco dinámico de 30 GB.
El disco es un límite ampliable: no reserva los 30 GB en el SSD del host al
crearse.

**Acceso operativo aplicado el 2026-08-23:** la administración desde el Mac se
realiza por OpenSSH a través de la red privada compartida de UTM, autenticada
exclusivamente con una clave pública local. `PasswordAuthentication` y
`KbdInteractiveAuthentication` están desactivados, igual que el acceso de
`root`. No se configura una carpeta compartida ni se cambia la VM a red puente.

## Consecuencias

### Positivas

- El laboratorio usa virtualización ARM64 nativa y no requiere suscripciones.
- Quedan recursos suficientes para macOS, Docker Compose y herramientas de
  desarrollo.
- La elección mantiene una frontera explícita: K3s local enseña Kubernetes y
  el laboratorio EKS enseñará servicios e integraciones de AWS.
- El acceso por clave evita depender del portapapeles de la consola Server y
  mantiene la administración fuera de la LAN del usuario.

### Negativas y deuda aceptada

- La VM es de un nodo, sin alta disponibilidad y dependiente del Mac.
- Ubuntu no reproduce Bottlerocket ni el ciclo de vida de nodos de EKS Auto
  Mode.
- La VM, su disco y sus snapshots requieren operación y copias explícitas.

## Validación

1. UTM crea la VM ARM64 con un disco dinámico cuyo uso real inicial es inferior
   a 30 GB. Tras el particionado guiado con LVM, el volumen raíz se amplía al
   espacio libre del grupo antes de instalar workloads; así el sistema de
   archivos aprovecha la capacidad lógica de 30 GB.
2. Ubuntu arranca, recibe actualizaciones y expone conectividad de red.
3. K3s inicia como servicio y muestra un único nodo sano.
4. La VM puede apagarse y arrancarse sin compartir datos, secretos ni tráfico
   con Compose o AWS.
5. Una clave autorizada del Mac accede por SSH; un intento sin clave y con
   autenticación por contraseña es rechazado.

## Disparadores de revisión

- Se necesitan más de un nodo, alta disponibilidad o disponibilidad continua.
- 6 GB de RAM o 30 GB de disco impiden una carga aceptada.
- EKS adopta Auto Mode como destino principal y justifica un laboratorio
  específico de Bottlerocket.

## Documentación afectada

- [ROADMAP.md](../project/ROADMAP.md)
- [DEPLOYMENT.md](../operations/DEPLOYMENT.md)
- [DECISIONS.md](../governance/DECISIONS.md)
- [DECISIONS_TO_REVISIT.md](../governance/DECISIONS_TO_REVISIT.md)
- [LEARNING.md](../project/LEARNING.md)
- [CHANGELOG.md](../../CHANGELOG.md)
