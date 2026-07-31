# ADR-0044: Usar sesiones opacas, SMTP local y Argon2id

- **Estado:** Aceptado
- **Fecha:** 2026-07-26
- **Decisor:** Usuario, mediante aceptación explícita de la propuesta original
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** ADR-0018, exclusivamente al añadir Mailpit como dependencia local
- **Superado por:** ADR-0062, exclusivamente para la duración y renovación de
  sesiones.

## Problema

El primer incremento de backend necesita registrar cuentas locales, verificar su
email, mantener la sesión en web y móvil, y revocarla de forma inmediata. Debe
hacerlo sin almacenar contraseñas recuperables, secretos de sesión en logs o URLs,
ni proveedores reales de correo antes de validar el flujo.

## Contexto y restricciones

- ADR-0010 fija identidad propia con sesiones emitidas por el backend.
- ADR-0031 exige cuentas pendientes sin sesión ni permisos hasta verificar el
  email, conservando el borrador asociado.
- ADR-0043 limita el primer incremento a identidad local, publicación y lectura
  de liga.
- PostgreSQL local y el ciclo de Compose están validados; las aplicaciones se
  ejecutan en el host conforme a ADR-0018.
- Apple, Google, recuperación de contraseña, MFA y sesiones multi-dispositivo
  administrables no entran en este incremento.

## Criterios de decisión

1. revocación inmediata de sesión y autorización siempre actual;
2. renovación silenciosa coherente entre web y móvil;
3. contraseñas no recuperables, resistentes al cracking offline;
4. verificación manual local mediante un email realista sin credenciales cloud;
5. coste operativo y complejidad proporcionales a un monolito inicial.

## Alternativas

### Alternativa A — Sesiones opacas persistidas, Argon2id y SMTP local

El backend emite secretos aleatorios opacos y guarda solo su huella en
PostgreSQL. La web usa una cookie segura; móvil usa el secreto como Bearer en
almacenamiento seguro. Mailpit captura el correo SMTP exclusivamente en el
entorno local.

- **Ventajas:** revocación y rotación inmediatas; una fuente de verdad de sesión;
  flujo completo de email sin proveedor real; mínima infraestructura adicional.
- **Inconvenientes:** cada petición autenticada consulta o valida estado en
  PostgreSQL; Mailpit añade un servicio local de desarrollo.
- **Coste de adopción:** medio.
- **Coste de mantenimiento:** bajo o medio.
- **Riesgos:** rotaciones concurrentes o almacenamiento móvil incorrecto; se
  mitigan mediante una única capa cliente de sesión y pruebas de revocación.

### Alternativa B — JWT de acceso y refresh token

- **Ventajas:** validación local de access tokens sin consultar sesiones; encaja
  con consumidores autónomos y varios servicios.
- **Inconvenientes:** dos credenciales, rotación, revocación y sincronización de
  permisos añaden complejidad sin una necesidad distribuida demostrada.
- **Coste de adopción:** medio o alto.
- **Coste de mantenimiento:** medio o alto.
- **Riesgos:** JWT válido tras logout hasta expirar si no se introduce estado de
  revocación; secretos expuestos por almacenamiento o renovación defectuosa.

### Alternativa C — JWT de larga duración o sesión solo en memoria

- **Ventajas:** menos tablas o flujo inicial muy corto.
- **Inconvenientes:** no permite logout fiable, persiste credenciales de riesgo o
  fuerza login frecuente; no resuelve verificación local de correo.
- **Coste de adopción:** bajo.
- **Coste de mantenimiento:** medio por incidentes y excepciones futuras.
- **Riesgos:** sesiones robadas difíciles de revocar o experiencia frágil.

### No cambiar

La implementación de identidad queda bloqueada y no se puede demostrar el
recorrido aceptado de cuenta pendiente a publicación autorizada.

## Comparación

La B es razonable si varios servicios o terceros deben validar tokens sin una
base de sesiones compartida. En el monolito, la autorización ya consulta datos
de negocio en PostgreSQL y la revocación inmediata aporta más valor. La C evita
trabajo inicial a cambio de controles insuficientes. La A cubre el flujo con una
sola credencial de sesión y una dependencia local acotada.

## Recomendación

**Opinión/recomendación:** alternativa A.

## Decisión del usuario

**Aceptada el 2026-07-26:** adoptar la alternativa A, con estas reglas:

- contraseñas con Argon2id, mínimo 19 MiB de memoria, 2 iteraciones y
  paralelismo 1; nunca se cifran ni almacenan en claro;
- secretos de sesión opacos, aleatorios y de al menos 128 bits; PostgreSQL
  almacena solo su huella, usuario, creación, último uso, expiración y revocación;
- expiración por inactividad de 7 días y límite absoluto de 30 días;
- rotación al login, verificación de email, reautenticación y cambio futuro de
  contraseña; logout revoca la sesión actual;
- web entrega la sesión mediante cookie `__Host-`, `Secure`, `HttpOnly`,
  `SameSite=Lax`, `Path=/` y sin `Domain`;
- móvil usa Bearer y guarda el secreto exclusivamente mediante Keychain/Keystore;
  su gestor de sesión renueva de forma serializada al entrar en primer plano o
  antes de caducar;
- los tokens de verificación son aleatorios, opacos, de un solo uso, duran 24
  horas y se almacenan solo como huella;
- Mailpit se añade a Compose solo para desarrollo, publicado exclusivamente en
  loopback; captura el SMTP y no existe en producción;
- un entorno no local falla al arrancar si no tiene un adaptador de correo
  configurado explícitamente;
- registro, login y reenvío no enumeran cuentas y aplican límites por IP e
  identificador normalizado;
- no se introducen JWT, refresh tokens, Apple/Google, recuperación de contraseña
  ni MFA en este incremento.

## Consecuencias

### Positivas

- Cerrar sesión y revocar una sesión produce efecto inmediato.
- La experiencia puede renovar la sesión sin pedir contraseña de nuevo.
- El flujo de verificación se prueba de extremo a extremo en local sin secretos
  de proveedores externos.
- Las contraseñas no pueden recuperarse desde la base de datos.

### Negativas y deuda aceptada

- PostgreSQL participa en cada autenticación de petición y debe mantenerse
  disponible.
- Mailpit amplía Compose con una dependencia estrictamente de desarrollo.
- Recuperación, vinculación federada y cierre global de sesiones esperan a
  incrementos posteriores.

## Validación

- Un hash Argon2id no permite recuperar la contraseña y el login solo funciona
  con la entrada correcta.
- Una cuenta pendiente no recibe sesión ni puede publicar; un enlace válido la
  verifica una sola vez y crea una sesión.
- Logout, revocación y expiración bloquean inmediatamente o al vencimiento las
  peticiones posteriores.
- Web recibe los atributos de cookie acordados; móvil renueva y sustituye el
  secreto de forma serializada en almacenamiento seguro.
- Mailpit recibe el enlace de desarrollo y no expone su UI fuera de loopback.
- Las respuestas de alta, login y reenvío no revelan la existencia de una cuenta.

## Disparadores de revisión

- Varios servicios autónomos o consumidores externos necesitan validar tokens
  sin consultar la sesión.
- La latencia o disponibilidad de PostgreSQL hace insuficiente la validación de
  sesión actual.
- Se necesita SSO, dispositivos administrados, MFA o revocación global.
- Mailpit deja de ser suficiente para probar el flujo local de correo.

## Documentación afectada

- [IDENTITY.md](../engineering/IDENTITY.md)
- [SECURITY.md](../engineering/SECURITY.md)
- [DEVELOPMENT.md](../engineering/DEVELOPMENT.md)
- [local-postgresql.md](../runbooks/local-postgresql.md)
- [DECISIONS.md](../governance/DECISIONS.md)
- [DECISIONS_TO_REVISIT.md](../governance/DECISIONS_TO_REVISIT.md)
- [LEARNING.md](../project/LEARNING.md)
- [CHANGELOG.md](../../CHANGELOG.md)

## Fuentes técnicas

- [OWASP Password Storage Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html)
- [OWASP Session Management Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html)
