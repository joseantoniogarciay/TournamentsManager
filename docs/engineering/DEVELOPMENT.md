# Desarrollo

> Estado: toolchains Go y TypeScript aceptados en
> [ADR-0012](../adr/0012-pin-go-toolchain-and-isolate-tools.md) y
> [ADR-0014](../adr/0014-use-node-pnpm-and-strict-typescript.md); el resto del
> entorno local continúa pendiente.

## Principios

- Git conserva la trazabilidad entre decisión, cambio, prueba y documentación.
- El entorno local debe parecerse a producción en comportamiento, no replicar todo
  su coste ni complejidad.
- Cada automatización debe poder explicarse y tener una ruta de diagnóstico.
- Los comandos frecuentes tendrán una única entrada documentada cuando exista
  código.
- Configuración, secretos y datos de ejemplo se tratarán de forma explícita.
- La documentación cambia en el mismo conjunto de cambios que el comportamiento.
- OpenAPI es la fuente editable del contrato HTTP; el cliente TypeScript generado
  no se modifica manualmente.
- El equipo escribe las consultas SQL; el código de acceso generado por `sqlc`
  no se modifica manualmente.
- Las migraciones `goose` se ejecutan explícitamente y no al arrancar la API.
- Toda generación debe ser reproducible mediante un comando versionado y producir
  un diff limpio cuando las entradas no cambian.

## Toolchain Go

- Versión mínima del módulo: Go 1.26.0.
- Toolchain exacto inicial: Go 1.26.5.
- El backend usa `apps/backend/go.mod`.
- Las herramientas usan `apps/backend/go.tool.mod` y
  `apps/backend/go.tool.sum`.
- `goimports`, golangci-lint, `govulncheck`, `sqlc` y `goose` se ejecutan con
  `go tool -modfile=go.tool.mod`.
- El Makefile de la raíz encapsula las rutas del monorepo y el módulo alternativo.
- No se usa `@latest` en automatizaciones versionadas.

## Comandos Go

Durante un cambio:

```bash
go -C apps/backend test ./ruta/del/paquete/...
make check
```

Antes de subirlo:

```bash
make verify
```

Comandos que modifican archivos:

```bash
make format
make tidy
make tidy-tools
make tidy-all
```

Comandos que solo verifican:

```bash
make format-check
make tidy-check
make tidy-tools-check
make lint
make test
make build
make vuln
make check
make verify
```

`make check` agrupa formato, lint y tests. `make verify` añade la limpieza de
ambos módulos, build y vulnerabilidades. CI reutilizará `make verify` cuando se
decida su política.

`make vuln` consulta la base oficial `vuln.go.dev`. No envía el código fuente,
pero requiere red y usa información de módulos para resolver vulnerabilidades.

Durante la fase documental, si no existe ningún paquete Go, Make omite con un
mensaje explícito lint, test, build y vulnerabilidades. Estos checks se activan
automáticamente al añadir el primer archivo `.go`; no se mantienen tests vacíos
para simular cobertura.

## Editor

La configuración versionada de VS Code recomienda la extensión oficial de Go y
ejecuta al guardar `apps/backend/scripts/goimports`. El wrapper selecciona el
`goimports` pineado en `go.tool.mod`; el editor no necesita otro `goimports`
global.

También recomienda ESLint y Prettier, selecciona el TypeScript instalado en el
workspace y formatea JavaScript, TypeScript y sus variantes JSX al guardar.

## Toolchain TypeScript

- Node 24.18.0 LTS se selecciona mediante `devEngines.runtime`.
- pnpm 11.17.0 gestiona un lockfile y los workspaces `apps/*` y `packages/*`.
- El linker aislado exige que cada workspace declare sus dependencias directas.
- TypeScript 6.0.3 usa `strict`, `noUncheckedIndexedAccess`,
  `exactOptionalPropertyTypes`, `noImplicitOverride` y `noEmit`.
- ESLint 10 usa flat config y `typescript-eslint`; Prettier se ocupa únicamente
  del formato.
- `tsconfig.json` comprueba la configuración JavaScript del tooling y añadirá
  referencias al existir aplicaciones o paquetes TypeScript.

Comandos directos:

```bash
pnpm run format
pnpm run format:check
pnpm run lint
pnpm run typecheck
pnpm run check
pnpm run verify
```

El Makefile raíz incorpora estas comprobaciones en `make format`, `make check` y
`make verify`, junto con las de Go.

## Cliente Expo

El cliente usará Expo, Expo Router y CNG conforme a
[ADR-0015](../adr/0015-use-expo-router-and-continuous-native-generation.md).
`apps/client` no se crea hasta completar la decisión de rendering y adaptación.

- Las pantallas futuras vivirán en `apps/client/src/app`; sus rutas derivarán de
  los archivos.
- `apps/client/ios` y `apps/client/android` se generarán bajo demanda y están
  ignorados por Git.
- Una development build se compila con `expo run:ios` o `expo run:android`; los
  cambios exclusivos de TypeScript se recargan con `expo start`.
- No se edita a mano un directorio generado. Las necesidades nativas se declaran
  en configuración o config plugins.

## Flujo de trabajo

1. Formular el problema y comprobar si requiere decisión.
2. Actualizar o crear ADR si supera el umbral.
3. Definir criterios de aceptación y riesgos.
4. Hacer el cambio mínimo que produzca aprendizaje o valor.
5. Ejecutar comprobaciones automáticas y manuales relevantes.
6. Actualizar handbook, changelog y troubleshooting.
7. Registrar el aprendizaje cuando cambie el modelo mental.

## Definition of Done

Un cambio está terminado cuando:

- cumple criterios de aceptación;
- no contradice un ADR aceptado;
- incluye pruebas proporcionales al riesgo;
- no introduce una abstracción sin necesidad demostrable;
- permite observar y diagnosticar su comportamiento;
- actualiza la documentación afectada;
- no contiene secretos ni dependencias no justificadas.
- conserva alineados contrato OpenAPI, implementación Go y cliente TypeScript.

## Entorno local pendiente

Fase 1 decidirá y documentará:

- Docker Compose y ciclo de vida de servicios;
- variables de entorno y secretos locales;
- configuración y rutas de `sqlc`, migraciones `goose` y datos semilla;
- health checks;
- comandos de servicios y cleanup;
- soporte de plataforma.

Hasta entonces no deben inventarse comandos ni prerequisitos adicionales.
