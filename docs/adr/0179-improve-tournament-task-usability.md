# ADR-0179: Reducir el recorrido de consulta y gestión de torneos

- **Estado:** Aceptado
- **Fecha:** 2026-11-02
- **Decisor:** Usuario: «De las de usabilidad hacemos todas».
- **Supera a:** Ninguno; amplía ADR-0054/0055/0178 sin sustituirlos.

## Problema y evidencia

Las capturas muestran metadatos y explicaciones extensas antes de los encuentros,
y dos acciones grandes repetidas por partido. La revisión de código encuentra
listas paginadas sin búsqueda y formularios de programación sin guardia de salida.
Es una evaluación heurística, no un estudio de usuarios ni una medición de impacto.
El gate técnico está cerrado.

## Alternativas, coste y recomendación

Mantener la composición cuesta menos, pero conserva los recorridos largos.
Rediseñar navegación o añadir otro framework eleva mantenimiento y rompe patrones
aceptados. Mejorar las tareas sobre primitivas existentes tiene coste moderado y
mantiene el producto reconocible. Se recomienda y acepta esta última alternativa.

## Decisión del usuario y alcance aceptado

1. Acceso al partido personal relevante y al primer resultado pendiente para
   administradores, sin alterar permisos ni abrir el editor automáticamente.
2. Resultado como acción principal; programación como secundaria visualmente
   discreta, siempre con semántica y objetivo táctil de 44 px.
3. Resumen de partidos resueltos/pendientes, con reglas explícitas para byes e
   incidencias y sin confundir pendientes con encuentros futuros indeterminados.
4. Guardia al salir con cambios sin guardar en programación y resultados;
   guardar o descartar limpiará el estado. Conservar el borrador de creación.
5. Explicación contextual de entradas rechazadas y foco en el primer campo
   inválido al guardar, manteniendo validación habitual por blur.
6. Búsqueda por nombre en biblioteca y filtros de partidos por equipo, jornada
   o estado. Las colecciones paginadas buscan en servidor, no solo en la página.
7. Resumen breve de la relación personal con el torneo y ayuda accesible para
   los permisos extensos; conservar explicación de controles no disponibles.

## Restricciones

Sin nueva librería, separación de negocio e infraestructura, contrato OpenAPI
para nuevas consultas, sin PII en observabilidad. Traducciones es/en/it/fr,
componentes compartidos, temas y accesibilidad según AGENTS del cliente.
No se reinterpretan estados deportivos ni se promete agenda o actualización
instantánea. Esta aceptación no autoriza publicación productiva.

## Validación

Rol creador/delegado/participante/lector; fase liga/grupos/cuadro/desempate,
partido relevante inexistente; filtros sin coincidencias, paginación y sesión
cambiada; conservación de posición y vuelta; guardado, descarte, fallo y gestos
nativos; teclado y lector, texto ampliado, cuatro idiomas, temas y anchos.
Contratos/errores/observabilidad para endpoints tocados, pruebas de regresión,
typecheck y exportación web. QA visual real declarada separadamente.

## Estado y retrospectiva

Alcance aceptado, implementación pendiente. Los ajustes previos de Lugar y máscara
numérica siguen en el workspace y no equivalen a completar este ADR. Priorizar
recorridos con datos existentes reduce riesgo frente a introducir todo a la vez.

## Documentación afectada

PRODUCT, DESIGN_SYSTEM, API/OBSERVABILITY cuando haya contratos nuevos,
DECISIONS, LEARNING y plan de incremento.
