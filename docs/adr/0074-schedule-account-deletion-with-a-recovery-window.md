# ADR-0074: Programar la eliminación de cuenta con ventana de recuperación

- **Estado:** Aceptado
- **Fecha:** 2026-08-04
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

Una persona necesita solicitar la eliminación de su cuenta desde Datos de acceso
sin dejar sesiones utilizables ni relaciones personales activas. La eliminación
física no debe ser inmediata: debe existir una ventana limitada para una futura
recuperación y una purga operativa posterior. Una liga no puede quedar sin
organizador.

## Contexto y restricciones

- `leagues.organizer_account_id` usa `ON DELETE RESTRICT`; seguimientos y
  administraciones delegadas son relaciones separadas (ADR-0034 y ADR-0058).
- ADR-0068 ya define logout local y el diálogo compartido. Esta baja no exige
  reautenticación, por decisión explícita del usuario.
- La API es REST/OpenAPI contract-first y la mutación web por cookie usa CSRF.
- Los 30 días son una decisión de producto, no un plazo legal universal. Se
  revisarán con la política de conservación aplicable antes de producción.
- Recuperación y purga quedan fuera de este corte.

## Criterios de decisión

1. Invalidar inmediatamente acceso y relaciones personales activas.
2. No destruir ligas ni dejar una liga visible sin organizador.
3. Comunicar la fecha prevista de eliminación definitiva.
4. No adelantar transferencia, recuperación ni un job de purga.

## Alternativas

### A — Baja lógica programada, bloqueo por ligas propias

`DELETE /v1/me/account` marca la cuenta `deletion_pending` y fija
`deletion_requested_at`. La fecha efectiva es 30 días después y se devuelve en
la confirmación. La misma transacción elimina o revoca sesiones y tokens, y
elimina seguimientos y administraciones delegadas. Credenciales e identidades
permanecen inaccesibles para permitir una recuperación futura. Si la cuenta
organiza una liga, devuelve `409 account_has_owned_leagues` sin modificar nada.

- **Ventajas:** evita pérdida definitiva accidental, conserva propiedad y no
  adelanta cancelación, transferencia ni purga.
- **Inconvenientes:** añade un estado y fecha al ciclo de vida de cuenta.
- **Coste de adopción:** medio.
- **Coste de mantenimiento:** bajo hasta introducir recuperación y purga.

### B — Eliminación física inmediata, incluida la propiedad de ligas

- **Ventajas:** no deja datos de cuenta pendientes.
- **Inconvenientes:** destruye datos compartidos de competición y no recupera.
- **Coste de adopción:** bajo técnicamente, alto en riesgo de producto.
- **Coste de mantenimiento:** medio por soporte y restauraciones manuales.

### C — Transferir automáticamente las ligas propias

- **Ventajas:** permite completar la baja de un organizador.
- **Inconvenientes:** exige sucesor, consentimiento, autorización, excepciones
  y auditoría; la feature no existe.
- **Coste de adopción y mantenimiento:** alto.

### No cambiar

La cuenta solo puede cerrar sesión; no puede solicitar su baja ni retirar sus
relaciones de seguimiento y administración.

## Comparación

B sacrifica datos compartidos y C anticipa decisiones de dominio. A preserva la
propiedad, elimina acceso de inmediato y limita este corte a una transición de
estado verificable.

## Recomendación

**Opinión/recomendación:** alternativa A, la solución mínima suficiente.

## Decisión del usuario

**Aceptada el 2026-08-04:** se adopta A con una ventana de recuperación de
**30 días**, sin reautenticación. La respuesta confirma la fecha efectiva; el
cliente muestra un banner localizado, borra su estado de sesión y reconstruye
la navegación anónima igual que ante una sesión caducada.

## Consecuencias

- `accounts` incorpora `deletion_pending` y `deletion_requested_at`; este último
  es obligatorio para ese estado y `NULL` para cuentas activas.
- Una cuenta pendiente no autentica, no restaura sesión ni ejecuta operaciones.
- `DELETE /v1/me/account` exige sesión y CSRF por cookie, devuelve `200` con
  `deletionEffectiveAt`; `409 account_has_owned_leagues` es el único error de
  negocio con banner específico. El resto mantiene el fallback seguro común.
- Datos de acceso añade un control destructivo subrayado y alineado a la derecha
  bajo los items. Su diálogo explica los efectos y los 30 días.
- Cancelar o transferir una liga y las rutas de recuperación/purga se decidirán
  en ADR posteriores.

## Validación

- Una cuenta sin ligas propias recibe `200` y una fecha calculada por el servidor
  exactamente 30 días posterior a la solicitud.
- La transacción retira sesiones/tokens, seguimientos y administraciones, pero
  conserva cuenta, credenciales e identidades con la marca pendiente.
- La sesión iniciadora y las demás dejan de ser válidas de inmediato.
- Una cuenta con liga propia recibe `409`, no pierde nada y ve el banner.
- Tras éxito, las tres plataformas descartan secretos, resetean navegación y
  muestran un banner localizado sin exponer detalles internos.

## Disparadores de revisión

- Política de conservación o consulta legal que cambie los 30 días.
- Implementación de recuperación, purga, transferencia o cancelación de ligas.
- Requisito de reautenticación, MFA o notificación fuera de banda.

## Ampliación aceptada — 2026-08-12

La ventana de 30 días concluye con una purga física automática. El usuario ha
aceptado un comando interno de backend, ejecutado una vez al día por un
LaunchAgent del Mac sobre el proyecto `tournaments-manager-dev`; no se expone un
endpoint HTTP ni se introduce un temporizador dentro de la API.

La purga selecciona como máximo 100 cuentas `deletion_pending` con la fecha
vencida, las bloquea con `FOR UPDATE SKIP LOCKED` y las elimina mediante la base
de datos. Una cuenta que organizase una liga sigue sin poder solicitar la baja.
Las claves foráneas eliminan credenciales, sesiones, tokens, identidades y
relaciones personales. El historial de cambios de resultado se conserva, pero
su autora pasa a `NULL` con `ON DELETE SET NULL`; no se retiene la identidad de
una cuenta eliminada para preservar la auditoría deportiva.

La recuperación no está implementada todavía. Hasta que exista, los textos de
producto y de privacidad describirán una espera de 30 días antes de la
eliminación definitiva, no una capacidad de recuperar la cuenta.

## Documentación afectada

- `docs/governance/DECISIONS.md`
- `docs/project/PRODUCT.md`
- `docs/engineering/IDENTITY.md`
- `docs/engineering/INITIAL_DATA_MODEL.md`
- `docs/engineering/API.md`
- `docs/project/LEARNING.md`
- `contracts/openapi/v1/openapi.yaml`
- `CHANGELOG.md`
