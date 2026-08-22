# ADR-0103: Correlacionar eventos consentidos del cliente con trazas de la API

- **Estado:** Aceptado
- **Fecha:** 2026-08-22
- **Decisor:** Usuario, mediante decisión explícita
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** ADR-0060, solo al concretar la instrumentación de vistas y peticiones
- **Superado por:** Ninguno

## Problema

Una excepción de cliente o una pantalla vista no basta para diagnosticar una
operación que termina en la API. Es necesario partir de una interacción concreta
de PostHog y llegar, sin exportar PII ni construir trazas de larga duración en
el cliente, al log y la traza OpenTelemetry que procesaron su petición.

## Decisión del usuario

**Aceptada:** después del opt-in y solo en `development`, el cliente registra:

| Evento | Propiedades permitidas |
| --- | --- |
| `screen_viewed` | `screen`, nombre canónico y cerrado de la vista |
| `api_request_completed` | `interaction_id`, `method`, `status`, `duration_ms` |
| `api_request_failed` | `interaction_id`, `method`, `failure` (`network_error` o `cancelled`), `duration_ms` |

Cada petición de la API recibe un UUID aleatorio `interaction_id`. El cliente
lo adjunta al evento correspondiente y a la cabecera `X-Interaction-ID`; la API
solo acepta su forma UUID, nunca lo copia a un atributo de métrica o span, y lo
escribe exclusivamente como `interaction_id` en el log estructurado junto a su
propio `trace_id` y `span_id` OpenTelemetry.

`trace_id` es generado por el backend. El cliente no lo fabrica, no intenta
propagar una traza OpenTelemetry ni lo usa como identidad. Las URLs reales,
parámetros, IDs de liga, nombres de usuario, cuerpos, cookies, tokens y errores
brutos quedan fuera de PostHog, logs adicionales, métricas y atributos de span.

## Alternativas descartadas

- **Autocapture de PostHog:** podría registrar textos, controles o rutas que no
  forman parte de un catálogo revisado; se mantiene desactivado.
- **Una traza OTel desde el cliente:** aumenta el acoplamiento y permite
  reconstruir sesiones técnicas más allá del límite aceptado.
- **Solo correlacionar mutaciones:** reduce volumen, pero deja sin diagnóstico
  las cargas y restauraciones de sesión que el usuario ha decidido incluir.

## Validación

1. Sin consentimiento, no hay evento de producto ni cabecera; la captura mínima
   de fiabilidad se rige por ADR-0105.
2. Una ruta dinámica genera solo un nombre de pantalla canónico, sin su ID.
3. Una petición aceptada incluye un UUID nuevo en su evento y cabecera.
4. La API registra un UUID válido solo en su log, junto al `trace_id`; uno
   malformado se ignora y nunca altera la respuesta.
5. Las métricas y spans no contienen `interaction_id`.

## Documentación afectada

- [ADR-0060](0060-use-posthog-for-deferred-client-product-observability.md)
- [ADR-0102](0102-activate-posthog-before-k3s.md)
- [OBSERVABILITY.md](../operations/OBSERVABILITY.md)
- [DECISIONS.md](../governance/DECISIONS.md)
- [LEARNING.md](../project/LEARNING.md)
