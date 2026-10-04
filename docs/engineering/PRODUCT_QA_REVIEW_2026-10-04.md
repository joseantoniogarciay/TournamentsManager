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
Las capturas nativas se observaron en la conversación, sin archivo independiente.

## Verificación automatizada ejecutada

| Comprobación                               | Resultado y alcance                                                                                                                   |
| ------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------- |
| `go test -json ./...`                      | 483 tests/subtests aprobados, 56 omitidos, ningún fallo. Las integraciones opt-in están omitidas en esta ejecución.                   |
| PostgreSQL opt-in en BD desechable         | 58 tests/subtests aprobados, sin fallos. `tm_product_qa_20261004` separada de la BD persistente de fixtures: la suite trunca cuentas. |
| `node --test tests/*.test.mjs`             | 34 aprobados en la pasada final, sin omitidos ni fallos (30 anteriores + 4 de borrador de invitación).                                                                                                 |
| `python3 tests/operational-safety.test.py` | 11 aprobados.                                                                                                                         |
| Cliente                                    | `pnpm run check` aprobado (formato, lint, TypeScript y OpenAPI); exportación web completada; cliente generado actualizado.                                   |

Los contadores incluyen subtests y no equivalen a funciones independientes,
a cobertura de líneas ni al número de casuísticas del producto.

## Matriz funcional

| Área / casos                                         | Automatizado                                                                                                    | API local real                                                                                                                                         | Revisión de pantalla                                                                                |
| ---------------------------------------------------- | --------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------ | --------------------------------------------------------------------------------------------------- |
| Registro, verificación, duplicados, enlace consumido | Dominio, HTTP y persistencia                                                                                    | Cuenta ficticia, verificación y reutilización 409                                                                                                      | Registro sin envío en ambos; verificación por enlace pendiente                                      |
| Login pendiente y reenvío                            | Secuencial, 8 solicitudes concurrentes, cancelación, sin crear sesión                                           | Reproducción del 500 y 202 tras corregir                                                                                                               | Login verificado en ambos; login pendiente pendiente                                                |
| Login verificado, logout y refresh                   | HTTP, cookies, CSRF, persistencia                                                                               | 200/204, sesión revocada 401 y refresh reutilizado 401                                                                                                 | Creador en ambos; cambio al delegado y sesión visible en iOS                                        |
| Recuperación y reautenticación                       | Persistencia, ticket de un uso, fallos DB/SMTP, timeout/cancelación y respuestas seguras                        | Solicitud, inspección, cambio, repetición 409, credencial vieja 401, nueva 200, solicitudes consecutivas 202/202                                       | Cambio de credencial por UI requiere intervención humana; pantallas y enlaces pendientes            |
| Google / Apple                                       | Challenges, validación, errores seguros, atomicidad y concurrencia con dobles de prueba                         | No se acredita acceso real a proveedores                                                                                                               | Orden de botones comprobado en ambos; OAuth real aplazado por el usuario a futuras pruebas en dev/prod                       |
| Biblioteca, recientes, seguir/dejar de seguir        | HTTP, persistencia y relaciones sin duplicar                                                                    | Paginación, seguimiento idempotente y consultas                                                                                                        | Inicio del creador y biblioteca delegada iOS; listas extensas y todos los filtros pendientes        |
| Administradores y exclusividad del creador           | HTTP y persistencia                                                                                             | Asignación autorizada/idempotente, delegado puede editar; listar admins, cancelar, completar y transferir rechazados 403                               | iOS muestra Administro y guarda 5–1, visible al volver; menú y Android delegado pendientes                                |
| Invitación de equipo                                 | HTTP y persistencia                                                                                             | Crear/regenerar, anterior inválida, inspección anónima, inscripción 201, duplicado 409, sin sesión 401, revocación idempotente, después de empezar 409 | Enlace compartido, registro y feedback nativos pendientes                                           |
| Composición de equipos                               | HTTP, persistencia                                                                                              | Último equipo 409, añadir, duplicado normalizado 409, eliminar antes de empezar                                                                        | Nombres largos y gestión previa revisadas parcialmente; envío y errores pendientes                  |
| Ocho deportes × tres formatos                        | Dominio, HTTP y persistencia                                                                                    | 16 combinaciones admitidas terminadas con campeón; 8 rechazadas 400 según contrato                                                                     | Liga fútbol en ambos; cuadro bádminton completado en ambos; restantes deportes pendientes           |
| Liga y mixto                                         | Grupos, dos vueltas, composición impar, retirada de clasificado, desempate repetido, congelación y concurrencia | Una vuelta y mixto tabla única; empate de corte con liguilla de desempate resuelto                                                                     | Clasificación fútbol parcial en ambos; mixto y desempates en pantalla pendientes                    |
| Eliminatoria                                         | Bracket, byes, correcciones, desempate fútbol/balonmano, todos los deportes                                     | Cuadro de cuatro participantes en ocho deportes                                                                                                        | Semifinal/final, ganador, parciales y origen de plazas bádminton; byes y correcciones UI pendientes |
| Resultados e incidencias                             | Marcadores, sets, límites, formas exclusivas, historia                                                          | Resultado inválido 400, no comparecencia, abandono con parcial, corrección jugada y permisos 403                                                       | Edición fútbol enviada en ambos; incidencias y todos los deportes pendientes                        |
| Retirada, cancelación y finalización                 | Dominio, HTTP y persistencia; co-campeones y concurrencia                                                       | Retirada, repetición 409, cancelación conserva lectura pública, edición/completar cancelado 409, finalización anticipada 409                           | Finalizado bádminton y cancelado fútbol visibles sin edición en ambos; retirada en ficha de equipo pendiente                          |
| Notificaciones y sugerencias                         | HTTP y persistencia                                                                                             | Lista, marcar todas leídas, contador 0, sugerencia 201, corta 400                                                                                      | Lista, vacíos, borrar y sugerencia enviada UI pendientes                                            |
| Cuenta: transferencia, baja y purga                  | Transferencia, ticket, baja/purga y anonimización                                                               | Baja del creador con torneos rechazada 409                                                                                                             | Formularios, transferencia y baja UI pendientes                                                     |
| Errores de transporte y fallback                     | Suite cliente/HTTP; nuevo reset 500 seguro                                                                      | Rechazos de negocio reales                                                                                                                             | Reintento, offline, cancelación de navegación y recuperación en ambos pendientes                    |

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
