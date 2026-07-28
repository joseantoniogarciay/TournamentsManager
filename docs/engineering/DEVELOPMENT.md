# Desarrollo

> Estado: toolchains Go y TypeScript aceptados en
> [ADR-0012](../adr/0012-pin-go-toolchain-and-isolate-tools.md) y
> [ADR-0014](../adr/0014-use-node-pnpm-and-strict-typescript.md); el modelo de
> entorno local está aceptado en [ADR-0018](../adr/0018-use-compose-for-local-service-dependencies.md)
> y su implementación está validada localmente.

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
- Redocly valida el contrato y Orval regenera `apps/client/src/api/generated/`;
  la comprobación de deriva forma parte de `make verify`.
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
make sqlc-generate
```

Comandos que solo verifican:

```bash
make format-check
make tidy-check
make tidy-tools-check
make sqlc-generate-check
make lint
make test
make build
make vuln
make check
make verify
```

`make check` agrupa formato, lint y tests. `make verify` añade la exportación
web del cliente Expo a `/tmp/tournaments-manager-web-export`, la limpieza de
ambos módulos, build y vulnerabilidades. La exportación comprueba Expo Router y
el bundling sin crear un artefacto versionado. Conforme a
[ADR-0021](../adr/0021-use-advisory-ci-with-local-quality-gate.md), CI ejecuta
`make verify` como comprobación informativa; el mismo comando se ejecuta
localmente antes de promover un bloque a `main`.

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
pnpm run openapi:ui
pnpm run check
pnpm run verify
```

`pnpm run openapi:ui` (o `make openapi-ui`) abre Scalar localmente en
`http://127.0.0.1:8082` para explorar el contrato OpenAPI 3.1. No arranca el
backend ni publica documentación. El contrato configura `http://127.0.0.1:8080/v1`
como destino de la API local cuando el servidor Go esté disponible.

El Makefile raíz incorpora estas comprobaciones en `make format`, `make check` y
`make verify`, junto con las de Go.

## Cliente Expo

El cliente usará Expo, Expo Router y CNG conforme a
[ADR-0015](../adr/0015-use-expo-router-and-continuous-native-generation.md). La
web usará rendering client-side inicialmente conforme a
[ADR-0016](../adr/0016-use-client-side-web-rendering-initially.md).
`apps/client` existe como proyecto Expo SDK 57. Expo Router usa `src/app` como
raíz de rutas; las primitivas compartidas viven en `src/shared` y los tokens en
`packages/design-tokens`.

- Las pantallas futuras vivirán en `apps/client/src/app`; sus rutas derivarán de
  los archivos.
- `apps/client/ios` y `apps/client/android` se generarán bajo demanda y están
  ignorados por Git.
- Una development build se compila con `expo run:ios` o `expo run:android`; los
  cambios exclusivos de TypeScript se recargan con `expo start`.
- No se edita a mano un directorio generado. Las necesidades nativas se declaran
  en configuración o config plugins.
- Las diferencias web/native se aíslan en componentes, adaptadores o archivos
  específicos de plataforma cuando exista una razón concreta.

Para iniciar el cliente:

```bash
pnpm --filter @tournaments-manager/client start
pnpm --filter @tournaments-manager/client web
```

La exportación web forma parte de `make verify` y se puede ejecutar de forma
aislada con `make client-web-export`. Los directorios `.expo`, `ios` y `android`
siguen sin versionarse; se generan solo mediante operaciones explícitas de Expo.

## Configuración local

La configuración y los secretos siguen
[ADR-0017](../adr/0017-use-env-contracts-github-environments-and-oidc.md).

- Los `.env` reales son locales y están ignorados por Git.
- Los `.env.example` futuros serán contratos versionados por app o servicio.
- El backend fallará al arrancar si falta configuración obligatoria o tiene
  formato inválido.
- El cliente Expo solo podrá leer desde JavaScript variables `EXPO_PUBLIC_*`, que
  se tratarán como públicas.
- Docker Compose podrá usar `env_file` cuando se decida el entorno local.
- No se introduce gestor de secretos hasta que exista evidencia operativa.

### Arranque de la API Go

Para levantar PostgreSQL local y la API desde el host:

```bash
make api-up
```

`make api-up` espera a que PostgreSQL esté saludable, carga
`apps/backend/.env` y mantiene la API en primer plano. El contrato local exige
`DATABASE_URL` y `HTTP_ADDR` en ese archivo. El proceso comprueba PostgreSQL
antes de abrir el puerto y expone `GET /healthz` en
`HTTP_ADDR` (por defecto, `http://127.0.0.1:8080/healthz`). Las migraciones
siguen siendo un paso explícito mediante `make db-migrate`; la API se detiene
con `Ctrl+C` y PostgreSQL, si se desea, con `make db-down`.

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

## Entorno local aceptado; implementación pendiente

ADR-0018 delimita Docker Compose a dependencias de infraestructura: inicialmente
solo PostgreSQL. La API Go y el cliente Expo —web, iOS y Android— se ejecutarán
en el host durante el desarrollo. No habrá contenedores de frontend localmente.

La implementación fija PostgreSQL 18.4, `infra/local/compose.yaml`, contratos
`.env.example` separados para Compose y backend, health check, volumen y los
comandos `make db-*`. El [Makefile](../../Makefile) raíz es el índice de
automatización compartida; los comandos se separan por tecnología en
[`mk/go.mk`](../../mk/go.mk), [`mk/typescript.mk`](../../mk/typescript.mk) y
[`mk/postgres.mk`](../../mk/postgres.mk). Consulta el [runbook de PostgreSQL local](../runbooks/local-postgresql.md).
Las migraciones `goose` se ejecutan separadamente con `make db-migrate`. Los
datos semilla funcionales permanecen aplazados hasta cerrar el primer vertical
slice de producto.
