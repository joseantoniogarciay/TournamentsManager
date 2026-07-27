# Retrospectiva técnica — Fase 1

- **Fecha:** 2026-07-26
- **Objetivo:** preparar un entorno local reproducible que se parezca a
  producción sin introducir Kubernetes antes de necesitarlo.
- **Participantes:** Usuario y Codex.

## Resultado frente al objetivo

PostgreSQL 18.4 se opera mediante Docker Compose con un contrato local separado,
volumen nombrado, `healthcheck` y puerto limitado a loopback. Se validaron el
arranque, la salud, la parada, el reinicio con persistencia y el reset explícito
que recrea una base vacía. El runbook documenta preparación, inspección, logs,
migraciones y los fallos comunes.

No se añadieron API, cliente, Redis/Valkey, MinIO, observabilidad ni Kubernetes.
Tampoco existen aún migraciones: `make db-migrate` lo comunica como una omisión
intencional y ejecutará `goose` cuando haya archivos SQL versionados.

## Decisiones

- **Funcionaron:** Compose solo para PostgreSQL conserva la paridad útil
  (imagen, configuración externa, salud y volumen) sin degradar el bucle nativo
  de Go y Expo.
- **Coste inesperado:** `goose` trata la ausencia total de migraciones como
  error; el comando versionado distingue ahora ese estado previo al esquema de
  un fallo de migración real.
- **Revisar:** ampliar Compose solo si aparecen varios servicios coordinados o
  evidencia de que probar la imagen de API en cada cambio aporta valor.
- **ADR ausentes:** ninguno; ADR-0018 y ADR-0017 cubren el alcance.

## Aprendizaje

- El contenedor es efímero; el volumen nombrado es quien conserva los datos.
- Un `healthcheck` valida disponibilidad, no solo que un proceso haya arrancado.
- Un reset útil debe ser explícito, visible y limitado al volumen del proyecto.

## Calidad profesional

- **Seguridad:** no se expone PostgreSQL fuera de `127.0.0.1`; los contratos
  locales están ignorados y los ejemplos no contienen secretos reales.
- **Pruebas:** se ejecutaron `make verify`, arranque saludable, persistencia y
  recuperación destructiva confirmada. Las migraciones reales se comprobarán
  junto al primer esquema.
- **Observabilidad:** se cuenta con estado y logs de Compose; métricas, trazas y
  logs estructurados pertenecen a la Fase 3.
- **Operación y recuperación:** el runbook incluye parada no destructiva y
  reset confirmado; la base queda vacía y saludable tras la validación.
- **Coste:** no se crearon recursos cloud ni servicios locales adicionales.
- **Documentación:** se actualizaron runbook, desarrollo, changelog y diario de
  aprendizaje en el mismo cambio.

## Complejidad

Una sola dependencia y los comandos `make db-*` son suficientes. Añadir una
imagen de API, frontend contenido, seeds funcionales o un orquestador no aporta
evidencia adicional en esta fase.

## Acciones

| Acción | Propietario | Disparador | Destino |
| --- | --- | --- | --- |
| Diseñar esquema y primeras migraciones del vertical slice | Usuario/Codex | Inicio de Fase 2 | ADR y `apps/backend/db/migrations` |
| Verificar que migraciones reales se aplican e idempotencia | Usuario/Codex | Primera migración SQL | Runbook y pruebas de integración |
| Revisar Compose local | Usuario/Codex | Disparadores de ADR-0018 | [DECISIONS_TO_REVISIT.md](../governance/DECISIONS_TO_REVISIT.md) |

## Cierre

La Fase 1 queda cerrada. El siguiente paso es la Fase 2: diseñar e implementar
el primer vertical slice del backend sin adelantar infraestructura adicional.
