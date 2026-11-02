# ADR-0180: Sincronizar participaciones con Google Calendar

- **Estado:** Propuesto
- **Fecha:** 2026-11-02
- **Decisor:** Usuario
- **Supera parcialmente a:** ADR-0158/0161, en la exclusión de calendario.

## Problema y decisiones de producto aceptadas

El usuario pide sincronización automática de su cuenta, no exportación puntual.
Acepta calendario separado FastTourney y sincronización app → Google; solo los
encuentros de su equipo/participante vinculado con fecha y hora completas.
Administrar o seguir no supone participar. Duración estimada configurable,
inicialmente 60 minutos, identificada como estimación, no duración deportiva.

## Alternativas y mantenimiento

- Archivo puntual: coste bajo, no mantiene cambios ni satisface la sincronización.
- Suscripción a feed: coste moderado, depende de refresco de Google y no equivale
  al flujo OAuth solicitado.
- Integración OAuth Calendar API: coste alto inicial y moderado continuo por
  credenciales, expiración, cuota, reintentos, deduplicación y reconciliación.
- Bidireccional: mayor coste y conflictos de autoridad; fuera de esta propuesta.

## Recomendación técnica pendiente de cierre

Integración backend con consentimiento separado del login. Evaluar el scope
calendar.app.created, que permite gestionar calendarios secundarios creados por
la app, en lugar de acceso a todos los calendarios. No vincular cuentas de login
ni fusionar identidades por conectar Calendar. Guardar credenciales renovables
cifradas, con claves fuera de DB y Git. La sincronización debe funcionar sin
mantener un navegador o móvil abierto, con trabajo durable y reintentos acotados;
no llamar a Google dentro de la transacción deportiva.

La propuesta aún requiere concretar contrato, persistencia/clave, operación del
worker, OAuth/callback por plataforma y configuración/verificación de Google.
No implementar estas decisiones técnicas mientras este ADR esté Propuesto.

## Criterios de comportamiento para revisar

Alta inicial y reconciliación sin duplicados por cuenta/partido; reprogramación,
lugar heredado, pérdida de fecha/hora, cancelación, retirada y pérdida de vínculo.
Conservar hechos pasados; definir explícitamente la retirada de eventos futuros
y la política de eventos editados/borrados en Google y desconexión.
Sin invitaciones a rivales ni emails a terceros. Mostrar conexión, última
sincronización confirmada, errores seguros y reconexión. Cancelar consentimiento
no produce una falsa conexión. Fecha/hora/zona deben resolver un instante real:
revisar horas inexistentes y ambiguas por DST antes de crear eventos.

## Fuentes oficiales verificadas

- https://developers.google.com/workspace/calendar/api/auth
- https://developers.google.com/workspace/calendar/api/v3/reference/events/insert
- https://developers.google.com/identity/protocols/oauth2/web-server
- https://developers.google.com/identity/protocols/oauth2/production-readiness/sensitive-scope-verification

Calendar exige inicio y fin. El acceso offline requiere credencial renovable;
verificación y configuración reales son un gate externo, no un detalle de UI.

## Decisión

Requisitos de producto anteriores aceptados por respuesta explícita del usuario.
Diseño técnico y políticas de ciclo de vida pendientes; no se autoriza una
conexión real, subida de datos personales o publicación por este documento.

## Validación y retrospectiva

OAuth cancelado/revocado/expirado, cuentas aisladas, retries y carreras sin
duplicados, caída Google/DB, DST, reprogramación y desconexión; contrato cliente
generado, feedback seguro y salida completa de observabilidad sin secretos/PII.
Pruebas con cuenta controlada y configuración real antes de producción.
La sincronización implica mantenimiento continuo: separarla del login y del
resultado deportivo limita el impacto de una caída de Google.

## Documentación afectada

PRODUCT, IDENTITY, SECURITY, DATABASE, API, OBSERVABILITY, DEPLOYMENT, DECISIONS,
LEARNING, política de privacidad y runbook de sincronización.
