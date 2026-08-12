# Borde doméstico

`Caddyfile` es la configuración versionada del router HTTP local que se instala en
`/opt/homebrew/etc/Caddyfile` en el Mac anfitrión. El servicio Homebrew de Caddy
ejecuta esa ruta; no se edita directamente, para que los cambios permanezcan
revisables en Git.

## Estado inicial seguro

Cloudflare Tunnel termina TLS públicamente y reenvía cada hostname a Caddy por
`http://127.0.0.1:9080`; Caddy solo escucha en loopback. La primera configuración
redirige `www` al dominio canónico, sirve la web estática de desarrollo y enlaza
su API a `127.0.0.1:8081`. Los hosts aún no publicados `fasttourney.com` y
`api.fasttourney.com` devuelven `503`. No expone API directamente, PostgreSQL
ni ficheros `.well-known` incompletos.

No se requieren redirecciones de puertos en UniFi ni DDNS para este flujo. Tras
validar cada hostname por el túnel, se eliminan los forwards TCP 80/443 y la
configuración DDNS. Los proxies y archivos reales se añaden solo junto con sus
respectivos artefactos y configuración.

`deploy-dev-web.sh` exporta Expo con la base API de desarrollo y publica el
artefacto en `/opt/homebrew/var/www/fasttourney/dev`; Caddy no ejecuta Expo Metro.
Como es un entorno de desarrollo público, su host añade
`X-Robots-Tag: noindex, nofollow, noarchive` a todas sus respuestas. La web
mantiene una meta description para que el documento tenga metadatos completos,
pero esa descripción no solicita ni permite su indexación.

La subcarpeta `launchd/` conserva plantillas de tareas del Mac que son
operativamente separadas de Caddy. La purga de cuentas usa un LaunchAgent porque
Docker Desktop y el proyecto Compose de desarrollo pertenecen al usuario; no se
instala ni se actualiza automáticamente desde Git.
