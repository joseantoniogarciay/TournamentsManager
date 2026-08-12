# Asociación de enlaces HTTPS

Estos archivos son plantillas, no artefactos publicables: sus marcadores deben
reemplazarse solo al disponer del dominio, el Apple Team ID y las huellas SHA-256
de los certificados Android reales. No contienen secretos y pueden permanecer
versionados.

## Publicación futura

Cada origen publica por HTTPS, sin redirecciones, autenticación ni extensión de
archivo, únicamente los identificadores de su propia variante:

| Entorno | Origen | iOS | Android |
| --- | --- | --- | --- |
| Producción | `https://fasttourney.com` | `production/apple-app-site-association.template` | `production/assetlinks.json.template` |
| Desarrollo | `https://dev.fasttourney.com` | `development/apple-app-site-association.template` | `development/assetlinks.json.template` |

Los archivos se sirven como `/.well-known/apple-app-site-association` y
`/.well-known/assetlinks.json`. Sustituye `APPLE_TEAM_ID` y las huellas SHA-256
solo después de obtener los valores reales de firma; nunca se inventan.

Después de publicar, configura el mismo origen HTTPS en `PUBLIC_BASE_URL` del
backend de cada entorno: `https://fasttourney.com` en producción y
`https://dev.fasttourney.com` en desarrollo. El correo enlaza a
`/link/confirm?token=…`; abrirlo no consume nada y
la persona confirma explícitamente con `POST /v1/registration-verifications`.
En desarrollo loopback se permite HTTP solo para probar la web: los universal
links de iOS y Android requieren el dominio HTTPS asociado.

`webcredentials` permite que iOS relacione las credenciales guardadas del host
con su propia variante de app. `app.config.ts` declara el host correspondiente
como `applinks:` y `webcredentials:` según `APP_ENV`.
El restablecimiento usa `/link/password-reset`; queda cubierto por el componente
`/link/*` y no necesita un fichero de asociación adicional.
