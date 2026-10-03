# ADR-0141: Soportar bádminton con juegos por puntos

- **Estado:** Aceptado
- **Fecha:** 2026-10-03
- **Decisor:** Usuario
- **Supera a:** ADR-0135 y ADR-0136, solo para ampliar perfiles de raqueta

## Problema y autorización

El usuario ha solicitado incorporar bádminton y después evaluar mejoras de los
deportes existentes. Ha elegido explícitamente permitir 21 o 15 puntos por
torneo sobre el alcance inicial propuesto. La decisión autoriza este incremento.

## Contexto y evidencia

El gate técnico está cerrado. ADR-0135 y ADR-0136 ofrecen participantes,
eliminatoria directa, juegos ordenados, agregado derivado e historial atómico.
La estructura sirve para bádminton, pero el validador de tenis de mesa no:
bádminton tiene un tope deportivo que puede permitir terminar por un punto.

El formato internacional actual es al mejor de tres juegos a 21 puntos. Badminton
Europe y la federación japonesa confirman el cambio a tres juegos a 15 puntos
desde el 4 de enero de 2027. No debe cambiarse el reglamento de un torneo ya
creado por el mero paso del tiempo.

## Alternativas y coste

1. **Solo el formato actual a 21 puntos.** Adopción media y mantenimiento bajo;
   reutiliza la configuración de tres juegos. Requiere otra decisión para
   incorporar posteriormente el formato a 15 puntos.
2. **Elección entre 21 y 15 puntos por torneo.** Adopción y mantenimiento mayores:
   exige persistir la elección, ampliar contrato y borradores, localizar el
   selector y verificar ambos perfiles y sus topes. Evita otra ampliación cercana
   y permite conservar el reglamento de cada torneo.
3. **Motor configurable de puntuación.** Coste alto; combinaciones, contrato y UI
   sin demanda. No se recomienda para dos perfiles cerrados.

No añadir el deporte ahora reduciría trabajo, pero no satisface la solicitud.

## Recomendación de alcance inicial

La recomendación inicial fue la alternativa 1 por coste. El usuario eligió la
alternativa 2 para cubrir ambos perfiles; se conserva un catálogo cerrado.

- Perfil explícito `badminton`, reglas en dominio y sin nuevas dependencias.
- Eliminatoria directa, con personas o parejas representadas por un nombre.
- Partidos al mejor de tres juegos; victoria exactamente al ganar dos.
- Elección de `pointsPerGame: 21 | 15` al crear, con 21 por defecto en cliente.
  Obligatoria en contrato y persistencia para bádminton; se omite en otros deportes.
- Juego a 21: diferencia de dos hasta el tope de 30; 30–29 es válido.
  Juego a 15: diferencia de dos hasta el tope de 21; 21–20 es válido.
  No se admiten juegos parciales ni posteriores a la victoria.
- Agregado derivado de los juegos, corrección atómica, historial y propagación
  de ganadora con las restricciones existentes sobre resultados posteriores.
- Vocabulario de participantes y juegos en los cuatro idiomas.
- Ligas, mixto, abandonos e incomparecencias requieren políticas posteriores;
  no se promete soporte completo de todo reglamento de competición.

La elección se conserva durante todo el ciclo del torneo, incluidos borradores
de registro y acceso federado. No se altera automáticamente por fecha ni se
introduce un motor configurable. Se valida también en las respuestas cliente.

## Decisión del usuario

**Aceptada el 2026-10-03:** «Dejamos eleccion entre 21 y 15». Se implementa la
alternativa 2 con eliminatoria directa y personas o parejas, al mejor de tres
juegos, conforme al alcance presentado.

## Validación y cierre

- Dominio: puntuación normal, deuce, tope, ambos lados, negativos, empates,
  resultados incompletos, victoria anticipada y formatos incompatibles.
- PostgreSQL: restricciones, creación, lectura, resultados, correcciones,
  instantáneas y propagación del cuadro.
- HTTP: creación y borradores consistentes; recorrer éxito, validación, tasa,
  rechazo de negocio, fallos técnicos y cancelación con salida segura.
- Cliente: contrato generado, borrador, creación, consulta y registro de juegos;
  traducciones, typecheck, exportación web y revisión visual responsive.
- Actualizar producto, API, modelo, observabilidad, aprendizaje y retrospectiva.

Completado en código y migración incremental 00018. Suite Go, integración
PostgreSQL desechable, lint Go y cliente, typecheck, contrato y exportación web
aprobados. Generación OpenAPI idempotente. Revisión visual del selector a
390×844 y del JSX real del resultado en ruta efímera a 390×844 y 320×480: ambos
topes habilitan Guardar, su exceso lo bloquea y el scroll permite alcanzar la
acción. La ruta de prueba se retiró antes de exportar. No se desplegó ni se
migraron entornos persistentes. Queda pendiente la comprobación del teclado
en iOS/Android reales; la advertencia de compatibilidad Expo ya aplazada en
ADR-0138 se mantiene sin cambiar dependencias.

## Fuentes

- [Liga oficial Pays de la Loire: 15 puntos y diferencia de dos hasta 20–20](https://www.badminton-paysdelaloire.fr/scoring-match-badminton/)

- [Badminton Europe: cambio a 3×15 desde el 4 de enero de 2027](https://cms-prod.badmintoneurope.com/en/web/guest/w/new-scoring-system-3-x-15)
- [Federación japonesa: aprobación y fecha de entrada](https://www.badminton.or.jp/news/detail/2156)
- [Comité Olímpico de Irlanda: tanteo actual y tope de 30](https://olympics.ie/sport/badminton/)

## Retrospectiva del análisis y cierre

Reutilizar la estructura de juegos reduce el cambio, pero no sustituye revisar
las reglas. El cambio reglamentario cercano exige decidir la configuración antes
de programar; elegirla por fecha sin guardarla alteraría torneos históricos.
Tras cerrar bádminton se evaluarán mejoras concretas de los deportes existentes,
sin introducir un motor genérico ni aceptar todavía nuevas políticas.

## Evaluación posterior de mejoras (sin decisión funcional)

El catálogo cubre ocho deportes. Esa cobertura no implica completar todas las
modalidades o incidencias de cada reglamento. La siguiente comparación es una
recomendación de producto, no autoriza implementar sus políticas:

| Alternativa | Valor | Coste y mantenimiento |
| --- | --- | --- |
| Incomparecencia y abandono por partido | Permite cerrar partidos sin inventar tanteos | Medio: distinguir causas, ganador, efecto en fases y auditoría por deporte |
| Ligas de raqueta | Amplía torneos recurrentes con participantes existentes | Mayor: clasificación, desempates, ratios e incidencias para cada perfil |
| Horarios y pistas | Ayuda a organizar la jornada | Medio/alto: colisiones, reprogramación y alcance del calendario |

Recomendación: priorizar las incidencias por partido; primero definir su
significado y efecto por deporte, evitando generalizar la retirada actual de
un equipo en liga. Después evaluar ligas de raqueta según demanda. Antes de
cualquiera de estas ampliaciones se necesita decisión explícita y ADR aceptado.
