# ADR-0095: Transferir directamente la propiedad de un torneo

- **Estado:** Aceptado
- **Fecha:** 2026-08-14
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

Una organizadora necesita poder entregar un torneo a otra cuenta sin cancelar la
competición ni dejar una propiedad huérfana. Esto también permite resolver, de
forma explícita, el bloqueo de baja de una cuenta que aún organiza ligas.

## Contexto y restricciones

- Hoy el único formato persistido es una liga, aunque la interfaz usa «torneo»
  porque brackets y formatos mixtos podrán llegar después.
- El buscador público ya devuelve únicamente cuentas verificadas por `username`.
- La organizadora es la única que puede gestionar equipos, ciclo y
  administradores; las administradoras delegadas solo gestionan resultados.
- No se introduce aceptación, invitación, correo, push ni una abstracción de
  torneo genérica antes de que existan otros formatos.

## Alternativas

### A — Transferencia directa e inmediata

La organizadora elige una cuenta verificada distinta mediante el buscador. En
una transacción, la destinataria pasa a ser organizadora, deja de ser
administradora delegada si lo era, la anterior organizadora pierde toda
administración y recibe la destinataria una notificación interna.

- **Ventajas:** resuelve la continuidad y la baja sin estados pendientes;
  conserva partidos, resultados y el resto de administraciones.
- **Inconvenientes:** la transferencia es irreversible en este corte; la cuenta
  elegida recibe la propiedad sin aceptación.
- **Coste de adopción y mantenimiento:** bajo a moderado.

### B — Solicitud con aceptación

- **Ventajas:** evita asignar propiedad no solicitada.
- **Inconvenientes:** requiere estados, caducidad, recordatorios y gestión de
  solicitudes antes de haber demostrado esa necesidad.
- **Coste de adopción y mantenimiento:** alto.

### No cambiar

La organizadora debe cancelar la liga para poder eliminar su cuenta o conservar
la cuenta indefinidamente.

## Recomendación

**Opinión/recomendación:** A es la mínima solución suficiente y reutiliza la
delegación directa ya aceptada.

## Decisión del usuario

**Aceptada el 2026-08-14:** se adopta A. La acción se denomina «Transferir
torneo», opera sobre ligas en este corte y está disponible en cualquier estado.
La anterior organizadora no conserva permisos delegados automáticamente. La UI
vuelve al detalle y confirma el resultado con un banner tras la navegación.

## Consecuencias

- `POST /leagues/{leagueId}/transfer` exige a la organizadora y recibe el
  `username` de la destinataria.
- La mutación serializa la liga, valida ambas cuentas y cambia la propiedad
  atómicamente; no altera resultados, equipos ni otras administraciones.
- La destinataria recibe una notificación interna durable de transferencia.
- La futura abstracción de torneos podrá conservar el nombre de la acción sin
  obligar a generalizar tablas o rutas hoy.

## Validación

- Solo la organizadora puede transferir; una cuenta inexistente, no verificada
  o la propia organizadora no modifica nada.
- Transferir a una administradora existente no deja una relación delegada
  duplicada.
- Tras la transacción, la antigua organizadora no puede ejecutar acciones de
  organizadora y la nueva sí, incluso con una liga en curso.
- El buscador cierra al detalle y el banner se encola con
  `showAfterNavigation()` para aparecer con el host de la ruta enfocada.

## Disparadores de revisión

- Se exige consentimiento, vencimiento o revocación de la transferencia.
- Existen brackets u otros formatos persistidos y se puede extraer una operación
  común sin ocultar reglas específicas.
- Se necesitan auditoría visible, correo o push.

## Documentación afectada

- `docs/governance/DECISIONS.md`
- `docs/project/PRODUCT.md`
- `docs/engineering/API.md`
- `contracts/openapi/v1/openapi.yaml`
- `docs/project/LEARNING.md`
