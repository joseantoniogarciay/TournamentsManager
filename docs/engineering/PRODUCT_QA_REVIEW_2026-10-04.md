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

| Comprobación                               | Resultado y alcance                                                                                                                           |
| ------------------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------- |
| `go test -json ./...`                      | 483 tests/subtests aprobados, 57 omitidos, ningún fallo; ejecución del 2026-10-07. Las integraciones opt-in están omitidas en esta ejecución. |
| PostgreSQL opt-in en BD desechable         | 58 tests/subtests aprobados, sin fallos. `tm_product_qa_20261004` separada de la BD persistente de fixtures: la suite trunca cuentas.         |
| `node --test tests/*.test.mjs`             | 78 aprobados, sin omitidos ni fallos; última ejecución del 2026-10-08, incluidas reautenticación asíncrona y renovación web.                                    |
| `python3 tests/operational-safety.test.py` | 11 aprobados.                                                                                                                                 |
| Cliente                                    | `pnpm run check` aprobado (formato, lint, TypeScript y OpenAPI); exportación web completada; cliente generado actualizado.                    |

Los contadores incluyen subtests y no equivalen a funciones independientes,
a cobertura de líneas ni al número de casuísticas del producto.

## Matriz funcional

| Área / casos                                         | Automatizado                                                                                                    | API local real                                                                                                                                         | Revisión de pantalla                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| ---------------------------------------------------- | --------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Registro, verificación, duplicados, enlace consumido | Dominio, HTTP y persistencia                                                                                    | Cuenta ficticia, verificación y reutilización 409                                                                                                      | Android, web e iOS: registro enviado con validación, disponibilidad, aceptación y correo; activación y enlace consumido comprobados en web, iOS y Android                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| Login pendiente y reenvío                            | Secuencial, 8 solicitudes concurrentes, cancelación, sin crear sesión                                           | Reproducción del 500 y 202 tras corregir                                                                                                               | Login verificado en ambos; login pendiente y dos reenvíos comprobados en web, iOS y Android                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| Login verificado, logout y refresh                   | HTTP, cookies, CSRF, persistencia                                                                               | 200/204, sesión revocada 401 y refresh reutilizado 401                                                                                                 | Creador y participante en ambos; logout confirmado y nuevo login Android; renovación visual web tras 401 controlado e iOS con vencimiento próximo simulado y corte/reintento comprobada; iOS: refresh revocado 401 real, cierre persistido y nuevo login comprobados; web: corte de refresh corregido y reintento, 500 sin logout y revocación 401 real comprobados; Android: renovación 200, corte/reintento y 500/429 sin logout, revocación 401 real, cierre persistido tras reapertura y nuevo login comprobados; iOS conserva sesión tras nuevo arranque y reautenticación incorrecta; Android conserva sesión tras reinstalar la development build                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| Recuperación y reautenticación                       | Persistencia, ticket de un uso, fallos DB/SMTP, timeout/cancelación y respuestas seguras                        | Solicitud, inspección, cambio, repetición 409, credencial vieja 401, nueva 200, solicitudes consecutivas 202/202                                       | Web: solicitud inválida/válida, enlace real y formulario preparados; cambio de credencial espera intervención humana. Reautenticación correcta/incorrecta, doble clic y cierre durante espera comprobados; iOS: contraseña incorrecta segura con sesión conservada; Android: solicitud inválida/válida, enlace real hasta formulario, enlace inválido y corte/reintento comprobados; iOS: solicitud vacía/incorrecta/válida, corte/reintento, enlace inválido y enlace real hasta formulario comprobados; cambio efectivo de credencial pendiente de intervención humana                                                                                                                                                                                                                    |
| Google / Apple                                       | Challenges, validación, errores seguros, atomicidad y concurrencia con dobles de prueba                         | No se acredita acceso real a proveedores                                                                                                               | Orden de botones comprobado en ambos; OAuth real aplazado por el usuario a futuras pruebas en dev/prod                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| Biblioteca, recientes, seguir/dejar de seguir        | HTTP, persistencia y relaciones sin duplicar                                                                    | Paginación, seguimiento idempotente y consultas                                                                                                        | Administro/Sigo y recientes en web y ambos SO; paginación de 52 en web/iOS y 54 en Android; seguimiento manual y retorno a biblioteca en las tres plataformas; vacíos web, Android e iOS                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| Administradores y exclusividad del creador           | HTTP y persistencia                                                                                             | Asignación autorizada/idempotente, delegado puede editar; listar admins, cancelar, completar y transferir rechazados 403                               | Delegado edita en ambos SO; menú solo Compartir en web/iOS/Android; URL directa 403 revisada en web y búsqueda con creador                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| Invitación de equipo                                 | HTTP y persistencia                                                                                             | Crear/regenerar, anterior inválida, inspección anónima, inscripción 201, duplicado 409, sin sesión 401, revocación idempotente, después de empezar 409 | Web anónimo → login conserva nombre e inscribe; conflicto del ya inscrito. iOS: entrada corregida, login conserva nombre, inscripción confirmada, duplicado recuperable y segundo enlace con app abierta; Android: entrada con/sin sesión, nombre obligatorio/corrección, login conserva nombre e inscripción con destino y equipos comprobados                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| Composición de equipos                               | HTTP, persistencia                                                                                              | Último equipo 409, añadir, duplicado normalizado 409, eliminar antes de empezar                                                                        | Creación web y Android, primer equipo, duplicado y corrección; Android cancela y confirma baja previa al inicio, protege último equipo y repone visitante; nombres largos en ambos; web cancela y confirma eliminación previa al inicio, protege último equipo y repone visitante con persistencia; iOS cancela y confirma eliminación previa al inicio, protege último equipo, rechaza duplicado y repone visitante con GET persistido                                                                                                                                                                                                                                                                                                                                                     |
| Ocho deportes × tres formatos                        | Dominio, HTTP y persistencia                                                                                    | 16 combinaciones admitidas terminadas con campeón; 8 rechazadas 400 según contrato                                                                     | Web: 16 formatos admitidos completados por UI; preparación de fixtures por API. Nativo parcial: fútbol, bádminton, tenis de mesa y voleibol; Android añade edición de fixtures de tenis y pádel por sets; Android e iOS completan por UI ligas y mixtos de baloncesto, balonmano y voleibol desde el inicio, con fixtures preparados por API; no acredita todos los formatos                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| Liga y mixto                                         | Grupos, dos vueltas, composición impar, retirada de clasificado, desempate repetido, congelación y concurrencia | Una vuelta y mixto tabla única; empate de corte con liguilla de desempate resuelto                                                                     | Web: dos vueltas con co-campeones, grupos, empate de corte, segundo ciclo, clasificación congelada y transición hasta campeón; ligas y mixtos de baloncesto/balonmano/voleibol terminados. Android: liga de fútbol a dos vueltas con co-campeones; mixto con dos grupos, clasificación congelada y eliminatorias hasta campeón; lectura final de ligas y mixtos de baloncesto/balonmano/voleibol comprobada. Android e iOS: inicio, resultado, clasificación y finalización con campeón en ligas de esos tres deportes. Android e iOS completan además sus mixtos de liga general con cuatro equipos, dos clasificados y final; clasificación congelada y campeón confirmados                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| Eliminatoria                                         | Bracket, byes, correcciones, desempate fútbol/balonmano, todos los deportes                                     | Cuadro de cuatro participantes en ocho deportes                                                                                                        | Bádminton semifinal/final en ambos; ocho finales web; cuadro de cinco equipos con tres pases directos hasta campeón web; corrección fútbol iOS; Android crea y completa final de fútbol con tanda; lectura y navegación de byes finalizados Android; Android inicia un cuadro de cinco equipos, juega cuartos/semifinales/final y confirma campeón; iOS inicia cuadro de tres equipos, comprueba pase directo y final pendiente, juega semifinal/final y confirma campeón con reapertura                                                                                                                                                                                                                                                                                                    |
| Resultados e incidencias                             | Marcadores, sets, límites, formas exclusivas, historia                                                          | Resultado inválido 400, no comparecencia, abandono con parcial, corrección jugada y permisos 403                                                       | Web: no comparecencia y abandono parcial en ocho deportes; corrección a marcador con penaltis fútbol. Fútbol editado por delegado en ambos SO; Android: voleibol inválido/válido, escritura y finalización; balonmano no comparecencia, abandono parcial y corrección a tanda; tenis no comparecencia y abandono 2–3 con validación y reapertura; iOS: empate de baloncesto bloqueado y recuperación; añade validación de pádel, bádminton, balonmano, tenis y voleibol en la pasada del 7/10 y guardado, reapertura y restauración de esos cinco fixtures; iOS añade no comparecencia y abandono en tenis de mesa, bádminton, balonmano, pádel, voleibol, tenis y fútbol, con reapertura y restauración exacta de los siete fixtures; un fixture nuevo de baloncesto pasa alta por incidencia, abandono, corrección a marcador jugado y finalización autorizada con campeón y reapertura; quedan pendientes las incidencias y combinaciones nativas no acreditadas; Android tenis/pádel comprueba sets inválidos, corrección, sets extra tras victoria, guardado y restauración del fixture; Android bádminton a 15 comprueba diferencia de dos, tope 21, excepción 21–20 y juego extra bloqueado; Android tenis de mesa a siete juegos comprueba último parcial, bloqueo de dos incompletos, scroll hasta Guardar y restauración; Android añade no comparecencia y abandono en bádminton, pádel y fútbol, reapertura y restauración exacta de los tres fixtures; fútbol comprueba eliminación de penaltis y copy específico de parcial incompleto; Android añade incomparecencia local 0–4 de tenis de mesa a siete juegos y restauración exacta del abandono 5–3; copy de parcial por tanteos revisado en es/en/it/fr Android |
| Retirada, cancelación y finalización                 | Dominio, HTTP y persistencia; co-campeones y concurrencia                                                       | Retirada, repetición 409, cancelación conserva lectura pública, edición/completar cancelado 409, finalización anticipada 409                           | Creación web hasta campeón; bádminton finalizado y fútbol cancelado en ambos; retirada web, cancelación y persistencia comprobadas; iOS comprueba lectura de baja y protección 3–0 y finaliza baloncesto con campeón; Android finaliza voleibol con campeón; Android confirma retirada nueva en liga de tres equipos, cancela antes y comprueba sustitución de jugado/pendiente por 3–0, clasificación y persistencia; iOS cancela y confirma retirada en liga de tres equipos, comprueba sustitución 0–3/3–0, clasificación, editor no afectado y reapertura                                                                                                                                                                                                                               |
| Notificaciones y sugerencias                         | HTTP y persistencia                                                                                             | Lista, marcar todas leídas, contador 0, sugerencia 201, corta 400                                                                                      | Web: 31 notificaciones, desplazamiento hasta la última y destino comprobados; borrado preparado y pendiente de confirmación específica. Sugerencia: corto/corrección/éxito web y Android; iOS verifica éxito con recepción única en Mailpit, campo limpio, mínimo 7/8 y tope ASCII 1.000, además del límite de tasa con borrador conservado; vacío comprobado en web e iOS; iOS carga 31 enlaces y el último accesible abre destino; Android alcanza las 31 por gestos, abre la última y conserva posición al volver, y comprueba vacío; desplazamiento iOS y borrado pendientes                                                                                                                                                                                                                                                                                                                           |
| Cuenta: transferencia, baja y purga                  | Transferencia, ticket, baja/purga y anonimización                                                               | Baja del creador con torneos rechazada 409                                                                                                             | Web: datos de acceso y cancelación de baja; transferencia ficticia preparada y pendiente de confirmación; cambio de contraseña requiere intervención humana; iOS: datos de acceso, cancelación de baja y formulario vacío comprobados; Android: datos de acceso, reautenticación incorrecta segura, cancelación de baja y formulario vacío comprobados; baja efectiva/purga pendientes                                                                                                                                                                                                                                                                                                                                                                                                      |
| Errores de transporte y fallback                     | Suite cliente/HTTP; nuevo reset 500 seguro                                                                      | Rechazos de negocio reales                                                                                                                             | Corte real de API y reintento en web/iOS/Android; web adicional con 500, cuerpo 200 inválido, demora/504, navegación durante espera y recuperación. Android añade refresco de ficha con cambio externo, corte/reintento, 500, cuerpo 200 inválido, 429, 404 seguro y salida durante espera; web comprueba también el nuevo control, esos errores, bloqueo, salida durante espera y cambio externo real; iOS añade cambio externo, 429/500/cuerpo inválido, 404 y recuperación, bloqueo y salida durante espera, con banner global conservado al navegar. Fallos inyectados en proxy; no acredita timeout propio del cliente                                                                                                                                                                                                                                                                                                            |

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
11. Recuperación mostraba el error genérico ante una desconexión de transporte.
    Se reutiliza `getRequestFailure` para mostrar el mensaje común de conexión;
    se conserva el tratamiento localizado de 429 y el fallback seguro restante.
    Android acredita corte real, borrador conservado y reintento tras recuperar API.

12. El error compartido de resultado hablaba de sets también en bádminton, que
    presenta juegos. Se ajusta su copy en los cuatro idiomas para referirse a
    ganar el partido, manteniendo las ayudas específicas y la validación existente.
    Android reproduce el error y confirma el mensaje corregido.

13. Cabecera Android al 200 %: Cambiar contraseña quedaba pegado a Volver,
    con ambos límites en x=168. Se sustituye el título estándar de Cuenta en
    Android por Text compartido con dos líneas y ancho calculado según ventana,
    control y separación. Español confirma 53 px físicos (aprox. 20 dp) frente
    al botón, con y sin teclado. Datos de acceso y Cambiar contraseña se revisan
    al 200 % también en inglés, italiano y francés; los títulos quedan completos.
14. La etiqueta francesa de Guardar contraseña ocupa dos líneas al 200 % y
    quedaba alineada a la izquierda dentro del botón centrado. La primitiva
    Button centra también las líneas de su texto; captura real confirma ambas
    líneas completas y centradas, sin cambiar el escalado ni el envío.

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
texto ampliado autenticado en rutas aún no acreditadas, orientación, VoiceOver/TalkBack, otras versiones de
SO, builds de distribución y dispositivos físicos. La variante release local Android
se acredita parcialmente en la continuación del 2026-10-08. Android por tres botones tiene el alcance parcial descrito más abajo; orientación horizontal de la app no se acredita. El usuario aplaza explícitamente el OAuth real a futuras pruebas en dev/prod;
no se bloquea esta sesión local ni se solicita su configuración ahora. Mailpit
no prueba entrega externa.

La apertura de un enlace de torneo en Android produjo una caída de DevLauncher
el 2026-10-06. Recuperar la app permitió continuar y un segundo intento abrió
el destino; la incidencia de arranque/enlace permanece abierta, con evidencia
en la continuación de tenis de mesa. La tanda del 2026-10-08 añade arranques y
enlaces en una variante release local; no demuestra que el defecto de
development build esté corregido.

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

## Continuación — grupos, pases directos y resultados de baja (2026-10-05)

Se reanudan únicamente API, PostgreSQL, Mailpit y Metro locales; no se activa
observabilidad ni se arrancan simuladores mientras sigue pendiente resolver la
presión de memoria del IDE. Web inicia sesión con el propietario ficticio.
Se preparan por API dos torneos nuevos (cinco y ocho equipos): esta preparación
no acredita creación por interfaz y eleva la colección del propietario de 52 a
54; la prueba histórica de paginación con 52 conserva su alcance original.

Por interfaz web se selecciona e inicia la eliminatoria de cinco equipos. El
cuadro presenta un partido de cuartos y tres pases directos, sin acción de
resultado en esos pases, y conserva enlaces a semifinales y final. Evidencia:
web-byes-bracket.jpg. No acredita todavía el recorrido hasta campeón ni móvil.

El torneo mixto de ocho equipos se configura en Grupos: cuatro grupos y dos
clasificados muestran que faltan equipos y deshabilitan Iniciar torneo. Cambiar
a dos grupos permite iniciar; se observan doce partidos repartidos en dos
grupos y tres jornadas. Clasificación permite alternar entre ambos grupos,
cada uno con sus cuatro equipos correctos y valores iniciales cero. Evidencia:
web-group-standings.jpg. La transición a eliminatorias y los desempates por
el corte siguen pendientes de validación visual en este fixture.

En la liga ficticia original, Retirar Peñas del Mediterráneo abre confirmación;
Cancelar conserva el equipo. Aceptar mantiene el equipo como Retirado y cambia
ambos partidos al 3–0 para sus rivales. Recargar conserva la baja. Se reproduce
un defecto: Editar resultado seguía disponible y guardar 2–0 sobrescribía el
resultado administrativo, incumpliendo ADR-0041. Antes de reiniciar el backend
se restaura el 3–0 por API, conservando historial; esa corrección pierde la
marca administrativa en la versión anterior, por lo que también sirve como
fixture de compatibilidad con datos afectados por el defecto.

La protección existente de voleibol se extiende a todos los resultados
administrativos y, además, a partidos con un participante retirado aunque una
versión anterior haya perdido esa marca. La transacción ya serializa cambios
sobre el torneo. La interfaz oculta la edición y protege la apertura del editor;
el backend conserva la autoridad y rechaza la escritura con 409. OpenAPI
explicita este caso sin crear un estado nuevo. Tras reiniciar API, un intento
HTTP real de guardar 2–0 recibe 409 y la lectura conserva 3–0. Web muestra ambos
3–0 sin edición, mientras el partido de los dos equipos activos conserva Añadir
resultado (web-withdrawal-fixed.jpg). No se reparan automáticamente datos
históricos: fuera de este fixture requerirían una revisión específica.

La regresión PostgreSQL aislada verifica rechazo, marcador e historial intactos,
compatibilidad sin marca administrativa y edición de equipos activos. El test
de incidencias comprueba la protección tras baja para todos sus deportes.
Dominio y HTTP pasan; 67 tests Node pasan; check de cliente pasa. La exportación
web se completa con 36 rutas. Se actualiza el comentario generado desde OpenAPI.
El cierre de servicios se comprueba al terminar esta fase.

El aviso de expo-blur se debe a que el SDK instalado requiere blurTarget para
los métodos Android de Dimezis, pero el Modal compartido no lo declara. El
módulo ya deshabilitaba el blur y usaba su tinte de respaldo. Se declara
blurMethod="none" explícitamente, conservando esa atenuación y el blur nativo
iOS (el método es una propiedad Android). No se añade un target que atraviese
ventanas nativas ni se presenta como restauración del blur Android. La
comparación visual nativa sigue pendiente; se documenta esta limitación en el
diseño compartido.

Retrospectiva: una baja no termina al ver Retirado. Hay que intentar corregir
sus partidos, verificar igualdad de trato entre rivales y comprobar que la
protección sigue funcionando con datos previamente afectados. Separar la
preparación de fixtures de acciones de interfaz evita exagerar la cobertura.

Revisión del endpoint de resultado: éxito 200 conserva edición e historial en
partidos de equipos activos; entrada inválida 400 usa validation.rejected;
sesión inválida 401, autorización 403, ausencia 404 y conflicto 409 mantienen
sus contratos. No hay límite de tasa específico en esta operación. El rechazo
de baja usa tournament.result_conflict en el span HTTP raíz con plantilla de
ruta; no incorpora atributos por deporte, partido, equipo o resultado ni spans
por la subconsulta. Fallos de adquisición/transacción, lectura, escritura,
historial, actualización y commit, timeout y cancelación conservan las
categorías técnicas existentes y la respuesta segura; no cambia la frontera
instrumentada. El adaptador cliente sigue usando recordMatchResult generado y
authenticatedApiFetch, con recuperación localizada para 409, mensaje común de
red para rechazo de transporte y fallback seguro para 5xx, cuerpo inválido o
estado no tratado. Cancelaciones intencionadas no muestran feedback.

Cierre de fase: check del cliente (formato, lint, TypeScript y OpenAPI),
exportación web, 67 tests Node, regresiones aisladas de bajas/incidencias,
tests de dominio y HTTP, golangci-lint (0 incidencias) y goimports pasan.
OpenAPI genera únicamente la actualización esperada del comentario de la
operación. Checklist cliente: primitivas y tokens compartidos, sin textos
nuevos, navegación intacta, host de diálogo en cada Screen, adaptador generado
y feedback seguro; blur real Android y los recorridos móviles siguen pendientes.
Metro se detiene; Compose confirma que API, PostgreSQL y Mailpit están apagados
y no quedan servicios locales en ejecución, sin eliminar volúmenes ni evidencia.

## Continuación web — formatos, incidencias y recuperación (2026-10-06)

El alcance sigue en curso: OAuth real está aplazado por decisión del usuario;
transferencia del torneo ficticio original y borrado de sus 31 avisos esperan
confirmación específica, y el restablecimiento de contraseña está preparado
para intervención humana. No se presentan estos pasos como ejecutados. La
prioridad acordada es agotar web antes de continuar las apps.

### Recorridos deportivos por interfaz

La eliminatoria de cinco equipos se completa jugando sus cuatro partidos reales;
los tres pases directos no ofrecen editor. Finalizar muestra Equipo 1 campeón
(web-byes-completed.jpg). Los grupos de ocho equipos completan sus doce partidos
0–0 y generan dos liguillas de desempate de cuatro equipos. La primera necesita
un segundo ciclo entre tres equipos por empate en el corte; el ya clasificado no
vuelve a jugar. Tras resolver ambos cortes, se inician semifinales y final y se
finaliza el torneo (web-tiebreak-resolved.jpg, web-mixed-completed.jpg).

Se reproduce una incoherencia: el primer ciclo cerrado aún ofrecía Editar
resultado; el backend devuelve 409 sin cambiar el 2–0 existente. La interfaz
comprueba ahora fase activa, estado del pool y ciclo actual antes de ofrecer o
abrir edición. Se verifica que desaparecen los controles cerrados, permanecen
los actuales y los pools resueltos no permiten editar. La autoridad sigue en el
backend; no cambia el contrato ni la regla aceptada de congelación.

En la final de fútbol se guarda no comparecencia, se corrige a abandono del
visitante con parcial 1–2 (vence el rival aunque vaya por detrás), y después a
marcador 1–1 con penaltis 4–3. Sin penaltis o con 4–4, Guardar permanece
deshabilitado. Una semifinal con final dependiente tampoco admite corrección.
Evidencia adicional: web-mixed-retirement.jpg.

Siete fixtures nuevos preparados por API permiten cerrar por UI los formatos
restantes: fútbol liga a ida y vuelta con dos 2–2 termina con co-campeones;
baloncesto y balonmano completan liga y mixto; voleibol completa liga y mixto.
Los mixtos juegan seis encuentros de fase inicial y final entre clasificados.
En baloncesto, 80–80 bloquea Guardar; 80–70 es válido. En voleibol, el quinto
set 15–14 bloquea Guardar; 16–14 permite guardar y terminar. Junto a las ocho
eliminatorias ya revisadas y el mixto de fútbol, esto acredita las 16
combinaciones admitidas en web, sin atribuir la creación de fixtures a la UI.
Evidencia: web-cochampions-two-legs.jpg, web-volleyball-league-completed.jpg y
web-{basketball,handball,volleyball}-mixed-completed.jpg.

Otros siete fixtures de eliminatoria de dos participantes comprueban no
comparecencia y abandono en baloncesto, balonmano, tenis, pádel, tenis de mesa,
bádminton y voleibol. Primero falta Web Visitante y gana Web Local; corregir a
abandono de Web Local con parcial 3–2 hace ganar Web Visitante. Se finaliza cada
torneo y se comprueba ese campeón. No se convierte el parcial en un marcador
jugado ni se exige completar sets tras el abandono. Fútbol cubre el octavo
deporte en el recorrido anterior. Capturas web-tennis-incident-completed.jpg y
web-{basketball,handball,padel,table_tennis,badminton,volleyball}-retirement.jpg.
El texto que explica los deportes limitados a eliminatoria incorpora bádminton
en los cuatro idiomas, según ADR-0141.

### Fallos y continuidad

Un proxy local temporal conserva las peticiones reales y altera únicamente la
lectura del fixture de pases directos. Un 500 con title/detail sintéticos y un
200 de cuerpo inválido muestran common_request_error, sin publicar el contenido
interno. Una respuesta retrasada doce segundos y terminada en 504 muestra la
carga centrada y después el fallback. Volver a Cuenta durante otra espera no
muestra un aviso tardío. Retirar el fallo y pulsar Reintentar recupera el torneo
persistido. Esto prueba demora y fallo de gateway, no un timeout implementado
por el cliente. Captura web-500-safe.jpg; no se cambian endpoints, observabilidad
ni datos del torneo para inducir estos fallos.

La reautenticación de Vincular Google enviaba dos tickets con doble clic. Un
contador que registra solo el número de peticiones, con demora de un segundo,
reproduce dos antes de corregir y una después. Se bloquea sincrónicamente la
operación y se muestra loading; cerrar/inactivar el diálogo invalida su
generación y descarta respuestas tardías. Contraseña incorrecta conserva sesión
y muestra el fallback seguro; cerrar antes de esa respuesta no muestra feedback.
Las callbacks del diálogo son estables y cada error de proveedor se procesa una
vez: la prueba había descubierto un ciclo de renders que agotaba React. Sin
Google configurado no se solicita un challenge; tras confirmar identidad se
muestra Continuar con Google deshabilitado. No acredita vinculación real.
Evidencias: web-reauth-single-request.jpg y web-reauth-wrong-password-safe.jpg.
La feature conserva el adaptador generado y apiFetch para esta comprobación:
un 401 por contraseña incorrecta no debe invalidar la sesión existente.

### Listas, cuenta y accesibilidad

La lista del participante contiene 31 notificaciones: treinta avisos sintéticos
preparados por SQL sin cambiar permisos y el aviso original de delegación. Se
alcanza el último, que abre el torneo correcto. La lectura real posterior del contador devuelve 200 y cero avisos sin leer. El diálogo irreversible de
borrar se cancela; no se ejecuta aún ese borrado. Capturas
web-notifications-long-bottom.jpg y web-delete-notifications-prepared.jpg.

Recuperación valida correo mal formado, bloqueo del envío vacío y solicitud
válida con mensaje que no revela existencia de cuenta. El enlace real de
Mailpit abre el formulario de nueva contraseña, email no editable y Guardar
deshabilitado sin nueva credencial. La pestaña queda preparada para el usuario;
no se introduce ni se confirma una contraseña nueva por el agente. Un enlace
sin token muestra el estado inválido y salida a Inicio.

La pasada actual usa 702×756 px efectivos: el override solicitado de 390 px no
se aplicó y se retiró. No se acredita un viewport de 390 en esta pasada. Crear
torneo no desborda horizontalmente; controles de cierre e inputs/botones tienen
44 px y Tab alcanza cierre y deporte. Validar nombre vacío conserva el borrador
de equipo. Ajustes refleja Sistema sin modificar la preferencia. Términos y
privacidad permiten llegar al final y cerrar; sus textos legales versionados
conservan menciones a ligas, que no se editan silenciosamente en esta revisión.
Una ruta desconocida vuelve a Inicio y un UUID válido ausente muestra el estado
localizado de no disponible con cierre.

Retrospectiva: terminar una fase deportiva incluye volver a sus resultados
anteriores y tratar de corregirlos. Una operación rápida oculta duplicados y
respuestas tardías: retrasarla permite comprobar exclusión y cierre con evidencia
observable. Contar peticiones sin registrar cuerpos evita guardar credenciales.
La preparación por API, la ejecución por UI y la intervención humana mantienen
alcances distintos; ninguna suma equivale a cobertura exhaustiva del producto.

Verificación de esta continuación: cinco regresiones de comportamiento del
diálogo prueban doble evento antes del render, respuesta correcta/incorrecta
tras cierre, error de proveedor único y error tardío sin cerrar la reapertura.
La suite Node completa suma 72 aprobados. Check (incluido TypeScript) y
exportación de las 36 rutas web pasan; no se añaden dependencias ni se cambia
OpenAPI. Checklist cliente: se conservan primitivas, tokens, objetivos de 44 px,
localización en cuatro catálogos, host de diálogos por Screen, navegación y
adaptadores generados. Los permisos y resultados siguen siendo del backend.
Los aspectos nativos pendientes no se acreditan con esta exportación.

El intento de reanudar iOS encuentra el Mac bloqueado según el control nativo;
se solicita desbloqueo manual, sin sustituir la interacción por comandos de UI.

Cierre de sesión: se detiene el proxy, se restaura la publicación original de
API en 8080 y se apagan API, PostgreSQL y Mailpit. Metro queda detenido. Compose
no devuelve servicios locales en ejecución y no quedan listeners en 8080,
8082 ni 8084. Se conservan volúmenes, fixtures, capturas y formulario de
restablecimiento; al reanudar habrá que arrancar el entorno y comprobar si su
enlace sigue vigente. No se toca producción ni observabilidad.

Al vaciar la salida acumulada de Metro aparece también un aviso de animación
nativa por ausencia de RCTAnimation. La salida contiene logs de la sesión y no
identifica por sí sola una pantalla o un arranque actual; no se atribuye a esta
corrección web. Antes de validar animaciones iOS hay que verificar el binario y
sus módulos, y recompilar si se reproduce la ausencia, según la checklist de
cliente. El bloqueo del Mac impide esa comprobación en este cierre.

## Continuación apps — cuenta y resultados iOS (2026-10-06)

Se reanudan API, PostgreSQL, Mailpit y Metro locales sin observabilidad. Device
Hub queda inicialmente en Connecting display; Restart, conservando datos,
recupera el iPhone 18 Pro con iOS 27.0. Fast Tourney Local carga el bundle
actual desde Metro y conserva la sesión de qa_visual_player. En esta salida
nueva de Metro no se reproduce el aviso anterior de RCTAnimation; esto no
acredita por sí solo todas las animaciones ni un binario recompilado.

Datos de acceso muestra el correo esperado. Vincular Google abre la
reautenticación; una contraseña incorrecta muestra el fallback seguro y
conserva la sesión. Eliminar cuenta abre la explicación de sesiones, relaciones
y plazo de 30 días; Cancelar mantiene la cuenta. Cambiar contraseña presenta
ambos campos y Guardar deshabilitado al estar vacíos; Volver restaura Datos de
acceso. No se introduce una credencial nueva ni se ejecuta la baja. Capturas:
ios-reauth-wrong-password-safe.jpg, ios-account-delete-confirmation.jpg e
ios-password-form-empty.jpg en el directorio privado de evidencia.

Notificaciones carga los 31 enlaces accesibles del participante. Activar el
último por accesibilidad abre el torneo original; cerrarlo vuelve a la lista.
Los intentos de scroll y drag del control nativo no desplazan visiblemente
la lista, por lo que no se acredita alcanzar visualmente su final ni se
atribuye esa limitación al producto. No se elimina ningún aviso.

El torneo original conserva los dos 3–0 administrativos sin edición, el
partido entre equipos activos mantiene Añadir resultado y Equipos muestra
Peñas del Mediterráneo como Retirado. Evidencia: ios-withdrawal-fixed.jpg e
ios-withdrawn-team.jpg. Es una revisión de lectura y protección visual de
la baja ejecutada antes en web, no una retirada nueva realizada en iOS.

El mixto de voleibol ya finalizado abre Liga, Eliminatorias y Clasificación.
Los seis resultados iniciales y la final 3–0 con sets 25–0 no ofrecen edición;
la tabla conserva cuatro equipos y los puntos 9, 6, 3 y 0. Evidencia:
ios-volleyball-mixed-completed-read.jpg e ios-volleyball-mixed-standings.jpg.
No acredita jugar este mixto completo mediante iOS.

En QA formulario basketball, que estaba en curso con final 82–80, el editor
bloquea Guardar al escribir 80–80 y explica que el resultado final no puede
quedar empatado. Restaurar 82–80 permite guardar; cerrar por el fondo del
diálogo conserva el resultado sin enviarlo. Finalizar torneo abre confirmación
y, tras aceptarla, muestra Equipo local campeón. Cerrar el popup conserva
Torneo finalizado sin Editar resultado ni Finalizar torneo. Evidencia:
ios-basketball-tie-disabled.jpg, ios-basketball-champion.jpg e
ios-basketball-completed.jpg.

Retrospectiva: leer una final web desde iOS valida consulta y congelación, pero
no acredita su edición nativa. Separar ese recorrido de la validación de empate
y la finalización realmente ejecutadas mantiene el alcance verificable. Un
enlace accesible fuera del viewport puede abrir correctamente sin que el
control de gestos haya demostrado el desplazamiento de la lista.

### Intento Android y cierre de esta continuación

Se apaga iOS antes de abrir Android Studio. El IDE reabre el proyecto y
empieza a analizar archivos pese a Power Save Mode. Cerrar proyecto y
terminar su importación permite abrir Virtual Device Manager desde Welcome.
Pixel arranca, pero no muestra una ventana controlable. La lectura técnica
por adb confirma emulator-5554 online y sys.boot_completed=1; no acredita
ningún caso de interfaz. El IDE queda entonces en unos 280 MB residentes.

Reabrir el proyecto, detener Pixel desde Device Manager y arrancarlo desde
Running Devices recupera su pantalla de Inicio Android. Se restablecen
los forwards 8080/8082 hacia API y Metro. Separar Running Devices abre la
ventana del dispositivo, pero los siguientes eventos vuelven a la ventana
del proyecto y el IDE pierde respuesta fiable. No se acredita apertura de
FastTourney, login ni paginación. El análisis de símbolos C++ continúa y la
muestra del proceso Studio vuelve a 2,6 GB residentes y 665 % CPU; el
emulador ronda 540 MB y 31 % CPU. Se fuerza salida exclusivamente de Studio
desde Monitor de Actividad, sin eliminar cachés ni datos. La comprobación
posterior no encuentra procesos Studio ni qemu-system. Android sigue pendiente
por este bloqueo del entorno; no se registra como defecto del producto.

Se detiene Metro y se apagan API, PostgreSQL y Mailpit conservando volúmenes,
fixtures y capturas. No se activa observabilidad ni se toca producción. Esta
continuación modifica documentación, no código ni contratos. No se repiten
las suites anteriores ni se atribuyen sus contadores a una nueva ejecución.

Retrospectiva del entorno: cerrar el proyecto reduce el coste del IDE, pero
recuperar su visor puede volver a activar la importación y análisis. Un
arranque técnico correcto del emulador no elimina ese bloqueo. Antes de la
siguiente pasada Android habrá que disponer de una ventana controlable que
no dependa de cargar e indexar el proyecto completo.

Verificación documental: Prettier y git diff --check pasan. Compose no devuelve
servicios locales en ejecución al cerrar. No quedan procesos Android Studio ni
qemu-system; Device Hub confirma el iPhone apagado. Los resultados nuevos son
recorridos UI descritos en esta sección, sin cambios de implementación.

## Preparación Android independiente — 2026-10-06

El usuario indica usar el emulador y dejar el IDE. Se arranca directamente
Pixel_API_34 con el ejecutable oficial instalado y -no-snapshot-load, sin
Android Studio ni cambio permanente de configuración. iOS queda apagado.
La API local responde 200 en healthz y adb confirma emulator-5554 online con
sys.boot_completed=1; se restablecen los puertos 8080 y 8082 hacia el host.
El emulador selecciona renderizado por software por presión de memoria.
La muestra de qemu es 2,6 GB y 26,6 % CPU, sin proceso Studio.

El control nativo de Codex no reconoce el ejecutable independiente como app
controlable. Se solicita autorización específica para usar ADB en capturas,
lectura de controles y gestos. Hasta recibirla no se ejecutan operaciones de
interfaz por ese método ni se acredita ningún nuevo recorrido Android.
El arranque técnico no equivale a validar la app.

Cierre mientras falta esa autorización: se interrumpen Metro y el proceso
del emulador y se detienen API, PostgreSQL y Mailpit conservando volúmenes.
Retrospectiva: separar el emulador del IDE evita su indexación, pero la
herramienta de control disponible debe poder identificar su ventana; el
método alternativo se autoriza antes de usarlo y no se simula cobertura visual.

## QA Android con emulador independiente y ADB autorizado — 2026-10-06

Tras la autorización explícita del usuario se reabren Pixel_API_34, Metro y
API/PostgreSQL/Mailpit locales. Android Studio permanece cerrado y no se activa
observabilidad. ADB ejecuta gestos, entrada y enlaces; uiautomator y capturas
verifican el estado después de las acciones. Los campos de credenciales se
rellenan desde el fixture local ignorado sin publicar sus valores.

La app inicia sin sesión. Login del propietario abre Cuenta. En Torneos,
Administro muestra 50+; el desplazamiento alcanza Cargar más. Pulsarlo carga
las cuatro filas restantes y cambia el contador a 54. Al final se abre
QA Liga de equipos con acentos y nombres largos y se vuelve a la biblioteca.
Evidencia: android-library-bottom.png, android-library-54.png y
android-last-tournament.png. Este recorrido cierra la paginación visual Android
con la colección actual de 54; no reescribe la prueba histórica de 52.

Cuenta: Datos de acceso muestra el correo del propietario; Cambiar contraseña
vacío deshabilita Guardar contraseña. Vincular Google solicita reautenticación;
una contraseña ficticia incorrecta devuelve el mensaje común seguro y conserva
la sesión. Eliminar cuenta explica sesiones, relaciones y eliminación definitiva
tras 30 días; Cancelar conserva Cuenta y la sesión. El diálogo Android presenta
atenuación gris, contenido legible y botones completos sin recorte; acredita
el fallback visual declarado, no blur real. Evidencia: android-password-empty.png,
android-reauth-wrong-password.png y android-delete-confirmation.png.

Notificaciones del propietario muestra el estado vacío. Se cierra su sesión
mediante confirmación y se inicia la del participante. Los gestos alcanzan el
final de su colección de 31 notificaciones: la última, de 4/10/2026 16:34:26,
queda visible, abre la liga original y al volver se conserva el desplazamiento.
Evidencia: android-notifications-empty.png, android-notifications-bottom.png y
android-notification-destination.png. No se elimina ninguna notificación.

El participante ve Administro 23 y Sigo 5. En QA invitacion Android sin sesion,
Iniciar torneo está deshabilitado y explica que debe iniciarlo su propietario.
Dejar de seguir desde Acciones reduce Sigo a 4 y retira su tarjeta al volver.
Se abre nuevamente el fixture por su enlace local; Seguir torneo restaura Sigo 5
y la tarjeta. Evidencia: android-unfollow-library.png,
android-unfollowed-detail.png y android-follow-restored.png. La prueba es una
mutación reversible nativa con verificación de biblioteca, no solo de etiqueta.

Resultados nativos: QA formulario volleyball abre su final 3–2 con cinco sets.
Cambiar 16–14 a 15–14 muestra la validación localizada y deshabilita Guardar.
Con teclado visible el diálogo se desplaza; ocultarlo deja el formulario,
mensaje y botón completos. Restaurar 16–14 habilita Guardar; se confirma la
escritura desde Android. Finalizar torneo abre la advertencia de congelación;
confirmarlo muestra Equipo local como campeón. Tras cerrar, el estado es
Finalizado y desaparecen Editar resultado y Finalizar torneo. Evidencia:
android-volleyball-invalid-fifth.png,
android-volleyball-invalid-keyboard-hidden.png,
android-volleyball-valid-recovered.png, android-volleyball-champion.png y
android-volleyball-completed.png. La preparación y el marcador previo eran
fixtures web; edición válida y finalización sí se ejecutan en Android.

QA revisión descansos, preparado y finalizado previamente en web, abre su cuadro
nativo de cinco equipos. Cuartos presenta el partido disputado y los pases con
«Pase automático. No se disputa partido.». El enlace de un pase conduce a
Semifinales · Partido 1; se ven enlaces de origen y destino a Final. Este
recorrido acredita lectura y navegación de rondas, no jugar todo el torneo
por Android. Evidencia: android-byes-completed.png,
android-byes-direct-passes.png, android-byes-semifinals.png y android-byes-final.png.

Retrospectiva de fase: el emulador independiente permite validar los recorridos
que el IDE bloqueaba. La paginación necesita una acción explícita al final del
bloque; el desplazamiento solo no carga más. La recuperación de un formulario
se prueba pasando de inválido a válido y confirmando la operación. Se conserva
la distinción entre fixtures web y acciones nativas. No se observan nuevos
defectos de producto en este corte; la matriz global conserva los casos no
recorridos y OAuth real diferido por el usuario.

Cierre: Metro y Pixel se detienen; Compose no devuelve servicios locales en
marcha y no quedan procesos Studio ni qemu. Se conservan volúmenes, fixtures
y evidencia. Prettier y git diff --check pasan. Este corte actualiza informe y
aprendizaje; no modifica implementación ni repite las suites anteriores.

## Continuación Android: incidencias, tanda y tema/idiomas — 2026-10-06

Se reabre Pixel_API_34 sin Android Studio, junto con API/PostgreSQL/Mailpit
locales y Metro. No se activa observabilidad. La sesión del participante
sobrevive al apagado y nuevo arranque del emulador.

QA formulario handball parte de 25–25 y tanda 5–4. Desde el editor Android se
registra No se presenta del visitante; la card muestra incomparecencia y
victoria local. Se corrige a Abandono del visitante. Introducir solo 12 goles
locales bloquea Guardar y solicita completar ambos tanteos o dejar el parcial
vacío. Completar 12–10 recupera Guardar; la escritura devuelve abandono,
victoria local y el parcial correcto. Se vuelve a Marcador, se restaura 25–25
y se comprueba que tanda vacía y tanda 4–4 deshabilitan Guardar. Con 5–4 se
habilita; guardar restaura el partido jugado, sin etiqueta ni parcial de
incidencia. El fixture queda con su resultado original y sigue en curso.
Evidencia: android-handball-no-show-saved.png,
android-handball-retirement-saved.png,
android-handball-tie-empty-shootout.png, android-handball-shootout-tie.png,
android-handball-shootout-valid.png y android-handball-played-restored.png.

El botón fijo Finalizar puede cubrir parte del editor mientras una card pasa
por el borde inferior. Desplazar al final deja Editar completo por encima de
la acción fija; el recorrido conserva acceso a ambos controles. No se toma
el estado intermedio del scroll como una imposibilidad de usar el editor.

Ajustes cambia a Oscuro. Cuenta, biblioteca y confirmación de logout muestran
superficies oscuras y texto/botones legibles; Cancelar preserva sesión. Evidencia:
android-account-dark.png, android-library-dark.png y android-logout-dark.png.
Se usa el ajuste de idioma del propio emulador, sin selector dentro de la app.
Inglés y francés como primer idioma del sistema recrean la navegación en Inicio,
conservan sesión, recientes y preferencia oscura, y traducen títulos, tabs,
cuenta y confirmación. Las confirmaciones se cancelan. Evidencia:
android-home-en-dark.png, android-logout-en-dark.png,
android-home-fr-dark.png y android-logout-fr-dark.png. Esto es cobertura visual
parcial por idioma; no acredita todas las rutas ni un análisis cuantitativo
de contraste o TalkBack.

Italiano repite Inicio, Cuenta y confirmación cancelada con traducción correcta,
sesión y tema conservados. Evidencia: android-home-it-dark.png y
android-logout-it-dark.png. Se eleva font_scale del emulador de 1.0 a 1.3;
la recreación de actividad vuelve a Inicio sin cerrar la sesión. La ampliación
se comprueba en los textos y tabs, y se recorren Inicio, Cuenta y confirmación:
el mensaje ocupa dos líneas y los botones permanecen completos y utilizables.
Evidencia: android-home-it-dark-large.png, android-account-it-dark-large.png y
android-logout-it-dark-large.png. No acredita escalas superiores ni todas las
pantallas autenticadas. Se restaura font_scale=1.0, español como único idioma
y se retiran los tres idiomas añadidos durante QA.

Retrospectiva: una incidencia no termina al guardar: hay que reabrirla, probar
un parcial incompleto, recuperar uno válido y corregirla a resultado jugado.
Un cambio de idioma o escala puede recrear la actividad y la navegación;
conservar la sesión no implica conservar la ruta. La captura verifica el estado
resultante antes de la siguiente acción. Se actualiza la matriz inicial para
que los pendientes históricos no contradigan la evidencia Android reciente.

Cierre de fase: la app vuelve a español, apariencia Sistema y escala 1.0
(android-restored-es-system.png). Se apagan Pixel, Metro y API/PostgreSQL/Mailpit,
sin procesos del IDE ni servicios locales en marcha; se conservan volúmenes y
evidencia. Prettier y git diff --check pasan. No se modifica código ni se
repiten suites de implementación. No aparecen nuevos defectos en los casos
recorridos; la matriz global sigue abierta con su cobertura parcial explícita.

## Continuación Android: recuperación y colecciones vacías — 2026-10-06

Se usa Pixel_API_34 independiente, sin IDE ni observabilidad. Al arrancar aparece
un informe antiguo del development client, fechado el 4 de octubre, sobre el
contexto React. Force-stop y apertura de la URL exacta de Metro recuperan la app
con sesión conservada; no se atribuye ese informe histórico a un defecto nuevo.

Recuperación bloquea el envío vacío y muestra validación localizada al abandonar
un correo inválido. Corregirlo habilita la solicitud, que termina con el mensaje
seguro sin enumerar cuentas. Mailpit confirma recepción para el fixture. El enlace
real abre el formulario con el correo correcto y Guardar deshabilitado sin nueva
contraseña; no se confirma ningún cambio de credencial. Un enlace ficticio inválido
muestra el mensaje seguro de expiración y permite volver a Inicio. El enlace real
se conserva privado y no se incluye en el informe ni en Git.

El login de una cuenta verificada sin torneos permite comprobar Administro 0 y
Sigo 0, sus mensajes distintos, acceso a crear torneo y notificaciones vacías con
retorno. El segundo fixture también permite login: su estado observado es
verificado y no acredita login pendiente ni reenvío móvil.

Con API detenida, Recuperación mostraba `common_request_error` pese al rechazo de
transporte identificado por `apiFetch`. La pantalla ahora reutiliza
`getRequestFailure`: conexión usa `common_network_error`, 429 conserva su mensaje
localizado y los demás fallos conservan el fallback seguro. No se exponen cuerpos
ni detalles del backend. Android comprueba el mensaje corregido y el correo
conservado; levantar API y reintentar en el mismo formulario devuelve éxito y
login. Se restaura la sesión del participante con su credencial existente.
Evidencia privada: android-recovery-offline-fixed.png,
android-recovery-draft-preserved.png, android-recovery-retry-success.png,
android-recovery-returned-login.png y android-participant-restored.png.

Checklist del cliente: cambio limitado al feedback de forgot-password, sin nuevas
primitivas, estilos, textos literales ni reglas de negocio. Se reutilizan los
catálogos de los cuatro idiomas y el clasificador compartido aceptado. La operación
sigue pasando por requestRecovery, cliente generado y apiFetch. Se revisan contrato
y el inventario de POST /v1/password-resets: éxito 202, validación 400, tasa 429,
fallos DB/SMTP, timeout y cancelación conservan el comportamiento documentado;
esta pasada no vuelve a ejecutar todas esas salidas. No se cambia endpoint ni
instrumentación. No introduce una decisión arquitectónica que necesite nuevo ADR.
`pnpm run check` pasa formato, lint, TypeScript y OpenAPI. La exportación web pasa
con 36 rutas estáticas; no se presentan las suites históricas como reejecutadas.

Retrospectiva: probar desconexión en cada operación detecta pantallas que todavía
usan un fallback local. La solución mínima reutiliza la clasificación de transporte
existente, manteniendo el mapeo útil de 429 en la feature. Un fixture supuestamente
pendiente se clasifica por su comportamiento observado. Quedan pendientes el cambio
real de contraseña y el resto explícito de la matriz global; la recuperación iOS se amplía en la continuación posterior.

Cierre: Metro y Pixel se detienen; Compose confirma API/PostgreSQL/Mailpit
apagados y no quedan procesos de emulador ni IDE. Se conservan volúmenes,
fixtures y evidencia. Prettier y git diff --check pasan. La matriz global
permanece abierta; no se despliega ni publica el cambio.

## Continuación Android: creación, composición y final de fútbol — 2026-10-06

Pixel_API_34 arranca independiente del IDE con API/PostgreSQL/Mailpit y Metro;
no se activa observabilidad. El arranque en frío invalida el snapshot anterior
por cambio del renderer. Se espera a la pantalla usable, no solo al flag de boot,
y se abre la URL exacta del development client. La sesión del participante sigue
vigente. Esta circunstancia del entorno no se registra como defecto de producto.

Desde Inicio se abre Crear torneo. Enviar con nombre vacío muestra su error
localizado; corregirlo a QA creacion Android 20261006 permite crear fútbol con
el primer equipo, Web Local, precompletado por la preferencia existente.
La pantalla de composición muestra Torneo sin empezar y bloquea Iniciar con
un solo equipo. Añadir equipo vacío bloquea Guardar; duplicar Web Local devuelve
«Ya existe un equipo con ese nombre.» dentro del diálogo. Corregir a Android
Visitante permite guardar y habilita Iniciar.

Se desplaza hasta la segunda fila para que su acción no quede tras el botón
fijo. Eliminar abre confirmación con el nombre correcto. Cancelar mantiene los
dos equipos; repetir y confirmar retira al visitante, devuelve el requisito de
dos equipos y deshabilita Iniciar. La fila del último equipo ya no ofrece Eliminar.
Se repone Android Visitante mediante el mismo formulario. Evidencia privada:
android-create-required.png, android-created-one-team.png,
android-team-duplicate.png, android-team-delete-cancelled.png y
android-team-deleted.png. Esto cierra baja previa al inicio Android, no retirada
de un equipo durante una liga ni bajas en las otras plataformas.

Se selecciona Eliminatorias y se inicia desde Android: aparece una única Final.
En Añadir resultado, 1–1 exige una tanda y mantiene Guardar deshabilitado sin
ambos valores. Completar 5–4 habilita Guardar; la card muestra 1 (5) y 1 (4),
explica los paréntesis y marca ganador al local. Finalizar solicita confirmación;
confirmar muestra Web Local como campeón. Cerrar el popup deja Torneo finalizado
sin Editar resultado ni Finalizar. Volver a Inicio lo muestra en recientes;
reabrirlo conserva el estado terminal y el resultado. Todas estas operaciones
se ejecutan en UI Android, sin preparación deportiva mediante API. Evidencia:
android-football-tie-required.png, android-football-penalties-saved.png,
android-creation-champion.png, android-creation-completed.png,
android-created-recent.png y android-created-reopened.png.

Retrospectiva: la composición se prueba como transición 1 → 2 → 1 → 2 equipos,
observando habilitación y disponibilidad de acciones en cada paso. Cancelar
la confirmación y confirmar después son casos distintos. Un resultado con tanda
necesita observar marcador, ganador y estado al reabrir; el popup de éxito solo
no prueba persistencia. El teclado virtual puede cubrir una acción que todavía
aparece en la jerarquía accesible: la captura decide si es pulsable y se oculta
el teclado antes de continuar. No se detectan nuevos defectos en esta fase;
la cobertura restante de deportes, byes jugados y otras plataformas sigue abierta.
No se modifica implementación ni se repiten suites del corte anterior.

Cierre: Metro y Pixel detenidos; API/PostgreSQL/Mailpit apagados y sin procesos
del emulador ni IDE. Se conservan los volúmenes y el nuevo torneo ficticio terminado
para futuras lecturas. Prettier y git diff --check pasan.

## Continuación Android: retirada durante una liga — 2026-10-06

Se reanudan API/PostgreSQL/Mailpit, Pixel_API_34 y la development build instalada.
La revisión automática rechaza Metro LAN por exposición del bundle a otros equipos;
se usa `--localhost`, API IPv4 loopback y forwards ADB 8080/8082. No se activa
observabilidad. El control nativo no reconoce el emulador independiente. Se
recupera de la conversación anterior la autorización explícita del usuario para
capturas, lectura y gestos ADB. El intento de abrir Studio se cierra, incluida su
importación, y la prueba continúa sin IDE. Al cerrarlo desaparecen los forwards;
la solicitud inicial de creación falla con el mensaje seguro de conexión. Se
restablecen y se vuelve a abrir la app. No se atribuye ese fallo de entorno al
producto ni se acredita creación por UI en esta fase.

Se prepara por API un fixture nuevo del participante: QA retirada Android
20261006, fútbol con Android Local, Android Retirado y Android Tercero. Android
conserva la configuración de liga de una vuelta e inicia el torneo. Desde el
editor nativo guarda Android Retirado 2–1 Android Tercero. En Equipos, Retirar
explica el nombre y el efecto 3–0 para todos sus rivales. Cancelar conserva los
tres equipos y el 2–1, comprobado al regresar al torneo.

Repetir y confirmar marca Android Retirado como Retirado y elimina su acción de
baja. El partido jugado pasa a 0–3; el pendiente contra Android Local pasa a 3–0.
Ambos dejan de ofrecer edición. Android Local–Android Tercero conserva su estado
pendiente y abre su editor con Guardar deshabilitado mientras está vacío; se
cierra sin enviar. La liga permanece en curso. La clasificación concede tres
puntos y una victoria a cada rival; el retirado tiene dos derrotas y cero puntos.
Volver a Inicio y reabrir desde recientes conserva los resultados y su protección.
No se prueba en esta fase una nueva escritura tras la retirada ni el historial
mediante UI. La cobertura de retirada nueva iOS continúa pendiente.

Evidencia privada en el directorio QA existente: android-withdrawal-played.png,
android-withdrawal-confirmation.png, android-withdrawal-cancelled.png,
android-withdrawal-cancelled-score.png, android-withdrawal-confirmed.png,
android-withdrawal-results.png, android-withdrawal-standings.png,
android-withdrawal-reopened.png y android-withdrawal-active-editor.png.

Retrospectiva: usar tres equipos distingue el partido jugado, el pendiente del
retirado y el no afectado. Cancelar y confirmar son recorridos separados; reabrir
desde recientes comprueba persistencia. Un fallo de transporte exige verificar
los forwards actuales, incluso si la API sigue respondiendo en el host. La
preparación por API se documenta aparte de las acciones deportivas nativas.
No se detectan defectos nuevos ni se modifica implementación. No se repiten las
suites históricas; la matriz global conserva sus restantes pendientes.

Cierre: Android vuelve a Inicio; se detienen Pixel, Metro y API/PostgreSQL/Mailpit.
Docker y la lista de procesos confirman el apagado, incluido el IDE. Se conservan
volúmenes, fixture y capturas; no se despliega ni publica. Prettier y
`git diff --check` pasan.

## Continuación Android: tres botones y orientación — 2026-10-06

Se usa Pixel_API_34 independiente, con los servicios locales de QA y Metro,
sin observabilidad. Se registra la configuración inicial: overlay gestual,
accelerometer_rotation=1, user_rotation=0 y sensor 0:9.77631:0.812349.
Activar el overlay de tres botones conserva sesión y muestra la barra nativa
separada de las tabs. En biblioteca Administro 25 se llega por gestos hasta
la última fila, QA Liga de equipos con acentos y nombres largos: queda completa,
por encima de la acción flotante y las tabs. Pulsarla abre el destino. El botón
Atrás del sistema devuelve la biblioteca conservando el desplazamiento.
Evidencia privada: android-threebutton-home.png,
android-threebutton-library-bottom.png, android-threebutton-last-destination.png
y android-threebutton-return.png. El contador actual pertenece a los fixtures
existentes; no se presenta como la paginación de más de 50 de otra cuenta.

En Cuenta, Atrás cierra la confirmación de logout conservando sesión. En Crear
torneo, con teclado virtual y texto de prueba en el borrador, la primera pulsación
oculta el teclado y conserva la ruta/texto. Se restaura el borrador previo; otra
pulsación, sin teclado, cierra la ruta y vuelve a Inicio. No se envía la creación.
Evidencia: android-threebutton-dialog-back.png, android-threebutton-keyboard.png,
android-threebutton-keyboard-back.png y android-threebutton-form-back.png.

La petición inicial de rotación por settings no acredita un giro efectivo:
user_rotation vuelve a 0 y la pantalla sigue vertical. Se libera la rotación y
se cambia el sensor a 9.77631:0:0.812349. Ajustes del sistema aparece horizontal
con mCurrentOrientation=1; abrir la app vuelve a mCurrentOrientation=0 y Cuenta
vertical con sesión conservada mientras el sensor sigue lateral. La build se
comporta de acuerdo con orientation: portrait de app.config.ts. Esto comprueba
el bloqueo vertical observado, no un layout horizontal ni su soporte en todos los
SO. Evidencia: android-sensor-landscape-system.png y
android-sensor-portrait-account.png. La captura inicial llamada
android-system-landscape.png permanecía vertical y no se usa para acreditar giro.

Retrospectiva: una preferencia solicitada no equivale a la orientación efectiva;
se contrasta sensor, una pantalla del sistema y la app. Atrás tiene responsabilidades
distintas con teclado, diálogo y ruta: cada estado requiere su propia observación.
La reserva inferior se contrasta con el último elemento visible y su destino.
No se detectan nuevos defectos ni se modifica implementación; quedan pendientes
TalkBack, otras versiones, release, dispositivos físicos y las rutas no recorridas.
Se restaura overlay gestual, sensor inicial, user_rotation=0 y rotación automática
antes del apagado. No se repiten suites anteriores.

Cierre: Pixel y Metro detenidos; Compose sin API/PostgreSQL/Mailpit en marcha
y sin procesos del emulador ni IDE. Se conservan volúmenes, fixtures y evidencia.
Prettier y git diff --check pasan; no se despliega ni publica.

## Continuación Android: byes jugados e incidencias de tenis — 2026-10-06

Pixel_API_34 independiente, sin IDE, usa la development build instalada con
Metro/API IPv4 loopback y forwards ADB. API/PostgreSQL/Mailpit se levantan sin
observabilidad. Se conserva la autorización explícita de control ADB de la
sesión anterior. No se cambian dependencias ni se despliega.

Se prepara por API QA byes Android 20261006, fútbol con cinco equipos. Desde
Android se elige Eliminatorias y se inicia. Cuartos muestra Equipo 1–Equipo 5
como único partido disputado y tres pases directos para 2, 3 y 4, con mensaje
Pase automático y sin editor. El enlace del bye 2 abre semifinales: Equipo 2
ya ocupa su plaza, el rival figura Ganador por determinar y esa semifinal no
ofrece Añadir resultado. El enlace de origen vuelve al partido de cuartos.

Todas las escrituras deportivas siguientes se hacen desde UI Android: cuartos
1–5 termina 2–0; semifinales 1–2, 1–0; semifinales 3–4, 0–1; final 1–4, 3–1.
Se comprueba ganador en cada card y los enlaces entre rondas. La confirmación de
finalización muestra Android Equipo 1 campeón. Cerrar deja el estado finalizado
sin Editar ni Finalizar; volver a Inicio y reabrir desde recientes conserva el
estado y el 2–0 de cuartos sin editor. Los byes jugados iOS siguen pendientes.

QA formulario tennis, al mejor de cinco sets, parte de 7–6, 7–5, 6–4. Android
cambia a Incidencia y elige visitante ausente; sin lado afectado Guardar está
bloqueado y al elegirlo aparece Victoria para Equipo local. Guardar muestra
No se presenta, motivo y ganador sin presentar sets administrativos como jugados.
Reabrir conserva esa incidencia. Se corrige a Abandono del visitante. Introducir
solo 2 en el primer lado del set bloquea Guardar y pide completar ambos tanteos
o dejar el parcial vacío. Completar 2–3 habilita Guardar; el scroll permite
alcanzar error y acción. La card guarda Abandono de Equipo visitante, Victoria
para Equipo local y Tanteo parcial: 2–3. Reabrir conserva ambos valores. Se
cierra sin otra escritura y se vuelve a Inicio. El fixture queda en curso con
ese abandono para posteriores pruebas; no se acredita corrección a jugado en
esta fase ni incidencias de tenis iOS.

Evidencia privada en el directorio QA: android-byes-quarterfinal.png,
android-byes-auto-2-3.png, android-byes-auto-4.png,
android-byes-semifinal-pending.png, android-byes-quarter-saved.png,
android-byes-semi1-saved.png, android-byes-semi2-saved.png,
android-byes-final-ready.png, android-byes-final-saved.png,
android-byes-champion.png, android-byes-reopened.png,
android-tennis-no-show-ready.png, android-tennis-no-show-saved.png,
android-tennis-partial-incomplete.png, android-tennis-partial-valid.png,
android-tennis-retirement-saved.png y android-tennis-retirement-reopened.png.

Retrospectiva: leer un cuadro terminado no acredita su propagación por escritura
nativa. Un número impar permite separar pases automáticos, plazas pendientes y
partidos disputados. En un abandono, elegir un parcial desfavorable al ganador
administrativo verifica que el tanteo real no decide la victoria. La captura
visual y el scroll comprueban la acción disponible; un control incluido en la
jerarquía fuera del viewport no demuestra que se pueda pulsar.
No aparecen defectos nuevos ni cambios de implementación. No se repiten suites
históricas. Pixel se apaga antes de continuar con iOS; el stack de pruebas sigue
activo para la siguiente fase.

## Continuación iOS: recuperación, transporte y enlaces — 2026-10-06

Se usa iPhone 18 Pro con iOS 27.0 y la development build instalada. Metro arranca
con `--localhost`, sin observabilidad. Su listener loopback IPv6 y la URL IPv4
anunciada inicialmente impiden cargar el bundle. Arrancar con
`REACT_NATIVE_PACKAGER_HOSTNAME=localhost` permite cargarlo manteniendo loopback;
es un ajuste temporal de la sesión, sin modificar la configuración del proyecto.
Android ya está apagado. API/PostgreSQL/Mailpit permanecen locales.

La sesión del participante se conserva al cargar. Desde Cuenta se cierra sesión
con confirmación y se abre Recuperación. El correo vacío bloquea Enviar. Al
abandonar un correo mal formado aparece el error localizado y el envío sigue
bloqueado; corregir a la cuenta ficticia habilita la acción.

Detener API y enviar muestra el mensaje común de conexión, sin detalles internos.
El formulario conserva el correo. Levantar API y reintentar desde ese formulario
termina con el mensaje seguro de solicitud y vuelve a login. Mailpit acredita
recepción local para la cuenta ficticia; no demuestra entrega externa.

Safari abre un enlace nativo ficticio inválido: la app muestra enlace inválido
o caducado y Volver a la home recupera Inicio. Después, una página temporal
servida exclusivamente en loopback permite abrir el enlace real recibido en
Mailpit. La app muestra Crea una contraseña nueva, el correo correcto no editable,
contraseña vacía y Guardar contraseña deshabilitado. No se introduce una credencial
nueva ni se acredita el cambio efectivo: requiere intervención humana conforme a
la política de control de escritorio. El token y la página quedan en evidencia
privada, fuera de Git. Estas pantallas se observaron mediante Device Hub; no se
atribuyen archivos de captura nuevos a esta fase.

Retrospectiva: verificar el transporte de Metro por separado del API evita
atribuir a la app un desacuerdo entre hostname anunciado y listener. Una
recuperación completa necesita distinguir solicitud segura, entrega local,
apertura nativa, formulario y cambio efectivo. Llegar al formulario correcto
no acredita cambiar la contraseña. El corte y reintento en iOS confirman también
el feedback de transporte corregido previamente en Android.

No aparecen defectos nuevos ni se modifica implementación. Se conserva el alcance
pendiente de la matriz, incluyendo cambio de credencial, login pendiente Android,
cobertura completa de formatos nativos y accesibilidad
con lectores de pantalla. No se repiten suites históricas.

## Continuación iOS: verificación, vacíos, byes y retirada — 2026-10-06

Se prepara por API una cuenta ficticia pendiente local, qa_ios_pending_d0c222,
con credencial existente para las pruebas de login. La preparación por API no
acredita registro mediante UI ni aceptación de condiciones en pantalla. Dos
intentos de login desde iOS muestran el mensaje de nuevo enlace enviado y
conservan el formulario sin abrir Cuenta autenticada. Mailpit contiene el correo
inicial y ambos reenvíos.

El último enlace se abre desde Safari mediante una página privada en loopback.
La app muestra verificación en curso, vuelve a Inicio y Cuenta identifica el
username esperado. Reabrir el mismo enlace muestra que ya se ha utilizado;
Volver a la home recupera Inicio y la sesión sigue en Cuenta. Con esa cuenta se
comprueban actividad reciente vacía, Administro 0 y Sigo 0 con mensajes distintos,
acción de creación disponible, notificaciones vacías y retorno a Cuenta.
Las credenciales y los tokens permanecen en fixtures privados ignorados por Git.

Se prepara por API QA byes iOS 20261006, fútbol con tres equipos. Desde iOS se
selecciona Eliminatorias y se inicia. Semifinales contiene Equipo 1–Equipo 3 y
un pase directo de Equipo 2 sin editor. Su enlace abre una final con Equipo 2 y
Ganador por determinar, sin Añadir resultado. El enlace de origen vuelve al
pase automático y lo presenta en pantalla. Se escribe semifinal 2–0 desde UI,
se comprueba propagación a la final y se escribe 1–3. La confirmación de
finalización muestra iOS Equipo 2 campeón. Cerrar conserva Torneo finalizado,
sin Editar ni Finalizar. Inicio presenta el torneo finalizado en recientes;
reabrir conserva el estado, el 2–0 y el pase directo sin editor.

QA retirada iOS 20261006 se prepara por API como liga de tres equipos con un
partido jugado: iOS Retirado gana 2–1 a iOS Activo. iOS abre la retirada de ese
equipo y cancela; volver conserva el marcador. Repetir y confirmar muestra el
equipo Retirado, sustituye el jugado por 0–3 y resuelve el pendiente contra iOS
Local en 3–0, sin editores para ambos. La clasificación da tres puntos a cada
rival activo y cero al retirado. El partido iOS Local–iOS Activo sigue pendiente
y permite abrir el editor vacío con Guardar bloqueado. Se cierra tocando el fondo
sin escribir; Inicio y reapertura conservan los tres partidos y sus controles.
El fixture queda en curso: no se acredita escritura del partido no afectado ni
finalización de esta liga. Una lectura independiente del API conserva evidencia
de ambos fixtures tras las escrituras nativas.

El control accesible de Device Hub permite seleccionar y traer a pantalla
controles fuera del viewport; los intentos de scroll/drag genérico no acreditan
desplazamiento táctil. Por ello esta fase no cierra el pendiente de desplazamiento
de notificaciones iOS ni acredita VoiceOver. Las pantallas se observaron en la
conversación; no se inventan archivos de captura para esta fase.

Retrospectiva: una cuenta verificada recientemente permite cubrir vacíos sin
alterar colecciones de fixtures previos. El enlace consumido debe conservar la
sesión establecida. Un cuadro de tres equipos prueba bye, plaza pendiente,
propagación y victoria del equipo con pase directo con solo dos escrituras.
La retirada se prueba con tres equipos para separar partido jugado, pendiente y
no afectado, comprobando además cancelación y persistencia tras reapertura.
No aparecen defectos nuevos ni cambios de implementación en esta fase.

Cierre de sesión: iOS queda en Inicio con la cuenta ficticia nueva; no se restaura
la sesión anterior del participante. Android, iPhone, Metro, servidor temporal,
API, PostgreSQL y Mailpit se apagan conservando volúmenes y fixtures. No se inicia
observabilidad ni se toca producción. El cambio efectivo de contraseña permanece
pendiente de intervención humana; el formulario real se probó vacío antes de
continuar con los demás casos. Prettier y git diff --check pasan sobre la
actualización documental. Se conserva el cambio previo de feedback en
forgot-password sin modificaciones en esta fase.

## Continuación Android: sets de tenis y pádel — 2026-10-06

Se reabre Pixel_API_34 independiente, API/PostgreSQL/Mailpit y Metro, sin IDE ni
observabilidad. La sesión del participante se conserva. Se revisan las reglas
aceptadas de ADR-0135 y ADR-0142; no se propone ni cambia una regla deportiva.
Los fixtures existentes ya estaban iniciados: este corte no acredita creación,
inicio ni finalización de estos deportes desde Android.

QA formulario tennis está al mejor de cinco y parte de abandono del visitante
con parcial 2–3, victoria local. Abrir Editar recupera ese parcial. Cambiar a
Marcador conserva los campos y bloquea Guardar mientras el resultado es incompleto.
Se introducen 6–0, 6–5 y 6–4: el segundo set es inválido y Guardar permanece
bloqueado con feedback localizado. Corregir a 7–5 habilita Guardar; el guardado
nativo muestra 3–0, sets 6–0 · 7–5 · 6–4 y ganador local, sin etiqueta de abandono.
Reabrir Editar recupera los tres sets. Añadir un cuarto 6–0 después de las tres
victorias bloquea Guardar. Se limpian los sets sobrantes, se vuelve a Incidencia,
se selecciona abandono del visitante y se restaura el parcial 2–3. Desplazar el
diálogo permite observar victoria y Guardar; confirmar devuelve el estado original.
Evidencia privada: android-tennis-initial.png, android-tennis-invalid-set.png,
android-tennis-valid-sets.png, android-tennis-played-saved.png,
android-tennis-played-reopened.png, android-tennis-extra-set-blocked.png y
android-tennis-retirement-restored.png.

QA formulario padel está al mejor de tres, con 6–4 y 7–6 y victoria local 2–0.
Editar recupera dos sets completos y el tercero vacío. Cambiar el segundo a 6–6
bloquea Guardar; volver a 7–6 lo habilita. Añadir un tercer 6–0 después de las
dos victorias vuelve a bloquearlo. Vaciar ese set recupera la acción y el guardado
nativo conserva 6–4 · 7–6, 2–0 y ganador local. Volver a Inicio y abrir desde
recientes conserva la lectura. Evidencia: android-padel-initial.png,
android-padel-editor-initial.png, android-padel-tied-set-blocked.png,
android-padel-valid-recovered.png, android-padel-extra-set-blocked.png,
android-padel-original-recovered.png, android-padel-saved.png y
android-padel-reopened.png.

Una lectura independiente del API confirma ambos fixtures en curso: tenis con
incident retirement/away y partialSets 2–3; pádel sin incidencia, con sets 6–4 y
7–6 y agregado 2–0. La evidencia reducida queda en android-racket-final-api.json,
sin credenciales ni tokens. Los marcadores originales se conservan, con el
historial real de las escrituras de QA. No se confirma Finalizar en ninguno.

Retrospectiva: validar cada set no basta; también hay que detener la secuencia
cuando se alcanza el número de victorias. Se prueba ese límite con cinco y tres
sets por separado. Convertir abandono a jugado debe retirar su metadata de la
card; volver a abandono conserva el parcial como dato distinto del agregado
administrativo. La altura del diálogo cambia al desaparecer el error: se observan
las nuevas posiciones antes de pulsar Guardar. La restauración se verifica en UI
y lectura independiente. No se detectan nuevos defectos ni se modifica código;
no se repiten suites anteriores. La matriz global conserva los formatos y
recorridos no cubiertos.

Cierre: Android queda en Inicio con la sesión conservada antes de apagar Pixel
y Metro. API/PostgreSQL/Mailpit apagados y sin procesos del emulador ni IDE.
Se conservan volúmenes, fixtures y capturas. Prettier y git diff --check pasan;
no se despliega ni publica.

## Continuación Android: bádminton a 15 puntos — 2026-10-06

Se usa Pixel_API_34 independiente con API/PostgreSQL/Mailpit y Metro, sin IDE
ni observabilidad. La sesión del participante se conserva. ADR-0141 aceptado
establece la elección persistida por torneo: para juegos a 15, diferencia de
dos hasta el tope de 21 y excepción 21–20. No se cambia esa decisión ni código.
QA formulario badminton está en curso, al mejor de tres juegos a 15, con
marcador inicial 21–20 · 15–12 y victoria local 2–0. Abrir Editar recupera
los juegos y muestra las reglas del perfil correcto.

Cambiar el primer juego a 15–14 bloquea Guardar con feedback localizado.
Corregir a 16–14 recupera la acción; guardar desde Android muestra
16–14 · 15–12 y conserva ganador local. Reabrir y escribir 22–20 bloquea
Guardar pese a la ventaja de dos: supera el tope. Cambiar a 20–19 sigue
bloqueado por ventaja insuficiente antes del tope. Corregir a 21–20 habilita
Guardar por la excepción prevista. Evidencia privada:
android-badminton-initial.png, android-badminton-one-point-blocked.png,
android-badminton-two-point-valid.png, android-badminton-two-point-saved.png,
android-badminton-over-cap-blocked.png,
android-badminton-before-cap-one-point-blocked.png y android-badminton-cap-valid.png.

Añadir un tercer juego 15–0 cuando el local ya ganó los dos primeros bloquea
Guardar. Vaciarlo recupera la acción; guardar restaura 21–20 · 15–12. Inicio
presenta el torneo en recientes y reabrirlo conserva ambos juegos y ganador.
Una lectura independiente del API confirma in_progress, pointsPerGame=15,
bestOfSets=3 y agregado 2–0. Evidencia: android-badminton-extra-game-blocked.png,
android-badminton-original-recovered.png, android-badminton-original-saved.png,
android-badminton-reopened.png y android-badminton-final-api.json. La finalización
no se confirma y el historial conserva las escrituras reales de QA.

Se detecta una incoherencia de copy: racket_result_invalid decía «sets necesarios»
aunque bádminton presenta juegos. El mensaje compartido ahora pide un resultado
válido y detenerse cuando un participante gane el partido. Se actualizan es/en/it/fr,
sin claves nuevas ni cambios de reglas. Android con 15–14 reproduce el bloqueo y
muestra la redacción corregida (android-badminton-feedback-fixed.png). Cerrar sin
guardar conserva 21–20 · 15–12 (android-badminton-after-feedback-close.png).

Checklist del cliente: solo cambia el catálogo de los cuatro locales; se reutilizan
clave, primitivas, estilos y validadores actuales. Se aplica el vocabulario aceptado
de ADR-0141 y los catálogos de ADR-0054/0055/0056. No se modifica endpoint, adaptador,
transporte, DTO ni instrumentación; no requiere una decisión arquitectónica nueva.
`pnpm run check` pasa formato, lint, TypeScript y OpenAPI; exportación web de 36
rutas completada. El primer check detectó únicamente el salto final regenerado
por Expo en expo-env.d.ts; se normaliza y se repite satisfactoriamente. No se
presentan las suites históricas como reejecutadas.

Retrospectiva: un límite máximo puede admitir una ventaja distinta del caso
normal. Se separan ventaja insuficiente antes del tope, excepción en el tope,
valor superior al tope y secuencia posterior a la victoria. Compartir un feedback
no debe imponer a otro deporte su vocabulario: una redacción común puede expresar
el fin del partido, dejando el detalle en la ayuda del perfil. Esta pasada acredita
el perfil a 15; el perfil a 21 con tope 30 no se presenta como probado aquí. La
matriz global sigue abierta con los recorridos restantes.

Cierre: app en Inicio antes de apagar Pixel y Metro. API/PostgreSQL/Mailpit
apagados; no quedan procesos del emulador ni IDE. Se conservan volúmenes,
fixtures y evidencia. Prettier y git diff --check pasan. No se despliega ni publica.

## Continuación Android: abandono en tenis de mesa a siete juegos — 2026-10-06

Se reabre Pixel_API_34 independiente con API/PostgreSQL/Mailpit y Metro, sin IDE
ni observabilidad. Tras observar Inicio autenticado, abrir el enlace del torneo
lleva al development launcher. Su informe muestra fecha 6 de octubre, 18:20:17,
y «App react context shouldn't be created before», con stack de creación de
actividad en Expo DevLauncher. Es una incidencia nueva observada del entorno de
development build; no se descarta como el informe histórico del 4 de octubre.
Force-stop y apertura de la URL exacta de Metro recuperan Inicio y sesión. Se
entra al torneo mediante biblioteca y gestos, sin repetir el enlace. Evidencia:
android-table-tennis-devlauncher-crash.png y android-table-tennis-recovered.png.
El workaround permite seguir el QA, pero no resuelve la caída ni acredita su
comportamiento en release. El caso de arranque/enlace queda abierto para diagnóstico.

QA formulario table_tennis está en curso al mejor de siete juegos, con abandono
del visitante y parcial 5–3. El editor carga ese parcial y siete filas. Escribir
11–10 en el primer juego y 2–2 en el segundo produce dos juegos incompletos.
Desplazar el diálogo hasta el final deja visibles el aviso «Solo el último set o
juego puede estar incompleto.» y Guardar bloqueado. Corregir el primer juego a
12–10 conserva el segundo parcial 2–2 y recupera Guardar. La acción queda completa
y accesible en el final del scroll; guardar desde Android muestra abandono del
visitante, victoria local y parcial 12–10 · 2–2. Reabrir Editar recupera esos
valores. Evidencia privada: android-table-tennis-original.png,
android-table-tennis-invalid-bottom.png,
android-table-tennis-valid-partial-bottom.png,
android-table-tennis-partial-saved.png y android-table-tennis-partial-reopened.png.

Se restauran por el editor el primer parcial 5–3 y el segundo vacío, se guarda y
la card vuelve a mostrar el abandono original. Una lectura independiente del API
confirma in_progress, bestOfSets=7, agregado administrativo 4–0, sets vacío e
incidencia retirement/away con partialSets=[5–3]. El torneo no se finaliza y se
conserva el historial de las escrituras reales. Evidencia:
android-table-tennis-original-recovered.png,
android-table-tennis-original-saved.png y android-table-tennis-final-api.json.

Tras volver a Inicio se repite una vez el mismo enlace con la app abierta. Esta
vez abre el torneo, con abandono y parcial 5–3 conservados. La captura inicial
inmediata aún mostraba Inicio; una segunda captura tras cargar confirma el
destino. Evidencia: android-table-tennis-warm-link-result.png. El resultado no
reproduce la caída en ese segundo intento ni demuestra que esté corregida.
El stack observado sitúa el rechazo en el guard de contexto de React de
DevLauncherAppLoader; no establece la causa ni un defecto del flujo de negocio.
No se cambia código de arranque para ocultar el fallo.

Retrospectiva: en un abandono, solo el último juego presente puede estar
incompleto. Probar dos parciales y corregir únicamente el primero distingue la
regla de secuencia de la validez de un juego completo. Siete filas requieren
comprobar el desplazamiento real hasta la ayuda y Guardar; esta pasada no juega
un partido completo de siete juegos. Una recuperación del development build
permite continuar, pero debe mantener el fallo fechado y abierto para diagnóstico.

Cierre: app en Inicio antes de apagar el emulador independiente y Metro.
API/PostgreSQL/Mailpit apagados; compose no muestra servicios activos y no quedan
procesos de emulador, Metro ni IDE. Se conservan volúmenes, fixtures y evidencia.
Esta continuación no cambia código del cliente ni reejecuta suites históricas;
Prettier de los documentos y git diff --check pasan. No se despliega ni publica.

## Continuación Android: login pendiente, reenvío y verificación — 2026-10-06

Se arranca Pixel_API_34 independiente, Metro y API/PostgreSQL/Mailpit, sin IDE
ni observabilidad. Una cuenta ficticia nueva se prepara por API; esto no acredita
registro ni aceptación de condiciones desde UI. Se cierra con confirmación la
sesión de qa_visual_player y se introduce la credencial de la cuenta pendiente.
Dos intentos nativos de login muestran «Hemos enviado un nuevo enlace de
verificación a tu correo.» y mantienen el formulario, sin abrir Cuenta autenticada.
Mailpit contiene tres correos: registro y ambos reenvíos. Evidencia privada:
android-pending-login-first.png, android-pending-login-screen.png y
android-pending-login-second.png. Credenciales y enlaces permanecen en el fixture
ignorado y archivos privados; no se publican tokens.

El enlace más reciente de Mailpit se abre con el esquema local en Android. Se
observa «Estamos verificando tu cuenta», después Inicio con recientes vacíos y
Cuenta con la identidad esperada. Reabrir el mismo enlace muestra «Este enlace
de verificación ya se ha utilizado. Solicita uno nuevo.» y Volver a la home.
La acción recupera Inicio; Cuenta sigue autenticada con la misma identidad.
Evidencia: android-verification-link-result.png,
android-verification-final-state.png, android-verification-account.png,
android-verification-used-link.png y android-verification-session-preserved.png.
El esquema local acredita recepción y procesamiento en la app instalada, no la
asociación HTTPS de App Links ni entrega externa del correo.

Force-stop y apertura de la URL exacta de Metro permiten comprobar un arranque
nuevo. La primera captura PNG todavía era el splash, aunque el XML conservaba
Cuenta: no se usa esa pareja como prueba de sesión. Tras completar el arranque,
abrir Cuenta confirma la misma identidad autenticada
(android-verification-restored-account.png). No se presenta como un refresh
forzado ni como diagnóstico resuelto de la caída de DevLauncher de la fase anterior.
La biblioteca muestra Administro 0 y Sigo 0 con mensajes distintos y Crear torneo
disponible. Evidencia: android-verification-admin-empty.png y
android-verification-follow-empty.png.

Retrospectiva: login pendiente y login verificado deben separarse por su efecto
de sesión, no solo por el código HTTP. El correo real de Mailpit, la transición
nativa, la identidad final, el rechazo del enlace consumido y la sesión tras
reinicio aportan evidencias distintas. Si uiautomator no obtiene un árbol nuevo
mientras hay animación, el archivo anterior puede seguir en el dispositivo:
se contrastan imagen, árbol y pantalla posterior antes de acreditar el resultado.

Notificaciones muestra «No tienes notificaciones.» y su cierre vuelve a Cuenta
(android-verification-notifications-empty.png). Se cierra con confirmación la
sesión nueva y se restaura qa_visual_player; Cuenta comprueba su identidad antes
de volver a Inicio (android-verification-qa-identity.png). La cuenta nueva queda
verificada y disponible en el fixture privado para futuras pruebas. No se borra
ni se modifica su contraseña. Esta pasada no modifica código del cliente.

Al restaurar qa_visual_player, el login navega a Torneos y Administro pasa de 26
a 27. Aparece otro QA retirada Android 20261006 sin empezar, con un solo equipo
Web Local. La inspección nativa y una lectura independiente del API confirman
published/football y ese equipo; se conserva sin iniciar ni cancelar. Evidencia:
android-verification-qa-account-restored.png y
android-verification-transferred-draft.json. La sesión técnica utilizada para
esa lectura se revoca al acabar.

Hecho: el cliente envía el borrador local completo opcional al login y navega a
Torneos cuando lo transfiere; ADR-0127 aceptado exige ese comportamiento y no
transfiere durante el login pendiente. Inferencia: la aparición y navegación son
compatibles con un borrador local anterior de QA transferido al restaurar la
cuenta. Esta pasada no capturó el borrador ni la petición antes del login, por
lo que no acredita su origen exacto ni una regresión de duplicados. No se cambia
la regla aceptada ni se oculta el efecto sobre los datos. La sesión de la cuenta
recién verificada había mostrado Administro 0 y Sigo 0.

Cierre: Inicio con la cuenta habitual de QA antes de apagar Pixel y Metro.
API/PostgreSQL/Mailpit apagados y sin procesos de emulador ni IDE. Se conservan
volúmenes, correos y evidencia; no se activa observabilidad ni se despliega.
Prettier de ambos documentos y git diff --check pasan. No se reejecutan suites
históricas para esta continuación documental. El QA global y la incidencia de
DevLauncher continúan abiertos.

## Continuación Android: generación nativa y diagnóstico de enlaces — 2026-10-06

Se conserva el APK instalado antes de intervenir. Su manifiesto solo declara
HTTPS /link/; la fuente app.config.ts declara también /join-team y /tournament/.
La build anterior estaba instalada desde el 3 de octubre. Este desfase acredita
una configuración nativa antigua, no la causa del guard de React de DevLauncher.
Evidencia privada: android-installed-before-rebuild.apk y
android-installed-manifest.txt.

Antes de instalar otra build, un intent VIEW/BROWSABLE con la app ya cargada abre
QA formulario table_tennis sin FATAL EXCEPTION ni el guard de contexto en los
logs de ese intento. Force-stop y el mismo enlace abren DevLauncherActivity, sin
esa excepción. La documentación oficial de Expo exige tener el proyecto abierto
para los enlaces propios en development builds; el launcher al arrancar desde
cero no se etiqueta como un fallo de la ruta ni acredita release. Evidencia:
android-link-old-warm.png, android-link-old-warm-result.json,
android-link-old-cold.png y android-link-old-cold-result.json. Referencia consultada
el 2026-10-06: [Expo: app-specific deep links](https://docs.expo.dev/develop/development-builds/development-workflows/).

Se ejecuta expo prebuild --platform android --no-install desde las dependencias
y plugins existentes, sin editar Kotlin/Gradle/Manifest a mano ni cambiar versiones.
El comando regenera el directorio ignorado Android y no cambia package.json;
el nuevo manifiesto incluye las tres entradas HTTPS. No se invoca --clean.
Se compila assembleDebug para arm64-v8a, arquitectura de Pixel_API_34, con el
Java existente de la instalación del IDE, sin arrancar el IDE. Logs privados:
android-prebuild-current.log y android-rebuild-current.log. La build anterior
queda conservada para comparar o reinstalar; los datos del emulador se preservan.

assembleDebug termina satisfactoriamente (497 tareas, 1m31s) y adb install -r
instala la build sin borrar datos. El manifiesto del APK recompilado confirma
/link/, /join-team y /tournament/; lastUpdateTime pasa a 2026-10-06 19:03:29.
Con la app detenida vuelve a abrir DevLauncherActivity sin la excepción.
Después de cargar Metro, dos intents del esquema local (VIEW con BROWSABLE y
VIEW original sin esa categoría) abren el torneo esperado, sin FATAL EXCEPTION
ni guard de contexto en sus logs. Un intent HTTPS limitado a com.fasttourney.app.local
abre también ese torneo. Esto acredita filtro y routing, no asociación del
dominio ni selección automática de la app desde un navegador. Evidencia:
android-rebuilt-manifest.txt, android-link-new-cold.png,
android-link-new-warm-1.png, android-link-new-warm-2.png,
android-link-new-https.png y sus archivos result.json/logs.txt.

La sesión habitual sigue en Cuenta después de reinstalar y navegar
(android-link-new-session.png). No se editan resultados ni se finaliza el torneo.
No cambia app.config, package.json, lockfile ni código versionado del cliente en
esta fase. Se actualiza DEVELOPMENT con la comprobación del binario instalado.
La caída del 6 de octubre permanece abierta: estos intentos no la reproducen,
pero no identifican una causa ni prueban una corrección. La entrada en frío de
una build distribuida sigue pendiente.

Retrospectiva: primero se acredita que el binario coincide con la configuración
que se pretende probar. El launcher esperado, el filtro HTTPS del APK y una caída
intermitente son evidencias distintas. Recompilar elimina el desfase observado;
no autoriza inferir que todo fallo de arranque compartía esa causa.

## Continuación Android: registro enviado desde UI — 2026-10-06

Con la development build recompilada, se cierra con confirmación la sesión de
QA y se abre Crear cuenta. Se preparan únicamente datos ficticios en el archivo
privado; esta cuenta no se registra por API. Sin aceptación de términos Crear
cuenta está deshabilitado. Al aceptar para la cuenta ficticia y enviar vacío,
los tres campos muestran errores obligatorios. Introducir qa_visual_player
muestra nombre no disponible y bloquea Crear cuenta. Un nombre nuevo recupera
su disponibilidad; correo-invalido y una contraseña corta muestran sus errores.
Intentar enviar esos valores conserva el formulario; Mailpit tiene cero correos
para la dirección nueva antes del alta válida. Evidencia privada:
android-registration-empty.png, android-registration-required.png,
android-registration-username-taken.png y android-registration-invalid-fields.png.

Corregir correo y contraseña elimina esos errores y muestra Seguridad: fuerte.
Con los tres campos válidos, desmarcar términos vuelve a deshabilitar Crear cuenta;
remarcarlos recupera la acción. Evidencia: android-registration-valid.png y
android-registration-valid-without-terms.png. Se confirma el envío desde el botón
nativo. Los valores de prueba permanecen en fixtures privados ignorados por Git.

El alta muestra «Revisa tu correo y abre el enlace de verificación para continuar.»
y vuelve al login vacío sin sesión. Mailpit contiene un único correo para la
cuenta preparada. Abrir su enlace con el esquema local verifica la cuenta y
termina en Inicio con recientes vacíos. Cuenta muestra el username esperado y
la biblioteca tiene Administro 0 y Sigo 0; no hereda colecciones de qa_visual_player.
Evidencia: android-registration-sent.png,
android-registration-pending-login.png, android-registration-verification-result.png,
android-registration-activated-account.png y android-registration-empty-library.png.
La preparación de datos no hizo POST /registrations: el único alta válido de
esta cuenta fue enviado desde UI. No se atribuye al enlace local una prueba de
App Links HTTPS ni de entrega externa.

Retrospectiva: preparar cuentas por API permite probar login, pero no acredita
validación, aceptación y envío del registro. Se separan nombres no disponibles,
campos inválidos y consentimiento de la cuenta ficticia; la corrección debe
recuperar la acción sin rehacer el formulario. La cuenta activada y sus colecciones
vacías comprueban que el cambio de sesión no conserva proyecciones del usuario previo.

Cierre: se cierra la nueva sesión con confirmación y se restaura qa_visual_player.
Cuenta confirma su identidad y la biblioteca mantiene Administro 27 y Sigo 5,
sin un nuevo aumento en este login. Evidencia:
android-registration-qa-session-restored.png y
android-registration-qa-library-restored.png. Se deja Inicio antes de apagar
Pixel, Metro y API/PostgreSQL/Mailpit. No quedan procesos del emulador, Metro ni
IDE; se detiene el daemon Gradle de la compilación. Se conservan APKs, volúmenes,
fixtures y evidencia. Sin observabilidad, despliegue ni publicación.

Prettier de DEVELOPMENT, este informe y LEARNING, y git diff --check pasan.
La build nativa debug compiló; no se presentan suites históricas como reejecutadas
ni esta build como release. El QA global, App Links verificados y el diagnóstico
de la caída intermitente de DevLauncher permanecen abiertos.

## Continuación web: eliminación previa y registro completo — 2026-10-06

Se prepara por API el fixture «QA bajas equipos web iOS 20261006», publicado
y sin empezar, con QA Local y QA Visitante. En web se cancela la confirmación
de eliminar el visitante y ambos equipos permanecen. Al confirmar, queda QA Local,
desaparece la acción de eliminar el último equipo y el inicio se deshabilita
con la indicación de necesitar dos participantes. Recargar conserva ese estado;
un GET independiente confirma published y un único equipo. No se intenta
convertir esta protección visual en evidencia nueva del rechazo HTTP 409,
ya cubierto por las pruebas anteriores.

Añadir equipo abre el formulario con Guardar vacío deshabilitado. Repetir
QA Local muestra el error de duplicado; corregir a QA Visitante recupera el
guardado. Recargar y un GET independiente confirman los dos nombres y el estado
published. El visitante repuesto tiene un ID nuevo: se restituye la composición
del fixture, no la identidad de la entidad eliminada. El torneo no se inicia.
Evidencia privada: web-team-delete-last-protected.jpg,
web-team-delete-restored.jpg, web-team-deleted-api.json y
web-team-restored-api.json.

Se prepara únicamente la cuenta ficticia qa_web_register_263be8; su alta no se
hace por API. Crear cuenta sin términos está deshabilitado. Aceptar y enviar
vacío muestra los tres errores obligatorios. qa_visual_player se declara ocupado
y bloquea el envío; el nombre nuevo está disponible. Correo inválido y contraseña
corta mantienen sus errores; intentar enviar no emite correo (Mailpit: cero).
Corregir ambos elimina los errores y muestra contraseña fuerte. Con datos válidos,
desmarcar términos vuelve a bloquear el envío.

Durante esta prueba se observa un defecto: la casilla funciona visualmente,
pero su nodo web carece de aria-checked. La operación check del controlador
activa el formulario pero no puede acreditar el estado. La implementación de
react-native-web 0.21.2 instalada consume aria-checked/accessibilityChecked,
sin convertir accessibilityState.checked. Se añade aria-checked={checked}
a TermsAcceptance, conservando accessibilityState para nativo. El DOM expone
true/false y check/uncheck funcionan; no se cambia consentimiento ni validación.
La solución mínima comparte el mismo booleano y no requiere otra biblioteca
ni casillas distintas por pantalla. Esta comprobación no sustituye VoiceOver
y TalkBack.

Tras marcar términos para la cuenta ficticia, el envío desde UI devuelve el
aviso de revisar correo y el login vacío, sin sesión. Mailpit contiene un único
correo. Abrir su enlace local real verifica la cuenta, termina en Inicio y
Cuenta muestra el username esperado. La biblioteca tiene Administro 0 y Sigo 0.
Se cierra esta sesión con confirmación y se restaura qa_visual_player: Administro
28 y Sigo 5, incluido el fixture nuevo preparado al inicio de esta fase.
No se atribuye a Mailpit entrega externa ni al navegador QA nativo de registro.
Evidencia: web-registration-required.jpg, web-registration-invalid.jpg,
web-registration-valid-without-terms.jpg, web-registration-ready.jpg,
web-registration-pending-login.jpg, web-registration-activated-account.jpg y
web-registration-empty-library.jpg. Credenciales y enlace permanecen privados.

Retrospectiva: una casilla que habilita correctamente el formulario puede seguir
sin comunicar su estado accesible. La prueba comprueba por separado datos,
consentimiento, efecto enviado y sesión activada. Recargar y consultar la API
acredita persistencia de equipos, mientras que volver a añadir un nombre
no recupera la identidad original.

## Continuación iOS: eliminación previa al inicio — 2026-10-06

El inventario nativo inicialmente no se obtiene porque el Mac está bloqueado.
Al recuperar acceso, Device Hub arranca iPhone 18 Pro iOS 27.0 pero queda en
Connecting display. Un Restart sin borrado recupera la pantalla. Fast Tourney
Local carga la sesión ficticia previamente activada qa_ios_pending_d0c222.
No se cambia el tema ni se activa observabilidad. Los toques por coordenadas
sobre Torneos y Return en Safari no producen navegación; el enlace del fixture
web escrito en Safari no se abre. Se cancela su edición y se devuelve Capture
Keyboard a off. Esta limitación de control no se atribuye al producto.

Los controles accesibles sí permiten completar otro recorrido: Crear torneo
desde Inicio, con «QA bajas equipos iOS 20261006» y QA Local, y añadir
QA Visitante. Es un fixture publicado nuevo de la cuenta ficticia iOS; no
es el fixture web ni se conserva un borrador local como sustituto de persistencia.
Se cancela la confirmación de eliminar al visitante y ambos permanecen. Después
se confirma: queda QA Local, desaparece Eliminar equipo y se deshabilita
Iniciar torneo con la indicación de necesitar al menos dos equipos. Un GET
independiente confirma published y QA Local como único equipo.

Añadir equipo presenta Guardar equipo vacío deshabilitado. QA Local muestra
«Ya existe un equipo con ese nombre.»; corregir a QA Visitante guarda, recupera
los dos controles Eliminar y habilita Iniciar torneo. Otro GET confirma ambos
nombres persistidos y published. El torneo no se inicia. La reposición crea
un visitante nuevo; no recupera el ID eliminado. No se acredita reapertura
de esta ruta ni registro completo iOS en esta fase. Evidencia privada:
ios-team-delete-last-protected.jpg, ios-team-delete-restored.jpg,
ios-team-deleted-api.json e ios-team-restored-api.json.

Retrospectiva: una limitación de gesto no impide todos los recorridos nativos.
Crear un fixture independiente desde Inicio permite probar los controles
accesibles sin inventar navegación ni permisos. Una consulta pública posterior
confirma persistencia, pero no sustituye una reapertura visual.

Validación del cambio compartido: typecheck, ESLint del componente, Prettier
y exportación web de 36 rutas pasan. La checklist de cliente conserva textos
localizados, tokens, objetivo de 44 px, consentimiento existente y adaptación
HTTP sin cambios; no se toca ninguna operación OpenAPI. El atributo comparte
el booleano de accesibilidad nativo y no introduce estado alternativo.
Se documenta el contrato semántico en DESIGN_SYSTEM y el aprendizaje reutilizable.
El QA global, lectores de pantalla, registro iOS completo y diagnóstico de la
caída intermitente Android continúan abiertos.

Cierre verificado: no hay simuladores booted, Metro no escucha en 8082 y
compose local no tiene servicios running. Se conservan volúmenes, fixtures y
evidencias, con qa_visual_player restaurado en web y la cuenta ficticia anterior
en iOS. Sin observabilidad, despliegue ni publicación. Se normaliza únicamente
el salto final de expo-env.d.ts generado por Metro. git diff --check pasa.

## Continuación iOS: registro enviado y activación — 2026-10-06

Se arranca únicamente la sesión local de QA y Metro para la development build.
Se prepara una página temporal de enlaces en loopback 127.0.0.1:8098 para abrir
Cuenta desde Safari sin depender del envío de Return a una dirección. El enlace
custom scheme en frío llega primero al launcher; seleccionar el servidor Metro
carga el proyecto y abre Cuenta con la sesión ficticia qa_ios_pending_d0c222.
No acredita asociación HTTPS ni comportamiento de una build distribuida.

Se cierra la sesión con confirmación y se abre Crear cuenta. Para
qa_ios_register_4eadc0 se preparan solo datos privados: no se registra por API.
La casilla anuncia unchecked y mantiene Crear cuenta deshabilitado; marcarla
anuncia checked. Enviar vacío muestra errores de username, correo y contraseña.
qa_visual_player aparece no disponible y bloquea el envío. El username nuevo
está disponible; correo-invalido y una contraseña corta mantienen sus errores.
Intentar enviar esos formatos no genera correo en Mailpit (cero). Corregir los
dos campos elimina errores y muestra Seguridad: fuerte. Desmarcar términos
con todos los campos válidos vuelve a bloquear Crear cuenta; remarcarlos
recupera el envío. Evidencia privada: ios-registration-required.jpg,
ios-registration-invalid-fields.jpg e ios-registration-valid-without-terms.jpg.

El envío válido desde UI muestra «Revisa tu correo y abre el enlace de
verificación para continuar.» y devuelve el login vacío, sin sesión. Mailpit
contiene un único correo para esta cuenta. Su enlace real se prepara en la
página temporal privada y se abre desde Safari en la app ya cargada. La
verificación termina en Inicio con recientes vacíos. Cuenta muestra el username
esperado y Torneos tiene Administro 0 y Sigo 0. Evidencia:
ios-registration-pending-login.jpg, ios-registration-activated-account.jpg e
ios-registration-empty-library.jpg. La activación no hereda torneos ni sesión
de la cuenta anterior. No se atribuye a Mailpit entrega externa.

Se cierra la nueva sesión con confirmación y se restaura qa_ios_pending_d0c222.
Cuenta confirma la identidad y Torneos muestra Administro 3 y Sigo 0, con los
fixtures de byes, retirada y bajas de equipos. Reabrir el último muestra
QA Local y QA Visitante y permite iniciar, sin hacerlo: cierra el pendiente de
reapertura visual de la reposición iOS de la fase anterior. Evidencia:
ios-registration-previous-account-restored.jpg,
ios-registration-previous-library-restored.jpg e ios-team-restored-reopened.jpg.

Los toques de las tabs funcionan en este arranque. Al restaurar sesiones, iOS
muestra un aviso de Guardar contraseña que no figura en el árbol accesible de
Device Hub: se descarta con Ahora no antes de seguir. Se registra como
observación de esta fase; no se atribuye retrospectivamente a ese aviso el
origen de todos los controles sin respuesta de fases anteriores.

Se cambia temporalmente a qa_visual_player para revisar el desplazamiento de
sus 31 notificaciones. La lista carga; drag y scroll sobre el viewport no
cambian las fechas visibles (primera 5/10/2026 23:58:24). No se acredita llegar
por gesto hasta el final, ni se convierte esa limitación del controlador en
defecto del producto. No se borran avisos ni se activan fuera del viewport.
Evidencia: ios-notifications-scroll-unverified.jpg. Se cierra esta sesión y
se restaura la cuenta iOS anterior, descartando el guardado de credenciales,
y se deja Inicio con sus tres recientes.

Retrospectiva: los datos preparados por API no sustituyen un alta enviada desde
UI. Separar validación, disponibilidad, consentimiento, correo y sesión activada
permite cerrar el recorrido nativo sin inferirlo del backend. Los avisos del SO
pueden no aparecer en la proyección accesible; una captura explica un bloqueo
que el árbol por sí solo no muestra. Un enlace servido solo en loopback permite
abrir rutas reales mediante Safari sin cambiar código ni recurrir a navegación
programática del cliente. Clicks exitosos no acreditan gestos de scroll.

En esta fase no cambia código del cliente, configuración nativa ni contrato.
No se presentan typecheck ni exportaciones históricas como reejecutadas. El
registro completo queda comprobado en las tres plataformas; QA global, scroll
iOS, lectores de pantalla y caída intermitente Android continúan abiertos.

Cierre verificado: simuladores booted vacíos, sin listeners en 8082/8098 y
compose local sin servicios running. Se conservan volúmenes, fixtures y
capturas; se retira el enlace con token de la página temporal. Sin observabilidad,
despliegue ni publicación. Prettier de este informe y LEARNING y git diff --check
pasan. Se normaliza el salto final de expo-env.d.ts generado por Metro.

## Continuación Android — dos vueltas, grupos y refresco explícito (2026-10-06)

Sesión autorizada en Pixel_API_34 independiente, sin IDE ni dispositivo físico.
Se usan la development build instalada, Metro y API/PostgreSQL/Mailpit locales;
no se activa observabilidad. La cuenta qa_visual_player conserva su sesión.
Dos fixtures nuevos publicados, con equipos ficticios, se preparan por API; la
configuración, inicio, transiciones y finalización se ejecutan desde Android.

### Liga de ida y vuelta

QA Android liga ida vuelta 20261006 tiene dos equipos. Se selecciona Liga e
Ida y vuelta y se inicia desde UI. Jornada 1 enfrenta Equipo 1 contra Equipo 2;
Jornada 2 invierte local y visitante. Ambos marcadores 2–2 se envían desde UI.
Clasificación muestra dos primeros puestos: PJ 2, PG 0, PE 2, PP 0, GF 4,
GC 4, DG 0 y Pts 2. La tabla mantiene su indicación de desplazamiento horizontal.

Finalizar solicita confirmación de cierre irreversible de resultados. Confirmar
muestra ambos equipos como «Campeones compartidos». Al cerrar el diálogo, el
estado es Finalizado y desaparece la edición. El GET independiente confirma
format league, roundRobinLegs 2, dos resultados completos 2–2 y dos championTeamIds.
Evidencia privada: android-league-tied-standings.png,
android-league-cochampions.png y android-formats-liga-ida-vuelta-snapshot.json.

### Mixto por grupos

QA Android mixto grupos 20261006 tiene ocho equipos. Liga + Eliminatorias y
Grupos, con cuatro grupos y dos clasificados por grupo, muestra que necesita
12 equipos y deshabilita Iniciar. Cambiar a dos grupos conserva dos clasificados,
elimina la advertencia y habilita el inicio. Se generan doce partidos.

Once resultados se preparan por API para reducir repetición; el duodécimo,
Equipo 6 contra Equipo 7, se introduce y guarda 2–0 desde UI. No se atribuyen
los once primeros envíos a interacción nativa. «Cerrar liga y continuar» abre
la confirmación de clasificación congelada. Cancelar conserva la edición;
confirmar genera dos semifinales sin partidos de desempate, al estar resuelto
el corte. La vista Liga conserva marcadores y retira controles de edición.

Clasificación permite alternar ambos grupos. Grupo 1 ordena equipos 1, 4, 5, 8;
Grupo 2, equipos 2, 3, 6, 7. Ambos muestran puntos 9, 6, 3, 0 y PJ 3.
Las semifinales 1–4 y 2–3 se guardan 2–0 desde UI. El enlace del cuadro conduce
a la final 1–2; guardar 2–0 habilita Finalizar. Confirmar muestra Equipo 1 como
campeón. Cerrar conserva el resultado y elimina Editar. El GET confirma torneo
y etapas completed, doce resultados de grupos y tres de cuadro completos, y el
championTeamId correspondiente. Evidencia: android-mixed-group-one-frozen.png,
android-mixed-semifinal-one-ready.png, android-mixed-semifinal-two-ready.png,
android-mixed-final-ready.png, android-mixed-champion.png y
android-formats-mixto-grupos-snapshot.json.

### Hueco detectado y corrección de refresco

Hecho: reabrir una ficha ya cargada conserva la entidad canónica en memoria y
no lee cambios externos. La ficha no ofrecía el refresh explícito comprometido
en ADR-0085, aceptado. Se concreta esa decisión mediante Actualizar en el menú
existente de acciones, con Button secundario y ModalDialog compartidos. Reutiliza
refreshTournament y la resolución de relación; no introduce dependencia de
caché, polling ni revalidación automática. El coste de mantenimiento es menor
que incorporar una librería global para este único mecanismo ya aceptado.

La primera instancia del emulador seguía ejecutando código anterior. Tras
recargar la build con Metro se abre la ficha, que muestra 2–0. Un cambio externo
real a 3–0 no se refleja hasta pulsar Actualizar; entonces aparece 3–0. Detener
solo la API local y repetir muestra common_network_error y conserva el marcador.
Recuperar API, restaurar el resultado externo a 2–0 y actualizar confirma la
recuperación. El marcador original queda restaurado antes de cerrar los grupos.
android-refresh-network-error.png acredita visualmente banner y contenido.

Un proxy temporal, limitado a loopback y a este GET del fixture, inyecta un 500
con problema desconocido. El prefijo real /v1 se respeta antes de considerar
válida la prueba; el proxy confirma la inyección y Android muestra únicamente
common_request_error («Estamos teniendo problemas…»), sin title ni detail de
prueba. El contenido finalizado se conserva. Los intentos anteriores sin ese
prefijo no cuentan como prueba de error. La captura tardía del banner ya no lo
muestra; el árbol del envío correcto y el registro del proxy acreditan el caso.

La prueba de cuerpo 200 inválido se preparó pero no se ejecutó: la revisión
automática de aprobación falló temporalmente por límite de uso, sin considerar
la operación insegura. HTTP 429 tampoco se ejecutó. Esos casos, 404 y cancelación
visual de este nuevo refresco quedan pendientes; no se infieren de los casos
web históricos. La aprobación posterior permitió apagar el compose local.

Revisión de cliente: GET conserva 200 y 404 declarados en OpenAPI. El adaptador
getTournament usa getPublicTournament y apiFetch, valida la proyección y mapea
solo 404 a TournamentUnavailableError; el resto mantiene fallback seguro. La
relación usa el adaptador existente paginado. No cambian endpoints, contratos ni
instrumentación del servidor. El menú evita duplicados con ref y loading; al
cambiar ID, cuenta o desmontar invalida commits tardíos de relación/feedback.
AbortError e invalidación de sesión no muestran un banner nuevo. La etiqueta
common_refresh existe en es/en/fr/it; no hay textos, colores, medidas ni popups
locales nuevos. Se conserva el patrón de cierre y objetivo táctil compartido.

Validación reejecutada: pnpm run check pasa formato, lint, typecheck y OpenAPI;
expo export --platform web pasa con 36 rutas. El primer check falló únicamente
por el salto final de expo-env.d.ts generado por Metro, normalizado antes del
check que pasó. La exportación no acredita revisión visual web/iOS del control.
El resto de deportes y formatos nativos, desempate de corte nativo, lectores de
pantalla, scroll iOS, refresh de sesión, OAuth real diferido y caída intermitente
Android siguen con sus pendientes previos. No se observó otra caída en esta fase.

Retrospectiva: compartir caché de mutaciones y leer cambios externos son dos
recorridos que requieren evidencia independiente. Una recarga de la build no
prueba refresh de producto: se debe cargar una ficha, cambiar después el dato
fuera de la app y usar su control explícito. Un control flotante de Expo tapaba
las acciones; moverlo permitió probar el menú sin cambiar el layout del producto.
El fixture amplio preparado por API se distingue de los partidos guardados por
UI y permite invertir tiempo en transición, congelación y cierre.

Cierre: compose local sin servicios running, sesiones de Metro/proxy cerradas,
emulador terminado y sin listeners 8082/8085. Se conservan volúmenes, fixtures y
evidencia. Sin despliegue, publicación ni cambios en producción. Se normaliza
expo-env.d.ts tras detener Metro. Cambios permanecen sin commit en develop.

## Continuación Android — cierre de errores del refresco (2026-10-06)

Se retoman los casos que no se ejecutaron por el fallo temporal de aprobación.
La nueva sesión sí permite arrancar y cerrar los servicios de QA autorizados.
No se cambia código del producto ni dependencias; se usan el mismo fixture mixto
finalizado, emulador Pixel_API_34 independiente y proxy limitado a loopback.
Las inyecciones afectan solo al GET /v1/tournaments del fixture; no son fallos
observados del backend ni acreditan políticas reales de rate limiting.

Al restaurar el emulador, la app conservaba un menú abierto. Recargar y seleccionar
Metro muestra una excepción de inicialización: RNSModule, NullPointerException,
conversión a FabricUIManager y ScreensModule.setupFabric. La evidencia privada
android-fabric-reload-failure.png/.xml conserva la traza. Reiniciar solo el proceso
con force-stop, sin borrar datos, volver al launcher y seleccionar el mismo Metro
recupera Inicio. Cuenta confirma qa_visual_player al terminar. Este síntoma no
establece la causa del guard de DevLauncher registrado antes y no se presenta
como resuelto; queda pendiente diagnóstico de arranque/recarga nativa. El emulador
avisó de presión de memoria, sin demostrar relación causal con la excepción.

Con la ficha cargada desde la API real se ejecuta Actualizar en estos casos:

- **200 con cuerpo inválido:** el proxy entrega un objeto que no cumple
  PublicTournament. Se muestra common_request_error y permanece la ficha
  finalizada con sus marcadores. Evidencia: android-refresh-invalid-body.png/.xml.
- **429 no tratado:** el problema sintético incluye title y detail de prueba.
  Se muestra el mismo mensaje común, sin exponer ninguno de esos campos.
  El proxy confirma 429 y la captura conserva banner y contenido:
  android-refresh-http-429.png/.xml. No se añade un mensaje de negocio por estado.
- **404:** se sustituye el contenido cacheado por «Este torneo ya no está
  disponible.» y Cerrar. Desaparecen equipos y marcadores de la ficha. Cerrar
  vuelve a Inicio; retirar la inyección permite reabrir el torneo.
  Evidencia: android-refresh-unavailable.png/.xml.
- **Espera y salida:** con un 500 demorado, reabrir el menú muestra el control
  ocupado y deshabilitado. Evidencia: android-refresh-in-flight.png/.xml. Una
  primera demora de doce segundos no permitió acreditar la salida: al llegar el
  banner, el toque no cerró la ficha. Se repite con treinta segundos y se confirma
  Inicio mientras el indicador privado del proxy aún marca la respuesta pendiente.
  Tras concluir el 500, Inicio conserva sus recientes y no muestra el error de la
  ruta cerrada. Evidencia: android-refresh-left-while-pending.png/.xml y
  android-refresh-left-after-error.png/.xml. Es descarte de feedback tardío al
  desmontar, no prueba de aborto del transporte ni de timeout propio del cliente.

El proxy usado es secuencial: durante la demora también retiene las lecturas
posteriores de Inicio, que muestra Cargando hasta que termina. Esa espera no se
atribuye a un defecto del producto. Un banner transitorio tapa controles de la
cabecera: se verifica el menú resultante antes de activar el siguiente botón;
un toque durante el banner no acredita que se abrió el menú ni una petición.

Se validan las evidencias XML: mensajes comunes presentes, title/detail sintéticos
ausentes y 404 sin equipos obsoletos. Un GET directo, fuera del proxy, devuelve
una proyección idéntica al snapshot final de la fase anterior: ningún marcador,
etapa ni campeón se altera en esta pasada. Se restaura el mapeo directo de ADB a
la API, modo pass del proxy e Inicio con la cuenta original antes de apagar.

Retrospectiva: cada fallo sintético debe corroborarse en el proxy y en la UI; la
navegación durante espera requiere demostrar el orden entre salida y respuesta.
Capturar inmediatamente el banner preserva evidencia que una captura tardía
pierde. Recuperar un proceso nativo permite continuar QA, pero no equivale a
corregir la excepción ni autoriza atribuirle una causa sin evidencia.

Los casos 200 inválido, 429 y 404 de este refresco Android quedan cerrados, junto
con carga deshabilitada y salida durante espera. Siguen pendientes la revisión
visual del control en web/iOS, cambios de cuenta durante refresh, aborto de
transporte, refresh de sesión y los demás casos de la matriz global. No se
reejecutan typecheck/export de la fase anterior porque aquí no cambia código;
se verifica formato documental y git diff --check. La caída de inicialización
Fabric se añade a las incidencias nativas abiertas.

Cierre verificado: compose local sin servicios running, Metro/proxy cerrados,
emulador terminado y sin listeners 8082/8085. Se conservan volúmenes y evidencia,
sin observabilidad, despliegue ni publicación. Se normaliza el salto final del
archivo expo-env.d.ts generado por Metro. El trabajo permanece sin commit en develop.

## Continuación web — refresco explícito y regresión automatizada (2026-10-07)

Se continúa el QA autorizado con la exportación actual del cliente, API,
PostgreSQL y Mailpit locales. El navegador conserva la sesión ficticia existente;
no se cambian credenciales, permisos, equipos ni resultados. Un proxy temporal
concurrente, limitado a loopback, inyecta fallos únicamente en el GET /v1/tournaments
del fixture mixto Android finalizado. No se activan servicios de observabilidad.

El menú web presenta Actualizar con Button secundario y ModalDialog compartidos,
sin recortes observados a 1280 × 720. Una petición correcta cierra el menú y
conserva el torneo finalizado con sus marcadores 2–0. La evidencia visual queda en
web-refresh-menu-20261007.png, bajo el directorio privado de esta revisión.

Se recorren desde ese control las salidas siguientes:

- **200 inválido, 429 no tratado y 500 desconocido:** muestran exclusivamente
  common_request_error y conservan el cuadro y los resultados. Los árboles
  guardados verifican ausencia de los title/detail sintéticos; el proxy registra
  cada inyección. Evidencia web-refresh-invalid/429/500-20261007.txt y .png.
- **404 declarado:** retira equipos y resultados obsoletos, muestra «Este torneo
  ya no está disponible.» y permite cerrar a Inicio. Retirar la inyección y
  reabrir recupera la ficha. Evidencia web-refresh-404-20261007.txt y .png.
- **Espera:** durante un 500 demorado, reabrir el menú muestra Actualizar ocupado
  y deshabilitado; isEnabled devuelve false. Evidencia web-refresh-busy-20261007.
- **Salida durante espera:** el primer intento de cerrar el menú y pulsar Volver
  inmediatamente no navegó y no cuenta como salida. Se repite sin reabrir el
  menú: Inicio queda acreditado a las 1791350382.642 s Unix, antes de finalizar
  el proxy a las 1791350390.099 s. Tras la respuesta, Inicio conserva recientes
  y no muestra feedback de la ficha cerrada. Evidencia
  web-refresh-left-confirmed/left-after-error-20261007.txt y registro
  web-refresh-requests.jsonl. Se valida descarte de feedback tardío, no aborto
  de transporte ni timeout del cliente.
- **Corte real y recuperación:** reabrir con API detenida falla en la carga de
  relación y presenta el error común de conexión con Reintentar; recuperarla
  permite cargar. En una prueba separada, con la ficha ya cargada, detener solo
  API y usar Actualizar muestra common_network_error conservando el contenido.
  Recuperar API y repetir Actualizar devuelve el cuadro sin banner nuevo.
  Evidencia web-reopen-network y web-refresh-network/final-recovered-20261007.

Las capturas y árboles se verifican contra los mensajes esperados y la presencia
o ausencia de equipos según cada caso. Las respuestas fallidas son sintéticas,
excepto el corte de API: no acreditan defectos del servidor ni rate limiting real.
No se modifica código del producto ni se vuelven a atribuir los datos del fixture
a nuevas interacciones deportivas. Esta fase cierra la revisión visual web del
control y sus errores; sigue pendiente iOS, cambio de cuenta durante refresco,
aborto de transporte y una prueba web de cambio externo real. El resto de la
matriz global conserva su alcance anterior, incluidos refresh de sesión, lectores
de pantalla, scroll iOS y diagnóstico de inicialización Android.

Regresión: pnpm run check pasa formato, lint, typecheck y OpenAPI; exportación web
completa 36 rutas. Node aprueba 72 pruebas y seguridad operacional 11. Go aprueba
483 tests/subtests y omite 57, sin fallos; no se ejecuta la integración PostgreSQL
opt-in. El primer intento Go bajo sandbox no pudo abrir servidores httptest ni
completar el setup; repetir con los permisos necesarios pasa, sin cambios de código.
Los contadores no equivalen a cobertura de líneas ni casuísticas independientes.

Retrospectiva: probar una carga sin API y refrescar contenido ya cargado son
recorridos distintos. Una acción de cierre inmediatamente después de ocultar un
menú necesita comprobar el destino real; el clic por sí solo no prueba navegación.
El proxy concurrente permite verificar feedback tardío sin bloquear las lecturas
de Inicio. Se conserva la solución existente, sin nuevas dependencias ni ADR.

Cierre verificado: local y dev sin contenedores activos; sin listeners en
8080/8082/8084. Se conservan volúmenes; proxy y servidor web cerrados, pestaña
temporal cerrada desde Inicio. Sin despliegue, publicación ni
cambios en producción. El trabajo anterior se conserva sin commit en develop.

### Reapertura y cambio externo web real (2026-10-07)

El usuario indica que se mantenga el entorno activo mientras queden comprobaciones
por hacer. Se reabren API/PostgreSQL/Mailpit, vista web y proxy, sin observabilidad;
el cierre previo fue un corte de esta pasada, no el cierre global del QA.

Se prepara por API un fixture ficticio exclusivo, QA Web refresh 20261007, con
dos equipos, liga de una vuelta y resultado 2–0. La ficha web se carga antes de
cambiar por API a 3–0: sigue mostrando 2–0 hasta pulsar Actualizar, tras lo cual
muestra 3–0. Se restaura por API a 2–0 y Actualizar confirma la restauración.
Evidencia privada: web-refresh-real-2/3.json,
web-refresh-external-before/after-20261007.txt y captura after. Se cierra así el
pendiente web de actualización externa real. Preparación, inicio y mutación del
fixture fueron por API; no se atribuyen a interacción de creación deportiva web.
Se continúa con iOS, manteniendo el entorno de pruebas abierto.


### Continuación iOS: refresco y banner global (2026-10-07)

La development build instalada carga Metro en loopback usando localhost tanto
para el manifest como para el bundle. No se habilita exposición LAN ni
observabilidad. El fixture exclusivo anterior permite acreditar Actualizar con
un cambio externo por API de 2–0 a 3–0 y su restauración a 2–0. Evidencia privada:
ios-refresh-external-before/after-20261007.txt y captura after.

Un cuerpo inválido durante el refresco conservaba el marcador pero el banner
raíz basado en Modal no aparecía sobre la ficha modal iOS. El primer enfoque
trasladó el host a Screen; el usuario señaló que eso alteraba una decisión
aceptada: en apps el banner es global y sobrevive a la navegación. Ese enfoque
se retiró íntegramente. La corrección final mantiene el provider y host raíz y
usa FullWindowOverlay, ya disponible en react-native-screens, para iOS.
Android y web conservan sus hosts existentes; Screen no tiene cambios por esta
corrección. No se modifican duración, mensajes ni reglas de recuperación.

Con el host final se comprueban 429, 500 y cuerpo inválido: el mensaje común
seguro aparece sobre la ficha y conserva 2–0. Tras el 429 se pulsa Equipos y el
banner permanece sobre la nueva pantalla hasta su cierre automático. La captura
muestra los equipos bajo el aviso y, tras expirar, el árbol confirma la pantalla
Equipos. Evidencia privada: ios-global-banner-before-navigation,
ios-global-banner-after-navigation, ios-global-banner-500 e
ios-global-banner-invalid-20261007, con capturas y árboles. La captura de cuerpo
inválido con el enfoque intermedio queda superada por esta evidencia final.
Los fallos HTTP y de decodificación son inyectados; no prueban fallos reales del
servidor. No se acredita aquí VoiceOver ni el resto de casuísticas iOS pendientes.

Retrospectiva: una corrección debe cumplir las decisiones vigentes antes de
reutilizar un patrón parecido de otro componente. Mover un host puede cambiar
su alcance y ciclo de vida aunque haga visible el aviso. El QA no autoriza esa
alteración; si resolver un defecto exige cambiar una decisión aceptada, se
expone el conflicto y se consulta al usuario antes de implementarlo. El entorno
sigue activo porque la sesión de QA continúa; se restaura el proxy a pass.


Validación de la corrección final: pnpm run check pasa formato, lint, typecheck y
OpenAPI; la exportación web completa 36 rutas y Node aprueba 72 pruebas, sin
fallos ni omisiones. git diff --check pasa. No se añaden dependencias, módulos
nativos ni cambios de contrato. Las suites Go y operacional conservan la
validación de esta pasada anterior al cambio exclusivamente visual.


### Continuación autónoma: esperas, indisponibilidad y renovación (2026-10-07)

**iOS, refresco de ficha:** durante un 500 demorado 15 segundos, el menú muestra
Actualizar ocupado y deshabilitado. El primer intento de cerrar el menú y la
ficha inmediatamente no salió y no cuenta como salida. En una segunda petición,
sin reabrir el menú, Inicio queda observado a las 1791352244.124 s Unix y el proxy
termina a las 1791352251.637 s. Tras finalizar, Inicio no muestra feedback tardío.
Esto acredita descarte del aviso aún no emitido; no modifica ni contradice la
permanencia al navegar de un banner ya visible. No acredita aborto de transporte.
Evidencia: ios-refresh-busy, ios-refresh-left-confirmed e
ios-refresh-left-after-error-20261007, más web-refresh-requests.jsonl.

El 404 inyectado retira equipos y marcador, muestra «Este torneo ya no está
disponible.» y permite cerrar a biblioteca. Al retirar la inyección, reabrir
recupera 2–0. Evidencia: ios-refresh-404-20261007.txt y .png. Al inicio de esta
pasada, el proxy había dejado de escuchar entre turnos: Actualizar mostró el
error común de conexión conservando la ficha. Se distingue ese corte real del
proxy de los fallos HTTP sintéticos; API y PostgreSQL seguían activos.

**Listas iOS:** se repite con qa_visual_player y sus 31 avisos. La primera fecha
visible es 5/10/2026 23:58:24 y no cambia tras drag ni scroll por Device Hub. No
se acredita llegar al final ni se declara un defecto de producto por esta
limitación. No se borran avisos. Evidencia: ios-notifications-scroll-start e
ios-notifications-scroll-unverified-20261007.png. Se restaura la cuenta iOS
anterior, qa_ios_pending_d0c222, sin guardar contraseñas en el sistema.

**Renovación iOS (ADR-0062):** el proxy modifica únicamente expiresAt de un login
ficticio para situarlo a un minuto, conservando los tokens reales emitidos por
API y su duración real. La app solicita automáticamente POST /v1/sessions/refresh,
que responde 200 real, y carga sus tres torneos con GET 200. La identidad queda
conservada. No se espera siete días ni se cambia la política de sesiones.

En otra sesión con el mismo vencimiento próximo simulado, el proxy corta la
conexión antes de remitir refresh a API. La app conserva identidad y muestra
common_network_error con Reintentar en biblioteca. Tras retirar el corte,
Reintentar obtiene refresh 200 real y recupera Administro 3/Sigo 0. Evidencia:
ios-session-refresh-library/account, ios-session-refresh-network,
ios-session-refresh-network-identity e ios-session-refresh-recovered-20261007.
Los dos intentos cortados pertenecen a consultas consecutivas, no prueban ni
refutan coordinación concurrente. Sigue pendiente el recorrido visual Android,
refresh rechazado/reutilizado y concurrencia visual explícita.

**Renovación web:** una lectura protegida de biblioteca recibe un único 401
sintético; apiFetch renueva cookies con POST /v1/sessions/refresh 200 real y repite
la lectura con GET 200. La biblioteca muestra Administro 31/Sigo 5 y Cuenta
conserva qa_visual_player. El primer intento encontró una carrera en el marcador
de un solo uso del proxy concurrente, causando un corte adicional y el error de
conexión. Se excluye ese intento como prueba limpia de renovación y se repite
tras hacer atómica la inyección. La segunda pasada recupera sin otro login.
Evidencia: web-session-refresh-library/account-20261007 y
session-refresh-requests.jsonl. El registro conserva solo método, ruta, estado,
tiempo y marca de inyección; no credenciales ni cuerpos de sesión. No acredita
caducidad real de cookies, revocación de backend ni renovación fallida web.

**Continuidad del entorno:** los procesos efímeros de herramientas no conservaron
Metro ni proxy al acabar el turno anterior. Se trasladan a PTY mantenidas por un
helper temporal, qa-pty-host.py, fuera del repositorio; PID y logs quedan en el
directorio privado de evidencia. Se verifica que sobreviven a ejecuciones
independientes. Proxy/vista web escuchan en 127.0.0.1:8080/8082 y Metro en ::1:8083.
API/PostgreSQL/Mailpit siguen activos, sin observabilidad ni tareas programadas.
Todas las inyecciones vuelven a pass. Esta continuidad responde a la instrucción
del usuario durante QA y no habilita desarrollo permanente fuera de esa sesión.

Retrospectiva: una prueba visual de sesión necesita correlacionar identidad,
contenido y renovación HTTP; un 200 aislado no la acredita. Simular el dato de
vencimiento y cortar antes de API permite comprobar recuperación sin tocar
secretos ni configuración de producto. Las carreras del inyector se corrigen y
se separan de defectos de producto. Esta fase no modifica código del cliente ni
decisiones aceptadas, y conserva los pendientes globales del inventario.


### Continuación iOS: validación de cinco editores deportivos (2026-10-07)

Se cambia de nuevo temporalmente a qa_visual_player para usar sus fixtures.
Los torneos de biblioteca se activan mediante controles accesibles; eso puede
llevar la lista a la fila elegida, pero no acredita scroll por gesto. Los campos
se editan con accesibilidad: no se acredita teclado virtual ni VoiceOver.
No se envía ningún resultado en esta fase.

- **Pádel, mejor de tres:** el original 6–4/7–6 habilita Guardar. Añadir un tercer
  6–0 después de la victoria o cambiar el primero a 6–5 lo bloquea; restaurar
  los parciales válidos lo habilita. En abandono, seleccionar al local muestra
  victoria del visitante. 6–4 seguido de 2–3 habilita Guardar; 3–3 seguido de
  2–3 lo bloquea por dos sets incompletos. El scroll enviado por Device Hub no
  lleva visiblemente al botón inferior del popup; su estado se acredita en el
  árbol, no se presenta como botón alcanzado por gesto. El cierre accesible no
  tuvo efecto observado; tocar el fondo exterior sí descarta el borrador. La
  ficha conserva 2–0, con 6–4/7–6. Evidencia ios-padel-extra-set, invalid-set,
  abandon-valid, abandon-two-incomplete, abandon-scroll y discarded-20261007.
- **Bádminton, mejor de tres a 15:** 15–14 bloquea por falta de diferencia de dos,
  16–14 habilita, 22–20 bloquea por superar el tope y 21–20 habilita la excepción
  del máximo. Cerrar y reabrir conserva 21–20/15–12. Evidencia ios-badminton-
  cap-valid, win-by-two, deuce-valid, over-cap, cap-restored y discarded-20261007.
- **Balonmano, eliminación:** con 25–25, cambiar la tanda de 5–4 a 5–5 bloquea
  Guardar. Volver a 5–4 lo habilita. Cambiar a 25–24 retira los campos de tanda.
  Cerrar descarta el cambio y conserva 25 (5)–25 (4). Evidencia ios-handball-
  shootout-tied, shootout-restored, no-shootout y discarded-20261007.
- **Tenis, mejor de cinco:** el fixture persistido contiene abandono del visitante
  con parcial 2–3. Cambiar solo el borrador a Marcador y completar 7–6, 7–5 y 6–4
  habilita Guardar; añadir 6–0 en el cuarto set lo bloquea. Cerrar devuelve el
  abandono original con 2–3. Evidencia ios-tennis-baseline, three-sets-valid,
  extra-set y discarded-20261007. No se sustituye la incidencia persistida.
- **Voleibol, mejor de cinco:** el fixture anterior estaba finalizado. Se prepara
  por API uno exclusivo, QA iOS voleibol 20261007, en curso con 3–2 y parciales
  25–20/20–25/26–24/20–25/16–14. Se recarga el runtime para cargarlo; esa recarga
  no se cuenta como actualización de producto. En el editor, el quinto 15–14
  bloquea y 15–13 habilita; un primer set 24–20 bloquea. Restaurar el marcador
  original habilita Guardar y cerrar conserva el 3–2 y sus cinco parciales.
  Evidencia ios-volleyball-baseline, deciding-invalid, deciding-valid,
  regular-invalid, restored y discarded-20261007. La preparación del fixture
  no acredita creación, inicio ni escritura de resultados desde UI iOS.

El Mac se bloqueó antes del ensayo 24–20. Se detuvo la automatización visual y
el usuario lo desbloqueó; se volvió a observar el editor antes de continuar.
No se acredita ninguna acción durante el bloqueo. Las referencias accesibles
pueden caducar tras renderizar un campo: antes de repetir se observa el valor
real, porque la escritura puede haber ocurrido aunque falle la captura siguiente.

Retrospectiva: habilitar Guardar, alcanzar el botón y persistir son tres pruebas
diferentes. Esta pasada amplía validación nativa y descarte sin cambiar datos ni
política deportiva. El comportamiento global del banner sigue intacto. Quedan
pendientes los envíos nativos y casuísticas específicas que el inventario no
acredita, teclado, lectores de pantalla, scroll por gesto y demás formatos.
La sesión de fixtures queda preparada para continuar; todas las inyecciones
están desactivadas y el entorno local sigue activo, sin observabilidad.


Lectura pública independiente posterior: los cinco fixtures conservan sus
resultados, incluida la incidencia de tenis y sus partialSets. Evidencia
inspeccionada: ios-editors-preserved-20261007.json. El primer helper empleaba
nombres incorrectos para los goles de tanda y buscaba los parciales de abandono
en sets; se ajustó al contrato (homePenalties/awayPenalties e
incident.partialSets) antes de completar la comprobación. Esos errores del
helper no corresponden a regresiones de producto. git diff --check pasa; no se
repiten suites de código por esta fase de QA y documentación sin cambios de
producto. Se conserva la validación automatizada de la corrección anterior.


### Continuación: rechazo real de refresh iOS y regresión de transporte web (2026-10-07)

En iOS se establece una sesión exclusiva del participante y el proxy revoca esa
misma sesión mediante DELETE real 204 antes de entregar el login. Solo la fecha
de acceso presentada al cliente se aproxima a un minuto; la revocación y el
refresh posterior 401 son reales del backend. La app muestra el mensaje seguro
de sesión caducada, retira las colecciones privadas y mantiene el formulario de
login tras recargar el runtime. La ficha pública abierta por el enlace inicial
no tiene acciones de edición. Un nuevo login sin inyecciones recupera la identidad
y Administro 32 / Sigo 5. Evidencia ios-session-refresh-revoked-immediate,
anonymous, after-reload, account-after-reload y recovered-20261007; metadatos en
session-refresh-requests.jsonl, sin secretos.

En web, una lectura de colección con 401 inyectado seguida de un corte del socket
de refresh provocó reset a raíz y formulario anónimo. La cookie de sesión había
respondido 200 al restaurar justo antes. Esto contradice ADR-0062: un fallo de red
no provoca logout. La causa está en refreshWebSession: convertía cualquier rechazo
en false, devolvía el 401 inicial y el coordinador lo interpretaba como expiración.
Se conserva la barrera concurrente y se propagan los errores de transporte,
parseo y estados inesperados. Solo un refresh 401 devuelve false; 200 permite
repetir la operación. No se cambia la política aceptada ni el host de banners.

Las seis pruebas nuevas de transporte ejecutan el módulo real con límites
simulados: red y reintento, 429, 500, cuerpo inválido, 401 y barrera concurrente.
La suite Node completa pasa 78 pruebas. pnpm run check pasa tras corregir las
referencias a globals Node del test y estrechamiento del tipo generado 200/401;
esos ajustes no cambian el contrato. La primera exportación de esta fase usó
127.0.0.1 como API desde una página localhost y el navegador no conservó cookies
utilizables. Se excluye ese login del QA de producto y se reconstruye con caché
limpia y origen localhost consistente antes de repetir el recorrido.


Repetición visual sobre el bundle corregido con API localhost:

- Lectura protegida con 401 de un solo uso y corte de socket durante refresh:
  RequestErrorCard muestra common_network_error y Reintentar. Cuenta conserva
  qa_visual_player. Se retira el corte, se induce otro 401 de acceso y Reintentar
  completa un refresh real 200 y recupera Administro 32 / Sigo 5. Evidencia
  web-session-refresh-network-fixed, identity-kept y recovered-20261007.
- Refresh 500 inyectado: aparece common_request_error, sin title/detail privados;
  Cuenta conserva identidad. Evidencia web-session-refresh-500-fixed e
  identity-kept-20261007.
- Se revoca la cookie actual mediante DELETE real 204 justo antes del refresh.
  La renovación real responde 401, se resetea a raíz y Cuenta muestra login.
  Evidencia web-session-refresh-revoked-real y anonymous-20261007. La captura
  inmediata contiene la transición; no se acredita aquí la duración del banner.
  Un nuevo login normal recupera Administro 32 / Sigo 5, evidencia recovered.

Retrospectiva: conservar el 401 de acceso solo es correcto cuando la renovación
confirma rechazo de credenciales. Los errores técnicos deben llegar al adaptador
para ofrecer recuperación segura, conservando la decisión aceptada sobre sesión.
La revisión de cliente conserva adaptadores OpenAPI, mensajes existentes y
coordinador único; no añade dependencias, rutas, textos ni reglas deportivas.
La exportación web final genera 36 rutas. La suite Node pasa 78/78 y el check
completo pasa. No se modifican endpoints ni sus spans. Sigue pendiente la cobertura
global indicada en el inventario, incluida renovación visual Android y envíos
nativos de los editores revisados. Las inyecciones quedan en pass, las dos sesiones
de participante preparadas y los servicios locales de QA activos sin observabilidad.


### Guardado iOS preparado; interrupción por bloqueo del Mac (2026-10-07)

Se abre QA iOS voleibol 20261007 y se cambia el borrador del quinto set de
16–14 a 17–15. El árbol accesible confirma ambos valores y Guardar habilitado.
La llamada para activar Guardar y capturar el resultado termina al detectar el
Mac bloqueado. No se acredita envío ni éxito visual. La lectura pública posterior
confirma los cinco sets originales, incluido 16–14 y resultado 3–2; evidencia
ios-volleyball-save-interrupted-20261007.json. No requiere restauración por API.
El borrador nativo debe observarse de nuevo antes de continuar tras desbloquear.

Mientras la pantalla no está disponible se verifican por GET los cinco fixtures
previstos: pádel, bádminton, balonmano, tenis y voleibol están en curso y conservan
sus datos iniciales. Se guardan baselines para verificar cada escritura y su
restauración, ios-write-baselines-20261007.json. No se acredita esa preparación
como guardado nativo. Se revisa el adaptador compartido de resultados: usa la
operación generada y authenticatedApiFetch, trata 409 de forma recuperable y
conserva fallback seguro para otros estados. No se modifica producto ni se
repiten suites sin cambios. git diff --check pasa.

Retrospectiva: una interrupción de automatización durante un envío exige consultar
persistencia antes de repetirlo; no permite asumir ni éxito ni descarte. La
lectura independiente evita escribir de nuevo a ciegas. QA visual espera
el desbloqueo solicitado; los servicios de pruebas siguen activos sin observabilidad.


### Guardado y restauración de cinco resultados iOS (2026-10-07)

Tras desbloquear, se observa el borrador de voleibol 17–15 antes de enviar.
El primer intento posterior no persiste: el proxy temporal solo admitía GET,
POST, DELETE y OPTIONS; PUT terminaba en 501 antes del backend. Se añade el
reenvío PUT al helper privado y se repite. Este fallo de la herramienta se
excluye de los defectos del producto. No se modifica código del cliente.

Los controles se activan mediante accesibilidad. Se acredita envío nativo,
respuesta visible, reapertura y lectura pública independiente, sin inferir
teclado virtual, VoiceOver ni scroll por gesto.

- Voleibol: quinto set 17–15 guardado, ficha 3–2 y parciales actualizados;
  reapertura conserva 17–15. Se restaura 16–14 y se guarda por UI.
- Bádminton a 15: segundo juego 15–13 guardado y reabierto, conservando primero
  21–20. Se restaura 15–12 por UI.
- Balonmano: empate 25–25 con tanda 6–4 guardado, visible como 25 (6)–25 (4),
  y reabierto con ambos campos de tanda. Se restaura 5–4 por UI.
- Pádel: 6–4/7–5 guardado y reabierto; se restaura el segundo set 7–6 por UI.
- Tenis: abandono del visitante con parcial 3–3 guardado y reabierto. La respuesta
  conserva tipo retirement, lado away y victoria local. Se restaura 2–3 por UI.

Evidencia privada: ios-{volleyball,badminton,handball,padel,tennis}-saved-api-
20261007.json; capturas/árboles ios-*-saved-20261007 e ios-*-restored-save-
20261007 (voleibol saved-17-15). Cada cambio se confirma por GET antes de
restaurar. La lectura conjunta final ios-editors-preserved-20261007.json confirma
los cinco datos originales y estados en curso. No se finalizan los torneos.

### Error al guardar y banner por encima del popup iOS (2026-10-07)

Se inyecta un PUT 500 con problema desconocido y title/detail privados. En tenis,
el borrador 3–3 permanece, Guardar vuelve a estar habilitado y el banner muestra
common_request_error. La captura inmediata confirma que el banner está por
encima del popup nativo. Una captura posterior al primer intento ya había
perdido la ventana del banner; no se usa como evidencia de ausencia. Se repite
con captura inmediata, ios-result-save-500-global-banner-immediate-20261007.png.

Se repite el rechazo y se cierra el editor por el fondo: la ficha muestra el
parcial persistido 2–3 y el mismo banner continúa por encima tras desmontar el
popup, ios-result-save-500-banner-after-close-20261007.png/txt. Se confirma así
el comportamiento global aceptado en una segunda estructura nativa, además de
la ruta fullScreenModal previamente revisada. El GET independiente confirma
que los rechazos no alteraron el resultado. La inyección se desactiva al terminar.

Retrospectiva: la evidencia de guardado completa la validación de formulario,
pero no acredita todos los deportes, formatos ni incidencias nativas. Una prueba
de escritura necesita que el proxy soporte su método HTTP. Para feedback de
vida corta se captura durante su ventana visible; una captura tardía no prueba
un fallo de superposición. Los banners nativos mantienen su host global y
sobreviven al cierre del popup. Sin cambios de producto en esta fase, no se
repiten suites de código; git diff --check pasa. El entorno permanece activo y
sin observabilidad para continuar QA.


### Renovación de sesión Android: éxito y fallos técnicos (2026-10-07)

Se reutiliza la development build instalada en Pixel_API_34 y el Metro activo,
con ADB autorizado previamente por el usuario y reverse de 8080/8083. No se
cambia código de producto, contrato de banners, dependencias ni servicios.
El arranque inicial compila el bundle Android; la pantalla de splash durante
esa carga y los fallos de captura XML transitorios no acreditan un crash.

El proxy modifica únicamente expiresAt en un login real 200 para provocar la
renovación inmediata; los secretos y la renovación de backend son reales.
POST /sessions/refresh responde 200 y la biblioteca muestra Administro 32 y
Sigo 5. Evidencia android-refresh-success-library-20261007.xml/png.

En otra sesión exclusiva, el proxy corta la conexión antes de remitir refresh.
La biblioteca muestra common_network_error con Reintentar. Sobre la misma
sesión, refresh 500 y 429 muestran common_request_error sin exponer el título
ni detalle sintéticos del backend. Al retirar la inyección, Reintentar obtiene
refresh real 200 y vuelve a mostrar 32/5 sin otro login. Evidencia
android-refresh-network-error-20261007, android-refresh-500-20261007,
android-refresh-429-20261007 y android-refresh-network-recovered-20261007.
La cuenta seguía identificada después del primer corte; recuperar su biblioteca
con los mismos secretos acredita que los fallos técnicos no borraron la sesión.

La primera entrada automatizada del segundo login respondió 401; hubo acciones
solapadas y el teclado no había terminado. Se excluye de los casos de refresh.
Se vacían y rellenan los campos secuencialmente, y el login real 200 precede a
los fallos de renovación registrados. No se guarda el contenido de campos como
evidencia entregable. session-refresh-requests.jsonl conserva solo metadatos.

Retrospectiva: esperar la finalización de cada interacción es parte de la prueba.
Una captura durante carga o un login mal introducido no demuestra un defecto de
renovación. Los fallos técnicos y el rechazo 401 se recorren por separado con
la misma política de identidad ya aceptada. Esto no acredita caducidad temporal
real, accesibilidad completa ni concurrencia visual de varias renovaciones.


### Revocación real de refresh Android (2026-10-07)

El proxy revoca una sesión exclusiva con DELETE /sessions 204 antes de entregar
el login real 200. expiresAt próximo es simulado; la revocación y el posterior
POST /sessions/refresh 401 proceden del backend. Cuenta vuelve al formulario y
Torneos retira las colecciones privadas y muestra el estado sin sesión. Evidencia
android-refresh-revoked-20261007 y android-refresh-revoked-library-20261007.
La captura inmediata falló durante la transición; no se acredita visualmente
el banner de caducidad en esta pasada.

Se reabre la app con force-stop/start sin borrar datos. El launcher tarda en
arrancar y conserva una conexión de desarrollo antigua; se reconecta al Metro
8083 de QA. Studio muestra aviso de falta de memoria durante la indexación. Se cierra el IDE
y se arranca Pixel_API_34 independiente, conservando datos. Hasta observar el
contenido cargado, se mantiene pendiente la persistencia tras reapertura.
Finalmente se recupera System UI y se observa Home anónima, Torneos sin
colecciones privadas y Cuenta con login, sin borrar datos de la app. Evidencia
android-refresh-revoked-reopened-library/account-20261007. Un nuevo login normal responde 200, recupera la cuenta y la biblioteca 32/5.
Evidencia android-refresh-relogin-account/library-20261007. No se atribuye el splash prolongado
al rechazo de sesión ni se presenta como causa raíz resuelta.


Recuperación del entorno Android: la pasada independiente muestra un ANR de
com.android.systemui; no es un ANR acreditado de com.fasttourney.app.local.
Se conservan captura android-independent-current.png y árbol de System UI.
Tras Esperar con destino explícito display 0, el contenido vuelve a estar
operativo. Las primeras pulsaciones no mostraron efecto inmediato; no se
atribuye causalidad exclusiva al display ni al cierre de Studio. No se cambia
código de producto para acomodar este bloqueo del entorno.


### Banner global Android sobre popup: bloqueo del backdrop (2026-10-07)

Se abre el resultado de tenis con abandono visitante y parcial persistido 2–3.
Se cambia únicamente el borrador a 3–3 y se inyecta PUT 500 antes de reenviar al
backend. El banner seguro se ve por encima del popup y el borrador sigue 3–3.
La captura android-result-500-banner-immediate-20261007 aún muestra la carga;
la evidencia del banner es android-result-500-popup-20261007.xml/png.

Se repite el rechazo, se espera a que aparezca el banner y se pulsa el backdrop
lateral (20,1200) del display 0. El popup sigue abierto con el banner encima:
android-result-500-after-popup-close-immediate-20261007.png y
android-result-500-after-popup-close-20261007.xml/png. El nombre de esos archivos
describe la acción intentada; no acredita un cierre exitoso. Tras desaparecer el
banner, la misma pulsación cierra el popup y la ficha sigue mostrando 2–3:
android-result-popup-closed-after-banner-20261007.xml/png.

Incidencia abierta: el host Android usa un Modal raíz transparente, y la ventana
visible intercepta la interacción destinada al popup inferior. La observación
comparada confirma el bloqueo del backdrop; atribuir todos los gestos o todas
las rutas al mismo fallo necesitaría más recorridos. No se ha podido acreditar
persistencia del banner tras cerrar este popup, porque el propio host impide el
cierre durante su vida. No se mueve el host a Screen ni se altera el contrato
aceptado de superposición global y supervivencia a navegación. Se conserva como
problema técnico a resolver respetando esa decisión.

Las inyecciones vuelven a pass. Las cinco lecturas independientes confirman los
fixtures originales, incluido tenis 2–3. El emulador queda independiente del IDE,
con el participante autenticado; API/proxy, web y Metro siguen disponibles (200),
sin observabilidad. No hay cambios nuevos de producto en esta pasada y no se
repiten suites sin cambios de código. git diff --check pasa.

Retrospectiva: ver un banner encima de un popup no basta para dar por validado
su comportamiento. La prueba debe incluir interacción en la capa inferior y
supervivencia al cierre. Un arreglo debe preservar el alcance global aceptado;
reducirlo a la pantalla evitaría el síntoma cambiando el comportamiento decidido.


### Control iOS previo al adaptador Android (2026-10-07)

El usuario autoriza el adaptador Android y pide comprobar que no rompa iOS ni
oculte un defecto previo. Antes de editar, Device Hub/iPhone 18 Pro reproduce
PUT 500 en el popup de tenis: aviso sobre popup, backdrop cierra y aviso sigue
en ficha. Una repetición cierra popup y navega a Participantes: aviso permanece
sobre la ruta nueva. Una llamada conjunta confirma aviso presente y después toque
sobre él: desaparece y el editor sigue abierto. Capturas y árboles en esta
conversación. Los resultados no se guardan; el proxy impide llegar al backend.

El arrastre CUA intentado no descarta el aviso en la primera captura; no se da
por aprobado ni se atribuye todavía al host. El primer cierre capturado en dos
llamadas separadas ocurrió ya sin banner, por lo que se excluye como evidencia
de supervivencia. Se repite en una llamada conjunta para evitar el autocierre.


### Adaptador Android autorizado y regresión iOS (2026-10-07)

ADR-0150 aceptado tras «Autorizar adaptador Android». El módulo local Expo
TMGlobalFeedback compila e instala en Pixel API 34. El provider sigue siendo dueño
del aviso, cierre y temporizador; ModalDialog registra solo el token de su ventana.
La implementación exclusiva .android.tsx no requiere el módulo en iOS/web.

Evidencia Android en /tmp/tm-product-qa-20261004/:

- android-adapter-fast-closed-20261007.png: backdrop cierra y aviso sigue en ficha.
- android-adapter-fast-participants-20261007.png: aviso sobre la ruta nueva.
- android-adapter-stable-before/after-tap-20261007.png: aviso observado, toque lo
  descarta y conserva el popup.
- android-adapter-final-before/after-swipe-20261007.png: aviso observado antes del
  arrastre, desaparece después y conserva el popup. Las capturas anteriores de
  toque tomadas durante la carga no se usan como evidencia.
- android-adapter-replacement-after-old-deadline-20261007.png: el aviso reemplazado
  permanece tras el plazo del anterior; replacement-expired acredita autocierre.
- android-adapter-resumed-active/expired-20261007.png: volver desde Home conserva
  el aviso vigente y después lo elimina al vencer.
- android-adapter-new-popup-active/keyboard-active-20261007.png: un popup abierto
  con aviso previo queda debajo; el campo puede enfocarse. La segunda captura
  aún no muestra teclado: no acredita la matriz completa de teclado.

Control iOS posterior, en Device Hub/iPhone 18 Pro: PUT 500, cierre por backdrop y
navegación a Participantes con aviso aún visible; toque sobre el aviso lo descarta
y conserva el editor; autocierre observado en la ruta destino. No hay regresión
observada en esos recorridos respecto al control previo. No equivale a declarar
iOS completamente validado.

El arrastre iOS no descarta el aviso en repetición con captura previa que confirma
su presencia. Se probaron captura del responder y el reconocedor ya incluido en
react-native-gesture-handler; ninguno produjo una mejora acreditada y todos los
experimentos se retiraron. Se conserva FullWindowOverlay y el gesto original.
Queda abierto diagnosticar el recorrido de toques y distinguir un defecto del
producto de una limitación del arrastre en Device Hub. El control de desplazamiento
del popup sí mueve contenido; no basta para localizar la causa en el host del
banner. No se introduce ni se da por necesario otro adaptador iOS.

Validación automática: pnpm check, 78 pruebas Node, exportación web de 36 rutas y
compilación Kotlin/Android pasan. No se toca una operación OpenAPI con el adaptador;
las mutaciones del fixture siguen usando su feature y apiFetch. Los PUT de esta
pasada son 500 inyectados antes del backend.

Pendientes de esta primera pasada (actualizados por la repetición final inferior):
teclado realmente visible, recreación de Activity, accesibilidad nativa y prueba
de giro dentro del alcance permitido: app.config.ts fija orientación portrait y no se cambia para este QA.
También permanece abierto el diagnóstico del gesto iOS. Se mantienen los servicios
de QA mientras continúa el trabajo, sin observabilidad.

Retrospectiva: ventana visible y ciclo de vida son comprobaciones diferentes.
El control iOS previo evita atribuir al adaptador Android un fallo anterior.
Las hipótesis de arreglo que no pasan el recorrido se retiran antes de entregar.


#### Repetición sobre WindowManager, mecanismo final

La prueba ampliada de Atrás descubre que PopupWindow registra su propio callback
en API 34: inicialmente oculta el banner; restaurarla conserva el aviso pero
consume Atrás. Se descarta ese mecanismo dentro del adaptador aceptado. El host
final usa WindowManager.addView/TYPE_APPLICATION_SUB_PANEL con NOT_FOCUSABLE y
NOT_TOUCH_MODAL, sin interceptar Atrás. Build recompilada e instalada, sin cambios
en JavaScript iOS/web. Las capturas anteriores de PopupWindow son exploratorias;
no sustituyen esta repetición.

Evidencia final en el mismo directorio:

- android-window-manager-before/after-tap-20261007.png y before/after-swipe:
  descarte observado con popup conservado.
- android-window-manager-origin-closed-frame-0/2-20261007.png: popup y ficha
  desmontados, Inicio visible con el banner todavía activo. Frame 3 ya acredita
  su autocierre. La prueba anterior que acababa en Inicio sin aviso no demuestra
  supervivencia y se excluye.
- android-window-manager-replacement-after-old-deadline/expired-20261007.png:
  reemplazo sobrevive al plazo anterior y después vence el suyo.
- android-window-manager-resumed-active/expired-20261007.png: retorno desde Home
  con aviso vigente y desaparición posterior, sin ventana huérfana.
- android-window-manager-new-popup-active/keyboard-active-20261007.png: nuevo
  popup bajo un aviso previo, campo enfocado y teclado numérico visible con el
  banner aún arriba. dumpsys confirma InputMethod isOnScreen=true. Se cierra el
  teclado sin cambiar el valor 2 y luego vence el aviso.

El gesto nativo conserva el ID de ACTION_DOWN: una liberación de un gesto iniciado
sobre un aviso anterior no descarta su reemplazo. El provider raíz mantiene su
propia protección por ID.

Control iOS adicional sobre código final: banner observado, backdrop cierra el
editor, pulsación en el borde no cubierto del cierre de ficha inicia el regreso
a la biblioteca con el banner todavía encima (capturas de esta conversación).
No se conserva ninguno de los experimentos de gesto iOS.

La build final de WindowManager pasa en 13 s y se instala antes de repetir estas
pruebas. Check final y exportación de 36 rutas pasan. La salud correcta de Metro
es localhost:8083/status (IPv6), API/proxy /healthz devuelve 200 y web 8082 devuelve
200. Un sondeo IPv4 a Metro produjo un falso negativo; no se reinicia el servicio.

Retrospectiva: sin foco no significa sin callback de Atrás. Una API pública puede
tener políticas de navegación ajenas al contrato del componente. La ventana hija
resulta el mecanismo suficiente; no se recurre a reflection ni parches internos.

#### Recreación y límites de la validación final

La build con título de ventana localizado compila en 18 s y se instala. El
TextView conserva liveRegion polite y semántica clickable; el título usa el mismo
mensaje localizado, sin nombre técnico expuesto. No se usa accessibilityTitle:
el SDK instalado lo marca @hide, por lo que no forma parte de la API pública.
Esto no sustituye una prueba completa con TalkBack/VoiceOver, que sigue pendiente.

Desde Inicio cargado, font_scale 1.1 y vuelta a 1.0 recrean la Activity y conservan
la sesión. Se vuelve a abrir la ficha por URL canónica y su editor; PUT 500 genera
un nuevo aviso y el toque lo descarta conservando el popup. Evidencia:
android-window-manager-loaded-recreation-1.1/1.0-20261007.png y
android-window-manager-post-recreation-active/dismissed-20261007.png. Escala final
1.0, sin nuevo fatal en esta repetición.

Una repetición anterior, cambiando la escala mientras arrancaba Expo después de
instalar, produjo a las 17:48:30 un fatal en DevLauncherAppLoader:
«App react context shouldn't be created before». Se recuperó abriendo de nuevo
el development client. Log android-window-manager-runtime-review-20261007.log.
No hay control suficiente para atribuirlo al adaptador ni para descartarlo como
fallo del entorno. Queda abierto el arranque solapado con recreación; no se declara
resuelto todo el ciclo de Activity ni se parchea Expo como parte de este cambio.

Tras desbloquear el Mac se repite iOS con observación del banner y arrastre en la
misma llamada de automatización, antes del plazo de autocierre. El aviso sigue
visible. No se cambia código iOS ni se acredita la causa: defecto del gesto o
limitación de Device Hub. Sigue sin justificarse otro adaptador nativo.

Checklist cliente: módulo solo Android y stub iOS/web; textos desde catálogos,
tokens compartidos, provider global dueño del aviso, anclajes sin estado funcional,
sin nuevos permisos ni dependencias externas; check, 78 pruebas y exportación web
válidos para el JavaScript final. No hay nuevas operaciones OpenAPI. La matriz
completa permanece abierta por accesibilidad, gesto iOS y arranque con recreación.

Retrospectiva: probar tras una recreación estable y provocar una durante el
arranque ejercitan caminos distintos. Una pasada correcta no borra el fatal de
la otra. Tampoco debe confundirse aislar el módulo Android con validar todos los
gestos y lectores de pantalla de iOS.

#### Diagnóstico y corrección del gesto iOS, código final

Se usan trazas temporales de touch/responder y el inspector Hermes local, sin
textos de negocio ni credenciales. El probe del inspector valida la captura.
La primera ausencia de trazas no se interpreta como pérdida de toques: el runtime
no contenía aún la instrumentación. Se recarga iOS explícitamente desde el menú
de Expo y se confirma QA_BANNER_PROVIDER_RENDER. Las hipótesis y controles iOS
anteriores conservan su valor exploratorio, pero no acreditan que cada ajuste de
Fast Refresh hubiera llegado al runtime; se excluyen como prueba de su eficacia.

Con código instrumentado confirmado, el gesto produce TOUCH_START, movimiento
(0, -79.666667), TOUCH_MOVE, RELEASE (0, 0) y TOUCH_END. El código instalado de
React Native Libraries/Interaction/PanResponder.js pone dx/dy a cero en
onResponderGrant: un arrastre rápido con un único move pierde en la evaluación
final la distancia que hizo reclamar el responder. La evidencia local vive en
ios-banner-touch-events-20261007.jsonl. El host recibe los toques; estos datos no
justifican otro adaptador iOS.

Se conserva grantDy al reclamar el gesto y se suma a movimientos/liberación
posteriores; se limpia al liberar o cancelar. No cambian slop, distancia, velocidad,
temporizador, superposición ni dueño global. Se retiran todas las trazas y se
recarga explícitamente el código corregido. Validación visual en Device Hub:

- Aviso observado antes de arrastrar desde (700,495) a (700,439): desaparece y el
  editor sigue abierto con 2–0.
- Arrastre corto de (700,495) a (700,480): el aviso permanece.
- Nuevo aviso, cierre por backdrop y cierre de ficha: Inicio visible con banner
  todavía encima. Se acredita la supervivencia al desmontaje de ambas estructuras.

Se usa el fixture QA Web refresh 20261007 sin cambiar 2–0; cada PUT recibe el 500
inyectado antes del backend. Las capturas quedan en esta conversación. No se
introduce adaptador iOS, dependencia nueva ni modificación nativa iOS. El cambio
JavaScript corrige un cálculo técnico para cumplir el gesto ya decidido.

Retrospectiva: confirmar el código que realmente ejecuta el dispositivo precede
al diagnóstico. Recibir toques y medir correctamente el gesto son límites
diferentes. Un movimiento anterior al grant no puede desaparecer del umbral de
cierre. La matriz deja de tener abierto este caso concreto de arrastre iOS;
siguen pendientes lectores de pantalla y arranque Android con recreación.

Cierre de esta pasada: check (formato, lint, TypeScript y OpenAPI), 78/78 pruebas
Node y exportación web de 36 rutas pasan después de la corrección del gesto.
La compilación nativa final sigue siendo la validada de 18 s; el ajuste posterior
es JavaScript. No quedan trazas QA_BANNER en el cliente. Se verifican de nuevo
los cinco fixtures y el marcador 2–0 de refresh; se retira la inyección 500.
API, web y Metro responden 200 y permanecen activos para continuar el QA.
No se declara cerrada toda la matriz ni se integra/publica este workspace sucio.


#### Control del fallo de arranque Android con recreación

Se repite la secuencia en el emulador API 34: force-stop de la app de QA, apertura
del development client, espera de 0.3/1/3 s, font_scale 1.1, espera de 2 s, vuelta a
1.0. La escala y la app se restauran en finally; no se borran datos ni se detienen
API, web, Metro o bases de datos. Con el adaptador: fatal Expo en 2/3 recorridos
(1 y 3 s). Evidencia android-startup-recreation-adapter-20261007.json/log.

Para aislarlo se compila e instala un control con android.modules vacío y host
JavaScript Android stub. Se comprueba que ExpoModulesPackageList generado no
contiene GlobalFeedbackModule: el controlador y sus hooks no se registran ni se
ejecutan. El mismo fatal se reproduce en 3/3 recorridos del control. Evidencia
android-control-native-module-registry-20261007.txt y
android-startup-recreation-control-no-adapter-20261007.json/log. La build de control
compila en 15 s. El código original, configuración y APK se conservaron antes.

El código fijado de expo-dev-launcher 57.0.14, DevLauncherAppLoader.kt:49, exige
currentReactContext == null al crear el delegado de Activity; la recreación
solapada con el arranque llega con contexto existente. La traza y el control
confirman que este fallo no requiere el adaptador. No demuestran que todos los
caminos de arranque o una build de producción estén validados. El código del
cargador observado pertenece a src/debug; no se parchea Expo como parte de ADR-0150.

Se restauran los originales y se recompila el adaptador para no dejar instalada
la build de control. Queda la incidencia del development client para su propio
seguimiento, separada de la validación del host global.

Retrospectiva: un fallo en una dependencia no se atribuye por proximidad temporal
al último módulo añadido. Desregistrar realmente el módulo y repetir la misma
secuencia aporta un control más fuerte que dejar de mostrar el banner. La ausencia
del adaptador en el registro generado forma parte de la evidencia del control.


#### Texto ampliado y Android Lint

La build completa restaurada compila en 8 s. Se confirma GlobalFeedbackModule en
el registro generado antes de reinstalar y se espera Inicio cargado antes de
cambiar font_scale a 1.5. Ficha y popup conservan texto legible, los controles
refluyen y Guardar sigue alcanzable mediante desplazamiento. PUT 500 inyectado
muestra el banner completo, y el toque lo descarta conservando editor y 2–3.
Evidencia android-banner-large-text-*-20261007.xml/png y
android-window-manager-large-text-active/dismissed-20261007.png. La captura active
incluye la transición de guardado; dismissed acredita el popup recuperado. No se
usan para afirmar una prueba de lector de pantalla. Escala restaurada a 1.0 y
modo de inyección devuelto a pass.

Se amplía la verificación con :global-feedback:lintDebug. Android Lint aborta en
:react-native-worklets:lintAnalyzeDebug con «Cannot find a KaModule for the
VirtualFile», dentro de UastGradleVisitor al analizar build scripts. Pedir
:global-feedback:lintAnalyzeDebug con arquitectura arm64 repite el mismo fallo de
la dependencia. Logs android-global-feedback-lint-20261007.log y
android-global-feedback-lint-analysis-20261007.log. No hay informe completo del
módulo: Lint queda bloqueado por esta excepción de herramienta, no se presenta
como aprobado ni se deshabilitan reglas o se parchean dependencias para ocultarlo.

Retrospectiva: compilar Kotlin y pasar el check TypeScript no equivale a pasar
Android Lint. Una excepción interna de análisis tampoco se trata como un defecto
funcional demostrado del banner. Se conservan ambos alcances por separado.


#### Francés, tema y conservación del plazo

Idioma por app Android fr-FR, tema del sistema oscuro y fixture de tenis: ficha,
editor y error común aparecen en francés. La captura
android-window-manager-french-theme-changed-active-20261007.png acredita el mensaje
completo y los tokens oscuros. french-dark-active se tomó durante la transición y
no muestra el banner; se excluye como evidencia de su visibilidad. La captura
posterior al primer cambio de tema aún conserva colores oscuros: no se usa para
acreditar la actualización de tokens en el aviso activo.

Se repite en español empezando en claro, con espera tras observar la ventana.
android-window-manager-theme-light-before-20261007.png acredita el aviso claro;
theme-dark-frame-2 y frame-4 acreditan el mismo mensaje ya oscuro a 3.74/5.32 s.
A 6.7 s no queda su ventana (theme-dark-expired), sin renovar el plazo al actualizar
los colores. El formulario conserva 2–3. Los PUT son 500 inyectados antes de llegar
al backend. En finally se restaura night=no y se retira la inyección; el idioma por
app vuelve a su lista original vacía y font_scale permanece 1.0.

La activación temporal de TalkBack/VoiceOver se ha solicitado separadamente: no
se habilitan lectores ni nuevos accesos a la pantalla sin respuesta explícita.
Esta pasada valida legibilidad, localización y tokens; no acredita lectura hablada,
foco ni navegación por lector. Los servicios de QA permanecen activos.

Retrospectiva: observar una ventana en dumpsys no garantiza que el primer frame ya
sea visible. Cambiar un ajuste del SO tampoco acredita que el cliente ya haya
renderizado sus nuevos tokens. Las capturas de transición se conservan, pero se
excluyen de las afirmaciones que no demuestran.

#### Arrastre corto y análisis aislado del módulo Android

Se comprueba un arrastre ascendente de 50 px durante 800 ms sobre el aviso:
permanece visible y vuelve a su posición; el toque posterior lo descarta y
conserva el editor con el parcial 2–3. Evidencia
android-window-manager-short-swipe-retained/short-swipe-then-tap-20261007.png y
android-banner-short-swipe-20261007.json, en el directorio temporal de QA.
El PUT 500 se intercepta antes del backend y la inyección vuelve a pass en finally.
Los cinco fixtures de resultados se verifican intactos mediante GET.

Se obtiene un informe real del módulo con :global-feedback:lintAnalyzeDebug y
:global-feedback:lintReportDebug, arquitectura arm64-v8a y exclusión explícita de
:react-native-worklets:lintAnalyzeDebug. No se cambian reglas, fuentes ni
dependencias. El informe lint-results-debug.txt del módulo contiene 0 errores y
6 avisos: ViewConstructor para el constructor ExpoView con AppContext y cinco
recomendaciones UseKtx (dos ubicaciones duplicadas) para Color.parseColor.
No se añade un constructor artificial sin AppContext ni se modifica código solo
para silenciar recomendaciones de estilo. El análisis completo con la tarea de
worklets continúa bloqueado; el resultado aislado no lo sustituye.
Logs android-global-feedback-isolated-lint{-report}-20261007.log. Los informes
generados se conservan también en el directorio temporal de QA.

iOS visual queda temporalmente detenido por el bloqueo del Mac, comunicado al
usuario; Android y los servicios de QA siguen activos. No se habilitan lectores
de pantalla mientras siga pendiente su autorización específica.

Retrospectiva: aislar una tarea defectuosa permite obtener evidencia útil del
módulo propio, siempre que se declare la exclusión y no se presente como un
análisis completo. Un arrastre insuficiente y el descarte posterior verifican
también la recuperación del gesto, además del caso de cierre directo.

#### Incidencia abierta: retorno tras vencimiento en segundo plano

Con el error inyectado visible se envía la app a Inicio del sistema, se esperan
7 s y se retorna mediante el intent MAIN/LAUNCHER de MainActivity. La captura
android-window-manager-background-expiry-resumed-20261007.png no muestra el
banner, pero Android presenta después «Fast Tourney Local no responde».
Este recorrido no se considera aprobado: la ausencia de ventana no acredita
una recuperación funcional cuando el proceso no responde.

El ANR de las 20:03:17 indica Input dispatching timed out al recibir
FocusEvent(hasFocus=true), con espera de 5126 ms. Logcat registra cientos de
frames omitidos. Se conservan lastanr, logcat y bugreport del emulador en
android-background-expiry-*-20261007.* dentro del directorio temporal de QA.
La traza de la app contiene fallo de tombstoned y timeout de debuggerd; no
aporta una pila que atribuya el bloqueo al adaptador o a otra dependencia.
No se deduce causalidad de la secuencia temporal ni se cambia el contrato.

Esperar no recupera la app. Se reinicia únicamente su proceso, se reabre la
development build y se confirma Inicio autenticado con actividad reciente
(android-recovered-home-stable-20261007.xml/png). No se borran datos y los
servicios permanecen activos. La inyección vuelve a pass en finally.
La investigación y repetición controlada del retorno quedan abiertas.

Controles posteriores con la misma espera de 7 s y el mismo intent de retorno:
Inicio sin banner y editor de tenis abierto/desplazado sin banner recuperan el
contenido y no cambian el registro lastanr. Evidencia
android-background-home-control-20261007.json/xml/png y
android-background-editor-control-20261007.json/xml/png. Estos controles acotan
el caso, pero no demuestran todavía la causa del primer bloqueo.

La primera repetición con aviso activo, 7 s en segundo plano y retorno al mismo
editor tampoco genera otro ANR; el banner vencido está ausente y el editor
conserva el parcial. Evidencia android-background-expiry-repeat-20261007.json y
android-background-expiry-repeat-stable-20261007.xml/png. Se registra como
repetición satisfactoria, sin cerrar el bloqueo inicial intermitente ni atribuir
su causa. No se modifica código por una hipótesis todavía no demostrada.
El toque en el backdrop cierra después el popup y devuelve la ficha con 2–3
(android-background-expiry-repeat-close-20261007.xml/png), confirmando interacción
real en esta repetición además de la lectura del árbol.

#### Ampliación del control Android y notificaciones iOS

Tres ciclos adicionales con aviso inyectado, 7 s en segundo plano y retorno
MAIN/LAUNCHER recuperan el editor sin nuevos ANR y sin resucitar el aviso vencido.
android-background-expiry-series-20261007.json registra los tres resultados;
android-window-manager-background-expiry-series-{1,2,3}-20261007.png conserva
las capturas. Estos controles no eliminan la incidencia inicial sin causa
confirmada. Se retira la inyección en finally y se conservan los datos.

Con el Mac accesible de nuevo se retoma iOS, sesión qa_visual_player. La lista
expone 31 enlaces y 31 acciones individuales. La última acción de apertura
accesible llega a QA Liga de equipos con acentos y nombres largos; la ficha
muestra dos 3–0 y un partido pendiente. Clasificación muestra Norte y Sur con
3 puntos/+3 y Peñas con 0/-6, coherente con esos partidos. Al cerrar clasificación
y ficha se recupera la lista en su tramo final, con la notificación más antigua
del 4/10/2026 16:34:26 visible.

Se abren las confirmaciones individual y masiva y se cancelan ambas. Diálogos
centrados sobre la ruta, texto localizado, aviso de irreversibilidad y Cancelar
operativo. El recuento final del árbol sigue siendo 31 enlaces/31 acciones.
No se acepta ningún borrado; la ejecución destructiva sigue fuera de esta pasada.
Capturas y árboles de Device Hub quedan en la conversación de QA.

Los intentos de drag/scroll no desplazan visualmente notificaciones, ficha ni
estadísticas horizontales en esta sesión de control. La activación AX del último
elemento sí lo lleva a la vista, pero no acredita el gesto manual. Se conserva
pendiente la prueba de desplazamiento real y no se atribuye todavía un defecto
a una pantalla concreta ni se modifica su implementación por esta observación.

Retrospectiva: ampliar repeticiones sin fallo mejora la evidencia de un caso
intermitente, pero no explica su primer fallo. Acceder por AX y recuperar la
posición de la lista son resultados distintos de desplazarla mediante gesto;
cancelar una confirmación tampoco valida su operación destructiva.

#### Sugerencias iOS: éxito real y límites del borrador

Se verifica el destino SMTP_ADDR del backend local: mailpit:1025. Desde Inicio
en iOS se envía una sugerencia ficticia identificada QA local iOS 20261007.
La interfaz muestra «¡Gracias! Hemos recibido tu sugerencia.», limpia el campo
y deshabilita Enviar. El aviso desaparece después. Mailpit recibe exactamente
un mensaje nuevo cuyo cuerpo coincide con el texto de prueba. Evidencia técnica
ios-suggestion-success-20261007.json y baseline de IDs en el directorio temporal
de QA; capturas y árboles del simulador en la conversación. No se utiliza un
destino externo ni se imprimen direcciones o credenciales.

Sin más envíos se comprueban 7 caracteres ASCII (Enviar deshabilitado), 8
(habilitado) y entrada ASCII de 1.001 (el valor observable queda en 1.000 y Enviar
habilitado). Se selecciona todo y se borra: campo vacío y Enviar deshabilitado.
No se presenta esta prueba como validación Unicode: el canal de entrada no
introdujo los emoji intentados. El setter AX del grupo tampoco está soportado;
la comprobación se hace mediante entrada de teclado y valor realmente observado.

Se amplía la calibración del arrastre fuera de FastTourney: en la pantalla de
inicio de iOS, Página 2 de 2, el gesto horizontal tampoco cambia de página.
Se vuelve a la app sin tocar ajustes ni otras apps. La prueba manual de gesto
permanece pendiente por el canal de control, sin cambiar ScrollView ni el host
global a partir de esta observación.

Retrospectiva: confirmar un banner de éxito no demuestra por sí solo la entrega;
contrastar el buzón local cierra el recorrido. Los límites se acreditan con el
valor observado, no con el texto solicitado al automatizador. Un control fuera
del producto permite acotar un fallo de entrada antes de proponer una corrección.

#### Android: lectura de ligas y mixtos ya finalizados

Se leen por API siete fixtures web-formats finalizados y se abren por sus enlaces
canónicos en Android. La liga de fútbol a dos vueltas conserva dos empates 2–2
y clasificación con dos posiciones 1, dos puntos y cuatro goles a favor/en contra.
En baloncesto, balonmano y voleibol se recorren liga y mixto: título correcto,
estado finalizado y ausencia de acciones Añadir/Editar resultado en el árbol.
Las seis clasificaciones abren mediante su botón de UI y presentan los equipos,
estadísticas propias del deporte y puntos. Se inspeccionan sus capturas.

Las finales mixtas muestran Web Equipo 1 ganador: 80–70 en baloncesto, 20–19
en balonmano y 3–0 con tres sets 25–0 en voleibol. La liga de voleibol conserva
el 3–2 y sus cinco parciales 25–23, 23–25, 25–23, 23–25, 16–14. Las tablas de
los mixtos conservan tres partidos de liga por equipo tras la final, sin sumar
la eliminatoria a esa clasificación. No se guardan resultados durante esta pasada.

Evidencia native-formats-read-baseline/preserved-20261007.json,
android-native-formats-read/standings-20261007.json y
android-native-format-*-20261007.xml/png en el directorio temporal de QA.
La lectura posterior confirma los siete estados y listas de campeones intactos.
API, proxy, preview, Metro y Mailpit responden 200 y permanecen activos.

En la tabla del mixto de voleibol se realiza un swipe horizontal real: aparecen
SC/CS/TF/TC/CT, incluidos infinito y cocientes localizados 0,5. Equipos, posiciones
y puntos 9/6/3/0 permanecen fijos y alineados. Evidencia
android-native-format-volleyball-mixed-stats-scrolled-20261007.xml/png. La captura
confirma el contenido renderizado; las cajas recortadas de nodos fuera del
viewport en el XML no se interpretan como un solapamiento visual.

Alcance: valida lectura nativa del resultado final y navegación a las tablas,
no acredita creación, edición ni transiciones completas de estos seis formatos
desde Android. La matriz conserva esa cobertura parcial.

Retrospectiva: reutilizar fixtures terminados permite contrastar rápidamente
representación y navegación entre plataformas sin modificar datos. Separar esa
evidencia del ciclo completo evita convertir una revisión de lectura en una
afirmación de cobertura funcional total.


#### Ciclos de liga nativos: baloncesto, balonmano y voleibol

Se preparan por API seis torneos publicados nuevos de dos equipos, tres para
Android y tres para iOS. La preparación no acredita el formulario de creación.
En cada plataforma, la interfaz inicia la liga a una vuelta, añade y guarda el
resultado, abre la clasificación, solicita y confirma la finalización y muestra
el campeón. La ficha finalizada conserva el marcador y deja de ofrecer edición.
En iOS también se abre la tabla final desde la celebración en los tres casos.

| Plataforma | Deporte | Resultado | Campeón |
| --- | --- | --- | --- |
| Android | Baloncesto | 80–70 | Ciclo Local |
| Android | Balonmano | 20–19 | Ciclo Local |
| Android | Voleibol | 3–0: 25–20, 26–24, 25–23 | Ciclo Local |
| iOS | Baloncesto | 81–79 | iOS Local |
| iOS | Balonmano | 20–18 | iOS Local |
| iOS | Voleibol | 3–0: 25–21, 25–23, 26–24 | iOS Local |

Las tablas iOS muestran 2/1 puntos en baloncesto, 2/0 en balonmano y 3/0 en
voleibol. En balonmano el árbol confirma GF/GC 20/18 y 18/20, diferencias +2/-2.
En voleibol confirma 76/68 puntos de tantos, cocientes 1,118 y 0,895 e infinito
en sets del ganador; las columnas adicionales se acreditan en el árbol accesible,
sin presentar esta lectura como un swipe manual iOS.

Tras un bloqueo del Mac, typeText deja vacíos los campos de voleibol. Se comprueba
la captura y se utiliza setValue sobre los campos de texto que sí lo soportan;
la captura visual confirma los seis valores y los sets cuarto/quinto vacíos antes
de guardar. La prueba no acredita el teclado manual después de ese bloqueo.

Una lectura independiente de la API comprueba en los seis torneos formato league,
estado completed, resultado, parciales de voleibol y lista exacta de campeones.
La comprobación de los cinco fixtures anteriores de editores iOS confirma que sus
resultados e incidencias se conservan. Evidencia privada:
native-league-cycles-persisted-20261007.json,
android-native-league-cycle-*-20261007.json,
android-cycle-*-20261007.xml/png e ios-editors-preserved-20261007.json; árboles y
capturas iOS en esta conversación. No se cambia código de producto en esta fase.

Alcance: estos seis ciclos amplían las ligas nativas; no cierran los recorridos
mixtos de estos deportes ni la matriz completa de formatos. Los servicios de QA
permanecen activos mientras continúan las pruebas autorizadas.

Retrospectiva: un torneo mínimo de dos equipos permite verificar la transición
completa desde inicio hasta campeón con pocas mutaciones. Complementa, pero no
sustituye, las ligas de varias jornadas, empates y grupos. Ante animaciones o
identificadores AX caducados, observar el destino real antes de actuar; ante un
fallo de entrada, comprobar el valor visible antes de guardar.


#### Mixtos nativos: liga general, dos clasificados y final

Se preparan por API seis fixtures nuevos de cuatro equipos, para baloncesto,
balonmano y voleibol en Android e iOS. La UI selecciona Liga + eliminatorias,
una vuelta, liga general y dos clasificados. Los seis encuentros de cada liga
se introducen desde su editor nativo, sin prellenar resultados por API. Gana
siempre el equipo de menor ordinal; se incluye la victoria visitante del Equipo 2
frente al Equipo 4. No se acredita aquí el formulario de creación del torneo.

Los seis recorridos están completados: 42 resultados introducidos desde los
editores nativos, 21 por plataforma. No se prellenó ningún encuentro por API.

Las clasificaciones antes y después de la final mantienen tres partidos por
equipo y la misma puntuación: 6/5/4/3 en baloncesto, 6/4/2/0 en balonmano y
9/6/3/0 en voleibol, en ambas plataformas. La confirmación Cerrar liga y continuar explica la
congelación y los posibles desempates. En estos casos sin empate de corte se
genera una final entre Mixto Equipo 1 y Mixto Equipo 2. La fase de liga deja de
ofrecer edición. La final no se suma a la clasificación de liga.

Se guarda la final 80–70 en baloncesto, 20–19 en balonmano y 3–0 en voleibol,
con parciales 25–20, 25–21 y 25–22. Tras confirmar Finalizar torneo se muestra
Mixto Equipo 1 como campeón y la ficha finalizada sin edición de resultados.
En los tres iOS se abre también Ver clasificación final desde la celebración.
Se inspeccionan las capturas de las tablas de los tres deportes en ambas
plataformas y las celebraciones de campeón.

La lectura independiente comprueba para los seis terminados siete partidos
completed, marcadores, sets de voleibol, fases completed y lista exacta del
campeón. La API confirma además cuatro filas con tres partidos, orden 1–4,
victorias 3/2/1/0 y los puntos propios del deporte; las seis ligas de la tanda
anterior y los cinco fixtures de editores iOS siguen intactos. Una expectativa
inicial del script llamaba played al estado del partido;
se corrige a completed, el valor real del contrato. No era un fallo del producto.

En iOS se usa el setter accesible de los campos, observando sus valores antes
de guardar. Con el teclado numérico abierto, el botón puede quedar fuera del
área visible del editor de voleibol; tocar contenido no editable dentro del
popup quita el foco, oculta el teclado y devuelve Guardar. El intento de scroll
por el canal actual no mueve el editor; no se acredita scroll táctil ni se
cambia código a partir de esa limitación. Se restaura Capture Keyboard apagado.

La biblioteca iOS conserva el snapshot de 36 torneos pese a los fixtures creados
externamente mientras permanecía montada. Una recarga del cliente de desarrollo,
sin cerrar servicios ni sesión, carga los 44 y permite abrir el mixto de
baloncesto. No se presenta esa recarga como validación del gesto de refresco.

Evidencia privada: native-mixed-cycles-persisted-20261007.json,
android-mixed-*-completed-persisted-20261007.json y android-mixed-*-20261007.xml/png;
árboles y capturas iOS en la conversación. native-mixed-services-20261007.json
confirma API, proxy, preview, Metro y Mailpit con respuesta 200 y activos.
Los helpers registran acciones UI y
usan API solo para preparación de fixtures y lectura de verificación.

Alcance: se cierra este caso de liga general a una vuelta con dos clasificados
en los tres deportes y las dos plataformas. No acredita todos los grupos,
desempates, incidencias, lectores de pantalla ni release. No se modifica código
de producto ni se altera el contrato del banner global durante esta tanda.

Retrospectiva: verificar un mixto exige comparar la tabla antes y después de la
final, además del campeón. Observar los controles renderizados evita actuar
sobre una fila tapada por una cabecera fija; el XML puede conservar bounds de
texto recortado. Separar fallos del automatizador, expectativas del script y
comportamiento del producto evita proponer cambios funcionales injustificados.


#### Incidencias iOS: tenis de mesa y bádminton

En tenis de mesa al mejor de siete juegos se guarda desde iOS una no
comparecencia del local. La ficha y su editor reabierto señalan victoria
visitante; la API confirma no_show/home, 0–4 y ausencia de sets jugados.
Se cambia a abandono local con un juego completo 11–9 y un último parcial 2–4;
la ficha, reapertura y GET conservan ambos juegos, ganador visitante y 0–4
administrativo. Se restaura desde UI el abandono visitante original 5–3.
La comparación de todos los partidos con la línea base resulta idéntica.

En bádminton al mejor de tres juegos a 15 puntos, pasar de Marcador a Incidencia
mantiene Guardar deshabilitado hasta elegir el lado. Se guarda no comparecencia
visitante (2–0 administrativo), se reabre y se cambia a abandono local sin
ningún tanteo opcional (0–2). La API confirma el lado, ganador y ausencia de
sets en ambos casos; la reapertura del abandono conserva los campos vacíos.
Se vuelve a Marcador y se repone 21–20 · 15–12, con tercer juego vacío. El GET
final coincide exactamente con todos los partidos de la línea base.

La lectura final confirma también los cinco fixtures previos de editores iOS
intactos. Baloncesto se consulta únicamente: su torneo ya finalizado muestra
82–80 y no ofrece edición; no se muta ni se presenta como prueba de incidencias.

Se usan setters accesibles para los tanteos y se observan los valores antes de
guardar. El intento de scroll del editor largo de tenis de mesa no lo desplaza;
Guardar se activa mediante AX fuera del viewport. Esto acredita la acción
accesible y persistencia, no alcance táctil, teclado ni scroll manual. Algunos
IDs AX caducan tras editar; se recupera el árbol antes de actuar y se confirma
el valor real. No se modifica código de producto por esta limitación.

Evidencia privada en /private/tmp/tm-product-qa-20261004:
ios-table-tennis-incidents-{baseline,walkover,retirement,restored}-20261007.json,
ios-badminton-incidents-{baseline,no-show,retirement-empty,restored}-20261007.json
y scripts verify-ios-{table-tennis,badminton}-incidents.py. Capturas y árboles
de Device Hub en esta conversación. La expectativa inicial del verificador
llamaba walkover a la incidencia; se corrige a no_show, valor del contrato,
sin atribuir ese error del script al producto. Los servicios permanecen activos
para continuar las pruebas autorizadas.

Retrospectiva: una incidencia puede cambiar el ganador sin convertir el parcial
en resultado jugado. Comprobar tipo, lado, ganador, marcador administrativo y
parciales por separado; la restauración exige comparar el partido completo,
no solo el texto mostrado en la ficha.


#### 2026-10-08 — Incidencias iOS: balonmano y pádel

Balonmano parte de 25–25 con tanda 5–4. Se guarda no comparecencia local;
la ficha y reapertura señalan victoria visitante, y GET confirma 0–10,
no_show/home y ausencia de los dos tanteos de 7 metros anteriores.
Se cambia a abandono visitante con parcial 12–14. Al vaciar solo el tanteo
visitante y abandonar el campo, aparece el error localizado y Guardar queda
deshabilitado. Completar 14 elimina el error y permite guardar. Ficha, reapertura
y GET conservan 12–14, victoria local y 10–0 administrativo, sin tanda anterior.
Se vuelve a Marcador, se repone 25–25 y se confirma Guardar deshabilitado hasta
introducir la tanda 5–4. Se guarda y el GET compara todos los partidos con la
línea base: restauración exacta, sin incidencia residual.

Pádel parte de 6–4 · 7–6. Se guarda abandono local durante el primer set, 4–2,
con los restantes vacíos. Ficha, reapertura y API conservan ese parcial;
resultado administrativo 0–2, ganador visitante y sets jugados vacíos.
Se cambia a no comparecencia visitante: desaparecen los campos de parcial y se
guarda. La ficha ya no muestra el 4–2; GET confirma no_show/away, 2–0,
ganador local y ausencia de partialSets. La reapertura mantiene el lado y tipo.
Se vuelve a Marcador y se restaura 6–4 · 7–6, tercer set vacío. El GET final
coincide exactamente con los partidos originales y mantiene in_progress.

La verificación final acredita los cinco fixtures previos de editores iOS
intactos y el tenis de mesa restaurado. Entradas por setter AX, valores observados
antes del envío y capturas de Device Hub en esta conversación; no acredita
teclado manual, scroll táctil ni lector de pantalla. No se modifica código de
producto ni se añade cobertura de otras plataformas por estos resultados.

Evidencia privada: ios-{handball,padel}-incidents-
{baseline,no-show,retirement,restored}-20261008.json y
ios-incidents-restoration-20261008.json, en el directorio temporal de QA.

Retrospectiva: cambiar entre resultado jugado, no comparecencia y abandono debe
sustituir también los datos que ya no aplican. Revisar la ausencia de tandas y
parciales residuales en persistencia, además del ganador y del texto visible.
Un abandono puede dar la victoria al equipo que iba perdiendo; ese caso ayuda
a distinguir el marcador parcial del administrativo.


#### Cierre operativo de la tanda — 2026-10-08

Tras restaurar y verificar los fixtures, make dev-down detiene local sin borrar
volúmenes. Se suspenden los seis LaunchAgents conocidos de dev y se cierran Metro,
el proxy y la preview privados de QA. La comprobación posterior confirma cero
contenedores activos local/dev, cero tareas dev cargadas y ningún listener en
5432, 1025, 8025, 8080, 8082, 8083 ni 8084. Se conservan el volumen PostgreSQL local,
los volúmenes dev y toda la evidencia temporal. El renderer de producción en
8091 continúa escuchando; no se opera sobre producción/K3s. No se activa
observabilidad. Evidencia: qa-shutdown-20261008.json.

La matriz global continúa abierta: esta tanda amplía incidencias iOS y no cierra
el ANR Android intermitente, lectores, scroll táctil iOS, proveedores reales,
release, dispositivos físicos ni acciones que esperan intervención humana.
Solo se actualizan informe y aprendizaje; git diff --check pasa. No se repiten
builds ni suites generales por esta ampliación documental sin cambios de producto.


#### 2026-10-08 — Incidencias iOS: voleibol y baloncesto

Se reabre local sin observabilidad y se restablece el override privado de API en
8084, proxy en 8080, preview en 8082 y Metro en 8083. El arranque ordinario mapea
API a 8080; el primer intento de proxy falla por puerto ocupado. Se aplica el
override de QA ya existente, sin cambiar archivos del repositorio. Los cinco
servicios responden 200 antes de probar. Todas las inyecciones quedan en pass.

Voleibol conserva como línea base el 3–2 de cinco sets. Desde UI se guarda no
comparecencia visitante; ficha y reapertura indican ganador local, y GET confirma
3–0 con tres sets administrativos 25–0. Cambiar a abandono abre el tanteo opcional
vacío, sin prellenarlo con esos sets administrativos. Se introduce 25–20 y 12–10,
y se selecciona abandono local. Añadir un tercer set 3–2 produce dos parciales
incompletos: el árbol confirma error localizado y Guardar deshabilitado.
Vaciar el tercero corrige la validación. Se guarda y reabre: parcial
25–20 · 12–10, ganador visitante y 0–3 administrativo, tres sets 0–25 en API.
Los parciales quedan dentro de incident.partialSets, separados de sets.
Se vuelve a Marcador y se restauran los diez tanteos originales; GET confirma
igualdad exacta de todos los partidos con la línea base, resultado played 3–2,
cinco sets y torneo in_progress.

Los dos fixtures disponibles de baloncesto ya estaban finalizados. Se prepara
por API un torneo ficticio nuevo de dos equipos y eliminatoria única:
QA iOS incidencias baloncesto 20261008. Esto no acredita creación ni inicio por UI.
La recarga desde el menú de development build conserva sesión, retorna a su ruta
inicial previa y permite abrir el fixture nuevo desde Actividad reciente.
No se presenta esa recarga como pull-to-refresh de biblioteca.

Desde Añadir resultado se guarda no comparecencia local: la ficha y GET indican
victoria visitante 0–20. Se reabre y cambia a abandono visitante con parcial
35–40; el ganador administrativo es local, 20–0, aunque iba perdiendo.
La reapertura conserva tipo, lado y parcial. Se cambia a Marcador y guarda 80–70;
la API confirma played, ganador local y ausencia de incident residual.
El torneo nuevo se conserva en curso con ese resultado.

Se abre la confirmación Finalizar torneo. El intento de confirmar es rechazado
por la revisión automática: considera que el cierre irreversible que impide
corregir resultados requiere autorización específica adicional a esta tanda de
QA. No se ejecuta por API ni otro canal. Se cancela el diálogo y GET confirma
in_progress con 80–70. La finalización de este fixture queda pendiente de permiso;
no se acredita campeón ni estado completed.

Entradas por setter AX y valores observados antes de guardar. La acción Guardar
del editor largo se activa mediante AX: no acredita alcance táctil ni scroll.
No se habilitan lectores ni se cambia código de producto. Los cinco fixtures
previos de editores iOS conservan sus resultados al cierre.

Evidencia privada: verify-ios-volleyball-incidents.py,
ios-volleyball-incidents-{baseline,no-show,retirement,restored}-20261008.json,
prepare-ios-basketball-incidents.py, verify-ios-basketball-new-incidents.py e
ios-basketball-new-incidents-{baseline,no-show,retirement,played}-20261008.json,
en el directorio temporal de QA. Árboles y capturas de Device Hub en la conversación.
El archivo ios-basketball-incidents-baseline-20261008.json corresponde a la
consulta del fixture web finalizado; no se usa como línea base del torneo nuevo.

Retrospectiva: voleibol persiste sets administrativos incluso en incidencias;
comprobarlos separados del parcial real y confirmar que no reaparecen como
borrador al editar. Un torneo nuevo permite probar Añadir y después Editar sin
reabrir datos finalizados. Distinguir esa preparación por API, el recorrido UI
y el cierre irreversible que sigue pendiente de autorización.


Cierre de esta ampliación: se detienen Metro, proxy y preview; make dev-down
cierra local conservando PostgreSQL y se suspenden las tareas dev conocidas.
Verificación posterior: cero contenedores local/dev activos, cero tareas dev
cargadas y puertos de QA sin listener. El renderer de producción 8091 permanece
escuchando; no se opera sobre producción/K3s. Evidencia:
qa-volleyball-basketball-shutdown-20261008.json. git diff --check pasa; sin cambios
de producto durante esta tanda. La finalización del fixture nuevo de baloncesto
queda pendiente de autorización explícita tras el rechazo automático.


#### 2026-10-08 — Finalización autorizada del fixture iOS de baloncesto

El usuario autoriza explícitamente el cierre pendiente con «Si». Se reactiva
local para esta comprobación, con el override privado de QA y sin observabilidad.
La finalización se confirma desde la UI iOS del torneo ficticio
QA iOS incidencias baloncesto 20261008. La pantalla de celebración muestra
«¡Torneo finalizado!» e «Incidencia Local» como campeón.

GET confirma state=completed y championTeamIds exactamente igual al equipo local
ganador del partido. Todos los partidos son idénticos a la evidencia played
anterior: 80–70, sin incidencia residual. Al cerrar la celebración, la ficha
muestra Torneo finalizado y ya no ofrece Añadir, Editar ni Finalizar resultado.
Se cierra la ficha y se abre otra vez desde Actividad reciente: persisten el
estado finalizado, el ganador local y el 80–70, sin controles de edición.
Este cierre resuelve la autorización pendiente de la tanda anterior; la
preparación e inicio por API siguen sin acreditar esos recorridos por UI.

Evidencia privada: ios-basketball-new-incidents-completed-20261008.json,
contrastada con ios-basketball-new-incidents-played-20261008.json; árboles y
capturas de Device Hub en la conversación. Solo se actualizan informe y
aprendizaje, sin cambios de producto ni repetición de suites generales.

Cierre operativo verificado: Metro, proxy y preview detenidos; make dev-down
conserva el volumen PostgreSQL local; las seis tareas dev conocidas quedan
suspendidas. Cero contenedores local/dev activos y ningún listener en los puertos
5432, 1025, 8025, 8080, 8082, 8083 ni 8084. El renderer de producción en 8091
continúa escuchando; no se opera sobre producción/K3s. Evidencia:
qa-basketball-completed-shutdown-20261008.json. git diff --check pasa.

Retrospectiva: acreditar un resultado no acredita el cierre del torneo.
Contrastar campeón por identidad exacta, inmutabilidad de todos los partidos y
reapertura de la ficha finalizada. Este caso queda cerrado; la matriz global
mantiene sus pendientes de ANR Android, lectores, scroll táctil iOS, proveedores
reales, release, dispositivos físicos y demás combinaciones no acreditadas.


#### 2026-10-08 — Incidencias iOS: tenis al mejor de cinco sets

Se inicia una sesión local autorizada sin observabilidad, con el override
privado de API y proxy de QA. Se conserva la línea base de QA formulario tennis:
torneo in_progress, abandono visitante con parcial 2–3 y ganador local.

Desde UI iOS se cambia a no comparecencia local. La ficha muestra ganador
visitante y no conserva el parcial anterior. GET confirma no_show/home,
marcador administrativo 0–3, sets vacío y ausencia de partialSets. Reabrir el
editor conserva tipo y lado. Cambiar a abandono abre los campos opcionales
vacíos, sin reutilizar el parcial anterior ni el marcador administrativo.

Se introduce 6–4, 3–2 y 1–0 con abandono local. El árbol accesible confirma el
error localizado de dos sets incompletos y Guardar deshabilitado. Al vaciar el
tercer set, Guardar vuelve a estar habilitado. Se guarda 6–4 · 3–2: ficha y
reapertura conservan abandono local, parcial y ganador visitante. GET confirma
retirement/home, 0–3 administrativo, sets vacío y los dos tanteos reales en
incident.partialSets. El ganador administrativo es distinto de quien llevaba
ventaja en el tanteo parcial.

Se restaura desde UI el abandono visitante con único parcial 2–3. GET confirma
igualdad exacta de todos los partidos respecto a la línea base y torneo aún
in_progress. La ficha muestra nuevamente ganador local y parcial 2–3. No se
finaliza el torneo ni se cambia código de producto.

Canal de entrada: setters AX, con valores observados antes de guardar. Algunas
llamadas notifican un identificador caducado después de aplicar la edición;
la relectura completa confirma los valores y permite continuar. No se interpreta
ese fallo del canal como defecto del producto. Guardar se activa por AX fuera
del viewport: no acredita scroll táctil, teclado ni lector de pantalla.
No se reproduce un defecto de producto en este recorrido.

Evidencia privada: verify-ios-tennis-incidents.py e
ios-tennis-incidents-{baseline,no-show,retirement,restored}-20261008.json.
Árboles y capturas de Device Hub en la conversación. Se actualizan matriz e
informe; git diff --check pasa, sin repetir builds o suites por cambios solo
documentales.

Cierre verificado: Metro, proxy y preview detenidos; cero contenedores local/dev,
ningún listener en 5432, 1025, 8025, 8080, 8082, 8083 ni 8084 y las seis tareas dev
conocidas suspendidas. PostgreSQL conserva su volumen y evidencia. El renderer
producción 8091 sigue escuchando; no se opera sobre producción/K3s.
Evidencia: qa-tennis-shutdown-20261008.json.

Retrospectiva: para abandono en tenis, combinar un set completo y un último
parcial permite distinguir un tanteo válido de dos sets incompletos. Verificar
la sustitución del tipo de incidencia, el ganador por identidad y la restauración
exacta evita confundir marcador administrativo, tanteo real y datos residuales.
La matriz global mantiene los pendientes previamente documentados.


#### 2026-10-08 — Incidencias iOS: fútbol y sustitución de tanda

Se conserva la línea base de QA formulario football: torneo in_progress,
marcador played 1–1, tanda 4–6 y ganador visitante. Sesión local autorizada,
sin observabilidad, usando el override y proxy privados de QA.

Desde UI iOS se cambia de Marcador a Incidencia. Guardar permanece deshabilitado
hasta seleccionar el equipo afectado. Se guarda no comparecencia visitante:
la ficha muestra ganador local; GET confirma no_show/away, 3–0 administrativo,
sin homePenalties ni awayPenalties y sin parcial residual. Reabrir conserva
tipo y lado. Cambiar a abandono abre ambos tanteos opcionales vacíos.

Se selecciona abandono local. Introducir solo 2 en goles locales muestra el
error localizado y bloquea Guardar. Completar visitante con 1 recupera el envío.
Se guarda y reabre: abandono local, parcial 2–1 y victoria visitante. GET confirma
retirement/home, 0–3 administrativo, partialHomeScore=2 y partialAwayScore=1,
sin penaltis residuales. El marcador parcial no decide la victoria administrativa.

Se vuelve a Marcador. Cambiar el tanteo local a 1 produce empate 1–1 y presenta
los campos de penaltis vacíos; Guardar se bloquea hasta completar la tanda.
Se introduce 4–6 y guarda desde UI. Ficha y GET confirman el resultado original,
ganador visitante, played y ausencia de incidencia. Todos los partidos coinciden
exactamente con la línea base y el torneo sigue in_progress; no se finaliza.

Observación de claridad pendiente: el error compartido de parcial incompleto
muestra «Completa ambos tanteos o deja el parcial vacío. Solo el último set o
juego puede estar incompleto.» también en fútbol. La primera frase permite
recuperarse, pero la segunda no corresponde a este deporte. No es un fallo de
persistencia ni de validación; conviene adaptar esa segunda frase al tipo de
tanteo en una corrección posterior. No se cambia producto en esta pasada.

Canal de entrada: setters y acciones AX con relectura de valores. No acredita
teclado, lectores ni scroll táctil iOS. Evidencia privada:
verify-ios-football-incidents.py e
ios-football-incidents-{baseline,no-show,retirement,restored}-20261008.json;
árboles de Device Hub en la conversación. Matriz e informe actualizados;
git diff --check pasa, sin builds o suites generales por cambios documentales.

Cierre verificado: Metro, proxy y preview detenidos; cero contenedores local/dev,
ningún listener en 5432, 1025, 8025, 8080, 8082, 8083 ni 8084 y seis tareas dev
conocidas suspendidas. Se conserva el volumen PostgreSQL y la evidencia. El
renderer producción 8091 sigue escuchando; no se opera sobre producción/K3s.
Evidencia: qa-football-shutdown-20261008.json.

Retrospectiva: alternar tanda, no comparecencia, abandono y tanda nuevamente
comprueba la sustitución completa del resultado. Verificar ausencia de penaltis
e incidencia además del marcador y ganador. El copy compartido debe seguir
siendo comprensible en cada tipo de deporte, incluso cuando la validación sea
correcta. La matriz global conserva sus pendientes restantes.


#### 2026-10-08 — Corrección del error de parcial en deportes por goles o puntos

Se cierra la observación de texto registrada en fútbol. El editor reutiliza
isSetSport para elegir el mensaje: mantiene result_incident_partial_invalid
con la ayuda del último set/juego incompleto en deportes por sets y usa la nueva
clave result_incident_partial_score_invalid en goles/puntos. Esta última pide
completar ambos tanteos o dejar el parcial vacío, sin mencionar sets o juegos.
Los cuatro catálogos es/en/it/fr contienen la nueva clave. No cambia validación,
contrato, reglas de negocio ni transporte; no requiere una nueva decisión de
arquitectura ni implementa un ADR propuesto.

Alternativas: quitar la segunda frase para todos perdería una ayuda útil en
sets; separar un mensaje por deporte duplicaría copy innecesariamente. Se aplica
la distinción mínima ya existente entre tipos de tanteo.

Validación: pnpm run typecheck pasa; exportación web completada con 36 rutas en
web-export-partial-copy-20261008; cuatro pruebas existentes de match-incidents
pasan. ESLint del archivo de ruta y Prettier de los cinco archivos afectados
pasan. No se añaden tests que solo reproduzcan la selección de copy.

Regresión iOS con development build recargada desde su menú: retorna al fixture
QA Web refresh 20261007, con resultado 2–0. Se abre su editor, elige abandono
local y vacía solo el tanteo visitante. El árbol confirma «Completa ambos tanteos
o deja el parcial vacío.» y Guardar deshabilitado. Se cierra sin guardar;
ficha y GET conservan played 2–0 sin incidencia. Esta pasada acredita el texto
español en iOS; no se afirma inspección visual nueva de otros idiomas o SO.
Los textos de sets permanecen sin cambios y su validación conserva las pruebas.
Evidencia: ios-partial-copy-unchanged-20261008.json y árboles de Device Hub.

Checklist de apps/client/AGENTS.md aplicada al cambio: texto exclusivamente en
catálogos planos, cuatro locales completos, selección con semántica existente,
sin nuevos literales visibles ni cambios de tokens, layout, cierre, navegación,
accesibilidad, providers, dependencias o permisos. No se toca una operación
OpenAPI: se conserva la adaptación generada y apiFetch existentes. Typecheck y
exportación web ejecutados; documentación y aprendizaje actualizados. Los
pendientes globales previamente documentados siguen abiertos.

Cierre verificado: Metro, proxy y preview detenidos; cero contenedores local/dev,
cero listeners de QA y seis tareas dev conocidas suspendidas. Se conserva el
volumen PostgreSQL y la evidencia. Producción 8091 permanece escuchando, sin
operar sobre producción/K3s ni activar observabilidad. Evidencia:
qa-partial-copy-shutdown-20261008.json. git diff --check pasa.

Retrospectiva: el mensaje de recuperación debe describir los controles que ve la
persona. Reutilizar la clasificación de tanteo del editor evita introducir
reglas nuevas por deporte y mantiene la ayuda específica donde resulta útil.


#### 2026-10-08 — Regresión automatizada y acceso al emulador Android

Se prepara la ampliación de incidencias Android de bádminton, guardando por GET
la línea base de QA formulario badminton: in_progress, played 2–0 y juegos
21–20, 15–12. No se escribe ningún resultado en esta tanda.

El inventario UI no muestra el emulador como aplicación accesible. Un intento de
arrancar el AVD Pixel_API_34 termina porque ya hay una instancia activa; no se
usa read-only ni se crea otra. La inspección de procesos en lectura y Device
Manager de Android Studio confirman el emulador existente en ejecución.
El canal UI no acepta su ejecutable como aplicación controlable. Se consulta el
administrador y se cierra su menú sin cambiar, borrar, duplicar ni reiniciar el
dispositivo. Se solicita cerrar Android Studio tras el diagnóstico. No se accede
a la pantalla del producto, por lo que no se acredita no comparecencia ni abandono
vacío nuevos en Android, ni se interpreta esto como fallo del producto o ANR.
La prueba visual permanece pendiente de acceso al emulador.

Regresión ejecutada sobre el estado actual: node --test tests/*.test.mjs completa
78 pruebas aprobadas, cero fallos, omitidas o canceladas. pnpm run lint pasa sin
warnings. Los avisos de Node sobre inferencia de tipo de módulo no son fallos de
las pruebas ni motivan cambiar la configuración de paquetes durante QA.

Se revisan los cuatro catálogos es/en/fr/it: 527 claves por idioma, conjuntos de
claves idénticos, cero duplicados, valores no vacíos y marcadores entre llaves
coincidentes con inglés. Evidencia: locale-consistency-20261008.json. Esta
comprobación estructural no acredita calidad lingüística, layout ni ejecución
visual en los cuatro idiomas. No se añaden tests redundantes ni cambia producto.

Evidencia privada de preparación:
android-badminton-incidents-baseline-20261008.json; Device Manager y resultados
de herramientas en la conversación. El fixture permanece sin escrituras de esta
tanda. Se actualizan informe y aprendizaje; git diff --check pasa.

Cierre verificado: Metro, proxy y preview detenidos; cero contenedores local/dev,
cero listeners de QA y las seis tareas dev conocidas suspendidas. Se conserva el
volumen PostgreSQL y la evidencia. El renderer de producción en 8091 continúa
escuchando; sin operar sobre producción/K3s ni activar observabilidad.
Evidencia: qa-automated-shutdown-20261008.json. La instancia Android preexistente
no se detiene ni se altera su almacenamiento.

Retrospectiva: un proceso o dispositivo activo no garantiza acceso a su ventana
mediante el canal de automatización disponible. Registrar la limitación y no
convertir preparación por API o una suite verde en cobertura visual. La igualdad
de claves y marcadores localizados ofrece una regresión útil y barata después
de ampliar un catálogo, pero no sustituye la revisión lingüística y visual.


#### 2026-10-08 — Detector de carreras, seguridad operativa y contrato

Se amplía la evidencia de backend sin arrancar API, PostgreSQL, Metro ni
observabilidad. make test-race pasa, pero incluye resultados de caché; para
obtener evidencia de esta tanda se ejecuta go test -race -count=1 -json ./...
desde apps/backend y se conserva la salida completa.

El primer intento sin caché bajo sandbox termina con fallos al abrir listeners
efímeros localhost: bind: operation not permitted en pruebas de shutdown,
proveedores con httptest y renderer. Se conserva esa salida como intento
restringido, sin atribuirla a un defecto funcional. Se repite la misma ejecución
con acceso a puertos locales: exit 0, 483 tests/subtests aprobados, 57 omitidos,
18 paquetes aprobados y cuatro sin pruebas. Cero informes WARNING: DATA RACE y
cero resultados de caché. Las integraciones PostgreSQL opt-in no están activadas;
esta pasada no renueva su evidencia previa ni acredita ausencia universal de
carreras fuera de los recorridos ejecutados.

python3 tests/operational-safety.test.py completa 11 pruebas con OK. Usa dobles
de Docker y temporales para comprobar límites de limpieza, aislamiento de
perfiles y desarrollo bajo petición; no borra volúmenes ni opera contra K3s.
pnpm run openapi:lint pasa; Redocly indica un problema explícitamente ignorado
por la configuración existente. No se modifica contrato, generación ni reglas
para ocultar ese aviso. No se amplía cobertura visual con estas comprobaciones.

Evidencia privada: backend-race-20261008.jsonl (intento restringido),
backend-race-host-20261008.jsonl (ejecución válida sin caché),
backend-race-summary-20261008.json y qa-race-services-off-20261008.json.
Se actualizan informe y aprendizaje; git diff --check pasa. Sin cambios de
producto en esta tanda.

Verificación final: cero contenedores local/dev activos, cero listeners en
5432, 1025, 8025, 8080, 8082, 8083 ni 8084, seis tareas dev conocidas suspendidas
y volumen PostgreSQL local conservado. Renderer producción 8091 permanece
escuchando; no se opera sobre producción/K3s ni sobre el emulador preexistente.

Retrospectiva: una suite con -race puede reutilizar caché. Para una nueva tanda
de QA, -count=1 y la salida JSON permiten distinguir ejecución fresca, omitidos
y resultados reales. Un bind rechazado por sandbox requiere resolver el permiso
del entorno, no cambiar la implementación ni declarar un fallo de concurrencia.
Los pendientes visuales y de intervención humana de la matriz siguen abiertos.


#### 2026-10-08 — Incidencias Android mediante ADB autorizado

El usuario autoriza explícitamente ADB para leer, capturar y operar Pixel_API_34.
Se reutiliza la instancia existente y se reactiva únicamente el entorno local
de QA y Metro, sin observabilidad. Esto resuelve el acceso pendiente de la tanda
anterior. Recargar y abrir los tres enlaces con la app activa funciona en esta
sesión; no cierra el diagnóstico histórico de ANR o recreación del launcher.

En QA formulario badminton se convierte el resultado jugado en incomparecencia
del visitante: victoria local 2–0, sin sets. Reabrir conserva tipo y lado. Se
cambia a abandono local, con los seis campos del parcial vacíos, y se guarda:
victoria visitante 0–2, sin sets ni partialSets. La reapertura conserva los campos
vacíos. Se recupera el marcador 21–20, 15–12, con los campos extra vacíos.

En QA formulario padel se guarda incomparecencia local: victoria visitante 0–2,
sin sets ni parcial. Se cambia a abandono visitante con parcial 6–4, 2–3:
victoria local administrativa 2–0, sets vacíos y partialSets exactos. La ficha y
el editor reabierto muestran el parcial. Cambiar a marcador y completar 6–4, 7–6
restaura el resultado original. Se desplaza el modal para alcanzar Guardar.

En QA formulario football se parte de 1–1 con penaltis 4–6. Incomparecencia
visitante produce 3–0; abandono local con parcial 2–1 produce 0–3. Ambos eliminan
homePenalties y awayPenalties. Reabrir el abandono conserva 2–1. Un borrador con
solo un tanteo muestra «Completa ambos tanteos o deja el parcial vacío.», sin
referencia a sets/juegos; la captura confirma Guardar visualmente desactivado.
Volver a marcador 1–1 exige una tanda nueva; introducir 4–6 restaura el original
y elimina la incidencia. Se comprueba el nuevo copy español también en Android.

Los verificadores GET comprueban cada escritura y la igualdad exacta de todos
los partidos con su línea base al restaurar. Los tres torneos siguen in_progress;
no se finalizan. Se revisa visualmente la captura del error de parcial. No se
modifica producto en esta tanda ni se repiten las suites ya aprobadas sin cambios
nuevos. La matriz mantiene sus pendientes de lectores de pantalla, dispositivos
físicos, arranque/recreación Android y demás combinaciones no acreditadas.

Evidencia privada en /private/tmp/tm-product-qa-20261004: familias
android-badminton-*, android-padel-* y android-football-*, con sufijo 20261008
(XML, PNG, líneas base y respuestas GET); verificadores
verify-android-{badminton,padel,football}-incidents.py. Las capturas pueden
contener datos de las cuentas ficticias y no se incorporan al repositorio.

Retrospectiva: cambiar de resultado jugado a incidencia y volver requiere
comprobar la ausencia de campos incompatibles en la respuesta, además de mirar
el ganador. Comparar todos los partidos contra una línea base detecta pérdidas
de datos fuera del marcador editado. En UIAutomator el texto hijo puede figurar
enabled aunque el botón esté desactivado; contrastar contenedor y captura antes
de interpretar ese atributo como un fallo funcional.

Cierre verificado: Metro, proxy y preview detenidos, cero contenedores local/dev,
cero listeners de QA y seis tareas dev conocidas suspendidas. Volumen PostgreSQL
y evidencia conservados. El renderer de producción en 8091 sigue escuchando;
no se opera sobre producción/K3s ni se detiene el emulador preexistente. Evidencia:
qa-android-incidents-shutdown-20261008.json. git diff --check pasa.


#### 2026-10-08 — Tenis de mesa Android y copy en tres idiomas adicionales

Se reutiliza Pixel_API_34 mediante ADB autorizado, reactivando API/PostgreSQL/
Mailpit, proxy y Metro para esta sesión, sin observabilidad. El enlace de tenis
de mesa con force-stop abre DevLauncherActivity, Status ok, cero FATAL EXCEPTION
y cero guard de contexto en los logs de ese intento. La pantalla conserva un
aviso histórico de caída: no demuestra una nueva excepción. Reconectar la URL
exacta de Metro recupera Inicio autenticado; el enlace VIEW/BROWSABLE con la app
cargada abre el torneo, sin nuevo fatal ni guard. No se repite la provocación
conocida de recreación durante el arranque ni se declara corregida; arranque en
release y dispositivos físicos continúan pendientes.

QA formulario table_tennis parte de abandono visitante con partialSets=[5–3],
agregado 4–0, al mejor de siete juegos. Desde el editor se cambia a
incomparecencia local y se guarda. La ficha muestra victoria visitante; GET
confirma no_show/home, 0–4, sets vacío e incidencia sin partialSets. Reabrir
conserva tipo y lado. Se vuelve a abandono visitante y se introducen 5–3, con
los otros doce campos vacíos. Desplazar el modal permite alcanzar Guardar.
La respuesta posterior coincide exactamente con todos los partidos de la línea
base y conserva in_progress; no se finaliza el torneo.

Se amplía la revisión del copy result_incident_partial_score_invalid a inglés,
italiano y francés usando la preferencia de idioma por app de Android. Cada
cambio se hace con la app ya cargada y la recreación vuelve a Inicio conservando
la sesión. En fútbol, desde 1–1 con penaltis 4–6, se prepara un borrador de
abandono local con tanteo 2 y visitante vacío, sin guardar. El mensaje coincide
con el catálogo de cada idioma; las tres capturas revisadas muestran texto
completo en dos líneas y Guardar desactivado, sin referencia a sets/juegos. Se
descarta cada borrador por backdrop. GET confirma todos los partidos idénticos
a la línea base, incluido 1–1 y penaltis 4–6. Con español acreditado en la tanda
anterior queda comprobado este mensaje concreto en los cuatro idiomas Android;
no se extrapola a todas las rutas o traducciones ni a lectores de pantalla.

La lista inicial de idiomas por app era vacía y se restaura a vacía, recuperando
español del sistema; font_scale permanece 1.0. No se cambia código de producto.
Se actualizan matriz y aprendizaje; no se repiten suites históricas sin cambios.

Evidencia privada: android-link-cold-current-20261008*,
android-startup-reconnected-20261008*, android-table-tennis-*-20261008*,
android-football-{en,it,fr}-*-20261008*, android-locale-*-20261008*,
android-partial-copy-locales-20261008.json y android-locales-runtime-20261008.log.
El log desde el intento caliente hasta restaurar idiomas contiene cero fatals
y cero guard de contexto; esto no reproduce el solapamiento con arranque.

Cierre verificado: cero contenedores local/dev, cero listeners de QA y seis
tareas dev conocidas suspendidas. Metro, proxy y preview detenidos; volumen
PostgreSQL y evidencia conservados. Renderer producción 8091 sigue escuchando;
no se opera sobre producción/K3s ni se detiene el emulador preexistente. Evidencia:
qa-android-table-tennis-locales-shutdown-20261008.json. git diff --check pasa.

Retrospectiva: la traducción de un error nuevo merece una prueba pequeña por
idioma en el formulario real, además de comprobar la igualdad de claves. Cambiar
la preferencia del SO permite validar la localización sin añadir un selector
al producto. Un aviso persistido por DevLauncher y un fatal nuevo requieren
evidencias distintas; una reapertura correcta no cierra el fallo histórico.


#### 2026-10-08 — Verificación del bloque acumulado antes de subir develop

A petición del usuario se revisa el conjunto de cambios acumulados y el remoto.
develop coincide con origin/develop antes del commit; ADR-0150 está aceptado.
Se excluye apps/client/modules/*/android/build/ para versionar únicamente fuente
del módulo, no cachés, informes ni binarios Gradle. Se actualiza Unreleased del
changelog. Los 28 archivos candidatos no incluyen configuración privada,
capturas, APK ni artefactos de build; el barrido de patrones comunes no encuentra
claves privadas, tokens GitHub, claves AWS ni JWT. No equivale a una garantía
universal de ausencia de secretos. Los fixtures y credenciales siguen ignorados.

node --test tests/*.test.mjs pasa 78/78. pnpm run check pasa. make verify completo
termina con exit 0: formato, lint, TypeScript, contrato, tests, exportación web
de 36 rutas, generación OpenAPI/sqlc sin diff, tidy de ambos grafos y build Go.
La integración PostgreSQL opt-in se omite por no haber URL de BD aislada; no se
apunta a fixtures persistentes. govulncheck informa cero vulnerabilidades
alcanzables y una en módulos requeridos sin llamadas afectadas; no se declara
una auditoría universal a cero. Logs privados pre-push-check-20261008.log,
pre-push-tests-20261008.log y pre-push-verify-20261008.log.

Checklist cliente contra el conjunto: localización es/en/it/fr y tokens
compartidos; host Android aislado con stub iOS/web, provider dueño del plazo y
anclajes sin reglas de negocio; transportes por adaptadores generados y apiFetch,
sin nuevas operaciones ni DTO manual; recuperación segura y control de envíos
duplicados. Se mantiene el inventario de pendientes: lectores de pantalla,
Android Lint bloqueado por excepción de herramienta en dependencia, recreación
durante arranque de DevLauncher y distribución/dispositivos reales. Subir a
develop no acredita cierre de estos casos ni publicación de producción.

No se arranca local/dev, observabilidad ni tareas durante esta verificación.
Retrospectiva: cerrar bloques validados con código, ADR y evidencia juntos reduce
el riesgo de perder contexto; excluir outputs de módulos locales es necesario
aunque los directorios nativos generados principales ya estén ignorados.


#### 2026-10-08 — Regresión de renovación web incorporada a CI

Tras subir 14d3edd a origin/develop, se revisa el workflow Verify: ejecuta
make verify con PostgreSQL efímero para las integraciones. La nueva suite
session-refresh.test.mjs estaba acreditada por la pasada local de 78 pruebas,
pero no era prerrequisito de verify. Se añade test-session-refresh al gate
existente. Sus seis pruebas pasan: red y reintento, 429, 500, cuerpo inválido,
401 con invalidación única y lecturas concurrentes con una renovación.
La comprobación en seco confirma que el gate incluye la suite.

Retrospectiva: tener un test versionado y aprobado no asegura que CI lo ejecute.
Mantener la nueva regresión en la entrada compartida evita perder esa protección
en futuras subidas; no requiere otro workflow ni una herramienta nueva.


#### 2026-10-08 — Android autenticado con texto al 200 %

Se arranca Pixel_API_34 para esta tanda; no había dispositivo ADB activo.
API/PostgreSQL/Mailpit locales y Metro se encienden sin observabilidad. La build
instalada es com.fasttourney.app.local: Metro se ejecuta con APP_ENV=local y los
reverses 8080/8083 de esta sesión, sin reconstruir ni cambiar dependencias.

Con la app ya cargada se cambia font_scale de 1.0 a 2.0. La recreación conserva
la sesión y vuelve a Inicio. Se recorre Actividad reciente, QA formulario football,
la ficha y Editar resultado. El título de la ficha usa dos líneas; las etiquetas
largas del formulario se ajustan, y el scroll permite alcanzar los campos y el
botón inferior. El alcance es español, tema claro y emulador Android API 34.

En el editor se elige Incidencia, Abandono y Equipo local. Se conserva el tanteo
local 1 y se vacía el visitante. Tras ocultar el teclado y desplazar el diálogo,
«Completa ambos tanteos o deja el parcial vacío.» se lee completo en dos líneas;
Guardar resultado queda desactivado tanto en el contenedor del árbol como en la
captura revisada. No se guarda: se descarta por backdrop. GET confirma igualdad
exacta de todos los partidos y estado con la línea base previa. El log de esta
sesión del emulador registra cero FATAL EXCEPTION y cero guard de contexto; no
reproduce ni cierra el fallo de solapamiento durante arranque de DevLauncher.

Se restaura font_scale=1.0 y se confirma Inicio autenticado. Se detienen Metro,
los tres servicios locales y el emulador que se arrancó para esta prueba;
se retiran únicamente los reverses ADB de la sesión. Volúmenes y evidencia
permanecen. La suspensión productiva de esta fecha responde a una autorización
explícita separada, documentada en DEPLOYMENT.md; no es parte del QA local.

Evidencia privada: android-large-font-*-20261008.{png,xml,json,log},
qa-large-font-{runtime,metro,shutdown}-20261008.log y
qa-large-font-final-state-20261008.json, bajo /private/tmp/tm-product-qa-20261004.
La evidencia no acredita lectores de pantalla, todas las rutas/idiomas, otros
SO, release ni dispositivos físicos. No se modifica código de producto.

Retrospectiva: un árbol puede incluir controles fuera del viewport con bounds
invertidos; desplazar y revisar la captura evita confundir contenido todavía no
visible con recorte definitivo. Probar el error y su botón juntos, restaurar la
preferencia del SO y comparar datos después acredita una tanda reproducible.


#### 2026-10-08 — Biblioteca y Cuenta Android con texto al 200 %

Se continúa la sesión autorizada con Pixel_API_34, API/PostgreSQL/Mailpit locales
y Metro en 8083, APP_ENV=local; sin observabilidad. Se arranca el emulador para
esta tanda y se conecta la build instalada por ADB, sin reinstalar ni actualizar
dependencias. El arranque carga Inicio autenticado. font_scale pasa de 1.0 a
2.0; la recreación conserva la sesión. Alcance: español, tema claro, API 34,
orientación vertical y development build.

Biblioteca muestra Administro 45 y Sigo 5. Los títulos largos se ajustan a
varias líneas. Se cambia a Sigo y se desplaza hasta el quinto torneo, visible
completo por encima de la botonera; el botón flotante permite abrir Crear torneo.
En creación se ven los ocho deportes y los campos; el scroll alcanza el equipo
precompletado y Crear torneo. Se cierra sin introducir nombre ni enviar.
No se acredita en esta tanda el envío, la validación ni la paginación de Administro.

Cuenta conserva el usuario ficticio y permite abrir Datos de acceso y Cambiar
contraseña. El correo y la explicación se ajustan a varias líneas. El formulario
vacío muestra ambas etiquetas, los campos y Guardar contraseña deshabilitado;
enfocar Nueva contraseña mantiene el formulario y su acción visibles con teclado.
No se introduce ni cambia contraseña, no se vincula proveedor, no se cierra sesión
ni se solicita baja.

Se detecta la incidencia 13 de cabecera: Cambiar contraseña queda pegado al botón
Volver al 200 %, tanto con teclado como sin él. El título y el control comparten
x=168 como límite en el árbol; la captura confirma falta de separación. No se
presenta este recorrido como aprobado íntegramente. La inspección del layout
nativo de Cuenta muestra el título estándar centrado y un headerLeft propio.
Recomendación: adaptar la cabecera al espacio disponible reutilizando el patrón
compartido y conservando el escalado; reducir globalmente el texto penalizaría
la accesibilidad. Falta implementar y verificar la corrección, incluidas las
otras cabeceras largas. No se toma una decisión nueva de producto.

Se restaura font_scale=1.0 y se confirma Inicio autenticado. Se detienen Metro,
el emulador arrancado aquí y los servicios locales; se retiran los reverses
8080/8083 de esta sesión. La comprobación final encuentra cero contenedores,
cero listeners de QA y las seis tareas dev conocidas suspendidas; el volumen
PostgreSQL local permanece. No se opera sobre producción/K3s ni se altera su
suspensión autorizada preexistente.

Evidencia privada bajo /private/tmp/tm-product-qa-20261004:
android-large-navigation-*-20261008.{png,xml,log},
qa-large-font-navigation-{emulator,runtime}-20261008.log,
qa-large-navigation-shutdown-20261008.log y
qa-large-navigation-final-state-20261008.json. Se revisan las capturas de
Biblioteca, final de Sigo, final de creación, Cuenta, Datos de acceso y contraseña
con/sin teclado. El log del emulador de esta tanda contiene cero FATAL EXCEPTION;
no reproduce ni cierra el fallo histórico de DevLauncher. No se modifica código
de producto ni se repiten suites aprobadas sin cambios. git diff --check pasa.

Retrospectiva: los campos pueden adaptarse correctamente al texto ampliado
mientras la cabecera conserva una restricción distinta. Revisar ambos evita
acreditar una pantalla entera solo por su contenido. Esperar y volver a leer
el árbol tras una recreación o transición evita interpretar un toque prematuro
como un fallo de navegación. Esta evidencia no sustituye TalkBack, otros idiomas,
release ni dispositivos físicos.


#### 2026-10-08 — Corrección de cabecera de Cuenta y etiqueta multilínea

Se reproduce de nuevo la incidencia 13 antes de editar. El layout de Cuenta
Android conserva NavigationHeaderButton y reemplaza solo su título estándar por
Text compartido, semántica header y dos líneas. El máximo se calcula con el ancho
actual menos el control de 44 dp, margen exterior y separación de 20 dp a cada
lado, siguiendo el patrón existente en web. No se fija un ancho de pantalla ni
se desactiva el escalado. iOS conserva su configuración nativa y web su layout
específico; no se modifica la navegación, autorización ni operaciones OpenAPI.

Al 200 %, Cambiar contraseña en español deja 53 px frente a Volver y se muestra
completo en dos líneas, también con Nueva contraseña enfocada y teclado visible.
Datos de acceso queda igualmente completo; Volver permite regresar a Cuenta.
En inglés, italiano y francés se recorren Cuenta → Datos de acceso → Cambiar
contraseña; las separaciones del editor son 53 px en los tres casos y las
capturas confirman título íntegro, sin solapamiento. Ajustes y Notificaciones
se revisan en español al 200 % y no muestran el defecto de separación.

En francés se detecta y corrige la incidencia 14 añadiendo textAlign center a la
etiqueta de Button compartido. La segunda captura confirma Enregistrer le mot de
passe en dos líneas centradas, con Guardar deshabilitado en formulario vacío.
No se introduce ni cambia contraseña, no se guarda ningún formulario ni se
modifican datos de torneos. El truncado de la etiqueta nativa inglesa de Torneos
al 200 % permanece como observación de alcance; no se acredita su locución.

pnpm run typecheck, pnpm run check (formato, lint, tipos y OpenAPI) y exportación
web de 36 rutas pasan con ambos cambios. Expo regenera expo-env.d.ts sin salto
final: se restaura únicamente ese salto antes del gate, sin incluir un cambio
funcional generado. No se añaden tests que reproduzcan estilos; la regresión
se contrasta con el árbol y las capturas reales del caso reproducido.

Checklist cliente: catálogos existentes es/en/it/fr; tokens, Text y Button
compartidos; cierre circular existente; reserva dinámica y dos líneas sin
reducir texto; rutas y transporte intactos; no se tocan endpoints ni se añaden
reglas de negocio. Siguen abiertos lectores, otras rutas/versiones de SO,
dispositivos físicos, distribución y proveedores reales aplazados.

Evidencia privada bajo /private/tmp/tm-product-qa-20261004:
android-header-*-20261008.{png,xml,json}, qa-header-locales.py,
qa-header-locales-20261008.log y qa-header-{button-check,button-web-export,
typecheck}-20261008.log. La corrección se prueba sin editar JavaScript durante
las secuencias de navegación.

Retrospectiva: el título estándar de una toolbar y un control propio pueden
medirse con límites diferentes. Reutilizar el cálculo de espacio aceptado es
suficiente; no hace falta cambiar el router ni reducir la preferencia de texto.
Revisar el idioma más largo descubre también etiquetas multilínea cuyo bloque
está centrado pero cuyas líneas no lo están.


#### 2026-10-08 — Intento de TalkBack, sin acreditar recorrido completo

TalkBack 14.2.0.618048417 está instalado en Pixel_API_34. Se guardan los valores
previos enabled_accessibility_services=null, accessibility_enabled=0 y
touch_exploration_enabled=0. Tras activar el servicio y resolver su solicitud
inicial de notificaciones, dumpsys acredita servicio enlazado y exploración
táctil activa. La notificación se rechaza; no se habilitan permisos del producto.

El primer intento combina UIAutomator events/dump con el lector. La inspección
altera el estado de exploración: no se usa como evidencia de navegación TalkBack.
Se detiene el proceso UIAutomator y se contrasta servicio activo con capturas.
Los intentos con gestos ADB y atajos de teclado no acreditan un recorrido fiable
del foco dentro de la app; el borde de foco permanece en Tools de la development
build aunque otros eventos naveguen. No se equipara navegar mediante una
inyección a usar correctamente el lector ni se acredita calidad de locución.
Los [atajos oficiales](https://support.google.com/accessibility/android/answer/6110948?hl=en)
son referencia, no prueba de ejecución satisfactoria en este emulador.

Durante la reconexión y el lanzamiento explícito de MainActivity se registran
dos FATAL EXCEPTION de DevLauncher con el guard App react context shouldn't be
created before. Reconectar una vez la URL exacta de Metro recupera Inicio
autenticado. Es evidencia nueva del fallo ya abierto de development build;
no se atribuye al formulario ni se declara corregido. Los logs y capturas
android-talkback-*-20261008 se conservan privados. Se restauran los tres valores
seguros previos de accesibilidad y el estado sin decisión del permiso de
notificaciones de TalkBack. El caso de lector permanece abierto.

Retrospectiva: validar un lector exige comprobar que el instrumento no lo
suspende y que las acciones realmente recorren su foco. Un servicio activo,
un árbol completo o una pulsación que navega no bastan por separado.


#### 2026-10-08 — APK release local, bundle incluido y recuperación de conexión

Se compila assembleRelease para ARM64 con el JDK incluido en Android Studio,
APP_ENV=local y API loopback. La APK com.fasttourney.app.local no es debuggable y
comparte el certificado de QA con la development build: adb install -r conserva
sus datos. Se guarda antes una copia de la APK debug. Es una prueba local,
firmada con la clave de desarrollo; no es un artefacto de tienda ni una
publicación de producción. La fuente es ea57150 más las dos correcciones de
cabecera y Button descritas arriba, todavía sin commit al compilar.

La primera APK, con política de red por defecto, abre Inicio en tres arranques
pero no conecta con la API HTTP local. Para aislar bundle y navegación, se
crea un overlay release temporal en el Android generado e ignorado: base HTTP
denegada y excepción únicamente para 127.0.0.1, sin subdominios. El primer
intento del overlay falla lintVitalRelease por omitir includeSubdomains; se
corrige explícitamente a false y la compilación pasa sin omitir tareas. El
overlay se conserva privado y se retira del árbol de trabajo. No cambia
app.config.ts, la política del producto ni la infraestructura pública.

Con Metro apagado y esa APK de QA se comprueba:

- Tres arranques fríos recuperan Inicio autenticado y actividad reciente.
  am start informa 5560, 6652 y 6748 ms; son medidas de este emulador, no un
  presupuesto de rendimiento. El log acotado registra cero FATAL EXCEPTION.
- El enlace fasttourney-local de QA formulario football abre la ficha con la
  app abierta y cerrada. La captura conserva 1–1, penaltis 4–6 y ganador visitante.
- Al detener solo la API local y abrir el enlace en frío, RequestErrorCard
  muestra common_network_error y Reintentar, sin cuerpo interno ni aviso
  duplicado. Tras restaurar la API, Reintentar recupera la ficha. La proyección
  pública completa coincide con la lectura previa, sin cambios en partidos.
- Un nuevo arranque con API disponible recupera la sesión. Cuenta → Datos de
  acceso → Cambiar contraseña al 200 % mantiene el título completo y 53 px
  físicos frente a Volver. El formulario permanece vacío y no se envía.

La compilación normal y la variante temporal pasan las tareas vitales de lint
release. Esto no cierra el gate anterior de lint Android completo ni acredita
iOS release, otras arquitecturas, firma de distribución, App Links verificados
por el SO, OAuth real, dispositivos físicos o todas las rutas de la matriz.
La caída de DevLauncher permanece abierta: no se reproduce en estos arranques
release, pero esa muestra no corrige ni invalida la evidencia de development.

Se reintenta TalkBack en la APK release sin Tools ni UIAutomator. El servicio
muestra foco en el título de Inicio, pero los siguientes gestos ADB no acreditan
avances sucesivos fiables. Se conservan las capturas, se restauran los valores
previos y no se declara aprobado el lector ni su locución.

Evidencia privada: android-local-release-{build,loopback-rebuild}-20261008.log,
android-local-release-{default-policy,loopback}-20261008.apk,
android-local-release-artifacts-20261008.json, android-release-*-20261008,
qa-release-{startup-loopback,recovery}-20261008.log y scripts asociados.
SHA256 de la APK temporal: 67cc5cee4501fdd4251834e9c3bd2d0949f3e205246066d2b90632b7f96255f2.
Producción/K3s permanece apagada y no se activa observabilidad.

Retrospectiva: un bundle release puede arrancar correctamente y fallar por la
política de transporte del entorno local. Separar ambos hechos permite probar
navegación y recuperación con una excepción estrecha, sin debilitar la
configuración distribuible ni presentar un binario de QA como producción.


#### 2026-10-08 — TalkBack con entrada táctil del emulador y enlaces negativos

Se resuelve el bloqueo instrumental de la muestra anterior con `adb emu event
mouse`: los eventos atraviesan el dispositivo táctil del emulador, acreditado
con getevent. No se ejecuta UIAutomator mientras TalkBack está activo. Capturas
sucesivas muestran avance real del borde de foco y la doble pulsación activa
el elemento previamente verificado. Esto amplía la evidencia anterior; no
convierte sus intentos fallidos en pruebas aprobadas.

- Inicio: foco en descripción, Crear torneo y Actividad reciente. Crear abre
  el formulario; el recorrido alcanza deportes y el campo Nombre vacío. La
  activación abre el teclado, sin escribir ni crear un torneo.
- Cuenta: Cerrar sesión abre su diálogo. El recorrido alcanza título, cuerpo,
  confirmación y Cancelar; activar Cancelar conserva la cuenta autenticada.
- Datos de acceso → Cambiar contraseña: el foco alcanza cabecera y campo
  actual. Al 200 %, alcanza además el campo nuevo y Guardar deshabilitado;
  la cabecera conserva sus dos líneas y separación. No se introducen ni
  cambian credenciales.
- Enlaces a torneo: UUID inexistente devuelve 404 y muestra «Este torneo ya
  no está disponible» con Cerrar. Identificador malformado devuelve 400 y
  muestra common_request_error con Reintentar, sin exponer el cuerpo interno.

La muestra acredita foco y activación, no calidad de locución, todas las rutas
ni dispositivos físicos. Un primer toque no adquirió el foco esperado y una
URL enviada durante recreación por font_scale terminó en Inicio: se descartan
esas capturas como prueba de contraseña y se repite después de verificar foco
y pantalla estables. No se atribuyen estos fallos del harness al producto.

Al cerrar Android se restauran exactamente los tres valores de accesibilidad,
font_scale=1 y locales vacíos; se reinstala la APK debug guardada con install -r,
se retiran los reverses propios y se apaga el emulador iniciado para esta tanda.
La proyección completa del torneo sigue idéntica. El log release desde las
20:43 registra cero FATAL nuevos y cero ANR observados: no cierra incidencias
históricas ni el fallo de DevLauncher. Metro queda apagado; la API local se
conserva únicamente para continuar el QA iOS. Producción continúa apagada.

Evidencia privada: android-release-negative-links-20261008.json,
android-talkback-hardware-*-20261008, android-release-cleanup-20261008.json,
android-release-complete-logcat-20261008.log y scripts qa-talkback-hardware y
qa-android-release-cleanup asociados. El directorio privado tiene modo 700.

Retrospectiva: verificar el foco antes de activar evita confundir navegación
por coordenadas con uso del lector. Comprobar respuestas HTTP junto al mensaje
visible distingue el rechazo de negocio recuperable del fallback común seguro.


#### 2026-10-08 — iOS release de simulador y VoiceOver acotado

Xcode 27 compila FastTourneyLocal Release para iPhone 17, iOS 27.0, ARM64,
con bundle incluido y Metro apagado. Se interrumpe deliberadamente el primer
build de dos arquitecturas y se usa ONLY_ACTIVE_ARCH=YES. La build sin firma
compila y abre Inicio, pero la ficha muestra common_request_error sin petición
correlacionada. Recompilar con CODE_SIGNING_ALLOWED=YES y CODE_SIGN_IDENTITY=-
genera entitlements de simulador y permite cargar la ficha. Es una limitación
del harness sin firma; no se demuestra una causa interna concreta de SecureStore.
No se añaden perfiles ni excepciones de transporte: ATS conserva
NSAllowsArbitraryLoads=false y NSAllowsLocalNetworking=true. No había app
FastTourney instalada en este simulador. Es firma ad hoc, no distribución.
Fuente 80a0d41, seguida solo por documentación. SHA256 de main.jsbundle:
3f2be6d7f6a588a88c5c9e19972cc9e6b2734e02124b6ff139054345f5d51e10.

Con la build firmada se comprueba visualmente:

- Tres arranques fríos llegan a Inicio anónimo sin Metro. No acredita sesión
  persistida en esta instalación nueva ni rendimiento de dispositivo físico.
- Enlace tournament con app abierta y cerrada: QA formulario football, 1–1,
  penaltis 4–6 y ganador visitante. El primer comando usó una ruta plural
  incorrecta; se descarta como prueba del enlace válido.
- UUID inexistente: HTTP 404 real, torneo no disponible y Cerrar. Identificador
  malformado: HTTP 400 real, common_request_error y Reintentar. Logs de API
  corroboran ambos estados; no se muestran cuerpos internos.
- API local detenida: common_network_error y Reintentar, sin aviso duplicado.
  Tras restaurar HTTP 200, el toque recupera la ficha. Las proyecciones públicas
  completas anterior y posterior coinciden byte a byte.
- Text Size 11 mantiene separados los controles, pero recorta verticalmente
  el título en la toolbar nativa. Reproduce la limitación abierta a Text Size 7,
  sin acreditar lectura completa. Text Size vuelve a 3.

Antes de recompilar con firma, VoiceOver muestra avance real desde el título
de Inicio a descripción y Crear torneo. La doble pulsación abre el formulario
y el siguiente gesto alcanza su cabecera. No se escribe ni envía. Es evidencia
de foco y activación sobre la build sin firma, no locución ni toda la matriz.
VoiceOver vuelve a off; Capture Keyboard permanece off. Tras reinstalar,
Device Hub deja de exponer el subárbol de la app en esta sesión: las pruebas
posteriores usan capturas y coordenadas observadas. La falta del árbol
instrumental no se presenta como una auditoría accesible del producto.

Evidencia privada: ios-local-release-{arm64,adhoc}-build-20261008.log,
ios-release-artifacts-20261008.json, ios-release-{cold-1,cold-2,cold-3,
football-warm,football-cold,link-missing,link-malformed,offline,retry,
text-size-11}-20261008.png, ios-voiceover-{create-focus,form-heading}-20261008.png,
ios-api-links-log-20261008.log e ios-release-complete-log-20261008.log.
El log acotado no contiene coincidencias fatal/uncaught/crash; no certifica
ausencia de cualquier fallo. Producción y observabilidad permanecen apagadas.

Se repite :app:lintDebug sin excluir dependencias. Falla nuevamente en
:react-native-worklets:lintAnalyzeDebug con Cannot find a KaModule for the
VirtualFile, una excepción del analizador. Log privado
android-full-lint-retry-20261008.log. El gate completo continúa abierto;
no se desactivan reglas ni se actualizan dependencias para ocultar el fallo.

Retrospectiva: xcodebuild aprobado no garantiza un harness instalado adecuado
para servicios nativos. Contrastar firma, HTTP y pantalla evita modificar
negocio o transporte por un problema de instrumentación. Conservar explícito
el límite de altura de cabecera, sin desactivar el escalado para ocultarlo.


Cierre operativo verificado el 2026-10-09: make dev-down retira API, PostgreSQL
y Mailpit sin borrar volúmenes; no quedan contenedores locales activos ni
listeners de QA en 8080/5432/8025/1025/8083. La app iOS de QA se termina; se
conserva el simulador que ya estaba arrancado antes de la tanda. Text Size=3,
VoiceOver=off y Capture Keyboard=off restaurados. Android permanece apagado
con la APK debug restaurada. UTM informa K3s stopped y las tareas de dev,
producción y autostart siguen disabled. No se reinicia producción ni se publica
un nuevo release. El primer cierre fue rechazado por un fallo de revisión
automática debido al límite de uso; no ejecutó la acción y se completó tras
reanudar la sesión. Evidencia y backups se conservan.


#### 2026-10-09 — Parche de seguridad descubierto durante el cierre de CI

CI de 787444c falla en govulncheck con once avisos publicados el 8 de octubre:
GO-2026-6617, 6613, 6612, 6611, 6610, 6608, 6607, 6605, 6603, 6600 y 6599.
La CI anterior aprobada no invalida una detección posterior de la base viva.
Se aplica la política de parches aceptada en ADR-0012: toolchain Go 1.26.9
en ambos módulos y x/net 0.60.0; go get resuelve además los mínimos requeridos
de crypto, sync, sys, text y mod, con tidy y checksums separados.
Docker Hub publica 1.26.9-bookworm para amd64 y arm64: se alinea la base.
No se construyen ni despliegan imágenes de servicio en esta tanda. Los binarios
de producción siguen siendo los de v1.10.0 y K3s permanece apagado.

GOTOOLCHAIN=go1.26.9 make verify local termina con exit 0: formato, lint,
tests, generación, tidy de ambos grafos, build, exportación web y govulncheck.
La integración PostgreSQL opt-in queda omitida localmente con los servicios
apagados; CI la ejecuta en su base efímera. El análisis detallado conserva
GO-2026-5932 de OpenPGP en el módulo x/crypto, sin versión corregida: cero
vulnerabilidades en símbolos y paquetes importados, un aviso de módulo cuyo
paquete no usa este backend. No se declara limpio todo el grafo por ese exit 0.
Fuente: [aviso oficial](https://pkg.go.dev/vuln/GO-2026-6617).

Se revisa el disparador de lint de tests: con --tests=true ahora carga los
paquetes y muestra 38 hallazgos (informe limitado por regla) (errcheck 3, errorlint 3, gosec 5,
misspell 8, noctx 7, revive 8, staticcheck 3, unused 1). Se actualiza la deuda
y el comentario de configuración; tests:false permanece hasta resolverlos.
No se atribuye ya ese pendiente a una excepción del cargador ni se suprimen
reglas para declararlo aprobado. Las pruebas funcionales siguen ejecutándose.

Evidencia privada: qa-close-ci-failed-20261009.log,
qa-go-security-verify-20261009.log, qa-go-security-vuln-verbose-20261009.log,
qa-go-lint-tests-20261009.log y go-1.26.9-docker-manifest-20261009.json.
Retrospectiva: la base de vulnerabilidades cambia sin cambios de código.
Una actualización mínima puede corregir el gate de seguridad y, a la vez,
permitir revisar una excepción antigua; distinguir ambas deudas evita
confundir un parche probado con una certificación global o un despliegue.


#### 2026-10-09 — Cierre del análisis de tests Go

Se restaura run.tests: true y se retira la deuda de DECISIONS_TO_REVISIT.
El informe inicial de 38 incidencias estaba limitado por regla; la revisión
posterior usa --max-same-issues=0 --max-issues-per-linter=0 y acaba con 0 issues.
No se eliminan reglas, aserciones ni escenarios y no se añaden nolint.

Peticiones y sockets de test usan contexto explícito; el cierre HTTP comprueba
lectura y cierre antes de publicar su resultado. Los errores de dominio se
comparan con errors.Is. Los helpers siguen validando las mismas cuentas y
resultados; los ocho tokens concurrentes conservan sus bytes únicos mediante
un índice byte acotado por constante. Las cookies entrantes se representan
como cabecera Cookie: Secure corresponde a la cookie emitida por servidor.
Se elimina parser.ParseDir obsoleto conservando la inspección de imports de
todos los archivos Go de producción en cada directorio. Los falsos positivos
del corrector inglés se resuelven reformulando diagnósticos españoles.

Evidencia privada: qa-tests-lint-unlimited-20261009.log (inventario restante)
y qa-tests-lint-complete-20261009.log (cero incidencias).

Retrospectiva: un informe resumido no es un inventario exhaustivo. Revisar sin
límites de presentación permite cerrar la excepción sin debilitar el gate.
Esta fase no acredita Android Lint completo, la cabecera iOS a texto máximo
ni el recorrido global con lectores. Local, dev y producción siguen apagados.

Validación final: GOTOOLCHAIN=go1.26.9 make verify y go test -race ./...
terminan con exit 0. PostgreSQL permanece apagado y las pruebas que exigen
TM_INTEGRATION_DATABASE_URL se omiten localmente; CI ejecutará esa integración.
Govulncheck conserva el aviso de módulo OpenPGP previamente documentado,
con cero vulnerabilidades alcanzables. Logs privados:
qa-tests-enabled-verify-20261009.log y qa-tests-enabled-race-20261009.log.


#### 2026-10-09 — Títulos iOS adaptables al espacio real (ADR-0151)

El usuario acepta ADR-0151 y precisa que el cambio de posición solo ocurra
cuando no haya otra opción; además extiende la regla a todos los títulos de
navegación. La primitiva mide el texto completo con ancho reservado para
controles y separación, y compara su altura con useHeaderHeight del Router
sin el inset superior. No limita el escalado ni decide por fontScale.
La clave de medición cambia con nombre, ancho, escala y peso tipográfico.

Screen.navigationTitle aplica la regla en rutas de Cuenta, Ajustes, avisos,
crear, unirse, legales y gestión del torneo. La ficha usa el mismo hook y
sitúa el nombre en su cabecera desplazable. Las pantallas sin título no
reciben uno. La medición oculta queda fuera de accesibilidad y de los toques.
La primera prueba detecta que devolver null permite al stack recuperar el
nombre técnico de ruta; se corrige devolviendo un View vacío y declarando
el título nativo vacío cuando se presenta el nombre debajo.

Xcode 27 compila Release ARM64 con firma ad hoc para iPhone 17/iOS 27, sin
Metro. Pruebas visuales en español sobre el simulador existente:

- Ficha football: tamaño 3 conserva el nombre completo en la barra; tamaños
  7 y 11 lo muestran completo debajo sin cruzar Cerrar ni Acciones.
- Equipos a tamaño 7: permanece completo en la barra. Demuestra que texto
  ampliado no activa por sí solo la adaptación.
- Crear torneo: tamaño 3 mantiene el título en la barra; tamaño 11 lo muestra
  completo debajo. El toque CUA sobre Cerrar vuelve a Inicio, sin enviar.
- Ajustes a tamaño 11: título completo debajo y sin identificador técnico.
- Las proyecciones públicas del fixture antes y después son idénticas por cmp.

Cambiar Text Size en caliente mediante Device Hub deja algunos textos del
contenido con cajas anteriores; se registran capturas separadas y se contrasta
mediante arranque frío. El arranque frío recompone el contenido. No se presenta
la prueba del inspector como certificación del cambio de tamaño en dispositivo
físico. Device Hub sigue sin exponer el subárbol de la app tras reinstalar:
esto acredita disposición y cierre por capturas/toque, no locución VoiceOver
ni toda la matriz. Idiomas restantes y todas las rutas aún requieren recorrido
nativo; la implementación compartida no equivale a verificar cada combinación.
La app admite orientación portrait: no se fuerza otra orientación para el QA.

Checklist apps/client/AGENTS.md revisada: catálogos existentes, tokens y
primitivas compartidos, márgenes, separación frente a botones, cierre nativo,
semántica de cabecera y escalado intactos. Ninguna operación OpenAPI cambia;
los adaptadores y apiFetch se conservan. No se añade dependencia ni módulo
nativo. La compilación final incluye la ruta de vinculación de Google; no se
prueba aquí su acceso social real.

Evidencia privada: ios-title-reflow-{normal,size7-fixed,size11-fixed}-20261009.png,
ios-title-teams-size7-20261009.png, ios-title-create-{normal,size11,close-size11}
-20261009.png, ios-title-settings-size11-20261009.png; logs de build y verify
ios-title-reflow-*-20261009.log. Se conserva también el intento inicial para
mostrar el fallback detectado.

Android: :react-native-worklets:lintAnalyzeDebug --rerun-tasks --stacktrace
--no-daemon --max-workers=2, con LINT_PRINT_STACKTRACE=true, vuelve a fallar
en Cannot find a KaModule for the VirtualFile. La traza pasa por análisis de
script Gradle Kotlin; no identifica aquí un arreglo validado. Evidencia:
android-worklets-lint-diagnostic-20261009.log. El gate completo sigue abierto,
sin supresiones ni actualización de dependencias.

Retrospectiva: medir texto y espacio reales permite adaptar solo cuando hace
falta. Una devolución vacía de un renderer puede activar un fallback nativo;
comprobar barra y contenido juntos evita mostrar identificadores internos.
Las pruebas de título corto/largo al mismo tamaño son más útiles que un umbral
arbitrario de escala. No confundir el arreglo del título con toda la matriz
de accesibilidad o con un despliegue.

Cierre local de la fase: typecheck, make verify y compilación final Xcode
Release ARM64 terminan con exit 0; Xcode registra BUILD SUCCEEDED. El análisis
de vulnerabilidades no encuentra casos alcanzables ni paquetes importados
afectados; mantiene un aviso en un módulo requerido no llamado por el código.
La integración PostgreSQL completa corresponde al CI, no se atribuye a una
ejecución local que omita su URL. git diff --check pasa.

Se restaura Text Size a 3, VoiceOver y captura de teclado permanecen apagados
y se termina la app de QA. make dev-down completa el cierre, docker ps queda
vacío y utmctl confirma producción/K3s stopped. Se conservan volúmenes, backups
y evidencia; no hay publicación de producción ni activación de observabilidad.

#### 2026-10-09 — Contraste del motor Android Lint e inventario alcanzable

El [análisis de herramientas](ANDROID_LINT_ENGINE_ANALYSIS_2026-10-09.md)
registra los contrastes y fuentes primarias. La propiedad efectiva de AGP
-Pandroid.lint.useK2Uast=false permite completar Worklets. El contraste completo
:app:lintDebug :global-feedback:lintDebug alcanza lintReportDebug y termina
exit 1: 924 tareas, 191 ejecutadas. La app tiene dos errores y 45 avisos,
contados desde el XML, sin asumir que un build exitoso equivalga a cero avisos.
MissingPrefix procede de la marca data-generated del generador de filtros de
Expo; NewApi procede del atributo API 33 del generador de estilos de splash.

El módulo local se ejecuta después por separado con K1: exit 0, 166 tareas,
28 ejecutadas; cero errores, seis avisos (ViewConstructor una vez y UseKtx
cinco). ViewConstructor concierne al constructor requerido por herramientas
XML, no demuestra un fallo de creación del ExpoView por su AppContext. No se
introduce un constructor inválido ni una dependencia para silenciar avisos.
Los avisos de la app incluyen recursos generados, iconos, APIs, orientación,
permisos heredados y recomendaciones de actualización; el XML conserva todo
el inventario para priorizarlo con contexto, sin actualizaciones automáticas.

ADR-0152 queda Propuesto para decidir el coste de dos parches de generación.
No se aplican; ningún archivo fuente del cliente, dependencia, configuración
nativa o regla de lint cambia en esta fase. K1 aporta evidencia complementaria;
el gate K2 continúa abierto. Lint 9.0.1 también falla por una segunda excepción
FIR. El cambio upstream integrado el 7 de octubre no prueba disponibilidad de
un artefacto publicado, ni autoriza saltar la espera de ADR-0138.

Se vuelve a comprobar iOS desde CUA: Home visible, Text Size 3, VoiceOver y
captura de teclado apagados; Device Hub sigue sin el subárbol de la app. Se
termina la app después. Captura ios-accessibility-channel-recheck-20261009.png.
No se atribuye a ese contraste locución, orden de foco ni una matriz accesible.
API, PostgreSQL, Metro, emulador Android y producción no se arrancan.

Evidencia privada: android-full-lint-k1-20261009.log, android-feedback-lint-k1
-20261009.log e informes android-{app,global-feedback}-lint-k1-report-20261009.xml.
Retrospectiva: un crash de herramienta puede ocultar errores reales de recursos.
Separar motor, ejecución, informe y decisión permite avanzar el inventario sin
rebajar el gate ni confundir un workaround diagnóstico con un arreglo aprobado.

Decisión posterior del usuario: «Esperar versiones corregidas de Expo».
ADR-0152 pasa a Aceptado con alternativa B; no se implementan los dos parches.
Los dos errores, los avisos y el bloqueo K2 se conservan abiertos.


#### 2026-10-09 — Títulos de creación por idioma y revisión de updates

Se instala el artefacto final Release ARM64 ya compilado y se prueba creación
con Text Size 11 y idiomas por proceso: inglés muestra Create tournament en
dos líneas; italiano Crea torneo en una; francés Créer un tournoi en dos.
En los tres el nombre completo aparece bajo la barra y Cerrar queda separado.
La pulsación de Cerrar en inglés vuelve a Home. No se envían formularios ni
se afirma que se haya probado el cierre en italiano/francés.

Capturas privadas: ios-title-create-{en,it,fr}-size11-20261009.png,
ios-title-create-en-close-20261009.png e ios-locale-restore-es-20261009.png.
Se restaura Text Size 3, español, VoiceOver y captura de teclado apagados;
se termina la app. El canal de accesibilidad de la app continúa ausente,
por lo que esta evidencia visual no cierra foco ni locución VoiceOver.

El usuario solicita revisar actualizaciones pendientes. La
[revisión Expo/RN](EXPO_UPDATE_REVIEW_2026-10-09.md) encuentra una matriz madura
57.0.26 / RN 0.86.3, pero los generadores inspeccionados aún conservan los dos
errores Android. No se modifica paquete, lockfile, parche ni motor lint.
La expectativa CLI más reciente incluye versiones jóvenes. La resolución
transitiva y las builds de una nueva matriz quedan pendientes de su adopción.

Retrospectiva: completar evidencia visual por idioma y comprobar paquetes
oficiales permite avanzar QA sin certificar accesibilidad o correcciones no
probadas. API, PostgreSQL, Metro, emulador Android, observabilidad y producción
no se arrancan en esta fase; se conservan evidencia y datos.


#### 2026-10-09 — Instalación autorizada de la matriz madura

El usuario pide instalar lo actualizable tras revisar Expo/RN. Se aplica Expo
57.0.26 / RN 0.86.3, se actualizan paquetes compatibles y transitivas maduras,
sin exclusiones de edad. El [informe de actualización](EXPO_UPDATE_REVIEW_2026-10-09.md)
registra versiones, pares, parches, compatibilidad y evidencia privada.
Instalación congelada, typecheck, web, make verify y regresiones de dependencias
pasan; el check online conserva los cinco directos jóvenes. Se mantienen los
avisos altos de node-forge/braces con correcciones todavía no publicadas.

Android Debug ARM64 compila, se instala conservando datos y recorre Home ->
Crear torneo -> Cerrar -> Home con Router 57.0.24 tras limpiar el caché Metro.
No se envían formularios ni se modifican fixtures. El emulador de esta sesión
queda apagado. K2 sigue fallando en Worklets; el XML diagnóstico K1 conserva dos
errores y 45 avisos de app. Global-feedback pasa con cero errores y seis avisos.
No se debilita el gate ni se aplican los dos parches rechazados por el usuario.

Las primeras builds nativas se interrumpen por disco lleno. Se conserva el
último binario iOS y la evidencia y se elimina su DerivedData antiguo; la
recompilación Android pasa. Después del desbloqueo confirmado por el usuario,
iOS Debug ARM64 compila, se instala y muestra Home con Router 57.0.24. Crear
torneo conserva título en barra a Text Size 3 y debajo a 11; al relanzar a 11
los chips hacen reflow y el cierre nativo vuelve a Home. Cambiar la escala en
caliente recorta otros textos: se registra el contraste sin atribuirlo al
upgrade ni cerrar accesibilidad. El árbol AX de la app sigue ausente.

Se restaura tamaño 3 y español; se termina la app iOS y Metro. No se arranca API,
PostgreSQL, observabilidad ni producción. La revisión vinculada conserva toda
la evidencia y la retrospectiva; no se presenta este smoke test como QA total.


#### 2026-10-09 — Recalcular texto iOS al cambiar Dynamic Type en caliente

Problema reproducido en la development build Expo 57.0.26 / RN 0.86.3: pasar
Text Size 3 -> 11 con Crear torneo montado conserva las alturas antiguas de
etiquetas y chips, aunque los glifos crecen. La captura privada
`dynamic-type-live-before-20261009.png` vuelve a mostrar el recorte; el arranque
previo usa el binario actualizado y Metro con Router 57.0.24.

El [reporte upstream #57512](https://github.com/react/react-native/issues/57512)
describe el mismo patrón y documenta que una actualización de propiedades basta
para volver a medir. Es evidencia de contexto, no prueba de que toda su causa
interna coincida con esta app. La dependencia instalada conserva contenido del
párrafo y marca su medición como sucia cuando cambia props.

Se concreta la accesibilidad aceptada en ADR-0054/0055/0151 en la primitiva
`Text`: `useWindowDimensions().fontScale` actualiza un `nativeID` único (`useId`
+ escala) solo en iOS. No se altera la etiqueta accesible ni se remonta el texto,
la ruta o el formulario. Se conserva `allowFontScaling` nativo y los tokens;
no se introduce un parche de dependencia ni una librería. Frente a remontar con
`key={fontScale}`, evita sustituir nodos; frente a un parche Fabric, tiene menor
coste de actualización. La adaptación se retira cuando el renderer pase esta
misma prueba sin la propiedad de invalidación.

Prueba con la pantalla montada tras cargar el cambio: 3 -> 11 muestra Deporte y
chips completos y reordenados (`dynamic-type-live-candidate-20261009.png`),
conserva el deporte Tenis y el borrador QA escala, comprobados también en AX.
11 -> 3 -> 7 muestra etiquetas completas y campo activo; escribir después de
cambiar a 7 añade texto al mismo borrador sin reenfocarlo. La captura
`dynamic-type-size7-focus-20261009.png` muestra cursor, borde de foco y
QA escala viva 7. Otra transición 7 -> 11 -> 3 conserva edición y añade 11 al
valor. A 7 el título cabe en barra; a 11 pasa completo debajo: la regla sigue
basada en medición y no en un umbral de escala.

El árbol AX de la app está disponible en esta sesión: hay una sola cabecera
Crear torneo, Cerrar es botón, Tenis consta seleccionado y los campos tienen
nombre localizado. Esto avanza la evidencia semántica, pero no certifica locución
ni recorrido VoiceOver. El gesto de scroll solicitado a Device Hub no desplaza
el formulario; no se acredita aquí lectura visual del último control a 11.
No se envían formularios ni se cambia ningún fixture. Se limpia el borrador,
Cerrar vuelve a Inicio y se restaura Text Size 3, VoiceOver apagado y captura
de teclado apagada.

Checklist de cliente: cambio compartido y acotado a iOS, sin copy nuevo ni
colores/medidas locales, semántica y escalado conservados, sin operación OpenAPI
ni endpoint modificado. Typecheck y exportación web pasan. Los bloqueos Android
K2/generadores de ADR-0152 y el resto de la matriz global siguen abiertos.

Retrospectiva: una captura tras relanzar no cubre el cambio de preferencia con
contenido montado. La comprobación de altura debe acompañarse de estado y
edición retenidos; una propiedad de invalidación permite corregir el defecto
observado sin asumir la complejidad de mantener el renderer nativo.

Cierre del gate: `make verify` pasa, además de typecheck y exportación web
independientes. La integración PostgreSQL local se omite sin URL de pruebas;
la comprobación remota se realiza en CI tras subir a develop. Metro queda
apagado y se termina la app y el simulador arrancado en esta sesión. API,
PostgreSQL, observabilidad y producción no se han arrancado.


#### 2026-10-09 — Safe area inferior en Crear torneo

El usuario identifica la franja inferior desaprovechada y precisa la regla:
aprovechar el safe area y añadir la separación dentro del contenido. El código
confirma que `Screen` sumaba inset + `space[4]` como padding del padre: el
viewport de `KeyboardAwareScrollView` terminaba antes de esa franja.

Se conserva el cálculo y se cambia su propietario: Crear torneo usa
`Screen bottomInset="none"`; la primitiva compartida admite
`bottomInset="safe-area"` y reserva inset nativo + 16 px al final de
`contentContainerStyle`. Web conserva 16 px sin inset nativo. El valor por
defecto de la nueva opción es none para respetar los cálculos específicos de
las rutas con tabs. No se cambia navegación, títulos, escalado ni operación
OpenAPI. La decisión concreta del usuario mantiene los tokens y la regla de
layout compartido de ADR-0054/0055; no requiere una nueva dirección de stack.

La vista compacta de Device Hub permite observar desplazamiento del formulario
previo a la corrección. Arranque con Text Size 11 alcanza el botón completo
(`accessibility-create-bottom-size11-cold-20261009.png`); pulsarlo vacío muestra
los errores localizados de ambos campos y sus etiquetas AX incluyen el mensaje.
No se crea un torneo ni se inicia sesión. En el cambio en caliente, la captura
previa `accessibility-create-bottom-size11-20261009.png` muestra el viewport
cortado antes del borde inferior. No se declara que ese gesto hubiese alcanzado
su límite máximo ni se infiere de él un fallo del motor de scroll.

Tras el cambio, `accessibility-create-safe-area-after-20261009.png` muestra el
contenido llegando a la zona inferior antes vacía a Text Size 11.
`accessibility-create-safe-area-size3-20261009.png` muestra formulario completo,
botón y card a tamaño normal. La separación final queda dentro del scroll,
según el código. Los gestos posteriores del canal no aportan evidencia adicional
del último control a 11 después del cambio; esa comprobación visual y VoiceOver
continúan pendientes. Se restaura tamaño 3 y se cierra a Inicio sin conservar
borrador; VoiceOver y captura de teclado permanecen apagados.

Typecheck, formato de los archivos modificados y exportación web pasan. La
checklist de cliente conserva copy localizado, tokens, cierre nativo y API
sin cambios; no se desactiva el escalado. Metro, app y simulador de esta sesión
se apagan. API, PostgreSQL, observabilidad y producción no se arrancan.

Retrospectiva: safe area y separación no obligan a recortar el viewport. En
contenido desplazable, reservarlos dentro de su contenido permite recorrer la
superficie hasta el borde y mantiene protegido el último control. La captura
de viewport y la de final de scroll responden a preguntas distintas.

Gate final: `make verify` pasa. Integración PostgreSQL local omitida por no
definir TM_INTEGRATION_DATABASE_URL; no se presenta como ejecutada localmente.
El commit anterior de Dynamic Type `ebb6649` cerró CI con éxito en
[run 37939253169](https://github.com/joseantoniogarciay/TournamentsManager/actions/runs/37939253169).

#### 2026-10-09 — Cierre visual del safe area a tamaño 11

Se carga el commit `456d9a3` en la development build instalada del simulador
iPhone 17 (iOS 27.0), sin arrancar API ni PostgreSQL. Desde Crear torneo a
Text Size 3 se cambia a 11 con la pantalla montada y se desplaza el formulario
en la ventana compacta de Device Hub. La captura privada
`accessibility-create-safe-area-bottom-live11-20261009.png` acredita el botón
«Inicia sesión para crearlo» completo, con sus dos líneas dentro de la
superficie y separación inferior. Cierra la comprobación visual pendiente
tras el ajuste del safe area; no certifica VoiceOver ni todas las rutas.

Pulsar el botón con los campos vacíos muestra los dos errores localizados;
las etiquetas AX de los campos incluyen su respectivo mensaje. No se inicia
sesión ni se crea un torneo. Cerrar vuelve a Inicio y se restaura Text Size 3.
VoiceOver y captura de teclado permanecen apagados; el simulador ya estaba
arrancado al comenzar esta comprobación y se conserva así. Metro y la app se
terminan. Dev, observabilidad y producción no se arrancan.

CI del cambio `456d9a3` termina correctamente en
[run 37944204703](https://github.com/joseantoniogarciay/TournamentsManager/actions/runs/37944204703).
Esta fase añade evidencia y documentación, sin cambios de implementación.
Los bloqueos Android de ADR-0152 y los recorridos pendientes de la matriz
global continúan abiertos.

Retrospectiva: la prueba en caliente acredita el último control después de
la nueva medición de texto. Combinar captura y validación local permite
comprobar su lectura y acción sin introducir datos ni depender de servicios.

#### 2026-10-09 — VoiceOver: foco y scroll de Crear torneo

Development build instalada en iPhone 17, iOS 27.0, con JavaScript de
`fd4f83d` y Metro local. API, PostgreSQL y observabilidad permanecen apagados.
VoiceOver se activa temporalmente en Device Hub. En Inicio, gestos horizontales
avanzan el rectángulo de foco desde título a descripción y Crear torneo.
En el formulario avanzan por cabecera, Deporte y los deportes en orden.

El foco está en Tenis al cambiar Text Size 3 → 11 con la vista montada:
permanece en ese control y el layout recompone el contenido. Pádel está
seleccionado por una pulsación del canal, sin introducir nombres. Los gestos
del lector continúan por Voleibol, Tenis de mesa, Bádminton, Pádel, explicación
de sets, etiqueta y campo del torneo, sección y ayuda del participante,
etiqueta y campo del participante y botón final. VoiceOver desplaza el
formulario automáticamente al alcanzar contenido fuera del viewport.
El botón completo queda enfocado y visible a 11. El gesto inverso alcanza el
campo anterior y la navegación inversa completa alcanza Cerrar nativo.

Evidencia privada en el directorio habitual:
`ios-voiceover-create-bottom-focus-size11-20261009.png`,
`ios-voiceover-create-reverse-field-size11-20261009.png` y
`ios-voiceover-create-close-focus-size11-20261009.png`.
Las capturas del canal durante el recorrido corroboran los focos intermedios.
No se acredita locución, escritura con lector ni activación por doble toque:
un doble clic de CUA en otro punto cambió Pádel mientras el foco seguía en
Tenis. Por eso las pulsaciones del canal no se equiparan a la activación
del elemento enfocado. Esto acota también la evidencia de apertura del
formulario de esta pasada; la navegación por gesto sí muestra avance real.
No se infiere un defecto de activación del producto por esa diferencia.

Se restaura Text Size 3 y VoiceOver off; Capture Keyboard continúa off y se
conserva el volumen original. Cerrar vuelve a Inicio sin guardar ni enviar.
Metro y la app se terminan, conservando el simulador previamente arrancado.
No cambia implementación ni se repiten los gates de código ya aprobados.
Producción sigue apagada por la decisión previa del usuario. El inventario
global conserva locución, activación/edición con lector, otras rutas y gates
Android pendientes.

Retrospectiva: un árbol AX estable puede coexistir con foco y scroll reales
del lector. Sus capturas permiten acreditar navegación; una pulsación enviada
por instrumentación no demuestra por sí sola activación VoiceOver.

#### 2026-10-09 — TalkBack y safe area Android tras la actualización

Pixel_API_34 (API 34) se arranca para esta tanda con la APK Debug actualizada
del 9 de octubre y JavaScript de `d1c83e3`. Metro está activo; API, PostgreSQL,
dev, observabilidad y producción permanecen apagados. Se conservan antes de
probar los tres ajustes de accesibilidad, font_scale=1.0 y reverses vacíos.
Cambiar a font_scale=2.0 recrea la actividad y vuelve a Inicio; se abre Crear
torneo desde esa pantalla estable. No se presenta esa recreación como un
recorrido aprobado de conservación de ruta.

UIAutomator se usa únicamente antes de activar TalkBack. Con el lector activo
se emplea entrada táctil `adb emu event mouse` y capturas. La petición de
notificaciones de la Suite de Accesibilidad se rechaza; no es necesaria para
el recorrido. La muestra acredita foco en cabecera, Deporte, Fútbol,
Baloncesto, Pádel y campo del torneo. El doble toque fuera de Pádel activa
el deporte enfocado, conserva el foco y actualiza su descripción y etiquetas.
Activar el campo del torneo vacío abre el teclado; no se escribe. Cerrar
teclado permite continuar hasta la sección de participante y el botón final.

El recorrido de foco desplaza automáticamente el formulario. La captura
`android-talkback-safe-area-focus-19-20261009.png` muestra Crear torneo
completo y enfocado, con separación respecto a la barra de gestos y final de
card visible. Se cierra la comprobación Android de acceso al último control
en este formulario al 200 %, sin acreditar locución ni todos los lectores,
rutas o versiones. La variante autenticada conserva un borrador previo con
nombre de torneo vacío y participante «Incidencia Local». No se pulsa Crear,
no se guarda ni se crea ningún fixture.

La navegación inversa vuelve a Fútbol; el doble toque restaura la selección
original. El foco alcanza Cerrar y activarlo vuelve a Inicio. El aviso común
de conexión allí es coherente con la API apagada y no muestra detalles internos.
No se modifica la sesión ni se cambian nombres. Se restauran y comprueban
exactamente los ajustes originales, se retira únicamente reverse tcp:8082,
se termina la app y se apaga el emulador iniciado para QA. Metro queda apagado.

Evidencia privada: `android-talkback-safe-area-{before,restored}-20261009.json`,
capturas `android-talkback-safe-area-*`, estado del servicio, logs de Metro y
emulador y script `qa-talkback-safe-area-20261009.py` en el directorio habitual.
La secuencia de navegación inversa rápida no avanzó un elemento por cada
evento; se vuelve a verificar visualmente Fútbol y Cerrar antes de activarlos.
No se deduce un defecto del producto de la cadencia del instrumento.
No cambia implementación. El gate lint Android de ADR-0152 continúa abierto;
no se repite con la misma matriz ni se aplican los parches rechazados.

Retrospectiva: comprobar entrada táctil, foco y resultado permite distinguir
activación del lector de un toque por coordenadas. El acceso al teclado y al
último botón aporta evidencia específica tras el upgrade sin enviar un
formulario ni arrancar servicios adicionales.

#### 2026-10-09 — Inventario del safe area restante

La siguiente pasada visual iOS no puede iniciarse: CUA informa que el Mac está
bloqueado y no puede desbloquearlo. Se solicita desbloqueo manual; no se arranca
Metro, API ni otro servicio mientras no haya acceso al simulador. Se continúa
con revisión estática del propietario del inset inferior. El cierre de Crear
torneo no se extiende implícitamente al resto del cliente.

| Ruta | Evidencia en código | QA y aplicación restantes |
| --- | --- | --- |
| join-team | Screen declara safe-area; el formulario usa KeyboardAwareScrollView sin reserva propia | Comprobar formulario y último control; conservar protección de las ramas estáticas de carga/error |
| link/password-reset | Las ramas de error y formulario usan Screen por defecto y KeyboardAwareScrollView sin inset propio | Revisar ambos layouts; el cambio efectivo de credencial continúa reservado a intervención humana |
| tournament/[id]/administrators/add | Screen por defecto; scroll con paddingBottom space[5] | Revisar viewport, teclado y último resultado; evitar duplicar separación al mover el inset |
| tournament/[id]/transfer | Screen por defecto; scroll con paddingBottom space[5] | Revisar viewport y último resultado sin ejecutar transferencia |
| account/notifications | Screen por defecto; ScrollView con paddingBottom space[8] | Completar desplazamiento real iOS; conservar padding de contenido y no borrar notificaciones |
| tournament/[id]/standings | Screen por defecto; ScrollView de tabla con paddingBottom space[5] | Revisar final vertical y control horizontal/sticky antes de adaptar layout |

En estas seis rutas, Screen reserva inset + space[4] fuera del scroll y por
tanto recorta su viewport. Es un hecho de implementación; no acredita por sí
solo un control inaccesible ni una reproducción visual de recorte. No se
modifica código durante este inventario. Las rutas bajo tabs ya usan su
cálculo específico dentro del contenido; equipos y administradores declaran
bottomInset none y reserva propia en su scroll. Los documentos legales tienen
bottomInset none y padding propio, y requieren una revisión distinta del
inset nativo; no se clasifican como la misma franja fija del padre.

Las tres integraciones anteriores de evidencia cierran CI con éxito:
`fd4f83d` ([37946222813](https://github.com/joseantoniogarciay/TournamentsManager/actions/runs/37946222813)),
`d1c83e3` ([37947317466](https://github.com/joseantoniogarciay/TournamentsManager/actions/runs/37947317466)) y
`03ed43c` ([37958638867](https://github.com/joseantoniogarciay/TournamentsManager/actions/runs/37958638867)).
No se publica producción ni se repite un gate nativo con la misma matriz.

Retrospectiva: localizar quién reserva el inset permite priorizar la siguiente
pasada sin declarar una migración global terminada. Las ramas sin scroll y
los paddings propios deben revisarse antes de trasladar la reserva.
