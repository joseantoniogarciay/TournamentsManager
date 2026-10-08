# Preparar el acceso social Google y Apple

Decisión: [ADR-0147](../adr/0147-add-apple-browser-login-on-all-clients.md).
El código puede compilar con placeholders; eso no acredita acceso real.

## Presentación móvil

Google usa Expo AuthSession; Apple usa `openAuthSessionAsync` de Expo WebBrowser.
iOS presenta ASWebAuthenticationSession y Android Custom Tabs, con retorno a la
app. No se usa WebView ni `Linking.openURL` para abrir el proveedor. El sistema
conserva sus controles, cookies y gestor de contraseñas. En web se abre un popup
desde el gesto; el layout completa su retorno en el mismo origen.
Android prioriza un paquete compatible con Custom Tabs. Si no existe ninguno,
Expo abre el navegador del sistema y conserva el retorno de autenticación a la app,
según la aclaración explícita del usuario; la ausencia de Custom Tabs no bloquea el login.

## Google

El backend ya verifica los ID tokens y consume un nonce de un solo uso. Configura
`GOOGLE_CLIENT_IDS` en API y los IDs públicos WEB/IOS/ANDROID en el build cliente.
La cuenta Play Console es para distribución; OAuth se configura aparte en Google
Cloud/Google Auth Platform. Web exige sus orígenes HTTPS autorizados. Para móvil,
registra bundle/package y certificados reales de firma; verifica que el cliente
Android permita el retorno de custom URI scheme del flujo AuthSession vigente.
Si no permite ese flujo, no se sustituye el client ID Android por un ID web: se
revisa el adaptador antes de distribuir. No se introduce un SDK silenciosamente.

## Apple

Por entorno, registra el App ID primario con Sign in with Apple y un Services ID
agrupado con ese App ID. Registra dominio y retorno HTTPS exacto:

| Entorno | Callback proveedor | Retorno web | Retorno móvil |
| --- | --- | --- | --- |
| dev | `https://dev-api.fasttourney.com/v1/apple-callback` | `https://dev.fasttourney.com/oauth/apple-complete` | `fasttourney-dev://oauth/apple-complete` |
| prod | `https://api.fasttourney.com/v1/apple-callback` | `https://fasttourney.com/oauth/apple-complete` | `fasttourney://oauth/apple-complete` |

El callback es un POST `application/x-www-form-urlencoded` de Apple. No configura
CORS general para Apple ni emite sesión. Valida state, intercambia código y verifica
nonce/ID token; solo entonces el cliente iniciador consume su prueba por JSON.
Un callback inválido no redirige; un fallo tras reclamar state vuelve con `failed`;
una denegación vuelve con `cancelled` y no produce feedback. Reintenta desde la app.

Variables API (plantillas `.env.example`, nunca valores inventados en runtime):

```dotenv
APPLE_SERVICE_ID=REPLACE_WITH_APPLE_SERVICE_ID
APPLE_TEAM_ID=REPLACE_WITH_APPLE_TEAM_ID
APPLE_KEY_ID=REPLACE_WITH_APPLE_KEY_ID
APPLE_PRIVATE_KEY_FILE=/run/secrets/REPLACE_WITH_APPLE_PRIVATE_KEY.p8
APPLE_REDIRECT_URI=https://dev-api.fasttourney.com/v1/apple-callback
APPLE_NATIVE_SCHEME=fasttourney-dev
```

La clave privada P-256 descargada de Apple se guarda fuera de Git y se monta en
lectura dentro del proceso API en la ruta indicada. No se copia al chat, .env
público ni bundle. En K3s/Compose hay que preparar ese montaje cuando exista la
clave; estas plantillas no crean ni montan una clave ficticia. Una configuración
parcial o con placeholders deja todos los endpoints Apple en `503`; una clave
real inválida impide arrancar. En cliente y export web del mismo entorno:

```dotenv
EXPO_PUBLIC_APPLE_SERVICE_ID=REPLACE_WITH_APPLE_SERVICE_ID
```

Debe coincidir exactamente con el Services ID servidor. No hace falta incorporar
Team ID/Key ID al cliente. El flujo navegador iOS no incorpora el SDK Apple ni
activa un entitlement nativo ficticio. Configura además en Apple los dominios y
emisores de correo para el relay privado; prueba verificación y recuperación
sobre un email relay real.

## Gates de publicación

1. Aplicar 00020 con la identidad de migración. PostgreSQL admite únicamente los
   pares Google/issuer Google y Apple/issuer Apple. No hay rollback destructivo.
2. Configurar valores reales por entorno y exportar/reconstruir las apps. Mantener
   local/dev apagados fuera de la sesión autorizada y observabilidad solo a petición.
3. Probar Google y Apple en builds firmadas: alta, acceso existente, username y
   términos, email repetido, email relay, borrador, cancelación, reintento,
   expiración, arranque en frío y retorno al modal original. Comprobar que no se
   prioriza la presentación interna y permite navegador externo si no hay Custom
   Tabs; comprobar que otra app no puede consumir el challenge.
4. Completar la revocación Apple al eliminar cuenta y revisar los requisitos de
   App Review antes de distribuir iOS. Este incremento descarta los tokens Apple
   de acceso/refresh y no implementa esa revocación ni gestión de vínculos Apple.
5. Superar los gates de asociaciones HTTPS y PostHog de builds distribuidas. Las
   pruebas sintéticas no sustituyen firmas, credenciales ni aprobaciones reales.

## Retrospectiva de implementación

Compartir la transacción de cuenta/sesión/borrador evita copiar reglas de negocio.
Apple necesita un callback distinto por `form_post`; abstraer todos los OAuth
antes de necesitarlo añadiría complejidad. State protege el callback, nonce liga
el token al intento y la prueba del cliente protege la vuelta a un esquema móvil:
son controles distintos. Un éxito 200 también necesita validar el cuerpo en cliente.

Fuentes: [Apple autorización](https://developer.apple.com/documentation/signinwithapplerestapi/request-an-authorization-to-the-sign-in-with-apple-server),
[Apple token validation](https://developer.apple.com/documentation/signinwithapplerestapi/generate-and-validate-tokens),
[Expo WebBrowser](https://docs.expo.dev/versions/latest/sdk/webbrowser/),
[Google OAuth móvil](https://developers.google.com/identity/protocols/oauth2/native-app).
