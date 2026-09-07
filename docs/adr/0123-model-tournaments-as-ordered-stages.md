# ADR-0123: Modelar los torneos como fases ordenadas

- **Estado:** Aceptado
- **Fecha:** 2026-09-06
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** ADR-0122, exclusivamente en la decisión de no introducir fases
- **Superado por:** Ninguno

## Problema

Un torneo podrá combinar en el futuro una liga única o grupos con una
eliminatoria posterior. Las plazas de un cuadro deberán poder proceder de la
clasificación de una fase anterior, además de equipos sembrados o ganadores de
partidos. Si los partidos pertenecen directamente al torneo, incorporar esa
composición obligaría a migrar de nuevo su identidad y sus relaciones.

## Contexto y restricciones

- ADR-0122 adopta `tournament` como recurso raíz, conserva `league` y añade
  `single_elimination`.
- La entrega actual conserva la liga a una o dos vueltas e incorpora, como
  alternativa, un torneo con una fase de eliminatoria directa a partido único.
- La futura primera fase puede ser una liga con todos los equipos o varios
  grupos; el creador configurará únicamente combinaciones que tengan sentido.
- Todavía no están decididas las reglas de grupos, clasificación, desempate,
  cupos, siembra, calendario ni qué combinaciones de fases se permitirán.

## Criterios de decisión

1. conservar una relación explícita entre una plaza de cuadro y su origen;
2. evitar migrar otra vez todos los partidos al añadir una fase previa;
3. no construir un motor de competición, grupos o clasificación especulativo;
4. mantener las reglas concretas dentro de cada tipo de fase.

## Alternativas

### Alternativa A — Partidos directamente bajo el torneo

- **Ventajas:** modelo inicial más pequeño.
- **Inconvenientes:** liga, grupos y cuadro comparten una colección sin frontera
  de reglas; los clasificadores futuros no tienen destino estructural.
- **Coste de adopción:** bajo.
- **Coste de mantenimiento:** alto al incorporar la segunda fase.
- **Riesgos:** migración posterior de identidad y dependencias de partidos.

### Alternativa B — Fases ordenadas mínimas

- **Ventajas:** cada partido pertenece a una fase y los huecos de una fase
  posterior pueden declarar su procedencia sin conocer aún la implementación de
  grupos.
- **Inconvenientes:** añade una entidad y una relación a la entrega actual.
- **Coste de adopción:** medio.
- **Coste de mantenimiento:** bajo; solo aumenta con tipos de fase que se
  acepten expresamente.
- **Riesgos:** convertir la entidad en un motor genérico prematuro.

### Alternativa C — Motor configurable de fases y reglas desde ahora

- **Ventajas:** máxima flexibilidad aparente.
- **Inconvenientes:** requiere decidir configuraciones, clasificación,
  validación, dependencias y UX antes de tener casos de uso concretos.
- **Coste de adopción:** alto.
- **Coste de mantenimiento:** alto.
- **Riesgos:** sobreingeniería y reglas deportivas incoherentes.

### No cambiar

El cuadro actual funcionaría, pero su evolución hacia clasificación por grupos
requeriría otra migración estructural de alto impacto.

## Comparación

La A simplifica el primer cuadro pero traslada el coste al requisito ya conocido
de fases consecutivas. La C adelanta decisiones de producto que el usuario no
ha tomado. La B fija solo la frontera necesaria para expresar el futuro sin
decidirlo: una secuencia de fases y fuentes de plazas tipadas.

## Recomendación

**Opinión/recomendación:** alternativa B, fases ordenadas mínimas.

## Decisión del usuario

**Aceptada el 2026-09-06:** un torneo contiene fases ordenadas. En esta entrega
cada torneo tiene una única fase, que puede ser `league` (una o dos vueltas) o
`single_elimination` (partido único). Liga por grupos, clasificación y
combinaciones de fases se decidirán e implementarán después.

Los partidos pertenecen a una fase. Cada plaza de partido declara una fuente:
equipo sembrado, ganador de partido o *bye* ahora; una posición clasificada de
grupo será una fuente futura. El servidor, no el cliente, valida que la fuente
esté disponible y que su uso sea compatible con la fase destino.

## Consecuencias

### Positivas

- El árbol de brackets soporta de forma nativa participantes aún desconocidos.
- Una futura clasificación puede alimentar un cuadro sin alterar la pertenencia
  de los partidos ni inventar identificadores de equipo temporales.
- Las reglas de liga y eliminatoria permanecen separadas por fase.

### Negativas y deuda aceptada

- Un torneo de una sola fase tiene una relación adicional que no es visible para
  la persona usuaria.
- No existe aún configuración de grupos ni transición automática entre fases.
- La selección de fases que tendrá el creador permanece pendiente de un ADR de
  producto posterior.

## Validación

- Toda fase tiene un orden único dentro del torneo y todo partido una fase.
- En la eliminatoria actual, cada plaza posee una fuente válida y visible en la
  proyección pública.
- Ningún resultado puede propagar una ganadora a una plaza que no la declare
  como fuente.
- Añadir en el futuro una fuente `group_ranking` no requiere cambiar la relación
  entre torneo, fase y partido.

## Disparadores de revisión

- Se decide la configuración de grupos o la clasificación hacia una fase
  posterior.
- Se aceptan fases simultáneas, repesca, doble eliminación o terceros puestos.
- Una fase necesita un ciclo de vida distinto del torneo completo.

## Documentación afectada

- [Producto](../project/PRODUCT.md)
- [Modelo inicial de datos](../engineering/INITIAL_DATA_MODEL.md)
- [API](../engineering/API.md)
- [OpenAPI v1](../../contracts/openapi/v1/openapi.yaml)
- [Observabilidad](../operations/OBSERVABILITY.md)
- [Decisiones](../governance/DECISIONS.md)
- [Aprendizaje](../project/LEARNING.md)
