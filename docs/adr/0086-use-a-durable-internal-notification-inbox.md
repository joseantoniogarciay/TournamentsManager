# ADR-0086: Usar un buzón interno de notificaciones duradero

- **Estado:** Aceptado
- **Fecha:** 2026-08-11
- **Decisor:** Usuario
- **Propietario del análisis:** Codex
- **Supera a:** ADR-0034, solo respecto al aviso de una nueva administración delegada
- **Superado por:** Ninguno

## Problema

Una cuenta debe saber que se le ha asignado la administración de una liga, incluso
si no estaba usando la aplicación en ese momento.

## Alternativas

### A — Buzón interno persistente

Se guarda una notificación por cuenta y se muestra desde Cuenta.

- Ventajas: funciona entre sesiones y dispositivos; coste bajo; el alta de la
  delegación y el aviso comparten transacción.
- Inconvenientes: no avisa fuera de la aplicación.
- Coste de mantenimiento: bajo.

### B — Push, outbox y worker desde el inicio

- Ventajas: entrega externa fiable.
- Inconvenientes: obliga a decidir permisos, dispositivos, preferencias,
  proveedor, reintentos y operación antes de necesitarlos.
- Coste de mantenimiento: alto.

## Decisión del usuario

**Aceptada el 2026-08-11:** alternativa A. Al crear una administración delegada
nueva se crea una notificación interna durable. Abrir el listado las marca todas
como leídas; se pueden eliminar individualmente o en bloque y pulsarlas abre la
liga. Las reasignaciones idempotentes no duplican avisos.

Push, tokens de dispositivo, preferencias y outbox/worker quedan aplazados. Una
decisión posterior podrá añadirlos al mismo evento sin sustituir el buzón.

## Aclaración — 2026-08-11

El contador se refresca al restaurar o crear una sesión. Los dos mecanismos de
actualización posterior son excluyentes: sin permiso push, al volver la app a
primer plano si han pasado al menos cinco minutos desde la última lectura; con
permiso push activo, exclusivamente al recibir una push. La recepción disparará
la misma actualización inmediata del contador. El indicador de la tab Cuenta
será un punto, nunca el número.

## Validación

- Una nueva asignación crea exactamente un aviso en la misma transacción.
- El contador baja a cero al abrir el buzón.
- Borrar una cuenta no afecta al buzón de otra.
