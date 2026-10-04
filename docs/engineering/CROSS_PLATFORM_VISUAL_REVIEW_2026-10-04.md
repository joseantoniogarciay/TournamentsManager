# Coherencia visual entre plataformas — 2026-10-04

> Estado: investigación de código y cobertura previa. No acredita una auditoría
> completa de renderizado. Referencia de código: `645b55b`.

## Problema y marco aceptado

El usuario solicita investigar la homogeneidad entre sistemas operativos,
separando sus características propias y comprobando que el contenido se pinta
correctamente. ADR-0008 admite una interfaz adaptativa con paridad funcional;
ADR-0054, ADR-0055, ADR-0056 y el sistema de diseño definen la base común.
Figtree procede de ADR-0096; el idioma automático, de ADR-0143.

La coherencia se evalúa por márgenes, jerarquía, legibilidad, estados y acciones
alcanzables. Las capturas de dos dispositivos con distinta densidad no se
comparan directamente en píxeles físicos. Deben registrar tamaño lógico, escala
de texto, tema, idioma y estado; se comparan distancias en unidades de layout y
comportamientos equivalentes.

## Inventario de lo común y lo adaptativo

| Área | Base compartida comprobada en código | Adaptación existente que debe conservarse |
| --- | --- | --- |
| Contenido | `Card`: 20 px exteriores; `Screen`: 12 px superiores; tokens semánticos | Insets reales del dispositivo y cabecera nativa, sin sumar dos veces el área segura |
| Tipografía y campos | Figtree, variantes de `Text`, `TextField`, validación y feedback | Renderizado de texto, selección, cursor y teclado nativos |
| Botones | Primitiva `Button`, mínimo de 44 px, estados busy/disabled | Indicador de actividad y respuesta táctil nativos |
| Cabeceras | Intención de cierre/atrás, títulos y destino de retorno | Toolbar y Liquid Glass en iOS reciente; control compartido en Android y versiones anteriores |
| Tabs | Tres destinos, selección y colores del tema resuelto | SF Symbols/Material; material iOS reciente; apariencia explícita Android; `Tabs` web |
| Formularios | `KeyboardAwareScrollView` mantiene visible el foco | Insets automáticos iOS; medición de oclusión restante tras resize Android |
| Diálogos | `ModalDialog`, superficie, cierre y host dentro de `Screen` | Blur nativo, capa de oscurecimiento Android y scrim web |
| Acceso social | Dos círculos de 48 px, gap de 16 px y etiquetas accesibles | iOS Apple → Google; Android Google → Apple; web Apple → Google |

Fuentes principales: `shared/ui`, `(tabs)/_layout.tsx`, su variante web y la ruta
de Cuenta. La adaptación se concentra mayormente en primitivas y navegación;
las rutas todavía repiten la selección de toolbar según `usesLiquidGlassNavigation`.
Esa repetición se registra como punto de revisión, sin justificar por sí sola
otra abstracción ni duplicar pantallas por SO.

## Evidencia disponible y huecos

Las auditorías [iOS](IOS_VISUAL_AUDIT_2026-10-03.md),
[Android](ANDROID_VISUAL_AUDIT_2026-10-03.md) y
[web](WEB_VISUAL_AUDIT_2026-10-03.md) son registros históricos con alcances
distintos. Los estados de sus hallazgos deben contrastarse con el código actual;
una recomendación antigua no se copia como defecto vigente sin reproducirla.

- iOS 27 y Android 14: recorridos iniciales sin sesión, tema y Crear torneo;
  regresión del foco con teclado software comprobada en ambos.
- iOS 18.5: arranque y apertura/cierre de Crear torneo; no acredita la regresión
  posterior del teclado ni todas las adaptaciones de cabecera.
- Cierre Android: margen corregido tras comprobar la contribución de la toolbar.
  La compensación de 16 px es evidencia de la build probada, no una constante
  universal certificada para cualquier toolbar o versión futura.
- Acceso social: el cambio `645b55b` se observó en iPhone 18 Pro/iOS 27 y
  Pixel API 34/Android 14, con el orden distinto confirmado por el usuario.
- Web: recorrido anterior con estados ficticios, tamaños pequeños y nombres
  largos; no prueba la API ni equivale a una comprobación nativa autenticada.
- Pendientes comunes: texto ampliado, orientación, lectores de pantalla reales,
  variantes de idioma y recorridos nativos con sesión. Release y dispositivos
  físicos tienen cobertura separada de las development builds.

## Casos prioritarios para continuar la revisión

Estas son hipótesis de riesgo derivadas del código, no defectos confirmados.

| Prioridad | Caso comparable | Motivo y criterio de aceptación |
| --- | --- | --- |
| 1 | Cabecera con título largo y controles a ambos lados | Revisar todas las rutas que eligen toolbar, incluidos ajustes, invitación, participantes y transferencia. Sin superposición; margen exterior correcto y cierre operable. Contrastar también iOS anterior a Liquid Glass. |
| 1 | Último campo y acción final con teclado abierto | El cálculo depende de medidas y eventos nativos. Probar registro, creación y búsquedas con autoFocus; foco completo, envío alcanzable y espacio temporal eliminado al cerrar teclado. |
| 1 | Botón con traducción larga, texto ampliado y loader | `Button` compone texto y loader en una fila sin una regla explícita de contracción del texto. Verificar ancho pequeño y todos los idiomas; contenido dentro del control y acción legible. No ajustar estilos sin reproducir. |
| 1 | Texto y TextInput con la misma escala | `Text` define lineHeight; TextInput mantiene métricas nativas y altura mínima. Comparar baseline, acentos, placeholder y texto multilínea; sin recorte ni pérdida del error inline. |
| 1 | Modal sobre ruta fullScreenModal | Confirmación visible sobre la ruta activa, sin quedar detrás; texto largo y teclado sin ocultar acciones. |
| 2 | Tabs y último elemento de listas | `useTabContentBottomPadding` usa inset más reserva fija. Probar poco/mucho contenido y navegación por gestos/botones Android; último elemento y acción flotante alcanzables. |
| 2 | Tema forzado contrario al SO | Comprobar contenido, cabeceras, tabs, diálogos y transiciones en ambos sentidos; sin destello ni controles con contraste perdido. |
| 2 | Torneos autenticados con datos extremos válidos | Nombres largos, clasificación, cuadro, resultados y menús; ninguna acción inaccesible ni columnas cortadas sin scroll disponible. |

## Protocolo de comparación y cierre

1. Identificar SHA del bundle y build nativa, SO, tamaño lógico, tema, idioma,
   escala de texto, teclado software y precondiciones. No comparar un bundle
   antiguo con uno actualizado.
2. Recorrer primero Inicio, Torneos sin sesión, Cuenta, Registro, Ajustes y Crear
   torneo, sin envíos. Repetir con claro/oscuro y un tamaño de texto ampliado.
3. Recorrer después pantallas autenticadas con los mismos fixtures de prueba,
   nunca interpretando los datos ficticios como prueba end-to-end de la API.
4. Para cada diferencia, registrar esperado, observado y captura: adaptación
   prevista, defecto reproducido o caso pendiente. Una lectura de estilos sola
   no obtiene el estado «visualmente comprobado».
5. Corregir el mínimo compartido cuando el fallo pertenezca al contenido;
   mantener la corrección en el adaptador cuando proceda del host nativo.
6. Repetir el caso en los dos SO y exportar web si cambia el cliente compartido.
   Cerrar con documentación, servicios de pruebas apagados y Git sincronizado.

## Alternativas y recomendación

Una revisión manual con matriz y capturas tiene coste inicial bajo y permite
descubrir qué estados son realmente sensibles. Una suite de regresión visual
por plataforma añade mantenimiento de builds, fixtures y baselines; conviene
evaluarla después, para casos estables y repetidos, conforme a ADR-0019.
Igualar cada píxel de los controles nativos duplicaría trabajo y contradiría la
navegación adaptativa aceptada.

Recomendación: continuar con la matriz manual y casos prioritarios sobre las
primitivas actuales. Esta investigación no incorpora herramientas, dependencias
ni una nueva decisión de arquitectura.

Retrospectiva: compartir componentes reduce divergencias, pero no acredita su
renderizado. La cobertura debe vincularse al estado y versión observados; separar
contenido común y host nativo ayuda a corregir la causa sin perder las convenciones
del SO.
