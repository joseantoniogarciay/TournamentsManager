# Backend Engineering Handbook

> Estado: Fase 0 — documentación
>
> Última revisión: 2026-07-24

Este repositorio empieza por el handbook porque el objetivo no es solo entregar una
aplicación: es aprender a diseñar, construir, desplegar y operar un backend con
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

Esta silueta no autoriza todavía carpetas internas, frameworks ni herramientas.

## Mapa del handbook

### Propósito y gobierno

- [Índice de documentación](docs/README.md)
- [PROJECT_MANIFESTO.md](PROJECT_MANIFESTO.md): principios innegociables y dirección.
- [WHY.md](docs/project/WHY.md): problema de aprendizaje, resultados y no objetivos.
- [PRODUCT.md](docs/project/PRODUCT.md): alcance funcional, actores y flujos aceptados.
- [TECHNICAL_BASELINE.md](docs/governance/TECHNICAL_BASELINE.md): gate técnico activo y estado.
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
| Go, Docker y AWS | Dirección objetivo | Manifiesto; requieren decisiones de implementación |
| Producto web y mobile de torneos | Alcance aceptado | [PRODUCT.md](docs/project/PRODUCT.md) |
| React Native universal | Estrategia aceptada; framework pendiente | [ADR-0008](docs/adr/0008-use-a-universal-react-native-client.md) |
| TypeScript y cliente API generado | Aceptada; toolchain pendiente | [ADR-0009](docs/adr/0009-use-rest-and-openapi-contract-first.md) |
| Redis o Valkey | Pendiente de evaluación | [DECISIONS_TO_REVISIT.md](docs/governance/DECISIONS_TO_REVISIT.md) |
| Stack de observabilidad | Pendiente de evaluación | [OBSERVABILITY.md](docs/operations/OBSERVABILITY.md) |
| Framework y rendering del cliente | Pendiente de decisión | [TECHNICAL_BASELINE.md](docs/governance/TECHNICAL_BASELINE.md) |

El gate activo es [TECHNICAL_BASELINE.md](docs/governance/TECHNICAL_BASELINE.md). Las decisiones
funcionales permanecen pausadas hasta cerrarlo.

## Regla de precedencia

Ante una contradicción se aplica este orden:

1. `docs/source/PROJECT_MANIFESTO.docx`.
2. ADR aceptado más reciente que no contradiga el manifiesto.
3. Documentación especializada del handbook.
4. Implementación.

Una divergencia detectada no se resuelve silenciosamente: se registra, se decide y
se corrige.
