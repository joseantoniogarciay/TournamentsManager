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
| `node --test tests/*.test.mjs`             | 67 aprobados, sin omitidos ni fallos; última ejecución del 2026-10-05, incluido seguimiento y paginación.                             |
| `python3 tests/operational-safety.test.py` | 11 aprobados.                                                                                                                         |
| Cliente                                    | `pnpm run check` aprobado (formato, lint, TypeScript y OpenAPI); exportación web completada; cliente generado actualizado.            |

Los contadores incluyen subtests y no equivalen a funciones independientes,
a cobertura de líneas ni al número de casuísticas del producto.

## Matriz funcional

| Área / casos                                         | Automatizado                                                                                                    | API local real                                                                                                                                         | Revisión de pantalla                                                                                                                                                                                                                                                                                                                            |
| ---------------------------------------------------- | --------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Registro, verificación, duplicados, enlace consumido | Dominio, HTTP y persistencia                                                                                    | Cuenta ficticia, verificación y reutilización 409                                                                                                      | Registro sin envío en ambos; activación y enlace consumido comprobados en web; enlace nativo pendiente                                                                                                                                                                                                                                          |
| Login pendiente y reenvío                            | Secuencial, 8 solicitudes concurrentes, cancelación, sin crear sesión                                           | Reproducción del 500 y 202 tras corregir                                                                                                               | Login verificado en ambos; login pendiente y reenvío comprobados en web; pendiente en móvil                                                                                                                                                                                                                                                     |
| Login verificado, logout y refresh                   | HTTP, cookies, CSRF, persistencia                                                                               | 200/204, sesión revocada 401 y refresh reutilizado 401                                                                                                 | Creador y participante en ambos; logout confirmado y nuevo login Android; refresh visual pendiente                                                                                                                                                                                                                                              |
| Recuperación y reautenticación                       | Persistencia, ticket de un uso, fallos DB/SMTP, timeout/cancelación y respuestas seguras                        | Solicitud, inspección, cambio, repetición 409, credencial vieja 401, nueva 200, solicitudes consecutivas 202/202                                       | Cambio de credencial por UI requiere intervención humana; pantallas y enlaces pendientes                                                                                                                                                                                                                                                        |
| Google / Apple                                       | Challenges, validación, errores seguros, atomicidad y concurrencia con dobles de prueba                         | No se acredita acceso real a proveedores                                                                                                               | Orden de botones comprobado en ambos; OAuth real aplazado por el usuario a futuras pruebas en dev/prod                                                                                                                                                                                                                                          |
| Biblioteca, recientes, seguir/dejar de seguir        | HTTP, persistencia y relaciones sin duplicar                                                                    | Paginación, seguimiento idempotente y consultas                                                                                                        | Administro/Sigo y recientes en web y ambos SO; paginación de 52 en web/iOS, pendiente Android; seguimiento manual web/Android/iOS, retorno a biblioteca web/iOS; vacíos web                                                                                                                                                                     |
| Administradores y exclusividad del creador           | HTTP y persistencia                                                                                             | Asignación autorizada/idempotente, delegado puede editar; listar admins, cancelar, completar y transferir rechazados 403                               | Delegado edita en ambos SO; menú solo Compartir en web/iOS/Android; URL directa 403 revisada en web y búsqueda con creador                                                                                                                                                                                                                      |
| Invitación de equipo                                 | HTTP y persistencia                                                                                             | Crear/regenerar, anterior inválida, inspección anónima, inscripción 201, duplicado 409, sin sesión 401, revocación idempotente, después de empezar 409 | Web anónimo → login conserva nombre e inscribe; conflicto del ya inscrito. iOS: entrada corregida, login conserva nombre, inscripción confirmada, duplicado recuperable y segundo enlace con app abierta; Android: entrada con/sin sesión, nombre obligatorio/corrección, login conserva nombre e inscripción con destino y equipos comprobados |
| Composición de equipos                               | HTTP, persistencia                                                                                              | Último equipo 409, añadir, duplicado normalizado 409, eliminar antes de empezar                                                                        | Creación web, primer equipo, duplicado y segundo válido; nombres largos en ambos; bajas UI pendientes                                                                                                                                                                                                                                           |
| Ocho deportes × tres formatos                        | Dominio, HTTP y persistencia                                                                                    | 16 combinaciones admitidas terminadas con campeón; 8 rechazadas 400 según contrato                                                                     | Ocho formularios guardados en web; formatos largos y desempates. Nativo parcial: fútbol, bádminton, tenis de mesa y voleibol; no acredita todos los formatos                                                                                                                                                                                    |
| Liga y mixto                                         | Grupos, dos vueltas, composición impar, retirada de clasificado, desempate repetido, congelación y concurrencia | Una vuelta y mixto tabla única; empate de corte con liguilla de desempate resuelto                                                                     | Clasificación fútbol parcial en ambos; mixto y desempates en pantalla pendientes                                                                                                                                                                                                                                                                |
| Eliminatoria                                         | Bracket, byes, correcciones, desempate fútbol/balonmano, todos los deportes                                     | Cuadro de cuatro participantes en ocho deportes                                                                                                        | Bádminton semifinal/final en ambos; ocho finales web, corrección fútbol iOS; byes visuales pendientes                                                                                                                                                                                                                                           |
| Resultados e incidencias                             | Marcadores, sets, límites, formas exclusivas, historia                                                          | Resultado inválido 400, no comparecencia, abandono con parcial, corrección jugada y permisos 403                                                       | Ocho deportes web; fútbol editado por delegado en ambos SO; no comparecencia y abandono parcial guardados Android; combinaciones restantes pendientes                                                                                                                                                                                           |
| Retirada, cancelación y finalización                 | Dominio, HTTP y persistencia; co-campeones y concurrencia                                                       | Retirada, repetición 409, cancelación conserva lectura pública, edición/completar cancelado 409, finalización anticipada 409                           | Creación web hasta campeón; bádminton finalizado y fútbol cancelado en ambos; retirada en ficha de equipo pendiente                                                                                                                                                                                                                             |
| Notificaciones y sugerencias                         | HTTP y persistencia                                                                                             | Lista, marcar todas leídas, contador 0, sugerencia 201, corta 400                                                                                      | Notificación web y destino comprobados; confirmación de borrar cancelada. Sugerencia: corto/corrección/éxito web y Android, límite con borrador conservado iOS; vacío comprobado en web; lista extensa, vacío móvil y borrar pendientes                                                                                                         |
| Cuenta: transferencia, baja y purga                  | Transferencia, ticket, baja/purga y anonimización                                                               | Baja del creador con torneos rechazada 409                                                                                                             | Datos de acceso, formulario vacío de contraseña y cancelación de baja comprobados en web; transferencia, baja efectiva y móvil pendientes                                                                                                                                                                                                       |
| Errores de transporte y fallback                     | Suite cliente/HTTP; nuevo reset 500 seguro                                                                      | Rechazos de negocio reales                                                                                                                             | Corte real de API, mensaje seguro y reintento exitoso en web/iOS/Android; 5xx, timeout y cancelación de navegación visual pendientes                                                                                                                                                                                                            |

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

| Deporte       | Caso visual recorrido                                      | Resultado mostrado             |
| ------------- | ---------------------------------------------------------- | ------------------------------ |
| Fútbol        | Empate 1–1 bloqueado hasta introducir tanda 4–5            | Visitante ganador, 1 (4)–1 (5) |
| Baloncesto    | Empate 80–80 bloqueado con mensaje; corrección a 82–80     | Local ganador, 82–80           |
| Balonmano     | Empate 25–25 con lanzamientos de 7 m 5–4                   | Local ganador, 25 (5)–25 (4)   |
| Tenis         | Mejor de cinco: 7–6, 7–5, 6–4 y restantes vacíos           | 3–0                            |
| Pádel         | Mejor de tres: 6–4, 7–6 y tercero vacío                    | 2–0                            |
| Tenis de mesa | Mejor de siete: 12–10, 11–5, 11–7, 11–9 y restantes vacíos | 4–0                            |
| Voleibol      | Cinco sets: 25–20, 20–25, 26–24, 20–25, 16–14              | 3–2                            |
| Bádminton     | A 15 puntos: 14–12 rechazado; 21–20, 15–12 aceptado        | 2–0                            |

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

## Delegado Android: resultado, menú y lectura

Con la delegación explícitamente autorizada de `qa_visual_player` en el torneo
ficticio «QA Liga de equipos con acentos y nombres largos», se abrió el torneo
desde Administro. Editar resultado muestra el teclado numérico sin ocultar Guardar;
se corrigió el primer partido de 5–1 a 5–2. El diálogo se cierra y la ficha muestra
5–2; una lectura independiente de la API confirma el resultado persistido.

El menú del delegado contiene únicamente Compartir: no ofrece Administradores,
Cancelar ni Finalizar. Equipos presenta los tres nombres completos y no muestra
acciones de composición. En Clasificación, Peñas tiene dos partidos, dos goles
a favor y cero puntos. Deslizar las estadísticas permite ver GC y DG conservando
posición, nombre y puntos: Sur 5/2/+3, Norte 3/0/+3 y Peñas 2/8/−6. Los nombres
de la tabla siguen truncados según su ancho; no se expanden al tocarlos y pueden
consultarse completos en Equipos.

Evidencia privada: `delegated-android-result-saved.jpg`,
`delegated-android-menu.jpg`, `delegated-android-standings-scroll.jpg` y
`delegated-android-teams.jpg`, bajo el directorio privado de esta revisión.
Los JSON antes/después corroboran el guardado; no acreditan permisos de otras
rutas ni sustituyen los rechazos 403 ya probados. La superposición del engranaje
de Expo interceptaba el centro del menú; tocar el borde superior visible abrió
el control del producto. No se registra como defecto de la build de producción.

Retrospectiva: para acreditar un rol hay que comprobar una mutación permitida,
su persistencia y la ausencia de acciones exclusivas; mostrar Administro no basta.
La lectura de clasificación y equipos complementa esa evidencia sin ampliar
permisos ni alterar los demás partidos.

## Reanudación del 5 de octubre: apariencia Android

Se relanzaron API, PostgreSQL y Mailpit locales, sin observabilidad. Metro y el
servidor auxiliar seguían activos. El Pixel restaurado no respondía a toques;
al reiniciar aparecieron ANR de Pixel Launcher y UI del sistema antes de abrir
Fast Tourney. Se pausó el análisis del IDE y se apagó temporalmente iOS; después
se recuperó la navegación Android. Es evidencia de una limitación del entorno,
no una causa demostrada ni un defecto atribuido al producto.

Se cambió la preferencia del producto de Sistema a Oscuro. Ajustes, Cuenta,
Inicio y Crear torneo mostraron fondos y textos del tema; la barra de estado
usó iconos claros. El diálogo de logout mantuvo título, explicación, confirmar
y cancelar legibles; Cancelar conservó la sesión. El envío vacío de Crear torneo
mostró ambos mensajes obligatorios junto a sus campos, sin ocultar la acción
ni desalinear el cierre respecto a la tarjeta. Se restauró Sistema y se confirmó
el retorno a claro con iconos oscuros.

Capturas privadas: `android-dark-settings.jpg`, `android-dark-session-dialog.jpg`,
`android-dark-create.jpg` y `android-dark-create-required.jpg`. No acreditan
contraste WCAG medido, todas las pantallas oscuras, tipografía aumentada ni la
paridad iOS.

Retrospectiva: recuperar una sesión de emulador no garantiza que su sistema
responda. Verificar primero el launcher y separar las pruebas por SO permite
registrar evidencia sin confundir un ANR del entorno con navegación de la app.

## Apariencia y texto aumentado iOS — 5 de octubre

Con solo iOS activo, el arranque de Fast Tourney Local recuperó la sesión del
participante. Se seleccionó Oscuro desde Ajustes: tarjetas, texto y cierre
adoptaron el tema y la barra de estado mostró iconos claros. Inicio mantuvo las
tarjetas recientes legibles. El diálogo de logout mostró sus dos acciones y
Cancelar conservó la sesión. Enviar Crear torneo vacío mostró los dos errores
obligatorios completos junto a sus campos, con la acción visible. Se restauró
Sistema antes de continuar, recuperando el tema claro.

Se aumentó Text Size del simulador de 3 a 7. Cuenta mantuvo sus acciones; el
header nativo agrupó sus controles en More. En Crear torneo, las ocho opciones
de deporte se redistribuyeron sin recortar sus etiquetas. El desplazamiento
permitió alcanzar el campo de equipo y Crear torneo; enviar sin nombre de torneo
mostró su error, legible al volver mediante desplazamiento. El campo de equipo
contenía el nombre sugerido existente; no se creó un torneo. Se restauró Text
Size a 3 y se verificó el valor del simulador.

Capturas privadas: `ios-dark-settings.jpg`, `ios-dark-session-dialog.jpg`,
`ios-dark-create-required.jpg` y `ios-large-text-create-error.jpg`. Esta pasada
no acredita contraste WCAG medido, todos los tamaños de accesibilidad, lector
de pantalla, teclado con texto aumentado ni todas las pantallas del producto.

Se repitió el corte de la API local desde Inicio. Abrir un torneo mostró en el
árbol nativo «No hemos podido conectarnos. Revisa tu conexión e inténtalo de
nuevo.» y Reintentar. El Mac se bloqueó antes de capturar la pantalla o repetir
la recuperación; esa repetición queda pendiente, sin invalidar la evidencia de
recuperación ya registrada en la matriz. Se apagaron API, PostgreSQL, Mailpit,
Metro y el servidor auxiliar conservando los datos. El análisis de Android
Studio había quedado pausado y no pudo restaurarse por UI con el Mac bloqueado.

Retrospectiva: aumentar el texto cambia tanto el flujo del contenido como los
controles nativos del header. Comprobar que las acciones y los errores siguen
siendo alcanzables aporta evidencia más útil que una captura inicial aislada;
no sustituye una revisión completa de accesibilidad.

## Reanudación: recuperación y cabecera de torneo iOS

Tras desbloquear el Mac se relanzaron API, PostgreSQL, Mailpit y Metro sin
observabilidad. Reintentar desde el error conservado cargó el torneo original
y sus marcadores 5–2 y 3–0. Se completó así la repetición interrumpida del corte.
Se reanudó el análisis de Android Studio; el Pixel siguió apagado.

Se reprodujo una colisión del nombre largo de torneo con los controles de la
toolbar iOS. La ficha limita ahora el título con el ancho de ventana, insets
seguros y reserva semántica de controles y separación. Con tamaño estándar se
ve completo en dos líneas sin invadirlos. Con Text Size 7 conserva la separación
y aparece truncado por el espacio nativo de la cabecera; no se acredita lectura
completa del título a ese tamaño. También se corrigió la división de
Clasificación en una línea y una letra: las acciones conservan su ancho de
contenido y permiten envolver la fila. Text Size volvió a 3.

Evidencia privada: `ios-long-title-bounded.jpg`,
`ios-long-title-large-text.jpg` y `ios-large-text-summary-actions.jpg`.
Checklist: se mantienen tokens, margen exterior, cierre nativo, tema,
localización y ruta canónica; ninguna operación OpenAPI ni permiso cambia.
La revisión visual del ancho nativo se realizó en iOS; web también mostró el
nombre largo y ambas acciones completas a tamaño estándar en 702 × 762
(`web-summary-actions.jpg`). El menú iOS siguió abriendo únicamente Compartir
y se cerró sin enviar datos. Android queda pendiente de una nueva pasada visual.
`pnpm run check` y exportación web pasan con la versión final; los logs privados
son `check-bounded-title.log` y `export-bounded-title.log`. Metro regeneró
`expo-env.d.ts` sin el formato esperado; normalizarlo dejó ese archivo sin diff.

Retrospectiva: un título con dos líneas no garantiza que el host nativo reserve
su ancho. El nombre largo y el texto aumentado descubren problemas diferentes;
la corrección debe respetar la reserva de controles y el ancho del contenido,
sin introducir tamaños fijos por nombre o idioma.

## Biblioteca: huecos identificados en el cliente

En iOS, Sigo muestra cinco torneos de la cuenta participante. Abrir «QA invitacion
nativa iOS» mantiene la espera del creador y el inicio deshabilitado; su menú
contiene únicamente Compartir. La búsqueda de código confirma que las operaciones
generadas de seguir/dejar de seguir no tienen un adaptador ni controles de
feature. La API ya probada y el seguimiento automático al inscribirse no acreditan
esas acciones manuales en pantalla. Quedan como funcionalidad cliente pendiente,
no como una casuística visual aprobada.

`listRelatedTournaments` y `getTournamentRelationship` solicitan una sola página
con límite 50 e ignoran `nextCursor`. ADR-0058 exige solicitar páginas posteriores.
Por inspección, un torneo administrado fuera de la primera página no se encuentra
al resolver la relación, además de quedar fuera de la biblioteca. Se trata de una
limitación comprobada en código; la reproducción visual con más de 50 relaciones
queda pendiente. Una lectura real devuelve 31 torneos administrados del propietario
y nueve del participante, sin cursor posterior: esos fixtures no reproducen
todavía el umbral. No se interpreta la ausencia de acciones de ese caso como
una regla de autorización del backend.

Retrospectiva: registrar un endpoint como probado no implica que esté conectado
al producto. Separar una acción inexistente, un límite estático y una reproducción
visual evita inflar la cobertura de la matriz.

## Biblioteca: paginación y límite de permisos — 5 de octubre

La reproducción local con más de 50 torneos encontró además un error de frontera
backend: el servicio devolvía como cursor el primer elemento no entregado, mientras
SQL usa `id < cursor`. Ese elemento desaparecía de todas las páginas. El cursor
ahora es el último ID entregado, conforme a ADR-0058; la fila adicional solo
confirma que hay otra página. La colección real de QA contiene 52 torneos: 50 en
la primera página y dos en la segunda, todos únicos, sin cursor terminal.

La biblioteca conserva cada colección y su cursor. «Cargar más», localizado en
los cuatro idiomas, pide solo la página siguiente; `50+` indica el mínimo cargado
y se convierte en `52` al agotar la colección. Actualizar vuelve a la primera
página conservando el segmento seleccionado. Un fallo parcial conserva los datos
y el cursor para reintentar; uno inicial usa RequestErrorCard sin banner duplicado.
Se bloquean envíos duplicados y se descartan respuestas anteriores a un refresh,
cambio de cuenta o desmontaje. La consulta de relación recorre páginas hasta
hallar el torneo o agotar la colección; no declara ausencia desde la primera página.

Web e iOS mostraron `50+`, permitieron Cargar más y terminaron con `52`. Abrir en
iOS el torneo original de la segunda página mantuvo Editar resultado y Añadir
resultado para su creador. Capturas privadas: `web-pagination-complete.jpg` e
`ios-pagination-complete.jpg`; evidencia HTTP: `pagination-evidence.json`.

La regresión PostgreSQL aislada recorre cinco torneos en tres páginas y exige
que no haya omisiones ni duplicados. Las pruebas cliente cubren búsqueda de un
administrador delegado después de 50, agotamiento, páginas inválidas y cursores
que no avanzan, fallos HTTP, reintento, doble pulsación, cambio de segmento,
refresh frente a append tardío, logout y recuperación de la carga inicial.

Checklist cliente: adaptador generado con authenticatedApiFetch, estado en hook
de feature, tokens y Button compartidos, márgenes de 20 px, reserva de tabs,
locales es/en/it/fr y errores seguros. No hay nueva dependencia ni decisión
funcional; se completa el comportamiento aceptado en ADR-0058. Seguir y dejar de
seguir manualmente permanecen pendientes. La matriz general continúa en curso.

Retrospectiva: probar una sola página no acredita paginación. Hay que contrastar
el cursor con la desigualdad SQL y exigir que la unión de páginas recupere todos
los elementos, además de comprobar la UI y los permisos fuera de la primera.

Validación automatizada del corte: `pnpm run check`, exportación web, 59 pruebas
Node y tests Go de dominio/HTTP pasan. La nueva regresión PostgreSQL pasa con
`TM_RUN_INTEGRATION=1` en `tm_product_qa_20261004`, separada del producto local.
Se revisaron las salidas de GET /v1/me/tournaments: éxito paginado y vacío,
validación, sesión inválida, ausencia de limitador o rechazo de negocio propio,
fallos PostgreSQL y cancelación. Se conservan las categorías seguras del span
raíz y el fallback localizado del cliente; cursor, cuenta y errores internos
quedan fuera del feedback y de atributos. No se activó observabilidad.

El Pixel se relanzó y se restablecieron los puertos 8080/8082. La build conservaba
Ajustes y respondía con retrasos; también el launcher tardó en procesar Home y el
cajón de aplicaciones. Se apagó iOS para liberar recursos. Volver a abrir Fast
Tourney Local restauró la pantalla anterior; no se logró verificar una recarga
ni la paginación Android. Al intentar revisar el arranque en frío el Mac se
bloqueó. Este caso sigue pendiente: no se atribuye el retraso a la paginación ni
se presenta Android como aprobado. Se cierran los servicios locales conservando
volúmenes, fixtures, base aislada y evidencia.

## Continuación — seguimiento manual (2026-10-05)

Se completa el control cliente de ADR-0034, reutilizando el menú del torneo y
las operaciones PUT/DELETE de seguimiento ya existentes. Solo se ofrece con
sesión y relación resuelta sin administración; no modifica equipos, inscripción
ni permisos. La consulta de seguidores también recorre toda la paginación.
La biblioteca se invalida al confirmar una mutación y se actualiza al volver.

Web: la cuenta ficticia participante deja de seguir QA invitacion nativa iOS;
Sigo pasa de cinco a cuatro y desaparece su card. Desde la ficha pública lo
sigue de nuevo y Sigo vuelve a cinco con la card restaurada. Evidencia privada:
web-unfollow-library.jpg y web-follow-library.jpg. Android: el menú de QA
invitacion nativa Android cambia de Dejar de seguir a Seguir torneo tras DELETE,
y vuelve a Dejar de seguir tras PUT; android-unfollow-menu.jpg y
android-follow-menu.jpg. El ciclo conserva la participación existente.

La continuación del Pixel reprodujo UI del sistema no responde en el launcher
tras arranque en frío (android-system-anr.jpg); cerrar System UI permitió abrir
la app. Se restablecieron los puertos. Android Studio también mostró Low memory
mientras indexaba C++. Se intentó pausar la indexación, pero no se acreditó el
cambio. La comprobación visual iOS del seguimiento y la paginación Android de
52 torneos siguen pendientes; el arranque iOS llegó a abrir Fast Tourney, sin
acreditar aún el recorrido. No se presentan como aprobados.

Checklist cliente: adaptadores generados con authenticatedApiFetch; lógica en
hook de feature; ModalDialog y Button compartidos; copy en es/en/it/fr; controles
sin envíos duplicados; feedback seguro y ausencia de banner de éxito redundante.
67 pruebas Node, pnpm run check y exportación web pasan. Se cubren respuestas
204, 404 de PUT, rechazos HTTP y transporte, bloqueo de doble envío, respuestas
tardías tras logout/ruta/desmontaje y recarga de biblioteca por invalidación.

Revisión de salidas PUT/DELETE: éxito idempotente, UUID inválido 400, sesión 401,
CSRF 403, PUT no visible 404, PostgreSQL 5xx y cancelación; no hay rate limit
propio. PUT 404 usa el mensaje localizado de torneo no disponible; los estados
restantes no tratados y problemas desconocidos usan fallback seguro, transporte
usa common_network_error y sesión invalidada no duplica feedback. No se activa
observabilidad. La matriz general sigue en curso, con OAuth real diferido por
el usuario y los demás recorridos pendientes expresados en su inventario.

Retrospectiva: una proyección en caché necesita invalidación tras la mutación.
El contador y la presencia de la card al volver son la evidencia del resultado;
el cambio de etiqueta por sí solo no acredita esa actualización.

Al cerrar este corte, las capturas y acciones de coordenadas sobre simuladores
empezaron a devolver ausencia de ventana incluso tras recuperar sus bindings;
las lecturas AX seguían disponibles y Back del Pixel sí se procesaba. No se
supone un bloqueo de pantalla sin evidencia. Se conserva el pendiente nativo
y se apagan API, PostgreSQL, Mailpit y Metro sin eliminar volúmenes.

## Continuación — seguimiento iOS y activación web (2026-10-05)

Se relanzan API, PostgreSQL, Mailpit y Metro para esta sesión autorizada, sin
observabilidad. La cuenta ficticia participante abre QA invitacion nativa iOS,
deja de seguir desde el menú y vuelve a la biblioteca: Sigo pasa de cinco a
cuatro y desaparece la card (ios-unfollow-library.jpg). Para volver a abrir la
ficha se restaura el seguimiento por API como preparación de fixture; no se
presenta esa preparación como interacción nativa. Desde la ficha se repite
DELETE y se ejecuta PUT en iOS; al volver, Sigo mantiene cinco y la card está
restaurada (ios-follow-library.jpg). Se conserva la participación. El intento
de refresco por arrastre no acredita una actualización y no se marca aprobado.

Web: qa_lifecycle_682d52 estaba pendiente. Dos intentos de login muestran el
reenvío de verificación y conservan el formulario sin crear sesión
(web-pending-login.jpg). Se obtiene el enlace del Mailpit local, se abre en el
navegador y se verifica la sesión con ese username en Cuenta
(web-verified-account.jpg). Reutilizar el mismo enlace muestra el mensaje
localizado de enlace ya utilizado (web-verification-link-used.jpg); volver a
Inicio conserva la sesión anterior. Logout y nuevo login con la contraseña
existente funcionan. No se cambia ninguna credencial ni se aceptan condiciones.
Los enlaces con token permanecen solo en archivos privados de permisos 0600.

La cuenta nueva muestra No tienes notificaciones y vuelve correctamente a
Cuenta (web-empty-notifications.jpg). Administro 0 y Sigo 0 muestran sus
mensajes vacíos; seleccionar Sigo cambia el mensaje y conserva ambos contadores
(web-empty-followed-library.jpg). Estos vacíos no acreditan las pantallas móviles.

Pixel: se recupera una captura de Cuenta como qa_visual_player, pero Android
Studio muestra Low memory y al intentar cerrar sesión el control vuelve a
perder la ventana. No se acredita logout ni paginación de 52 en Android; no se
atribuye el fallo de control al producto. La matriz sigue en curso. OAuth real
sigue diferido por el usuario y los restantes pendientes conservan su alcance.

Este corte añade evidencia y corrige el inventario, sin cambios de código ni
contrato. Las 67 pruebas y el check/export anteriores corresponden a 3301c65;
no se presentan como una ejecución nueva. Se verifica formato y diff del
inventario actualizado y se cierran los servicios locales conservando volúmenes.

Retrospectiva: comprobar un flujo requiere observar también su destino y el
segundo uso de un enlace de un solo uso. Preparar un fixture por API permite
continuar una revisión, pero debe quedar separado de la evidencia de UI. Un
problema de memoria del host mantiene pendientes sus casos nativos, sin
invalidar las comprobaciones independientes de web e iOS.

## Continuación — recuperación de memoria y acceso web (2026-10-05)

El usuario solicita liberar memoria y continuar. La inspección de procesos
identifica Android Studio con aproximadamente 2,4 GB residentes y 580–650 %
de CPU; el Monitor de Actividad confirma No responde. No queda Metro de la
sesión anterior. Tras desbloqueo manual se apaga iOS por Shut Down en Device
Hub y se fuerza la salida de Android Studio desde Monitor de Actividad. Se
confirma por procesos que Studio y el emulador terminaron; swap pasa de unos
6,6 a 4,8 GB tras ese cierre, sin limpiar cachés ni eliminar datos.

Se relanzan API, PostgreSQL, Mailpit y Metro sin observabilidad. Android Studio
abre Welcome y consume unos 400 MB; se arranca Pixel desde Virtual Device
Manager, pero no se obtiene su pantalla controlable hasta abrir el proyecto y
reiniciar el dispositivo desde Running Devices. Los puertos 8080/8082 se
restablecen. Power Save Mode muestra explícitamente On; la pausa de indexación
se intenta, pero el análisis C++ sigue avanzando y no queda acreditada.

FastTourney abre Inicio con recientes, conserva qa_visual_player y permite
cerrar sesión con su confirmación. El primer intento de escribir correo pierde
caracteres; se corrige por tramos y se observa completo antes de continuar.
No se acredita el envío de login del propietario ni la paginación Android.
Studio vuelve a unos 2,3 GB y 650–700 % de CPU, con fallo de captura de pantalla.
Se cierra de nuevo por Monitor de Actividad y se verifica la ausencia de sus
procesos y del emulador. iOS queda apagado. No se atribuye este bloqueo al
producto ni se da por resuelta la estabilidad del IDE al reiniciarlo.

Web, con qa_lifecycle_682d52: Datos de acceso muestra el correo de la cuenta y
sus acciones. Eliminar cuenta abre el diálogo que explica cierre de sesiones,
seguimientos, delegaciones y eliminación definitiva tras 30 días; Cancelar
cierra el diálogo y conserva Datos de acceso. Evidencia privada:
web-account-delete-confirmation.jpg y web-account-access-after-cancel.jpg.
No se ejecuta baja, transferencia ni purga. Cambiar contraseña abre los campos
de contraseña actual y nueva con Guardar deshabilitado cuando están vacíos;
Volver regresa a Datos de acceso (web-password-form-empty.jpg). No se introduce
una credencial nueva ni se acredita cambio o reautenticación por este recorrido.

Se conservan los pendientes de la matriz, incluido OAuth real diferido. Este
corte solo añade documentación y evidencia; se comprueban formato y diff y se
apagan servicios locales y Metro al cerrar, conservando volúmenes y fixtures.
Power Save Mode del IDE se mantiene activo para limitar tareas del entorno.

Retrospectiva: comparar memoria y CPU antes y después del cierre ayuda a
identificar un proceso bloqueado. El arranque de Welcome no acredita el coste
del proyecto cargado; la indexación puede reproducir la presión aunque el
modo de ahorro esté activo. Verificar destino y cancelación mantiene separados
los formularios revisados y las operaciones de cuenta todavía no ejecutadas.

Al cerrar Metro aparece un aviso de expo-blur: se usa
`dimezisBlurViewSdk31Plus` sin `blurTarget` y el módulo anuncia fallback sin
blur. `shared/ui/confirmation-dialog.tsx` declara ese método sin target. Queda
pendiente revisar el contrato del SDK y el fondo del diálogo en Android; no se
interpreta este aviso como causa de la falta de memoria del IDE.
