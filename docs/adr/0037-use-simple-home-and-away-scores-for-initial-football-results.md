# ADR-0037: Usar marcadores simples local-visitante en resultados iniciales

- **Estado:** Aceptado
- **Fecha:** 2026-07-26
- **Decisor:** Usuario
- **Propietario del análisis:** Codex
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

Para aplicar un resultado en una liga de fútbol hay que definir el dato mínimo
que representa el desenlace de un partido sin adelantar reglas de formatos o
incidencias deportivas que el primer corte no requiere.

## Contexto y restricciones

- La liga inicial es a una vuelta y usa puntuación 3-1-0, según ADR-0032.
- Administradores delegados registran y corrigen resultados inmediatamente con
  historial, según ADR-0035 y ADR-0036.
- No hay eliminatorias, calendario completo ni arbitraje en el primer alcance.

## Criterios de decisión

1. representar el resultado habitual de fútbol con validación simple;
2. permitir victorias, derrotas y empates;
3. evitar datos sin reglas de clasificación aceptadas;
4. conservar una evolución explícita para incidencias futuras.

## Alternativas

### Alternativa A — Goles de local y visitante

Cada resultado tiene dos enteros no negativos: goles del equipo local y goles del
visitante. `0-0` es válido.

- **Ventajas:** modelo mínimo, inequívoco y suficiente para la puntuación 3-1-0.
- **Inconvenientes:** no describe incomparecencias, sanciones, prórroga ni
  penaltis.
- **Coste de adopción:** bajo.
- **Coste de mantenimiento:** bajo.
- **Riesgos:** necesitar una extensión explícita al incorporar casos especiales.

### Alternativa B — Marcador con incidencias desde el inicio

- **Ventajas:** cubre más sucesos deportivos.
- **Inconvenientes:** obliga a decidir efectos de cada incidencia sobre puntos,
  goles, clasificación y correcciones.
- **Coste de adopción:** alto.
- **Coste de mantenimiento:** alto.
- **Riesgos:** reglas especulativas y resultados difíciles de explicar.

### Alternativa C — Texto libre de resultado

- **Ventajas:** gran flexibilidad aparente.
- **Inconvenientes:** no se puede validar ni calcular una clasificación fiable.
- **Coste de adopción:** bajo.
- **Coste de mantenimiento:** alto por interpretación manual.
- **Riesgos:** datos inconsistentes.

### No cambiar

No se podría validar ni aplicar un resultado de forma consistente.

## Comparación

La B anticipa excepciones sin reglas acordadas y la C elimina la semántica que
una liga necesita. La A es suficiente para el formato vigente y deja la
extensión futura como una decisión de negocio explícita.

## Recomendación

**Recomendación:** alternativa A, marcador simple local-visitante.

## Decisión del usuario

**Aceptada el 2026-07-26:** un resultado inicial consta solo de goles locales y
goles visitantes, ambos enteros no negativos; `0-0` es válido. Prórroga,
penaltis, incomparecencias, sanciones y otros resultados especiales quedan fuera
del primer corte.

## Consecuencias

### Positivas

- Los resultados alimentan directamente la regla 3-1-0 y futuras vistas de
  clasificación.
- Validar un marcador es simple y no depende de texto libre.

### Negativas y deuda aceptada

- Una incidencia deportiva no puede registrarse todavía.
- Los formatos sin empate requerirán otra decisión de modelo.

## Validación

- El sistema acepta dos goles enteros no negativos, incluido `0-0`.
- Los valores negativos, ausentes o no enteros se rechazan.
- Ningún resultado inicial contiene campos de prórroga, penaltis o sanciones.

## Disparadores de revisión

- Incomparecencias, sanciones o partidos anulados.
- Incorporación de eliminatorias o deportes con otra semántica de resultado.
- Necesidad de registrar incidencias más allá del marcador.

## Documentación afectada

- [PRODUCT.md](../project/PRODUCT.md)
- [LEARNING.md](../project/LEARNING.md)
- [DECISIONS.md](../governance/DECISIONS.md)
