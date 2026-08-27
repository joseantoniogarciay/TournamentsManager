# ADR-0109: Diferir la validación distribuible de PostHog hasta disponer de cuentas móviles

- **Estado:** Aceptado
- **Fecha:** 2026-08-23
- **Decisor:** Usuario, mediante decisión explícita
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** ADR-0102, solo en el orden que bloqueaba la Fase 4
- **Superado por:** Ninguno

## Problema

ADR-0102 situaba la activación y validación distribuible de PostHog antes de
K3s. Esa prueba requiere las cuentas de distribución de iOS y Android, que no
estarán disponibles previsiblemente hasta el mes siguiente. Esperarlas impide
iniciar un laboratorio local de VM Linux y K3s que no necesita esas cuentas ni
añade gasto externo.

## Contexto y restricciones

- ADR-0101 ya fija una VM Linux ligera de un nodo con K3s para la Fase 4.
- La integración de PostHog permanece preparada y limitada por ADR-0105:
  región UE, coste máximo de 0 €, sin identificación, replay, autocapture,
  GeoIP ni flags remotos.
- La validación de excepciones JavaScript y crashes nativos simbolizados en
  builds distribuibles sigue siendo obligatoria antes de distribuir la beta;
  este ADR no la elimina ni autoriza distribuir sin ella.
- Compose sigue siendo el entorno de desarrollo y el runtime público `dev`.

## Criterios de decisión

1. no bloquear aprendizaje local por una cuenta externa aún no contratada;
2. no rebajar los controles de privacidad ni la evidencia de PostHog;
3. no introducir otro proveedor, build temporal ni coste para simular la
   distribución móvil;
4. mantener trazabilidad explícita del trabajo pendiente.

## Alternativas

### Alternativa A — Esperar las cuentas móviles antes de K3s

- **Ventajas:** conserva literalmente el orden de ADR-0102.
- **Inconvenientes:** deja inactiva la Fase 4 sin reducir riesgo técnico de
  PostHog ni generar aprendizaje adicional.
- **Coste de adopción:** nulo.
- **Coste de mantenimiento:** bajo, pero introduce tiempo de espera sin valor.
- **Riesgos:** confundir una dependencia comercial de distribución con un
  requisito del laboratorio local.

### Alternativa B — Iniciar VM Linux + K3s y diferir solo la validación distribuible

- **Ventajas:** desbloquea la Fase 4 con el alcance ya aceptado, sin relajar la
  validación ni la política de privacidad de PostHog.
- **Inconvenientes:** dos líneas de aprendizaje quedan temporalmente en fases
  distintas y habrá que retomar la validación móvil al disponer de las cuentas.
- **Coste de adopción:** bajo; solo actualización de trazabilidad.
- **Coste de mantenimiento:** bajo; una comprobación pendiente y explícita.
- **Riesgos:** olvidar la validación si no se mantiene como gate de distribución.

### No cambiar

K3s permanece bloqueado hasta el mes siguiente sin aportar evidencia adicional.

## Comparación

La alternativa B satisface la restricción real —las cuentas no están
disponibles— y conserva intacto el criterio de salida de PostHog para una beta
distribuida. No añade infraestructura, proveedor ni excepción de privacidad.

## Recomendación

**Opinión/recomendación:** alternativa B. Es la mínima separación entre una
dependencia de distribución móvil y un laboratorio de orquestación local.

## Decisión del usuario

**Aceptada el 2026-08-23:** iniciar ahora la Fase 4 con una VM Linux local y
K3s, conforme a ADR-0101. Se difiere la validación de PostHog en builds
distribuibles de iOS y Android hasta disponer de las cuentas necesarias,
previsiblemente el mes siguiente. Esa validación continúa siendo obligatoria
antes de distribuir la beta pública móvil.

## Consecuencias

### Positivas

- La Fase 4 puede empezar sin gasto ni cuenta externa adicionales.
- PostHog conserva sus límites de datos y su validación completa.

### Negativas y deuda aceptada

- La evidencia de crashes nativos simbolizados queda pendiente y debe retomarse.
- La Fase 4 no prueba todavía la entrega distribuida del cliente móvil.

## Validación

1. El roadmap presenta VM Linux + K3s como siguiente fase y no condiciona su
   inicio a PostHog distribuible.
2. La Fase 4 cumple su propio criterio de salida y cierra retrospectiva.
3. Al disponer de las cuentas, una build distribuible de iOS y otra de Android
   demuestran excepción JavaScript y crash nativo simbolizados; se revisan
   región UE, límite de gasto, filtrado, retención y privacidad antes de beta.

## Disparadores de revisión

- Las cuentas móviles siguen sin estar disponibles cuando la Fase 4 termine.
- Se pretende distribuir una beta móvil antes de completar la validación.
- PostHog introduce coste, cambia su configuración de privacidad o no permite
  simbolizar correctamente los fallos.

## Documentación afectada

- [ROADMAP.md](../project/ROADMAP.md)
- [DECISIONS.md](../governance/DECISIONS.md)
- [DECISIONS_TO_REVISIT.md](../governance/DECISIONS_TO_REVISIT.md)
- [LEARNING.md](../project/LEARNING.md)
- [README.md](../../README.md)
- [CHANGELOG.md](../../CHANGELOG.md)
