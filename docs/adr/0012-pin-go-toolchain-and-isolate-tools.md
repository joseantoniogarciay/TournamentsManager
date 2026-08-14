# ADR-0012: Fijar el toolchain Go y aislar las herramientas

- **Estado:** Aceptado
- **Fecha:** 2026-07-24
- **Decisor:** Usuario, mediante confirmación explícita
- **Propietario del análisis:** Asistente como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

El backend necesita una versión reproducible de Go y una política común de
formato, análisis estático, herramientas y comandos locales. Las mismas
comprobaciones deben poder ejecutarse desde el editor, la terminal y CI sin
depender de binarios globales con versiones distintas.

## Contexto y restricciones

- El backend vive dentro de un monorepo con otros ecosistemas.
- El manifiesto prioriza simplicidad, aprendizaje y entornos reproducibles.
- PostgreSQL, `pgx`, `sqlc` y `goose` están aceptados por ADR-0011.
- Las herramientas Go modernas pueden declararse mediante `tool` y ejecutarse con
  `go tool`.
- Las dependencias de herramientas participan en el grafo del módulo que las
  declara.
- El editor es una ayuda local; los comandos versionados y CI son la autoridad.

## Criterios de decisión

1. reproducir la misma versión en desarrollo y CI;
2. evitar instalaciones globales implícitas;
3. mantener separado el grafo de ejecución del grafo de herramientas;
4. proporcionar feedback rápido y comandos fáciles de recordar;
5. cubrir formato, errores, seguridad y recursos de PostgreSQL;
6. conservar un coste de mantenimiento razonable.

## Alternativas

### Alternativa A — Herramientas en el `go.mod` principal

Declarar `goimports`, golangci-lint, `sqlc`, `goose` y `govulncheck` con
directivas `tool` en el módulo del backend.

- **Ventajas:** mecanismo oficial directo y comandos cortos con `go tool`.
- **Inconvenientes:** las dependencias de herramientas participan en el mismo
  grafo que las dependencias de ejecución.
- **Coste de adopción:** bajo.
- **Coste de mantenimiento:** bajo.
- **Riesgos:** una actualización de tooling puede modificar la selección de
  versiones del backend.

### Alternativa B — `go.tool.mod` alternativo

Mantener el módulo de aplicación en `go.mod` y seleccionar un módulo alternativo
de herramientas mediante `-modfile=go.tool.mod`.

- **Ventajas:** grafos y checksums independientes; herramientas pineadas; no
  requiere instalaciones globales.
- **Inconvenientes:** `go.tool.mod` no se descubre automáticamente y todos los
  comandos de herramientas necesitan `-modfile`.
- **Coste de adopción:** medio; requiere Makefile y wrapper para el editor.
- **Coste de mantenimiento:** bajo o medio.
- **Riesgos:** olvidar el parámetro y ejecutar una herramienta diferente.

### Alternativa C — Binarios globales

Instalar cada herramienta mediante Homebrew o `go install ...@latest`.

- **Ventajas:** invocaciones cortas y configuración inicial familiar.
- **Inconvenientes:** versiones distintas por máquina, estado fuera del
  repositorio y actualizaciones no revisadas.
- **Coste de adopción:** bajo.
- **Coste de mantenimiento:** medio por deriva.
- **Riesgos:** resultados diferentes entre desarrollo y CI.

## Comparación

La alternativa A es la mínima suficiente para pocos comandos, pero el conjunto
aceptado incluye herramientas con grafos amplios como golangci-lint, `sqlc` y
`goose`. La alternativa B añade un parámetro mecánico que puede ocultarse tras
Make y conserva una frontera verificable. La alternativa C no satisface la
reproducibilidad.

## Recomendación

**Opinión/recomendación:** alternativa B. Usar Go 1.26.6, un módulo de aplicación
en `apps/backend/go.mod`, herramientas declaradas en
`apps/backend/go.tool.mod`, comandos de Make desde la raíz y un wrapper
versionado para `goimports` en VS Code.

## Decisión del usuario

**Aceptada:** establecer:

- Go 1.26.6 como toolchain exacto;
- `go 1.26.0` como versión mínima del módulo y `toolchain go1.26.6`;
- Go modules para el backend, sin `go.work` mientras exista un solo módulo Go;
- `goimports` como formateador;
- golangci-lint v2 con lista explícita de linters;
- `govulncheck` como análisis de vulnerabilidades conocidas;
- herramientas declaradas con `tool` en un `go.tool.mod` separado;
- versiones exactas y checksums independientes en `go.tool.sum`;
- Make como entrada común para formato, lint, test, build y verificación;
- formato al guardar en VS Code mediante el `goimports` pineado.

Versiones iniciales:

| Componente | Versión |
|---|---|
| Go | 1.26.6 |
| `goimports` (`golang.org/x/tools`) | 0.48.0 |
| golangci-lint | 2.12.2 |
| `govulncheck` (`golang.org/x/vuln`) | 1.6.0 |
| `sqlc` | 1.31.1 |
| `goose` | 3.27.1 |

## Reglas de implementación

- `go.mod` contiene dependencias necesarias para compilar el backend.
- `go.tool.mod` contiene herramientas y sus dependencias.
- `go.sum` y `go.tool.sum` no se sustituyen entre sí.
- No se usa `@latest` en comandos versionados ni en CI.
- `make format`, `make tidy`, `make tidy-tools` y `make generate` pueden modificar
  archivos.
- `make check` y `make verify` solo verifican; no corrigen silenciosamente.
- `make check` agrupa formato, lint y tests rápidos.
- `make verify` añade limpieza de ambos módulos, build y vulnerabilidades.
- VS Code puede automatizar el formato, pero CI vuelve a comprobarlo.
- Las excepciones `//nolint` deben nombrar el linter y explicar el motivo.
- Las actualizaciones de patch o minor requieren cambio revisado y verificación;
  una major o un cambio de política reabre este ADR.

**Actualización de seguridad — 2026-08-14:** el usuario autorizó adelantar Go
1.26.6 sin esperar una ventana de maduración, ya que `govulncheck` identificó
vulnerabilidades alcanzables corregidas en ese parche. Mantiene la misma línea
minor y la misma política; se verifica con `make verify` antes de promoción.

## Ruta canónica del módulo

La autenticación del propietario y la disponibilidad del repositorio se
verificaron antes de la primera publicación. El módulo usa
`github.com/joseantoniogarciay/TournamentsManager/apps/backend`, que coincide con
la URL canónica del remoto.

## Configuración inicial de lint

La lista explícita evita que una actualización habilite reglas nuevas sin
decisión:

- `bodyclose`;
- `errcheck`;
- `errorlint`;
- `gosec`;
- `govet`;
- `ineffassign`;
- `misspell`;
- `nilerr`;
- `nilnesserr`;
- `noctx`;
- `nolintlint`;
- `revive`;
- `rowserrcheck`;
- `sqlclosecheck`;
- `staticcheck`;
- `unused`.

`gosimple` no se configura por separado porque golangci-lint v2 lo integra en
`staticcheck`. `typecheck` tampoco se habilita: es una comprobación estructural,
no un linter configurable.

## Consecuencias

### Positivas

- Desarrollo y CI ejecutan herramientas versionadas.
- Las herramientas no modifican el grafo de ejecución del backend.
- `goimports` aplica formato e imports de forma consistente.
- Los comandos frecuentes tienen nombres cortos y documentados.
- Los fallos de errores, recursos, SQL y seguridad aparecen antes del merge.

### Negativas y deuda aceptada

- Cada invocación real de herramienta necesita `-modfile=go.tool.mod`.
- El Makefile y el wrapper de VS Code forman parte del contrato de desarrollo.
- `go.tool.sum` puede ser considerable por las dependencias de golangci-lint.
- Algunos linters necesitarán exclusiones justificadas al aparecer casos reales.
- `govulncheck` consulta `vuln.go.dev` y puede tardar más que los checks rápidos.
  No envía el código fuente, pero necesita red y comparte la información de
  módulos necesaria para consultar vulnerabilidades conocidas.

## Validación

- `go version` devuelve Go 1.26.6.
- El módulo principal y el de herramientas pasan sus respectivos `tidy -diff`.
- `go tool -modfile=go.tool.mod` localiza las herramientas aceptadas.
- `make check` falla ante código sin formato, lint inválido o tests fallidos.
- `make verify` añade build y análisis de vulnerabilidades.
- Mientras no exista ningún paquete Go, los checks que necesitan cargar código se
  omiten con un mensaje explícito; se activan automáticamente con el primer
  archivo `.go`.
- Guardar un archivo Go en VS Code aplica el wrapper versionado de `goimports`.
- La configuración de golangci-lint v2 se valida con la versión pineada.

## Disparadores de revisión

- Una herramienta no puede ejecutarse de forma fiable mediante `go tool`.
- La separación de módulos introduce más incidencias que aislamiento útil.
- Los tiempos de `make check` dejan de ofrecer feedback rápido.
- Un linter produce ruido sostenido o no detecta una categoría de fallo relevante.
- Se añade un segundo módulo Go y aparece una necesidad real de `go.work`.
- Una nueva versión major de Go o golangci-lint cambia el comportamiento acordado.

## Documentación afectada

- [TECHNICAL_BASELINE.md](../governance/TECHNICAL_BASELINE.md)
- [DEVELOPMENT.md](../engineering/DEVELOPMENT.md)
- [STYLEGUIDE.md](../engineering/STYLEGUIDE.md)
- [TESTING.md](../engineering/TESTING.md)
- [DECISIONS.md](../governance/DECISIONS.md)
- [LEARNING.md](../project/LEARNING.md)

## Fuentes técnicas

- [Go 1.26 release notes](https://go.dev/doc/go1.26)
- [Go release history](https://go.dev/doc/devel/release)
- [Directiva `tool` y `go tool`](https://go.dev/ref/mod#tool-directive)
- [`go` y el flag `-modfile`](https://pkg.go.dev/cmd/go)
- [Módulo `goimports`](https://pkg.go.dev/golang.org/x/tools/cmd/goimports)
- [Configuración de golangci-lint](https://golangci-lint.run/docs/configuration/file/)
- [Linters de golangci-lint](https://golangci-lint.run/docs/linters/)
- [Gestión de vulnerabilidades de Go](https://go.dev/doc/security/vuln/)
- [Configuración de la extensión Go para VS Code](https://github.com/golang/vscode-go/wiki/settings)
