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

`deploy-dev.sh` es el único despliegue manual de dev: exige `develop` limpio y
alineado con `origin/develop`, construye la API runtime, aplica las migraciones
SQL pendientes y llama a `deploy-dev-web.sh`. Las migraciones son solo hacia
delante: un rollback de imagen no revierte el esquema. Cada export se guarda bajo
`/opt/homebrew/var/www/fasttourney/dev/releases/<SHA>/` y contiene un manifiesto
sin secretos con commit, imagen y fecha. Caddy sirve el enlace `current`; el
cambio del enlace es atómico mediante `mv -h`, para reemplazar el enlace y no el
directorio al que apunta en macOS, y se retienen solamente la versión actual y
su predecesora. Git conserva fuente y configuración, no artefactos, imágenes ni
backups. `deploy-dev-web.sh` ignora el `.env` local y limpia la caché de Metro
para que ambas URL públicas se incorporen al bundle. También fija
`APP_ENV=development`, de modo que la exportación usa la identidad **Fast
Tourney Dev** en lugar de la variante local. Caddy no ejecuta Expo
Metro.

Si la espera de salud de Compose falla, `deploy-dev.sh` no conmuta `current` ni
escribe el manifiesto. Antes de salir muestra el estado de los servicios y los
últimos logs de API y PostgreSQL, para diagnosticar el fallo sin imprimir los
contratos `.env` ni secretos.

Cada release de dev también incorpora la referencia Scalar y la copia exacta de
OpenAPI con la que se construyó. Caddy la sirve públicamente, sin DNS ni Access
adicionales, en `https://dev.fasttourney.com/api-docs/`; la interfaz llama a
`https://dev-api.fasttourney.com/v1` bajo el mismo origen web ya autorizado por
CORS. La referencia no se indexa, como el resto del host de desarrollo.

Para recuperar el último artefacto anterior, se usa
`make dev-public-rollback SHA=<SHA-completo>`. La reversión cambia API y web al
mismo SHA conservado, pero no sustituye una restauración de PostgreSQL.
Como es un entorno de desarrollo público, su host añade
`X-Robots-Tag: noindex, nofollow, noarchive` a todas sus respuestas. La web
mantiene una meta description para que el documento tenga metadatos completos,
pero esa descripción no solicita ni permite su indexación.

Todos los hosts públicos añaden `Strict-Transport-Security: max-age=31536000`.
Aunque Caddy recibe HTTP de loopback desde el túnel, Cloudflare termina la
conexión HTTPS pública y reenvía la cabecera al cliente. No se usa
`includeSubDomains` ni `preload`: ambas opciones comprometerían subdominios
presentes y futuros, y se revisarán solo cuando todos puedan sostener HTTPS.
Además, `X-Content-Type-Options: nosniff` evita interpretaciones de tipo
inesperadas, `Referrer-Policy: strict-origin-when-cross-origin` limita la URL
que se comparte al navegar fuera del sitio y `Permissions-Policy` desactiva
cámara, geolocalización, micrófono, pagos y USB mientras la web no los requiere.

La web de desarrollo publica una `Content-Security-Policy` que permite solo el
propio host, su API de desarrollo y los orígenes de Google necesarios para el
inicio de sesión federado; además niega objetos, embebido por iframes,
formularios externos y bases ajenas. Antes de endurecerla, se publicó en modo
`Content-Security-Policy-Report-Only` y se recorrieron los flujos públicos sin
avisos; todo origen adicional debe justificarse por una capacidad real, nunca
añadirse por comodidad.

La subcarpeta `launchd/` conserva plantillas de tareas del Mac que son
operativamente separadas de Caddy. La purga de cuentas usa un LaunchAgent porque
Docker Desktop y el proyecto Compose de desarrollo pertenecen al usuario; no se
instala ni se actualiza automáticamente desde Git.

Los templates `com.fasttourney.dev-postgresql-backup-full.plist.template` y
`com.fasttourney.dev-postgresql-backup-incremental.plist.template` programan,
respectivamente, la copia completa semanal y los incrementales de lunes a
sábado. Se instalan manualmente tras `make dev-public-backup-init`; el
[runbook de backup](../../docs/runbooks/postgresql-backup-dev.md) define la
restauración aislada.

Los templates equivalentes de `prod` se instalan con
`infra/k3s/scripts/install-backup-launch-agents.sh`. El instalador deja una
copia ejecutable y su configuración privada fuera de `Desktop`, y un helper
sandboxed publica en iCloud mediante permisos explícitos de carpeta conforme a
ADR-0115; la fuente y el procedimiento siguen versionados en el repositorio.
Consulta el
[runbook PostgreSQL de K3s](../../docs/runbooks/k3s-postgresql.md) para la
verificación posterior.
