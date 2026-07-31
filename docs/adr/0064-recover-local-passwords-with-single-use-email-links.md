# ADR-0064: Recuperar contraseñas locales con enlaces de email de un solo uso

- **Estado:** Aceptado
- **Fecha:** 2026-07-31
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** ADR-0044, exclusivamente al incorporar la recuperación de contraseña local y su política de longitud
- **Superado por:** ADR-0065, exclusivamente para la sesión emitida tras completar el restablecimiento

## Problema

Una persona que no recuerda su contraseña local no puede volver a acceder a su
cuenta. La pantalla Cuenta necesita ofrecer una recuperación segura, sin revelar
si un email tiene cuenta y sin que la aplicación maneje directamente el llavero o
el gestor de contraseñas del dispositivo.

## Contexto y restricciones

- ADR-0044 excluyó la recuperación de contraseña del primer incremento; esta
  propuesta incorpora ese alcance y conserva Argon2id, tokens opacos y correo
  local mediante Mailpit para desarrollo.
- El producto define “recordar contraseña” como establecer una nueva, nunca
  recuperar la anterior.
- La web, iOS y Android reciben el enlace HTTPS como web, Universal Link o App
  Link, siguiendo el límite de confianza y las reglas de deep link de
  `IDENTITY.md`.
- El cliente no guarda ni escribe credenciales en Keychain, Keystore ni un
  gestor de contraseñas. Declara la semántica de los campos para que el sistema
  o proveedor elegido por la persona proponga guardar o actualizar la entrada.
- El contrato OpenAPI continuará siendo la fuente de verdad y todo endpoint se
  implementará primero en backend y después mediante el adaptador de feature
  del cliente.
- El mínimo actual de 12 caracteres cambia tanto en OpenAPI como en la
  validación de backend y cliente. La decisión de permitir 8 caracteres escritos
  manualmente es una concesión de experiencia frente a la guía NIST vigente para
  contraseñas de un solo factor, que recomienda un mínimo de 15.

## Criterios de decisión

1. recuperar acceso sin exponer contraseñas ni enumerar cuentas;
2. impedir que un enlace filtrado se reutilice y limitar su ventana de abuso;
3. ofrecer una experiencia clara en web, iOS y Android, incluido el guardado o
   reemplazo de la credencial por el proveedor del sistema;
4. revocar credenciales potencialmente comprometidas sin introducir una nueva
   dependencia de identidad;
5. mantener lógica de dominio independiente de SMTP, URLs y plataformas.

## Alternativas

### Alternativa A — Enlace HTTPS opaco de un solo uso y nueva contraseña

La persona solicita recuperación con su email. La API devuelve la misma respuesta
aceptada para cualquier email y, solo si existe una credencial local activa, crea
un intento con token aleatorio, opaco, almacenado únicamente como huella, de un
solo uso y con una duración de 30 minutos. El email contiene una URL HTTPS con el
token. El cliente lo extrae, lo elimina inmediatamente de URL e historial y envía
el token en el cuerpo de un `POST` que establece la nueva contraseña.

La pantalla de nueva contraseña muestra el email asociado como `TextField`
deshabilitado con teclado/semántica de email y el campo nuevo como
`new-password`. Registro usa la misma semántica. Login usa `current-password`.
Los tres campos de contraseña ofrecen el control accesible de mostrar u ocultar
el valor. Registro y recuperación muestran un medidor orientativo de fortaleza;
el backend acepta 8 a 1024 caracteres, mientras que el gestor del sistema recibe
`new-password` y puede sugerir una contraseña de 15 caracteres o más.

Al completar el cambio, una transacción consume el intento, sustituye el hash
Argon2id y revoca todas las demás familias de sesión de la cuenta. Si la petición
presenta una sesión vigente de esa misma cuenta, se rota y conserva únicamente
esa sesión del dispositivo que completó el enlace; si no presenta una sesión,
el resultado no emite una nueva y la persona inicia sesión con la contraseña
nueva.

- **Ventajas:** flujo conocido, no precisa contraseña anterior, limita la vida
  del secreto y fuerza el cierre de las otras sesiones potencialmente
  comprometidas;
  `new-password` habilita la sugerencia y el guardado/actualización gestionados
  por el sistema.
- **Inconvenientes:** depende de que la persona controle el buzón y la obliga a
  iniciar sesión de nuevo; 8 caracteres manuales reducen la resistencia frente a
  ataques offline respecto al mínimo NIST actual.
- **Coste de adopción:** medio: modelo persistente de intentos, endpoints,
  plantillas localizadas, deep link, pantallas y pruebas de seguridad.
- **Coste de mantenimiento:** bajo o medio: los cuatro locales y la política de
  token se mantienen junto al flujo de identidad.
- **Riesgos:** filtración del enlace, previsualizadores de email y abuso por
  solicitudes masivas; se mitigan con POST, consumo atómico, `Referrer-Policy:
no-referrer`, ausencia de recursos de terceros, límites por IP/email y una
  respuesta no enumeradora.

### Alternativa B — Código temporal introducido manualmente

El correo incluye un código y el cliente pide email, código y contraseña nueva.

- **Ventajas:** el secreto no aparece en la URL y puede transcribirse entre
  dispositivos.
- **Inconvenientes:** añade pasos, errores de transcripción y dos campos más;
  no satisface la experiencia de deep link solicitada.
- **Coste de adopción:** medio.
- **Coste de mantenimiento:** medio por UX adicional, reintentos y soporte.
- **Riesgos:** los códigos siguen siendo secretos y requieren los mismos
  controles de expiración, límite y no enumeración.

### Alternativa C — Cambio desde una sesión autenticada

La persona solo puede cambiar contraseña desde ajustes tras presentar la actual.

- **Ventajas:** no añade correo ni tokens de recuperación.
- **Inconvenientes:** no resuelve el olvido de contraseña ni el acceso perdido.
- **Coste de adopción:** bajo.
- **Coste de mantenimiento:** bajo.

### No cambiar

La Cuenta no ofrece recuperación; perder la contraseña equivale a perder el
acceso local a la cuenta.

## Comparación

La alternativa A cubre el problema con una capacidad de identidad estándar y
reutiliza el canal SMTP, las asociaciones HTTPS y el patrón de token opaco ya
existentes. B mejora la ausencia de token en URL, pero empeora la experiencia
solicitada sin eliminar los controles de seguridad. C y no cambiar no recuperan
el acceso. El medidor es informativo: no sustituye validación del backend ni una
futura lista de contraseñas comprometidas.

## Recomendación

**Recomendación:** alternativa A, con 30 minutos de vida, consumo atómico,
respuesta no enumeradora y revocación de las demás sesiones. Recomiendo conservar
12 caracteres mínimos, pero la preferencia del usuario es permitir 8 manualmente
y usar el gestor de contraseñas para sugerir 15 o más; se documenta como deuda de
seguridad explícitamente aceptada si se aprueba este ADR.

## Decisión del usuario

**Aceptada el 2026-07-31:** alternativa A, con recuperación por email y deep
link, sugerencia del gestor de contraseñas de 15 caracteres o más, mínimo manual
de 8 con medidor de fortaleza y revocación de las demás sesiones. Si el
dispositivo que completa el enlace ya tiene una sesión de la misma cuenta, esa
sesión se rota y conserva; si no, el cambio no inicia una sesión nueva.

## Consecuencias

### Positivas

- Una cuenta local puede recuperar acceso sin soporte manual.
- iOS, Android y web pueden proponer generar, guardar o actualizar credenciales
  mediante sus gestores, incluso si no existía una entrada previa.
- Una contraseña comprometida deja inutilizables las demás sesiones existentes,
  sin expulsar innecesariamente al dispositivo que completa el cambio.

### Negativas y deuda aceptada

- El mínimo manual de 8 no cumple la recomendación NIST actual para login de un
  solo factor.
- La persona debe volver a iniciar sesión después de cambiar la contraseña.
- Se añaden plantilla de email, datos temporales, límites y pruebas de seguridad.

## Validación

- La solicitud responde igual para emails existentes, inexistentes, federados o
  sin credencial local, y queda sometida a límite por IP e identificador.
- El token es aleatorio, opaco, solo se persiste como huella, vence a los 30
  minutos, no aparece en logs, analytics, Referer ni historial y solo un POST
  puede consumirlo.
- Un intento válido sustituye el hash Argon2id y revoca todas las otras familias
  de sesión; solo rota y conserva la sesión presentada por el dispositivo que
  completa el cambio cuando pertenece a esa misma cuenta. Un token repetido,
  inválido o vencido no cambia nada.
- Login muestra `current-password`; Registro y Nueva contraseña usan
  `new-password`, ojo accesible y sugerencia de gestor de credenciales.
- Registro y Nueva contraseña rechazan menos de 8 caracteres en cliente y
  backend, permiten hasta 1024 y muestran el medidor sin tratarlo como control
  de autorización.
- Las cuatro localizaciones incluyen los textos, feedback y plantilla de email.

## Disparadores de revisión

- Un incidente de credenciales débiles o abuso de recuperación exige elevar el
  mínimo manual o incorporar lista de contraseñas comprometidas.
- Soporte evidencia que la revocación total de sesiones impide recuperar acceso
  de forma razonable.
- Se necesita recuperar cuentas sin acceso al email o se incorpora MFA.

## Documentación afectada

- `docs/engineering/IDENTITY.md`
- `docs/engineering/SECURITY.md`
- `docs/engineering/API.md`
- `docs/project/PRODUCT.md`
- `docs/project/LEARNING.md`
- `CHANGELOG.md`
- `contracts/openapi/v1/openapi.yaml`

## Fuentes técnicas

- [NIST SP 800-63B](https://pages.nist.gov/800-63-4/sp800-63b.html)
- [OWASP: Forgot Password Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Forgot_Password_Cheat_Sheet.html)
- [Apple: Password AutoFill](https://developer.apple.com/documentation/security/enabling-password-autofill-on-a-text-input-view)
- [Android: Autofill hints](https://developer.android.com/reference/androidx/autofill/HintConstants)
