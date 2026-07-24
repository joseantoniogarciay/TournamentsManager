# ADR-0000: Registrar decisiones arquitectónicas

- **Estado:** Aceptado
- **Fecha:** 2026-07-23
- **Decisor:** Usuario, mediante el manifiesto
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

El proyecto busca aprendizaje profesional y necesita conservar no solo qué se
eligió, sino por qué, frente a qué alternativas y con qué consecuencias.

## Contexto y restricciones

El manifiesto exige comparar alternativas, dar al usuario la decisión final,
registrar toda decisión importante en un ADR y mantener la documentación junto al
código.

## Alternativas

### ADR versionados

Ventajas: contexto cercano al repositorio, revisión junto al cambio, historial y
enlaces. Inconvenientes: disciplina de mantenimiento y posible exceso si se usan
para decisiones triviales.

### Solo documentación temática

Ventajas: menos archivos. Inconvenientes: se pierde el contexto histórico y se
mezclan estado actual y debate.

### Solo historial de commits o conversaciones

Ventajas: sin proceso adicional. Inconvenientes: baja descubribilidad, contexto
fragmentado y decisiones sin estado claro.

## Decisión del usuario

Usar ADR para todas las decisiones importantes. El usuario conserva la autoridad
final y la documentación se actualiza junto a cada decisión.

## Consecuencias

- `docs/governance/DECISIONS.md` mantiene el índice y el umbral.
- Los ADR usan estados explícitos.
- Una propuesta no autoriza implementación.
- Las decisiones triviales no necesitan ADR hasta que sean transversales o
  costosas de revertir.

## Validación

En cada fase, la retrospectiva comprobará que decisiones e implementación son
trazables y que no existen ADR huérfanos.

## Disparadores de revisión

- El proceso bloquea cambios pequeños de forma recurrente.
- Se pierden decisiones importantes a pesar del proceso.
- El equipo necesita una herramienta adicional para gobierno.
