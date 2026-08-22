# Diagnóstico del refresh de sesión

> Estado: validado en entorno Compose local el 2026-08-21.
>
> Alcance: diagnóstico seguro del entorno Compose local y, cuando se indique,
> de `tournaments-manager-dev`; no autoriza cambios de datos.

## Síntoma

Una persona pierde la sesión, el cliente recibe un error al renovar la sesión o
la operación tarda más de lo esperado.

## Prerrequisitos

- contratos locales preparados con `make dev-init`;
- esquema inicial aplicado y una sesión web válida creada;
- `make dev-up` en ejecución;
- Grafana disponible en `http://127.0.0.1:3000`.

## Diagnóstico seguro

1. En Grafana, abre **Explore > Prometheus** y consulta:

   ```promql
   sum by (status) (rate(tournaments_manager_http_requests_total{route="POST /v1/sessions/refresh"}[5m]))
   ```

2. Para latencia, consulta el percentil 95:

   ```promql
   histogram_quantile(0.95, sum by (le) (rate(tournaments_manager_http_request_duration_seconds_bucket{route="POST /v1/sessions/refresh"}[5m])))
   ```

3. En **Explore > Tempo**, busca el servicio `tournaments-manager-api` y la
   operación `HTTP server`. La traza correcta contiene el span hijo
   `postgresql.query`.

4. Desde la traza, abre los logs correlacionados o, en **Explore > Loki**, busca
   `{service="tournaments-manager-api"} |= "HTTP request completed"`. Abre el
   JSON y usa `trace_id` para llegar a Tempo.

Los tres pasos muestran datos técnicos. No copies cookies, cabeceras
`Authorization`, cuerpos ni tokens a búsquedas, tickets o logs.

## Alertas y entrega

Grafana concentra las dos partes de la operación:

1. Abre **Alerting > Alert rules** para comprobar las seis reglas cargadas por
   Prometheus: cuatro alertas evaluables y dos series de grabación. Son de solo
   lectura porque su fuente es `infra/observability/prometheus-rules.yml`.
2. Abre **Alerting > Active notifications** y, en **Choose Alertmanager**,
   selecciona `Alertmanager`. Ahí se ven las alertas activas, su agrupación y el
   receptor local (`mailpit-warning` o `mailpit-critical`).
3. Para detener temporalmente un aviso conocido, usa **Alerting > Silences** y
   el mismo selector `Alertmanager`. Crea un silencio con alcance y caducidad
   explícitos; no modifiques las rutas ni los receptores desde Grafana.

La configuración declarativa sigue siendo
`infra/observability/alertmanager.yml`: espera 30 segundos antes del primer
aviso, agrupa por `alertname` y repite los `warning` cada cuatro horas y los
`critical` cada diez minutos. Mailpit permite inspeccionar el correo local en
`http://127.0.0.1:8025`; no entrega correo externo.

## Desarrollo público

Para `tournaments-manager-dev`, aplica el mismo diagnóstico desde Grafana en
`http://127.0.0.1:3001`. Prometheus y Alertmanager están en loopback en los
puertos `9091` y `9094`; no se consultan desde el túnel. El receptor de las
alertas es Resend SMTP con STARTTLS y una clave separada en
`infra/dev/alertmanager.smtp-password`.

Antes de probar una alerta, confirma que el secreto existe, contiene solo la
clave `re_...` en una línea, no conserva comentarios ni el texto de ejemplo y que
`make dev-public-deploy` ha terminado. Provoca la misma caída controlada de
PostgreSQL solo si `dev` no tiene usuarios que puedan verse afectados; comprueba
el aviso recibido y su resolución tras recuperar la base. No copies la clave SMTP
ni contenido de correos a Grafana, tickets o logs.

## Prueba de fallo controlada

1. Con una sesión web válida, abre el cliente y fuerza una renovación de sesión
   o espera a que el cliente la solicite.
2. Detén solo PostgreSQL: `docker compose --env-file infra/local/.env -f infra/local/compose.dev.yaml stop postgres`.
3. Repite el refresh y verifica una respuesta segura de error genérico; el
   cliente no debe exponer detalles de PostgreSQL.
4. En Grafana confirma: incremento de errores HTTP, traza con fallo de
   `postgresql.query` y log correlacionado sin secretos ni datos personales.
   El span raíz debe incluir una razón cerrada `database.*`; si el borde no
   conoce el límite concreto, usa `request.failed`.
5. Recupera PostgreSQL: `docker compose --env-file infra/local/.env -f infra/local/compose.dev.yaml start postgres`.
6. Repite el refresh y confirma la recuperación mediante las tres señales.

## Riesgos y recuperación

Detener PostgreSQL interrumpe temporalmente las operaciones de la API local,
pero no elimina el volumen. No ejecutes `down --volumes`, `db-reset` ni ningún
comando de reinicialización durante esta prueba.

Si la API no se recupera después de que PostgreSQL esté saludable, consulta
[PostgreSQL local con Docker Compose](local-postgresql.md). Si Tempo, Loki o
Prometheus no están disponibles, la API debe seguir funcionando; revisa sus
logs de contenedor y vuelve a iniciar `make dev-up`.

## Verificación desde la persona usuaria

Tras recuperar PostgreSQL, una renovación de sesión vuelve a mantener el acceso
sin mostrar detalles internos ni requerir un inicio de sesión nuevo.

## Última prueba

El 2026-08-21 se creó una cuenta de prueba exclusivamente local, se obtuvo una
sesión bearer, se detuvo únicamente `postgres` y se renovó la sesión. La API
respondió `500`, el log JSON correlacionado registró
`failure_reason=database.query_failed` y Tempo confirmó la misma causa en el
span HTTP raíz. PostgreSQL se inició de nuevo antes de cerrar la prueba.
