# ADR-0047: Organizar el esquema PostgreSQL inicial y la generación sqlc

- **Estado:** Superado parcialmente por ADR-0053
- **Fecha:** 2026-07-27
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** ADR-0053, únicamente en la política de evolución del esquema local

## Problema

El modelo lógico aceptado necesita una única fuente de verdad ejecutable antes de
implementar adaptadores. Hay que organizar migraciones, futuras consultas y código
generado sin duplicar el esquema ni introducir repositorios genéricos.

## Alternativas

### A — Migraciones secuenciales, sqlc desde migraciones y salida versionada

- **Ventajas:** orden legible; una sola fuente de esquema; revisión local de la
  generación; UUIDv7 nativo de PostgreSQL 18.
- **Inconvenientes:** el directorio generado debe regenerarse y revisarse.
- **Coste de mantenimiento:** bajo o medio.

### B — Migraciones por timestamp y esquema SQL duplicado para sqlc

- **Ventajas:** menor riesgo de colisión con muchos equipos.
- **Inconvenientes:** dos representaciones del esquema pueden divergir.
- **Coste de mantenimiento:** medio.

### C — Migraciones y repositorios por módulo desde el inicio

- **Ventajas:** separación física máxima.
- **Inconvenientes:** multiplica configuración y paquetes antes de necesitarlos.
- **Coste de mantenimiento:** alto inicialmente.

## Decisión del usuario

**Aceptada el 2026-07-27:** adoptar alternativa A.

- Migraciones SQL secuenciales en `apps/backend/db/migrations/`.
- Consultas en `apps/backend/db/queries/` y código sqlc versionado bajo
  `apps/backend/internal/adapters/postgres/sqlc/`.
- `apps/backend/sqlc.yaml` usa directamente las migraciones como esquema y
  `pgx/v5`, sin interfaz generada ni sentencias preparadas globales.
- PostgreSQL genera IDs con `uuidv7()` y las fechas usan `timestamptz`.
- `make sqlc-generate` y `make sqlc-generate-check` regeneran y detectan deriva.
- La migración inicial protege con restricciones, FKs e índices las invariantes
  expresables por PostgreSQL; las transacciones siguen perteneciendo a futuros
  casos de uso.

## Consecuencias

- No se necesita extensión PostgreSQL ni dependencia Go adicional para UUIDv7.
- La migración posee identidad, datos temporales y liga, sin datos semilla,
  resultados, roles delegados ni auditoría futura.
- No se escriben consultas de relleno; sqlc generará código al aparecer la
  primera operación de persistencia.

## Validación

- Goose aplica la migración desde una base vacía y se repite sin cambios.
- La base rechaza duplicados dentro de un borrador o liga, pares de equipos
  repetidos y huellas de token de tamaño inválido.
- sqlc genera de forma determinista cuando haya consultas SQL.

## Disparadores de revisión

- Colisiones recurrentes de numeración entre ramas.
- sqlc no analiza de forma mantenible el esquema desde migraciones.
- UUIDv7 o los índices de las consultas reales revelan un coste inaceptable.

## Documentación afectada

- [Datos y persistencia](../engineering/DATABASE.md)
- [Desarrollo](../engineering/DEVELOPMENT.md)
- [Aprendizaje](../project/LEARNING.md)
- [Decisiones](../governance/DECISIONS.md)
