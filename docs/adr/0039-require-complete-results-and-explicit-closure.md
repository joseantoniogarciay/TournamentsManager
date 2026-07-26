# ADR-0039: Requerir resultados completos y cierre explícito de la liga

- **Estado:** Aceptado
- **Fecha:** 2026-07-26
- **Decisor:** Usuario
- **Propietario del análisis:** Codex
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

Una liga debe llegar a `finalizado` sin partidos pendientes ni una finalización
automática provocada por una entrada accidental.

## Contexto y restricciones

- Los resultados son goles locales y visitantes no negativos (ADR-0037).
- Solo se registran en `en_curso` (ADR-0038).
- Los administradores aplican y corrigen resultados inmediatamente (ADR-0035 y
  ADR-0036).
- `cancelado` sigue siendo el cierre alternativo (ADR-0032).

## Criterios de decisión

1. impedir clasificaciones incompletas;
2. preservar una revisión final del creador;
3. no introducir incidencias deportivas con semántica propia todavía.

## Alternativas

### Alternativa A — Cierre explícito con resultados completos

El creador solo finaliza cuando todos los partidos tienen resultado. Una
excepción se anota como marcador normal, por ejemplo `3-0`, sin tipo especial.

- **Ventajas:** consistencia, revisión final y modelo pequeño.
- **Inconvenientes:** no conserva el motivo especial de un marcador.
- **Coste de adopción:** bajo.
- **Coste de mantenimiento:** bajo.
- **Riesgos:** un resultado administrativo no se distingue de uno jugado.

### Alternativa B — Cierre automático al último resultado

- **Ventajas:** una acción menos.
- **Inconvenientes:** un error puede cerrar la liga antes de corregirse.
- **Coste de adopción:** bajo.
- **Coste de mantenimiento:** bajo.
- **Riesgos:** finalizaciones involuntarias.

### Alternativa C — Cierre parcial o incidencias explícitas

- **Ventajas:** cubre abandonos, sanciones e incomparecencias ricas.
- **Inconvenientes:** exige reglas de clasificación y responsables no decididos.
- **Coste de adopción:** alto.
- **Coste de mantenimiento:** alto.
- **Riesgos:** sobreingeniería del primer corte.

### No cambiar

No habría una regla verificable para cerrar una liga.

## Comparación

La B ahorra una acción, pero pierde la revisión final. La C adelanta reglas que
no existen. La A conserva una clasificación completa con la menor semántica.

## Recomendación

**Recomendación:** alternativa A, resultados completos y cierre explícito.

## Decisión del usuario

**Aceptada el 2026-07-26:** el creador finaliza explícitamente la liga solo con
todos los partidos resueltos. Las excepciones se anotan por ahora como marcador
normal —por ejemplo, `3-0`— sin motivo ni tipo de resultado. Si una liga no puede
terminar así, se cancelará o se aplicará una decisión futura específica.

## Consecuencias

### Positivas

- Una liga finalizada contiene resultados para todos sus partidos.
- El creador tiene una última revisión antes de congelarla.

### Negativas y deuda aceptada

- No se distingue un `3-0` jugado de uno administrativo.
- No se puede corregir una liga finalizada en este corte.

## Validación

- Solo el creador puede ejecutar `en_curso → finalizado`.
- La transición se rechaza si hay un partido sin resultado.
- Una liga finalizada no admite nuevos resultados ni correcciones.

## Disparadores de revisión

- Necesidad de conservar motivos de incomparecencia, sanción o anulación.
- Necesidad de corregir una liga finalizada.
- Clasificación que admita partidos no disputados.

## Documentación afectada

- [PRODUCT.md](../project/PRODUCT.md)
- [LEARNING.md](../project/LEARNING.md)
- [DECISIONS.md](../governance/DECISIONS.md)
