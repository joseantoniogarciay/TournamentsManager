# ADR-0068: Gestionar métodos de acceso y cierre local de sesión

- **Estado:** Aceptado
- **Fecha:** 2026-08-01
- **Decisor:** Usuario, mediante aceptación explícita de la alternativa A
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** Ninguno

## Problema

Una persona con sesión necesita ver los datos mínimos de acceso de su cuenta,
añadir o cambiar una contraseña, vincular Google a su propia cuenta y cerrar la
sesión actual. El sistema aún no define el contrato de reautenticación de esas
mutaciones ni implementa la revocación de la sesión actual.

## Contexto y restricciones

- ADR-0067 permite añadir métodos solo desde `Cuenta > Seguridad`, tras una
  reautenticación reciente; no hay fusión ni traslado de identidades.
- La sesión es opaca, revocable y se transporta por cookie `HttpOnly` en web o
  Bearer en móvil. La cookie solo puede expirar de forma fiable desde la
  respuesta del backend.
- La sesión actual solo contiene `id` y `username`; el email y los métodos no
  se deben inferir en el cliente.
- La persona ha solicitado que logout sea silencioso, no espere la red y que la
  navegación reconstruida permanezca en `Cuenta`.
- El contrato OpenAPI sigue siendo la fuente de verdad; backend y cliente se
  implementan después de aceptarlo.

## Criterios

1. Evitar que una sesión robada añada o reemplace credenciales persistentes.
2. Mantener una única cuenta y prohibir mover identidades entre cuentas.
3. No exponer secretos, métodos de terceros ajenos ni errores internos.
4. Resolver logout de inmediato en la interfaz sin sacrificar la revocación
   remota cuando la red esté disponible.
5. Añadir el mínimo de estado compartido y de UI reutilizable.

## Alternativas

### A — Ticket de reautenticación de un solo uso, limitado a sesión y operación

`GET /v1/me/access-methods` devuelve el email, username y una lista mínima de
métodos presentes (`password`, `google`). Una operación sensible empieza con
una reautenticación del método ya vinculado: contraseña actual o prueba Google
validada contra una identidad que pertenezca a la misma cuenta. El backend emite
un ticket opaco, solo almacenado como huella, de un solo uso, válido cinco
minutos y ligado a la sesión y a la cuenta actuales.

`PUT /v1/me/local-credential` consume ese ticket y establece una contraseña
nueva; la misma operación sirve para añadir o cambiarla. `POST
/v1/me/google-identities` consume un ticket y una prueba Google con challenge,
y enlaza únicamente una identidad nueva o ya vinculada a la misma cuenta. Ambos
flujos invalidan el ticket y no cambian la sesión actual. Los 401, 403, 409 y
errores no reconocidos solo reciben el feedback seguro común, salvo que el
contrato defina una recuperación distinta.

`DELETE /v1/sessions` revoca la sesión presentada y expira la cookie cuando
corresponde. El cliente inicia esa petición en segundo plano con la credencial
capturada, borra inmediatamente su sesión local y se queda en `/account`; no
muestra éxito ni error. Si el envío falla, el secreto local permanece borrado y
la revocación remota se resolverá por expiración o en el siguiente uso.

- **Ventajas:** un mismo límite simple para contraseña y Google; el ticket no
  reutiliza la sesión como autorización suficiente; la UI no conoce secretos ni
  reglas de propiedad; logout responde inmediatamente.
- **Inconvenientes:** añade tabla, hash, endpoints y una segunda confirmación
  antes de cada mutación; el logout sin espera no garantiza entrega si no hay
  red.
- **Coste de adopción:** medio.
- **Coste de mantenimiento:** bajo o medio; el ticket puede reutilizarse para
  futuras mutaciones sensibles.

### B — Autorizar las mutaciones con la sesión vigente

La sesión autenticada llama directamente a cambiar/añadir credenciales y
vincular Google.

- **Ventajas:** menos endpoints y una experiencia más corta.
- **Inconvenientes:** una sesión robada permite tomar persistencia en la cuenta;
  contradice el requisito de reautenticación reciente de ADR-0067.
- **Coste de adopción:** bajo.
- **Coste de mantenimiento:** medio por excepciones de seguridad posteriores.

### C — Reautenticar mediante un login completo que rota la sesión

Cada cambio usa el flujo de login existente, crea otra sesión y realiza después
la mutación.

- **Ventajas:** evita el nuevo tipo de ticket.
- **Inconvenientes:** mezcla autenticación, navegación y mutación; puede dejar
  sesiones innecesarias y duplica diferencias entre cookie y Bearer.
- **Coste de adopción:** medio.
- **Coste de mantenimiento:** medio o alto.

### No cambiar

Cuenta sigue sin salida de sesión ni gestión de métodos; Google y contraseña
solo se resuelven desde recorridos públicos incompletos.

## Comparación

La B es demasiado permisiva para cambiar autenticadores. C aparenta reutilizar
infraestructura, pero aumenta los estados de sesión y complica la navegación.
A separa prueba reciente de posesión, autorización de la operación y sesión de
producto, conservando el modelo opaco ya aceptado.

## Recomendación

**Recomendación:** alternativa A. Cinco minutos y un solo uso son límites
pequeños, comprensibles y suficientes para este primer corte. No se introduce
MFA, fusión de cuentas, administración de dispositivos ni reintentos de logout:
serían sobreingeniería sin necesidad demostrada.

## Decisión del usuario

**Aceptada el 2026-08-01:** alternativa A. Se usarán tickets opacos de
reautenticación de un solo uso, con vida de cinco minutos y ligados a la sesión
y cuenta actuales. Las mutaciones de contraseña y la vinculación Google los
consumen. El cierre de sesión borra el estado local inmediatamente, intenta la
revocación sin bloquear y deja la navegación en Cuenta.

## Consecuencias si se acepta

- La pantalla `Cuenta` autenticada muestra una lista `Datos de acceso` con
  email, username, contraseña y Google, y una fila destructiva de cerrar sesión.
- Un diálogo compartido de confirmación ofrece título, descripción, aceptar
  primario y cancelar secundario con borde. El logout lo usa antes de ejecutar
  su acción silenciosa.
- El modelo persiste tickets de reautenticación solo como huellas y los invalida
  al consumirlos, expirar la sesión o cerrar sesión.
- La cuenta conserva al menos un método de acceso; no se implementa todavía su
  eliminación, por lo que esta regla no crea una UI nueva.

## Validación

- Un ticket caducado, repetido, de otra sesión o de otra cuenta no muta nada.
- Una contraseña actual o una identidad Google ya vinculada a la misma cuenta
  puede crear el ticket; una identidad de otra cuenta no.
- Añadir Google nunca cambia `account_id`; repetir el mismo vínculo es idempotente.
- Añadir o cambiar contraseña respeta la política vigente y no registra el
  secreto.
- Logout elimina de inmediato el secreto Bearer local o el estado de sesión,
  intenta revocar el backend sin bloquear y la pantalla resultante es Cuenta sin
  sesión. La API revoca la sesión y expira la cookie cuando recibe la petición.
- Web, iOS y Android tienen textos localizados, controles accesibles y no
  muestran errores internos.

## Disparadores de revisión

- Un requisito de MFA, passkeys, dispositivos administrados o mayor assurance.
- Evidencia de que cinco minutos es insuficiente o que la reautenticación frena
  una tarea habitual.
- Necesidad demostrada de eliminar métodos o de revocar todas las sesiones.

## Documentación afectada

- `docs/engineering/IDENTITY.md`
- `docs/engineering/SECURITY.md`
- `docs/engineering/API.md`
- `docs/project/PRODUCT.md`
- `docs/project/LEARNING.md`
- `docs/governance/DECISIONS.md`
- `contracts/openapi/v1/openapi.yaml`
- `CHANGELOG.md`

## Fuentes técnicas

- [NIST SP 800-63B: reauthentication and password verifiers](https://pages.nist.gov/800-63-4/sp800-63b.html)
- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
- [OWASP Session Management Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html)
