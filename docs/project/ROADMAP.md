# Roadmap de aprendizaje e implementación

El roadmap es secuencial por aprendizaje, no necesariamente por entregas. No se
avanza de fase porque exista código: se avanza cuando se cumplen sus criterios de
salida y se completa la retrospectiva.

## Fase 0 — Documentación — completada

**Objetivo:** establecer propósito, proceso y baseline técnica antes del código.

**Entregables:**

- manifiesto preservado y transcrito;
- handbook navegable;
- proceso ADR y plantillas;
- arquitectura acordada registrada;
- mapa de decisiones pendientes;
- plantillas de aprendizaje, playbook, runbook y retrospectiva.

**Salida:** completada. Los enlaces internos se validan, las decisiones aceptadas
se distinguen de propuestas y el primer vertical slice queda definido. Véase la
[retrospectiva de Fase 0](PHASE_0_RETROSPECTIVE.md).

### Gate 0A — Confirmación de la base técnica — completado

Antes de Fase 1 se decide:

- topología del repositorio;
- topología del backend;
- estrategia web/mobile;
- contrato API e identidad;
- persistencia y toolchains;
- configuración, entorno local y pruebas;
- observabilidad, CI y despliegue.

El estado y orden están en
[TECHNICAL_BASELINE.md](../governance/TECHNICAL_BASELINE.md).

### Gate 0B — Definición del primer vertical slice — completado

Después de confirmar la base técnica se decide:

- formato inicial o inscripción/publicación sin competición;
- datos mínimos y ciclo de vida de la liga de fútbol inicial;
- usuario, equipo o ambos como participante;
- visibilidad del torneo;
- mecanismo de incorporación.

El formato y el ciclo de vida inicial están aceptados en
[ADR-0032](../adr/0032-define-minimum-football-league-data-and-lifecycle.md) y
revisados en
[ADR-0040](../adr/0040-make-published-leagues-editable-until-start.md).
La lectura pública por ID está aceptada en
[ADR-0049](../adr/0049-use-public-league-ids-for-read-only-access.md).
Los equipos como participantes, el seguimiento autenticado y la administración
delegada directa están aceptados en
[ADR-0034](../adr/0034-use-teams-as-competitors-and-direct-delegated-administration.md).
La gestión de resultados, bajas y cancelación se completa en ADR-0035 a
ADR-0042. El detalle vigente está en [PRODUCT.md](PRODUCT.md).

## Fase 1 — Entorno local — completada

**Objetivo:** crear un entorno reproducible que se parezca a producción sin
introducir Kubernetes antes de necesitarlo.

**Decisiones previas:** versiones soportadas, Docker Compose, gestión de
configuración y secretos, estrategia de migraciones, datos de desarrollo y
comandos de trabajo.

**Salida:** completada. Una persona puede preparar, arrancar, comprobar, detener
y limpiar PostgreSQL local siguiendo el runbook; arranque saludable, persistencia
y recuperación tras reset quedaron demostrados. Véase la
[retrospectiva de Fase 1](PHASE_1_RETROSPECTIVE.md).

## Fase 2 — Backend — completada

**Objetivo:** entregar un primer vertical slice de negocio en Go.

**Decisiones previas:** alcance, API, modelo de dominio, persistencia, errores,
validación y estrategia de pruebas.

**Primer incremento aceptado:** borrador local, identidad local verificada,
login federado con Google, publicación de liga con equipos y consulta por
organizador o ID público; resultados, Apple y ciclo avanzado quedan para un
incremento posterior. Véanse [ADR-0043](../adr/0043-deliver-publish-and-read-league-first-backend-increment.md)
y [ADR-0050](../adr/0050-include-google-federated-login-in-first-increment.md).

**Salida:** completada. El primer vertical slice recorre identidad verificada,
sesión, publicación y lectura de liga desde una base vacía; la autorización, la
dirección de dependencias, el contrato OpenAPI y las transacciones PostgreSQL
cuentan con pruebas proporcionales al riesgo. El backend continuó después con el
ciclo deportivo, administración, recuperación y notificaciones sin convertir
esas ampliaciones en requisitos retroactivos del slice inicial. Véase la
[retrospectiva de Fase 2](PHASE_2_RETROSPECTIVE.md).

## Fase 3 — Observabilidad — completada

**Objetivo:** poder explicar el estado del sistema y diagnosticar fallos.

**Decisiones previas:** señales y objetivos de servicio, instrumentación,
retención, stack mínimo y coste.

**Salida:** completada. El refresh de sesión dispone de logs, métricas y trazas
correlacionables, dashboard, SLO y runbook. Una caída controlada de PostgreSQL
activó y resolvió la alerta crítica real en local; la entrega externa de `dev`
se verificó mediante Alertmanager, Resend, Cloudflare y el buzón final. Véase la
[retrospectiva de Fase 3](PHASE_3_RETROSPECTIVE.md).

## Fase 4 — Kubernetes

**Objetivo:** aprender orquestación cuando el servicio ya sea operable.

**Secuencia vigente:** la Fase 4 comienza con una VM Linux ligera de un nodo y
K3s, que será el runtime doméstico de `prod` conforme a ADR-0111. Compose
conserva desarrollo y el runtime público `dev`. La validación distribuible de
PostHog para iOS y Android se difiere hasta disponer de sus cuentas de
distribución, conforme a ADR-0109; sigue siendo un gate antes de distribuir la
beta móvil, no de Kubernetes.

**Puerta de entrada:** las fases de backend y observabilidad están cerradas y
la selección VM Linux + K3s está aceptada. No se implementa una capacidad de
Fase 4 sin su diseño concreto, plan de verificación y rollback.

**Decisiones previas:** UTM, Ubuntu Server 24.04 LTS ARM64 y recursos de la VM
están aceptados en ADR-0110; el destino doméstico de `prod`, en ADR-0111.
Quedan aprovisionamiento reproducible de la VM,
empaquetado y manifests, health checks, recursos de los workloads,
configuración, secretos y estrategia de despliegue.

**Salida:** despliegue local reproducible, recuperación demostrada y comparación
documentada frente a Docker Compose.

### Itinerario de aprendizaje y entrega de K3s

Este itinerario permanece vigente entre sesiones. Cada módulo se cierra con:
problema operativo, concepto Kubernetes, comparación con Compose, manifiestos
explicados, verificación, fallo controlado cuando proceda y retrospectiva breve.
No autoriza a publicar `prod` hasta completar los gates de ADR-0111.

1. **Host y control plane:** verificar VM, servicio K3s, `kubectl`, nodo y
   almacenamiento local; distinguir host, runtime de contenedores y control
   plane de los procesos Docker Compose.
2. **Aislamiento:** crear el namespace `prod` y el mínimo RBAC necesario;
   separar sus recursos de `local` y `dev`.
3. **Configuración:** aplicar `ConfigMap` y `Secret`; comparar objetos de
   configuración con los archivos `.env` y secretos montados por Compose.
4. **API sin estado:** desplegar la imagen runtime mediante `Deployment`,
   `Service`, requests/limits y probes; observar reconciliación, rollout y
   recuperación de pods frente al arranque de contenedores de Compose.
5. **PostgreSQL con estado:** usar `StatefulSet` y `PersistentVolumeClaim`;
   estudiar identidad, orden y persistencia, y por qué difiere de la API.
6. **Recuperación de datos:** aplicar a `prod` el patrón pgBackRest de ADR-0108
   con volumen, repositorio y clave propios; demostrar una restauración aislada.
7. **Entrada pública:** enrutar Cloudflare Tunnel → Caddy → ingress K3s;
   diferenciar exposición interna por `Service` del enrutamiento HTTP por
   `Ingress` y conservar la ausencia de puertos LAN/WAN.
8. **Observabilidad y operación:** recorrer logs, eventos, métricas, alertas,
   rollouts y rollbacks; provocar fallos controlados de API y dependencia.
9. **Gate de publicación:** comprobar persistencia, restauración, secretos,
   recursos/probes, ingress, rollback y alertas antes de retirar el `503` de
   los hosts de producción.

## Fase 5 — Cloud

**Objetivo:** desplegar y operar en AWS mediante Terraform sin acoplar el dominio
al proveedor.

**Decisiones previas:** cuenta y seguridad, red, cómputo, datos, almacenamiento,
coste, backup, CI/CD y estrategia de rollback.

**Salida:** infraestructura reproducible, despliegue verificable, observabilidad,
presupuesto y procedimiento de recuperación documentados.

## Retrospectiva obligatoria

Cada fase termina usando
[phase-retrospective.md](../playbooks/phase-retrospective.md). Sus conclusiones
actualizan `LEARNING.md`, el handbook y, si procede, los ADR.
