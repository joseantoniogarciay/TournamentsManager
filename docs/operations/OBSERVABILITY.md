# Observabilidad

> Estado: Fase 3 cerrada y stack operativo en `local`, `dev` y `prod`.
> OpenTelemetry, Prometheus, Grafana, Loki y Tempo forman la base. El
> OpenTelemetry Collector queda aplazado hasta que una necesidad medida lo
> justifique. Véase [ADR-0020](../adr/0020-use-minimal-correlated-observability.md).

## Resultado buscado

La fiabilidad de cliente reserva el proyecto PostHog UE 255144 a producción,
según [ADR-0149](../adr/0149-reserve-the-single-posthog-project-for-production.md).
Beta/local no inicializan SDK y la analítica de uso queda apagada. Símbolos y
prueba de entrega/simbolización siguen el [runbook cliente](../runbooks/client-error-tracking.md);
la prueba real se difiere y no se acredita por configurar el SDK.

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
  recorrido de una operación. Una traza se compone de _spans_ —por ejemplo,
  HTTP entrante y consulta PostgreSQL—, no solo de llamadas de red.
- **Correlación:** cada log incluirá el identificador de traza y span cuando el
  contexto exista. No se registran secretos, tokens, credenciales ni PII.

Cuando una persona haya consentido la analítica de producto de PostHog en la beta pública, cada petición de
la API puede incluir un `X-Interaction-ID` UUID aleatorio. La API lo valida y
lo escribe exclusivamente como `interaction_id` en el log HTTP, junto a su
propio `trace_id` y `span_id`; no llega a atributos de span ni etiquetas de
métricas. Así PostHog permite localizar una interacción concreta y Grafana/Loki
la traza técnica correspondiente, sin aceptar un `trace_id` creado por cliente.

La instrumentación automática cubre HTTP y PostgreSQL. Quien implementa el
código decide los spans manuales solo cuando representen una operación
operativamente significativa que no esté cubierta; no se añade un span por
función. Los nombres y atributos técnicos siguen las convenciones semánticas de
OpenTelemetry. Los eventos de negocio se decidirán junto al caso de uso que los
necesite.

El servicio debe degradarse de forma segura si un backend de telemetría no está
configurado o no está disponible. El dominio no importa SDKs ni tipos de los
backends.

## Recorrido de referencia — refresh de sesión

La primera pregunta operativa es: **¿por qué falló o se degradó un refresh de
sesión web?** La ruta observada es `POST /v1/sessions/refresh`; atraviesa HTTP,
la protección CSRF y PostgreSQL sin incorporar todavía SMTP, Google ni lógica de
ligas.

ADR-0146 separa ejecución y diagnóstico: `make dev-up` inicia API, PostgreSQL y
Mailpit solo para pruebas. Únicamente por petición explícita,
`make dev-observability-up` activa el stack local con la API ya en marcha:

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

El receptor `POST /v1/risc/events` mantiene `credential.risc_invalid` como
causa de su span HTTP. Añade al log correlacionado, solo para diagnóstico,
`reason` con uno de `jwt.malformed`, `jwt.header_invalid`,
`jwt.claims_invalid`, `jwt.header_rejected`, `jwt.jti_missing`,
`jwt.issuer_missing`, `jwt.audience_missing`, `event.count_rejected`,
`jwt.audience_rejected`, `jwt.issuer_rejected`, `jwt.signature_invalid`,
`event.invalid` o `jwt.invalid`. Son categorías cerradas y no contienen el
SET, su `jti`, el estado de prueba ni identificadores de personas.

El primer objetivo de servicio aceptado es `POST /v1/sessions/refresh`:

- disponibilidad de al menos **99,5 %** en ventana móvil de 30 días; una
  respuesta `5xx` consume presupuesto y cualquier otra respuesta no;
- latencia **p95 inferior a 500 ms**, evaluada sobre una ventana de cinco minutos.

Prometheus calcula la disponibilidad y el presupuesto consumido como series de
grabación. Expone alertas `warning` cuando los `5xx` superan el 7,2 % durante
cinco minutos o el p95 supera 500 ms durante quince minutos. Dos alertas
`critical` cubren una tasa de `5xx` superior al 20 % durante dos minutos y una
API que Prometheus no puede monitorizar durante dos minutos.

Alertmanager agrupa por `alertname`, espera 30 segundos antes del primer aviso y
usa Mailpit solo en local. Repite `warning` cada cuatro horas y `critical` cada
diez minutos mientras sigan activas. Grafana aprovisiona el dashboard **SLO —
Refresh de sesión**, muestra las reglas de Prometheus y consulta Alertmanager
para alertas y silencios. No hay notificación remota, retención de producción
ni SLOs generales **en local**: véanse
[ADR-0098](../adr/0098-define-local-session-refresh-slo.md) y
[ADR-0099](../adr/0099-route-local-alerts-through-alertmanager.md).

## Espacio y retención de desarrollo

ADR-0139 mantiene las señales técnicas de `local` y `dev` durante un día:
Loki con borrado efectivo mediante compactor y Tempo a 24 horas; Prometheus
con 24 horas y 128 MB de retención TSDB. Los contenedores rotan la copia de
consola en dos archivos de 5 MB y comprimen el rotado. Ese límite es por tamaño,
no un TTL, y la eliminación de bloques de los backends es asíncrona.

Promtail excluye chequeos GET /healthz y GET /metrics correctos, conserva sus
fallos y recoge solo la API del proyecto correspondiente. No añade PII ni
etiquetas por petición. La configuración de producción y PostHog es independiente.

`make dev-observability-clean` y `make dev-public-observability-clean` retiran
solo volúmenes técnicos sin contenedores y sin archivos de las últimas 24 horas.
PostgreSQL, evidencia legal, backups, Grafana y Alertmanager quedan fuera.
Véase [LOG_RETENTION.md](LOG_RETENTION.md) para el procedimiento y los límites.

## Desarrollo público

`tournaments-manager-dev` replica el stack local —Prometheus, Alertmanager,
Loki, Tempo, Promtail y Grafana— con volúmenes y red propios. Reutiliza reglas,
dashboard y fuentes de datos versionadas; Promtail filtra explícitamente el
proyecto Compose `tournaments-manager-dev` para no mezclar los logs de local.
Solo `make dev-public-observability-up` lo activa bajo petición explícita;
`make dev-public-up`, despliegue y rollback arrancan API/PostgreSQL sin él.
La API exporta OTLP/HTTP a Tempo interno únicamente durante ese diagnóstico.

Alertmanager entrega los mismos avisos mediante Resend SMTP en
`smtp.resend.com:587`, con STARTTLS obligatorio. Usa una clave _Sending access_
distinta de `SMTP_PASSWORD`, montada desde el secreto local
`infra/dev/alertmanager.smtp-password`; el archivo no se versiona. El remitente
visible es `FastTourney Dev Alerts <alerts@mail.fasttourney.com>` y el asunto
empieza por `[DEV]`; el receptor inicial es `alerts@fasttourney.com`.

Grafana (`127.0.0.1:3001`), Prometheus (`127.0.0.1:9091`) y Alertmanager
(`127.0.0.1:9094`) no tienen ruta Caddy ni Cloudflare: una alerta se entrega por
correo, pero la operación detallada conserva acceso solo en el Mac. Prometheus
y Tempo retienen 24 horas; Prometheus añade un límite de retención TSDB de
128 MB. Loki también conserva un día con compactor activo. No hay HA, on-call ni
alertas nuevas por completitud.
Véase [ADR-0100](../adr/0100-deliver-public-development-alerts-through-resend.md).

## K3s de producción — desplegado privado, no publicado

El perfil aceptado para `prod` usa Helm solo para software de terceros, según
ADR-0112: Prometheus y Alertmanager, Loki monolítico, Tempo monolítico, Grafana
y Alloy. Alloy recoge exclusivamente logs de Pods de `prod` mediante la API de
Kubernetes y los entrega a Loki; no monta logs de host ni sustituye al Collector
de OpenTelemetry, que sigue aplazado. Los charts, versiones fijadas, PVC,
retención y límites están en [`infra/k3s/observability`](../../infra/k3s/observability/).

El 2026-10-03 se aplicó ADR-0140: siete días de diagnóstico, noventa de
seguridad, compactor persistente, rechazo únicamente de probes correctos y
retención transitoria de noventa días para histórico mezclado. Prometheus mantiene
un día o 128 MB TSDB. Alloy expone agregados de espacio mediante su collector
textfile; solo monta en lectura el directorio generado por un timer del host.
Las alertas cubren raíz, volúmenes técnicos, medición vencida y ciclos de retención
Loki ausentes. Son umbrales de intervención, no cuotas físicas de local-path.

El 2026-09-05 se instaló y validó el perfil en la VM. Prometheus hace scrape
estático de `api.prod.svc.cluster.local:8080` y no recibe token ni RBAC para
descubrimiento de Kubernetes: el coste es añadir manualmente cada target nuevo;
el beneficio aceptado es menor privilegio y carga en esta VM de un nodo.

Grafana, Prometheus, Loki, Tempo y Alertmanager siguen siendo privados. Caddy,
Cloudflare y el `503` de `api.fasttourney.com` permanecen fuera de este módulo.
El runbook registra el acceso por túnel SSH, la prueba de alerta y el rollback
por release.

## Validación y cierre de Fase 3

El 2026-08-21 se completaron los dos recorridos que responden preguntas
distintas:

- **Regla real local:** al detener solo PostgreSQL y mantener peticiones de
  refresh, la API respondió `500` con `database.query_failed` sin detalles
  internos. Prometheus activó `SessionRefreshFailureRateCritical`, Alertmanager
  entregó el aviso a Mailpit y envió su resolución tras recuperar PostgreSQL.
- **Canal externo de `dev`:** una alerta sintética `critical`, marcada
  `test=true`, fue aceptada por Alertmanager y Resend y llegó al buzón final a
  través de `alerts@fasttourney.com` y Cloudflare Email Routing. La prueba se
  resolvió sin detener API ni PostgreSQL de `dev`.

El secreto `infra/dev/alertmanager.smtp-password` contiene exclusivamente la
clave, en una sola línea. A diferencia de un archivo `.env`, sus comentarios no
se interpretan: formarían parte literal de la contraseña SMTP. La comprobación
operativa valida su presencia y forma sin imprimirlo.

Con esta evidencia se cumple el criterio de salida documentado en
[ROADMAP.md](../project/ROADMAP.md). Véase la
[retrospectiva técnica de Fase 3](../project/PHASE_3_RETROSPECTIVE.md).

## Recorridos revisados

| Ruta                                                                                                                 | Spans hijos relevantes                                                                                    | Decisión                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| -------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `POST /v1/registrations`                                                                                             | `auth.password.hash`, operaciones PostgreSQL de alta, `smtp.send.verification`                            | Instrumentar CPU costosa, transacción y dependencia SMTP; los rechazos usan `validation.rejected` o `rate_limit.exceeded`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| `POST /v1/password-resets`                                                                                           | `postgresql.CreatePasswordReset`, `smtp.send.password_reset` si existe una cuenta elegible                | No revelar la existencia de la cuenta en los atributos ni crear un span cuando no se envía correo; rechazos: `validation.rejected` y `rate_limit.exceeded`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| `POST /v1/password-reset-links`                                                                                      | `postgresql.InspectPasswordReset`                                                                         | Los rechazos usan `validation.rejected` o `credential.reset_link_invalid`, sin exponer el token.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| `POST /v1/password-reset-confirmations`                                                                              | `auth.password.hash`, `postgresql.ConsumePasswordReset`                                                   | La actualización de credencial, revocación y sesión es una sola consulta atómica; rechazos: `validation.rejected` o `credential.reset_link_invalid`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| `POST /v1/registration-verifications`                                                                                | `postgresql.VerifyRegistrationAndCreateSession`                                                           | No añadir spans para SHA-256, aleatoriedad o CTE internos; rechazos: `validation.rejected` o `credential.verification_link_invalid`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| `POST /v1/google-login-challenges`                                                                                   | `postgresql.CreateGoogleLoginChallenge`                                                                   | El nonce y el identificador opacos no salen como atributos; el único límite relevante es PostgreSQL y sus categorías técnicas cerradas. Si Google no está configurado, la salida `503` es `identity.google_unavailable`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| `POST /v1/google-sessions`                                                                                           | Operaciones PostgreSQL de challenge, identidad, torneo opcional idempotente y sesión                      | La prueba OIDC, token, nonce, email, subject, `draftId`, borrador y sesión quedan fuera de atributos. Torneo, equipos y sesión se confirman en la misma transacción; un `draftId` repetido para la cuenta reutiliza el torneo. Los rechazos usan `validation.rejected`, `credential.google_challenge_invalid` o `identity.email_conflict`; un fallo técnico no clasificable usa `request.failed`. El alta pendiente (`202`) no es un fallo; Google no configurado usa `identity.google_unavailable`.                                                                                                                                                                                                                                  |
| `POST /v1/me/google-identities`                                                                                      | Operaciones PostgreSQL de ticket, challenge e identidad                                                   | La sesión, ticket, prueba Google e identidad no se exportan; rechazos: `validation.rejected`, `credential.reauthentication_invalid` e `identity.google_conflict`. La autenticación y CSRF se resuelven en middleware sin spans adicionales; `503` por Google no configurado es `identity.google_unavailable`.                                                                                                                                                                                                                                                                                                                                                                                                                         |
| `DELETE /v1/me/google-identities`                                                                                    | `postgresql.ConsumeReauthenticationTicketAndRemoveGoogle`                                                 | La eliminación es una transición atómica que conserva la credencial local; el ticket no se registra. Rechazos: `validation.rejected` o `credential.reauthentication_invalid`; errores de PostgreSQL usan la categoría técnica cerrada. Google no configurado usa `identity.google_unavailable`.                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| `POST /v1/risc/events`                                                                                               | Configuración RISC/JWKS de Google y transacción de deduplicación + revocación                             | El SET, `sub`, `jti`, audiencia y cuerpo no salen del proceso. Un SET inválido usa `credential.risc_invalid`; RISC/JWKS no disponible usa `identity.google_unavailable` y PostgreSQL `database.query_failed`. El `202` se entrega solo al validar y aplicar o reconocer la revocación idempotente, sin spans hijos por JWT o consulta de claves.                                                                                                                                                                                                                                                                                                                                                                                      |
| `POST /v1/sessions`                                                                                                  | `postgresql.FindLocalAccountForLogin`, `auth.password.verify`, transacción de torneo idempotente y sesión | Torneo, equipos y sesión se confirman juntos; un `draftId` repetido para la cuenta no duplica el torneo y nunca se exporta como atributo. Rechazos: `validation.rejected`, `rate_limit.exceeded` o `authentication.credentials_rejected`, sin distinguir cuenta, contraseña, borrador o estado; un fallo técnico no clasificable usa `request.failed`.                                                                                                                                                                                                                                                                                                                                                                                |
| `GET /v1/sessions`                                                                                                   | `postgresql.GetCurrentSession`                                                                            | Una sesión que deja de ser válida después de autenticarse se marca como `session.invalid`; la autenticación previa no añade un span ni atributos de token.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| `POST /v1/sessions/refresh`                                                                                          | `postgresql.RotateSessionTokens`                                                                          | La rotación del token opaco y la sesión se diagnostican como una única transición atómica. Un refresh ausente, duplicado o ya consumido se marca como `session.refresh_invalid`, sin exportar el token.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| `DELETE /v1/sessions`                                                                                                | `postgresql.RevokeSession`                                                                                | La revocación es idempotente. Los rechazos de autenticación se resuelven en el middleware sin crear spans adicionales ni registrar la credencial.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| `GET /v1/me/access-methods`                                                                                          | `postgresql.GetAccessMethods`                                                                             | La consulta no añade razones de negocio: email, username y métodos forman parte de la respuesta pero no de atributos de traza.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| `POST /v1/me/reauthentication-tickets`                                                                               | `auth.password.verify`, `postgresql.GetCurrentPasswordHash`, `postgresql.CreateReauthenticationTicket`    | La validación usa `validation.rejected`; una identidad federada de otra cuenta es `reauthentication.identity_conflict` y un desafío, contraseña o sesión no válidos es `reauthentication.invalid`. No se incluyen challenge, ticket, identidad, contraseña ni token.                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| `PUT /v1/me/local-credential`                                                                                        | `auth.password.hash`, `postgresql.ConsumeReauthenticationTicketAndSetPassword`                            | Los rechazos se limitan a `validation.rejected` y `reauthentication.invalid`; no se crean spans para SHA-256 del ticket ni aleatoriedad.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| `DELETE /v1/me/local-credential`                                                                                     | `postgresql.ConsumeReauthenticationTicketAndRemovePassword`                                               | Un ticket inválido es `reauthentication.invalid`; intentar dejar la cuenta sin método de acceso es `access_method.last_remaining`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| `DELETE /v1/me/account`                                                                                              | `postgresql.ScheduleAccountDeletion`                                                                      | Una cuenta con ligas propias no puede programar su borrado y se marca como `account.deletion_owned_leagues`; el ID y la fecha efectiva no son atributos.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| `GET /v1/usernames/{username}/availability`                                                                          | `postgresql.IsUsernameAvailable`                                                                          | La plantilla de ruta y el nombre de consulta son estáticos; username e IP quedan fuera de los atributos; rechazos: `validation.rejected` y `rate_limit.exceeded`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| `GET /v1/users`                                                                                                      | `postgresql.SearchPublicUsernames`                                                                        | La búsqueda conserva la query string fuera de HTTP y el texto buscado fuera de PostgreSQL; rechazos: `validation.rejected` y `rate_limit.exceeded`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| `GET /v1/me/tournaments`, `GET /v1/me/recent-tournaments`                                                            | Consultas PostgreSQL estáticas de colecciones                                                             | Éxito devuelve las proyecciones; en la colección paginada el cursor corresponde al último ID entregado, validado con recorridos PostgreSQL sin omisiones ni duplicados. La fila adicional no crea spans ni atributos. Filtros, cursor y límite inválidos usan `validation.rejected`. No hay limitador específico. Un fallo del límite PostgreSQL o cancelación usa `database.*`, `request.cancelled` o `request.timeout` en el span HTTP; no se registran cuenta, cursor ni filtros.                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| `POST /v1/me/suggestions`                                                                                            | `postgresql.CreateProductSuggestion`; `smtp.send.suggestion` después del guardado                         | El éxito depende solo del insert. Entrada inválida usa `validation.rejected`; el cuarto envío por cuenta y hora usa `rate_limit.exceeded`; PostgreSQL usa las categorías técnicas comunes. Un fallo SMTP queda en su span hijo y no cambia el `201`. Texto, username, cuenta, destinatario e identificador quedan fuera de logs, trazas y analítica.                                                                                                                                                                                                                                                                                                                                                                                  |
| `PUT` \| `DELETE /v1/me/tournaments/{tournamentId}/follow`                                                           | `postgresql.FollowVisibleTournament`, `postgresql.UnfollowTournament`                                     | Éxito es idempotente. ID inválido: `validation.rejected`; torneo no visible: `tournament.not_found`. No hay límite de tasa propio ni spans por las ramas de seguimiento.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| `POST /v1/tournaments`, equipos, inicio, cancelación, resultado y finalización; `GET /v1/tournaments/{tournamentId}` | Operaciones PostgreSQL estáticas del ciclo del torneo                                                     | La entrada inválida —incluido deporte ausente/desconocido, `bestOfSets` incompatible, tanteo de set ausente o nulo, set imposible o posterior a la victoria, tanteo de baloncesto empatado o desempate incompatible de fútbol o balonmano— usa `validation.rejected`; los rechazos de negocio distinguen `tournament.forbidden`, `tournament.not_found`, y conflictos cerrados de inicio, equipos, retirada, resultado, cancelación o finalización. Deporte, configuración, sets, juegos de tenis de mesa, parciales de voleibol, tanteos e IDs no se exportan como atributos. Sin limitador específico. Los fallos técnicos y cancelaciones conservan las categorías seguras comunes en el span raíz.                                                                                                                                                                                    |
| Inicio mixto y `POST /v1/tournaments/{tournamentId}/stages/elimination/start`                                        | Transacción PostgreSQL de composición, clasificación congelada, desempate condicional y generación del cuadro | El éxito crea la liga, abre atómicamente los desempates exactos del corte o, cuando todas las plazas están resueltas, inicia el cuadro. JSON o parámetros inválidos usan `validation.rejected`; una composición o plantilla incompatible usa `tournament.configuration_rejected`; una transición prematura, repetida o sin suficientes equipos no retirados usa `tournament.stage_transition_conflict`; autorización y ausencia conservan `tournament.forbidden` y `tournament.not_found`. Un resultado de desempate sin ganador usa `validation.rejected`, igual que cualquier resultado eliminatorio inválido. No se exportan equipos, grupos, ciclos, cantidades, resultados, retiradas ni IDs. No hay límite de tasa específico. Cada fallo técnico usa `database.*` o `request.failed`, el timeout `request.timeout` y la cancelación `request.cancelled`, sin error bruto y sin spans por validación, clasificación, desempate, CTE o generación de IDs. |
| Invitaciones de equipo: crear, revocar, inspeccionar e inscribir                                                     | Operaciones PostgreSQL estáticas sobre invitación, equipo, vínculo y seguimiento                          | Éxito crea o revoca la capacidad, devuelve su proyección segura o confirma equipo y seguimiento. Entrada inválida usa `validation.rejected`; enlace desconocido, rotado o cerrado usa `tournament.invitation_not_found`; cuenta repetida, nombre duplicado o aforo completo usa `tournament.invitation_conflict`; autorización de la organizadora conserva `tournament.forbidden`. Token, hash, nombre, cuenta e IDs no se exportan. No hay span por aleatoriedad o SHA-256 ni limitador específico; fallos técnicos y cancelaciones usan las categorías comunes.                                                                                                                                                                     |
| Administración de torneos: listar, asignar, eliminar y transferir                                                    | Operaciones PostgreSQL estáticas de administración                                                        | Éxito devuelve o modifica únicamente la proyección contractual. Entrada inválida: `validation.rejected`; negocio: `tournament.forbidden`, `tournament.not_found`, `tournament.administrator_conflict` o `tournament.ownership_transfer_conflict`. Nombres de usuario e IDs quedan fuera de atributos, logs y nombres de spans; no existe un límite de tasa específico.                                                                                                                                                                                                                                                                                                                                                                |
| Bandeja de notificaciones: lista, contador, marcar leídas y borrar una o todas                                       | Consultas y mutaciones PostgreSQL estáticas de `account_notifications`                                    | Las mutaciones son idempotentes y no añaden un rechazo de negocio. Sus fallos técnicos y cancelaciones usan `database.*`, `request.cancelled` o `request.timeout` en el span HTTP, sin cuenta, notificación ni error bruto.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |

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

No se añaden paneles, alertas, SLO, retenciones ni perfiles por completitud. El
SLO de refresh existente responde a una pregunta operativa concreta; la unidad
mínima sigue siendo una pregunta respondida de extremo a extremo y validada
provocando un fallo. La desviación de retención de seguridad se registra en
[LOG_RETENTION.md](LOG_RETENTION.md), no se disfraza ampliando señales sin un
destino adecuado.

### Recorrido de resultados por sets (ADR-0136 y ADR-0137)

- Éxito: `200` conserva un único span HTTP con plantilla y el límite PostgreSQL;
  corrección, retirada y avance mantienen historial atómico. Sin spans por set,
  validación, rama ni comparación de cocientes.
- Rechazo de entrada: `400` por campos ausentes/nulos, configuración o secuencia
  inválidas; `validation.rejected` cerrado. Pruebas de forma HTTP y de dominio
  comprueban los límites; cero explícito sigue siendo válido.
- No hay limitador específico en estos endpoints. Un eventual límite común
  conserva `rate_limit.exceeded`; no se inventa otro flujo de feature.
- Sesión/autorización, ausencia y estado: `401/403/404/409` seguros, causas
  cerradas existentes. El `409` de resultado cubre la fase congelada, dependencia
  del cuadro y rechazo de sobrescribir un resultado administrativo de voleibol.
- Límites técnicos: adquirir conexión, consulta y transacción/commit conservan
  las categorías seguras de `RecordDatabaseEndpointFailure`, sin error bruto.
  El `500` no expone cuerpos internos; cliente no mapea estos estados a negocio.
- Cancelación: misma categoría segura común en raíz; una cancelación intencional
  del cliente no muestra feedback. Sin diagnósticos con deporte, sets, ratios,
  inputs, IDs o PII.

Validación: forma de payload con span seguro; secuencias de sets y cocientes
exactos en dominio; integración PostgreSQL para corrección, retirada, campeón
por tantos, grupos, desempate repetido y final. Se reutiliza el mapeo cerrado del
adaptador cliente de resultados y las pruebas existentes de fallbacks.

### Recorrido de bádminton con perfil 15/21 (ADR-0141)

- Éxito: creación `201`, lectura y resultado `200`; configuración persistida,
  historial atómico y ganadora derivados. Se conservan span HTTP de plantilla y
  límites PostgreSQL, sin span por perfil, juego, validación o rama.
- Validación `400`: perfil ausente/ajeno, mejor de tres inválido, secuencia o
  tope inválidos; `validation.rejected` en raíz, cerrado y sin valores de input.
- Tasa: no hay limitador específico de creación/resultado; los límites de
  autenticación al transferir borradores mantienen `rate_limit.exceeded`.
- Negocio: `401/403/404/409` seguros y causas existentes de acceso, ausencia,
  transición o dependencia del cuadro. No se crean causas por puntuación.
- Fallos de adquisición de conexión, consulta y transacción/commit conservan
  las categorías de `RecordDatabaseEndpointFailure`; `500` seguro sin error
  bruto. No se añaden mappings cliente por estado técnico.
- Cancelación: categoría segura común en raíz, sin feedback si el cliente
  cancela intencionadamente; sin exportar puntos, nombres, IDs, secretos o PII.

Validación añadida: HTTP rechaza configuración inválida con span seguro, creación
y borradores conservan el perfil; dominio cubre ambos topes; integración
PostgreSQL desechable verifica creación, lectura, restricciones, correcciones,
historial, dependencia y campeón. Se reutilizan las pruebas existentes de
fallos técnicos, cancelación y feedback común, sin nuevos límites técnicos.

### Recorrido de incidencias por partido (ADR-0142)

- Éxito `200`: resultado administrativo, parcial separado e historial atómico;
  se conserva el span HTTP con plantilla de ruta y los límites PostgreSQL.
- Validación `400`: variante mezclada, tipo/lado desconocido, campos no
  admitidos o parcial imposible; `validation.rejected` seguro en raíz. No se
  añade span por motivo, set, validación o rama.
- Tasa: la operación no tiene un limitador específico; no se inventa una
  salida `429`. Los límites de autenticación conservan su recorrido vigente.
- Negocio `401/403/404/409`: permisos, ausencia, fase congelada, dependencia o
  resultado administrativo bloqueado; se mantienen las causas cerradas
  existentes, sin inferir reglas centrales por estado HTTP.
- Límites técnicos: adquisición de conexión, consulta y transacción/commit
  mantienen las categorías de `RecordDatabaseEndpointFailure`; respuesta
  segura `500`, sin exportar error bruto, tanteos, nombres, IDs o PII.
- Cancelación: conserva la categoría común segura en raíz; el cliente no
  muestra feedback cuando la cancelación es intencional.

HTTP comprueba formas cerradas, feedback y razón de validación; dominio cubre
los ocho deportes; PostgreSQL desechable cubre correcciones, historial, lectura,
clasificación, dependencias y ciclos de desempate. Los límites técnicos y
cancelación reutilizan las pruebas existentes, sin nuevos límites ni causas.


### Recorrido de terminación de la API — 2026-10-03

ADR-0145 no modifica las salidas de negocio, validación o tasa de los endpoints.
El éxito HTTP en vuelo puede concluir durante diez segundos de drenaje. Si vence,
Server.Close cancela las peticiones restantes: los límites técnicos mantienen sus
categorías seguras existentes (`request.cancelled`, `request.timeout`, `database.*`),
sin exportar errores brutos ni añadir spans por cierre. Los logs de ciclo de vida
son mensajes cerrados sin inputs: inicio, finalización y plazo agotado. El cierre
de trazas tiene un máximo independiente de cinco segundos después del pool.

`cmd/api/shutdown_test.go` valida éxito en vuelo, listener cerrado y cancelación
forzada con sockets reales. Las pruebas existentes de HTTP conservan el contrato
de categorías seguras. La propagación de endpoints, pérdida de respuestas en el
borde y terminación por SIGKILL requieren el recorrido de rollout del runbook;
no quedan demostradas por los logs de drenaje ni por un rollout exitoso.

## Apagado fuera de pruebas — ADR-0146

`make dev-observability-down` y `make dev-public-observability-down` desactivan
el exportador y detienen los seis servicios técnicos conservando volúmenes. No
levantan una API que ya esté apagada. `make dev-down` y `make dev-public-down`
apagan todo el carril, incluidos perfiles técnicos; público suspende además sus
LaunchAgents. Ningún servicio dev reinicia automáticamente con Docker Desktop.
Mientras la observabilidad está apagada no hay histórico central ni alertas dev;
la API conserva su consola JSON segura y métricas internas durante las pruebas.
La retención de 24 h se aplica cuando los servicios técnicos están activos; la
purga no avanza con Loki/Tempo apagados. Producción conserva su funcionamiento.

### Recorridos sociales revisados — ADR-0147

| Endpoint | Éxito y negocio | Rechazos y límites técnicos seguros |
| --- | --- | --- |
| `POST /v1/apple-login-challenges` | 201; solo digests en persistencia | 400 `validation.rejected`/`credential.apple_challenge_invalid`; 429 `rate_limit.exceeded`; configuración ausente 503 `identity.apple_unavailable`; PostgreSQL usa categorías técnicas existentes. |
| `POST /v1/apple-callback` | 303 `ready`; denegación 303 `cancelled` sin causa de fallo ni sesión | State inválido 400 `credential.apple_challenge_invalid`; formulario 400 `validation.rejected`; 429 `rate_limit.exceeded`; intercambio/JWKS 303 `failed` con `identity.provider_unavailable`; token/nonce rechazado con `credential.apple_challenge_invalid`; fallos DB con categorías existentes. Nunca se redirige sin state reclamado. |
| `POST /v1/apple-sessions` | 200; 202 pide alta y no consume prueba; 409 `identity.email_conflict` | 400 validación/prueba; 429 tasa; configuración ausente 503; fallo DB 500 seguro. La transacción abarca prueba, cuenta, evidencia legal, borrador y sesión. |
| `POST /v1/google-login-challenges`, `POST /v1/google-sessions` | Conservan sus recorridos | Se aplica el 429 declarado con `rate_limit.exceeded`; no exporta IP. El contrato challenge añade 500 real de persistencia. |
| `GET /v1/me/access-methods` | Incluye presencia booleana de Apple | No exporta issuer, subject ni email como atributos; conserva autenticación y fallos técnicos anteriores. |

`identity.provider_unavailable` es la categoría técnica cerrada del límite HTTP
Apple (intercambio de código y claves públicas): no contiene el error remoto.
Los spans raíz conservan las plantillas de ruta; no hay spans por hash, rama,
aleatoriedad o JWT. Ninguna salida exporta state, nonce, proof, código, tokens,
email, subject, clave privada ni cuerpos Apple. Los callbacks y las sesiones usan
no-store. Cancelar una petición no escribe feedback; cancelar en Apple vuelve con
un estado cerrado, sin fallo. Pruebas HTTP verifican todos estos recorridos y
atributos seguros; la integración comprueba consumo concurrente y rollback de 202.

## Revisión de rotación de enlaces — 2026-10-04

Se revisan `POST /v1/sessions` y `POST /v1/password-resets` al corregir las
colisiones de tokens activos. Los locks de cuenta y las consultas de rotación
son límites PostgreSQL; no se añaden spans de validación, CTE, hash ni generación
de secreto. No cambian las categorías ni se exportan email, cuenta, token o SQL.

| Salida                                                                  | Evidencia y tratamiento                                                                                                                                                                                                                                                           |
| ----------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Login verificado 200 / pendiente 202                                    | HTTP real; el pendiente rota email y conserva cero sesiones en las regresiones de persistencia.                                                                                                                                                                                   |
| Recuperación elegible/desconocida 202                                   | Cuerpo vacío idéntico; la cuenta desconocida no invoca SMTP. Prueba HTTP contractual y recorrido Mailpit local.                                                                                                                                                                   |
| Validación 400 / tasa 429                                               | Pruebas HTTP; `validation.rejected` / `rate_limit.exceeded`; 429 conserva Retry-After.                                                                                                                                                                                            |
| Credenciales de login rechazadas 401                                    | Recorrido HTTP y suite `access_observability_test.go`; `authentication.credentials_rejected`, sin distinguir cuenta o contraseña. Recuperación no tiene rechazo de negocio revelador.                                                                                             |
| Adquisición de conexión, lock, invalidación/inserción o commit fallidos | El span de PostgreSQL conserva su categoría `database.*`; la raíz de estos recorridos usa `request.failed` como diagnóstico común sin exportar el error. Respuesta 500 segura. Las regresiones de persistencia prueban la transición y la cancelación sin dejar tokens parciales. |
| SMTP fallido                                                            | El span SMTP conserva `smtp.delivery_failed`; la raíz usa `request.failed`. Respuesta 500 sin mensaje del proveedor. Prueba HTTP con un error privado centinela.                                                                                                                  |
| Timeout o cancelación de PostgreSQL/SMTP                                | Categorías comunes `request.timeout` / `request.cancelled` en el span raíz y en los límites instrumentados; pruebas HTTP comprueban que sus errores no aparecen en el cuerpo. El cliente no muestra feedback por una cancelación intencionada.                                    |

`password_recovery_contract_test.go` comprueba las diez salidas de recuperación:
éxito elegible/desconocido, validación, tasa y fallo/cancelación/timeout de cada
una de sus dos dependencias. Las pruebas no activan exporters ni observabilidad
local/dev. No se afirma haber inducido cada fallo técnico en un servidor vivo:
esa evidencia procede de inyección de dependencias y de la suite de observabilidad.

Aprendizaje: revisar únicamente el primer envío ocultaba un 500 al repetirlo.
Probar transiciones consecutivas y simultáneas descubre orden de CTE y relojes
de transacción que una prueba aislada de éxito no ejercita.
