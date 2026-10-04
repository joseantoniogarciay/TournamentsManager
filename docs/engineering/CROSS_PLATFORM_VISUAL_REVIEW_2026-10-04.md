# Coherencia visual entre plataformas — 2026-10-04

> Estado: investigación de código y revisión manual nativa sin sesión.
> No acredita una auditoría completa de todas las rutas y configuraciones.
> Bundle inicial de la revisión manual: `9ba3351`; corrección de iconos posterior
> incluida en el mismo cierre que este registro.

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

## Ejecución manual autorizada — 2026-10-04

Development builds `com.fasttourney.app.local` ya instaladas: iPhone 18 Pro con
iOS 27 y Pixel API 34 con Android 14. Ambos cargan el mismo código mediante
Metro LAN en 8082, con español y orientación vertical. Se revisa sin sesión,
sin activar backend ni observabilidad, sin enviar formularios ni aceptar términos.
Las capturas y árboles accesibles se observaron en la conversación; no se
conserva un archivo de captura independiente en el repositorio. La evidencia
acredita estos estados concretos, no todos los estados posibles de cada ruta.

| Recorrido | iOS 27 | Android 14 |
| --- | --- | --- |
| Inicio, claro y texto habitual | Cards, copy y CTA legibles | Cards, copy y CTA legibles |
| Torneos sin sesión, claro y texto habitual | Estado vacío y acción flotante sobre tabs | Estado vacío y acción flotante sobre tabs |
| Crear torneo, claro y texto habitual | Cierre y card alineados; opciones y campos visibles | Cierre y card alineados; opciones y campos visibles |
| Cuenta, claro y texto habitual | Apple → Google, círculos y etiquetas accesibles | Google → Apple, círculos visibles |
| Registro, claro y texto habitual | Campos, textos legales y CTA deshabilitada sin solaparse | Campos, textos legales y CTA deshabilitada sin solaparse |
| Último campo de Crear torneo con teclado software | Foco visible; al intentar desplazar se cerró el teclado, por lo que no se acredita CTA con teclado aún abierto | Foco visible; al desplazar, CTA visible sobre Gboard; al cerrar, layout recuperado |
| Registro con teclado software | Pendiente | Contraseña visible, bloque legal y CTA deshabilitada sobre Gboard |
| Texto ampliado | Dynamic Type al extremo superior del rango habitual, tamaños de accesibilidad adicionales desactivados; Registro legible en claro; Inicio, Torneos, Cuenta, Ajustes y Crear torneo legibles en oscuro | Tamaño de fuente al extremo superior del slider, tamaño de visualización original; Inicio y Crear torneo legibles en claro; CTA larga alcanzable al desplazar |
| Contenido bajo tabs con texto ampliado | Crear cuenta se desplaza por encima de las tabs | Pendiente en Cuenta ampliada |
| Tema forzado contrario al SO | Oscuro en app con SO claro: superficies, cabeceras y tabs coherentes; retorno a Sistema comprobado | Pendiente en esta ejecución |

### Defecto confirmado y corrección mínima

Con el tamaño de fuente máximo de Android, la X de Crear torneo quedaba
recortada dentro de su caja y parecía una flecha diagonal. La implementación
instalada de `expo-symbols` 57.0.2 usa un `Text` con fontSize/lineHeight iguales
al tamaño de una `View` fija, pero permite por defecto el escalado de texto.
El glifo crecía mientras su caja no lo hacía. Esto explica la divergencia frente
al SF Symbol nativo de iOS.

Se registra un parche pnpm en fuente y build JavaScript que desactiva
`allowFontScaling` únicamente para el glifo Material. Mantiene el token de
icono, el objetivo táctil y las etiquetas localizadas. No se reduce el tamaño
de los textos ni se duplican componentes por pantalla. Una compensación local
por fontScale sería menos fiable con el escalado no lineal Android; copiar otra
implementación de iconos añadiría mantenimiento y riesgo de divergencia.

### Cobertura que sigue abierta

No se han verificado en esta ejecución: sesión y fixtures autenticados,
títulos de entidades largos, loader sobre traducciones largas, contenido
TextInput con acentos a escala máxima, diálogo sobre fullScreenModal, otros
idiomas, orientación horizontal, rango extra de accesibilidad iOS, VoiceOver/
TalkBack, iOS anterior a Liquid Glass, otras modalidades de navegación Android,
build Release y dispositivos físicos. No interpretar esta revisión sin sesión
como cierre de todos los criterios de los ADR ni prueba end-to-end de la API.

Retrospectiva: probar texto ampliado aportó un defecto que la revisión estática
y el tamaño habitual no revelaban. El tamaño lógico del control y la escala del
glifo deben contrastarse juntos. Un parche de dos líneas tiene coste menor que
un adaptador nuevo, pero exige revisión y retirada al actualizar la dependencia.


### Cierre y límite de verificación

Pasan typecheck, exportación web y las cinco pruebas de compatibilidad de
dependencias. El parche se aplica reproduciblemente mediante pnpm y ambos
bundles nativos se regeneraron tras reiniciar Metro con caché vacía.
Tras desbloquear el Mac, se cargó el bundle corregido en Pixel API 34:
con fuente máxima, la X completa queda centrada en el control circular y su
borde exterior mantiene la alineación con la card. El texto del formulario
conserva su escala ampliada. Se confirma visualmente la regresión corregida.

Se restauraron y verificaron los ajustes originales: iOS Dynamic Type al 50 %
y tamaños adicionales desactivados; Android fuente en posición 2 de 7 y tamaño
de visualización conservado en posición 2 de 5. La restauración se completó con
los menús de accesibilidad de Device Hub y Device UI Shortcuts de Android Studio,
tras respuestas irregulares de los gestos sobre los sliders. La captura de
teclado y el zoom habitual de Device Hub quedaron restaurados.

Crear torneo en iOS, después de recargar y restaurar el tamaño, conserva la X
nativa, los márgenes y la CTA «Inicia sesión para crearlo». Android recupera
Inicio y Crear torneo con su tipografía habitual, X completa y CTA sin recorte. Tema iOS restaurado a Sistema; tema Android
sin cambios. La revisión sigue limitada a la cobertura sin sesión descrita en
la matriz, sin cerrar los casos abiertos de los ADR.

Metro temporal apagado al cerrar; no se han arrancado backend, dev ni
observabilidad.

### Ampliación con sesión real local

Por petición del usuario se crearon y verificaron dos cuentas ficticias en la
API local: `qa_visual_owner@example.test` y `qa_visual_player@example.test`.
Se usa registro, correo capturado por Mailpit y verificación reales. No hay
login hardcodeado, bypass de autorización ni cuentas en producción. Las
contraseñas aleatorias y sesiones viven exclusivamente en
`apps/client/.env.qa-review.json`, ignorado por Git y con permisos `0600`.
No copiar ese archivo a documentación, bundles ni variables `EXPO_PUBLIC_*`.
Las cuentas permanecen en el volumen local al detener Compose.

En una base local nueva, aplicar primero `make db-schema-apply` y después
`make dev-migrate`; ejecutar migraciones sobre una base sin esquema inicial
falla por ausencia de `accounts`. Esto no autoriza reinicializar una base existente.

Se verificaron 17 respuestas HTTP esperadas: login y sesión de ambas cuentas,
métodos de acceso, creación, invitación e inspección pública, inscripción del
participante, rechazo de una segunda inscripción (409), seguimiento y biblioteca,
lectura de administradores y notificaciones, rechazo del inicio por participante
(403) e inicio por organizador. No se modifica contrato, endpoint ni adaptadores
cliente con esta revisión. Esta muestra no sustituye la suite de salidas de
observabilidad ni cubre todos los errores técnicos de cada endpoint.

Fixture local: `01a10743-eafc-713f-9dc4-c8660ea5581f`, «QA Liga de equipos con
acentos y nombres largos», fútbol/liga en curso, tres equipos. El participante
está vinculado a «Peñas del Mediterráneo». No tiene administración delegada.

| Recorrido visual                             | iOS 27 / iPhone 18 Pro                                                                        | Android 14 / Pixel API 34                                                          |
| -------------------------------------------- | --------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------- |
| Login real y cuenta del organizador          | Acceso confirmado; Datos de acceso muestra el correo                                          | Acceso confirmado; nombre, campana y Datos de acceso visibles                      |
| Biblioteca Administro                        | Título largo en dos líneas y relación Creador                                                 | Título largo en dos líneas y relación Creador                                      |
| Detalle de liga en curso                     | Permisos, equipos y resultados presentes                                                      | Cabecera de dos líneas, card y partidos sin desbordamiento                         |
| Equipos con acentos y nombres largos         | Filas de dos líneas y retirada visible                                                        | Filas de dos líneas y retirada visible                                             |
| Diálogo de resultado sobre pantalla completa | Centrado con fondo desenfocado; guardado 2–1 confirmado                                       | Centrado con oscurecimiento; teclado no oculta campos/CTA; guardado 3–0 confirmado |
| Clasificación                                | Orden y tres puntos del primer resultado confirmados; nombres truncados en columnas estrechas | Orden y puntos de los dos resultados confirmados; nombres truncados en columnas estrechas                                                |

Se detectó que `DisclosureIndicator` enviaba exclusivamente `chevron.right`:
Expo Symbols 57 requiere el nombre Material explícito en Android, por lo que la
flecha de apertura no se dibujaba allí. Se añade `chevron_right` manteniendo SF
Symbols en iOS y WebIcon en web. La misma omisión de `xmark` en
`DialogCloseButton` se corrige con `close` para Android. Es una reparación de
las primitivas existentes, no una nueva decisión de diseño ni una dependencia.

La vista Android del diálogo mostró el fondo oscurecido pero sin desenfoque
apreciable, frente al blur claro de iOS. Mantener como punto abierto la
verificación del blur Android en una build nativa y su superficie de captura;
no dar por cumplida esa parte de la homogeneidad a partir del scrim solamente.
OAuth real, cambios de credenciales, eliminación de cuenta, transferencia,
retirada/cancelación/finalización, todos los deportes y estados, rol delegado,
lectores de pantalla e idiomas siguen fuera de esta ejecución. La revisión
automática exige confirmación específica antes de asignar delegación a la
segunda cuenta. Las comprobaciones API no acreditan la presentación UI de ese rol.

Retrospectiva: una sesión real descubre componentes que el recorrido anónimo
no monta. Mantener cuentas y fixtures en local permite revisar permisos reales
con coste menor que introducir un acceso especial y evita que una simulación
de sesión oculte problemas de autorización.

Cierre de esta ampliación: pasan typecheck, exportación web y formato de los
componentes modificados. Tras cargar el bundle nuevo, Pixel muestra la flecha
Material en Actividad reciente. Metro y Compose local se apagan conservando
volúmenes, cuentas y fixture; se retira el puente temporal de API Android.
No se activa observabilidad ni se modifica dev público o producción.

## Ampliación autenticada y matriz de producto

La [matriz de producto](PRODUCT_QA_REVIEW_2026-10-04.md) distingue pruebas
automatizadas, API real y observación nativa. Se amplía la sesión con los ocho
deportes y 24 combinaciones de formato: 16 admitidas completadas y 8 rechazadas
según contrato. No acredita sus 24 recorridos visuales.

La delegación de `qa_visual_player` en el torneo local original fue autorizada
explícitamente por el usuario y aplicada por API; el estado previo sin delegación
de esta revisión es histórico. En iOS se comprobó login, biblioteca Administro,
apertura del torneo y formulario de edición. El resumen de permisos se corrigió
en los cuatro catálogos y se observó en español tras recargar.

El torneo de bádminton finalizado se abrió en ambos simuladores: semifinales,
final, ganador y parciales legibles. En iOS también se comprobó el retorno al
partido de origen. Los diálogos mantienen una diferencia abierta de blur; los
experimentos que daban desenfoque parcial o fondo vacío en Android se retiraron.
Los detalles, causas y cobertura que falta constan en la matriz enlazada.

El delegado guardó un 5–1 en iOS y lo vio al volver al torneo. El fixture de
fútbol cancelado se abrió en ambos sistemas: partidos conservados, marcadores
administrativos y ningún botón de edición visible. La cancelación y retirada
se realizaron antes por API; no se acredita su envío mediante UI.
