# ADR-0066: Denegar vinculación y fusión entre cuentas distintas

- **Estado:** Aceptado
- **Fecha:** 2026-08-01
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** ADR-0010 y ADR-0050, exclusivamente en la vinculación de una identidad externa nueva cuando su email coincide con otra cuenta interna
- **Superado por:** Ninguno

## Problema

El flujo previsto permitía confirmar por email la vinculación de una identidad
externa a una cuenta local candidata. También se consideró una futura fusión de
cuentas distintas. Ambos caminos añaden estados, pruebas de posesión y traslado
de relaciones de negocio cuyo coste no se justifica para el producto inicial.

## Contexto y restricciones

- La identidad de proveedor se identifica únicamente por `(issuer, subject)`.
- Una cuenta interna conserva sus relaciones de torneos y sus métodos de acceso.
- Una coincidencia de email no identifica de forma estable a una persona ni
  autoriza acceso a otra cuenta.
- Google es el único proveedor federado del primer incremento; Apple y otros
  proveedores quedan fuera de alcance.

## Criterios de decisión

1. seguridad y límites de propiedad fáciles de explicar;
2. menor superficie de identidad y recuperación;
3. ausencia de migraciones de datos de torneos entre cuentas;
4. experiencia recuperable sin revelar cuentas innecesariamente.

## Alternativas

### Alternativa A — Vinculación confirmada y futura fusión

- **Ventajas:** permite reunir métodos y datos de cuentas que parezcan de la misma persona.
- **Inconvenientes:** exige acreditar dos cuentas, resolver credenciales, sesiones,
  conflictos de datos, auditoría y transferencia de permisos de torneos.
- **Coste de adopción:** alto.
- **Coste de mantenimiento:** alto y permanente.
- **Riesgos:** escalada de privilegios o traslado incorrecto de relaciones.

### Alternativa B — Denegar vínculos entre cuentas

- **Ventajas:** cada cuenta y sus datos mantienen un único propietario; no hay
  fusión, cambio de IDs ni estados de vinculación entre cuentas.
- **Inconvenientes:** una persona con varias cuentas no puede unificarlas.
- **Coste de adopción:** bajo.
- **Coste de mantenimiento:** bajo.
- **Riesgos:** duplicidad voluntaria o accidental de cuentas.

### No cambiar

- **Consecuencias:** se conserva la complejidad de confirmación de vínculos y se
  deja abierta una futura fusión de datos de torneos.

## Comparación

La alternativa A prioriza comodidad para casos poco frecuentes, pero convierte
la identidad en un mecanismo de migración de datos de negocio. B mantiene el
límite más simple: una identidad externa ya pertenece a una cuenta o no se puede
usar para acceder a otra. La duplicidad es preferible a trasladar permisos y
propiedad sin una necesidad demostrada.

## Recomendación

**Recomendación:** alternativa B por simplicidad, menor superficie de seguridad
y ausencia de operaciones irreversibles sobre datos de torneos.

## Decisión del usuario

**Aceptada el 2026-08-01:** alternativa B.

- Si el email verificado de una identidad externa nueva ya pertenece a una cuenta
  interna, se deniega el alta o la vinculación; no se envía desafío de enlace ni
  se crea sesión.
- Si `(issuer, subject)` ya pertenece a una cuenta, solo puede iniciar sesión en
  esa cuenta; no se puede vincular a otra.
- No existe ni se planifica una fusión de cuentas ni la transferencia de IDs o
  relaciones de torneos entre ellas.

## Consecuencias

### Positivas

- La titularidad de ligas, administraciones y seguimientos nunca se desplaza
  entre cuentas por un flujo de identidad.
- Se eliminan los intentos de vinculación y sus enlaces de confirmación.
- Los errores de conflicto tienen una recuperación sencilla: usar el método de
  acceso que ya pertenece a la cuenta correspondiente.

### Negativas y deuda aceptada

- Las cuentas creadas con distintos emails o proveedores permanecen separadas.
- Una futura necesidad real de unificación requerirá una decisión nueva, no una
  extensión implícita de esta regla.

## Validación

- Una identidad Google ya vinculada abre sesión únicamente en su cuenta.
- Una identidad Google nueva cuyo email coincide con otra cuenta no crea cuenta,
  identidad ni sesión.
- Ninguna operación modifica `account_id` en ligas, administraciones o
  seguimientos para resolver una identidad.

## Disparadores de revisión

- Evidencia de que la duplicidad de cuentas impide una tarea esencial del
  producto.
- Requisito legal u operativo de consolidar datos de dos cuentas con autorización
  verificable.

## Documentación afectada

- `docs/engineering/IDENTITY.md`
- `docs/engineering/INITIAL_DATA_MODEL.md`
- `docs/governance/DECISIONS_TO_REVISIT.md`
- `docs/governance/DECISIONS.md`
- `contracts/openapi/v1/openapi.yaml`
