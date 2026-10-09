# Architecture Decision Records

Los ADR conservan el contexto y las consecuencias de decisiones importantes. No
son actas extensas ni documentación de una tecnología.

## Convención

- Nombre: `NNNN-titulo-en-kebab-case.md`.
- Numeración: secuencial, cuatro dígitos; no reutilizar números.
- Idioma: español.
- Un ADR aceptado es inmutable en su decisión; las aclaraciones menores se añaden
  con fecha. Un cambio de decisión crea un ADR sucesor.
- El índice canónico está en [DECISIONS.md](../governance/DECISIONS.md).

## Ciclo de vida

`Propuesto → Aceptado | Rechazado → En revisión → Superado`

Solo el usuario cambia una propuesta a **Aceptado**. El autor registra la evidencia
de esa decisión en el campo correspondiente.

## Contenido obligatorio

- problema;
- contexto y restricciones;
- criterios;
- alternativas;
- comparación;
- recomendación;
- decisión del usuario;
- consecuencias;
- validación;
- disparadores de revisión;
- documentos afectados.

Usa [template.md](template.md) y el
[playbook de decisiones](../playbooks/decision-process.md).

- [ADR-0138: Esperar siete días salvo vulnerabilidades críticas](0138-wait-seven-days-except-critical-vulnerabilities.md) — Aceptado.

- [ADR-0143: Idioma automático desde sistema y navegador](0143-detect-client-language-from-system-and-browser.md) — Aceptado.

- [ADR-0144: Ciclo de vida UIScene mediante CNG](0144-adopt-ios-scene-lifecycle-through-cng.md) — Aceptado.

- [ADR-0145: Drenaje de la API antes de terminar el contenedor](0145-drain-api-before-container-termination.md) — Aceptado.

- [ADR-0146: Desarrollo y observabilidad bajo petición](0146-run-development-and-observability-on-demand.md) — Aceptado.

- [ADR-0147: Acceso Apple mediante navegador en todos los clientes](0147-add-apple-browser-login-on-all-clients.md) — Aceptado.

- [ADR-0148: Preparar fiabilidad de producción con proyectos PostHog separados](0148-prepare-production-error-tracking-with-separated-posthog-projects.md) — Superado por ADR-0149.
- [ADR-0149: Reservar el único proyecto PostHog para producción](0149-reserve-the-single-posthog-project-for-production.md) — Aceptado; supera ADR-0148.

- [ADR-0150: Host Android local para el banner global](0150-use-a-local-android-host-for-global-feedback.md) — Aceptado.

- [ADR-0151: Títulos iOS adaptables al espacio real](0151-reflow-ios-entity-titles-for-accessibility.md) — Aceptado.

- [ADR-0152: Corregir errores en recursos Android generados por Expo](0152-correct-expo-generated-android-lint-errors.md) — Propuesto.
