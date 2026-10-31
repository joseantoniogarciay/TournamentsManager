# Migración a Expo SDK 58 — 2026-10-31

La corrección posterior de ADR-0168 cierra los dos errores Android Lint que se
registran en esta revisión inicial. Véase [el informe de corrección](ANDROID_LINT_FIX_2026-10-31.md)
para el resultado actual: cero errores y los avisos previos conservados.

## Decisión y alcance

El usuario elige Expo 58.0.22 y confirma expresamente que la excepción puntual
a la espera de siete días incluye sus dependencias necesarias, Expo CLI, React
Native y herramientas nativas, incluido noxcturnal. ADR-0167 registra esa decisión
tras el rechazo inicial de la revisión automática. La política general conserva
minimumReleaseAge=10080, modo estricto, rechazo de fechas ausentes y lockfile
congelado. Las 65 exclusiones son versiones exactas presentes en el lockfile,
sin permisos para publicaciones futuras; sus fechas de retirada aparecen abajo.

Se conservan los cambios locales anteriores de ADR-0166. El importer raíz del
lockfile permanece idéntico al anterior a esta tarea. No quedan paquetes Expo
SDK 57 en la resolución final. No se publica ni despliega producción.

## Implementación

- Expo 58.0.22, Router 58.0.29, React 19.3.0, React Native y Metro Config 0.88.1.
  Las dependencias directas actualizadas quedan fijadas a versiones exactas.
- Worklets 0.13.0 satisface Reanimated 4.7.1; System UI 58.0.7 completa el soporte
  de apariencia configurado. Expo Status Bar se declara como plugin de SDK 58.
- Ref de View/ScrollView/TextInput con sus tipos de instancia; SectionList usa
  ComponentRef y la forma explícita de su sección. Sin opt-out de la API estricta.
- NativeTabs usa el entrypoint estable. Los stacks conservan la cabecera opaca
  con headerTransparent=false para preservar la geometría del contenido.
- La home exporta generateMetadata: descripción, canonical, Open Graph y Twitter
  existen en el HTML inicial. Head conserva las actualizaciones de idioma en
  cliente. El viewport con viewport-fit=cover vive en +html. Se conserva salida
  estática y el documento sigue incluyendo headNodes/bodyNodes para hidratación.
- La prueba real revela que iOS 27 ignora la API global de barra de estado cuando
  UIViewControllerBasedStatusBarAppearance=false. Se fija true en app.config,
  conservando el componente StatusBar raíz y el tema compartido. La build final
  se regenera, recompila e instala: Inicio es visible con iconos de estado
  contrastados y ya no aparece el aviso del runtime.
- SDK 58 genera SceneDelegate basado en ExpoAppSceneDelegate. Se retiran el
  plugin UIScene local, la prueba de su plantilla antigua y su target de Make.
  Se comprueba la plantilla y el manifiesto oficiales generados.
- JSI 58.0.21 incorpora la limpieza de ambas sintaxis Swift; se retira el parche
  57.1.1 y se mantiene la regresión sobre el script upstream con sed -E.
- Router abandona query-string: se retiran su override y parche sin consumidor;
  las regresiones de URLs ahora ejecutan el parser/decoder instalado de Router.
- El parche de Symbols se porta a 58.0.4: upstream todavía escala el glifo Material
  dentro de una caja fija. Solo ese parche permanece en patchedDependencies.

## Verificación ejecutada

| Comprobación | Resultado |
| --- | --- |
| Expo install --check | Dependencias compatibles |
| Expo Doctor 1.21.5 | 20/20 comprobaciones correctas |
| pnpm run typecheck | Correcto |
| ESLint | Sin errores ni avisos |
| Prettier | Todos los archivos coinciden con el formato |
| Regresiones JavaScript del cliente | 119 pruebas correctas |
| Seguridad operativa Python | 11 pruebas correctas |
| Instalación frozen-lockfile | Correcta, sin nueva resolución |
| Exportación web | Correcta; metadatos verificados dentro de head |
| Exportación Hermes iOS y Android | Correcta |
| Prebuild iOS y Android | Correcto, sin plugin UIScene local |
| Compilación Android Debug arm64 | Correcta: 523 tareas, 6 min 3 s |
| Compilación iOS Debug simulador | Correcta arm64/x86_64 (941,785 s); build final con ajuste de barra en arm64 (280,444 s) |
| Android Lint por defecto, arm64 | 2 errores previos (MissingPrefix/NewApi), 46 avisos; alcanza el informe |
| Prueba de arranque con Metro | Autorizada y ejecutada: Inicio en ambas apps; formulario y cierre en iOS |

La generación conserva UIApplicationSupportsMultipleScenes=false, cuatro
orientaciones en iPhone/iPad, UIRequiresFullScreen=false, familia 1,2 y el
componente StatusBar raíz; el control nativo cambia al modo por controlador. Android conserva fullUser,
resizeableActivity=true, adjustResize, configChanges y filtros de enlaces.
No cambia el mínimo de plataforma: iOS 16.4 y Android API 24 en ambas generaciones.

La inspección de los generadores de SDK 58 sigue encontrando data-generated en
intent-filter y windowSplashScreenBehavior en values/styles.xml sin values-v33.
El lint normal termina con esos dos errores y 46 avisos; el crash previo del
motor no se reproduce en esta invocación. No se certifica el cierre de ADR-0152
ni se añaden supresiones, baseline o parches de esos recursos. La compilación
Android no sustituye Android Lint ni la matriz visual. El informe queda en
apps/client/android/app/build/reports/lint-results-debug.xml; la invocación se
acota a arm64 para usar la arquitectura ya compilada.

Expo Doctor pasa con sus versiones fijadas y no requiere otro servicio. pnpm
peers check conserva los avisos ajenos al cliente de @hono/node-server para
@hono/node-ws y TypeScript para neverpanic; no se modifican herramientas ajenas
para resolverlos. El peer de Metro del cliente sí queda alineado a 0.88.1.

## Evidencia y límites

Los logs locales de esta sesión están en /tmp/tm-expo58-*.log; los proyectos
iOS y Android de SDK 57, y los parches retirados, se guardan completos en
/tmp/tm-expo58-before. Las salidas estáticas están en /tmp/tm-expo58-web-export-final
y /tmp/tm-expo58-native-export; la build iOS usa /tmp/tm-expo58-ios-build y el
APK está en apps/client/android/app/build/outputs/apk/debug/app-debug.apk.
Son evidencia local y artefactos ignorados, no nuevas fuentes versionadas.

La revisión automática pidió autorización explícita para Metro en LAN; el
usuario la concedió y pidió continuar. La sesión usa APP_ENV=local y una API
inalcanzable en loopback, sin arrancar backend/dev ni observabilidad. Ambas builds
se instalan conservando los datos de los dispositivos y se abren con el enlace
exacto devuelto por GET /_expo/open del CLI 58 (el formato cambió a __expo_url).

Pixel_API_31 muestra inicialmente un diálogo de bloqueo de System UI. Después de
reiniciar ese proceso desde su diálogo, Inicio queda visible con tema oscuro y
barra de estado legible; no se recoge una excepción ReactNativeJS/AndroidRuntime.
El iPhone Duo informa Data Migration Failed en bootstatus, pero permite instalar
y ejecutar la app. La captura por defecto elige la pantalla secundaria LCD-1 y
queda negra; la captura explícita de la pantalla principal 1 muestra Inicio.
No se trata esa captura negra como un fallo de la app. Se confirma el aviso del
sistema para abrir la app local; Crear torneo abre el formulario y su cierre
restaura Inicio. No se introducen datos ni se envían formularios.

Las capturas están en /tmp/tm-expo58-android-screen-final.png y
/tmp/tm-expo58-ios-statusbar-final.png; Device Hub confirma navegación y cierre. El log iOS conserva diagnósticos de UIKit del simulador (medición de lanzamiento,
PointerUI, caché de foco y la lectura de statusBarFrame deprecada de dependencias);
no se confunden con el aviso de configuración de StatusBar que sí se corrige.
La prueba es de arranque/navegación básica: no acredita tablet, Split View, todas las
orientaciones, autenticación, enlaces distribuidos ni operaciones de negocio.
La sesión temporal queda apagada: Metro sin listener en 8082, emulador Android
cerrado, iPhone Duo en Shutdown y daemon ADB detenido. Se conservan datos,
volúmenes, respaldos y evidencia; API local/dev y observabilidad no se arrancan.
No se autoriza publicación de producción.

## Exclusiones exactas y retirada

Fechas obtenidas del registro npm oficial. Retirar cada entrada cuando alcance
la última columna; una actualización posterior vuelve a aplicar ADR-0138.

| Versión | Publicación UTC | Cumple siete días UTC |
| --- | --- | --- |
| `expo-system-ui@58.0.7` | 2026-10-28T15:33:53.914Z | 2026-11-04T15:33:53.914000Z |
| `@react-native/babel-preset@0.88.1` | 2026-10-27T12:15:47.882Z | 2026-11-03T12:15:47.882000Z |
| `@react-native/metro-babel-transformer@0.88.1` | 2026-10-27T12:14:40.177Z | 2026-11-03T12:14:40.177000Z |
| `@expo/metro-runtime@58.0.16` | 2026-10-31T10:32:18.339Z | 2026-11-07T10:32:18.339000Z |
| `@react-native/community-cli-plugin@0.88.1` | 2026-10-27T12:14:18.331Z | 2026-11-03T12:14:18.331000Z |
| `@react-native/gradle-plugin@0.88.1` | 2026-10-27T12:14:13.036Z | 2026-11-03T12:14:13.036000Z |
| `@react-native/virtualized-lists@0.88.1` | 2026-10-27T12:14:36.327Z | 2026-11-03T12:14:36.327000Z |
| `expo-dev-launcher@58.0.18` | 2026-10-31T10:30:26.409Z | 2026-11-07T10:30:26.409000Z |
| `expo-dev-menu@58.0.12` | 2026-10-27T12:14:46.144Z | 2026-11-03T12:14:46.144000Z |
| `hermes-compiler@260318099.0.5` | 2026-10-26T14:55:45.115Z | 2026-11-02T14:55:45.115000Z |
| `react-native-screens@4.30.1` | 2026-10-28T13:21:33.097Z | 2026-11-04T13:21:33.097000Z |
| `@expo/cli@58.1.18` | 2026-10-31T10:33:41.402Z | 2026-11-07T10:33:41.402000Z |
| `@expo/config-plugins@58.0.11` | 2026-10-28T15:32:21.785Z | 2026-11-04T15:32:21.785000Z |
| `@expo/config@58.0.10` | 2026-10-31T10:17:59.027Z | 2026-11-07T10:17:59.027000Z |
| `@expo/fingerprint@0.21.8` | 2026-10-31T10:17:05.011Z | 2026-11-07T10:17:05.011000Z |
| `@expo/image-utils@0.12.6` | 2026-10-28T15:32:18.781Z | 2026-11-04T15:32:18.781000Z |
| `@expo/inline-modules@0.2.8` | 2026-10-28T15:33:57.662Z | 2026-11-04T15:33:57.662000Z |
| `@expo/local-build-cache-provider@58.0.10` | 2026-10-31T10:30:58.026Z | 2026-11-07T10:30:58.026000Z |
| `@expo/log-box@58.0.14` | 2026-10-31T10:20:18.846Z | 2026-11-07T10:20:18.846000Z |
| `@expo/metro-config@58.0.20` | 2026-10-31T10:31:38.814Z | 2026-11-07T10:31:38.814000Z |
| `@expo/modules-macros-linux-arm64@0.27.1` | 2026-10-28T11:19:51.059Z | 2026-11-04T11:19:51.059000Z |
| `@expo/modules-macros-linux-x64@0.27.1` | 2026-10-28T11:21:40.752Z | 2026-11-04T11:21:40.752000Z |
| `@expo/modules-macros-win32-arm64@0.27.1` | 2026-10-28T11:19:57.518Z | 2026-11-04T11:19:57.518000Z |
| `@expo/modules-macros-win32-x64@0.27.1` | 2026-10-28T11:18:22.498Z | 2026-11-04T11:18:22.498000Z |
| `@expo/prebuild-config@58.0.18` | 2026-10-31T10:31:06.251Z | 2026-11-07T10:31:06.251000Z |
| `@expo/require-utils@58.0.5` | 2026-10-28T15:29:47.745Z | 2026-11-04T15:29:47.745000Z |
| `@expo/router-server@58.0.20` | 2026-10-31T10:33:24.700Z | 2026-11-07T10:33:24.700000Z |
| `@expo/schema-utils@58.0.3` | 2026-10-31T10:16:20.010Z | 2026-11-07T10:16:20.010000Z |
| `@noxcturnal/noxcturnal-darwin-arm64@0.1.11` | 2026-10-27T19:11:34.729Z | 2026-11-03T19:11:34.729000Z |
| `@noxcturnal/noxcturnal-darwin-x64@0.1.11` | 2026-10-27T19:10:34.892Z | 2026-11-03T19:10:34.892000Z |
| `@noxcturnal/noxcturnal-linux-arm64-gnu@0.1.11` | 2026-10-27T19:11:04.371Z | 2026-11-03T19:11:04.371000Z |
| `@noxcturnal/noxcturnal-linux-arm64-musl@0.1.11` | 2026-10-27T19:10:33.595Z | 2026-11-03T19:10:33.595000Z |
| `@noxcturnal/noxcturnal-linux-x64-gnu@0.1.11` | 2026-10-27T19:11:04.952Z | 2026-11-03T19:11:04.952000Z |
| `@noxcturnal/noxcturnal-linux-x64-musl@0.1.11` | 2026-10-27T19:10:33.671Z | 2026-11-03T19:10:33.671000Z |
| `@noxcturnal/noxcturnal-win32-arm64-msvc@0.1.11` | 2026-10-27T19:11:34.903Z | 2026-11-03T19:11:34.903000Z |
| `@noxcturnal/noxcturnal-win32-x64-msvc@0.1.11` | 2026-10-27T19:14:35.597Z | 2026-11-03T19:14:35.597000Z |
| `@react-native/asset-utils@0.88.1` | 2026-10-27T12:11:31.892Z | 2026-11-03T12:11:31.892000Z |
| `@react-native/babel-plugin-codegen@0.88.1` | 2026-10-27T12:18:45.273Z | 2026-11-03T12:18:45.273000Z |
| `@react-native/codegen@0.88.1` | 2026-10-27T12:15:17.712Z | 2026-11-03T12:15:17.712000Z |
| `@react-native/debugger-frontend@0.88.1` | 2026-10-27T12:16:07.607Z | 2026-11-03T12:16:07.607000Z |
| `@react-native/debugger-shell@0.88.1` | 2026-10-27T12:18:25.627Z | 2026-11-03T12:18:25.627000Z |
| `@react-native/dev-middleware@0.88.1` | 2026-10-27T12:14:50.514Z | 2026-11-03T12:14:50.514000Z |
| `@react-native/js-polyfills@0.88.1` | 2026-10-27T12:14:37.600Z | 2026-11-03T12:14:37.600000Z |
| `@react-native/metro-config@0.88.1` | 2026-10-27T12:13:15.848Z | 2026-11-03T12:13:15.848000Z |
| `@react-native/normalize-colors@0.88.1` | 2026-10-27T12:14:54.695Z | 2026-11-03T12:14:54.695000Z |
| `babel-preset-expo@58.0.23` | 2026-10-31T10:32:22.913Z | 2026-11-07T10:32:22.913000Z |
| `expo-asset@58.0.19` | 2026-10-28T15:35:29.223Z | 2026-11-04T15:35:29.223000Z |
| `expo-auth-session@58.0.16` | 2026-10-28T15:34:54.611Z | 2026-11-04T15:34:54.611000Z |
| `expo-build-properties@58.0.12` | 2026-10-31T10:31:02.710Z | 2026-11-07T10:31:02.710000Z |
| `expo-constants@58.0.14` | 2026-10-28T15:35:30.588Z | 2026-11-04T15:35:30.588000Z |
| `expo-dev-client@58.0.18` | 2026-10-31T10:32:20.181Z | 2026-11-07T10:32:20.181000Z |
| `expo-file-system@58.0.13` | 2026-10-28T15:39:27.892Z | 2026-11-04T15:39:27.892000Z |
| `expo-font@58.0.11` | 2026-10-28T15:33:20.136Z | 2026-11-04T15:33:20.136000Z |
| `expo-linking@58.0.15` | 2026-10-28T15:34:41.340Z | 2026-11-04T15:34:41.340000Z |
| `expo-modules-autolinking@58.0.16` | 2026-10-28T15:33:25.829Z | 2026-11-04T15:33:25.829000Z |
| `expo-modules-cli@0.4.0` | 2026-10-28T15:31:34.614Z | 2026-11-04T15:31:34.614000Z |
| `expo-modules-core@58.0.29` | 2026-10-31T10:20:39.133Z | 2026-11-07T10:20:39.133000Z |
| `expo-modules-jsi@58.0.21` | 2026-10-28T15:31:58.715Z | 2026-11-04T15:31:58.715000Z |
| `expo-modules-macros@0.27.1` | 2026-10-28T11:19:38.066Z | 2026-11-04T11:19:38.066000Z |
| `expo-router@58.0.29` | 2026-10-31T10:33:33.883Z | 2026-11-07T10:33:33.883000Z |
| `expo-splash-screen@58.0.12` | 2026-10-28T15:34:56.357Z | 2026-11-04T15:34:56.357000Z |
| `expo-web-browser@58.0.7` | 2026-10-27T12:14:44.758Z | 2026-11-03T12:14:44.758000Z |
| `expo@58.0.22` | 2026-10-31T10:33:20.523Z | 2026-11-07T10:33:20.523000Z |
| `noxcturnal@0.1.11` | 2026-10-27T19:11:37.017Z | 2026-11-03T19:11:37.017000Z |
| `react-native@0.88.1` | 2026-10-27T14:43:35.156Z | 2026-11-03T14:43:35.156000Z |

## Retrospectiva técnica

Actualizar el paquete raíz no completa un SDK nativo: la matriz, peers, plugins,
parches, tipos y metadatos de exportación deben evolucionar juntos. Las pruebas
contra una implementación retirada dejan de proteger el producto; se retiran
junto con esa implementación y se contrasta la salida oficial que la sustituye.
La excepción puntual queda acotada a versiones y a una fecha verificable sin
alterar la protección global. Compilación, arranque y QA visual se registran
por separado; los resultados de SDK 57 no se atribuyen a SDK 58.

## Fuentes

- [Expo SDK 58](https://expo.dev/changelog/sdk-58).
- [Migración Router 57 a 58](https://docs.expo.dev/router/migrate/sdk-57-to-58/).
- [Registro Expo 58.0.22](https://registry.npmjs.org/expo/58.0.22).
- [Registro React Native 0.88.1](https://registry.npmjs.org/react-native/0.88.1).


## Cierre de la checklist cliente

Se revisan los ADR aplicables de tema, navegación y primitivas, además de CNG,
maduración y desarrollo bajo petición. La migración adapta compatibilidad y
conserva los textos en sus catálogos, tokens, arquitectura por feature y transporte
compartido. No modifica operaciones OpenAPI, reglas de negocio ni feedback HTTP.
Typecheck y exportación web están ejecutados; el cambio nativo de barra de estado
se valida en la build reinstalada, con prueba visual de Inicio. Los límites de
ADR-0152 y de la matriz completa de ADR-0166 se declaran arriba, sin certificar
QA completo por una prueba de arranque. Se actualizan ADR, índices, Desarrollo,
sistema de diseño, checklist cliente, Changelog y Aprendizaje.
