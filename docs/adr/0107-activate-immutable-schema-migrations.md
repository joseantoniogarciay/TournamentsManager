# ADR-0107: Activar migraciones inmutables de esquema

- **Estado:** Aceptado
- **Fecha:** 2026-08-22
- **Decisor:** Usuario
- **Supera a:** ADR-0072 y ADR-0053, únicamente en la política de evolución del esquema

## Problema

La base pública de desarrollo ya contiene datos que no deben perderse y la
evidencia legal de ADR-0106 requiere evolucionarla sin resetearla.

## Contexto y restricciones

- ADR-0011 ya acepta Goose y migraciones SQL explícitas.
- ADR-0097 separa la identidad de migración de la identidad de ejecución.
- El esquema inicial es la base 0 de una base vacía; `sqlc` analiza esa base y
  todas las migraciones posteriores.
- Las migraciones nunca se ejecutan al arrancar la API.

## Alternativas

### A — Resetear cada base y reescribir el esquema inicial

- **Ventaja:** operación mínima mientras los datos son desechables.
- **Inconveniente:** destruye datos y no permite una evolución segura.
- **Mantenimiento:** bajo, pero inaceptable desde que hay evidencia que conservar.

### B — Activar Goose con migraciones SQL inmutables

- **Ventaja:** actualiza una base existente de forma ordenada, auditable y sin
  acoplar el dominio al mecanismo de despliegue.
- **Inconveniente:** cada cambio exige una migración, sus permisos y una prueba.
- **Mantenimiento:** bajo y predecible.

## Recomendación

**Recomendación:** B. Es el mecanismo ya elegido en ADR-0011 y el mínimo
necesario para no perder datos.

## Decisión del usuario

**Aceptada el 2026-08-22:** desde ahora cada cambio de esquema se añade como
migración SQL inmutable de Goose. En una base vacía se aplica primero
`initial_schema.sql` y después todas las migraciones pendientes. En una base
existente se ejecutan únicamente las migraciones pendientes.

## Consecuencias

- `initial_schema.sql` permanece como base de bootstrap y, junto con las
  migraciones, como entrada de `sqlc`.
- La primera migración activa crea la evidencia legal y concede sus permisos de
  runtime dentro de la misma operación.
- El servicio de migración usa la credencial de migrador y se invoca de forma
  explícita; la API no la recibe.
- No se reescriben migraciones ya aplicadas. Un error se corrige con una nueva
  migración hacia delante.

## Validación

1. Una base existente alcanza la versión actual sin borrar datos.
2. Una base vacía aplica el esquema inicial y llega a la misma versión.
3. La identidad de ejecución puede operar la tabla nueva y no alterar el
   esquema.
4. Ejecutar de nuevo la migración no realiza cambios.

## Disparadores de revisión

- Una migración incompatible requiera expand/contract, ventana de mantenimiento
  o restauración.
- Producción necesite un runner distinto de Compose o controles adicionales.

## Documentación afectada

- [Datos y persistencia](../engineering/DATABASE.md)
- [Runbook PostgreSQL local](../runbooks/local-postgresql.md)
- [Decisiones](../governance/DECISIONS.md)
