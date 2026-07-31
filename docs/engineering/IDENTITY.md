# Identidad y acceso

> Estado: identidad propia federada aceptada; sesiones locales opacas,
> verificación SMTP local y contraseñas Argon2id aceptadas en ADR-0044.

## Vocabulario

- **Usuario interno:** identidad estable de TournamentsManager.
- **Cuenta pendiente:** registro temporal de alta local que aún no puede iniciar
  una sesión de producto ni ejecutar acciones de negocio.
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

Para el primer incremento, [ADR-0044](../adr/0044-use-opaque-sessions-local-smtp-and-argon2id.md)
fija sesiones opacas persistidas y revocables. La web usa cookie segura; móvil
usa Bearer en almacenamiento seguro. [ADR-0050](../adr/0050-include-google-federated-login-in-first-increment.md)
incorpora Google como único proveedor federado inicial. JWT, refresh tokens y
Apple quedan fuera de este incremento.

```text
Email / password ──> local_credentials ─┐
                                         ├──> accounts ──> sessions
Google ──> external_identities ──────────┘

Antes de entregar una credencial Google, el cliente solicita un challenge de
cinco minutos y usa el nonce devuelto al iniciar Google. El backend consume ese
challenge una sola vez tras validar el ID token; no es una sesión ni una cuenta.
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

Si el email ya corresponde a otro usuario, no se actualiza ni fusiona
automáticamente: se exige demostrar acceso a ambas cuentas.

## Primera entrada con Google para una cuenta local existente

Una coincidencia de email es una señal para iniciar vinculación, no autorización
para completarla.

```text
Google verificado
      │
      ├── ya existe (google, subject) ──> login
      │
      └── no existe
             │
             ├── no hay cuenta candidata ──> alta
             └── existe email local ───────> prueba fresca
                                                  │
                                                  ├── contraseña actual
                                                  └── enlace/código de un solo uso
```

Después de la prueba fresca se vincula `(google, subject)` al usuario interno. A
partir de entonces puede autenticarse con ambos métodos.

La validación enviada al registrar la cuenta demostró control del email en aquel
momento. La vinculación es una acción sensible posterior y exige control actual.

## Estado pendiente de vinculación

Autenticar correctamente con Google o Apple no concede acceso si el proveedor
todavía no está vinculado y existe una cuenta local candidata. Se crea un intento
de vinculación, no una sesión de usuario.

```text
provider_authenticated
          │
          ▼
pending_email_confirmation
       │             │
       │             ├── caduca / se cancela ──> expired
       │
       └── enlace válido
                 │
                 ▼
               linked
                 │
                 ▼
           session_eligible
```

Mientras está `pending_email_confirmation`:

- no existe todavía la identidad externa definitiva;
- no se emite una sesión normal ni se autorizan acciones;
- el intento queda ligado al usuario candidato, proveedor y `subject`
  previamente verificado;
- el enlace contiene un token aleatorio, con caducidad y de un solo uso;
- se almacena una representación no reutilizable del token, no el secreto en
  claro.

Abrir el enlace no modifica la cuenta. La confirmación explícita hace que el
backend compruebe token, estado y expiración. En una operación atómica crea el
vínculo si no existe conflicto y consume el intento. Si el enlace caduca o ya fue
usado, no se modifica ninguna cuenta.

## Deep link y establecimiento de sesión

El enlace de confirmación será una URL HTTPS del producto:

- iOS podrá abrirla como Universal Link si la aplicación está instalada y la
  asociación o preferencia del usuario lo permite;
- Android podrá abrirla como App Link verificado si la aplicación está instalada
  y asociada al dominio;
- en cualquier otro caso se resolverá en la aplicación web.

No se usarán custom schemes como mecanismo primario. El dominio y las
aplicaciones demostrarán su asociación mediante los mecanismos de plataforma.

El enlace transporta únicamente el token opaco y de un solo uso del intento. No
contiene access tokens, refresh tokens ni identificadores de sesión.

```text
GET /auth/link/confirm?token=...
          │
          ├── iOS associated ────> Universal Link
          ├── Android associated ─> App Link
          └── fallback ──────────> Web
          │
          ▼
show blocking transition
          │
          ▼
POST confirmation to backend
          │
          ▼
consume attempt + link identity + create session
          │
          ▼
replace web URL with home / or reset native navigation
```

La ruta del enlace es una pantalla transitoria del cliente, no el endpoint REST
que modifica la cuenta. Su forma conceptual es:

```text
https://<base-url>/auth/link/confirm?token=<opaque-token>
```

`GET` es seguro: puede validar lo necesario para presentar la pantalla, pero no
consume el intento, no vincula la identidad y no crea una sesión. Esto protege el
flujo frente a aperturas repetidas, previsualizaciones e inspecciones automáticas
del enlace.

Mientras el cliente confirma el enlace y reemplaza la sesión, una capa global
bloquea la interacción para que no quede visible ni operable el estado de la
identidad anterior. Tras éxito, la web reemplaza la URL por `/`; las aplicaciones
reconstruyen las raíces de Inicio, Torneos y Cuenta, descartando modales y pilas
previas. Cada raíz carga sus datos al recibir foco, no como efecto del reset.

La persona confirma mediante una acción explícita. El cliente realiza entonces
un `POST` al backend; la ruta y los DTO concretos se incorporarán a OpenAPI antes
de implementarse y el token se enviará en el cuerpo de la petición. El backend
valida y consume el intento una sola vez, vincula la identidad y emite la nueva
sesión. Las peticiones concurrentes o repetidas no pueden producir más de un
consumo válido.

Tras confirmar y crear el vínculo:

- sin sesión previa, se crea la sesión del usuario vinculado;
- si la sesión previa pertenece al mismo usuario, se mantiene la identidad y se
  rota la sesión cuando corresponda;
- si pertenece a otro usuario, el cliente sustituye automáticamente su sesión
  local por una sesión nueva del usuario vinculado.

No se modifica el `user_id` de una sesión existente. El backend crea una sesión
nueva y la sustitución solo ocurre después de haberla emitido correctamente. Las
sesiones de otros dispositivos no se cambian por este switch.

Después del éxito, web y aplicaciones sustituyen la ruta de confirmación por la
home `/`; no añaden otra entrada al historial y el token deja de estar visible.
Si el token es inválido, ha caducado, fue cancelado o ya se consumió, el cliente
muestra un estado de error y una recuperación posible en vez de redirigir
silenciosamente.

En web, la sesión se entregará mediante el mecanismo seguro que se decida para
cookies. En aplicaciones, el cliente canjeará un resultado de confirmación por la
sesión mediante la API. Los secretos de sesión no aparecerán en URLs, historial,
analytics ni logs.

La URL base se configura por entorno y nunca se deriva de un encabezado `Host`
no confiable. La pantalla de confirmación evitará recursos de terceros, no
registrará el token y usará `Referrer-Policy: no-referrer`.

## Invariantes de seguridad

- Los sujetos externos no son identificadores del dominio.
- Email no es clave de una identidad federada.
- Toda entrada del cliente es no confiable.
- Vincular, desvincular y cambiar email requiere reautenticación.
- Debe permanecer al menos un método de acceso.
- La recuperación local no crea ni vincula proveedores.
- Los mensajes públicos no enumeran cuentas.
- Tokens, códigos y secretos no se escriben en logs.
- Un intento pendiente no es una sesión ni concede permisos.
- Una sesión nunca cambia de propietario; un switch crea y selecciona otra.
- `GET` no consume un intento ni cambia identidad o sesión.
- Solo una confirmación explícita mediante `POST` puede consumir el intento.
- El token se elimina de la navegación tras el éxito.

## Fuentes técnicas

- [Apple: verificar un usuario](https://developer.apple.com/documentation/signinwithapple/verifying-a-user)
- [Apple: identificadores con alcance de equipo](https://developer.apple.com/documentation/signinwithapple/bringing-new-apps-and-users-into-your-team)
- [Google: referencia OpenID Connect](https://developers.google.com/identity/openid-connect/reference)
- [OpenID Connect Core](https://openid.net/specs/openid-connect-core-1_0.html)
- [RFC 9110: métodos seguros](https://www.rfc-editor.org/rfc/rfc9110.html#name-safe-methods)
- [OWASP: Forgot Password Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Forgot_Password_Cheat_Sheet.html)
