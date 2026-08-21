# Observabilidad

> Base aceptada: OpenTelemetry, Prometheus, Grafana, Loki y Tempo. El
> OpenTelemetry Collector queda aplazado hasta que una necesidad medida lo
> justifique. Véase [ADR-0020](../adr/0020-use-minimal-correlated-observability.md).

## Resultado buscado

La observabilidad debe permitir responder:

- ¿está funcionando el servicio para el usuario?
- ¿qué cambió?
- ¿dónde está el cuello de botella o fallo?
- ¿qué usuarios, operaciones o dependencias están afectados?
- ¿qué acción reduce el impacto?

## Señales

- **Logs:** eventos estructurados y accionables, sin secretos.
- **Métricas:** comportamiento agregado, capacidad y objetivos de servicio.
- **Trazas:** recorrido y latencia entre límites.
- **Perfiles/eventos:** solo cuando respondan una pregunta concreta.

Las señales deben compartir contexto de correlación y convenciones de nombres.

## Base mínima aceptada

- **Logs:** JSON a salida estándar mediante `log/slog`; Loki los almacena y
  Grafana permite buscarlos. Un log es un evento discreto, no una traza ni una
  sustitución de `fmt.Println` en producción.
- **Métricas:** Prometheus recopila medidas agregadas; Grafana las visualiza.
- **Trazas:** OpenTelemetry instrumenta límites técnicos y Tempo conserva el
  recorrido de una operación. Una traza se compone de *spans* —por ejemplo,
  HTTP entrante y consulta PostgreSQL—, no solo de llamadas de red.
- **Correlación:** cada log incluirá el identificador de traza y span cuando el
  contexto exista. No se registran secretos, tokens, credenciales ni PII.

La instrumentación automática cubre HTTP y PostgreSQL. Quien implementa el
código decide los spans manuales solo cuando representen una operación
operativamente significativa que no esté cubierta; no se añade un span por
función. Los nombres y atributos técnicos siguen las convenciones semánticas de
OpenTelemetry. Los eventos de negocio se decidirán junto al caso de uso que los
necesite.

El servicio debe degradarse de forma segura si un backend de telemetría no está
configurado o no está disponible. El dominio no importa SDKs ni tipos de los
backends.

## Primer corte ejecutable — refresh de sesión

La primera pregunta operativa es: **¿por qué falló o se degradó un refresh de
sesión web?** La ruta observada es `POST /v1/sessions/refresh`; atraviesa HTTP,
la protección CSRF y PostgreSQL sin incorporar todavía SMTP, Google ni lógica de
ligas.

`make dev-up` inicia, además de API, PostgreSQL y Mailpit, el stack local:

- Grafana en `http://127.0.0.1:3000`;
- Prometheus en `http://127.0.0.1:9090`;
- Loki, accesible desde Grafana;
- Tempo, que recibe OTLP/HTTP en `127.0.0.1:4318`;
- Promtail, que recoge exclusivamente el `stdout` JSON del contenedor `api`.

Grafana provisiona las tres fuentes de datos. La API expone métricas agregadas
en `/metrics` y registra cada petición terminada con método, plantilla de ruta,
estado, duración y, si existe, `trace_id` y `span_id`. No registra cuerpos,
cookies, tokens, query strings, SQL ni argumentos SQL. El identificador de
traza no se convierte en etiqueta de Loki o Prometheus, para no elevar la
cardinalidad.

Las trazas priorizan nombres operativos, no contenido sensible: el span HTTP se
llama, por ejemplo, `POST /v1/sessions`; las operaciones locales Argon2id se
ven como `auth.password.hash` o `auth.password.verify` y declaran únicamente
`argon2id`; PostgreSQL usa el nombre estático de la operación —generada por
sqlc o anotada en una consulta manual—, por ejemplo
`postgresql.FindLocalAccountForLogin`; y la entrega de correo se ve como
`smtp.send.verification` o `smtp.send.password_reset`. Los decoradores técnicos
no añaden destinatarios, tokens, contraseñas, hashes, SQL, argumentos SQL ni
contenido del mensaje. El mismo decorador Argon2id cubre la creación de cuenta
y el cambio de contraseña tras consumir un enlace de restablecimiento.

Cuando un límite técnico falla, el span no exporta el error bruto. Registra
solamente `tournaments_manager.failure.reason`, con valores cerrados como
`database.unavailable`, `database.constraint_failed`,
`database.query_failed`, `smtp.delivery_failed`, `request.cancelled` o
`request.timeout`. Si el borde HTTP recibe un `5xx` que la feature no pudo
clasificar, registra `request.failed`: conserva una causa segura en el span raíz
y en el log correlacionado sin inventar una dependencia concreta. Cada feature
puede añadir una causa de negocio segura cuando aporte recuperación distinta; no
se deduce centralmente del código HTTP.

La revisión de un endpoint cubre sus salidas de éxito, validación, límite de
tasa, negocio, límites técnicos y cancelación. Un rechazo esperado que necesite
diagnóstico añade en el span HTTP la misma clave con una causa cerrada —por
ejemplo `validation.rejected` o `rate_limit.exceeded`—, no un span adicional.
La feature decide las causas de negocio: no se centralizan por estado HTTP ni
se incluyen valores introducidos, longitudes, mínimos, identificadores o PII.

`OTEL_TRACES_ENDPOINT` es opcional. Cuando falta o Tempo deja de estar
disponible, la API mantiene los logs JSON y las métricas y no deja de servir
peticiones por un error de exportación.

El procedimiento de diagnóstico y la prueba de indisponibilidad controlada de
PostgreSQL están en el [runbook de refresh de sesión](../runbooks/session-refresh-observability.md).

## SLO local — refresh de sesión

El primer objetivo de servicio aceptado es `POST /v1/sessions/refresh`:

- disponibilidad de al menos **99,5 %** en ventana móvil de 30 días; una
  respuesta `5xx` consume presupuesto y cualquier otra respuesta no;
- latencia **p95 inferior a 500 ms**, evaluada sobre una ventana de cinco minutos.

Prometheus calcula la disponibilidad y el presupuesto consumido como series de
grabación. Expone una alerta local cuando los `5xx` superan el 7,2 % durante
cinco minutos y otra cuando el p95 supera 500 ms durante quince minutos.
Grafana aprovisiona el dashboard **SLO — Refresh de sesión**. No hay
Alertmanager, notificación remota, retención de producción ni SLOs generales:
véase [ADR-0098](../adr/0098-define-local-session-refresh-slo.md).

## Recorridos revisados

| Ruta | Spans hijos relevantes | Decisión |
| --- | --- | --- |
| `POST /v1/registrations` | `auth.password.hash`, operaciones PostgreSQL de alta, `smtp.send.verification` | Instrumentar CPU costosa, transacción y dependencia SMTP; los rechazos usan `validation.rejected` o `rate_limit.exceeded`. |
| `POST /v1/password-resets` | `postgresql.CreatePasswordReset`, `smtp.send.password_reset` si existe una cuenta elegible | No revelar la existencia de la cuenta en los atributos ni crear un span cuando no se envía correo; rechazos: `validation.rejected` y `rate_limit.exceeded`. |
| `POST /v1/password-reset-links` | `postgresql.InspectPasswordReset` | Los rechazos usan `validation.rejected` o `credential.reset_link_invalid`, sin exponer el token. |
| `POST /v1/password-reset-confirmations` | `auth.password.hash`, `postgresql.ConsumePasswordReset` | La actualización de credencial, revocación y sesión es una sola consulta atómica; rechazos: `validation.rejected` o `credential.reset_link_invalid`. |
| `POST /v1/registration-verifications` | `postgresql.VerifyRegistrationAndCreateSession` | No añadir spans para SHA-256, aleatoriedad o CTE internos; rechazos: `validation.rejected` o `credential.verification_link_invalid`. |
| `POST /v1/google-login-challenges` | `postgresql.CreateGoogleLoginChallenge` | El nonce y el identificador opacos no salen como atributos; el único límite relevante es PostgreSQL y sus categorías técnicas cerradas. Si Google no está configurado, la salida `503` es `identity.google_unavailable`. |
| `POST /v1/google-sessions` | Operaciones PostgreSQL de challenge, identidad y sesión | La prueba OIDC, token, nonce, email, subject y sesión quedan fuera de atributos; `validation.rejected`, `credential.google_challenge_invalid` e `identity.email_conflict` distinguen recuperaciones útiles. El alta pendiente (`202`) no es un fallo; Google no configurado usa `identity.google_unavailable`. |
| `POST /v1/me/google-identities` | Operaciones PostgreSQL de ticket, challenge e identidad | La sesión, ticket, prueba Google e identidad no se exportan; rechazos: `validation.rejected`, `credential.reauthentication_invalid` e `identity.google_conflict`. La autenticación y CSRF se resuelven en middleware sin spans adicionales; `503` por Google no configurado es `identity.google_unavailable`. |
| `DELETE /v1/me/google-identities` | `postgresql.ConsumeReauthenticationTicketAndRemoveGoogle` | La eliminación es una transición atómica que conserva la credencial local; el ticket no se registra. Rechazos: `validation.rejected` o `credential.reauthentication_invalid`; errores de PostgreSQL usan la categoría técnica cerrada. Google no configurado usa `identity.google_unavailable`. |
| `POST /v1/sessions` | `postgresql.FindLocalAccountForLogin`, `auth.password.verify`, `postgresql.CreateLocalLoginSession` | Rechazos: `validation.rejected`, `rate_limit.exceeded` o `authentication.credentials_rejected`, sin distinguir cuenta, contraseña o estado. |
| `GET /v1/sessions` | `postgresql.GetCurrentSession` | Una sesión que deja de ser válida después de autenticarse se marca como `session.invalid`; la autenticación previa no añade un span ni atributos de token. |
| `POST /v1/sessions/refresh` | `postgresql.RotateSessionTokens` | La rotación del token opaco y la sesión se diagnostican como una única transición atómica. Un refresh ausente, duplicado o ya consumido se marca como `session.refresh_invalid`, sin exportar el token. |
| `DELETE /v1/sessions` | `postgresql.RevokeSession` | La revocación es idempotente. Los rechazos de autenticación se resuelven en el middleware sin crear spans adicionales ni registrar la credencial. |
| `GET /v1/me/access-methods` | `postgresql.GetAccessMethods` | La consulta no añade razones de negocio: email, username y métodos forman parte de la respuesta pero no de atributos de traza. |
| `POST /v1/me/reauthentication-tickets` | `auth.password.verify`, `postgresql.GetCurrentPasswordHash`, `postgresql.CreateReauthenticationTicket` | La validación usa `validation.rejected`; una identidad federada de otra cuenta es `reauthentication.identity_conflict` y un desafío, contraseña o sesión no válidos es `reauthentication.invalid`. No se incluyen challenge, ticket, identidad, contraseña ni token. |
| `PUT /v1/me/local-credential` | `auth.password.hash`, `postgresql.ConsumeReauthenticationTicketAndSetPassword` | Los rechazos se limitan a `validation.rejected` y `reauthentication.invalid`; no se crean spans para SHA-256 del ticket ni aleatoriedad. |
| `DELETE /v1/me/local-credential` | `postgresql.ConsumeReauthenticationTicketAndRemovePassword` | Un ticket inválido es `reauthentication.invalid`; intentar dejar la cuenta sin método de acceso es `access_method.last_remaining`. |
| `DELETE /v1/me/account` | `postgresql.ScheduleAccountDeletion` | Una cuenta con ligas propias no puede programar su borrado y se marca como `account.deletion_owned_leagues`; el ID y la fecha efectiva no son atributos. |
| `GET /v1/usernames/{username}/availability` | `postgresql.IsUsernameAvailable` | La plantilla de ruta y el nombre de consulta son estáticos; username e IP quedan fuera de los atributos; rechazos: `validation.rejected` y `rate_limit.exceeded`. |
| `GET /v1/users` | `postgresql.SearchPublicUsernames` | La búsqueda conserva la query string fuera de HTTP y el texto buscado fuera de PostgreSQL; rechazos: `validation.rejected` y `rate_limit.exceeded`. |
| `GET /v1/me/leagues`, `GET /v1/me/recent-leagues` | Consultas PostgreSQL estáticas de colecciones | Éxito devuelve las proyecciones; filtros, cursor y límite inválidos usan `validation.rejected`. No hay limitador específico. Un fallo del límite PostgreSQL o cancelación usa `database.*`, `request.cancelled` o `request.timeout` en el span HTTP; no se registran cuenta, cursor ni filtros. |
| `PUT` \| `DELETE /v1/me/leagues/{leagueId}/follow` | `postgresql.FollowVisibleLeague`, `postgresql.UnfollowLeague` | Éxito es idempotente. ID inválido: `validation.rejected`; liga no visible: `league.not_found`. No hay límite de tasa propio ni spans por las ramas de seguimiento. |
| `POST /v1/leagues`, equipos, inicio, cancelación, resultado y finalización; `GET /v1/leagues/{leagueId}` | Operaciones PostgreSQL estáticas del ciclo de liga | La entrada inválida usa `validation.rejected`; los rechazos de negocio distinguen `league.forbidden`, `league.not_found`, y conflictos cerrados de inicio, equipos, retirada, resultado, cancelación o finalización. Sin limitador específico. Los fallos técnicos y cancelaciones conservan las categorías seguras comunes en el span raíz. |
| Administración de ligas: listar, asignar, eliminar y transferir | Operaciones PostgreSQL estáticas de administración | Éxito devuelve o modifica únicamente la proyección contractual. Entrada inválida: `validation.rejected`; negocio: `league.forbidden`, `league.not_found`, `league.administrator_conflict` o `league.ownership_transfer_conflict`. Nombres de usuario e IDs quedan fuera de atributos, logs y nombres de spans; no existe un límite de tasa específico. |
| Bandeja de notificaciones: lista, contador, marcar leídas y borrar una o todas | Consultas y mutaciones PostgreSQL estáticas de `account_notifications` | Las mutaciones son idempotentes y no añaden un rechazo de negocio. Sus fallos técnicos y cancelaciones usan `database.*`, `request.cancelled` o `request.timeout` en el span HTTP, sin cuenta, notificación ni error bruto. |

## Orden de diseño

1. definir el flujo crítico y su resultado correcto;
2. definir indicadores y objetivos;
3. enumerar modos de fallo;
4. elegir señales y atributos necesarios;
5. decidir instrumentación y backend;
6. crear visualización, alerta y runbook;
7. provocar un fallo y verificar el diagnóstico.

## Criterios para evaluar el stack

- estándares abiertos y portabilidad;
- integración con Go y Kubernetes;
- correlación entre señales;
- coste de operación y almacenamiento;
- retención y cardinalidad;
- experiencia local;
- seguridad de datos;
- facilidad de backup, upgrade y diagnóstico.

No se crearán paneles, alertas, SLO, retenciones de producción ni perfiles por
completitud. La excepción aceptada es el SLO local de refresh, porque responde a
una pregunta operativa concreta; la unidad mínima sigue siendo una pregunta
respondida de extremo a extremo y validada provocando un fallo.
