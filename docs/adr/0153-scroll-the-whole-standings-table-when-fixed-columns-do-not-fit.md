# ADR-0153: Desplazar la tabla completa cuando las columnas fijas no dejen espacio legible

- **Estado:** Propuesto
- **Fecha:** 2026-10-09
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** Ninguno; concreta la adaptación accesible de ADR-0054/0055.
- **Superado por:** Ninguno

## Problema y evidencia

La clasificación divide cada fila entre equipos, estadísticas desplazables y
puntos. En iPhone 17/iOS 27 a Text Size 11, los anchos fijos originales provocan
saltos de línea y alturas distintas: los valores dejan de coincidir con el equipo.
Compartir alturas medidas corrige esa relación. Ampliar las celdas conserva el
escalado, pero las dos columnas fijas agotan el ancho disponible y dejan un
viewport de estadísticas ilegible. La prueba de esta sesión no pasa QA.

## Contexto y restricciones

Mantener la tabla, sus datos y reglas, Figtree, escalado completo, tokens,
márgenes, navegación y botones nativos. A tamaño normal se conserva la tabla
actual con equipos y puntos fijos. No cambiar HTTP, ordenar datos de otra forma,
introducir librerías ni ocultar columnas. El baseline técnico está cerrado.

## Alternativas y costes

### A — Desplazar la tabla completa solo cuando no quepa

- Comparar el ancho disponible con las columnas fijas y al menos una columna
  estadística completa. La activación depende del espacio, no de un umbral
  arbitrario de fontScale.
- En esa rama, equipo, estadísticas y puntos participan en un único recorrido
  horizontal; la cabecera acompaña a sus columnas y cada fila comparte altura.
- Ventaja: conserva la tabla y permite leer todas sus celdas con texto ampliado.
- Inconveniente: equipo y puntos pueden salir del viewport durante el gesto.
- Coste de adopción y mantenimiento: bajo o medio, con una rama adaptativa y
  pruebas de ambas direcciones, cambios de tamaño y cabecera.

### B — Sustituir la tabla ampliada por cards o una ficha por equipo

- Ventaja: permite lectura lineal de nombres y valores.
- Inconveniente: dificulta comparar equipos y añade una composición y recorrido.
- Coste: mayor; dos vistas, etiquetas y pruebas de paridad que mantener.

### No cambiar

Conservar columnas fijas ilegibles o desactivar escalado no resuelve la
accesibilidad. La alineación por sí sola no constituye un cierre válido.

## Recomendación

**Opinión/recomendación:** A. Corrige el límite de espacio sin sustituir la
presentación existente y solo cambia el gesto cuando resulte necesario.

## Decisión del usuario

**Pendiente:** se ha presentado la alternativa A al usuario en esta sesión.
No implementar esta rama mientras el ADR siga Propuesto.

## Validación y revisión

Comprobar tamaño normal y máximo, scroll horizontal hasta todas las columnas,
relación entre equipos y valores, cabecera, acceso al final vertical, restauración
del tamaño y cierre. Completar después Android, web y lector sin atribuirles la
evidencia iOS. Typecheck, lint y exportación web obligatorios.
Revisar si la tabla completa pierde contexto o aparece una mejor disposición
con evidencia que justifique el coste de mantenerla.

## Documentación afectada

DESIGN_SYSTEM.md, PRODUCT_QA_REVIEW_2026-10-04.md, LEARNING.md e índices ADR.
