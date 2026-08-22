# ADR-0105: Separar la fiabilidad esencial del cliente de la analítica opcional de producto

- **Estado:** Aceptado
- **Fecha:** 2026-08-22
- **Decisor:** Usuario, mediante decisión explícita
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** ADR-0060 y ADR-0102, solo en el requisito de consentimiento para excepciones y crashes del cliente
- **Superado por:** Ninguno

## Problema

Un crash nativo o una excepción no controlada impide prestar correctamente el
servicio, pero el opt-in de analítica de producto vigente también desactiva esa
señal. Así se pierde el diagnóstico de los fallos de quienes rechazan medir su
uso de la interfaz.

## Contexto y restricciones

- PostHog Cloud sigue limitado a la región UE, a la beta pública
  `development` y a gasto máximo de 0 €; `local` y `production` continúan
  bloqueados.
- OpenTelemetry, Prometheus, Grafana, Loki y Tempo siguen siendo la fuente de
  diagnóstico técnico del backend.
- El SDK no identifica cuentas y no recibe credenciales, tokens, formularios,
  cuerpos HTTP, URLs reales, IDs de entidades ni eventos de producto sin
  consentimiento.
- Las excepciones pueden contener datos imprevistos en su mensaje o stack. La
  configuración efectiva del proyecto, su DPA, retención, filtrado y la base
  jurídica aplicable requieren revisión del responsable antes de distribuir la
  build; este ADR no sustituye dicha revisión.

## Alternativas

### A — Mantener un único opt-in para todo PostHog

Es la implementación anterior.

- **Ventaja:** límite de privacidad muy fácil de explicar.
- **Inconveniente:** se pierde el diagnóstico de fallos de quienes rechazan la
  analítica, por lo que no sirve al objetivo de fiabilidad.

### B — Dos proveedores SaaS separados

Usar un error tracker solo para crashes y otro para analítica.

- **Ventaja:** aislamiento fuerte de finalidades.
- **Inconvenientes:** dos SDK, dos contratos de privacidad, símbolos y costes
  operativos; es complejidad prematura para la primera beta.

### C — Un cliente PostHog con dos canales lógicos

Inicializar PostHog únicamente para error tracking mínimo en la beta pública y
habilitar la fachada de producto solo tras el consentimiento revocable.

- **Ventajas:** conserva la señal de fiabilidad, no añade proveedor ni
  identidad, y mantiene fuera del consentimiento vistas, resultados de producto,
  replay, autocapture y correlación con API.
- **Inconveniente:** requiere una política de privacidad y configuración de
  cuenta más precisas, y no permite retirar la captura mínima mediante el
  switch de analítica.
- **Coste de mantenimiento:** bajo: una instancia y dos límites explícitos en
  el código y la documentación.

## Decisión del usuario

**Aceptada:** alternativa C.

En `development`, si existe la clave pública configurada, el cliente inicializa
PostHog para capturar únicamente excepciones JavaScript no controladas,
rechazos de promesas y crashes nativos. Desactiva replay, autocapture, eventos
de ciclo de vida, carga/evaluación remota de flags y resolución GeoIP. No
identifica la cuenta ni añade propiedades de producto.

El switch existente conserva una única finalidad: habilita o deshabilita la
analítica opcional de producto. Antes de aceptarlo no hay `screen_viewed`,
eventos de resultados, eventos de red ni cabeceras `X-Interaction-ID`; al
retirarlo se detienen sin apagar la captura mínima de fiabilidad.

## Validación

1. Sin consentimiento, una build distribuible de iOS y otra de Android envían
   un fallo JavaScript controlado y un crash nativo simbolizados para su
   release, sin replay ni eventos de navegación/producto.
2. En web, una excepción no controlada se registra sin iniciar autocapture ni
   vistas; el rechazo del switch no impide su captura.
3. Al activar el switch aparecen únicamente los eventos consentidos de ADR-0103
   y ADR-0104; al desactivarlo vuelven a detenerse.
4. El proyecto de PostHog está en región UE, gasto máximo 0 €, con filtrado,
   retención y documentación de privacidad revisados antes de distribuir la
   build.

## Consecuencias

- Se reduce el riesgo de no poder diagnosticar un fallo real de cliente.
- La preferencia deja de prometer que desactiva todo PostHog; solo controla la
  analítica de producto y debe expresarse así en la política de privacidad.
- No se añade un segundo proveedor ni se altera el diagnóstico del backend.

## Documentación afectada

- [README.md](../../README.md)
- [DECISIONS.md](../governance/DECISIONS.md)
- [DECISIONS_TO_REVISIT.md](../governance/DECISIONS_TO_REVISIT.md)
- [OBSERVABILITY.md](../operations/OBSERVABILITY.md)
- [LEARNING.md](../project/LEARNING.md)
- [CHANGELOG.md](../../CHANGELOG.md)
- [ADR-0060](0060-use-posthog-for-deferred-client-product-observability.md)
- [ADR-0102](0102-activate-posthog-before-k3s.md)
