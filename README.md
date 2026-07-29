# TournamentsManager — Engineering Handbook

> Estado: Fase 2 — primer vertical slice de backend en curso
>
> Última revisión: 2026-07-28

Este repositorio empieza por el handbook porque el objetivo no es solo entregar una
aplicación: es aprender a diseñar, construir, desplegar y operar un producto con
criterio profesional. El archivo
[PROJECT_MANIFESTO.docx](docs/source/PROJECT_MANIFESTO.docx) es la fuente de
verdad. [PROJECT_MANIFESTO.md](PROJECT_MANIFESTO.md) es su transcripción
versionable y navegable.

## Cómo usar este handbook

1. Lee [WHY.md](docs/project/WHY.md) para entender el propósito y los límites.
2. Consulta [DECISIONS.md](docs/governance/DECISIONS.md) antes de proponer una tecnología o patrón.
3. Sigue el [playbook de decisiones](docs/playbooks/decision-process.md).
4. Registra en un ADR toda decisión importante.
5. Actualiza la documentación afectada en el mismo cambio.
6. Cierra cada fase con una retrospectiva técnica.

## Estructura del monorepo

La raíz conserva solo los puntos de entrada y autoridad. El handbook se organiza
desde [docs/README.md](docs/README.md). Las unidades de código se crearán cuando
sus decisiones de implementación estén aceptadas:

```text
apps/
  backend/       # Backend y adaptador REST en Go
  client/        # Cliente universal React Native + TypeScript
contracts/
  openapi/       # Fuente de verdad del contrato HTTP
infra/           # Entorno, despliegue e infraestructura
docs/            # Handbook, ADR y material operativo
```

Esta estructura no autoriza por sí misma nuevos frameworks, herramientas ni
decisiones de implementación.

La guía de uso del cliente está en [apps/client/README.md](apps/client/README.md).

## Mapa del handbook

### Propósito y gobierno

- [Índice de documentación](docs/README.md)
- [PROJECT_MANIFESTO.md](PROJECT_MANIFESTO.md): principios innegociables y dirección.
- [WHY.md](docs/project/WHY.md): problema de aprendizaje, resultados y no objetivos.
- [PRODUCT.md](docs/project/PRODUCT.md): alcance funcional, actores y flujos aceptados.
- [TECHNICAL_BASELINE.md](docs/governance/TECHNICAL_BASELINE.md): base técnica cerrada y trazabilidad de decisiones.
- [SYSTEM_OPTIONS.md](docs/governance/SYSTEM_OPTIONS.md): alternativas y orden de decisiones.
- [DECISIONS.md](docs/governance/DECISIONS.md): índice y reglas de decisión.
- [ROADMAP.md](docs/project/ROADMAP.md): fases, puertas de entrada y criterios de salida.
- [LEARNING.md](docs/project/LEARNING.md): competencias y diario de aprendizaje.
- [DECISIONS_TO_REVISIT.md](docs/governance/DECISIONS_TO_REVISIT.md): decisiones y supuestos con fecha
  o disparadores de revisión.

### Ingeniería

- [ARCHITECTURE.md](docs/engineering/ARCHITECTURE.md)
- [DEVELOPMENT.md](docs/engineering/DEVELOPMENT.md)
- [DATABASE.md](docs/engineering/DATABASE.md)
- [API.md](docs/engineering/API.md)
- [IDENTITY.md](docs/engineering/IDENTITY.md)
- [SECURITY.md](docs/engineering/SECURITY.md)
- [OBSERVABILITY.md](docs/operations/OBSERVABILITY.md)
- [DEPLOYMENT.md](docs/operations/DEPLOYMENT.md)
- [STYLEGUIDE.md](docs/engineering/STYLEGUIDE.md)
- [TESTING.md](docs/engineering/TESTING.md)

### Operación y colaboración

- [CONTRIBUTING.md](CONTRIBUTING.md)
- [TROUBLESHOOTING.md](docs/operations/TROUBLESHOOTING.md)
- [CHANGELOG.md](CHANGELOG.md)
- [ADR](docs/adr/README.md)
- [Knowledge base](docs/knowledge/README.md)
- [Playbooks](docs/playbooks/README.md)
- [Runbooks](docs/runbooks/README.md)
- [Diagramas](docs/diagrams/README.md)

## Estado de las decisiones

| Área | Estado | Fuente |
|---|---|---|
| Handbook antes que código | Aceptada | [ADR-0002](docs/adr/0002-handbook-before-code.md) |
| Proceso ADR | Aceptada | [ADR-0000](docs/adr/0000-record-architecture-decisions.md) |
| Arquitectura pragmática clean/hexagonal | Aceptada | [ADR-0001](docs/adr/0001-pragmatic-clean-architecture.md) |
| Git para control de versiones | Aceptada | [ADR-0003](docs/adr/0003-use-git-for-version-control.md) |
| Base técnica antes del detalle de producto | Aceptada | [ADR-0004](docs/adr/0004-technical-baseline-before-product-design.md) |
| Monorepo de producto | Aceptada | [ADR-0005](docs/adr/0005-use-a-product-monorepo.md) |
| GitHub público con secretos protegidos | Aceptada | [ADR-0006](docs/adr/0006-public-github-repository-security-boundary.md) |
| Backend como monolito modular | Aceptada | [ADR-0007](docs/adr/0007-use-a-modular-monolith-backend.md) |
| Cliente universal web, iOS y Android | Aceptada | [ADR-0008](docs/adr/0008-use-a-universal-react-native-client.md) |
| REST con OpenAPI contract-first | Aceptada | [ADR-0009](docs/adr/0009-use-rest-and-openapi-contract-first.md) |
| Identidad propia federada con Apple y Google | Aceptada | [ADR-0010](docs/adr/0010-own-identity-with-federated-login.md) |
| PostgreSQL con pgx, sqlc y goose | Aceptada | [ADR-0011](docs/adr/0011-use-postgresql-pgx-sqlc-and-goose.md) |
| Go 1.26.5 y herramientas aisladas | Aceptada | [ADR-0012](docs/adr/0012-pin-go-toolchain-and-isolate-tools.md) |
| `develop` como rama de integración | Aceptada | [ADR-0013](docs/adr/0013-use-develop-as-integration-branch.md) |
| Node LTS, pnpm y TypeScript estricto | Aceptada | [ADR-0014](docs/adr/0014-use-node-pnpm-and-strict-typescript.md) |
| Expo, Expo Router y CNG | Aceptada | [ADR-0015](docs/adr/0015-use-expo-router-and-continuous-native-generation.md) |
| Rendering web client-side inicial | Aceptada | [ADR-0016](docs/adr/0016-use-client-side-web-rendering-initially.md) |
| Configuración y secretos | Aceptada | [ADR-0017](docs/adr/0017-use-env-contracts-github-environments-and-oidc.md) |
| Estrategia de pruebas por riesgo y capas | Aceptada | [ADR-0019](docs/adr/0019-use-risk-based-layered-testing.md) |
| Go, Docker y AWS | Dirección objetivo | Manifiesto; requieren decisiones de implementación |
| Producto web y mobile de torneos | Alcance aceptado | [PRODUCT.md](docs/project/PRODUCT.md) |
| React Native universal | Expo, Expo Router y CNG aceptados | [ADR-0008](docs/adr/0008-use-a-universal-react-native-client.md), [ADR-0015](docs/adr/0015-use-expo-router-and-continuous-native-generation.md) |
| TypeScript y cliente API generado | Aceptada; toolchain fijado | [ADR-0009](docs/adr/0009-use-rest-and-openapi-contract-first.md), [ADR-0014](docs/adr/0014-use-node-pnpm-and-strict-typescript.md) |
| Redis o Valkey | Pendiente de evaluación | [DECISIONS_TO_REVISIT.md](docs/governance/DECISIONS_TO_REVISIT.md) |
| Observabilidad mínima | OpenTelemetry, Prometheus, Grafana, Loki y Tempo aceptados; Collector aplazado | [ADR-0020](docs/adr/0020-use-minimal-correlated-observability.md) |
| Observabilidad de producto del cliente | PostHog Cloud diferido para beta distribuida; región UE, límite de gasto 0 € y sin autoridad sobre negocio | [ADR-0060](docs/adr/0060-use-posthog-for-deferred-client-product-observability.md) |
| Empaquetado de la API | Imagen OCI/Docker aceptada | [ADR-0022](docs/adr/0022-package-backend-as-oci-image.md) |
| Runtime cloud de la API | ECS con Fargate en Fase 5; todo el trabajo actual continúa local y sin coste AWS | [ADR-0023](docs/adr/0023-use-ecs-fargate-as-future-cloud-runtime.md) |
| Registry y promoción de la API | ECR privado, digest inmutable y releases selectivas futuras; sin recursos AWS aún | [ADR-0024](docs/adr/0024-use-ecr-and-digest-based-image-promotion.md) |
| Infraestructura como código | Terraform aceptado; cuenta, estado remoto y red AWS pendientes, sin recursos creados | [ADR-0025](docs/adr/0025-use-terraform-for-infrastructure-as-code.md) |
| Fundación AWS | Organizations con cuentas `management`, `nonprod` y `prod`; acceso temporal con Identity Center y MFA | [ADR-0026](docs/adr/0026-use-aws-organizations-and-temporary-identities.md) |
| Estado Terraform | Local solo antes de AWS real; backend remoto obligatorio desde el primer apply cloud | [ADR-0027](docs/adr/0027-keep-local-state-until-first-cloud-apply.md) |
| Backend remoto de Terraform | HCP Terraform Free con ejecución local inicial; sin recursos AWS ni auto-apply | [ADR-0028](docs/adr/0028-use-hcp-terraform-free-for-remote-state.md) |
| Entrada pública y egress inicial | ALB público; API Fargate solo accesible desde el ALB; base de datos privada y sin NAT | [ADR-0029](docs/adr/0029-use-public-alb-restricted-fargate-and-no-nat-initially.md) |
| Región y red AWS | España (`eu-south-2`); VPC separada por cuenta en dos AZ; gasto solo tras autorización explícita | [ADR-0030](docs/adr/0030-use-spain-region-and-two-az-cost-gated-network.md) |
| Framework y rendering del cliente | Aceptado | [ADR-0015](docs/adr/0015-use-expo-router-and-continuous-native-generation.md), [ADR-0016](docs/adr/0016-use-client-side-web-rendering-initially.md) |

La [base técnica](docs/governance/TECHNICAL_BASELINE.md), el gate 0B y la Fase 1
están cerrados. El proyecto se encuentra en la [Fase 2: Backend](docs/project/ROADMAP.md):
el primer vertical slice implementa identidad local, publicación y consulta de
ligas de forma incremental. El corte activo es el registro local de una cuenta
pendiente; la verificación, sesión y publicación se incorporarán en los cortes
siguientes. Las decisiones nuevas que alcancen el umbral definido siguen
requiriendo ADR aceptado.

## Cliente universal

El cliente Expo está disponible para web, iOS y Android. Su punto de entrada es
la [guía de `apps/client`](apps/client/README.md): explica cómo iniciarlo, su
navegación actual y las reglas de localización y diseño que aplican a cada
pantalla. La referencia operativa completa permanece en
[DEVELOPMENT.md](docs/engineering/DEVELOPMENT.md) para evitar duplicar comandos
y decisiones técnicas.

## Regla de precedencia

Ante una contradicción se aplica este orden:

1. `docs/source/PROJECT_MANIFESTO.docx`.
2. ADR aceptado más reciente que no contradiga el manifiesto.
3. Documentación especializada del handbook.
4. Implementación.

Una divergencia detectada no se resuelve silenciosamente: se registra, se decide y
se corrige.
