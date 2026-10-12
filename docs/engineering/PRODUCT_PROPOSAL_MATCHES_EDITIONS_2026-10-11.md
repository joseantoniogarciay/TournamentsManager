# Propuesta de producto: Mis partidos y Ediciones

Fecha: 2026-10-11. Estado: diseño aprobado; implementación según ADR-0161/0162.
La maqueta documenta la revisión previa y no sustituye las pruebas del cliente real.

## Recorridos

Mis partidos → Pendientes/Resultados → encuentro → detalle del torneo situado
en la fase, ronda y partido correspondientes → volver a la misma lista.

Detalle → Ediciones (solo con dos o más) → colección compartida de la familia
→ ficha de otra edición → Ediciones, con la nueva edición identificada.

La maqueta interactiva de conversación usa datos ficticios. Permite cambiar
estados iniciales, vacío por vista, error con reintento y error de actualización.
Pulsar un partido abre el destino resaltado; Ediciones permite recorrer fichas.
No representa QA funcional ni llamadas al contrato de producto.

## Composición visual propuesta

Cuarta tab Mis partidos antes de Cuenta. Cabecera y dos controles Pendientes /
Resultados; tarjetas existentes con torneo, equipos y programación/resultado.
Carga compartida centrada; error inicial con tarjeta y acción de reintento;
vacío sin tarjeta falsa de partido. Fallo de actualización conserva contenido.
20 px exteriores y entre cards; objetivos de 44 px, margen de botonera,
anchos web y adaptación nativa según AGENTS.md y DESIGN_SYSTEM.md.

El detalle mantiene su estructura y añade Ediciones junto a sus acciones
existentes cuando la relación esté confirmada. La lista de ediciones reutiliza
cards, estados deportivos y navegación canónica; marca Edición consultada.
La maqueta es esquemática: no sustituye los controles nativos de iOS/Android.

### Correcciones de revisión

El usuario concreta pull to refresh en apps para Mis partidos y Ediciones,
también en listas vacías; se retira el botón de actualización de cabecera de la
maqueta móvil. La revisión permite arrastrar hacia abajo la lista con dedo o
ratón. El error inicial mantiene Reintentar y carga la colección de esa pantalla.
Los escenarios Una edición, Varias ediciones y Error al consultar ediciones
abren directamente su pantalla: detalle sin botón, lista de tres ediciones y
error de esa lista respectivamente. Reintentar en Ediciones muestra su carga y
después sus ediciones, sin cambiar a Mis partidos. Las fichas históricas muestran
resultados ficticios propios de su edición, no el encuentro pendiente actual.

Revisión comprobada en Chrome headless, viewport 390×844: Una edición sin botón;
Varias ediciones con tres entradas; Error al consultar ediciones → Cargando
ediciones → lista de la misma colección; ficha histórica sin partido pendiente
de la edición actual; arrastre hacia abajo → Actualizando → contenido conservado
y recuperación del banner de error. Sin errores JavaScript. Inspeccionadas
visualmente la lista y la pantalla de error. El gesto se verifica con ratón en
la maqueta, no en dispositivos nativos; QA de RefreshControl iOS/Android futura.

## Reglas que requieren revisión

- Pendientes excluye byes, rivales todavía indeterminados, equipos retirados
  y torneos cancelados; Resultados conserva hechos resueltos, incluidas incidencias.
- Orden de pendientes por fecha del partido, manteniendo su zona; sin fecha al
  final. No se presenta como agenda de instantes convertidos a la zona del lector.
- Ediciones agrupa toda la familia, incluso repeticiones desde ediciones antiguas.
- Conocer una edición permite recorrer las demás mediante la lectura pública
  vigente. No introduce una audiencia privada para la familia.
- Reutilizar procedencia persistida antes de introducir una entidad Serie.

## Matriz mínima de revisión y QA futura

| Recorrido | Casos |
| --- | --- |
| Mis partidos | Sin sesión; carga; ninguna participación/encuentro; vacío por vista |
| Recuperación | Error inicial/reintento; actualización fallida; paginación si existe |
| Datos | Fecha/hora incompletas; zonas distintas; incidencia; resultado administrativo |
| Navegación | Liga, grupos, cuadro y desempate; cambio de fase; partido desaparecido |
| Regreso | Vista y posición conservadas; enlace en frío; historial web |
| Ediciones | Una/dos/muchas; cadena y ramas; nombres iguales sin vínculo; canceladas |
| Interfaz | Claro/oscuro; texto ampliado; lector; movimiento reducido; cuatro locales |
| Seguridad | Sesión cambiada; permisos independientes; errores seguros; cancelación |

## Retrospectiva de preparación

Comprobación de la maqueta en Chrome headless: abrir encuentro resaltado,
recorrer tres ediciones y destino elegido, error inicial con reintento, vacío
de Pendientes con Resultados disponible, fallo de actualización conservando
filas y botón Ediciones ausente con una sola edición. Sin errores JavaScript
en esa muestra. Inspección visual de la lista en viewport 460×850. Enlaces
locales de los documentos y git diff --check correctos. Estas comprobaciones
son de la maqueta, no del producto, backend, iOS ni Android.

Un estado vacío requiere evidencia de una consulta correcta, no ausencia de filas
por un error. Participar y administrar son relaciones diferentes. La procedencia
de repetición ya existe en datos: inspeccionar antes de proponer otra estructura
reduce mantenimiento y permite incluir repeticiones anteriores sin adivinarlas.
Siguiente gate: aceptación de diseños y ADR; después contrato, implementación,
pruebas y documentación de cierre.

Referencias: [ADR-0161](../adr/0161-add-personal-matches-tab-and-match-deep-links.md)
y [ADR-0162](../adr/0162-browse-related-tournament-editions.md).
