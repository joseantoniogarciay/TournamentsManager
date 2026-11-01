# ADR-0178: Usar secciones abiertas y filas con separadores

- **Estado:** Aceptado
- **Fecha:** 2026-11-01
- **Decisor:** Usuario
- **Supera parcialmente:** ADR-0054 y ADR-0176, en el uso de tarjetas como
  contenedor de texto, formularios y elementos simples de una lista.

## Problema y evidencia

El usuario percibe que toda la información vive en cajas planas. Se comparan
tres conceptos de preparación de equipos: Lista abierta, Hoja de inscripción
y Recorrido guiado, manteniendo la identidad Grada. La repetición de una tarjeta
por equipo da el mismo peso visual a acciones, ayudas y elementos de la lista.
El gate técnico está cerrado.

## Criterios, alternativas y coste

- Lista abierta: jerarquía mediante texto, espacio y separadores; adopción
  moderada por revisión transversal, mantenimiento bajo con primitivas comunes.
- Hoja de inscripción: una superficie con profundidad; mantenimiento moderado
  por elevaciones y adaptación visual en cada plataforma.
- Recorrido guiado: tareas sobre un eje; mayor mantenimiento y riesgo de sugerir
  un orden obligatorio que el producto no exige.

## Recomendación y decisión explícita

Se recomienda Lista abierta por eliminar la repetición sin añadir interacción.
Tras presentarla como recomendación, el usuario responde: «Me gusta apliquemos
esto en todos los sitios donde sea evidente». Se registra la aceptación de
Lista abierta y su extensión a casos claros del mismo patrón.

## Alcance y consecuencias

- `ContentSection` organiza texto, ayudas y formularios sobre el canvas.
- `ListRow` separa equipos, administradores, notificaciones, sesiones y ediciones
  con una línea semántica, sin fondos, radios ni elevaciones individuales.
- Preparación y resumen de torneo, estados informativos, Inicio, ajustes y
  documentos legales usan secciones abiertas cuando solo agrupan contenido.
- La lista de equipos precede a la invitación en móvil. En web ancha las
  acciones pueden ocupar la columna lateral derecha, sin remontar el contenido.
- Las fichas independientes de torneos y encuentros, cuadros, diálogos y
  errores recuperables conservan sus superficies. No se cambia `Card` globalmente.
- Se conservan tokens Grada, controles, navegación nativa, URLs, localización,
  escalado, permisos, contrato y feedback. No se adopta el copy ficticio ni la
  tipografía de tamaño aumentado de la maqueta.
- Cada primitiva abierta reserva exactamente 20 px horizontales. El scroll y
  `Screen` no duplican ese margen; las columnas comparten la misma compensación
  que Card. Las filas tienen padding vertical y los grupos separación propia.

## Validación

Typecheck, lint, exportación web y regresiones del cliente. Revisar márgenes,
reflow, controles de 44 px y jerarquía de equipos, formularios y listas en claro,
oscuro y anchos estrechos/amplios. La compilación web no acredita QA nativo;
registrar por separado cualquier validación visual que no se haya ejecutado.
No requiere modificar operaciones OpenAPI ni iniciar backend u observabilidad.

## Disparadores de revisión

Una fila contiene tareas independientes que necesitan superficie propia; pérdida
de jerarquía, foco o legibilidad; problemas de reflow con nombres largos y letra
ampliada. No introducir tarjetas por defecto para resolver separación.

## Documentación afectada

AGENTS del cliente, DESIGN_SYSTEM, índice de ADR, DECISIONS, LEARNING y evidencia
de implementación con retrospectiva técnica.
