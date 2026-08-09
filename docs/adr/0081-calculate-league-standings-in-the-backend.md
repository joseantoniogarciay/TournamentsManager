# ADR-0081: Calcular la clasificación de liga en el backend

- **Estado:** Aceptado
- **Fecha:** 2026-08-09
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

La aplicación debe mostrar una clasificación verificable tras cada resultado y,
en el futuro, acreditar torneos ganados en la cuenta o perfil sin permitir que
un cliente declare resultados o posiciones.

## Contexto y restricciones

- La liga inicial usa tres puntos por victoria, uno por empate y cero por derrota
  (ADR-0032 y ADR-0070), con una o dos vueltas elegidas antes de empezar.
- Los resultados se guardan y corrigen en el servidor (ADR-0035, ADR-0036,
  ADR-0037 y ADR-0080).
- El Reglamento de Competiciones de la RFEF vigente distingue el desempate según
  una o dos vueltas. Sirve de referencia, no convierte el producto en una
  competición federada.
- No se registran tarjetas, juego limpio, sorteos ni partidos de desempate.

## Criterios de decisión

1. La clasificación y una futura victoria deben ser verificables desde datos persistidos y reglas del dominio.
2. La tabla debe resultar reconocible para fútbol español sin introducir datos que el producto no recoge.
3. El cliente recibe una proyección para pintar, no una regla que pueda alterar.
4. La solución debe permanecer pequeña y revisable.

## Alternativas

### A — Calcular y ordenar en el cliente

- **Ventajas:** no amplía el contrato inicialmente.
- **Inconvenientes:** distintos clientes pueden discrepar y un cliente alterado podría presentar como ganada una liga que no lo está.
- **Coste de adopción:** bajo.
- **Coste de mantenimiento:** medio; duplica reglas entre plataformas y perfil.
- **Riesgos:** fuente de verdad falsa y resultados incoherentes.

### B — Proyección calculada por el backend

- **Ventajas:** una sola regla verificable; la web, móvil y futuros perfiles consumen el mismo resultado; facilita acreditar campeonas después.
- **Inconvenientes:** amplía la proyección de lectura y exige pruebas de dominio.
- **Coste de adopción:** moderado.
- **Coste de mantenimiento:** bajo; una regla pura y una única representación.
- **Riesgos:** cualquier extensión deportiva futura debe pasar por una nueva decisión.

### C — Persistir posiciones o títulos al registrar cada resultado

- **Ventajas:** lectura aparentemente inmediata.
- **Inconvenientes:** obliga a reconciliar correcciones, bajas y reintentos; persiste un derivado que puede recalcularse.
- **Coste de adopción:** alto.
- **Coste de mantenimiento:** alto.
- **Riesgos:** desincronización entre marcador, tabla y perfil.

### No cambiar

La pantalla de clasificación no puede ser una fuente fiable para la futura cuenta y se mantiene bloqueada.

## Comparación

La A es rápida pero no protege la integridad que requiere el perfil. La C añade estado derivado y casos de corrección prematuros. La B mantiene la autoridad en el dominio y es la mínima solución suficiente.

## Recomendación

**Opinión/recomendación:** alternativa B, con una proyección calculada en cada lectura y tras cada mutación de resultados.

## Decisión del usuario

**Aceptada el 2026-08-09:** la clasificación se calcula en backend y el cliente solo la presenta. Se aplican estos cuatro criterios, en orden:

1. puntos: victoria 3, empate 1, derrota 0;
2. con dos vueltas, mini-clasificación de los enfrentamientos entre equipos empatados (puntos, diferencia de goles y goles a favor);
3. con una vuelta, diferencia general y goles a favor generales antes de los enfrentamientos directos; con dos vueltas, esas métricas generales siguen a la mini-clasificación;
4. si permanece la igualdad, los equipos comparten posición. No se aplican juego limpio, sorteo ni partido adicional.

Los títulos o torneos ganados de una cuenta o perfil se derivarán en backend de la clasificación final; este ADR no crea aún ese perfil ni persiste títulos.

## Consecuencias

### Positivas

- La tabla se actualiza tras cada resultado o corrección desde la misma fuente.
- Un perfil futuro puede acreditar solo campeonas derivadas de ligas finalizadas.
- La interfaz explica reglas reales sin tener capacidad de modificarlas.

### Negativas y deuda aceptada

- El empate múltiple usa la mini-clasificación disponible y puede compartir posición cuando los datos actuales no resuelven la igualdad.
- No hay reglas de sanciones, tarjetas o desempate presencial.
- **Medición pendiente:** antes de materializar la clasificación se medirá el cálculo de lectura y tras una corrección con 64 equipos y dos vueltas (4.032 partidos), incluyendo percentiles de latencia y consultas de base de datos.

## Validación

- Cambiar un marcador recalcula puntos, goles, posición y desempates en backend.
- La API devuelve filas calculadas y el cliente no recalcula ni ordena la tabla.
- Dos equipos totalmente iguales reciben la misma posición.
- Empates triples y cuádruples se evalúan contra una única mini-clasificación del
  grupo completo y no mediante comparaciones por parejas.
- Una futura consulta de perfil obtiene victorias solo de esta proyección final o de una derivación equivalente del mismo dominio.

## Disparadores de revisión

- Se añaden sanciones, juego limpio, premios o un requisito federativo.
- Se necesitan desempates oficiales que no puedan expresarse con los datos actuales.
- El perfil requiere historial materializado por motivos medidos de rendimiento.

## Documentación afectada

- [Producto](../project/PRODUCT.md)
- [API](../engineering/API.md)
- [OpenAPI v1](../../contracts/openapi/v1/openapi.yaml)
- [Decisiones](../governance/DECISIONS.md)
- [Aprendizaje](../project/LEARNING.md)
