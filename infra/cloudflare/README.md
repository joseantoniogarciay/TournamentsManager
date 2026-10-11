# Hosting web Cloudflare

ADR-0155 aloja los archivos de producción en Workers Static Assets. La API sigue
en K3s doméstico vía Tunnel/Caddy. Dominio y audiencia OAuth no cambian.

## Estado efectivo

2026-10-11: rutas productivas activadas en Workers Free. Se conserva la app
`b002b4b99fc44a6ec0662b94ec962b8cce9cbea5`; no se promocionan cambios funcionales
de develop. K3s, PostgreSQL y dos réplicas API están recuperados, con healthz
público 200 y CORS productivo 204. Backups y renderer de rollback rehabilitados.
No se contrató pago ni se cambió el DNS/API. La comprobación de corte detectó
`noindex` en la home del borde, pese a pasar localmente: la home pasa primero
por el Worker para retirar esa cabecera solo en el host canónico. El resto de
archivos sigue usando el servicio estático y los previews conservan noindex.

El certificado ajeno recibido el 10 de octubre coincide con la interceptación
Digi documentada por [OONI](https://ooni.org/post/2026-laliga-collateral/), incluida
su huella SHA-1. No demostraba ausencia de certificado en Cloudflare. El día 11
la conexión vuelve a validar HTTPS con Google Trust Services WE1. Universal SSL
queda activado; no se repite su reemisión ni se omite validación TLS.

## Preparar y comprobar

Wrangler 4.147.0 está fijado en el lockfile; usa el runtime oficial `workerd`,
cuyo instalador está autorizado en `allowBuilds`. Se respeta la espera de siete
días. `make test-cloudflare-web` prueba ese runtime en loopback con API simulada.

Una nueva versión se exporta con `infra/home/stage-prod-web.sh <SHA>` después
del gate y promoción de ADR-0119. Ese script sigue siendo solo un constructor;
guardar el release en disco no exige que Caddy sea el hosting activo. La primera
migración reutiliza el artefacto ya publicado, sin promover mejoras de `develop`.

Con checkout limpio:

```sh
make cloudflare-web-stage SHA=<SHA-web>
make cloudflare-web-check RELEASE=<directorio-impreso>
```

El candidato `.config/cloudflare/releases/<SHA-web>-<SHA-infra>/` contiene assets,
Worker, cabeceras y configuración. `release.json`, fuera de assets, registra ambos
SHA y hashes SHA-256; los comandos rechazan modificaciones posteriores. El
manifiesto público `deployment.json` conserva el SHA de la app. Se rechazan
source maps, claves y `.env` en los archivos. Preparar no actualiza `current`.

## Subir y publicar

Autenticar Wrangler en la cuenta propietaria de la zona; no imprimir ni versionar
tokens. Comprobar Workers Free y permisos de Workers/rutas antes de publicar.
No seleccionar planes de pago.

La autorización mínima utilizada es `account:read user:read
workers_scripts:write workers_routes:write zone:read`; Wrangler añade
`offline_access`. El permiso genérico `workers:write` no permitió consultar los
despliegues con la API actual. No ampliar a los demás scopes del login por defecto.

La primera subida de versión exige que el Worker exista. Si Cloudflare confirma
que no existe, crear primero una configuración independiente del candidato, con
el mismo nombre, código y assets, `routes: []`, `workers_dev: false` y
`preview_urls: true`, y ejecutar `wrangler deploy --config <config-inicial>`.
Guardar fuera del candidato esta configuración, resolviendo `main` y el
directorio de assets a sus rutas absolutas. El resultado debe indicar **No
targets deployed**. Este paso es solo para la creación inicial: no repetirlo
contra un Worker con rutas activas porque retiraría sus triggers. Después usar
el candidato inmutable para `preview` y `publish`.

```sh
make cloudflare-web-preview RELEASE=<directorio>
# Revisar la URL de versión antes de activar el dominio.
make cloudflare-web-publish RELEASE=<directorio>
```

`preview` usa `wrangler versions upload`: sube sin activar rutas. `publish` usa
`wrangler deploy` para `fasttourney-web` y después `fasttourney-web-redirect`.
El segundo conserva el 308 de `www` incluso con el Mac apagado. Se usan rutas
sobre DNS proxied existente, sin retirar registros de Tunnel ni cambiar API.
`workers.dev` queda desactivado; las URLs de versión están habilitadas y llevan
`noindex`. No sustituyen el gate de login: CORS solo autoriza el dominio productivo.

Los archivos se sirven antes de la función, salvo la home y `/tournament/*`:
la home retira noindex solo en el dominio canónico; el UUID exacto de torneo consulta la API pública con timeout de tres segundos; sus subrutas
usan el shell de los mismos assets. Nunca se reenvían sesión, cookies ni token
privado de borde. Otras rutas sin archivo conservan el fallback Expo. Archivos y
asociaciones inexistentes devuelven 404. Cabeceras estáticas y dinámicas comparten
CSP, HSTS y política de indexación. Se conserva el copy del renderer Go,
incluida su descripción histórica de liga, evitando cambios de producto aquí.

## Coste y recuperación

Los archivos estáticos tienen almacenamiento y peticiones gratuitos. La función (incluida la home)
y `www` consumen la cuota compartida Workers Free de 100.000 peticiones/día por
cuenta y su límite de CPU. No se usa `run_worker_first=true` global. Agotar cuota
afecta a documentos dinámicos/fallback y `www`, sin contratar pago automáticamente.
API sigue dependiendo del Mac. Estos comandos desactivan métricas de Wrangler
y Workers Logs; no activan observabilidad dev.

Comprobar manifiesto, home indexable, cuenta noindex, tipos de assets, rutas
legales, subrutas, torneo público con canonical/previews, UUID inexistente 404,
`www` 308 conservando query, API `/healthz` y CORS productivo.

Conservar IDs de versiones de ambos Workers y sus rutas previas. Para volver a
otra versión Cloudflare usar `wrangler rollback <version-id>` con su configuración.
En la primera migración, si no existe versión anterior, retirar **solo las rutas
de estos dos Workers** devuelve el tráfico al Tunnel/Caddy conservado. No borrar
Workers, DNS ni datos. Si un deploy falla, revisar qué rutas/versiones quedaron
activas antes de reintentar: los dos deploys no son una transacción atómica.

Tras la conmutación `activate-prod-web.sh`/`rollback-prod-web.sh` solo cambian
el origen Caddy de recuperación. Los siguientes despliegues usan este flujo.
Dev y su renderer Go conservan ADR-0146 y permanecen apagados fuera de pruebas.

## Retrospectiva

Una exportación exige verificar routing y metadatos además de copiar archivos.
La prueba con el runtime de Wrangler detectó que `redirect: "error"` no es
aceptado; `manual` permite rechazar los saltos de API con respuesta segura.
La URL de versión tiene dos etiquetas antes de `workers.dev`; una regla con una
sola no la protegía de indexación. Separar assets y renderer conserva gratuito
el tráfico estático sin cambiar el cliente universal ni lógica de negocio. La comprobación del
dominio real sigue siendo necesaria: la indexación de la home no coincidió
con el runtime local y se trata explícitamente en el Worker.

## Evidencia local — 2026-10-10

- App productiva reutilizada: `b002b4b99fc44a6ec0662b94ec962b8cce9cbea5`.
- Implementación: `219dbf8`; candidato independiente bajo `.config/cloudflare`.
- Trece pruebas de hosting/dependencias y once de seguridad operacional pasan.
  ESLint, TypeScript, formato y export Expo de 36 rutas pasan. No se ejecutó el
  gate completo de promoción ni se creó un nuevo release de producto.
- Dry-run de ambos Workers correcto: 110 archivos y Worker de 7,03 KiB.
- Wrangler local verificó doce rutas del artefacto real: home, cuenta, ajustes,
  legales, subruta de torneo, iconos, robots, sitemap y manifiesto en 200;
  asociación móvil ausente en 404. Tipos MIME, SHA público, `nosniff` y política
  de indexación correctos. No se certificó login en URL de versión ni cuota/plan
  de la cuenta: requieren acceso y validación del dominio real.
- `wrangler whoami` confirma ausencia de autenticación. No se subió ni activó
  ninguna versión. El servidor local se detuvo al terminar; dev/observabilidad
  permanecieron apagados y producción doméstica no se modificó.

## Evidencia remota — 2026-10-10

- OAuth conectado con permisos de scripts/rutas y lectura de usuario/cuenta/zona;
  no se registran credenciales en Git. Plan de la cuenta: Workers Free, $0.
- Creación inicial sin targets: `e4ae4f3e-9657-46e5-9f9c-98e3a9cc51a9`.
  Versión de prueba: `d7feb929-0197-4c8e-a876-74e96685a05a`, infraestructura
  `1d760bb03f6cc479be074d6088769bad3c271ac0` y app `b002b4b…`.
- [URL de versión](https://d7feb929-fasttourney-web.joseantoniogarciay.workers.dev):
  home, manifiesto, cuenta, privacidad y subruta de clasificación en 200;
  asociación móvil ausente y JS inexistente en 404. `noindex`, MIME y `nosniff`
  presentes; SHA de `deployment.json` coincide con producción. Cloudflare añade
  su propio `noindex` a la URL de versión. Startup 1 ms.
- Dos fichas raíz de torneo devuelven el 503 seguro: API productiva suspendida.
  No certifica aún metadatos reales ni 404 de un torneo ausente desde el borde.
- La API Cloudflare confirma zona activa y cero rutas Workers en la zona.
  El TLS público presenta un certificado ajeno (`core1.netops.test`); DNS público
  y resolver HTTPS coinciden. No se omitió validación de certificados.
  Tabla Edge Certificates vacía; reemisión autorizada solicitada y Universal SSL
  reactivado. La publicación queda pendiente del gate TLS y salud de API.

Retrospectiva: un dry-run no comprueba permisos remotos ni el requisito de crear
el Worker antes de subir versiones. Separar esa creación sin rutas evita cortar
el origen antes de validar. El HTTPS de la URL de versión y el del dominio son
gates independientes; restaurar K3s no corrige una interceptación TLS del operador.
