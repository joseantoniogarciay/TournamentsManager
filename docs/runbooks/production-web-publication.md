# Publicación de la web de producción

- **Servicio/componente:** web estática `fasttourney.com`, Caddy y Cloudflare Tunnel
- **Propietario:** operador de FastTourney
- **Última prueba:** 2026-09-05 — TLS, CORS, publicación y rollback operativo verificados
- **Severidad aplicable:** alta mientras afecte a acceso, registro o recuperación

## Síntoma e impacto

Este runbook cubre una nueva publicación o la recuperación de la SPA de
producción. No despliega API, PostgreSQL ni Secrets. Ante una publicación que no
supere las verificaciones, `fasttourney.com` vuelve al `503` versionado.

## Prerequisitos y seguridad

- Una decisión explícita de publicar y un SHA completo, con árbol Git limpio.
- La API pública devuelve `200` en `/healthz`; su CORS permite el origen exacto
  `https://fasttourney.com`.
- Dentro del árbol de trabajo pero ignorados por Git existen, con permisos
  `0600`, `infra/home/secrets/production-web.env` (a partir de
  `infra/home/production-web.env.example`) con el Client ID OAuth web de
  producción.
- El gate web no depende de Apple Developer ni de una firma Android. Las
  asociaciones móviles reales `apple-app-site-association` y `assetlinks.json`
  solo se añaden desde `infra/home/secrets/app-links/` al preparar el primer
  release móvil; exigen entonces Team ID Apple y huella SHA-256 Android reales.
- La API ya recibe el Secret privado `api-integrations`, con SMTP autenticado y
  la audiencia del cliente web de Google, antes de probar registro, recuperación
  o federación web. Su fuente local ignorada es
  `infra/k3s/secrets/api-integrations.env`, creada a partir de
  `infra/k3s/secrets/api-integrations.env.example`.
  Se aplica sin reconstruir la imagen con
  `infra/k3s/scripts/apply-api-integrations.sh`.
- No imprimir archivos de configuración ni Secret; los IDs de cliente son
  públicos una vez exportados, pero la clave SMTP no lo es.

## Diagnóstico

1. Desde una red externa, validar TLS de `https://fasttourney.com` y
   `https://www.fasttourney.com`, sin `--insecure`. Ambos deben presentar un
   certificado válido para el dominio; `www` redirige `308` al host canónico.
2. Confirmar que Cloudflare Tunnel `fasttourney-home` está Healthy y que los
   hostnames de web apuntan a `http://127.0.0.1:9080`.
3. Validar localmente el Caddyfile antes de recargarlo. Mientras el bloque
   `fasttourney.com` siga con `respond ... 503`, el artefacto no es público.
4. Realizar el preflight CORS con origen `https://fasttourney.com` contra
   `https://api.fasttourney.com/v1/registrations`. Debe devolver `204`,
   `Access-Control-Allow-Origin` con el mismo origen y credenciales permitidas.

## Recuperación y publicación

1. Preparar, sin activar, el artefacto trazable:

   ```sh
   infra/home/stage-prod-web.sh <SHA-completo>
   ```

2. Inspeccionar `deployment.json`, `index.html`, `robots.txt` y `sitemap.xml`;
   confirmar que el bundle solo apunta a la API y al dominio de producción, que
   la home declara `https://fasttourney.com/` como canonical y que el sitemap
   solo enumera esa URL. Si el release incluye asociaciones móviles, inspeccionar
   también ambos ficheros `.well-known`.
3. Tras la autorización final, sustituir el `respond` del bloque
   `http://fasttourney.com` por `import production_web`, validar Caddy y
   recargarlo. Esto es el instante de publicación; no conmutar antes el DNS ni
   modificar K3s.
4. Activar el release ya preparado de forma atómica:

   ```sh
   infra/home/activate-prod-web.sh <SHA-completo>
   ```

5. Si no está instalado aún o si cambió el renderer, compilarlo y cargar su
   LaunchAgent desde
   `infra/home/launchd/com.fasttourney.prod-league-preview-renderer.plist.template`.
   Sustituir exclusivamente `__LEAGUE_PREVIEW_BINARY__` y `__LOG_DIRECTORY__`,
   validar el plist y confirmar que escucha solo en `127.0.0.1:8091`. El binario
   usa `prod/current`, por lo que no se reinicia en cada activación del release.
   Validar y recargar Caddy después de que el proceso esté sano.
6. Con una liga pública controlada, comprobar por HTTPS el HTML de
   `/league/<uuid>`: debe devolver `200`, el mismo canonical, `og:title` con el
   nombre, `og:image` de 1200×630 y `X-Robots-Tag: noindex, nofollow, noarchive`.
   Una liga inexistente debe devolver `404`; `/league/<uuid>/standings` debe
   seguir llegando al fallback de la aplicación. Probar la primera URL real en
   un inspector social y conservar solo la evidencia saneada.
7. Recorrer en navegador un registro con buzón de prueba controlado,
   confirmación por enlace, login local, recuperación de contraseña y login
   Google web. El gate móvil posterior verificará además que ambos recursos
   `/.well-known` devuelven `200`, JSON válido y sin redirección.
8. Confirmar por HTTPS que `/` no entrega `X-Robots-Tag: noindex` y que una
   ruta de aplicación, por ejemplo `/account`, sí lo entrega. Verificar el
   dominio en Search Console, inspeccionar `/` y enviar el sitemap tras activar
   el release.
9. Observar Grafana y Alertmanager durante la ventana inicial. Cualquier `5xx`,
   fallo SMTP, error de CORS, violación CSP o límite por IP incoherente detiene
   el gate.

## Rollback y escalado

- Si falla antes de importar `production_web`, no se ha publicado la web:
  eliminar el release preparado si corresponde, sin tocar API o datos.
- Si la web ya está abierta, restaurar de inmediato el `respond ... 503`,
  validar Caddy, recargar y comprobar el `503` por HTTPS público. Es el rollback
  seguro cuando hay dudas sobre TLS, contenido o identidad.
- Si el incidente se limita a un artefacto ya servido y el borde sigue sano,
  cambiar el enlace con `infra/home/rollback-prod-web.sh <SHA-conservado>`.
  Este rollback no modifica API, K3s, PostgreSQL, migraciones ni Secrets.
- Si el certificado público no es válido o el host sirve contenido ajeno, no
  continuar: revisar la ruta de hostname en Cloudflare Tunnel y la configuración
  TLS/DNS de Cloudflare antes de abrir el origen.

## Después del incidente

Conservar SHA, hora, respuestas HTTP saneadas, estado del túnel y de alertas.
Actualizar este runbook, `docs/operations/DEPLOYMENT.md` y
`docs/project/LEARNING.md` con la evidencia real de la primera prueba.

## Evidencia de publicación — 2026-09-05

Con la autorización explícita para la apertura, se publicó el release
`22b98367e32b9c165a636e5dcccf5b7bdba90e5c` y se instaló el Caddyfile
versionado que importa `production_web`. La comprobación externa posterior
confirmó `https://fasttourney.com/` en `200` con TLS válido,
`https://www.fasttourney.com/` en `308` al host canónico, y las rutas legales
en `200`. La API conservó `200` en `/healthz`; el preflight de registro devolvió
`204`, el origen exacto `https://fasttourney.com` y credenciales permitidas.
La inspección anónima del cliente cargó inicio y la pantalla de cuenta, incluido
el acceso local, recuperación, alta y la acción Google. Quedan como gate
posterior los recorridos que envían correo y el login real con una cuenta de
prueba controlada, además de los clientes y asociaciones móviles.

Rollback explícito: restaurar el `respond ... 503` en el bloque
`http://fasttourney.com`, validar y recargar Caddy; si el borde está sano y el
problema es solo el artefacto, usar `rollback-prod-web.sh` con un SHA conservado.

## Evidencia de recuperación TLS — 2026-09-05

La tabla de certificados de borde estaba vacía pese a que DNS, las cinco rutas
del Tunnel `fasttourney-home` y Caddy local eran correctos. Con autorización
explícita se desactivó y reactivó Universal SSL para solicitar una reemisión.
Durante la emisión, web y API rechazaron el handshake TLS; no se modificaron
Caddy, K3s, DNS ni la respuesta `503` de la web. Tras la propagación, las
comprobaciones externas verificaron certificado válido (`ssl_verify_result=0`),
`https://fasttourney.com` en `503`, `https://www.fasttourney.com` en `308` al
host canónico y `https://api.fasttourney.com/healthz` en `200`. Un `502` inicial
de API fue transitorio: Caddy local y el Ingress privado ya devolvían `200` y el
reintento público también devolvió `200`.

## Evidencia de publicación del favicon — 2026-09-21

El release `v1.7.1`, SHA
`e87ddab9cc1a9828e7e395114fe2440ad00cb6c5`, publicó el favicon en `dev` y
`prod`. `make verify` y las ejecuciones remotas de `Verify` en `develop` y
`main` terminaron correctamente. El artefacto inmutable contiene un ICO
cuadrado de 48 px con tamaños adicionales de 32 y 16 px, y la home declara
`<link rel="icon" href="/favicon.ico"/>`.

Tras activar el release y recargar el Caddyfile versionado, la comprobación
HTTPS de producción confirmó `/favicon.ico` en `200`, `Content-Type:
image/x-icon` y sin `X-Robots-Tag: noindex`; `/account` conservó `noindex,
nofollow, noarchive`. `www` mantuvo su redirección `308`, la API devolvió `200`
y el preflight CORS de registro devolvió `204` con el origen productivo exacto.

El host `dev` quedó en el mismo SHA y su API devolvió `200`. Cloudflare conservó
temporalmente en caché la antigua respuesta HTML de `/favicon.ico`; una petición
con URL no cacheada confirmó el ICO nuevo en origen. El host de desarrollo
mantiene deliberadamente `X-Robots-Tag: noindex, nofollow, noarchive` en todas
sus respuestas. El rollback web de producción permanece en `v1.7.0`, SHA
`b54578fe4bdd8798b292780d8b4463cf525cd44d`.

## Evidencia de publicación del icono de Safari — 2026-09-21

El release `v1.7.2`, SHA
`1c0c31c7868d46a333ad142219e6462811f61116`, publicó en `dev` y `prod` un
`apple-touch-icon` explícito sin añadir manifest, modo `standalone` ni otras
capacidades PWA. `make verify` y las ejecuciones remotas de `Verify` en
`develop` y `main` terminaron correctamente. El artefacto inmutable contiene
un PNG de 1024×1024, RGB y sin transparencia; la home declara tanto
`apple-touch-icon` como `/favicon.ico`.

Después de activar producción y recargar el Caddyfile versionado, las
comprobaciones HTTPS confirmaron `/apple-touch-icon.png` en `200` y
`Content-Type: image/png` en ambos entornos. Producción no aplica `noindex` al
icono y `/account` conserva `noindex, nofollow, noarchive`; desarrollo mantiene
deliberadamente esa cabecera en todas sus respuestas. Los manifiestos de
despliegue de ambos entornos exponen el mismo SHA y las APIs de `dev` y `prod`
respondieron `200` en `/healthz`.

El rollback web inmediato de producción es `v1.7.1`, SHA
`e87ddab9cc1a9828e7e395114fe2440ad00cb6c5`.
