# Identidad y acceso

> Estado: arquitectura propia federada aceptada; protocolos, almacenamiento y
> controles concretos pendientes.

## Vocabulario

- **Usuario interno:** identidad estable de TournamentsManager.
- **Credencial local:** email y secreto de autenticación gestionados por el
  backend.
- **Identidad externa:** vínculo con un proveedor mediante `issuer` y `subject`.
- **Autenticación:** demostrar quién inicia la sesión.
- **Sesión:** credencial emitida por TournamentsManager tras autenticar.
- **Autorización:** decidir qué puede hacer el usuario sobre un torneo.
- **Email de contacto:** canal verificado que puede cambiar sin alterar el
  identificador interno ni los vínculos externos.

## Decisión

[ADR-0010](../adr/0010-own-identity-with-federated-login.md) establece identidad
propia en Go con credenciales locales y login federado inicial mediante Apple y
Google.

```text
Apple / Google ── credencial firmada ──┐
                                       ▼
Email / password ────────────────> Backend Go
                                       │
                                       ├── autentica
                                       ├── resuelve User interno
                                       └── emite sesión propia
```

El cliente transporta credenciales y usa la sesión resultante. No decide la
identidad ni la autorización.

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
- el enlace contiene un token aleatorio, expirado y de un solo uso;
- se almacena una representación no reutilizable del token, no el secreto en
  claro.

Al abrir el enlace, el backend comprueba token, estado y expiración. En una
operación atómica crea el vínculo si no existe conflicto y consume el intento. Si
el enlace caduca o ya fue usado, no se modifica ninguna cuenta.

La forma de devolver la sesión al dispositivo original —continuación segura,
deep link o repetir el login social— se decidirá junto al modelo de sesión. La
regla ya aceptada es que nunca habrá sesión antes de confirmar el vínculo.

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

## Fuentes técnicas

- [Apple: verificar un usuario](https://developer.apple.com/documentation/signinwithapple/verifying-a-user)
- [Apple: identificadores con alcance de equipo](https://developer.apple.com/documentation/signinwithapple/bringing-new-apps-and-users-into-your-team)
- [Google: referencia OpenID Connect](https://developers.google.com/identity/openid-connect/reference)
- [OpenID Connect Core](https://openid.net/specs/openid-connect-core-1_0.html)
