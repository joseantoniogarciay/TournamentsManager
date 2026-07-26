# Gobierno de decisiones

## Autoridad

El usuario toma la decisión final. El asistente y cualquier colaborador preparan
el análisis, diferencian hechos de opiniones y formulan una recomendación. Una
recomendación no se convierte en decisión por aparecer en un documento.

## Cuándo hace falta un ADR

Se crea un ADR si la elección:

- cambia límites arquitectónicos o la dirección de dependencias;
- introduce una tecnología, servicio o dependencia operativa relevante;
- afecta datos, seguridad, disponibilidad, coste o portabilidad;
- es difícil o costosa de revertir;
- establece una convención transversal;
- reemplaza una decisión aceptada.

Las decisiones locales y reversibles pueden documentarse junto al cambio. Si se
repiten o empiezan a condicionar otras áreas, se elevan a ADR.

## Estados

- **Propuesto:** análisis abierto; no autoriza implementación.
- **Aceptado:** decisión final del usuario.
- **Rechazado:** analizado y no elegido.
- **Superado:** reemplazado por otro ADR.
- **En revisión:** un disparador exige reconsideración.

Solo un ADR aceptado puede presentarse como “decisión tomada”.

## Proceso obligatorio

1. Definir el problema y el resultado que se necesita.
2. Aclarar restricciones, supuestos y criterios.
3. Comparar al menos dos alternativas razonables, incluido “no hacer nada” cuando
   aplique.
4. Explicar ventajas, inconvenientes, riesgos y coste de mantenimiento.
5. Separar estándar de industria, evidencia y opinión.
6. Presentar una recomendación.
7. Obtener la decisión explícita del usuario.
8. Registrar el ADR y actualizar los documentos afectados.
9. Definir cómo se validará y cuándo debe revisarse.

Playbook completo: [decision-process.md](../playbooks/decision-process.md).

## Índice de ADR

| ADR                                                                     | Título                                                    | Estado   | Fecha      |
| ----------------------------------------------------------------------- | --------------------------------------------------------- | -------- | ---------- |
| [0000](../adr/0000-record-architecture-decisions.md)                    | Registrar decisiones arquitectónicas                      | Aceptado | 2026-07-23 |
| [0001](../adr/0001-pragmatic-clean-architecture.md)                     | Clean Architecture pragmática con principios hexagonales  | Aceptado | 2026-07-23 |
| [0002](../adr/0002-handbook-before-code.md)                             | Construir el handbook antes que el código                 | Aceptado | 2026-07-23 |
| [0003](../adr/0003-use-git-for-version-control.md)                      | Usar Git para control de versiones                        | Aceptado | 2026-07-23 |
| [0004](../adr/0004-technical-baseline-before-product-design.md)         | Confirmar la base técnica antes del diseño de producto    | Aceptado | 2026-07-23 |
| [0005](../adr/0005-use-a-product-monorepo.md)                           | Usar un monorepo de producto                              | Aceptado | 2026-07-23 |
| [0006](../adr/0006-public-github-repository-security-boundary.md)       | Publicar el monorepo en GitHub sin publicar secretos      | Aceptado | 2026-07-23 |
| [0007](../adr/0007-use-a-modular-monolith-backend.md)                   | Usar un monolito modular para el backend                  | Aceptado | 2026-07-24 |
| [0008](../adr/0008-use-a-universal-react-native-client.md)              | Usar un cliente universal con React Native                | Aceptado | 2026-07-24 |
| [0009](../adr/0009-use-rest-and-openapi-contract-first.md)              | Usar REST con OpenAPI contract-first                      | Aceptado | 2026-07-24 |
| [0010](../adr/0010-own-identity-with-federated-login.md)                | Gestionar identidad propia con login federado             | Aceptado | 2026-07-24 |
| [0011](../adr/0011-use-postgresql-pgx-sqlc-and-goose.md)                | Usar PostgreSQL con pgx, sqlc y goose                     | Aceptado | 2026-07-24 |
| [0012](../adr/0012-pin-go-toolchain-and-isolate-tools.md)               | Fijar el toolchain Go y aislar las herramientas           | Aceptado | 2026-07-24 |
| [0013](../adr/0013-use-develop-as-integration-branch.md)                | Usar `develop` como rama de integración                   | Aceptado | 2026-07-24 |
| [0014](../adr/0014-use-node-pnpm-and-strict-typescript.md)              | Usar Node LTS, pnpm y TypeScript estricto                 | Aceptado | 2026-07-24 |
| [0015](../adr/0015-use-expo-router-and-continuous-native-generation.md) | Usar Expo, Expo Router y CNG                              | Aceptado | 2026-07-24 |
| [0016](../adr/0016-use-client-side-web-rendering-initially.md)          | Usar rendering web client-side inicialmente               | Aceptado | 2026-07-24 |
| [0017](../adr/0017-use-env-contracts-github-environments-and-oidc.md)   | Usar contratos de entorno, GitHub Environments y OIDC     | Aceptado | 2026-07-24 |
| [0018](../adr/0018-use-compose-for-local-service-dependencies.md)       | Usar Docker Compose para dependencias locales de servicio | Aceptado | 2026-07-25 |
| [0019](../adr/0019-use-risk-based-layered-testing.md)                    | Usar pruebas por riesgo y capas                           | Aceptado | 2026-07-25 |
| [0020](../adr/0020-use-minimal-correlated-observability.md)              | Usar observabilidad mínima correlacionada                 | Aceptado | 2026-07-25 |
| [0021](../adr/0021-use-advisory-ci-with-local-quality-gate.md)            | Usar CI informativa con puerta de calidad local           | Aceptado | 2026-07-25 |
| [0022](../adr/0022-package-backend-as-oci-image.md)                        | Empaquetar la API como imagen OCI                         | Aceptado | 2026-07-25 |
| [0023](../adr/0023-use-ecs-fargate-as-future-cloud-runtime.md)             | Usar ECS con Fargate como runtime cloud futuro             | Aceptado | 2026-07-25 |
| [0024](../adr/0024-use-ecr-and-digest-based-image-promotion.md)            | Usar ECR y promoción por digest con releases selectivas    | Aceptado | 2026-07-25 |
| [0025](../adr/0025-use-terraform-for-infrastructure-as-code.md)            | Usar Terraform para la infraestructura como código         | Aceptado | 2026-07-25 |
| [0026](../adr/0026-use-aws-organizations-and-temporary-identities.md)      | Usar AWS Organizations e identidades temporales             | Aceptado | 2026-07-25 |
| [0027](../adr/0027-keep-local-state-until-first-cloud-apply.md)            | Mantener estado local hasta el primer apply cloud           | Aceptado | 2026-07-25 |
| [0028](../adr/0028-use-hcp-terraform-free-for-remote-state.md)              | Usar HCP Terraform Free para el estado remoto inicial       | Aceptado | 2026-07-25 |
| [0029](../adr/0029-use-public-alb-restricted-fargate-and-no-nat-initially.md) | Usar ALB público, Fargate restringido y sin NAT inicialmente | Aceptado | 2026-07-25 |
| [0030](../adr/0030-use-spain-region-and-two-az-cost-gated-network.md)         | Usar la región España y una red en dos AZ con gasto autorizado | Aceptado | 2026-07-25 |
| [0031](../adr/0031-preserve-pre-auth-tournament-drafts-until-verified.md)     | Conservar borradores previos al acceso hasta verificar la cuenta | Aceptado | 2026-07-26 |
| [0032](../adr/0032-define-minimum-football-league-data-and-lifecycle.md)      | Definir los datos mínimos y ciclo de vida de una liga de fútbol | Superado parcialmente | 2026-07-26 |
| [0033](../adr/0033-use-unlisted-read-only-links-for-published-leagues.md)     | Usar enlaces no listados de solo lectura para ligas publicadas | Aceptado | 2026-07-26 |
| [0034](../adr/0034-use-teams-as-competitors-and-direct-delegated-administration.md) | Usar equipos como participantes y administración delegada directa | Aceptado | 2026-07-26 |
| [0035](../adr/0035-apply-results-entered-by-delegated-administrators-immediately.md) | Aplicar inmediatamente los resultados de administradores delegados | Aceptado | 2026-07-26 |
| [0036](../adr/0036-allow-delegated-administrators-to-correct-results-with-history.md) | Permitir a administradores corregir resultados con historial | Aceptado | 2026-07-26 |
| [0037](../adr/0037-use-simple-home-and-away-scores-for-initial-football-results.md) | Usar marcadores simples local-visitante en resultados iniciales | Aceptado | 2026-07-26 |
| [0038](../adr/0038-require-creator-to-start-a-league-before-results.md) | Requerir que el creador inicie la liga antes de registrar resultados | Aceptado | 2026-07-26 |
| [0039](../adr/0039-require-complete-results-and-explicit-closure.md) | Requerir resultados completos y cierre explícito de la liga | Aceptado | 2026-07-26 |
| [0040](../adr/0040-make-published-leagues-editable-until-start.md) | Mantener la liga publicada editable hasta iniciarla | Aceptado | 2026-07-26 |
| [0041](../adr/0041-continue-league-after-team-withdrawal-with-3-0-results.md) | Continuar la liga tras una baja con resultados 3-0 | Aceptado | 2026-07-26 |
| [0042](../adr/0042-cancel-leagues-without-mandatory-reason-or-notifications.md) | Cancelar ligas sin motivo obligatorio ni avisos automáticos | Aceptado | 2026-07-26 |

## Trazabilidad de un cambio

Toda propuesta importante debe enlazar:

`problema → análisis → decisión → cambio → prueba → documentación → aprendizaje`

Si falta un eslabón, el cambio no está terminado.
