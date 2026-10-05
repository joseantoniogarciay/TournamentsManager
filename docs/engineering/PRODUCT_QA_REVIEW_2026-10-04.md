# Revisión de producto — 2026-10-04

> Estado: en curso. No acredita todas las casuísticas ni todas las plataformas.
> Base: `189acc7`, con las correcciones de rotación de enlaces y copy descritas
> aquí. Entorno local, cuentas ficticias; sin desplegar ni activar observabilidad.

## Criterio y evidencia

El usuario autoriza revisar todo el producto con sesión y datos ficticios, y
asignar `qa_visual_player` como administrador del torneo local original.
Se separan tres evidencias: dominio/HTTP automatizado, API real con PostgreSQL y
Mailpit, y recorrido visual nativo. Una respuesta API correcta no valida el
teclado, el layout, la navegación ni la accesibilidad de una pantalla.

Los scripts y respuestas de la sesión se conservan en
`/private/tmp/tm-product-qa-20261004`; las credenciales y fixtures están en el
archivo ignorado `apps/client/.env.qa-review.json`, con permisos restringidos.
No se incluyen secretos, enlaces de invitación ni cuerpos privados en Git.
Las primeras capturas nativas se observaron en la conversación; las pasadas
posteriores conservaron archivos privados de evidencia enumerados más abajo.

## Verificación automatizada ejecutada

| Comprobación                               | Resultado y alcance                                                                                                                   |
| ------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------- |
| `go test -json ./...`                      | 483 tests/subtests aprobados, 56 omitidos, ningún fallo. Las integraciones opt-in están omitidas en esta ejecución.                   |
| PostgreSQL opt-in en BD desechable         | 58 tests/subtests aprobados, sin fallos. `tm_product_qa_20261004` separada de la BD persistente de fixtures: la suite trunca cuentas. |
| `node --test tests/*.test.mjs`             | 41 aprobados, sin omitidos ni fallos (30 anteriores + 4 de borrador + 7 de entrada nativa).                                                                                                 |
| `python3 tests/operational-safety.test.py` | 11 aprobados.                                                                                                                         |
| Cliente                                    | `pnpm run check` aprobado (formato, lint, TypeScript y OpenAPI); exportación web completada; cliente generado actualizado.                                   |

Los contadores incluyen subtests y no equivalen a funciones independientes,
a cobertura de líneas ni al número de casuísticas del producto.

## Matriz funcional

| Área / casos                                         | Automatizado                                                                                                    | API local real                                                                                                                                         | Revisión de pantalla                                                                                |
| ---------------------------------------------------- | --------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------ | --------------------------------------------------------------------------------------------------- |
| Registro, verificación, duplicados, enlace consumido | Dominio, HTTP y persistencia                                                                                    | Cuenta ficticia, verificación y reutilización 409                                                                                                      | Registro sin envío en ambos; verificación por enlace pendiente                                      |
| Login pendiente y reenvío                            | Secuencial, 8 solicitudes concurrentes, cancelación, sin crear sesión                                           | Reproducción del 500 y 202 tras corregir                                                                                                               | Login verificado en ambos; login pendiente pendiente                                                |
| Login verificado, logout y refresh                   | HTTP, cookies, CSRF, persistencia                                                                               | 200/204, sesión revocada 401 y refresh reutilizado 401                                                                                                 | Creador y participante en ambos; logout confirmado y nuevo login Android; refresh visual pendiente |
| Recuperación y reautenticación                       | Persistencia, ticket de un uso, fallos DB/SMTP, timeout/cancelación y respuestas seguras                        | Solicitud, inspección, cambio, repetición 409, credencial vieja 401, nueva 200, solicitudes consecutivas 202/202                                       | Cambio de credencial por UI requiere intervención humana; pantallas y enlaces pendientes            |
| Google / Apple                                       | Challenges, validación, errores seguros, atomicidad y concurrencia con dobles de prueba                         | No se acredita acceso real a proveedores                                                                                                               | Orden de botones comprobado en ambos; OAuth real aplazado por el usuario a futuras pruebas en dev/prod                       |
| Biblioteca, recientes, seguir/dejar de seguir        | HTTP, persistencia y relaciones sin duplicar                                                                    | Paginación, seguimiento idempotente y consultas                                                                                                        | Administro/Sigo en web y ambos SO; recientes; paginación y seguir/dejar de seguir UI pendientes |
| Administradores y exclusividad del creador           | HTTP y persistencia                                                                                             | Asignación autorizada/idempotente, delegado puede editar; listar admins, cancelar, completar y transferir rechazados 403                               | Delegado edita en iOS; menú y URL directa 403 revisados en web, búsqueda con creador; Android delegado pendiente |
| Invitación de equipo                                 | HTTP y persistencia                                                                                             | Crear/regenerar, anterior inválida, inspección anónima, inscripción 201, duplicado 409, sin sesión 401, revocación idempotente, después de empezar 409 | Web anónimo → login conserva nombre e inscribe; conflicto del ya inscrito. iOS: entrada corregida, login conserva nombre, inscripción confirmada, duplicado recuperable y segundo enlace con app abierta; Android: entrada con/sin sesión, nombre obligatorio/corrección, login conserva nombre e inscripción con destino y equipos comprobados |
| Composición de equipos                               | HTTP, persistencia                                                                                              | Último equipo 409, añadir, duplicado normalizado 409, eliminar antes de empezar                                                                        | Creación web, primer equipo, duplicado y segundo válido; nombres largos en ambos; bajas UI pendientes |
| Ocho deportes × tres formatos                        | Dominio, HTTP y persistencia                                                                                    | 16 combinaciones admitidas terminadas con campeón; 8 rechazadas 400 según contrato                                                                     | Ocho formularios guardados en web; formatos largos y desempates. Nativo parcial: fútbol, bádminton, tenis de mesa y voleibol; no acredita todos los formatos |
| Liga y mixto                                         | Grupos, dos vueltas, composición impar, retirada de clasificado, desempate repetido, congelación y concurrencia | Una vuelta y mixto tabla única; empate de corte con liguilla de desempate resuelto                                                                     | Clasificación fútbol parcial en ambos; mixto y desempates en pantalla pendientes                    |
| Eliminatoria                                         | Bracket, byes, correcciones, desempate fútbol/balonmano, todos los deportes                                     | Cuadro de cuatro participantes en ocho deportes                                                                                                        | Bádminton semifinal/final en ambos; ocho finales web, corrección fútbol iOS; byes visuales pendientes |
| Resultados e incidencias                             | Marcadores, sets, límites, formas exclusivas, historia                                                          | Resultado inválido 400, no comparecencia, abandono con parcial, corrección jugada y permisos 403                                                       | Ocho deportes web; fútbol editado iOS; no comparecencia y abandono parcial guardados Android; combinaciones restantes pendientes |
| Retirada, cancelación y finalización                 | Dominio, HTTP y persistencia; co-campeones y concurrencia                                                       | Retirada, repetición 409, cancelación conserva lectura pública, edición/completar cancelado 409, finalización anticipada 409                           | Creación web hasta campeón; bádminton finalizado y fútbol cancelado en ambos; retirada en ficha de equipo pendiente |
| Notificaciones y sugerencias                         | HTTP y persistencia                                                                                             | Lista, marcar todas leídas, contador 0, sugerencia 201, corta 400                                                                                      | Notificación web y destino comprobados; confirmación de borrar cancelada. Sugerencia: corto/corrección/éxito web y Android, límite con borrador conservado iOS; lista extensa, vacíos y borrar pendientes |
| Cuenta: transferencia, baja y purga                  | Transferencia, ticket, baja/purga y anonimización                                                               | Baja del creador con torneos rechazada 409                                                                                                             | Formularios, transferencia y baja UI pendientes                                                     |
| Errores de transporte y fallback                     | Suite cliente/HTTP; nuevo reset 500 seguro                                                                      | Rechazos de negocio reales                                                                                                                             | Corte real de API, mensaje seguro y reintento exitoso en web/iOS/Android; 5xx, timeout y cancelación de navegación visual pendientes |

La pasada deportiva registró 208 peticiones: 205 coincidieron con la expectativa
inicial; tres expectativas del script eran incorrectas (formato prohibido o
transición prematura). Se ajustaron al contrato; no son tres fallos del producto.
Otras pasadas registraron 14 peticiones de delegación, 26 de identidad, 17 de
colecciones/sugerencias y 33 de invitaciones/incidencias, todas con el resultado
esperado. No se suman estos contadores como cobertura exhaustiva.

## Defectos reproducidos y resolución

1. Login de cuenta pendiente: segundo enlace devolvía 500 por insertar antes de
   invalidar el token anterior. Dos CTE de escritura independientes no garantizan
   orden. Se vincula la inserción a la invalidación y se serializa por cuenta.
2. Recuperación: segunda solicitud sufría el mismo defecto. Además, `now()` usa
   el inicio de transacción y puede preceder al token recién confirmado por otra
   transacción tras esperar un lock. Se usa el reloj de la sentencia posterior
   al lock para creación, invalidación y expiración. Tres regresiones de
   PostgreSQL acreditan reenvío, concurrencia y cancelación.
3. El contrato de solicitud de recuperación omitía su respuesta técnica 500.
   Se declara y regenera el cliente; la feature conserva el fallback seguro.
   Diez subcasos HTTP comprueban cuerpos seguros y causas cerradas del span raíz.
4. El resumen de torneo decía a cualquier persona «también puedes ... finalizarlo».
   Se corrige en los cuatro idiomas para distinguir creador y administradores
   sin cambiar permisos ni inventar una regla de negocio en el cliente.
5. Administradores mostraba Añadir tras un acceso directo rechazado 403. Se
   requiere también la consulta autorizada completada para mostrar el control.
6. La clasificación vacía de un torneo cancelado prometía disponibilidad futura.
   Se incorpora su mensaje terminal en los cuatro idiomas.
7. El nombre de equipo escrito sin sesión se perdía al volver del login. Un
   borrador temporal asociado a la invitación conserva la intención y se limpia
   al cerrar o confirmar; regresiones de ambos adaptadores y recorrido real web.
8. La card de clasificación vacía duplicaba el margen horizontal. La corrección
   del contenedor sin filas se comprobó en iOS, Android y web.
9. Sugerencias no explicaba visualmente la longitud mínima tras abandonar el
   campo. Se conecta el error localizado al `TextField` compartido; al guardar
   se reinicia también su estado de interacción para no señalar el campo vacío.
10. iOS rechazaba una invitación válida al abrir la app ya iniciada: el listener
    de la pantalla se montaba después del evento. La entrada nativa captura y
    persiste el fragmento antes de navegar; iOS confirma formulario, login e inscripción.

## Diferencias y cobertura visual abierta

[La revisión entre plataformas](CROSS_PLATFORM_VISUAL_REVIEW_2026-10-04.md)
conserva la evidencia de márgenes, iconos, escala, teclado y temas de fases
anteriores. Se amplía con fixtures reales autenticados; no se sustituyen aquellos
registros por una afirmación de cobertura total.

El diálogo iOS desenfoca el fondo completo. Android oscurece sin el mismo blur.
Se probaron referencias a BlurTargetView en contenido y en raíz: respectivamente
produjeron desenfoque parcial y fondo vacío. Se retiraron ambos experimentos.
La limitación de captura entre ventanas de React Native Modal fue descrita en
[una incidencia de Expo](https://github.com/expo/expo/issues/44165) para una
versión anterior; los resultados locales del SDK 57 son la evidencia aplicable
a esta revisión. La incidencia está cerrada y no acredita por sí sola una
limitación vigente universal. La solución queda abierta: evaluar una capa en la
misma ventana frente a una adaptación nativa sin romper fullScreenModal iOS.

Faltan recorridos completos de todas las rutas de la matriz, cuatro idiomas,
texto ampliado autenticado, orientación, VoiceOver/TalkBack, otras versiones de
SO, navegación Android por botones, release y dispositivos físicos. El usuario aplaza explícitamente el OAuth real a futuras pruebas en dev/prod;
no se bloquea esta sesión local ni se solicita su configuración ahora. Mailpit
no prueba entrega externa.

## Retrospectiva técnica

La revisión visual con cuentas reales encuentra problemas de persistencia que
los dobles de repositorio no detectan. El mínimo suficiente para la regresión
es probar PostgreSQL de forma aislada, mantener un único token activo y revisar
cada salida HTTP. Un experimento visual que mejora solo parte del fondo debe
retirarse; los casos abiertos permanecen visibles hasta reproducir su cierre.

## Segunda pasada web con API real

Web local en 833 × 726, español y tema oscuro resuelto existente. Se verificaron
login de delegado/creador, logout confirmado, biblioteca Administro/Sigo, menú
delegado con solo Compartir, equipos de lectura, clasificación y su popup de
reglas, notificación de delegación y búsqueda de administradores (mínimo, vacío
al excluir creador/delegado existente y resultado de otra cuenta ficticia).
No se concedieron permisos a esa otra cuenta.

Un acceso directo del delegado a Administradores devuelve 403 y el fallback
seguro, pero mostraba Añadir administrador. Se corrigió condicionándolo también
a la consulta autorizada completada, sin cambiar reglas del backend. Se
comprobó ausencia con delegado y presencia/lista con creador tras recargar.

La clasificación del torneo cancelado devolvía tabla vacía por la regla actual
del backend, pero el cliente prometía disponibilidad «cuando empiece». Se añade
copy de cancelación en los cuatro idiomas; web muestra el mensaje corregido.
No se modifica la proyección ni se presenta como aprobada una nueva regla de
clasificación histórica. ADR-0042 conserva el torneo de consulta y ADR-0081
mantiene la autoridad de la clasificación en backend.

La creación web autenticada se probó de principio a fin: nombre vacío con error
inline; publicación con primer equipo; segundo equipo duplicado rechazado;
segundo equipo válido; inicio; no comparecencia; marcador administrativo 3–0;
confirmación de finalización; popup de campeón y clasificación final. La API
independiente confirmó `completed`, campeón y dos filas de tabla. Fixture
`QA creación web autenticada`, conservado en el inventario privado.

Se abrieron rutas sin token de invitación, recuperación y verificación: mensaje
localizado y salida segura. Datos de acceso refleja el email ficticio; la
reautenticación con contraseña incorrecta muestra fallback seguro y permite
seguir en sesión. No se introdujo ninguna credencial nueva desde la UI.

Capturas locales: `/private/tmp/tm-product-qa-20261004/web-completed-champion.jpg`
y `cancelled-standings-web.jpg`. No acreditan tamaños móviles ni otros idiomas.
Retrospectiva: abrir directamente las URLs detecta controles que el menú normal
oculta; el estado vacío debe explicar el estado terminal que realmente devuelve
el backend, sin prometer una transición imposible.

La exportación web posterior a las correcciones de cabecera y estado cancelado
completó correctamente; `pnpm run check` y los 30 casos Node se repitieron sin
fallos. OAuth real queda aplazado por decisión explícita del usuario a futuras
sesiones en dev/prod.

## Invitación anónima → login → inscripción

Una invitación real de un nuevo torneo ficticio publicado se abrió en web con
el formato de fragmento que genera el producto. Tras capturarla, la dirección
visible quedó en `/join-team`. Con creador ya inscrito se vio el conflicto
localizado; sin sesión se mostró el CTA de login y se volvió a la invitación
tras autenticar al participante.

Se reprodujo pérdida del nombre escrito antes del login: lo sustituía la última
sugerencia de la cuenta al remontar la ruta. Se conserva ahora un borrador
temporal asociado al token concreto de invitación, con el almacenamiento ya
existente (SecureStore móvil / AsyncStorage web). No cambia la preferencia
sincronizada del último equipo: el borrador prevalece hasta inscripción o cierre
y se borra junto a la invitación. ADR-0130/0131 siguen vigentes.

La repetición web conservó «Cóndores invitados web» al volver del login; la
inscripción creó exactamente ese equipo y la ficha de equipos lo mostró sin
controles de administración. Cuatro regresiones ejecutan el módulo real con
adaptadores de almacenamiento de prueba: remontaje en web/iOS/Android, aislamiento
entre tokens, limpieza y compatibilidad con invitaciones anteriores o borradores
inválidos. No equivalen a repetir visualmente el login en los dos SO.

Captura de la regresión: `invitation-name-after-login.jpg` en el directorio privado
de evidencia. Retrospectiva: mantener el secreto pendiente no basta para mantener
la intención de la persona; también hay que conservar el campo que ya editó
antes de cruzar una ruta de autenticación.

## Formularios de los ocho deportes con API local

Se prepararon ocho torneos ficticios en curso, con dos participantes y una final,
para recorrer los formularios web autenticados. Se guardó y se volvió a mostrar
el resultado en los ocho casos:

| Deporte | Caso visual recorrido | Resultado mostrado |
| --- | --- | --- |
| Fútbol | Empate 1–1 bloqueado hasta introducir tanda 4–5 | Visitante ganador, 1 (4)–1 (5) |
| Baloncesto | Empate 80–80 bloqueado con mensaje; corrección a 82–80 | Local ganador, 82–80 |
| Balonmano | Empate 25–25 con lanzamientos de 7 m 5–4 | Local ganador, 25 (5)–25 (4) |
| Tenis | Mejor de cinco: 7–6, 7–5, 6–4 y restantes vacíos | 3–0 |
| Pádel | Mejor de tres: 6–4, 7–6 y tercero vacío | 2–0 |
| Tenis de mesa | Mejor de siete: 12–10, 11–5, 11–7, 11–9 y restantes vacíos | 4–0 |
| Voleibol | Cinco sets: 25–20, 20–25, 26–24, 20–25, 16–14 | 3–2 |
| Bádminton | A 15 puntos: 14–12 rechazado; 21–20, 15–12 aceptado | 2–0 |

En iPhone se recuperó el marcador de fútbol guardado en web, se bloqueó la tanda
4–4 y se guardó su corrección a 4–6. El ganador visitante siguió visible. Se
abrió además el formulario de siete juegos: un quinto juego posterior a las
cuatro victorias bloquea Guardar, el mensaje permanece dentro del popup y el
desplazamiento permite alcanzar el botón. Al limpiar esos campos y cerrar por
el fondo exterior no se alteró el resultado persistido. La ficha permite
desplazar la última card por encima del botón fijo de finalizar.

Pixel API 34 mostró el mensaje corregido de clasificación cancelada. No se
presentan como recorridos los ocho formularios en ambos SO ni el teclado virtual
de iOS: los campos de esta pasada se editaron mediante los controles accesibles
del simulador. El acceso real OAuth sigue aplazado por el usuario.

Evidencia privada: `football-penalties-web.jpg`, `football-edited-ios.jpg` y
`cancelled-standings-android.jpg`. El inventario privado conserva los fixtures
para repetir estos casos sin volver a crearlos.

Checklist de cierre de los cambios de cliente: se mantienen primitivas y tokens,
cuatro catálogos, navegación y autorización existentes, el adaptador generado y
`apiFetch`; no se añade dependencia ni nueva regla de negocio. `pnpm run check`
(incluido TypeScript), los 34 casos Node y la exportación web posterior al cambio
del borrador terminaron sin fallos. Permanece abierta la diferencia de blur de
Android y el resto de recorridos señalados en la matriz.

Retrospectiva: variar el número de sets y el punto final descubre límites de
formulario que una captura con el formato por defecto no cubre. La comprobación
de cierre requiere ver tanto la validación como el dato persistido al volver a
la ficha; comprobar una sola plataforma no acredita la otra.

## Margen del estado vacío de clasificación

La comparación del torneo cancelado en iPhone y Pixel reveló un margen exterior
doble: el contenedor de la tabla añadía 20 px y `Card` añadía otros 20 px. El
estado sin filas elimina ahora únicamente el padding horizontal del contenedor;
la card conserva sus 20 px. La tabla con filas conserva su layout.

Se confirmó visualmente el margen corregido y el copy de cancelación en iOS,
Android y web; se abrió después la clasificación web con tres equipos para
verificar que sus filas y columnas seguían visibles. Capturas privadas:
`cancelled-standings-ios-margin.jpg`, `cancelled-standings-android-margin.jpg` y
`cancelled-standings-web-margin.jpg`.

Retrospectiva: revisar solo tablas con datos no detecta la duplicación de margen
de una card vacía. Las reglas de `apps/client/AGENTS.md` distinguen explícitamente
el margen de `Card` del de un bloque sin superficie; la corrección aplica esa
regla existente y no introduce otro token o un ajuste exclusivo de plataforma.

## Android: teclado, incidencia y corrección

Pixel API 34 cambió desde el creador al participante ficticio mediante logout
confirmado y login con su credencial existente. La cuenta y la biblioteca
mostraron `qa_visual_player`, Administro 9 y Sigo 2. El formulario de siete juegos
recuperó el 4–0 guardado en web. Al enfocar el séptimo juego, el teclado numérico
desplaza el contenido; un juego extra incompleto bloquea Guardar y el scroll
permite llegar al mensaje y al botón. Atrás del host cerró el popup sin guardar;
se reabrió con el séptimo juego vacío y el resultado persistido original.

Se guardó no comparecencia del visitante y la ficha mostró ganador local e
incidencia. Se reabrió la selección persistida y se corrigió a abandono del
visitante con parcial 5–3. El teclado se ocultó con su control inferior sin cerrar
el popup, y el desplazamiento permitió guardar. La ficha muestra Abandono,
ganador local y Tanteo parcial: 5–3. No se acredita navegación por tres botones
del SO a partir del botón Atrás del host del emulador.

El símbolo `@` enviado desde el teclado del ordenador fue interpretado de forma
distinta por el emulador; el teclado en pantalla permitió introducirlo y el login
funcionó. Se registra como limitación de entrada de la herramienta, no como
defecto del formulario. Capturas privadas: `table-tennis-android-keyboard.jpg` y
`table-tennis-android-no-show.jpg`.

## Corte de API local y recuperación visual

Se detuvo únicamente el contenedor de API local, conservando PostgreSQL, Mailpit
y volúmenes. Web recargó la clasificación; Pixel solicitó la ficha de voleibol;
iPhone volvió a la ficha del torneo cancelado. Los tres mostraron el mensaje
común de conexión y una única card con Reintentar, sin cuerpo técnico ni banner
duplicado. Se restauró la API local saludable y se pulsó Reintentar en cada uno:
volvieron respectivamente la tabla con tres equipos, el voleibol 3–2 y el torneo
cancelado con sus jornadas. No se cerró la app para recuperarse.

Capturas privadas: `network-error-web.jpg`, `network-error-android.jpg` y
`network-error-ios.jpg`. Este recorrido acredita rechazo de transporte y
recuperación; no equivale a respuestas HTTP 5xx, timeout, red del SO desactivada
ni todas las cancelaciones de navegación.

Retrospectiva: el fallo de transporte debe comprobarse con la dependencia
realmente inaccesible y cerrarse con un reintento exitoso. El estado correcto de
error por sí solo no acredita que la persona pueda continuar su tarea.

Notificaciones web mostró el aviso ficticio de delegación y la confirmación de
Eliminar todas, con advertencia de irreversibilidad y Cancelar. No se ha ejecutado
el borrado: queda pendiente la confirmación específica solicitada al usuario.
La captura privada es `qa-notifications-delete-confirmation.jpg`.

## Sugerencias: longitud, éxito y límite entre plataformas

Con `QA`, web deshabilitaba Enviar sin mensaje visible incluso después del blur:
el requisito solo existía como accessibilityHint. La corrección usa la clave ya
localizada en los cuatro idiomas y `validationTrigger="blur"` de `TextField`.
La revisión de éxito detectó además que vaciar el texto conservaba la interacción
del campo y mostraba el error junto al agradecimiento: después de persistir se
remonta únicamente el campo para reiniciar esa interacción. Los fallos no
reinician el campo ni pierden el borrador. No cambia contrato, adaptador ni
transportes de ADR-0125.

Web y Pixel confirmaron texto corto, error tras blur, corrección que habilita
Enviar y limpieza tras éxito. Dos envíos web y uno Android de la misma cuenta
ficticia alcanzaron el límite; el cuarto desde iPhone mostró el mensaje
específico de espera conservando el texto y permitiendo editarlo. En iOS la
entrada con teclado del ordenador y el envío quedaron revisados; no se acredita
el blur del texto corto ni el recorrido con teclado táctil: pulsar la descripción
no terminó la edición en esta configuración del simulador.

Evidencias privadas: `suggestion-short-web.jpg`, `suggestion-short-android.jpg`,
`suggestion-success-web.jpg`, `suggestion-success-android.jpg` y
`suggestion-rate-limit-ios.jpg`. `pnpm run check` y exportación web completados
con éxito. El reinicio de Metro reescribió `expo-env.d.ts`; se normalizó el archivo
y no quedó diferencia de contenido. No se añadieron pruebas que repitan estas
cinco líneas de presentación; la validación usa recorridos reales y checks del
cliente. Checklist: primitiva, locales existentes, tokens/márgenes intactos,
error por campo, acción sin duplicados y adaptador generado con apiFetch.

Notificaciones web: se canceló el diálogo de borrado, se conservó el aviso y se
abrió el torneo original desde él. El borrado irreversible sigue sin ejecutar.

Retrospectiva: el botón correctamente deshabilitado no explica por sí solo cómo
recuperarse. Una prueba de formulario debe cerrar también el éxito y comprobar
que no persiste una validación de la edición anterior; compartir cuenta entre
SO permite revisar que el límite no depende del dispositivo.

## Entrada nativa de invitación: evento previo al montaje

Dos fixtures nuevos locales, uno por SO, permiten revisar la inscripción tras
login sin reutilizar el equipo web ya inscrito. Tras logout confirmado, Safari
en iPhone abrió el esquema local de la app con fragmento válido. La pantalla
mostró «Esta invitación ya no está disponible», aunque la inspección anónima
por API real respondió 200 con ese mismo torneo. La captura privada
`invitation-ios-unavailable.jpg` conserva la reproducción.

Evidencia de causa: el `useURL` instalado empieza su listener al montar y lee
`getInitialURL`, que no recupera ese evento posterior al arranque. Expo Router
recibe el enlace antes de montar la ruta. Su extensión `+native-intent` permite
esperar la persistencia y devolver `/join-team` sin secreto; se usa ese punto
para enlaces de esquema, universales y Expo Go. La pantalla nativa restaura el
almacenamiento; web conserva la captura de fragmento previa. Un fragmento
inválido limpia token y borrador anteriores; un fallo de almacenamiento vuelve
a Inicio sin propagar la URL secreta. El parser existente se mueve al módulo
de la feature, sin cambiar el contrato de token.

Siete regresiones del módulo real cubren arranque/app abierta por cada adaptador
nativo, formatos de enlace, invalidación del borrador, espera del guardado,
fallo seguro, conservación del gate inicial y retorno OAuth. La suite Node
completa pasa 41 casos; `pnpm run check` y exportación web terminan con éxito.

La recarga completa en iOS cargó la extensión existente `+native-intent.ts`.
Un primer experimento duplicó esa extensión con `.tsx`; se retiró al comprobar
que Expo Router seleccionaba la original. No queda extensión duplicada ni
logging de diagnóstico. Hot refresh no sustituye la recarga del punto de entrada.

Repetición real iOS: formulario del torneo correcto sin sesión; «Cóndores
invitados iOS» se conserva al autenticar la cuenta ficticia existente; inscripción
abre el torneo y mantiene Iniciar deshabilitado para el participante. El detalle
público confirma ambos equipos. Reabrir el enlace y repetir la inscripción
muestra conflicto localizado conservando el formulario. Un enlace del segundo
fixture recibido con la ruta de invitación ya abierta cambia al torneo correcto
sin reiniciar. Evidencias: `invitation-ios-restored-name.jpg` y
`invitation-ios-warm-second.jpg`.

Android: autorización específica del usuario para términos iniciales de Chrome,
usado sin cuenta. Una página ficticia servida solo en loopback entregó el enlace;
tras recarga completa de Expo, la app ya viva mostró el torneo correcto con
sesión. El nombre vacío muestra «Introduce el nombre de tu equipo»; corregirlo
retira el error. La inscripción de «Condores invitados Android» se confirma en
el detalle público de API con ambos equipos. Tras desbloquear, el destino
muestra el torneo correcto, Iniciar deshabilitado y los dos equipos sin acciones
de administración. Evidencias privadas: `invitation-android-required-name.jpg`,
`invitation-android-destination.jpg` e `invitation-android-teams.jpg`.

Web confirma que el equipo inscrito desde iOS aparece en Equipos y el torneo
en Sigo con permisos de participante; evidencia `invitation-ios-teams-web.jpg`.
No se acreditan asociaciones reales de enlaces universales de dev/prod mediante
un esquema local ni una terminación fría real mediante una recarga de Expo.

Retrospectiva: probar almacenamiento y retorno del login no detecta un evento
que llegó antes de existir la pantalla. La captura del enlace pertenece al
punto de entrada de navegación; la vista conserva formulario, inspección y
recuperación, sin inventar validez a partir de una URL.


## Continuación 2026-10-05: invitación anónima y retorno nativo

Un fixture nuevo local, «QA invitacion Android sin sesion», evita confundir el
conflicto del participante ya inscrito con un éxito nuevo. Tras logout confirmado,
Chrome entrega el enlace y la app muestra el torneo correcto sin sesión. El campo
vacío no reutiliza el último nombre de la cuenta anterior. «Halcones Android sin
sesion», escrito antes del login, se conserva al volver e inscribirse abre el
detalle correcto con Iniciar deshabilitado; la consulta pública confirma ambos
equipos. Evidencias privadas: `invitation-android-anonymous-name.jpg` e
`invitation-android-restored-name.jpg`. Esa primera prueba mostró iconos blancos
sobre canvas claro al regresar.

### Corrección de apariencia y navegación

Inicio era la única pantalla que montaba `StatusBar`. Ahora la raíz controla
su estilo con el tema resuelto y lo renueva por revisión de sesión. Es aplicación
de ADR-0056 aceptado; no cambia preferencias del SO, binario ni contratos.
No se usa `screenOptions.statusBarStyle`: el binario iOS actual requiere control
global y aquella alternativa produjo una excepción de configuración nativa.

La repetición contextual Android reprodujo `ScreenStackFragment added into a
non-stack container` incluso sin esa opción; traza privada `android-stack-crash.log`.
Separar actualización de tabs, dismiss, replace y fin de transición en frames
cancelables permitió volver con el nombre conservado. En iOS apareció pantalla
negra tras autenticar. El aislamiento confirmó que persistía con su secuencia
anterior y sin el nuevo StatusBar; no se atribuye a la espera Android ni a la
apariencia. Recargar JS y «Go home» no recuperaban el presentador, pero terminar
el proceso y cargar de nuevo Metro sí.

La solución conserva el stack raíz sin `key={revision}`: no se destruye el
contenedor nativo que presenta un modal mientras se solicita el destino nuevo.
`NativeTabs` mantiene su revisión de sesión y dismiss/replace eliminan las rutas
anteriores. Android conserva sus operaciones separadas en frames cancelables;
iOS y web conservan la secuencia inmediata. No se añaden librerías ni se cambian
las reglas de sesión o permisos.

### Evidencia de la versión final

- iOS: Safari → invitación anónima → nombre «Halcones iOS stack estable» → login
  vuelve al torneo correcto con el nombre conservado e iconos oscuros; cerrar
  vuelve a Inicio. El aviso nativo «Guardar contraseña» se descarta con «Ahora no».
  Evidencia privada: `invitation-ios-stable-stack-return.jpg`.
- Android: recarga completa → logout → Chrome → invitación → login vuelve al
  torneo correcto con «Halcones Android repeticion» e iconos oscuros, sin la
  excepción anterior. Evidencia: `invitation-android-stable-stack-return.jpg`.
- La entrada host de `@` se convirtió en `²`; los envíos masivos de contraseña
  perdían caracteres. Se corrigió mediante tecla táctil y escritura individual;
  no se atribuye ese comportamiento al producto.
- Diez regresiones ejecutan el efecto real de `SessionNavigator`: destino
  contextual, logout, sesión idle, conservación del contenedor raíz y cancelación
  antes/después de dismiss y replace. El conjunto Node completo pasa 51 pruebas.
  No sustituyen la evidencia UIKit/Fabric.
- `pnpm run check` y exportación web pasan para el stack estable; logs privados
  `check-stable-session-navigator.log` y `export-stable-session-navigator.log`.

Checklist cliente: tema compartido, apariencia raíz, tokens/márgenes/localización
intactos, cierre existente y URL canónica conservados, ninguna operación OpenAPI
modificada. La matriz general sigue abierta: contraste oscuro, accesibilidad,
varios formatos/roles y OAuth real aplazado por el usuario no quedan acreditados
por estos recorridos. El arranque en frío de una development build llevó primero
al launcher; cargar después Metro no acredita un arranque directo de producción
ni asociaciones universales reales.

Retrospectiva: conservar datos y resetear tabs no exige destruir su presentador
nativo. Un check estático no detecta una carrera UIKit ni incompatibilidad del
binario; aislar variables y repetir el recorrido original evita convertir una
hipótesis de apariencia en una causa no demostrada.
