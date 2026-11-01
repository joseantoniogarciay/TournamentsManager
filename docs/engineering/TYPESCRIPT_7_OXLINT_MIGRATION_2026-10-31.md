# Migración TypeScript 7 y Oxlint — 2026-10-31

## Decisión y alcance

El usuario acepta Oxlint + Prettier y TypeScript 7.0.2 tras comparar convivencia
con TS 6 y Biome. [ADR-0171](../adr/0171-use-typescript-7-and-oxlint-with-prettier.md)
registra la decisión antes de implementar. Se conservan los cambios previos del
workspace; no se integra una rama ni se publica o despliega.

| Responsabilidad | Antes | Después |
| --- | --- | --- |
| Compilador del proyecto | TypeScript 6.0.3 | TypeScript 7.0.2 nativo |
| Lint | ESLint 10.13.0/typescript-eslint 8.72.0 | Oxlint 1.89.0 |
| Formato | Prettier 3.9.10 | Prettier 3.9.10 |
| Transformación TS/TSX en pruebas | ts.transpileModule | esbuild 0.28.1 |
| AST de navegación en una prueba | ts.createSourceFile | Babel Parser 7.29.9 |
| Servicio de editor | tsserver/lib | Extensión oficial TypeScript 7 |
| Generación OpenAPI | Orval 8.44.0 | Orval 8.44.0 |

No se añade Java ni se sustituye el generador OpenAPI. El cliente generado ya
compila con TS 7. Se descarta OpenAPI Generator en esta sesión por preferencia
del usuario de evitar Java adicional. Las alternativas npm revisadas tampoco
acreditan retirada de la API anterior: openapi-typescript 7.13.0 requiere TS
`^5.x`; Hey API 0.99.0 declara un rango amplio, pero el código publicado importa
`createSourceFile`, `createPrinter` y `factory` de TypeScript.

## Reglas efectivas y diferencias

La conversión automática de @oxlint/migrate 1.89.0 se usa solo como evidencia;
el paquete se retira al terminar. Su salida pierde los overrides AND de archivos
TypeScript y omite reglas nursery. Se comparan las configuraciones efectivas
de ESLint para TSX y MJS antes de retirarlo, no solo los presets con nombre.
`.oxlintrc.json` conserva sus opciones explícitas y diferencia JS de TS.

| Baseline efectivo | Oxlint | Sustitución por parser |
| --- | --- | --- |
| JavaScript: 90 reglas activas | 88 reglas | no-dupe-args y no-octal |
| TypeScript: 76 reglas activas | 75 reglas | no-octal |

- `no-undef` y `no-useless-assignment` se activan explícitamente aunque estén
  clasificadas como nursery: ya estaban activas y no se pierde su cobertura.
- Las extensiones de TypeScript que Oxlint implementa con la regla base se
  mapean a ese nombre; las demás usan `typescript/`. Los overrides de TS
  conservan `no-var`, `prefer-const`, `prefer-rest-params` y `prefer-spread`,
  además de las desactivaciones de reglas que cubre el compilador.
- `preserve-caught-error` no acepta `errorClassNames`; se elimina la lista vacía
  de ESLint, conservando `requireCatchParameter: false` y el comportamiento
  predeterminado. No se trasladan opciones no soportadas silenciosamente.
- Los dos CommonJS del cliente declaran `"use strict"`. Oxlint no dispone de
  `no-octal` separado y permite octal en un CJS sin esa directiva. La regresión
  recorre los CJS escritos en el repo y exige modo estricto; otra prueba
  demuestra que octal y parámetros duplicados se rechazan con ese modo.
- Los globals de CommonJS y del script de tema se declaran por ámbito en el
  config; se retiran comentarios ESLint que Oxlint no interpreta. La excepción
  `no-require-imports` sigue limitada a CJS, requerido por Metro/Expo.
- Categorías automáticas apagadas y reglas explícitas: no se incorpora otro
  preset por accidente ni se añade tsgolint o lint con información de tipos.

La equivalencia funcional de estas reglas no implica que ambos motores
produzcan diagnósticos idénticos en todo programa posible. Las tres regresiones
de tooling comprueban rechazos reales y configuración crítica, además del lint
de todas las fuentes actuales.

## API antigua e instalación reproducible

Doce archivos de pruebas dejan de importar TypeScript. Un helper transforma
CommonJS/ES2022, con JSX automático cuando procede, y conserva los stubs/VM
existentes. esbuild asigna `module.exports`; la fixture de refresh se adapta
exponiendo `module`, sin cambiar sus expectativas. Babel Parser extrae la
declaración por nodos y offsets, sin sustituir el AST por una expresión regular.

El tsconfig raíz comprueba el helper y el hook de instalación, sustituyendo la
única fuente anterior, eslint.config.mjs, retirada. No se añaden flags para
ocultar errores de tipos. Probar además orval.config.ts descubrió declaraciones
incompletas de dependencias externas (MCP y picomatch); ampliar ese alcance no
formaba parte del baseline y no se añade skipLibCheck. Orval sigue validándose
mediante ejecución y comparación determinista de su salida.

Orval/core ya usa TS 6.0.3 como dependencia interna. TypeDoc 0.28.20 opcional
exige además la API anterior, pero pnpm enlaza por defecto su peer al TS 7 raíz.
Añadir una dependencia con packageExtensions o un override del rango del peer
no cambió ese enlace en la prueba real. `.pnpmfile.mjs` transforma exclusivamente
ese peer exacto en dependencia privada 6.0.3. La resolución comprobada muestra
raíz 7.0.2 y TypeDoc 6.0.3; no se amplían rangos para ocultar un aviso. Retirar
el hook al actualizar TypeDoc con soporte real del compilador nativo.

El grafo final añade 41 versiones y retira 74 frente al inicio de esta migración.
Se verifican las 41 fechas oficiales, incluidos binarios de otras plataformas,
y las dos versiones ya existentes declaradas directamente para pruebas:
43/43 maduras. No cambia la política de siete días ni sus excepciones Expo.

## Verificación y cierre

- `pnpm run check`: Prettier, Oxlint sin avisos, tipos y contrato OpenAPI correctos.
- 125 pruebas Node correctas: 122 anteriores y tres de tooling. Las pruebas
  incluyen transporte, errores seguros, refresco de sesión y carga CommonJS.
- `make test-toolchain`: tres pruebas correctas; integrado en `make verify`.
- Dos generaciones OpenAPI idénticas entre sí y con los 145 archivos de partida.
- Exportación web correcta en `/tmp/tm-typescript7-web-20261031`.
- Instalación congelada correcta sin volver a resolver dependencias.
- Peers previos de Hono/node-server y neverpanic/TypeScript siguen pendientes;
  no se silencian. neverpanic declara peer TS 5 pero su runtime publicado no
  importa la API de TypeScript. El peer nuevo de TypeDoc queda resuelto realmente.
- Auditoría npm ejecutada tras autorización explícita de envío del grafo:
  43 avisos, 15 altos, 18 moderados y 10 bajos; cero críticos. Todos los pares
  paquete/versión afectados ya figuran en el lockfile previo a esta migración.
  No hay avisos relativos a las 41 versiones añadidas. Esto acredita procedencia
  de versiones, no ausencia de exposición ni una auditoría limpia.

La checklist de cliente se recorre contra los dos CJS y el script público:
no cambian UI, traducciones, navegación, endpoints ni adaptadores. El preflight
de interfaz no aplica a pantallas porque no se modifica ninguna. Se ejecutan
tipos y exportación web obligatorios. No se recompilan de nuevo apps nativas:
este bloque cambia herramientas y modo estricto de configuración, sin modificar
SDKs nativos; sus builds PostHog están verificadas en el informe anterior.

Logs, mapas de reglas, fechas y hashes se conservan en
`.tmp/typescript7-migration-20261031/`, ignorado por Git. No se arrancan API,
Metro, local/dev ni observabilidad. Los servidores de pruebas loopback terminan
con sus fixtures. No hay prueba de recepción/simbolización PostHog: ADR-0109/0149
sigue pendiente, independiente de la actualización de dependencias.

## Inventario de seguridad pendiente

La [revisión separada](DEPENDENCY_SECURITY_REVIEW_2026-10-31.md) contrasta los
43 registros con 39 GHSA únicos y fechas oficiales. El lote posterior aceptado
en ADR-0172 elimina 28 registros: quedan 15 (6 altos, 6 moderados y 3 bajos),
incluido el recorrido Hono 1.x pendiente de compatibilidad.

La auditoría actual sustituye los conteos históricos de octubre para el grafo
local. Los 43 avisos se agrupan en 12 paquetes; no son 43 paquetes distintos.
Las versiones corregidas indicadas por npm necesitan revisión de publicación,
maduración, compatibilidad y alcance antes de instalarlas. No se ejecuta
`audit --fix`, no se añaden overrides ni se salta a majors automáticamente.

| Paquete afectado | Versiones locales afectadas | Alto / moderado / bajo |
| --- | --- | --- |
| @hono/node-server | 1.19.17, 2.0.10 | 0 / 1 / 0 |
| fast-uri | 3.1.8, 4.1.5 | 0 / 4 / 4 |
| fastify | 5.12.5 | 0 / 0 / 1 |
| glob | 13.0.6 | 0 / 1 / 0 |
| hono | 4.13.7 | 0 / 1 / 0 |
| markdown-it | 14.3.1 | 2 / 0 / 1 |
| minimatch | 3.1.5, 10.2.6 | 0 / 1 / 0 |
| node-forge | 1.4.0 | 6 / 3 / 3 |
| sharp | 0.35.4 | 1 / 0 / 0 |
| undici | 7.29.1 | 4 / 4 / 0 |
| vite | 8.0.16 | 1 / 3 / 1 |
| zod | 4.6.5 | 1 / 0 / 0 |

La revisión automática rechazó el primer envío del grafo por posible divulgación
de nombres internos. El usuario autorizó expresamente ese envío y la segunda
ejecución terminó con el resultado anterior (exit 1 por avisos, no por fallo
de conectividad). Los IDs y enlaces de cada aviso, junto a la comparación con
el lockfile de partida, quedan en `audit.json` y `audit-comparison.json` dentro
del directorio de evidencia. Cerrar la migración no cierra este inventario.

## Retrospectiva

Migrar un linter exige comparar reglas efectivas y probar código erróneo;
un lint verde por sí solo podría ser una configuración vacía. La clasificación
nursery o una conversión automática no deben rebajar controles existentes.
En una major del compilador hay que distinguir tsc del proyecto, la API que usan
las pruebas y los compiladores privados del tooling. Comprobar la resolución
real evita dar por válido un override que solo cambió el rango declarado.

## Fuentes

- [Microsoft: TypeScript 7](https://devblogs.microsoft.com/typescript/announcing-typescript-7-0/).
- [Oxlint: migración](https://oxc.rs/docs/guide/usage/linter/migrate-from-eslint.html).
- [TypeScript 7 para VS Code](https://marketplace.visualstudio.com/items?itemName=TypeScriptTeam.native-preview).
- [pnpm: readPackage y checksum](https://pnpm.io/pnpmfile).
- [esbuild: transform](https://esbuild.github.io/api/#transform).
- [Babel Parser](https://babeljs.io/docs/babel-parser).
- [Manifiesto npm de openapi-typescript](https://registry.npmjs.org/openapi-typescript/7.13.0).
- [Manifiesto npm de Hey API](https://registry.npmjs.org/@hey-api/openapi-ts/0.99.0).
