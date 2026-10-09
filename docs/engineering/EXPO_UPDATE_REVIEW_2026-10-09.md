# Revisión de actualizaciones Expo / React Native — 2026-10-09

## Resultado y alcance

**Hechos:** hay 17 dependencias directas desalineadas con la matriz más reciente
que devuelve `expo install --check` (exit 1). La matriz publicada dentro de
Expo 57.0.26 ofrece candidatos directos con más de siete días de antigüedad.
React Native 0.86.3 pertenece a esa matriz; 0.87.1 es el dist-tag latest de RN,
pero no es la línea objetivo del SDK 57. No se instala ni modifica ninguna
dependencia, parche, override o lockfile en esta revisión.

**Recomendación:** evaluar primero la matriz madura del SDK 57, con Expo
57.0.26 y React Native 0.86.3, en una fase de actualización explícita.
No ejecutar `--fix` sin revisar las versiones: hoy solicita Expo 57.0.27 y
cuatro paquetes directos publicados el 6 de octubre, que no cumplen ADR-0138.
Expo 57.0.27 cumple siete días el 13 de octubre a las 12:10:58 UTC (14:10:58
Europe/Madrid); eso no acredita por sí solo la madurez de todo su grafo.

La decisión aceptada de ADR-0152 sigue siendo esperar las correcciones oficiales
de los generadores Android. La actualización aquí identificada no cierra esos
errores. El usuario pidió revisar updates; esta revisión no decide instalar
una nueva matriz ni cambiar de SDK.

## Inventario directo

Fuente: registro oficial npm consultado el 9 de octubre y
`bundledNativeModules.json` del tarball exacto de Expo 57.0.26. La columna
CLI corresponde a la expectativa actual, no a una recomendación automática.
Las fechas son UTC. Todas las versiones de la columna candidata cumplen la
antigüedad directa de ADR-0138 en el momento de la consulta.

| Paquete | Instalado | Esperado por CLI | Candidato SDK 57 maduro | Publicado |
| --- | --- | --- | --- | --- |
| expo | 57.0.15 | 57.0.27 | 57.0.26 | 2026-09-29 |
| expo-auth-session | 57.0.8 | 57.0.14 | 57.0.13 | 2026-09-24 |
| expo-blur | 57.0.2 | 57.0.3 | 57.0.3 | 2026-09-11 |
| expo-build-properties | 57.0.12 | 57.0.22 | 57.0.22 | 2026-09-24 |
| expo-constants | 57.0.13 | 57.0.21 | 57.0.20 | 2026-09-29 |
| expo-crypto | 57.0.1 | 57.0.3 | 57.0.3 | 2026-09-11 |
| expo-dev-client | 57.0.14 | 57.0.19 | 57.0.19 | 2026-09-11 |
| expo-font | 57.0.1 | 57.0.4 | 57.0.4 | 2026-09-11 |
| expo-linear-gradient | 57.0.1 | 57.0.2 | 57.0.2 | 2026-09-11 |
| expo-linking | 57.0.7 | 57.0.12 | 57.0.11 | 2026-09-24 |
| expo-localization | 57.0.1 | 57.0.2 | 57.0.2 | 2026-09-11 |
| expo-router | 57.0.15 | 57.0.25 | 57.0.24 | 2026-09-29 |
| expo-secure-store | 57.0.1 | 57.0.4 | 57.0.4 | 2026-09-11 |
| expo-splash-screen | 57.0.7 | 57.0.9 | 57.0.9 | 2026-09-11 |
| expo-symbols | 57.0.2 | 57.0.3 | 57.0.3 | 2026-09-11 |
| expo-web-browser | 57.0.2 | 57.0.3 | 57.0.3 | 2026-09-11 |
| react-native | 0.86.2 | 0.86.3 | 0.86.3 | 2026-08-24 |

Expo 57.0.27, auth-session 57.0.14, constants 57.0.21, linking 57.0.12 y
router 57.0.25 aún son jóvenes. No confundir SDK 58 / canary ni Worklets 0.13
con una actualización compatible por el mero hecho de estar publicados.
La matriz 57.0.26 declara Worklets 0.10.1; el árbol actual contiene 0.10.3
con Reanimated 4.5.1. No se propone un downgrade automático.

## Qué aporta y qué no resuelve

- El changelog oficial identifica la corrección del arranque lento de desarrollo
  en Expo 57.0.17 / RN 0.86.3. No afecta al arranque de producción. La corrección
  de memoria Hermes V1 ya está en Expo 57.0.9 / RN 0.86.2, por tanto el conjunto
  instalado ya la incluye; no se atribuye un ahorro nuevo sin medirlo.
- Expo 57.0.23 introduce soporte oficial opt-in de escenas iOS mediante
  `ios.enableSceneSupport` de expo-build-properties. La candidata podría
  permitir retirar `with-ios-scene-lifecycle.cjs`; requiere contrastar generación,
  arranque iOS 27, foreground, deep links y OAuth antes de retirar el adaptador.
- Inspección sin ejecución de tarballs oficiales: config-plugins 57.0.9 conserva
  `data-generated` en IntentFilters.js. También lo conserva 57.0.10, todavía
  joven. Splash-screen 57.0.9 conserva `windowSplashScreenBehavior` en el estilo
  base. No hay evidencia de corrección de MissingPrefix y NewApi en estos
  candidatos. No se afirma que se haya ejecutado lint sobre una nueva matriz.
- RN 0.86.3 mantiene AGP 8.12.0 en su catálogo Gradle. Esta revisión no demuestra
  solución del crash K2. El contraste K1 sigue siendo diagnóstico complementario.
- Symbols 57.0.3 conserva el Text sin desactivar font scaling en la fuente
  inspeccionada. El parche actual de 57.0.2 no se retira automáticamente:
  requiere adaptar su versión, comprobar aplicación y repetir regresión nativa.

## Compatibilidad y coste de mantenimiento

La matriz madura requiere revisar overrides actuales de `@expo/log-box` 57.0.1
y `@expo/metro-runtime` 57.0.4 frente a los rangos del candidato. Expo 57.0.26
requiere log-box ^57.0.4 y Router 57.0.24 metro-runtime ^57.0.16; ambas versiones
mínimas están maduras (26 de agosto y 18 de septiembre respectivamente).
Expo modules-core 57.0.20 está publicado desde el 29 de septiembre.

No se ha resuelto un lockfile candidato completo ni compilado esa matriz.
Los rangos transitivos pueden seleccionar revisiones posteriores: la edad de
los directos no sustituye el control de todas las dependencias y herramientas.
Conservar `minimumReleaseAge: 10080`, modo estricto, exclusiones vacías y
lockfile congelado después de la resolución revisada; no añadir excepciones.

| Alternativa | Beneficio | Coste / límite |
| --- | --- | --- |
| Mantener conjunto actual | Sin regresión de upgrade; respeta espera de ADR-0152 | Conserva advertencia de matriz y problemas Android abiertos |
| Actualizar matriz madura SDK 57 | Corrige arranque dev; permite evaluar escenas oficiales | Resolución transitiva, revisión de overrides y parches, generación y QA nativos; no cierra por sí sola lint |
| Esperar matriz 57.0.27 madura | Accede a revisiones más recientes | Espera adicional y mismo proceso de validación; generador manifest inspeccionado sigue sin corregir |
| Migrar SDK / RN minor | Puede incorporar otros arreglos | Mayor superficie y mantenimiento; falta análisis específico y decisión, innecesario para esta revisión |

La solución mínima recomendada es una actualización coordinada de parches del
SDK actual cuando se decida implementarla. Evitar una migración mayor solo
para hacer desaparecer una advertencia de CLI.

## Validación requerida antes de integrar una actualización

1. Resolver versiones directas y transitivas maduras; revisar lockfile, peers,
   overrides de seguridad y parches existentes. Ejecutar instalación congelada.
2. Repetir Expo check/Doctor con herramientas maduras, typecheck y gates comunes.
3. Generar nativos limpios mediante CNG, revisar manifest y estilos y ejecutar
   lint completo. Registrar fallos K2 y avisos sin supresiones ni cambio de gate.
4. Builds Android/iOS, arranque y navegación, splash, deep links/OAuth sin
   mutaciones no autorizadas; regresión de títulos, símbolos y feedback local.
5. Retirar adaptadores solo tras equivalencia demostrada; actualizar documentos
   y aprendizaje. Mantener servicios apagados al cerrar la sesión.

## Evidencia y retrospectiva

Evidencia privada en `/private/tmp/tm-product-qa-20261004/`:
`expo-updates-check-20261009.log`, registros de fechas Expo/RN,
`expo-update-maturity-20261009.json` y fuentes seleccionadas de tarballs en
`expo-mature-candidates-20261009/`. No se ejecutan scripts de esos paquetes.

Retrospectiva: latest, compatibilidad, antigüedad y corrección real son cuatro
comprobaciones distintas. Consultar la matriz incluida en el paquete exacto y
leer los generadores evita prometer que cualquier upgrade cerrará el QA.
La inspección estática orienta la decisión; la certificación exige ejecutar
la nueva matriz y no se atribuye a esta revisión.

## Fuentes primarias

- [Expo SDK 57: regresiones y escenas iOS](https://expo.dev/changelog/sdk-57).
- [Actualización de SDK Expo](https://docs.expo.dev/workflow/upgrading-expo-sdk-walkthrough/).
- [React Native 0.86: upgrading](https://reactnative.dev/docs/0.86/upgrading).
- [Metadatos oficiales Expo](https://registry.npmjs.org/expo).
- [Metadatos oficiales React Native](https://registry.npmjs.org/react-native).
- [Metadatos config-plugins](https://registry.npmjs.org/@expo/config-plugins).
- [Metadatos splash-screen](https://registry.npmjs.org/expo-splash-screen).
- ADR-0138, ADR-0152 y [diagnóstico lint](ANDROID_LINT_ENGINE_ANALYSIS_2026-10-09.md).


## Aplicación autorizada después de la revisión

El usuario solicita «Deja instalado todo lo actualizable». Se aplica la matriz
madura SDK 57 de la tabla mediante Expo CLI, manteniendo siete días y cero
exclusiones. Los directos Expo quedan fijados; se resuelven transitivas y
actualizaciones dentro de los rangos compatibles existentes con pnpm update.
No se migra SDK, React, la línea RN ni herramientas a otro major.

Se alinean log-box 57.0.4 y metro-runtime 57.0.16. Se declara Metro Config
0.86.3 como herramienta del cliente para satisfacer el peer exacto de RN.
Symbols conserva el mismo parche de glifo para 57.0.3, comprobado en fuente
y build instaladas. Metro 0.84.5/0.84.6 incorporan lector propio, no dependen
de image-size; se retiran el parche 0.84.4 y su override específico. Las pruebas
existentes de imágenes válidas, escaladas, vacías y truncadas pasan.

También se actualizan, dentro de sus rangos, PostHog JS 4.78.4, su plugin
2.12.3, Screens 4.26.2, React Native Web 0.21.3 y tipos React 19.2.18.
El resolver prefiere plugin 2.12.3 por el manifiesto y conjunto instalado;
no se fuerza 2.12.4 con un override adicional. No se activa telemetría ni se
cambian sus reglas por actualizar sus dependencias.

Expo CLI instala los paquetes, pero devuelve exit 1 al no poder escribir
plugins automáticamente en app.config.ts. Se añaden explícitamente expo-font
y expo-localization a esa configuración; CNG limpio termina exit 0 en iOS y
Android. La advertencia de expo-system-ui ya existente se conserva inventariada;
no se incorpora otra dependencia durante este mantenimiento.

Instalación congelada, typecheck, exportación web, tests de compatibilidad y
make verify terminan exit 0. Expo check conserva únicamente los cinco directos
jóvenes de la revisión; no se oculta con excludes. Peers nativos quedan
alineados; persisten dos avisos previos de tooling (Hono Node Server y
TypeScript/neverpanic), sin degradar las correcciones de seguridad para ocultarlos.

CocoaPods se resuelve después de actualizar su índice: PostHog iOS 3.86.3
(publicado 1 de octubre, 16:37:26 UTC) cumple la espera; Android usa
posthog-android 3.71.4 (Maven Central Last-Modified: 30 de septiembre,
11:31:39 UTC). El control pnpm no se atribuye a gestores nativos: sus fechas
se contrastan en las fuentes oficiales.

Fuentes adicionales: [release PostHog iOS 3.86.3](https://github.com/PostHog/posthog-ios/releases/tag/3.86.3)
y [POM PostHog Android 3.71.4](https://repo1.maven.org/maven2/com/posthog/posthog-android/3.71.4/posthog-android-3.71.4.pom).
Los resultados nativos y de arranque se registran abajo.


La comprobación offline de la matriz incluida termina exit 0, con el aviso
propio de Expo de que la validación offline es menos fiable. Complementa la
comprobación online, no acredita que las revisiones remotas jóvenes estén
instaladas. `pnpm audit` conserva dos avisos altos y cero críticos: node-forge
(requiere >=1.4.1) y braces (>=3.0.4). El registro oficial sigue sin publicar
esas correcciones al consultar el 9 de octubre; no se modifica la política
ni se declara que la auditoría haya pasado sin avisos.


### Validación Android de la instalación

La primera compilación nativa se interrumpe por disco lleno. Se conservan logs,
capturas y el último binario válido iOS, y se elimina exclusivamente su antiguo
DerivedData de QA. Se recuperan unos cinco GiB; la recompilación Android para
la arquitectura ARM64 del emulador termina BUILD SUCCESSFUL (55 s).
El APK Debug se instala con `adb install -r`, sin borrar los datos.

Metro conserva inicialmente la ruta física de Router 57.0.15 en su caché.
Se reinicia con `--clear` y registra el bundle de Router 57.0.24. El emulador
muestra Home, abre Crear torneo desde su botón y regresa a Home al cerrar.
Fuentes, iconos y campos visibles quedan en capturas privadas; no se envían
formularios ni se cambian las fixtures. Se apaga el emulador iniciado para QA.

Lint habitual conserva el crash K2 en Worklets, `Cannot find a KaModule for the
VirtualFile`. El contraste temporal K1 alcanza el XML de app: dos errores
(MissingPrefix y NewApi), 45 avisos. Global-feedback se completa por separado
con exit 0. No se cambia el motor permanente ni se aplican los parches Android
rechazados. La instalación nativa funcional no equivale a cerrar ese gate.

Evidencias: expo-update-{android-build,android-lint,android-lint-k1,feedback-lint-k1}
-20261009.log, informes XML copiados y expo-update-android-{home,create}
-20261009.png. Los fallos iniciales de disco se conservan como *-disk-full.log.
El último binario anterior queda en ios-last-good-release-20261008.app.


### Validación iOS y cierre

Xcode Debug ARM64 termina BUILD SUCCEEDED, exit 0 (740 s tras recuperar
espacio). Se instala la nueva .app en iPhone 17 / iOS 27, conservando los datos.
Tras el desbloqueo del Mac confirmado por el usuario, CUA abre el launcher y
cierra su explicación inicial y Dev Menu. Metro registra Router 57.0.24 para
el bundle iOS. Home muestra fuentes e iconos y abre Crear torneo desde su botón.

Text Size 3 conserva el título en la barra. Cambiar a 11 con el formulario
abierto reubica correctamente el título, pero produce recortes en otros textos
con alturas antiguas. Tras relanzar la app y entrar por el esquema local,
los bloques y chips recuperan su reflow: título completo debajo de la barra,
Cerrar separado y texto de deportes visible. Se verifica Cerrar -> Home.
Este contraste no determina si el relayout en caliente procede del baseline,
del modo Debug o del upgrade; se registra para comparación, sin dar por cerrada
la accesibilidad. Home a 11 conserva el gran título partido ya inventariado.

Capturas privadas: expo-update-ios-home-20261009.png,
expo-update-ios-create-size{3,11}-20261009.png,
expo-update-ios-create-size11-cold-20261009.png,
expo-update-ios-close-size11-20261009.png y expo-update-ios-restored-20261009.png.
El árbol accesible de la app sigue ausente en Device Hub; estas capturas no
acreditan VoiceOver ni orden de foco. No se envían formularios, no se inicia
sesión y no se alteran credenciales o fixtures.

Se restaura Text Size 3 y español; VoiceOver y Capture Keyboard permanecen off.
La app iOS se termina y Metro se detiene. El emulador Android creado para esta
sesión está apagado; el simulador iOS preexistente se conserva. API, PostgreSQL,
observabilidad y producción no se arrancan. Los adaptadores de escenas vigentes
se mantienen; no se atribuye al upgrade su retirada ni la validación OAuth.

Checklist cliente contra los archivos cambiados: package.json, app.config.ts
y parche Symbols conservan las decisiones de arquitectura, tema, idioma,
feedback y navegación aceptadas. Los plugins añadidos permiten CNG, no añaden
flujos de producto. No se toca operación OpenAPI ni se crean fetch/DTOs, textos
visibles, controles o reglas de negocio. Se completan typecheck, web,
compatibilidad de dependencias y make verify; los gates nativos abiertos se
identifican expresamente. La integración PostgreSQL completa se valida en CI,
no se atribuye a una prueba local sin su URL.

Retrospectiva de cierre: retirar el parche ya sustituido por upstream reduce
mantenimiento; conservar Symbols evita reintroducir el defecto. Un caché de
Metro y una falta de espacio pueden invalidar un contraste del upgrade sin
ser defectos de producto. Medir con código nuevo y separar arranque frío del
cambio de escala en caliente deja evidencia reproducible y límites claros.
