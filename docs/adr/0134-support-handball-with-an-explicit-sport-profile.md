# ADR-0134: Soportar balonmano con un perfil deportivo explícito

- **Estado:** Aceptado
- **Fecha:** 2026-09-21
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** ADR-0126, exclusivamente en el enum inicial de deportes
- **Superado por:** Ninguno

## Problema

El producto admite fútbol y baloncesto mediante políticas deportivas cerradas.
Se necesita incorporar un tercer deporte sin etiquetar como balonmano unas
reglas futbolísticas ni anticipar un motor configurable.

## Contexto y restricciones

- El usuario ha elegido explícitamente balonmano como siguiente deporte.
- Balonmano reutiliza equipos, marcador entero local-visitante, liga a una o dos
  vueltas, eliminatoria directa y el formato mixto existente.
- Una liga admite empate; una eliminatoria debe producir una persona ganadora.
- El backend sigue siendo autoridad de validación, clasificación, desempate y
  retirada.
- No se desglosan periodos ni prórrogas y no se incorporan plantillas,
  estadísticas de jugadores o eliminatorias a doble partido.

## Criterios de decisión

1. representar honestamente puntuación y desempates de balonmano;
2. reutilizar solo los conceptos cuya semántica ya coincide;
3. conservar reglas cerradas, probables y verificables en el dominio;
4. evitar un lenguaje configurable de competiciones;
5. mantener contrato, datos, cliente y observabilidad coherentes.

## Alternativas

### Alternativa A — Reutilizar el perfil de fútbol

- **Ventajas:** cambio pequeño de enum y vocabulario.
- **Inconvenientes:** aplicaría 3-1-0 en vez de 2-1-0 y no explicaría los
  lanzamientos de siete metros.
- **Coste de adopción:** bajo.
- **Coste de mantenimiento:** alto por semántica falsa.
- **Riesgos:** clasificaciones incorrectas y soporte solo nominal.

### Alternativa B — Perfil explícito de balonmano

- **Ventajas:** conserva el modelo común y hace explícitas puntuación,
  desempates, retirada y vocabulario.
- **Inconvenientes:** afecta contrato, esquema, dominio, borradores, UI y
  documentación.
- **Coste de adopción:** medio.
- **Coste de mantenimiento:** bajo mientras la política permanezca cerrada.
- **Riesgos:** mantener supuestos binarios de fútbol y baloncesto en el cliente.

### Alternativa C — Reglamento configurable

- **Ventajas:** podría representar variantes de cada competición.
- **Inconvenientes:** exige lenguaje de reglas, combinaciones válidas y una UI
  de configuración antes de disponer de demanda.
- **Coste de adopción y mantenimiento:** alto.
- **Riesgos:** sobreingeniería y torneos incoherentes.

### No cambiar

Balonmano seguiría sin poder declararse y no se atendería la decisión de
producto.

## Comparación

La alternativa A es barata pero incorrecta. La C optimiza variabilidad no
demostrada. La B añade una tercera política cerrada y reutiliza honestamente el
marcador, las fases y los equipos existentes.

## Recomendación

**Opinión/recomendación:** alternativa B, perfil explícito de balonmano sin
motor configurable.

## Decisión del usuario

**Aceptada el 2026-09-21:** incorporar `handball` con estas reglas:

- el deporte es inmutable y usa equipos como participantes;
- ofrece liga, eliminatoria directa y liga más eliminatoria;
- una liga asigna 2 puntos por victoria, 1 por empate y 0 por derrota;
- los equipos empatados se ordenan por mini-clasificación directa, diferencia
  de goles y goles a favor generales; una igualdad total comparte posición;
- el resultado de liga admite empate y no admite lanzamientos de desempate;
- una eliminatoria registra el marcador final tras una posible prórroga; si
  continúa empatado exige lanzamientos de siete metros decisivos;
- una retirada sustituye uniformemente los partidos por 10-0 para el rival y
  conserva el resultado como administrativo;
- el cliente usa «goles» y nombra el desempate «lanzamientos de 7 metros»;
- no se persiste el desglose de periodos, prórroga o lanzamientos individuales.

## Consecuencias

### Positivas

- El tercer deporte valida la frontera de políticas cerradas sin un motor
  genérico.
- Calendario, fases, participantes y marcador conservan su semántica.
- La clasificación y la retirada son verificables mediante pruebas de dominio.

### Negativas y deuda aceptada

- El perfil es deliberadamente pequeño y no representa todas las normativas
  locales.
- Los campos contractuales de `penalties` conservan su nombre por compatibilidad,
  aunque el cliente los presenta como lanzamientos de siete metros.
- Una futura competición que puntúe o desempate de otro modo requerirá revisar
  esta política cerrada.

## Validación

- Contrato y PostgreSQL aceptan exclusivamente `football`, `basketball` y
  `handball`.
- Las pruebas demuestran 2-1-0, desempate directo y retirada 10-0.
- Liga admite empate; eliminatoria empatada exige lanzamientos de siete metros.
- Borradores locales y transferidos conservan balonmano.
- Los cuatro locales presentan deporte, marcador, clasificación, retirada y
  desempate con vocabulario coherente.
- Las salidas HTTP existentes conservan sus categorías seguras de observabilidad.

## Disparadores de revisión

- Una competición necesita otro sistema de puntos o desempate.
- Se solicita almacenar prórroga o lanzamientos individualmente.
- Se incorporan eliminatorias a doble partido.
- El número de perfiles hace difícil revisar las políticas cerradas actuales.

## Documentación afectada

- [Producto](../project/PRODUCT.md)
- [Modelo inicial](../engineering/INITIAL_DATA_MODEL.md)
- [API](../engineering/API.md)
- [Sistema de diseño](../engineering/DESIGN_SYSTEM.md)
- [Observabilidad](../operations/OBSERVABILITY.md)
- [Decisiones](../governance/DECISIONS.md)
- [Aprendizaje](../project/LEARNING.md)
