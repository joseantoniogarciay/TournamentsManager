# Identidad y acceso

> Estado: identidad propia federada aceptada; sesiones locales opacas,
> verificación SMTP local y contraseñas Argon2id aceptadas en ADR-0044.

## Vocabulario

- **Usuario interno:** identidad estable de TournamentsManager.
- **Cuenta pendiente:** registro temporal de alta local que aún no puede iniciar
  una sesión de producto ni ejecutar acciones de negocio.
- **Cuenta con baja programada:** cuenta sin sesión ni operaciones de producto,
  retenida durante 30 días desde `deletion_requested_at` antes de su purga
  definitiva conforme a ADR-0074. La recuperación aún no está implementada.
- **Credencial local:** email y secreto de autenticación gestionados por el
  backend.
- **Identidad externa:** vínculo con un proveedor mediante `issuer` y `subject`.
- **Autenticación:** demostrar quién inicia la sesión.
- **Sesión:** credencial emitida por TournamentsManager tras autenticar.
- **Autorización:** decidir qué puede hacer el usuario sobre un torneo.
- **Email de contacto:** canal verificado que puede cambiar sin alterar el
  identificador interno ni los vínculos externos.
- **Username:** identificador público, único e inmutable inicialmente de una
  cuenta verificada. Sirve para seleccionar usuarios sin exponer su email.

## Decisión

[ADR-0010](../adr/0010-own-identity-with-federated-login.md) establece identidad
propia en Go con credenciales locales y login federado inicial mediante Apple y
Google.

ADR-0044 fija sesiones opacas persistidas y revocables, y ADR-0062 las completa
con access y refresh opacos rotatorios. La web conserva ambos secretos en cookies
persistentes `HttpOnly`; móvil usa Bearer en almacenamiento seguro. [ADR-0050](../adr/0050-include-google-federated-login-in-first-increment.md)
incorpora Google como único proveedor federado inicial. No se introducen JWT ni
Apple en este incremento.

```text
Email / password ──> local_credentials ─┐
                                         ├──> accounts ──> sessions
Google ──> external_identities ──────────┘

Antes de entregar una credencial Google, el cliente solicita un challenge de
cinco minutos y usa el nonce devuelto al iniciar Google. El backend consume ese
challenge una sola vez tras validar el ID token; no es una sesión ni una cuenta.

Google solo se habilita en los entornos públicos `dev` y `prod`; el entorno
local no configura clientes ni acepta audiencias Google. El cliente web de
desarrollo registra `https://dev.fasttourney.com` como origen y
`https://dev.fasttourney.com/account` como URI de redirección. El layout raíz
resuelve el popup al regresar para que la ventana que lo inició conserve el
recorrido de Cuenta. iOS y Android usan clientes OAuth nativos separados, asociados
respectivamente al identificador de bundle o paquete y, en Android, a la huella
SHA-1 del certificado que firma la build. Los IDs de cliente son públicos: el
artefacto público de desarrollo los incorpora al exportarse y el backend admite
sus audiencias mediante `GOOGLE_CLIENT_IDS`.
```

El cliente transporta credenciales y usa la sesión resultante. No decide la
identidad ni la autorización.

## Username público

La [ADR-0048](../adr/0048-require-username-at-registration-and-rotate-verification.md)
establece que toda cuenta aporta un `username` público, único y en minúsculas al
crear su identidad. En un primer acceso con Apple o Google se elige después de
acreditar la identidad con el proveedor y antes de crear la cuenta. No forma
parte de una credencial, no sustituye al identificador interno y no se puede
cambiar en el primer corte.

Se usa para buscar y seleccionar administradores de una liga. Las reglas exactas
de formato, normalización, nombres reservados y un futuro cambio de `username`
se decidirán antes de implementarlas.

El cliente puede consultar `GET /v1/usernames/{username}/availability` cuando
el valor ya cumple el mínimo de tres caracteres y permanece sin cambios durante
400 ms. La respuesta informa del estado actual, no reserva el nombre: el alta
vuelve a aplicar la restricción única de PostgreSQL. Para contener sondeo o
enumeración, el endpoint admite 30 consultas por IP y minuto en cada proceso y
responde `429` con `Retry-After` al superar ese límite. La política se revisará
antes de escalar la API horizontalmente, pues el límite local no se comparte
entre réplicas.

## Alta local y borradores antes del acceso

Un invitado puede preparar un borrador de torneo en el cliente sin autenticarse.
Al enviar un alta con email, contraseña, `username` y locale efectivo, el backend
crea una cuenta pendiente y asocia el borrador a ella. El locale se valida contra
los idiomas soportados y se guarda como preferencia de la cuenta para localizar
el email de verificación y futuros emails. La verificación del correo activa la
cuenta; solo entonces se emite una sesión de producto y se permite publicar el
torneo.

La cuenta pendiente no es una sesión, ni concede autorización. Cuenta y borrador
caducan a los siete días y se eliminan mediante una purga explícita. Esta
decisión no crea persistencia de borradores anónimos en el servidor. Véase
[ADR-0031](../adr/0031-preserve-pre-auth-tournament-drafts-until-verified.md).

El contrato de registro, verificación y sesión se concreta en
[OpenAPI v1](../../contracts/openapi/v1/openapi.yaml); el modelo persistente está
en [INITIAL_DATA_MODEL.md](INITIAL_DATA_MODEL.md). La verificación consume el
token de un solo uso, activa la cuenta completa y crea la sesión en una única
transacción. ADR-0061 establece que el cliente la inicia automáticamente al
abrir el deep link, después de retirar el token de la URL, y reemplaza la sesión
preexistente si la hubiera. Un login correcto de cuenta pendiente invalida el token activo y
solicita otro correo, sin crear sesión.

## Subject de Apple y Google

En un login federado, el frontend puede recibir un identificador y un token, pero
el backend no confía en campos sueltos.

1. El cliente inicia el login con el proveedor.
2. El proveedor devuelve un código o token firmado y, según plataforma, datos de
   la credencial.
3. El cliente transmite esos artefactos al backend.
4. El backend verifica criptográficamente y comprueba issuer, audience,
   expiración y nonce.
5. El backend extrae el `subject` verificado.
6. Se busca el vínculo `(provider, subject)` y se emite una sesión propia.

En el primer incremento se implementa exclusivamente Google. El backend trata
`sub` como el identificador externo estable, no el email, y valida la credencial
antes de crear o consultar el vínculo. Apple reutilizará la misma frontera en un
incremento posterior.

Apple utiliza identificadores con alcance del equipo de desarrollo; las
aplicaciones correctamente agrupadas pueden correlacionar al mismo usuario. La
configuración exacta y las migraciones de equipo se tratarán como parte del
adaptador Apple.

## Email real tras usar Apple

El email Apple —real o relay— y el `subject` cumplen funciones diferentes. Si el
usuario quiere registrar después su email real:

1. inicia sesión mediante su vínculo Apple existente;
2. propone un nuevo email de contacto;
3. el backend envía un desafío de verificación a ese email;
4. tras verificarlo, actualiza el canal de contacto;
5. el vínculo `(apple, subject)` permanece sin cambios.

Si el email ya corresponde a otro usuario, se deniega el cambio: no se fusionan
ni se vinculan cuentas distintas.

## Primera entrada con Google

Una identidad externa se resuelve únicamente por `(issuer, subject)`:

```text
Google verificado
      │
      ├── ya existe (google, subject) ──> login en su cuenta
      │
      └── no existe
             │
             ├── email no usado ─────────> alta Google con username
             └── email ya usado ─────────> denegar; usar el método existente
```

No se envía un desafío de vinculación ni se crea sesión cuando una identidad
nueva declara un email que ya pertenece a otra cuenta. El cliente muestra un
aviso genérico para iniciar sesión con el método habitual y añadir otro acceso
desde `Cuenta > Seguridad`; no revela qué proveedor usa una cuenta existente.

Desde Seguridad, una persona con sesión y reautenticación reciente puede añadir
una contraseña o una identidad social todavía no vinculada a ninguna cuenta. Si
el `(issuer, subject)` ya pertenece a otra cuenta, la operación falla sin mover
ni duplicar la identidad; si pertenece a la misma cuenta, no crea un duplicado.
No existe fusión de cuentas: las ligas, administraciones y seguimientos
permanecen ligados al ID de la cuenta que los creó. Véanse ADR-0066 y ADR-0067.
La API y la interfaz de Seguridad requieren primero el mecanismo común de
reautenticación reciente; por eso no forman parte todavía del endpoint de inicio
de Google implementado en este incremento.

## Invariantes de seguridad

- Los sujetos externos no son identificadores del dominio.
- Email no es clave de una identidad federada.
- Toda entrada del cliente es no confiable.
- Un `(issuer, subject)` ya vinculado no se puede mover a otra cuenta; solo una
  adición idempotente a su misma cuenta puede aceptarse.
- Un email ya perteneciente a otra cuenta no permite crear, vincular ni fusionar
  cuentas.
- Debe permanecer al menos un método de acceso.
- Una cuenta con baja programada no puede autenticar hasta que se decida y
  complete un flujo de recuperación.
- La recuperación local no crea ni vincula proveedores.
- Los mensajes públicos no enumeran cuentas.
- Tokens, códigos y secretos no se escriben en logs.

## Recuperación de contraseña local

ADR-0064 incorpora una solicitud no enumeradora que entrega por email un enlace
HTTPS con token opaco, de un solo uso y de 30 minutos. El cliente elimina el
token de URL e historial antes de mostrar el email de solo lectura y pedir la
nueva contraseña. El consumo por `POST` sustituye el hash Argon2id y revoca las
otras sesiones y crea una sesión nueva para el dispositivo que completó el
cambio. Las contraseñas manuales admiten de 8 a 1024 caracteres y el
cliente declara `new-password` al gestor de credenciales para que sugiera y
guarde contraseñas largas.

## Fuentes técnicas

- [Apple: verificar un usuario](https://developer.apple.com/documentation/signinwithapple/verifying-a-user)
- [Apple: identificadores con alcance de equipo](https://developer.apple.com/documentation/signinwithapple/bringing-new-apps-and-users-into-your-team)
- [Google: referencia OpenID Connect](https://developers.google.com/identity/openid-connect/reference)
- [OpenID Connect Core](https://openid.net/specs/openid-connect-core-1_0.html)
- [RFC 9110: métodos seguros](https://www.rfc-editor.org/rfc/rfc9110.html#name-safe-methods)
- [OWASP: Forgot Password Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Forgot_Password_Cheat_Sheet.html)
# Métodos de acceso administrados

ADR-0094 completa la gestión de Google y contraseña: una cuenta conserva siempre
al menos un método. Crear la primera contraseña desde una cuenta solo Google
requiere una prueba Google reciente; cambiar una contraseña exige una credencial
actual. Para retirar un método se acredita el otro que permanecerá. Los tickets
opacos de reautenticación son de un uso, viven cinco minutos y solo autorizan la
finalidad con que se emitieron.

ADR-0068 separa la sesión de producto de una prueba reciente de posesión. Una
persona autenticada obtiene un ticket opaco, almacenado únicamente como hash,
con cinco minutos de vida, ligado a su sesión y consumido por una única
mutación: establecer la contraseña local o vincular Google. La reautenticación
acepta la contraseña Argon2id vigente o una identidad Google ya vinculada a la
misma cuenta; nunca usa el email para unir cuentas.
