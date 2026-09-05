# Publicación de la web de producción

- **Servicio/componente:** web estática `fasttourney.com`, Caddy y Cloudflare Tunnel
- **Propietario:** operador de FastTourney
- **Última prueba:** 2026-09-05 — recuperación TLS; publicación de la SPA pendiente
- **Severidad aplicable:** alta mientras afecte a acceso, registro o recuperación

## Síntoma e impacto

Este runbook cubre la apertura o la recuperación de la SPA de producción. No
despliega API, PostgreSQL ni Secrets. Hasta completar cada verificación,
`fasttourney.com` conserva el `503` versionado.

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

2. Inspeccionar `deployment.json` e `index.html`; confirmar que el bundle solo
   apunta a la API y al dominio de producción. Si el release incluye
   asociaciones móviles, inspeccionar también ambos ficheros `.well-known`.
3. Tras la autorización final, sustituir el `respond` del bloque
   `http://fasttourney.com` por `import production_web`, validar Caddy y
   recargarlo. Esto es el instante de publicación; no conmutar antes el DNS ni
   modificar K3s.
4. Activar el release ya preparado de forma atómica:

   ```sh
   infra/home/activate-prod-web.sh <SHA-completo>
   ```

5. Recorrer en navegador un registro con buzón de prueba controlado,
   confirmación por enlace, login local, recuperación de contraseña y login
   Google web. El gate móvil posterior verificará además que ambos recursos
   `/.well-known` devuelven `200`, JSON válido y sin redirección.
6. Observar Grafana y Alertmanager durante la ventana inicial. Cualquier `5xx`,
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
