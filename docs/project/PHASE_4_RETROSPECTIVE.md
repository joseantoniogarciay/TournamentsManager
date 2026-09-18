# Retrospectiva técnica — Fase 4

- **Fecha:** 2026-09-18
- **Objetivo:** aprender y operar Kubernetes en una VM Linux de un nodo sin
  adelantar la complejidad ni el coste de una plataforma cloud permanente.
- **Participantes:** Usuario y Codex.

## Resultado frente al objetivo

La fase cumple su criterio de salida. `prod` se ejecuta en K3s dentro de la VM
Ubuntu ARM64: namespace aislado, API sin estado con recursos y probes,
PostgreSQL con StatefulSet y PVC propios, configuración y secretos fuera de
Git, y observabilidad privada instalada con Helm. Cloudflare Tunnel y Caddy
publican `fasttourney.com` y `api.fasttourney.com` sin abrir puertos en la LAN
ni en la WAN; el rollback de API restaura `503` sin alterar K3s ni datos.

La recuperación se validó desde el repositorio persistente de K3s sin montar el
PVC activo y, el 2026-09-18, desde la réplica cifrada ya publicada en iCloud: el
contenedor efímero del Mac terminó con `fasttourney_prod|f`. La réplica
incremental programada también se ejecutó mediante `launchd` y finalizó con
código cero, publicando a través del helper sandboxed.

Quedan fuera alta disponibilidad, varios nodos, operador de base de datos,
despliegue blue/green, almacenamiento independiente del Mac/iCloud y EKS
permanente. Son límites deliberados, no requisitos incumplidos de la fase.

## Comparación con Docker Compose

| Aspecto | Compose (`local` / `dev`) | K3s (`prod`) | Aprendizaje |
| --- | --- | --- | --- |
| Ejecución | Contenedores iniciados por Compose | Pods reconciliados por Deployment o StatefulSet | Kubernetes declara el estado deseado y recupera procesos. |
| Red | Servicios de un proyecto Compose | Service e Ingress internos; Caddy como borde | Exponer una API no exige publicar el puerto ni la base de datos. |
| Estado | Volúmenes Docker del entorno | PVC de datos y repositorio separados | PostgreSQL necesita identidad y recuperación explícitas. |
| Configuración | Archivos locales ignorados | ConfigMap y Secrets por namespace | El contrato se conserva, pero el alcance y rotación cambian. |
| Operación | Logs y reinicios del proyecto | Rollouts, probes, eventos, métricas y alertas | La reconciliación aporta controles a cambio de más objetos. |
| Recuperación | Backup y restauración local | pgBackRest, réplica cifrada y restauración aislada | Un backup solo cuenta cuando se restaura desde otra copia. |

## Decisiones y aprendizaje

- **Funcionó:** manifiestos propios para el core y Helm solo para observabilidad
  de terceros (ADR-0112) hicieron visible el modelo de Kubernetes sin recrear
  charts ajenos.
- **Coste inesperado:** Docker Desktop no monta de forma fiable una ruta de
  iCloud Drive. El helper de ADR-0115 prepara staging local autorizado para la
  restauración, sin Full Disk Access para Bash.
- **Regla reutilizable:** un nodo `Ready` no demuestra una plataforma; hay que
  recorrer configuración, estado, red, observabilidad y recuperación. Un
  `StatefulSet` no es una API con un disco adjunto.
- **ADR ausentes:** ninguno. ADR-0101 y ADR-0110 a ADR-0118 cubren el runtime,
  operación, almacenamiento, entrada y recuperación actuales.

## Calidad profesional

- **Seguridad:** secretos fuera de Git, SSH por clave dedicada y Caddy
  autenticado ante K3s; PostgreSQL no se publica.
- **Operación:** manifests con dry-run, rollouts, probes, fallos controlados,
  logs, métricas, trazas, alertas y runbooks verificables.
- **Recuperación:** réplica publicada, cifrada y distinta del PVC activo;
  recursos temporales eliminados al terminar.
- **Coste:** la VM doméstica evita cloud permanente, a cambio de depender de un
  nodo, Mac, iCloud y una cuenta de usuario.

## Acciones

| Acción | Propietario | Disparador | Destino |
| --- | --- | --- | --- |
| Reabrir cloud solo con nuevo análisis y autorización | Usuario/Codex | Necesidad explícita de cloud | [ADR-0128](../adr/0128-close-roadmap-after-k3s-and-cancel-aws-phase.md) |
| Evaluar blue/green y migraciones compatibles | Usuario/Codex | Requisito de disponibilidad o cambio incompatible | [DEPLOYMENT.md](../operations/DEPLOYMENT.md) |
| Revisar independencia del backup | Usuario/Codex | RPO/RTO insuficiente o pérdida de Mac/cuenta | [DECISIONS_TO_REVISIT.md](../governance/DECISIONS_TO_REVISIT.md) |

## Cierre

La Fase 4 queda cerrada. K3s opera el runtime doméstico de `prod` con
publicación, observabilidad y recuperación demostrada. La Fase 5 no se inicia
automáticamente: cualquier `apply` de AWS requiere análisis de coste y
autorización explícita del usuario.
