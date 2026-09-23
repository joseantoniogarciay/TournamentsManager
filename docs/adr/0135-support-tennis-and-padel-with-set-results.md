# ADR-0135: Soportar tenis y pádel con resultados por sets

- **Estado:** Aceptado
- **Fecha:** 2026-09-23
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** ADR-0126 y ADR-0134, exclusivamente para ampliar los perfiles
  deportivos con resultados multidimensionales
- **Superado por:** Ninguno

## Problema

El torneo solo representa resultados mediante un tanteo entero por participante.
Ese dato no puede explicar ni validar un partido de raqueta: conocer únicamente
que terminó 2-1 no demuestra que cada set sea posible ni que el encuentro se
detuviera cuando una participante alcanzó la victoria.

## Contexto y restricciones

- El usuario ha solicitado deportes de raqueta, resultados por sets y elección
  entre partidos al mejor de tres o cinco sets al crear el torneo.
- Tenis admite el mejor de tres o cinco; pádel se mantiene al mejor de tres.
- Esta entrega usa sets normales con tie-break: `6-0` a `6-4`, `7-5` o `7-6`.
- Los nombres representan participantes individuales o parejas sin modelar
  plantillas ni cuentas de jugadores.
- Solo se habilita eliminatoria directa. Una liga necesitaría una decisión
  adicional sobre puntos, clasificación y desempates.
- No se incorporan abandonos, incomparecencias, super tie-break, advantage set
  ni el tanteo interno del tie-break.

## Alternativas

### Alternativa A — Guardar solo sets ganados

- **Ventajas:** reutiliza el marcador entero existente.
- **Inconvenientes:** acepta sets imposibles y no conserva el desarrollo real.
- **Coste de mantenimiento:** bajo, con semántica insuficiente.

### Alternativa B — Perfiles explícitos con cada set

- **Ventajas:** valida el resultado completo, conserva historial y mantiene las
  reglas dentro del dominio.
- **Inconvenientes:** modifica contrato, persistencia, dominio y cliente.
- **Coste de mantenimiento:** medio y acotado por perfiles cerrados.

### Alternativa C — Motor configurable de deportes de raqueta

- **Ventajas:** permitiría combinar formatos y tanteos arbitrarios.
- **Inconvenientes:** exige un lenguaje de reglas, validación de combinaciones y
  una interfaz compleja antes de conocer la demanda.
- **Coste de mantenimiento:** alto.

## Recomendación

**Opinión/recomendación:** alternativa B. Es el modelo mínimo que puede impedir
resultados incoherentes sin anticipar un motor configurable.

## Decisión del usuario

**Aceptada el 2026-09-23:** incorporar perfiles explícitos `tennis` y `padel`.

- Tenis permite elegir `bestOfSets: 3 | 5` al crear el torneo.
- Pádel fija `bestOfSets: 3` y no ofrece una elección inválida.
- Ambos deportes solo admiten `single_elimination` en esta entrega.
- El resultado contiene los sets en orden y el tanteo de juegos de cada lado.
- Cada set tiene una única ganadora y solo puede terminar `6-0` a `6-4`, `7-5`
  o `7-6`.
- El partido termina exactamente cuando una participante alcanza dos sets en un
  mejor de tres o tres sets en un mejor de cinco; no admite sets posteriores.
- El tanteo agregado se deriva de los sets y nunca lo introduce el cliente.
- El resultado y cada corrección se guardan atómicamente con su historial.
- La interfaz usa «participantes» para los nombres de persona o pareja y no
  introduce una entidad de plantilla.

## Consecuencias

### Positivas

- El backend impide resultados parciales, empatados o con sets imposibles.
- La ganadora del cuadro se deriva de una única fuente de verdad.
- Tenis y pádel comparten estructura sin fingir que comparten toda configuración.

### Negativas y deuda aceptada

- Los nombres internos históricos de `team` continúan en persistencia y contrato;
  el cliente presenta vocabulario neutral para los perfiles de raqueta.
- Pádel a cinco sets y variantes de set requieren una decisión posterior.
- Las ligas de raqueta quedan fuera hasta aceptar sus reglas de clasificación.

## Validación

- Contrato y PostgreSQL solo aceptan `bestOfSets` para tenis o pádel y fijan
  pádel a tres.
- El dominio rechaza sets empatados, marcadores inválidos, resultados sin
  ganadora y sets jugados después de decidir el encuentro.
- Una corrección sustituye los sets actuales y añade una instantánea histórica
  dentro de la misma transacción.
- La eliminatoria propaga exclusivamente la ganadora derivada de los sets.
- Web, iOS y Android crean y muestran la configuración y editan todos los sets
  mediante textos localizados.

## Disparadores de revisión

- Se solicitan ligas de tenis o pádel.
- Se necesitan super tie-break, advantage set, abandono o incomparecencia.
- Se incorporan bádminton, tenis de mesa u otro deporte con tanteo por puntos.
- Se necesita identificar por separado a las personas de una pareja.

## Documentación afectada

- [Producto](../project/PRODUCT.md)
- [Modelo inicial](../engineering/INITIAL_DATA_MODEL.md)
- [API](../engineering/API.md)
- [Sistema de diseño](../engineering/DESIGN_SYSTEM.md)
- [Observabilidad](../operations/OBSERVABILITY.md)
- [Decisiones](../governance/DECISIONS.md)
- [Aprendizaje](../project/LEARNING.md)
