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

`deploy-dev-web.sh` exporta Expo con la base API y el origen de enlaces de
desarrollo, y publica el artefacto en `/opt/homebrew/var/www/fasttourney/dev`;
ignora el `.env` local y limpia la caché de Metro para que ambas URL públicas se
incorporen al bundle. Caddy no ejecuta Expo Metro.
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
