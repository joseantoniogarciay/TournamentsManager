# ADR-0147: Añadir acceso con Apple mediante navegador en todos los clientes

- **Estado:** Aceptado
- **Fecha:** 2026-10-04
- **Decisor:** Usuario, mediante petición explícita de implementar Google y Apple también en Android con placeholders
- **Supera a:** ADR-0050 y el alcance v1 de PRODUCT.md, solo en el aplazamiento de Apple

## Problema y evidencia

Google ya tiene challenge, verificación OIDC, sesión y alta. Apple debe ofrecerse
además en Android; no hay todavía cuenta de distribución ni datos reales Apple.
El gate técnico está cerrado. La identidad sigue siendo `(issuer, subject)` y
ADR-0066 impide unir cuentas por coincidencia de email.

## Alternativas y coste

- SDK Apple nativo iOS y navegador Android/web: mejor integración iOS, pero dos
  flujos, un módulo nativo adicional y dos matrices de pruebas.
- Navegador del sistema en web/iOS/Android, callback HTTPS y verificación backend:
  un flujo sobre expo-web-browser existente; requiere Services ID, App ID primario,
  Team ID, Key ID y clave privada solo en servidor. Mantenimiento moderado.
- Proveedor de identidad gestionado: reduce implementación propia, pero cambia
  la arquitectura aceptada y añade una dependencia de servicio innecesaria.

## Recomendación y decisión

La petición del usuario acepta incorporar Apple también en Android y preparar
la configuración pendiente con placeholders. Se concreta con navegador del sistema
para Apple y se conserva el flujo Google existente, sin nueva plataforma de identidad.
No se usan WebViews. Los placeholders no activan proveedores ni se publican como
asociaciones reales. El futuro paso a SDK nativo iOS exige reevaluar su coste.

## Invariantes

- Apple retorna un código por `form_post` a un callback HTTPS fijo. El backend
  intercambia el código y valida RS256, JWKS Apple, emisor, audiencia, expiración,
  subject y nonce. ES256 firma el client secret con la clave privada del servidor.
- State y nonce aleatorios, cinco minutos, callback de un solo uso. Un secreto
  independiente generado en cliente liga el challenge a la app que lo inició;
  únicamente su digest se envía al crearlo. El retorno no transporta tokens,
  sesiones, códigos del proveedor ni ese secreto.
- Retornos cerrados por plataforma y entorno. Nunca se acepta una URL de retorno
  arbitraria del cliente. La web vuelve a su mismo origen.
- Alta: username, locale y términos aceptados; email verificado para una identidad
  nueva. Una identidad existente se resuelve por subject aunque Apple no vuelva
  a facilitar email. Conflicto de email nunca crea sesión ni vínculo automático.
- Sesión y borrador de torneo se persisten atómicamente como con Google. Apple no
  introduce en este incremento vinculación/desvinculación ni reautenticación Apple
  para administrar métodos; el login no implica autorización de esas capacidades.
- Los tokens de acceso/refresh de Apple no se conservan: solo se solicita acceso
  de identidad. La integración de revocación Apple al borrar la cuenta y su
  revisión de requisitos de tienda queda como gate explícito de distribución iOS.
- Errores seguros, cancelaciones sin banner y rechazo de transporte común. Cada
  endpoint declara y prueba validación, tasa, negocio, límites técnicos y cancelación.

## Validación y límites

Pruebas con claves y proveedor sintéticos comprueban seguridad y persistencia;
no acreditan configuración Apple/Google real, firma de las apps ni aprobación.
Antes de distribuir: registrar IDs/dominos/callback, configurar relay de correo,
revisar revocación de cuenta Apple y probar todos los recorridos con builds firmadas.
Google Android mantiene su gate de cliente OAuth y soporte de custom URI scheme;
no se atribuye al pago de Play Console la activación de Google OAuth.

## Fuentes

- [Apple: autorización](https://developer.apple.com/documentation/signinwithapplerestapi/request-an-authorization-to-the-sign-in-with-apple-server)
- [Apple: intercambio de código](https://developer.apple.com/documentation/signinwithapplerestapi/generate-and-validate-tokens)
- [Expo WebBrowser](https://docs.expo.dev/versions/latest/sdk/webbrowser/)
- [Google: restricciones Android](https://developers.google.com/identity/protocols/oauth2/native-app)

## Aclaración de interacción del usuario

El usuario solicita priorizar la presentación interna de autenticación.
Google y Apple usan ASWebAuthenticationSession en iOS y Custom Tabs en Android
mediante Expo; la presentación pertenece al sistema, sobre la app, y vuelve a
la ruta iniciadora. SFSafariViewController no es el adaptador OAuth elegido.
La aclaración posterior acepta el navegador del sistema como alternativa cuando
Android no disponga de Custom Tabs: su ausencia no debe impedir el login. Se
selecciona un paquete compatible si existe y se deja a Expo resolver el navegador
del sistema en caso contrario, conservando el mismo retorno y prueba de sesión.

## Consecuencias y revisión

Se incorpora golang-jwt/jwt/v5 5.3.1 para firma/verificación JWT mantenida;
la criptografía no se reimplementa. JWKS utiliza host Apple fijo, timeout y caché
con límite de refresco ante kids desconocidos. La pérdida del proceso cliente
pierde también su prueba: el intento se reinicia, sin recuperar credenciales por URL.

Revisar si se requiere SDK nativo, gestión de métodos Apple, revocación o si el
proveedor cambia el contrato. Documentos afectados: Identidad, Arquitectura,
Producto, cliente README, OpenAPI, Observabilidad, Despliegue y Aprendizaje.
