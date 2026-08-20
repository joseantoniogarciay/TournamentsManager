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
`request.timeout`. Cada feature puede añadir una causa de negocio segura cuando
aporte recuperación distinta; no se deduce centralmente del código HTTP.

`OTEL_TRACES_ENDPOINT` es opcional. Cuando falta o Tempo deja de estar
disponible, la API mantiene los logs JSON y las métricas y no deja de servir
peticiones por un error de exportación.

El procedimiento de diagnóstico y la prueba de indisponibilidad controlada de
PostgreSQL están en el [runbook de refresh de sesión](../runbooks/session-refresh-observability.md).

## Recorridos revisados

| Ruta | Spans hijos relevantes | Decisión |
| --- | --- | --- |
| `POST /v1/registrations` | `auth.password.hash`, operaciones PostgreSQL de alta, `smtp.send.verification` | Instrumentar CPU costosa, transacción y dependencia SMTP. |
| `POST /v1/password-resets` | `postgresql.CreatePasswordReset`, `smtp.send.password_reset` si existe una cuenta elegible | No revelar la existencia de la cuenta en los atributos ni crear un span cuando no se envía correo. |
| `POST /v1/password-reset-confirmations` | `auth.password.hash`, `postgresql.ConsumePasswordReset` | La actualización de credencial, revocación y sesión es una sola consulta atómica. |
| `POST /v1/registration-verifications` | `postgresql.VerifyRegistrationAndCreateSession` | No añadir spans para SHA-256, aleatoriedad o CTE internos: no son límites operativos independientes. |
| `POST /v1/sessions/refresh` | `postgresql.RotateSessionTokens` | La rotación del token opaco y la sesión se diagnostican como una única transición atómica; el runbook conserva la investigación extremo a extremo. |
| `GET /v1/usernames/{username}/availability` | `postgresql.IsUsernameAvailable` | La plantilla de ruta y el nombre de consulta son estáticos; username e IP quedan fuera de los atributos. |
| `GET /v1/users` | `postgresql.SearchPublicUsernames` | La búsqueda conserva la query string fuera de HTTP y el texto buscado fuera de PostgreSQL. |

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
completitud. La unidad mínima es una pregunta operativa respondida de extremo a
extremo y validada provocando un fallo.
