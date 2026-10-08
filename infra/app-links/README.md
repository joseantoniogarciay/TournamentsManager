# Asociación de enlaces HTTPS

Estos archivos son plantillas, no artefactos publicables: sus marcadores deben
reemplazarse solo al disponer del dominio, el App ID Prefix de Apple y las huellas SHA-256
de los certificados Android reales. No contienen secretos y pueden permanecer
versionados.

## Preparación y publicación

Cada origen publica por HTTPS, sin redirecciones, autenticación ni extensión de
archivo, únicamente los identificadores de su propia variante:

| Entorno | Origen | iOS | Android |
| --- | --- | --- | --- |
| Producción | `https://fasttourney.com` | `production/apple-app-site-association.template` | `production/assetlinks.json.template` |
| Desarrollo | `https://dev.fasttourney.com` | `development/apple-app-site-association.template` | `development/assetlinks.json.template` |

Los archivos se sirven como `/.well-known/apple-app-site-association` y
`/.well-known/assetlinks.json`. Sustituye `APPLE_TEAM_ID` y las huellas SHA-256
solo después de obtener los valores reales de firma; nunca se inventan.

Para producción, las copias ya materializadas permanecen dentro del árbol de
trabajo pero ignoradas por Git en `infra/home/secrets/app-links/`, con los
nombres finales de esos dos recursos. `infra/home/stage-prod-web.sh` las valida
y las incluye cuando se prepare un release destinado también a móvil; un release
web puede existir sin ellas.

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

## Flujo reproducible por entorno

Las plantillas AASA cubren `/link/*`, `/join-team` y `/tournament/*`; Android
limita las rutas desde los intent filters de `app.config.ts`. `/tournament/*`
incluye el detalle `/tournament/:id` y sus subrutas. Los parámetros y el fragmento
de invitación no cambian la asociación; nunca se usan tokens reales para probar
la publicación. La variante `.local` no se publica en ninguna asociación.

1. Obtén el `application-identifier` del binario iOS firmado o su perfil:
   `<App ID Prefix>.<bundleIdentifier>`. Habitualmente el prefijo es el Team ID,
   pero debe comprobarse; sustituye el marcador `APPLE_TEAM_ID` por ese prefijo.
   Verifica también el entitlement de Associated Domains en la build firmada.
2. Obtén la huella SHA-256 del certificado de la app Android que se instalará.
   Para una APK, `apksigner verify --print-certs <app.apk>` permite comprobarla.
   Si se distribuye con Play App Signing, usa el certificado de **firma de app**
   de Play Console, no la clave de subida. Cada huella consta de 32 pares
   hexadecimales mayúsculos separados por `:`. Se admiten varias para una
   rotación verificada del mismo paquete; no se añade la firma local por comodidad.
3. Guarda el par materializado de dev en
   `infra/home/secrets/app-links-development/` y el de prod en
   `infra/home/secrets/app-links/`. Ambos están ignorados por Git; contienen
   identificadores públicos, nunca claves privadas ni certificados de firma.
4. Valida antes de construir:

   ```sh
   FASTTOURNEY_REQUIRE_APP_LINKS=1 node infra/app-links/prepare.mjs development infra/home/secrets/app-links-development
   FASTTOURNEY_REQUIRE_APP_LINKS=1 node infra/app-links/prepare.mjs production infra/home/secrets/app-links
   ```

   El validador comprueba JSON, identidad única por entorno, rutas AASA y formato
   de todas las huellas. **No prueba que los datos pertenezcan a una firma real**;
   eso requiere compararlos con el binario o la consola de distribución.
5. `deploy-dev-web.sh` y `stage-prod-web.sh` validan antes de exportar y copian el
   par en el release antes de promoverlo. Se puede cambiar la carpeta mediante
   `FASTTOURNEY_DEV_APP_LINKS_DIR` o `FASTTOURNEY_PROD_APP_LINKS_DIR`.
   `FASTTOURNEY_REQUIRE_APP_LINKS=1` hace fallar un release móvil sin el par.
   Sin esa opción y sin ambos archivos sigue permitido un release solo web;
   un par parcial o inválido siempre se rechaza.
6. Publica por los procedimientos existentes de cada entorno, con su SHA y sus
   gates. Caddy sirve `/.well-known/*` fuera del fallback SPA en ambos hosts,
   con `Content-Type: application/json`; la ausencia devuelve 404. Actualizar el
   archivo versionado no recarga por sí solo el Caddy del anfitrión.

## Validación pública y nativa

Para cada host consulta ambos recursos mediante `curl --max-redirs 0 -i` a la
URL HTTPS exacta. Exige estado 200, `application/json`, JSON válido y el mismo
contenido que el release, sin redirección ni autenticación. Comprueba también
que un recurso inexistente bajo `/.well-known/` devuelve 404 y no la SPA.

En Android, con la app firmada correcta instalada:

```sh
adb shell pm verify-app-links --re-verify com.fasttourney.app.dev
adb shell pm get-app-links com.fasttourney.app.dev
```

Repite para `com.fasttourney.app` y exige `verified` en su propio host. En iOS
comprueba el AASA público y el diagnóstico de Associated Domains; la caché CDN
puede retrasar cambios. Prueba enlaces desde otra app (por ejemplo Notas), con
la app abierta y cerrada, para `/link/confirm`, `/link/password-reset`,
`/join-team` y `/tournament/<id-de-prueba>`. La prueba de esquema propio o una
navegación en Safari al mismo dominio no acredita Universal Links. Reconstruye
las variantes nativas después de cambiar intent filters; Expo Go y `.local`
no acreditan la asociación de las apps dev/prod.

## Estado y retrospectiva — 2026-10-04

La preparación por release, las rutas y el enrutamiento JSON quedan implementados.
Validación realizada: cinco tests de rechazo/copia (`make test-app-links`,
incluido en `make verify`), lint de archivos afectados, typecheck y exportación
web. Expo config resuelve identidades/hosts y los tres filtros en dev y prod.
Una instancia aislada de Caddy con raíces temporales sirve los dos recursos en
ambos hosts con `200 application/json`; un recurso ausente devuelve 404 sin SPA.
Los datos sintéticos se usan únicamente en pruebas; nunca se publican. La publicación real y la validación en
binarios firmados siguen pendientes del prefijo Apple y las huellas Android
verificadas de cada entorno, y de los gates de despliegue existentes.

Aprendizaje: recepción nativa, asociación pública y firma instalada son tres
comprobaciones independientes. Reutilizar un validador pequeño entre dev y prod
reduce deriva sin introducir otro servicio de publicación.

Fuentes: [Apple Associated Domains](https://developer.apple.com/documentation/xcode/supporting-associated-domains),
[Apple TN3155](https://developer.apple.com/documentation/technotes/tn3155-debugging-universal-links),
[Android App Links](https://developer.android.com/training/app-links/verify-applinks),
[Android troubleshooting](https://developer.android.com/training/app-links/troubleshoot).
