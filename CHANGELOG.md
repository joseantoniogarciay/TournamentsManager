# Changelog

Los cambios relevantes del handbook, producto y operación se registran aquí. El
formato seguirá categorías `Added`, `Changed`, `Deprecated`, `Removed`, `Fixed` y
`Security` cuando existan releases.

## [Unreleased]

### Added

- Registro completo en el cliente: alta local, correo de verificación HTML y
  texto alternativo, confirmación automática al abrir el deep link y sesión en cookie web o
  Keychain/Keystore móvil. Se añaden configuración CNG y plantillas documentadas
  de `apple-app-site-association` y `assetlinks.json` para asociar el dominio
  HTTPS cuando existan sus identificadores y certificados reales.
- `make local-api-up` valida los contratos locales, espera a PostgreSQL y
  Mailpit, y ejecuta la API Go en el host sin migraciones implícitas.
- ADR-0060: PostHog Cloud diferido para observabilidad de producto del cliente,
  con región UE, límite de gasto 0 €, replay restringido y correlación opaca con
  OpenTelemetry del backend; no se instala ni envía telemetría hasta una beta
  distribuida y la revisión de privacidad.
- README operativo de `apps/client`, con arranque web/iOS/Android, estructura,
  localización, reglas de cards y verificación; el README raíz enlaza al cliente
  y `CONTRIBUTING.md` concreta el alcance de `make verify`.
- `make verify` exporta el cliente Expo para web a un directorio temporal como
  comprobación de router y bundling.
- Botonera raíz nativa con Inicio, Torneos y Cuenta; en iOS 26 usa Liquid Glass
  del sistema y Cuenta dispone de su propio flujo de navegación.
- Catálogos planos JSON por locale (`es`, `en`, `it`, `fr`) con claves
  semánticas estables y fallback inglés para el cliente universal.
- Primera implementación de la home de invitado en `/`, con acción principal de
  crear torneo, acceso de cuenta y orientación sin datos personalizados
  simulados.
- ADR-0057: home contextual con creación de torneo, acceso a cuenta, accesos
  rápidos por relación y navegación adaptada entre web y apps.
- ADR-0058 y `GET /v1/me/leagues`: colección autenticada, paginada y filtrada
  por relaciones de administración o seguimiento; esquema con administradores y
  seguidores explícitos.
- ADR-0059: middleware de sesión para rutas HTTP protegidas, separado de la
  autorización por liga y de la futura protección CSRF.
- Guía de buenas prácticas para el cliente universal: arquitectura, rendimiento,
  virtualización de listas, accesibilidad, localización y condiciones de entrega.
- Cliente Expo SDK 57 con Expo Router, rutas universales, presentación adaptativa
  de deep links y primitives Pulse iniciales (`Screen`, `Text`, `Button`,
  `TextField` y feedback global). La comprobación TypeScript del cliente forma
  parte de `make verify`.
- ADR-0054 y `@tournaments-manager/design-tokens`: dirección Pulse, fundaciones
  compartidas y reglas de feedback de formularios para web e interfaces nativas.
- Primer corte del registro local: `POST /v1/registrations` valida la entrada,
  crea cuenta pendiente, credencial Argon2id y token de verificación hasheado en
  PostgreSQL mediante `sqlc`; Mailpit recibe el correo SMTP local. El consumo del
  token y la sesión se implementarán en el siguiente corte.
- Bootstrap ejecutable de la API Go: configuración explícita, comprobación de
  PostgreSQL al iniciar y endpoint local `GET /healthz`.
- `make api-up` para levantar PostgreSQL local y ejecutar la API Go en el host,
  sin ejecutar migraciones implícitamente.
- ADR-0053: un único esquema PostgreSQL inicial reescribible y reseteable
  mientras los datos sean exclusivamente locales y descartables.

### Changed

- Mailpit local queda fijado a `v1.30.5` con health check de disponibilidad;
  se corrige la etiqueta inexistente `v1.28.5`.
- Autenticación local definida con sesiones opacas revocables, Argon2id y
  renovación silenciosa; JWT y refresh tokens quedan aplazados.
- El primer incremento de backend queda acotado a identidad local, publicación
  y consulta de una liga; el ciclo deportivo avanzado se aplaza.
- Fase 1 cerrada: PostgreSQL local queda validado para arranque, salud,
  persistencia y recuperación mediante reset confirmado.
- `make db-migrate` omite de forma explícita y correcta el estado previo a la
  primera migración; el runbook de PostgreSQL queda validado para arranque,
  salud y persistencia local.
- Gate 0B completado: el primer vertical slice queda definido como una liga de
  fútbol con lectura pública por ID, ciclo de vida, seguimiento autenticado, administración
  delegada limitada a resultados y reglas explícitas de bajas y cancelación.
- Ruta del módulo Go alineada con el propietario canónico de GitHub antes de la
  primera publicación.
- Handbook reorganizado por proyecto, gobierno, ingeniería y operaciones; la
  raíz conserva únicamente documentos de entrada y autoridad.
- El primer login social sobre una cuenta local queda pendiente hasta consumir
  un enlace de vinculación de un solo uso.
- La confirmación usa HTTPS deep linking con fallback web y sustituye la sesión
  local sin cambiar el propietario de una sesión existente.
- ADR-0061: el deep link de registro abre una ruta mediante `GET` seguro y el
  cliente ejecuta el `POST` de verificación automáticamente, revoca la sesión
  anterior presentada y termina en Home sin conservar el token en navegación.
- El alcance inicial de torneos se ajusta a privacidad por defecto y acceso
  invitado limitado, dejando la visibilidad pública como decisión futura.

### Added

- Swagger UI local para explorar `contracts/openapi/v1/openapi.yaml` mediante
  `pnpm run openapi:ui` o `make openapi-ui`, sin incluirlo en la API desplegable.
- ADR-0050: login federado con Google incluido en el primer incremento; Apple
  queda aplazado.
- ADR-0049: lectura pública de ligas visibles por ID, sin token ni tabla de
  enlaces compartidos.
- ADR-0044: sesiones opacas, SMTP local y hashing de contraseñas.
- ADR-0043: alcance del primer incremento de backend.
- ADR-0031 a ADR-0042: borradores previos al acceso y decisiones de producto del
  primer vertical slice; retrospectiva de cierre de la Fase 0.
- ADR-0030: región AWS España (`eu-south-2`), VPC independientes en dos AZ y
  autorización explícita de coste antes de crear recursos.
- ADR-0029: ALB público como entrada, API Fargate restringida al ALB y
  PostgreSQL privado; sin NAT Gateway inicial y sin recursos AWS creados.
- ADR-0028: HCP Terraform Free como backend remoto inicial, con ejecución local
  y sin auto-apply; región, red y recursos AWS permanecen pendientes.
- ADR-0027: estado Terraform local mientras no exista infraestructura AWS real;
  backend remoto HCP Terraform Free o S3 pendiente antes del primer apply cloud.
- ADR-0026: AWS Organizations con cuentas `management`, `nonprod` y `prod`,
  IAM Identity Center, MFA y credenciales temporales; sin cuentas ni recursos
  AWS creados aún.
- ADR-0025: Terraform como herramienta IaC para la Fase 5; cuenta AWS,
  estado remoto, red y recursos siguen pendientes y no se crea gasto cloud.
- ADR-0024: ECR privado, tags inmutables y promoción por digest; futuros
  entornos dev, staging y prod con releases selectivas solo cuando exista un
  disparador real, sin crear recursos AWS aún.
- ADR-0023: ECS con Fargate como runtime cloud futuro; no se crearán recursos
  AWS ni gasto hasta abrir la Fase 5.
- ADR-0022: la API se empaquetará como imagen OCI/Docker de runtime mínimo;
  runtime, registry y promoción permanecen como decisiones separadas.
- ADR-0021: CI informativa que ejecuta `make verify` y puerta de calidad local;
  sin PR ni checks obligatorios mientras el trabajo sea individual.
- ADR-0020: observabilidad mínima correlacionada con OpenTelemetry,
  Prometheus, Grafana, Loki y Tempo; logs JSON, métricas y trazas consultables
  en una interfaz común, sin OpenTelemetry Collector inicialmente.
- ADR-0019: pruebas por riesgo y capas; `testing` y `httptest` como base,
  PostgreSQL real en una base efímera para integración y E2E mínimos para flujos
  críticos.
- ADR-0018: Docker Compose para PostgreSQL local; API Go y cliente Expo en host
  durante desarrollo, sin contenedores de frontend locales.
- Entorno PostgreSQL 18.4 reproducible mediante Compose, contratos locales
  separados, volumen persistente, health check y comandos `make db-*`.
- Runbook de arranque, migración, inspección, parada y reset explícito de
  PostgreSQL local.
- Automatización separada por tecnología en `mk/go.mk`, `mk/typescript.mk` y
  `mk/postgres.mk`; el `Makefile` raíz conserva los comandos transversales.
- ADR-0017: `.env` locales ignorados, `.env.example` como contrato, GitHub
  Environments y OIDC para configuración y secretos.
- ADR-0016: rendering web client-side inicial y adaptación por plataforma
  explícita para el cliente universal.
- ADR-0015: Expo, Expo Router y CNG; rutas universales y proyectos nativos
  generados que no se versionan.
- ADR-0014 y baseline ejecutable para Node 24 LTS, pnpm, TypeScript estricto,
  ESLint, Prettier, workspaces y formato al guardar.
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
