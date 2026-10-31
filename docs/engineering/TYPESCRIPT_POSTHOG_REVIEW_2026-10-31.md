# Revisión de TypeScript y actualización PostHog — 2026-10-31

Esta revisión inicial conserva el resultado de PostHog. El aplazamiento de
TypeScript se supera después por ADR-0171: véase la
[migración a TypeScript 7 y Oxlint](TYPESCRIPT_7_OXLINT_MIGRATION_2026-10-31.md).

## Alcance y evidencia

[ADR-0170](../adr/0170-update-mature-posthog-and-review-typescript-7.md) registra
el mantenimiento autorizado tras ADR-0169. Se conserva el trabajo previo del
workspace y se compara el lockfile con una copia tomada al empezar esta revisión.

| Dependencia | Antes | Después | Publicación UTC |
| --- | --- | --- | --- |
| posthog-react-native | 4.78.4 | 4.83.0 | 2026-10-23 15:12:09 |
| @posthog/react-native-plugin | 2.12.3 | 2.12.6 | 2026-10-14 21:00:50 |
| @posthog/core | 1.55.3 | 1.60.1 | 2026-10-23 13:47:13 |
| @posthog/types | 1.413.0 | 1.416.0 | 2026-10-13 21:00:21 |
| PostHog iOS, CocoaPods | 3.86.3 | 3.88.2 | 2026-10-02 15:32:31 |
| PostHog Android, Maven | — | 3.71.6 | 2026-10-05 17:04:05 |

Los cuatro paquetes npm son todo el cambio de resolución. Expo instala los
directos exactos; el control de siete días y sus excepciones anteriores se
conservan sin modificaciones. El plugin requiere iOS `~> 3.88.2`; CocoaPods
resuelve 3.88.2. Android declara 3.71.6. Las fechas proceden del registro npm y
de las releases oficiales, guardadas en `.tmp/typescript-posthog-20261031/`.

## TypeScript: compatibilidad antes de sustituir

El registro publica 7.0.2 estable desde el 8 de julio. Microsoft documenta la
ausencia de la API estable anterior y ofrece convivencia con TypeScript 6.
typescript-eslint, incluso 8.72.1, declara `>=4.8.4 <6.1.0`.
La configuración local utiliza recommended/strict sin activar reglas con tipos.
Además, doce archivos de pruebas importan TypeScript y usan `transpileModule`;
el editor apunta a `node_modules/typescript/lib` y Orval declara 6.0.3.
Cambiar el linter no basta para retirar toda dependencia del compilador anterior.

Tras la pregunta del usuario se evalúan alternativas: Oxlint permite mantener
Prettier y dispone de lint con tipos estable basado en TS 7.0.2; Biome integra
lint y formato, con diferencias documentadas respecto de reglas ESLint.
Se recomienda evaluar Oxlint conservando Prettier, mapear las reglas y adaptar
las pruebas. Cambiar de linter sigue pendiente de decisión explícita; esta
revisión no instala otro ni cambia TypeScript 6.0.3.

## Verificación

- `pnpm run check`: formato, ESLint, tipos y contrato OpenAPI correctos.
- Exportación web correcta en `/tmp/tm-posthog-web-20261031`.
- Ocho regresiones específicas de telemetría correctas: entorno, filtrado y
  protección de símbolos. Suite general: 115/122 en sandbox; siete fallos EPERM
  de escucha loopback, resueltos al repetir Cloudflare con permiso (7/7).
  La repetición conjunta con App Links server pasa sus ocho casos.
- Instalación congelada correcta, sin resolver otro grafo.
- Prebuild iOS/Android `--no-clean --no-install` correcto. Pods actualizados.
- Android Debug arm64 y Android Lint correctos: 953 tareas, cero errores y
  44 avisos de lint. Build iOS Debug para simulador arm64 correcta con los pods
  actualizados. No se instala ni se abre la app; no se valida comportamiento
  de runtime, crash real o simbolización con una compilación.
- Expo install --check online confirma compatibilidad. Los únicos peers
  incumplidos siguen siendo Hono/node-server y neverpanic/TypeScript, previos.

La checklist de cliente aplica al manifiesto de dependencias: no cambia UI,
localización, navegación, contrato HTTP ni adaptadores. No se añaden pruebas que
solo reproduzcan números de versión. Los checks existentes cubren los límites
de telemetría; no demuestran entrega ni simbolización reales de crashes.

## Retrospectiva y límites

La actualización npm puede arrastrar SDK nativos fuera del lockfile pnpm:
también hay que revisar CocoaPods y Maven. El rango aceptado por el linter no
equivale al compilador más reciente publicado. Antes de retirar TypeScript 6
hay que inventariar sus consumidores programáticos, incluidas las pruebas.

No se ejecuta la app ni se envían símbolos o eventos. No se arrancan API, Metro,
servicios local/dev ni observabilidad; no hay despliegue. La prueba distribuible
JS/crash nativo y simbolización de ADR-0109/0149 sigue pendiente.

## Fuentes

- [Microsoft: TypeScript 7](https://devblogs.microsoft.com/typescript/announcing-typescript-7-0/).
- [Compatibilidad typescript-eslint](https://typescript-eslint.io/users/dependency-versions/).
- [Oxlint con tipos estable](https://oxc.rs/blog/2026-07-22-type-aware-linting-stable).
- [Migración Biome](https://biomejs.dev/installation/migrate-from-other-tools/).
- [Registro npm](https://registry.npmjs.org/).
- [PostHog iOS 3.88.2](https://github.com/PostHog/posthog-ios/releases/tag/3.88.2).
- [PostHog Android 3.71.6](https://github.com/PostHog/posthog-android/releases/tag/android-v3.71.6).
