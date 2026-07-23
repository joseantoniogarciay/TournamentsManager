# Project Manifesto

> Transcripción estructurada de `docs/source/PROJECT_MANIFESTO.docx`.
>
> Fuente de verdad: el DOCX. Si existe una divergencia, prevalece el DOCX.

## Visión

Este proyecto no persigue únicamente construir una aplicación. Su objetivo es
servir como plataforma de aprendizaje para dominar el desarrollo backend moderno,
la arquitectura de software y la infraestructura cloud, entendiendo el motivo de
cada decisión.

## Objetivos

- Aprender el “por qué” antes que el “cómo”.
- Construir una plataforma cloud agnostic.
- Evitar sobreingeniería.
- Documentar todas las decisiones importantes.
- Conseguir un proyecto con calidad profesional.

## Principios innegociables

1. Simplicidad antes que complejidad.
2. No introducir abstracciones sin necesidad.
3. Comparar alternativas antes de decidir.
4. El usuario toma las decisiones finales.
5. Toda decisión importante tendrá un ADR.
6. La documentación evoluciona junto al código.
7. La lógica de negocio no depende de infraestructura.
8. El entorno local debe parecerse a producción.

## Proceso de decisión

Para cualquier tecnología o patrón:

- Problema.
- Alternativas.
- Ventajas.
- Inconvenientes.
- Coste de mantenimiento.
- Recomendación.
- Decisión del usuario.
- Registro en ADR.
- Actualización de documentación.

## Arquitectura acordada

Clean Architecture pragmática con principios hexagonales.

Interfaces solo cuando aporten valor: desacoplamiento, tests o varias
implementaciones reales.

No crear capas innecesarias.

## Stack objetivo

- Backend: Go.
- Base de datos: PostgreSQL.
- Cache: Redis, evaluando Valkey cuando corresponda.
- Storage: MinIO local y S3 en cloud.
- Contenedores: Docker y Docker Compose.
- Orquestación: Kubernetes, con k3d en local.
- Observabilidad: OpenTelemetry, Prometheus, Grafana, Loki y Tempo, a evaluar.
- Cloud inicial: AWS.
- IaC: Terraform.
- Frontend: pendiente de decisión tras comparar alternativas.

## Roadmap

- Fase 0: documentación.
- Fase 1: entorno local.
- Fase 2: backend.
- Fase 3: observabilidad.
- Fase 4: Kubernetes.
- Fase 5: cloud.

Cada fase termina con una retrospectiva técnica.

## Documentación

Documentos raíz:

- `README.md`
- `ARCHITECTURE.md`
- `DEVELOPMENT.md`
- `DATABASE.md`
- `API.md`
- `SECURITY.md`
- `OBSERVABILITY.md`
- `DEPLOYMENT.md`
- `STYLEGUIDE.md`
- `TESTING.md`
- `CONTRIBUTING.md`
- `ROADMAP.md`
- `CHANGELOG.md`
- `TROUBLESHOOTING.md`
- `DECISIONS.md`

Carpetas:

- `docs/adr`
- `docs/knowledge`
- `docs/playbooks`
- `docs/runbooks`
- `docs/diagrams`

Documentos especiales:

- `WHY.md`
- `LEARNING.md`
- `DECISIONS_TO_REVISIT.md`

## Reglas para el asistente

- Actuar como mentor técnico.
- Explicar fundamentos antes de automatizar.
- Mostrar siempre alternativas.
- Priorizar estándares de la industria.
- Avisar cuando aparezca sobreingeniería.
- Diferenciar claramente opinión, estándar y decisión tomada.

## Meta final

Ser capaz de diseñar, desarrollar, desplegar y operar una plataforma profesional
basada en Go, PostgreSQL, Docker, Kubernetes y observabilidad, comprendiendo cada
decisión arquitectónica y pudiendo adaptarla a distintos proveedores cloud.
