# ADR-0071: Ejecutar integración PostgreSQL en CI

- **Estado:** Aceptado
- **Fecha:** 2026-08-02
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** ADR-0021, exclusivamente en la exclusión de PostgreSQL en CI
- **Superado por:** ADR-0072, únicamente en el mecanismo de aplicación del esquema

## Problema

La creación e inicio de una liga dependen de restricciones y una transacción
PostgreSQL. Las pruebas unitarias no pueden demostrar que el esquema y el
calendario persisten correctamente en un runner limpio.

## Alternativas

### A — Mantener PostgreSQL solo local

- **Ventajas:** workflow mínimo.
- **Inconvenientes:** CI no demuestra la capa de persistencia real.
- **Coste de mantenimiento:** bajo.

### B — Servicio PostgreSQL efímero en GitHub Actions

- **Ventajas:** reproduce la integración en Linux, sin secretos ni recursos
  cloud propios; la base nace vacía en cada ejecución.
- **Inconvenientes:** añade arranque de servicio y algunos minutos al job.
- **Coste de mantenimiento:** bajo.

## Recomendación

**Opinión/recomendación:** alternativa B, porque ya existe una transacción de
negocio cuyo riesgo justifica la dependencia real.

## Decisión del usuario

**Aceptada el 2026-08-02:** CI incorpora PostgreSQL 18.4 como servicio efímero.
`make verify` ejecuta la integración cuando recibe
`TM_INTEGRATION_DATABASE_URL`; fuera de CI conserva el feedback rápido y muestra
que esa capa se ha omitido si no se proporciona una base aislada.

## Consecuencias

- El job aplica el esquema inicial a una base vacía y ejecuta las pruebas que declaran esa URL.
- No se usan secretos, volúmenes persistentes ni infraestructura AWS.
- Las pruebas de integración no apuntan a la base de desarrollo local.

## Validación

- Un push a GitHub ejecuta la prueba PostgreSQL junto a `make verify`.
- Sin URL de integración, ninguna prueba usa la base local por accidente.

## Documentación afectada

- [CI](../../.github/workflows/verify.yml)
- [Pruebas](../engineering/TESTING.md)
- [Decisiones](../governance/DECISIONS.md)
