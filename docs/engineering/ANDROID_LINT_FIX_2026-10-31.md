# Corrección de Android Lint en Expo SDK 58 — 2026-10-31

## Resultado y decisión

Android Lint pasa con **cero errores** en la app y en el módulo local
global-feedback. Conserva 46 avisos en la app y 6 en el módulo. La compilación
Debug arm64 termina correctamente. No se añaden supresiones ni baseline, no se
cambia el motor ni minSdk y no se instalan versiones nuevas.

El usuario pide solucionar los errores que la revisión de SDK 58 dejó abiertos.
ADR-0168 sustituye la espera decidida en ADR-0152 por dos parches exactos de los
generadores ya fijados. Se preserva todo el trabajo previo de SDK 58 y ventanas.

## Correcciones reproducibles

### Propiedad de filtros de App Links

El marcador data-generated es útil dentro del objeto de configuración de Expo,
pero no es un atributo válido del manifest. El lector/escritor de Manifest lo
representa en disco con el comentario reservado @generated expo-intent-filter.
Al leer, se recupera el indicador interno antes de que xml2js descarte comentarios.
Al escribir, se convierte el indicador en comentario. El setter de Expo conserva
su lógica de propiedad y se corrige además el alta cuando aún no existe un array
de filtros. Los manifiestos antiguos con data-generated se migran al escribir.

Se comprueba conservación de filtros manuales, autoVerify, acciones, categorías,
host y path; cambiar el dominio sustituye solo el filtro propio y retirar la
configuración elimina solo el generado. Leer/escribir y regenerar no duplican
marcas ni filtros; también se prueba la lectura de un manifest con CRLF.

La propuesta inicial tools:expoGenerated falla en el merger: tools no admite
metadatos arbitrarios, sino instrucciones conocidas. Se descarta esa propuesta;
no se usa una instrucción de merge o de supresión como identificador. El comentario
XML no afecta al merger y el manifest compilado no conserva la marca.

### Recursos de splash por nivel API

El estilo base Theme.App.SplashScreen conserva windowSplashScreenBackground,
windowSplashScreenAnimatedIcon y postSplashScreenTheme. El generador escribe el
estilo completo en values-v33/styles.xml y solo allí añade
android:windowSplashScreenBehavior=icon_preferred. Se comparten las referencias
al logo y colores; los overlays de color claro/oscuro siguen resolviéndose.

La variante se escribe mediante un mod Android del propio generador de splash,
con lectura/actualización XML que conserva otros estilos del archivo. No depende
de leer el estilo base antes de que se haya generado. Se prueban escritura
repetida, conservación de otros estilos y colores de ambos temas. No se edita
ningún recurso generado manualmente.

## Parches y mantenimiento

- @expo/config-plugins 58.0.11: patches/@expo__config-plugins@58.0.11.patch.
- expo-splash-screen 58.0.12: patches/expo-splash-screen@58.0.12.patch; fuente TS
  y JavaScript distribuido reciben el mismo cambio.
- pnpm registra los hashes junto al parche de Symbols ya existente. La instalación
  congelada pasa; las versiones directas y la política general no cambian.

Cada actualización de esos dos paquetes debe revisar y retirar su parche cuando
upstream cubra los casos y las regresiones pasen. No se añade una librería ni un
plugin local que duplique la generación de Expo.

## Verificación

| Comprobación | Resultado |
| --- | --- |
| Tres regresiones nuevas de generadores | Correctas |
| Suite completa del cliente | 122 pruebas correctas |
| Typecheck y ESLint | Correctos |
| Formato y git diff --check | Correctos |
| Prebuild Android y segunda pasada --no-clean | Manifest, estilos base y v33 idénticos byte a byte |
| assembleDebug, lintDebug de app y global-feedback | Correctos: 956 tareas; 2 min 4 s |
| Lint app | 0 errores, 46 avisos |
| Lint global-feedback | 0 errores, 6 avisos |
| Manifest compilado del APK | Sin data-generated, expoGenerated ni comentario de propiedad; autoVerify y hosts conservados |
| Recursos compilados del APK | Estilo base con tres entradas; variante v33 con cuatro; mismos logo/colores/post-theme |
| Mínimo del APK | API 24 |
| Exportación web | Correcta |
| Instalación frozen-lockfile | Correcta |

La comprobación del APK usa aapt2 dump xmltree y dump resources del SDK Android
instalado. No se deduce una prueba visual en API 24/34 de inspeccionar recursos:
esa matriz de dispositivos, splash release y enlaces distribuidos queda como QA
adicional de ADR-0152. La corrección y el gate de lint sí quedan comprobados.
No se cambia código de UI, operaciones OpenAPI ni lógica de negocio. iOS conserva
la build y el arranque validados en ADR-0167; los cambios de generadores son Android.

Los informes locales están en apps/client/android/app/build/reports y
apps/client/modules/global-feedback/android/build/reports. Logs, dumps del APK,
hashes de generación y exportación web se conservan en /tmp/tm-android-lint-fix.
El directorio Android anterior se conserva allí como respaldo. SDK 58 limpia por
defecto al ejecutar prebuild; para comprobar idempotencia sin destruir la caché
se usa explícitamente --no-clean. API local/dev, Metro y observabilidad no se
arrancan en esta corrección; no se publica producción.

## Retrospectiva técnica

Un namespace válido en XML no implica que el merger admita cualquier atributo
en él. La primera propuesta pasa pruebas del objeto, pero falla al compilar; por
eso se verifican serialización, regeneración real y APK además de los helpers.
El identificador de propiedad pertenece a la generación y se expresa como un
comentario que no llega a runtime. Los recursos por API preservan el diseño
sin ocultar una incompatibilidad ni elevar la plataforma mínima.

## Fuentes

- [ADR-0168](../adr/0168-fix-android-lint-at-expo-generators.md).
- [Atributos tools de Android](https://developer.android.com/studio/write/tool-attributes).
- [Recursos alternativos Android](https://developer.android.com/guide/topics/resources/providing-resources).
