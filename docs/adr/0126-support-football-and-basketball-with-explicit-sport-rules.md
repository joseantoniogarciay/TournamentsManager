# ADR-0126: Soportar fútbol y baloncesto con reglas deportivas explícitas

- **Estado:** Aceptado
- **Fecha:** 2026-09-13
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** ADR-0037, ADR-0041 y ADR-0081 exclusivamente en su alcance
  futbolístico como regla única para todos los torneos
- **Superado por:** Ninguno

## Problema

El recurso raíz ya se llama torneo y admite liga o eliminatoria, pero el deporte
permanece fijado a fútbol. El contrato no permite elegirlo y resultado,
clasificación, desempate y retirada aplican reglas y vocabulario de fútbol. Solo
ampliar un enum permitiría crear torneos etiquetados como baloncesto con una
clasificación falsa.

## Contexto y restricciones

- Las sugerencias privadas permiten observar demanda de deportes posteriores.
- El primer deporte adicional solicitado es baloncesto; fútbol sala permanece
  incluido en «Fútbol» mientras no necesite comportamiento propio.
- Ambos deportes usan equipos, marcador entero local-visitante, liga a una o dos
  vueltas y eliminatoria directa a partido único.
- La autoridad deportiva sigue en el backend; el cliente selecciona y presenta
  las reglas mediante el enum recibido.
- No se crea un motor configurable, estadísticas por periodo, prórrogas
  desglosadas, plantillas ni series al mejor de N.

## Criterios de decisión

1. no mostrar una clasificación incorrecta para el deporte elegido;
2. mantener explícitas y probables las reglas de cada deporte;
3. pedir a la organizadora solo decisiones que cambien realmente su torneo;
4. conservar un marcador común cuando su semántica sí coincide;
5. permitir añadir otro deporte sin repartir condicionales de negocio por la
   infraestructura o el cliente.

## Alternativas

### Alternativa A — Cambiar únicamente el vocabulario

- **Ventajas:** cambio pequeño; el marcador existente acepta tanteos altos.
- **Inconvenientes:** conserva 3-1-0, empates y desempates futbolísticos.
- **Coste de adopción:** bajo.
- **Coste de mantenimiento:** alto por datos semánticamente falsos.
- **Riesgos:** proclamar soporte que no existe.

### Alternativa B — Perfiles explícitos de fútbol y baloncesto

- **Ventajas:** enum cerrado, reglas verificables y una frontera clara por
  deporte sin lenguaje configurable.
- **Inconvenientes:** modifica contrato, persistencia, dominio, borradores y UI.
- **Coste de adopción:** medio.
- **Coste de mantenimiento:** bajo mientras se añadan deportes solo con demanda.
- **Riesgos:** dispersar `if` por deporte si no se centralizan sus políticas.

### Alternativa C — Motor genérico de reglas deportivas

- **Ventajas:** flexibilidad aparente para puntuaciones y desempates futuros.
- **Inconvenientes:** obliga a diseñar un lenguaje de reglas, validarlo y
  explicarlo antes de conocer los siguientes deportes.
- **Coste de adopción y mantenimiento:** alto.
- **Riesgos:** sobreingeniería e interacciones inválidas entre opciones.

### No cambiar

Baloncesto seguiría sin poder declararse o produciría UX y clasificación de
fútbol.

## Comparación

La A reduce trabajo pero incumple el criterio principal. La C anticipa una
variabilidad no demostrada. La B introduce únicamente las dos políticas
necesarias y mantiene el marcador común, que sí representa honestamente el
resultado final de ambos deportes.

## Recomendación

**Opinión/recomendación:** alternativa B. Elegir el deporte al crear, congelarlo
para toda la vida del torneo y seleccionar mediante él validación, clasificación,
desempate, retirada y vocabulario.

## Decisión del usuario

**Aceptada el 2026-09-13:** adoptar la alternativa B con estas reglas:

- el enum inicial es `football | basketball`; «Fútbol» engloba fútbol sala;
- el deporte se elige al crear el torneo y no se edita después;
- ambos deportes ofrecen liga a una o dos vueltas y eliminatoria directa;
- fútbol conserva resultado, 3-1-0, desempates y retirada administrativa 3-0;
- baloncesto no admite un resultado final empatado;
- su liga asigna 2 puntos por victoria, 1 por derrota jugada y 0 por derrota
  administrativa; desempata por mini-clasificación directa y después por
  diferencia y tanteo general;
- una retirada de baloncesto sustituye uniformemente todos sus partidos por
  20-0 para el rival, conserva el historial y no ofrece configurar ese valor;
- el resultado administrativo se distingue de uno jugado;
- el cliente usa «goles» para fútbol y «puntos» para baloncesto.

## Consecuencias

### Positivas

- Un torneo declara honestamente su deporte desde el borrador.
- Baloncesto reutiliza participantes, calendario y marcador sin heredar reglas
  futbolísticas.
- La configuración evita opciones que la organizadora no necesita decidir.

### Negativas y deuda aceptada

- La puntuación y retirada siguen un perfil FIBA deliberadamente pequeño, no
  todos los reglamentos locales posibles.
- No se desglosan periodos ni prórrogas; se registra el resultado final.
- Una retirada completa se trata como derrotas administrativas uniformes; no se
  ofrece todavía anular todos los partidos.
- Incorporar deportes de sets o participantes individuales requerirá otro
  modelo de resultado o participante.

## Validación

- Contrato y base de datos rechazan deportes fuera del enum.
- Los borradores locales y transferidos conservan el deporte elegido.
- Un empate de baloncesto se rechaza tanto en liga como en eliminatoria.
- Una retirada produce 3-0 en fútbol o 20-0 en baloncesto y queda marcada como
  administrativa.
- Las pruebas de dominio demuestran 3-1-0 para fútbol, 2-1 para un partido jugado
  de baloncesto, 2-0 para uno administrativo y sus desempates.
- Los cuatro locales muestran selector y vocabulario coherentes.

## Disparadores de revisión

- Los usuarios piden resultados administrativos configurables.
- Un reglamento necesita anular partidos tras una retirada.
- Fútbol sala necesita reglas propias.
- Se incorpora un deporte de sets, individual o con puntuación multidimensional.

## Documentación afectada

- [Producto](../project/PRODUCT.md)
- [Modelo inicial](../engineering/INITIAL_DATA_MODEL.md)
- [API](../engineering/API.md)
- [Sistema de diseño](../engineering/DESIGN_SYSTEM.md)
- [Observabilidad](../operations/OBSERVABILITY.md)
- [Decisiones](../governance/DECISIONS.md)
- [Aprendizaje](../project/LEARNING.md)
