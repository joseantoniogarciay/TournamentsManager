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

- [ADR-0152: Corregir errores en recursos Android generados por Expo](0152-correct-expo-generated-android-lint-errors.md) — Aceptado; esperar versiones corregidas de Expo.

- [ADR-0153: Tamaño fijo de las celdas de clasificación](0153-keep-standings-table-at-a-fixed-text-size.md) — Aceptado; excepción localizada al escalado de texto.

- [ADR-0154: Ancho web responsive](0154-constrain-responsive-web-content-width.md) — Aceptado.

- [ADR-0155: Hosting web en Cloudflare Static Assets](0155-host-production-web-on-cloudflare-static-assets.md) — Aceptado; conmutación pendiente de validación en la cuenta.

- [ADR-0157: PostgreSQL, pgBackRest y Caddy](0157-upgrade-postgresql-pgbackrest-and-caddy.md) — Aceptado; actualización operativa y excepción puntual de madurez.

- [ADR-0156: Valoraciones nativas en hitos de torneos](0156-request-native-store-reviews-at-tournament-milestones.md) — Aceptado; ambos disparadores comparten 60 días entre intentos.

- [ADR-0158: Fecha, hora y lugar opcionales](0158-add-optional-tournament-and-match-scheduling.md) — Aceptado.

- [ADR-0159: Repetir torneos finalizados como nuevas ediciones](0159-repeat-completed-tournaments-as-new-editions.md) — Aceptado.
- [ADR-0160: Elegir equipos libres o crear otros mediante invitación](0160-choose-unlinked-teams-or-create-through-invitations.md) — Aceptado.

- [ADR-0161: Mis partidos y acceso al encuentro dentro del torneo](0161-add-personal-matches-tab-and-match-deep-links.md) — Aceptado; implementación autorizada.
- [ADR-0162: Consultar ediciones relacionadas desde el torneo](0162-browse-related-tournament-editions.md) — Aceptado; implementación autorizada.

- [ADR-0163: Contener recursos API y revalidar invitaciones](0163-bound-api-resource-work-and-revalidate-invitations.md) — Aceptado; corrección autorizada por el usuario.

- [ADR-0164: Gestionar sesiones opacas activas](0164-manage-active-opaque-sessions.md) — Aceptado.
