# Changelog

Los cambios relevantes del handbook, producto y operación se registran aquí. El
formato seguirá categorías `Added`, `Changed`, `Deprecated`, `Removed`, `Fixed` y
`Security` cuando existan releases.

## [Unreleased]

### Changed

- Ruta del módulo Go alineada con el propietario canónico de GitHub antes de la
  primera publicación.
- Handbook reorganizado por proyecto, gobierno, ingeniería y operaciones; la
  raíz conserva únicamente documentos de entrada y autoridad.
- El primer login social sobre una cuenta local queda pendiente hasta consumir
  un enlace de vinculación de un solo uso.
- La confirmación usa HTTPS deep linking con fallback web y sustituye la sesión
  local sin cambiar el propietario de una sesión existente.
- El enlace de vinculación abre una ruta dedicada mediante `GET` seguro; una
  confirmación explícita mediante `POST` consume el intento y termina en la home
  sin conservar el token en la navegación.

### Added

- ADR-0013: `develop` como rama de integración diaria, con promociones por
  bloques a `main` y ramas temporales solo cuando aporten aislamiento real.
- ADR-0012: Go 1.26.5, módulos separados para aplicación y herramientas,
  `goimports`, golangci-lint v2 y `govulncheck`.
- Makefile con checks rápidos, verificación completa y mantenimiento simétrico de
  `go.mod` y `go.tool.mod`.
- Configuración compartida de VS Code para formatear al guardar con el
  `goimports` pineado.
- ADR-0011: PostgreSQL con `pgx` nativo, código de acceso generado por `sqlc` y
  migraciones SQL versionadas mediante `goose`.
- Reglas de separación entre dominio, adaptador, filas generadas y esquema, junto
  con evidencia mínima de generación y migración.
- ADR-0010: identidad propia con credenciales locales y login federado mediante
  Apple y Google.
- Guía de identidad, subject verificado, cambio de email y vinculación segura.
- ADR-0009: API REST con OpenAPI contract-first y cliente TypeScript generado.
- Distinción documental entre backend Go, adaptador HTTP y cliente TypeScript.
- ADR-0008: cliente universal React Native para web, iOS y Android.
- Paridad funcional y diseño adaptativo como límites de la experiencia
  multiplataforma.
- ADR-0007: backend como monolito modular en Go.
- Comparación activa de estrategias web/mobile.
- ADR-0005: monorepo de producto.
- ADR-0006: GitHub público con límites para secretos y despliegues.
- Comparación activa de monolito modular, microservicios y serverless.
- Gate ordenado para confirmar la base técnica antes del diseño funcional.
- ADR-0004 sobre la secuencia técnica primero.
- Comparación activa de monorepo, multirepo y topología híbrida.
- Repositorio Git local con rama `main`.
- Alcance inicial de producto: invitados, cuentas, torneos, web y mobile.
- Mapa comparado de decisiones del sistema.
- Contexto conceptual de actores y componentes.
- Próxima puerta de decisión: primer vertical slice e identidad.

## [0.1.0] — 2026-07-23

### Added

- Manifiesto preservado y transcrito.
- Backend Engineering Handbook inicial.
- Gobierno de decisiones y ADR.
- Baseline de arquitectura pragmática clean/hexagonal.
- Roadmap por fases y criterios de salida.
- Guías de arquitectura, desarrollo, datos, API, seguridad, observabilidad,
  despliegue, estilo y pruebas.
- Plantillas de knowledge base, playbooks, runbooks, diagramas y retrospectivas.
