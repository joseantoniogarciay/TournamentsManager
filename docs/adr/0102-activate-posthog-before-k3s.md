# ADR-0102: Activar PostHog antes de iniciar K3s

- **Estado:** Superado
- **Fecha:** 2026-08-22
- **Decisor:** Usuario, mediante decisión explícita
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** ADR-0060, solo en el momento de activación
- **Superado por:** ADR-0105, para el consentimiento de excepciones y crashes mínimos del cliente; ADR-0109, para el orden que bloqueaba K3s

## Problema

ADR-0060 eligió PostHog Cloud para la observabilidad de producto del cliente,
pero aplazó la integración hasta una primera beta distribuida. El usuario decide
ahora priorizar esa señal —en especial errores y crashes de iOS y Android— antes
de comenzar el laboratorio K3s de la Fase 4.

## Contexto y restricciones

- La elección de proveedor, límites de datos y límite de gasto siguen siendo los
  de ADR-0060: PostHog Cloud en la región UE, gasto máximo de 0 € y sin PII,
  credenciales, tokens, cuerpos HTTP ni campos de formularios.
- El cliente ya ofrece una preferencia local, revocable y desactivada por
  defecto. Antes del opt-in no se inicializa el SDK ni se captura señal.
- OpenTelemetry, Prometheus, Grafana, Loki y Tempo conservan el diagnóstico
  técnico del backend; PostHog no recibe ni sustituye sus trazas.
- Replay y crash reporting nativo requieren builds de desarrollo o distribuidas;
  Expo Go no es evidencia suficiente de su funcionamiento.
- La configuración de cuenta, DPA, límites, enmascarado, exclusiones y alertas
  de cuota requiere acceso a la cuenta PostHog del responsable y no puede
  automatizarse con secretos en el repositorio.

## Alternativas

### A — Activar PostHog ahora, antes de K3s

Preparar el proyecto PostHog de la región UE, integrar el SDK con el opt-in ya
existente y verificar errores JavaScript, crashes nativos y replay restringido
en una build distribuible. K3s permanece como la siguiente iniciativa.

- **Ventajas:** proporciona evidencia de experiencia real y reduce el tiempo de
  diagnóstico de la primera beta; no añade dependencia a la operación K3s.
- **Inconvenientes:** añade configuración de privacidad, símbolos y source maps
  antes de profundizar en infraestructura.
- **Coste de adopción:** bajo o medio; se limita por la cuota gratuita y gasto
  máximo de 0 €.
- **Coste de mantenimiento:** bajo mientras el catálogo de eventos y los flags
  siga siendo pequeño.

### B — Mantener el aplazamiento y comenzar K3s

Conservar ADR-0060 sin integración hasta que haya un incidente o una beta
distribuida posterior.

- **Ventajas:** retrasa la dependencia SaaS y su revisión de privacidad.
- **Inconvenientes:** un primer problema de cliente carece de diagnóstico
  específico y no valida la entrega móvil antes del laboratorio K3s.

## Recomendación

**Opinión/recomendación:** alternativa A. La integración conserva el alcance
limitado ya decidido, complementa la observabilidad técnica existente y aporta
un aprendizaje de cliente más inmediato que K3s.

## Decisión del usuario

**Aceptada:** activar PostHog antes de iniciar la Fase 4/K3s. La cuenta se crea
en región UE, se fija el gasto máximo en 0 € y se aplican las salvaguardas de
ADR-0060 antes de habilitar la captura. K3s continúa después de cerrar esta
integración y su retrospectiva técnica.

La primera entrega activa únicamente excepciones JavaScript y la preparación de
crashes nativos tras opt-in. Replay, autocapture de interacción, breadcrumbs,
eventos de producto y flags permanecen desactivados hasta validar su catálogo,
enmascarado, exclusiones y configuración efectiva en la cuenta.

Mientras el plan Free solo admita un proyecto, PostHog se limita al entorno
`development` público que representa la primera beta. El cliente no lo
inicializa en `local` ni en `production`, aunque una configuración accidental
incluya una clave. Antes de producción se decidirá una separación real de datos
y configuración en lugar de mezclar sus señales con la beta.

## Validación

1. El proyecto PostHog está en región UE, con gasto máximo de 0 € y alertas de
   cuota configuradas.
2. La política de privacidad y el consentimiento describen la telemetría que se
   active; antes del opt-in no hay inicialización ni llamadas a PostHog.
3. El catálogo inicial no contiene PII ni datos de formularios, y replay queda
   excluido de registro, acceso, recuperación y demás pantallas sensibles.
4. Una build de iOS y otra de Android simbolizan un fallo JavaScript y un crash
   nativo de prueba para su release o actualización correspondiente.
5. Los límites de correlación, flags y backend se mantienen como en ADR-0060.

## Documentación afectada

- [README.md](../../README.md)
- [DECISIONS.md](../governance/DECISIONS.md)
- [DECISIONS_TO_REVISIT.md](../governance/DECISIONS_TO_REVISIT.md)
- [ROADMAP.md](../project/ROADMAP.md)
- [LEARNING.md](../project/LEARNING.md)
- [ADR-0060](0060-use-posthog-for-deferred-client-product-observability.md)

## Fuentes técnicas

- [PostHog: precios](https://posthog.com/pricing)
- [Expo: usar PostHog](https://docs.expo.dev/guides/using-posthog/)
- [PostHog: React Native](https://posthog.com/docs/libraries/react-native)
