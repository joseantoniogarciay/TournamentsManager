# ADR-0164: Gestionar las sesiones opacas activas

- **Estado:** Aceptado
- **Fecha:** 2026-10-12
- **Decisor:** Usuario: «Ok vamos con eso», tras recomendar conservar sesiones opacas y permitir consultar y revocar sesiones propias.
- **Supera a:** ADR-0068, solo en su aplazamiento de la administración de sesiones.

## Problema y contexto

La persona necesita cerrar accesos de otras instalaciones o navegadores sin cambiar sus autenticadores. ADR-0044 y ADR-0062 ya conservan una sesión estable y refresh tokens asociados en PostgreSQL. El gate técnico está cerrado.

## Criterios y alternativas

1. Revocar acceso y renovación, sin afectar a otra cuenta ni a la sesión actual al cerrar las demás.
2. No exponer secretos, huellas, IP ni User-Agent bruto.
3. Mantener costes y consultas acotados.

A: ampliar sesiones opacas. Reutiliza persistencia, rotación y revocación; requiere contrato, proyección y UI. Mantenimiento bajo o medio.
B: JWT con registro de revocación. Permite el mismo resultado con estado consultado y añade claves, firma y validación de claims; mantenimiento mayor.
C: mantener solo logout local. Menor coste, pero no permite cierre remoto.

## Comparación y recomendación

**Recomendación aceptada:** A. JWT registrado no aporta una ventaja para este monolito. La funcionalidad administra sesiones, no identifica dispositivos físicos.

## Decisión y consecuencias

- Cuenta ofrece una lista paginada de sesiones propias activas (50 por página), con creación, última actividad y marca de sesión actual.
- Plataforma y cliente son categorías orientativas cerradas, obtenidas del User-Agent web o de un hint móvil; nunca autorizan operaciones. Los valores desconocidos se muestran como tales. Sesiones anteriores se enriquecen al volver a autenticar una petición.
- La actividad se actualiza como máximo una vez cada cinco minutos por sesión, salvo captura inicial de categorías; no se almacena IP ni agente bruto.
- Sesión activa significa acceso vigente o refresh sin consumir, sin revocar y vigente. Un access vencido con refresh útil sigue siendo revocable.
- Se puede revocar una sesión ajena a la actual, o todas las demás, previa confirmación en UI. La actual conserva el logout existente. Mutaciones idempotentes: una sesión inexistente o de otra cuenta no revela su existencia.
- El backend revalida y bloquea la sesión solicitante en la mutación, y revoca sesión y todos sus refresh tokens atómicamente. Cerrar las demás afecta a las sesiones presentes en la operación, no impide futuros logins.
- El cierre remoto espera confirmación del servidor; no elimina elementos ante fallo de red. No requiere ticket adicional porque no añade autenticadores persistentes. La sesión revocada pierde autorización en las peticiones posteriores; no se promete cancelar trabajo ya autorizado ni borrar a distancia el almacenamiento local.
- CSRF y límites de tasa protegen estas operaciones. Estados no tratados, cuerpos inválidos y problemas desconocidos usan feedback común seguro; cancelación intencional sin feedback.

## Validación

Aislamiento entre cuentas, persistencia del ID durante refresh, cierre selectivo y de las demás, refresh revocado inutilizable, reintento idempotente, sesión actual preservada, carreras con rotación; contrato, rechazo de validación, tasa, autenticación/CSRF, fallo técnico y cancelación sin PII; cliente generado, parsing, cuatro idiomas, typecheck y exportación web.

## Disparadores de revisión

Necesidad real de identidad de dispositivos, reautenticación/MFA para revocación, volumen medido de sesiones o consumidores que necesiten validación autónoma.

## Documentación afectada

IDENTITY, SECURITY, API, PRODUCT, DATABASE, OBSERVABILITY, DECISIONS, LEARNING y CHANGELOG.
