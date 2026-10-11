# ADR-0153: Mantener fijo el tamaño de texto de la tabla de clasificación

- **Estado:** Aceptado
- **Fecha:** 2026-10-09
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** Ninguno; excepción localizada al escalado de ADR-0054/0056.
- **Superado por:** Ninguno

## Problema y evidencia

En iPhone 17/iOS 27 a Text Size 11, las columnas de la clasificación aumentan
su altura de forma independiente y los valores dejan de coincidir con el equipo.
Un ensayo con alturas compartidas corrige la relación entre filas, pero las
columnas fijas siguen dejando un viewport de estadísticas ilegible.

## Contexto y restricciones

Conservar tabla, datos, reglas, Figtree, navegación, márgenes, cabecera sticky y
scroll horizontal. No cambiar HTTP, sesión, orden de equipos ni permisos.
La excepción solo afecta a las celdas de la tabla: encabezados, posición, nombre
de equipo, estadísticas y puntos. El título de pantalla, indicación de scroll,
selector de grupos y diálogo de reglas siguen respetando el tamaño del sistema.

## Alternativas y costes

### A — Escalado con desplazamiento de la tabla completa cuando no quepa

- Ventaja: todas las celdas se amplían y siguen siendo alcanzables.
- Inconveniente: equipo y puntos salen del viewport durante el gesto; la tabla
  crece y pierde densidad de comparación.
- Coste: bajo o medio; requiere una rama adaptativa y pruebas de cabecera,
  scroll y cambios de tamaño. Fue la recomendación inicial del asistente.

### B — Tamaño fijo únicamente en las celdas de la tabla

- Ventaja: conserva la tabla compacta, la correspondencia de filas y la
  navegación horizontal existente, con el menor cambio y mantenimiento.
- Inconveniente: la preferencia de texto ampliado no aumenta las celdas;
  limita su lectura visual para personas que necesitan esa ampliación.
- Coste: bajo; prop optativa en Text y aplicación explícita en las celdas.
- Mantener la semántica y los textos completos para lectores no certifica
  por sí solo la accesibilidad del recorrido; se conserva el pendiente de lector.

### C — Cards o ficha por equipo

Permite lectura lineal ampliada, pero reduce la comparación y añade otra
composición, etiquetas y pruebas. Coste mayor, sin necesidad aceptada.

## Recomendación y decisión del usuario

**Recomendación inicial:** A para conservar escalado. No constituye una regla
que prevalezca sobre la decisión explícita del usuario.

**Aceptada el 2026-10-09: B.** El usuario precisa: «La tabla de clasificación no
puede ser tan accesible, la dejamos usable pero que no crezca con texto porque
va a reventar». Esta instrucción sustituye la propuesta de tabla desplazable
completa y la interpretación anterior de «Continua». No se implementa la rama A.

Se usa allowFontScaling=false explícitamente en el texto de las celdas. La
primitiva compartida conserva el default nativo de escalado para todos los
consumidores que no opten por esta excepción. No se fija un máximo global ni se
cambia la configuración del sistema.

## Validación y revisión

Comprobar texto normal y máximo: mismos tamaños y correspondencia de filas,
scroll horizontal hasta GC/DG o columnas equivalentes, datos y puntos fijos,
apertura/cierre de reglas y escalado de los textos fuera de tabla. Restaurar
ajustes y apagar servicios. Typecheck, lint y exportación web obligatorios.
Android, web y lector requieren evidencia específica; la pasada iOS no los
certifica. Revisar la decisión ante problemas de lectura de las celdas o una
alternativa que conserve comparación sin el crecimiento rechazado.

## Documentación afectada

DESIGN_SYSTEM.md, PRODUCT_QA_REVIEW_2026-10-04.md, LEARNING.md e índices ADR.
