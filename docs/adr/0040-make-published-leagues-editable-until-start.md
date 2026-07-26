# ADR-0040: Mantener la liga publicada editable hasta iniciarla

- **Estado:** Aceptado
- **Fecha:** 2026-07-26
- **Decisor:** Usuario
- **Propietario del análisis:** Codex
- **Supera a:** ADR-0032, exclusivamente en generación de partidos y edición tras
  publicar
- **Superado por:** Ninguno

## Problema

El flujo deseado permite compartir una liga en cuanto se crea, pero el creador
debe poder completar o ajustar su estructura antes de que empiece realmente. La
decisión anterior generaba los partidos y fijaba equipos al publicar, lo que
impide ese flujo.

## Contexto y restricciones

- El borrador se prepara localmente y puede asociarse temporalmente a una cuenta
  pendiente; no es aún una liga persistida operable (ADR-0031).
- Una liga publicada es no listada y se puede consultar por enlace (ADR-0033).
- El creador inicia la liga y los resultados solo se gestionan en `en_curso`
  (ADR-0038).
- El formato vigente sigue siendo una vuelta con regla 3-1-0. Esta decisión no
  añade doble vuelta ni nuevas reglas de clasificación.

## Criterios de decisión

1. compartir pronto sin bloquear ajustes previos al juego;
2. generar partidos únicamente cuando la composición esté lista;
3. separar borrador local, preparación publicada y competición en curso;
4. preservar la coherencia de los resultados ya existentes;
5. evitar diseñar ahora configuraciones deportivas futuras.

## Alternativas

### Alternativa A — Fijar y generar al publicar

- **Ventajas:** invariantes simples desde la publicación.
- **Inconvenientes:** obliga a crear una liga definitiva antes de poder revisarla
  o compartirla con el grupo.
- **Coste de adopción:** bajo.
- **Coste de mantenimiento:** bajo.
- **Riesgos:** fricción y necesidad de cancelar para cambios normales.

### Alternativa B — Publicada editable; fijar y generar al iniciar

La liga publicada se comparte y puede modificarse por el creador. Iniciarla
valida su composición, genera los partidos y congela los datos estructurales.

- **Ventajas:** encaja con preparación colaborativa ligera; el inicio tiene una
  semántica clara; no hay partidos obsoletos que regenerar.
- **Inconvenientes:** una liga publicada no tiene aún calendario de partidos.
- **Coste de adopción:** bajo a moderado.
- **Coste de mantenimiento:** bajo.
- **Riesgos:** mantener una liga publicada sin iniciar durante mucho tiempo.

### Alternativa C — Permitir cambios incluso en curso y regenerar partidos

- **Ventajas:** máxima flexibilidad.
- **Inconvenientes:** requiere resolver resultados afectados, recalcular tabla y
  explicar cambios al grupo.
- **Coste de adopción:** alto.
- **Coste de mantenimiento:** alto.
- **Riesgos:** inconsistencia deportiva y reglas especulativas.

### No cambiar

La creación seguiría confundiendo preparación con inicio de competición.

## Comparación

La A protege la consistencia a costa del flujo deseado. La C añade un motor de
replanificación innecesario. La B aplaza la congelación al límite natural —el
inicio— y mantiene los resultados libres de cambios estructurales.

## Recomendación

**Recomendación:** alternativa B, publicada editable hasta iniciar.

## Decisión del usuario

**Aceptada el 2026-07-26:**

- El borrador local o asociado a una cuenta pendiente no forma parte del ciclo de
  vida persistido de una liga; se puede descartar.
- Una liga persistida nace en `publicado`, es compartible por enlace y el creador
  puede modificar sus equipos y datos estructurales mientras no esté en curso.
- Al iniciar, el creador valida los datos requeridos, genera los partidos una
  sola vez y congela equipos y reglas aplicables.
- La liga puede cancelarse por el creador desde `publicado` o `en_curso`.
- `finalizado` sigue requiriendo todos los resultados y una acción explícita del
  creador.

La edición futura de reglas —por ejemplo, número de vueltas— se decidirá en un
ADR posterior. Mientras tanto, la única configuración admitida sigue siendo una
vuelta y puntuación 3-1-0.

## Consecuencias

### Positivas

- El enlace puede compartirse mientras el creador termina de preparar la liga.
- Los partidos solo reflejan una composición ya congelada.
- Los cambios normales no requieren cancelar y empezar otra liga.

### Negativas y deuda aceptada

- Una liga publicada no muestra partidos hasta ser iniciada.
- Hay que validar los datos otra vez al iniciar, no solo al publicar.
- Las ligas publicadas sin iniciar necesitan una política futura de recordatorio
  o limpieza si se convierten en un problema.

## Validación

- Un borrador no tiene enlace ni transición de cancelación.
- Una liga publicada puede consultarse por enlace y editarse solo por el creador.
- Iniciar requiere una liga válida y crea los partidos una vez.
- Tras iniciar, equipos y reglas no se modifican.
- Solo el creador cancela desde `publicado` o `en_curso`; cancelar conserva datos.

## Disparadores de revisión

- Necesidad de configurar ida y vuelta u otras reglas antes de iniciar.
- Necesidad de editar una liga con resultados.
- Acumulación o abuso de ligas publicadas que nunca se inician.

## Documentación afectada

- [PRODUCT.md](../project/PRODUCT.md)
- [ROADMAP.md](../project/ROADMAP.md)
- [LEARNING.md](../project/LEARNING.md)
- [DECISIONS.md](../governance/DECISIONS.md)
