# Backend Engineering Handbook

> Estado: Fase 0 — documentación
>
> Última revisión: 2026-07-23

Este repositorio empieza por el handbook porque el objetivo no es solo entregar una
aplicación: es aprender a diseñar, construir, desplegar y operar un backend con
criterio profesional. El archivo
[PROJECT_MANIFESTO.docx](docs/source/PROJECT_MANIFESTO.docx) es la fuente de
verdad. [PROJECT_MANIFESTO.md](PROJECT_MANIFESTO.md) es su transcripción
versionable y navegable.

## Cómo usar este handbook

1. Lee [WHY.md](WHY.md) para entender el propósito y los límites.
2. Consulta [DECISIONS.md](DECISIONS.md) antes de proponer una tecnología o patrón.
3. Sigue el [playbook de decisiones](docs/playbooks/decision-process.md).
4. Registra en un ADR toda decisión importante.
5. Actualiza la documentación afectada en el mismo cambio.
6. Cierra cada fase con una retrospectiva técnica.

## Mapa del handbook

### Propósito y gobierno

- [PROJECT_MANIFESTO.md](PROJECT_MANIFESTO.md): principios innegociables y dirección.
- [WHY.md](WHY.md): problema de aprendizaje, resultados y no objetivos.
- [PRODUCT.md](PRODUCT.md): alcance funcional, actores y flujos aceptados.
- [TECHNICAL_BASELINE.md](TECHNICAL_BASELINE.md): gate técnico activo y estado.
- [SYSTEM_OPTIONS.md](SYSTEM_OPTIONS.md): alternativas y orden de decisiones.
- [DECISIONS.md](DECISIONS.md): índice y reglas de decisión.
- [ROADMAP.md](ROADMAP.md): fases, puertas de entrada y criterios de salida.
- [LEARNING.md](LEARNING.md): competencias y diario de aprendizaje.
- [DECISIONS_TO_REVISIT.md](DECISIONS_TO_REVISIT.md): decisiones y supuestos con fecha
  o disparadores de revisión.

### Ingeniería

- [ARCHITECTURE.md](ARCHITECTURE.md)
- [DEVELOPMENT.md](DEVELOPMENT.md)
- [DATABASE.md](DATABASE.md)
- [API.md](API.md)
- [SECURITY.md](SECURITY.md)
- [OBSERVABILITY.md](OBSERVABILITY.md)
- [DEPLOYMENT.md](DEPLOYMENT.md)
- [STYLEGUIDE.md](STYLEGUIDE.md)
- [TESTING.md](TESTING.md)

### Operación y colaboración

- [CONTRIBUTING.md](CONTRIBUTING.md)
- [TROUBLESHOOTING.md](TROUBLESHOOTING.md)
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
| Go, PostgreSQL, Docker y AWS | Dirección objetivo | Manifiesto; requieren decisiones de implementación |
| Producto web y mobile de torneos | Alcance aceptado | [PRODUCT.md](PRODUCT.md) |
| React y React Native | Dirección preferida | [SYSTEM_OPTIONS.md](SYSTEM_OPTIONS.md) |
| Redis o Valkey | Pendiente de evaluación | [DECISIONS_TO_REVISIT.md](DECISIONS_TO_REVISIT.md) |
| Stack de observabilidad | Pendiente de evaluación | [OBSERVABILITY.md](OBSERVABILITY.md) |
| Frontend | Pendiente de decisión | [DECISIONS_TO_REVISIT.md](DECISIONS_TO_REVISIT.md) |

El gate activo es [TECHNICAL_BASELINE.md](TECHNICAL_BASELINE.md). Las decisiones
funcionales permanecen pausadas hasta cerrarlo.

## Regla de precedencia

Ante una contradicción se aplica este orden:

1. `docs/source/PROJECT_MANIFESTO.docx`.
2. ADR aceptado más reciente que no contradiga el manifiesto.
3. Documentación especializada del handbook.
4. Implementación.

Una divergencia detectada no se resuelve silenciosamente: se registra, se decide y
se corrige.
