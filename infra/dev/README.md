# Entorno público de desarrollo

Este directorio define el proyecto Compose `tournaments-manager-dev`. Sus
servicios internos se llaman `api` y `postgres`; Docker añade el
namespace del proyecto, de modo que no colisionan con
`tournaments-manager-local` ni con el futuro `tournaments-manager-prod`.

La API usa el target `runtime` del Dockerfile: no incluye Air, fuentes ni bind
mounts. Solo se publica en `127.0.0.1:8081`, donde Caddy la alcanza desde el
Mac. PostgreSQL no publica ningún puerto. El correo se entrega mediante Resend
SMTP; Mailpit se mantiene solo en el proyecto `tournaments-manager-local` para
inspección y no tiene ruta Cloudflare/Caddy.

La red del proyecto fija `172.19.0.0/16` para que la API pueda reconocer la
gateway `172.19.0.1` como el único proxy de confianza. Caddy traduce
`CF-Connecting-IP` a `X-Client-IP`; la API usa esta última exclusivamente desde
esa gateway para sus límites de abuso.

Los servicios usan `restart: unless-stopped`. Cuando Docker Desktop se inicia
al abrir sesión, recupera el proyecto que estuviera en marcha antes de apagar el
Mac. Una parada explícita con `docker compose stop` conserva esa intención y no
lo reinicia automáticamente.

## Primer arranque

```bash
make dev-public-init
# Edita infra/dev/.env con tres contraseñas distintas. Copia exactamente
# POSTGRES_APP_PASSWORD en DATABASE_URL de infra/dev/api.docker.env.
make dev-public-bootstrap
make dev-public-up
```

`dev-public-bootstrap` se ejecuta solo sobre una base vacía. Crea el propietario
sin login, migrador y runtime; aplica el esquema mediante el migrador después de
que asuma temporalmente el propietario, y concede finalmente el acceso DML de
la API, comprobando además que esa identidad no puede crear tablas. Un cambio
futuro de esquema requiere su `GRANT` de runtime explícito; no se conceden
permisos por defecto.

Para reiniciar únicamente los datos de este entorno antes del bootstrap:

```bash
make dev-public-reset
# confirma escribiendo DEV_RESET
make dev-public-bootstrap
```

## Actualización y reversión de dev

Tras pasar `make verify` y la CI del commit de `develop`, el despliegue manual
es:

```bash
make dev-public-deploy
```

El primer despliegue tras adoptar ADR-0097 exige antes editar los dos contratos
de entorno con las credenciales nuevas y ejecutar `make dev-public-reset`,
`make dev-public-bootstrap`. Después, `make dev-public-deploy` recrea la API
con `POSTGRES_APP_USER`; no recibe ni la contraseña de migración ni la de
propiedad.

El comando exige un árbol limpio cuyo `HEAD` coincida exactamente con
`origin/develop`. Construye una imagen `runtime` etiquetada con el SHA completo,
exporta la web estática en un directorio por SHA y publica ambas versiones. La
web cambia mediante el enlace simbólico `current`, por lo que no mezcla archivos
de dos exports. Se conservan el despliegue actual y el anterior bajo
`/opt/homebrew/var/www/fasttourney/dev/releases/`; cada uno contiene un
`deployment.json` con SHA, imagen y fecha. Git conserva el código, no estos
artefactos locales.

Antes de construir o conmutar un release, el script de despliegue compara sin
mostrar secretos `DATABASE_URL` con la URL esperada para `POSTGRES_APP_USER`.
Un usuario administrador o migrador en el contrato de la API aborta el
despliegue.

Para recuperar uno de los dos despliegues conservados:

```bash
make dev-public-rollback SHA=<SHA-completo>
```

La reversión recoloca API y web en el mismo SHA; no modifica PostgreSQL. Por
tanto no sustituye una estrategia de backup/restauración cuando los datos de dev
dejen de ser descartables.

No hay GitHub Releases para los pushes normales de `develop`. Los tags y GitHub
Releases quedan reservados para producción o hitos distribuidos.

La web se exporta como estática con `infra/home/deploy-dev-web.sh`; no se usa
`expo start --web` como servidor público. Antes de invitar usuarios que deban
recibir emails, verifica `mail.fasttourney.com` en Resend y configura sus
SPF/DKIM/DMARC. Crea una API key *Sending access* restringida a ese dominio y
cópiala solo en `infra/dev/api.docker.env` como `SMTP_PASSWORD`; el usuario SMTP
es `resend` y el endpoint es `smtp.resend.com:587` con STARTTLS. No actives
invitaciones si el dominio aún figura como pendiente o si la clave ha aparecido
en una terminal, un log o Git: revócala y crea otra.

## Observabilidad y alertas

`dev` ejecuta el mismo stack correlacionado que local, pero con volúmenes propios.
Antes del primer `make dev-public-deploy`, crea en Resend una segunda API key de
tipo *Sending access*, restringida al dominio remitente y exclusiva para
Alertmanager. No reutilices `SMTP_PASSWORD`, que sigue reservado al correo
transaccional de la API.

```bash
cp infra/dev/alertmanager.smtp-password.example infra/dev/alertmanager.smtp-password
# Edita el archivo y deja únicamente la nueva clave de Resend.
make dev-public-deploy
```

El archivo real está ignorado por Git y se monta como secreto de Docker, no como
variable de entorno. Alertmanager conecta a `smtp.resend.com:587` con STARTTLS
antes de autenticar. El remitente visible es **FastTourney Dev Alerts** y el
asunto empieza por `[DEV]`; el receptor inicial es `alerts@fasttourney.com`.
Registro y restablecimiento usan **FastTourney Dev** y el mismo prefijo. Ambos
buzones necesitan ser válidos en Resend y en tu correo. Si cambia el receptor,
actualiza la configuración versionada
`infra/observability/alertmanager.dev.yml` y la documentación asociada.

Grafana, Prometheus y Alertmanager se consultan solo desde el Mac en
`http://127.0.0.1:3001`, `http://127.0.0.1:9091` y `http://127.0.0.1:9094`.
No los añadas a Caddy, Cloudflare Tunnel ni a la web pública. El procedimiento de
diagnóstico y prueba controlada está en el runbook de observabilidad.

## Purga de cuentas con baja vencida

`make dev-public-purge` ejecuta dentro del contenedor `api` el comando interno
`/api purge-expired-accounts`. Solo elimina hasta 100 cuentas con estado
`deletion_pending` cuya solicitud tiene al menos 30 días; no existe una ruta HTTP
para invocarlo. Las cuentas con ligas propias no llegan a ese estado. Si una
cuenta conserva autoría en el historial de resultados, la purga mantiene el
resultado y deja esa autoría a `NULL`.

El Mac lo programa mediante la plantilla
[`infra/home/launchd/com.fasttourney.dev-account-purge.plist.template`](../home/launchd/com.fasttourney.dev-account-purge.plist.template).
Sustituye su marcador por un directorio local de logs antes de instalarla como
`~/Library/LaunchAgents/com.fasttourney.dev-account-purge.plist`. La plantilla
invoca directamente `docker exec` sobre `tournaments-manager-dev-api-1`: no
depende del directorio de trabajo del repositorio, que `launchd` no puede usar
si está protegido por macOS.
Después, `launchctl bootstrap gui/$(id -u) ~/Library/LaunchAgents/com.fasttourney.dev-account-purge.plist`
la ejecuta al cargar la sesión y cada día a las 03:15. Como Docker Desktop es de
usuario, se usa un LaunchAgent y no un LaunchDaemon.
