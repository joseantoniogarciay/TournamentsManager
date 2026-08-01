# Asociación de enlaces HTTPS

Estos archivos son plantillas, no artefactos publicables: sus marcadores deben
reemplazarse solo al disponer del dominio, el Apple Team ID y las huellas SHA-256
de los certificados Android reales. No contienen secretos y pueden permanecer
versionados.

## Publicación futura

Con `EXPO_PUBLIC_APP_LINK_URL=https://enlaces.ejemplo.com`, publica por HTTPS,
sin redirecciones, autenticación ni extensión de archivo:

| Plataforma | Ruta pública                              | Fuente de este repositorio                                                                                  |
| ---------- | ----------------------------------------- | ----------------------------------------------------------------------------------------------------------- |
| iOS        | `/.well-known/apple-app-site-association` | `apple-app-site-association.template` tras sustituir `APPLE_TEAM_ID`; incluye `applinks` y `webcredentials` |
| Android    | `/.well-known/assetlinks.json`            | `assetlinks.json.template` tras sustituir las huellas                                                       |

Las dos variantes de la app están declaradas porque desarrollo y producción usan
identificadores distintos. Si no se publica una variante, se elimina su entrada;
nunca se inventa un Team ID ni una huella.

Después de publicar, configura el mismo origen HTTPS en `PUBLIC_BASE_URL` del
backend. El correo enlaza a `/link/confirm?token=…`; abrirlo no consume nada y
la persona confirma explícitamente con `POST /v1/registration-verifications`.
En desarrollo loopback se permite HTTP solo para probar la web: los universal
links de iOS y Android requieren el dominio HTTPS asociado.

`webcredentials` permite que iOS relacione las credenciales guardadas del
dominio con ambas variantes de la app. `app.config.ts` declara el mismo dominio
como `applinks:` y `webcredentials:` cuando existe `EXPO_PUBLIC_APP_LINK_URL`.
El restablecimiento usa `/link/password-reset`; queda cubierto por el componente
`/link/*` y no necesita un fichero de asociación adicional.
