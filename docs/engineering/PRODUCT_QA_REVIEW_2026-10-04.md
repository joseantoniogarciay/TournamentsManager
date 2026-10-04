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
| `node --test tests/*.test.mjs`             | 30 aprobados, sin omitidos ni fallos.                                                                                                 |
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
| Google / Apple                                       | Challenges, validación, errores seguros, atomicidad y concurrencia con dobles de prueba                         | No se acredita acceso real a proveedores                                                                                                               | Orden de botones comprobado en ambos; OAuth real y builds firmadas pendientes                       |
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
SO, navegación Android por botones, release y dispositivos físicos. OAuth real
requiere configuración y proveedores reales; Mailpit no prueba entrega externa.

## Retrospectiva técnica

La revisión visual con cuentas reales encuentra problemas de persistencia que
los dobles de repositorio no detectan. El mínimo suficiente para la regresión
es probar PostgreSQL de forma aislada, mantener un único token activo y revisar
cada salida HTTP. Un experimento visual que mejora solo parte del fondo debe
retirarse; los casos abiertos permanecen visibles hasta reproducir su cierre.
