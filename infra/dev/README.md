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

## Primer arranque

```bash
make dev-public-init
# Edita infra/dev/.env y infra/dev/api.docker.env con la misma contraseña única.
make dev-public-up
make dev-public-schema-apply
```

## Actualización y reversión de dev

Tras pasar `make verify` y la CI del commit de `develop`, el despliegue manual
es:

```bash
make dev-public-deploy
```

El comando exige un árbol limpio cuyo `HEAD` coincida exactamente con
`origin/develop`. Construye una imagen `runtime` etiquetada con el SHA completo,
exporta la web estática en un directorio por SHA y publica ambas versiones. La
web cambia mediante el enlace simbólico `current`, por lo que no mezcla archivos
de dos exports. Se conservan el despliegue actual y el anterior bajo
`/opt/homebrew/var/www/fasttourney/dev/releases/`; cada uno contiene un
`deployment.json` con SHA, imagen y fecha. Git conserva el código, no estos
artefactos locales.

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
