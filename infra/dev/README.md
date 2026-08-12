# Entorno público de desarrollo

Este directorio define el proyecto Compose `tournaments-manager-dev`. Sus
servicios internos se llaman `api`, `postgres` y `mailpit`; Docker añade el
namespace del proyecto, de modo que no colisionan con
`tournaments-manager-local` ni con el futuro `tournaments-manager-prod`.

La API usa el target `runtime` del Dockerfile: no incluye Air, fuentes ni bind
mounts. Solo se publica en `127.0.0.1:8081`, donde Caddy la alcanza desde el
Mac. PostgreSQL no publica ningún puerto. Mailpit queda en `127.0.0.1:8026`
para inspección local y no tiene ruta Cloudflare/Caddy.

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

La web se exporta como estática con `infra/home/deploy-dev-web.sh`; no se usa
`expo start --web` como servidor público. Antes de invitar usuarios que deban
recibir emails, sustituye Mailpit por un proveedor SMTP transaccional y configura
sus SPF/DKIM/DMARC.

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
