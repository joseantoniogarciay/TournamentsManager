# Actualización de dependencias — 2026-10-31

## Alcance y decisión

[ADR-0169](../adr/0169-upgrade-backend-and-development-dependencies.md) registra
la autorización del listado de mantenimiento y la excepción posterior de x/net.
Se conserva el trabajo previo del workspace, incluida la migración Expo SDK 58.
La revisión usa como partida los manifiestos locales, no certifica versiones
desplegadas. No hay publicación ni despliegue de producción.

| Dependencia | Antes | Después |
| --- | --- | --- |
| Go | 1.26.9 | 1.26.10 |
| pgx, runtime | 5.10.0 | 5.11.0 |
| pgx, módulo de herramientas | 5.9.2 | 5.11.0 |
| OpenTelemetry Go | 1.45.0 | 1.47.0 |
| otelhttp, runtime / herramientas | 0.61.0 / 0.68.0 | 0.72.0 |
| pnpm | 11.17.0 | 12.15.0 |
| Prettier | 3.9.6 | 3.9.10 |
| ESLint | 10.7.0 | 10.13.0 |
| typescript-eslint | 8.65.0 | 8.72.0 |
| Orval | 8.23.0 | 8.44.0 |
| Redocly CLI | 2.41.0 | 2.71.0 |
| x/net | 0.60.0 | 0.62.0 |

Node 24.18.0, TypeScript 6.0.3, PostHog React Native 4.78.4 y su plugin 2.12.3
conservan sus versiones, al igual que React y React Native. El Dockerfile y ambos
módulos Go seleccionan el compilador 1.26.10; el binario global de Go no se cambia.

## Compatibilidad y maduración

Se verifican fechas oficiales de 70 versiones npm nuevas y 25 módulos Go
cambiados, incluidos requisitos transitivos. Todas son maduras salvo x/net
0.62.0, autorizado explícitamente. Su actualización requiere x/crypto 0.58.0,
x/text 0.43.0, x/sync 0.24.0 y x/sys 0.49.0; el módulo de herramientas requiere
además x/tools 0.51.0 y una revisión de x/telemetry de septiembre. Sus fechas
cumplen la espera. No se desactiva la política de pnpm ni se amplía la excepción
Expo. La presencia transitiva del módulo `otel/log` no activa un pipeline de logs.

pnpm 12 rechaza los parches unified sin cabeceras Git. Se añaden exclusivamente
líneas `diff --git` a los tres parches Expo, conservando todos sus hunks. El
lockfile recalcula sus hashes. La instalación congelada posterior pasa sin
resolver de nuevo ni modificar el lockfile. pnpm acepta las opciones de política
existentes y confirma la verificación de suministro.

Los avisos de peers `@hono/node-ws`/`@hono/node-server` y `neverpanic`/TypeScript
ya estaban presentes en el grafo anterior; no se cambian overrides de seguridad
ni TypeScript para ocultarlos. Continúan avisos de tres dependencias obsoletas.

Orval regenera 145 archivos. El cambio funcional revisado normaliza cabeceras
Fetch recibidas como objetos `Headers` o pares iterables antes de combinarlas con
Content-Type, conservando `fetchFn` y el transporte común. Se retira la versión
del generador de las cabeceras por comportamiento upstream. Contrato HTTP,
estados de negocio y mensajes de usuario conservan su definición.

## Verificación

- `pnpm run check`: formato, ESLint, tipos y Redocly correctos. La única regla
  ignorada de Redocly ya estaba configurada.
- Suite Node: 122 pruebas correctas tras repetir las de loopback con permiso
  de escucha; el primer intento tuvo siete fallos EPERM del sandbox. La
  repetición de App Links/Cloudflare pasa sus 13 pruebas, incluidos esos siete
  casos. Las regresiones incluyen parches de dependencias y generación Android.
- Seguridad operativa: 11 pruebas Python correctas.
- Exportación web estática correcta en `/tmp/tm-dependency-upgrade-web-20261031`.
- Dos generaciones OpenAPI consecutivas producen los mismos 145 archivos.
  No se usa un diff contra HEAD como prueba de determinismo porque la
  actualización autorizada todavía está sin commit.
- Tests Go, detector de carreras, build, lint y formato; ambos módulos limpios
  con `tidy -diff`, y generación sqlc sin deriva.
- Integración PostgreSQL 18.6 sobre un contenedor efímero independiente:
  62 pruebas funcionales; la comparación de rendimiento opt-in queda omitida.
- Imagen oficial `golang:1.26.10-bookworm` comprobada mediante manifiesto;
  no se construye ni despliega una imagen productiva.

Los logs y auditorías locales se conservan en
`.tmp/dependency-upgrade-20261031/`, ignorado por Git. Los servicios local/dev y
observabilidad permanecen apagados; el contenedor efímero se retira al terminar.

## Seguridad y retrospectiva

La primera ejecución de govulncheck detectó GO-2026-6860 y GO-2026-6866 en
`html.Parse`/`html.Render` de la preview social. La corrección se publicó el
27 de octubre y todavía no cumplía siete días. Se pidió y obtuvo la excepción
exacta antes de modificar x/net; no se califican los avisos como críticos.

Con el grafo final, `govulncheck` termina correctamente: cero vulnerabilidades
alcanzables y cero avisos en paquetes importados. El único aviso a nivel de
módulo es [GO-2026-5932](https://pkg.go.dev/vuln/GO-2026-5932), relativo a
`golang.org/x/crypto/openpgp`, que no se importa. No tiene versión corregida;
no se afirma que todos los paquetes del módulo x/crypto estén libres de avisos.
El backend consume Argon2id, no OpenPGP. La repetición final de integración
mantiene 62 pruebas correctas y una comparación de rendimiento omitida.

La actualización de herramientas requiere validar su resolución y su salida:
el parser de parches y las cabeceras Fetch muestran cambios reales aunque el
contrato de producto no cambie. Separar pruebas funcionales, maduración y
seguridad evita presentar un build correcto como ausencia de vulnerabilidades.

## Fuentes

- [Historial Go](https://go.dev/doc/devel/release#go1.26.10).
- [pgx 5.11](https://github.com/jackc/pgx/releases/tag/v5.11.0).
- [OpenTelemetry Go 1.47](https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.47.0).
- [pnpm 12](https://github.com/pnpm/pnpm/releases/tag/v12.0.0).
- [Orval 8.44](https://github.com/orval-labs/orval/releases/tag/v8.44.0).
- [GO-2026-6860](https://pkg.go.dev/vuln/GO-2026-6860),
  [GO-2026-6866](https://pkg.go.dev/vuln/GO-2026-6866).
- [Registro npm](https://registry.npmjs.org/) y
  [proxy Go](https://proxy.golang.org/): fechas exactas de publicación.
