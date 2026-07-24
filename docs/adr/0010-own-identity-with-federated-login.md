# ADR-0010: Gestionar identidad propia con login federado

- **Estado:** Aceptado
- **Fecha:** 2026-07-24
- **Decisor:** Usuario
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

TournamentsManager necesita registro, login, verificación de correo,
recuperación de contraseña y sesiones. Además debe permitir autenticarse al menos
con Apple y Google sin perder un usuario interno único ni la autorización sobre
torneos.

## Contexto y restricciones

- El backend es un monolito modular en Go.
- Web, iOS y Android comparten un cliente universal.
- Crear o unirse a un torneo requiere cuenta.
- La autorización de negocio pertenece a TournamentsManager.
- Las credenciales, tokens y flujos de vinculación son una superficie crítica.
- Apple puede ocultar el correo real y los correos pueden cambiar.
- Un usuario puede utilizar varios métodos de acceso para una sola cuenta.
- No se implementará criptografía propia.

## Criterios de decisión

1. control del usuario interno y de la autorización;
2. seguridad verificable;
3. aprendizaje de autenticación y sesiones;
4. compatibilidad con email/password, Apple y Google;
5. experiencia coherente en web, iOS y Android;
6. capacidad de recuperar, vincular y revocar accesos;
7. coste de operación y mantenimiento.

## Alternativas

### Alternativa A — Identidad exclusivamente local

El backend Go almacena credenciales locales, verifica correo, recupera
contraseñas y gestiona sesiones.

- **Ventajas:** control y aprendizaje máximos sobre el flujo local.
- **Inconvenientes:** no satisface login con Apple y Google.
- **Mantenimiento:** alto en seguridad de credenciales y sesiones.

### Alternativa B — Proveedor de identidad gestionado

Un servicio externo gestiona credenciales, sesiones y federación; el backend
mantiene el usuario de negocio.

- **Ventajas:** controles maduros y menor superficie propia.
- **Inconvenientes:** dependencia, coste y menor aprendizaje del núcleo de
  identidad.
- **Mantenimiento:** bajo o moderado según personalización e integración.

### Alternativa C — Identidad propia federada

TournamentsManager gestiona usuarios, credenciales locales, recuperación y
sesiones. Apple y Google actúan como proveedores externos de autenticación que
se vinculan al mismo usuario interno.

- **Ventajas:** control del ciclo completo; login local y social; una sesión y
  autorización propias; alto valor de aprendizaje.
- **Inconvenientes:** combina el riesgo de credenciales locales con la complejidad
  de OAuth/OIDC, vinculación y revocación.
- **Mantenimiento:** alto y permanente.

### Alternativa D — Solo identidades externas

Apple y Google son los únicos métodos de acceso; no existen contraseñas locales.

- **Ventajas:** no se almacenan contraseñas.
- **Inconvenientes:** no satisface el registro y recuperación local aceptados;
  dependencia total de proveedores.
- **Mantenimiento:** moderado.

## Comparación

La alternativa B reduce el riesgo operativo y sería la opción más conservadora
para un equipo pequeño. La alternativa C es la única que combina el aprendizaje
y control buscados con email/password y login social, a cambio de asumir
explícitamente una superficie de seguridad mayor.

## Recomendación

**Opinión profesional:** utilizar identidad gestionada con usuario y autorización
internos minimiza el riesgo.

**Recomendación condicionada al objetivo de aprendizaje:** aceptar identidad
propia federada solo con threat model previo, librerías establecidas, pruebas de
abuso y revisión específica antes de producción.

## Decisión del usuario

Adoptar identidad propia federada:

- TournamentsManager gestiona en Go usuarios, credenciales locales,
  verificación, recuperación y sesiones.
- Apple y Google serán proveedores federados iniciales.
- La aplicación conservará un usuario interno independiente de los proveedores.
- Varias identidades podrán vincularse al mismo usuario.
- El backend emitirá su propia sesión después de cualquier autenticación válida.

## Modelo de identidad

El identificador interno de usuario será independiente de email y proveedor.
Conceptualmente:

```text
User
├── local credential: email/password
├── external identity: apple + subject
└── external identity: google + subject
```

Una identidad externa se identifica por el par `(issuer/provider, subject)`.
Nunca por email.

## Reglas de confianza

- El frontend inicia el flujo y transmite la credencial, código o token al
  backend.
- El backend verifica firma, emisor, audiencia, expiración, nonce y propiedades
  exigidas por el proveedor.
- El backend extrae el `subject` solo de una credencial verificada.
- Un `subject`, email o indicador `email_verified` enviado como texto por el
  frontend no es confiable.
- El email es un canal o atributo verificable, no la identidad externa.
- Cambiar el correo de contacto no modifica ni reemplaza un vínculo Apple/Google.
- Los secretos de proveedor y claves de sesión nunca se incorporan al cliente.

## Vinculación de cuentas

- No se vincularán cuentas automáticamente solo por coincidencia de email.
- Si ya existe una cuenta local, el usuario debe demostrar acceso actual mediante
  contraseña o un desafío nuevo de un solo uso enviado al canal verificado.
- Una verificación histórica de correo no sustituye esa prueba fresca.
- Vincular o desvincular exige reautenticación.
- No se puede eliminar el último método de acceso utilizable.
- Si un nuevo correo ya pertenece a otra cuenta, se inicia un flujo explícito de
  vinculación o recuperación; no se fusionan datos silenciosamente.

## Aclaración de vinculación — 2026-07-24

Cuando un proveedor autentica al usuario pero todavía no existe el vínculo y hay
una cuenta local candidata, el login social queda en
`pending_email_confirmation`.

- Se crea un intento temporal ligado al usuario candidato, proveedor y `subject`
  verificado.
- Se envía un enlace de un solo uso al correo ya verificado de la cuenta local.
- El estado pendiente no constituye una sesión ni concede autorización.
- El vínculo externo se crea únicamente al consumir correctamente el enlace.
- Confirmación, consumo del intento y creación del vínculo serán atómicos.
- Un intento caducado, cancelado o consumido no puede modificar la cuenta.
- La confirmación usará una URL HTTPS asociada a web, iOS y Android.
- Si la aplicación está instalada y asociada, el sistema operativo podrá abrir el
  target mediante Universal Link o App Link; la web será el fallback.
- El token de sesión nunca viajará en la URL.
- Tras confirmar se creará una sesión nueva. Si ya había otra sesión local, el
  cliente realizará un switch automático solo después de recibir la nueva.
- Una sesión existente nunca cambia de usuario; se sustituye por otra sesión.
- Las sesiones de otros dispositivos no se ven afectadas por el switch.

### Aclaración de confirmación y destino — 2026-07-24

El enlace no apuntará a la home ni ejecutará la vinculación mediante un `GET`.
Usará una ruta HTTPS dedicada y transitoria:

```text
https://<base-url>/auth/link/confirm?token=<opaque-token>
```

- La ruta `GET` abre el cliente asociado o la web y presenta el estado de
  confirmación; no consume el intento, no vincula la identidad y no crea sesión.
- Una acción explícita de confirmación inicia una petición `POST` al backend. La
  ruta y el contrato concretos de esa operación se definirán contract-first en
  OpenAPI; el token se enviará en el cuerpo, no en otra URL.
- El backend valida y consume el intento, crea el vínculo y emite una sesión
  nueva dentro de los límites atómicos que se definan para el caso de uso.
- Solo después del éxito, el cliente reemplaza la navegación por la home `/` y
  elimina el token del historial visible.
- Un token inválido, caducado, cancelado o ya consumido muestra un estado de error
  recuperable; no se redirige silenciosamente a la home.
- `token` es el nombre conceptual del parámetro. Nunca contiene una sesión,
  email, proveedor ni identificador de usuario.
- La URL base se obtiene de configuración confiable del entorno; no se construye
  desde un encabezado `Host` recibido.
- La pantalla de confirmación usará `Referrer-Policy: no-referrer`, evitará
  recursos de terceros y no registrará el token.

Separar apertura y mutación conserva la semántica segura de `GET` y evita que la
inspección, previsualización o repetición de un enlace produzca una vinculación o
un cambio de sesión involuntarios.

## Consecuencias

### Positivas

- Una única identidad de producto soporta varios métodos de acceso.
- El dominio y la autorización no dependen de Apple ni Google.
- Un usuario puede cambiar su correo sin perder su identidad federada.
- El backend normaliza todos los métodos en una sesión propia.

### Negativas y deuda aceptada

- TournamentsManager asume contraseñas, sesiones, recuperación y federación.
- Habrá configuraciones y credenciales distintas por target y entorno.
- Se necesitan flujos de vinculación, desvinculación, revocación y recuperación.
- Cambios de proveedor, cuentas transferidas y clientes desactualizados requieren
  procedimientos operativos.
- El lanzamiento exige revisión de seguridad específica.

## Validación

Antes de producción deberá existir evidencia de:

- threat model de registro, login, vinculación, recuperación y cambio de email;
- almacenamiento de contraseña con algoritmo y parámetros aceptados;
- respuestas resistentes a enumeración y límites frente a abuso;
- verificación negativa de tokens con firma, issuer, audience, nonce o expiración
  incorrectos;
- vinculación imposible sin prueba fresca de la cuenta existente;
- revocación de sesiones y proveedores;
- auditoría de acciones sensibles sin registrar secretos ni tokens;
- apertura repetida del enlace mediante `GET` sin consumir el intento ni cambiar
  ninguna cuenta;
- consumo mediante `POST` de un solo uso, resistente a concurrencia y repetición;
- eliminación del token de la navegación y del historial tras el éxito;
- pruebas de compatibilidad en web, iOS y Android.

## Decisiones pendientes

- formato del identificador interno;
- hashing y política de contraseñas;
- verificación y recuperación por email;
- modelo de sesión para web y aplicaciones;
- flujo OAuth/OIDC concreto y librerías;
- almacenamiento de identidades externas y sesiones;
- políticas de vinculación, desvinculación, cambio de email y borrado;
- MFA y passkeys;
- tratamiento de revocaciones y notificaciones de proveedores.

## Disparadores de revisión

- Un incidente o hallazgo de seguridad.
- El mantenimiento impide cumplir objetivos de seguridad u operación.
- Requisitos regulatorios o de assurance superiores.
- Incorporación de muchos proveedores.
- Necesidad de SSO empresarial, MFA avanzada o gestión administrativa.

## Documentación afectada

- [IDENTITY.md](../engineering/IDENTITY.md)
- [SECURITY.md](../engineering/SECURITY.md)
- [ARCHITECTURE.md](../engineering/ARCHITECTURE.md)
- [PRODUCT.md](../project/PRODUCT.md)
- [TECHNICAL_BASELINE.md](../governance/TECHNICAL_BASELINE.md)
- [SYSTEM_OPTIONS.md](../governance/SYSTEM_OPTIONS.md)

## Fuentes técnicas

- [RFC 9110: métodos seguros](https://www.rfc-editor.org/rfc/rfc9110.html#name-safe-methods)
- [OWASP: Forgot Password Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Forgot_Password_Cheat_Sheet.html)
