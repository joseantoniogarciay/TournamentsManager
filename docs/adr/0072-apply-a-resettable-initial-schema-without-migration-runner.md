# ADR-0072: Aplicar el esquema inicial reescribible sin ejecutor de migraciones

- **Estado:** Aceptado
- **Fecha:** 2026-08-02
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** ADR-0053 y ADR-0071, solo en el mecanismo de aplicar el esquema durante la primera versión
- **Superado por:** Ninguno

## Problema

Durante la primera versión los datos locales y de CI son desechables. Mantener
un directorio de migraciones y ejecutar Goose sugiere una historia de versiones
que todavía no existe y permite confundir esa etapa con una evolución segura de
producción.

## Alternativas

### A — Mantener Goose con una única migración editable

- **Ventajas:** conserva el futuro comando de migración desde el inicio.
- **Inconvenientes:** mezcla un esquema reescribible con semántica de historial
  inmutable; añade estado técnico que no aporta valor ahora.
- **Coste de mantenimiento:** bajo, pero confuso.

### B — Aplicar directamente un único esquema SQL reescribible

- **Ventajas:** refleja exactamente que no hay migraciones activas; el mismo SQL
  sirve a PostgreSQL local, sqlc, integración y CI.
- **Inconvenientes:** exige resetear las bases desechables antes de reaplicar el
  esquema.
- **Coste de mantenimiento:** mínimo.

## Decisión del usuario

**Aceptada el 2026-08-02:** elegir la alternativa B mientras se desarrolla la
primera versión. `apps/backend/db/schema/initial_schema.sql` es la única fuente
de verdad y se reescribe cuando cambie el modelo. No se ejecutan migraciones ni
Goose en local o CI.

Cuando exista un entorno, backup o dato que deba preservarse, se abrirá un ADR
sucesor para introducir migraciones incrementales e inmutables y aplicarlas
automáticamente en el despliegue, antes de iniciar la aplicación.

## Consecuencias

- `make db-schema-apply` aplica el esquema a una base PostgreSQL local vacía.
- `make test-integration` y CI aplican ese mismo archivo a una base efímera.
- `sqlc` analiza `db/schema`; no hay directorio de migraciones activo.
- Un reset local borra todos los datos y es requisito antes de reaplicar cambios
  incompatibles.

## Validación

- Una base vacía acepta `initial_schema.sql` con `psql`.
- CI crea PostgreSQL efímero, aplica el esquema y ejecuta la integración.
- Sin `TM_INTEGRATION_DATABASE_URL`, las pruebas nunca usan la base local.

## Disparadores de revisión

- Primer entorno compartido, despliegue, backup o datos que haya que conservar.
- Necesidad de actualizar una base existente sin pérdida de datos.

## Documentación afectada

- [Datos y persistencia](../engineering/DATABASE.md)
- [Desarrollo](../engineering/DEVELOPMENT.md)
- [Pruebas](../engineering/TESTING.md)
- [Runbook PostgreSQL local](../runbooks/local-postgresql.md)
