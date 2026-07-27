# ADR-0046: Validar y generar OpenAPI con Redocly CLI y Orval

- **Estado:** Aceptado
- **Fecha:** 2026-07-26
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

El contrato OpenAPI debe validarse de forma reproducible y producir un cliente
TypeScript sin copiar DTOs u operaciones HTTP a mano. Sin comprobar deriva,
contrato y cliente pueden representar APIs distintas.

## Contexto y restricciones

- ADR-0009 establece OpenAPI como fuente de verdad y cliente TypeScript generado.
- ADR-0045 fija OpenAPI 3.1 en `contracts/openapi/v1/openapi.yaml`.
- El cliente universal usa TypeScript y Fetch; React Query y Axios no están
  aceptados.
- Cookie/Bearer sigue siendo responsabilidad de una capa de transporte manual,
  no del dominio ni de la generación.

## Alternativas

### A — Redocly CLI para lint y Orval con cliente Fetch

- **Ventajas:** validación de especificación y reglas de diseño; tipos y
  funciones Fetch generadas; sin Java ni framework de datos.
- **Inconvenientes:** dos dependencias Node y revisión del resultado generado.
- **Coste de mantenimiento:** bajo o medio.

### B — OpenAPI Generator con `typescript-fetch`

- **Ventajas:** ecosistema amplio y generadores para muchos lenguajes.
- **Inconvenientes:** JVM y configuración mayor para un único consumidor
  TypeScript.
- **Coste de mantenimiento:** medio.

### C — Solo tipos generados y cliente HTTP manual

- **Ventajas:** menos dependencia y control máximo del transporte.
- **Inconvenientes:** duplica URLs y serialización; la deriva es más probable.
- **Coste de mantenimiento:** medio.

## Decisión del usuario

**Aceptada el 2026-07-26:** adoptar alternativa A.

- `@redocly/cli` 2.40.0 aplica el ruleset recomendado y reglas explícitas.
- `orval` 8.23.0 genera un cliente Fetch por tags en
  `apps/client/src/api/generated/`, ruta reservada a generación.
- `pnpm run openapi:lint` valida; `openapi:generate` actualiza; y
  `openapi:generate:check` detecta deriva.
- `make check` ejecuta lint OpenAPI y `make verify` añade la deriva.
- No se aprueban React Query, Axios, mocks ni capa de sesión automática.

## Consecuencias

- La salida generada se versiona y no se edita manualmente.
- Si hay deriva, la comprobación falla tras regenerar para que el cambio sea
  visible y se revise explícitamente.
- La capa futura que configure URL base, cookies web y Bearer móvil invocará las
  funciones generadas sin cambiar el contrato ni el dominio.

## Validación

- Redocly valida `tournaments-manager@v1` sin errores.
- Orval genera el cliente desde el contrato sin servidor activo.
- Tras versionar la salida, la comprobación de deriva no muestra cambios.
- Formato, lint y typecheck siguen pasando con los artefactos generados.

## Disparadores de revisión

- El cliente universal necesita caché, reintentos, React Query o un adaptador
  distinto de Fetch.
- La generación genera diffs difíciles de revisar o no cubre OpenAPI 3.1.
- Aparece un consumidor que justifique otro lenguaje o un SDK publicado.

## Documentación afectada

- [API](../engineering/API.md)
- [Desarrollo](../engineering/DEVELOPMENT.md)
- [Aprendizaje](../project/LEARNING.md)
- [Decisiones](../governance/DECISIONS.md)
