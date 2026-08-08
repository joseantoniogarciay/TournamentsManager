# Desarrollo

> Estado: toolchains Go y TypeScript aceptados en
> [ADR-0012](../adr/0012-pin-go-toolchain-and-isolate-tools.md) y
> [ADR-0014](../adr/0014-use-node-pnpm-and-strict-typescript.md); el modelo de
> entorno local está aceptado en [ADR-0076](../adr/0076-run-the-local-api-in-compose-with-air.md)
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
- El esquema inicial se aplica explícitamente y no al arrancar la API; Goose no
  se ejecuta durante la primera versión.
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
- La instalación local y CI usan el lockfile congelado; una instalación normal
  no puede alterar la resolución. Los cambios intencionados usan `pnpm add`,
  `pnpm update` o `pnpm install --no-frozen-lockfile` y requieren revisar el
  diff resultante.
- pnpm exige que cada versión nueva, directa o transitiva, tenga al menos siete
  días (10 080 minutos). Si faltan metadatos temporales o no hay una versión
  madura compatible, falla en lugar de instalar una versión más reciente.
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

El cliente declara dos variantes de aplicación en `apps/client/app.config.ts`:
`development` muestra **Fast Tourney Dev** y `production` muestra **Fast
Tourney**. Se seleccionan con `APP_ENV` y comparten un icono de 1024 × 1024.
Los comandos `start:dev`, `ios:dev` e `ios:prod` del workspace cliente evitan
configurar la variante a mano. La configuración de Expo y sus config plugins
siguen siendo la fuente de verdad: no se crean targets Xcode persistentes para
separar entornos mientras CNG pueda generar sus diferencias de forma declarativa.
Las variantes iOS declaran `com.fasttourney.app.dev` y `com.fasttourney.app`.
Antes de distribuir una build nativa se verificará que ambos identificadores
estén registrados y controlados en la cuenta de Apple.

La splash nativa se configura mediante el plugin `expo-splash-screen`: usa el
icono local de Fast Tourney sobre las superficies `canvas` clara y oscura. En
arranque nativo se conserva únicamente hasta que se hidrata la preferencia local
de tema y se puede pintar la primera ruta; después se oculta con un fundido de
240 ms. No espera red ni se sustituye por una pantalla React de carga. Sus
propiedades nativas se validan en una build release, ya que Expo Go y las
development builds no reproducen fielmente la splash distribuida.

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
- `EXPO_PUBLIC_API_BASE_URL` indica la base pública de la API para el cliente;
  en desarrollo, si no se declara, usa `http://127.0.0.1:8080/v1`. En Android
  físico o emulador debe apuntar a una dirección alcanzable desde el dispositivo,
  no a su propio `127.0.0.1`.
- `CORS_ALLOWED_ORIGINS` es una lista separada por comas de orígenes web
  completos sin path. El backend falla si falta o incluye un valor inválido. En
  local se permiten `http://localhost:8081` y `http://127.0.0.1:8081`; cada
  entorno desplegado declara solo sus dominios web reales.
- El contrato de la API en Compose vive en `infra/local/api.docker.env`; usa
  nombres de servicio (`postgres` y `mailpit`). El contrato host
  `apps/backend/.env` conserva `127.0.0.1` y solo sirve para ejecutar la API
  fuera de Docker de forma puntual.
- No se introduce gestor de secretos hasta que exista evidencia operativa.

### Arranque de la API Go en Compose

En un clon nuevo, crea los dos contratos de Compose una sola vez:

```bash
make dev-init
```

Después, para levantar API, PostgreSQL y Mailpit:

```bash
make dev-up
```

`make dev-up` valida los contratos sin sobrescribirlos, espera a que PostgreSQL
y Mailpit estén saludables y mantiene los logs en primer plano. El servicio API
selecciona el target Docker `dev`: Air recompila y reinicia la API al guardar un
archivo Go. El contrato `infra/local/api.docker.env` exige `DATABASE_URL`,
`HTTP_ADDR`, `SMTP_ADDR`, `SMTP_FROM`, `PUBLIC_BASE_URL` y
`CORS_ALLOWED_ORIGINS`.
`PUBLIC_BASE_URL` es la URL del cliente a la que llega el correo, no la de la
API: en local usa `http://localhost:8081`; así el navegador puede usar la
excepción de desarrollo para cookies `Secure`. Fuera de loopback debe ser HTTPS y
coincidir con `EXPO_PUBLIC_APP_LINK_URL` de la build móvil.
El proceso comprueba PostgreSQL antes de abrir el puerto y expone `GET /healthz`
en `HTTP_ADDR` (publicado como `http://127.0.0.1:8080/healthz`). El esquema inicial
se aplica explícitamente con `make db-schema-apply`; `Ctrl+C` detiene los
contenedores y `make dev-down` los detiene de forma explícita conservando datos.
`make api-image-build` construye el target `runtime`, que no contiene Air,
fuentes ni compilador.

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

## Entorno local aceptado

ADR-0076 ejecuta API, PostgreSQL y Mailpit en Docker Compose. El Dockerfile de
la API comparte una etapa de módulos: `dev` añade Air y recibe el código mediante
bind mount; `runtime` solo recibe el binario. El cliente Expo —web, iOS y
Android— sigue en host; no hay contenedor de frontend local.

La implementación fija PostgreSQL 18.4, `infra/local/compose.dev.yaml`,
contratos `.env.example` separados y health checks, volúmenes y comandos
`make dev-*`/`make db-*`. El [Makefile](../../Makefile) raíz es el índice de
automatización compartida; los comandos se separan por tecnología en
[`mk/go.mk`](../../mk/go.mk), [`mk/typescript.mk`](../../mk/typescript.mk) y
[`mk/postgres.mk`](../../mk/postgres.mk). Consulta el [runbook de PostgreSQL local](../runbooks/local-postgresql.md).
El esquema inicial se aplica separadamente con `make db-schema-apply`. Los
datos semilla funcionales permanecen aplazados hasta cerrar el primer vertical
slice de producto.
