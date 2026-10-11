# Revisión de parches Expo — 2026-10-11

## Resultado

Hay cinco parches directos publicados y recomendados hoy por `expo install --check`.
Siguen siendo demasiado jóvenes para instalarlos conforme a ADR-0138. No se
modifican package.json, lockfile, overrides, parches ni proyectos nativos.

El conjunto instalado 57.0.26 ya es la matriz madura adoptada el 9 de octubre.
Esta revisión trata solo la siguiente revisión del SDK 57, sin migrar de SDK/RN.

## Fechas verificadas en el registro oficial npm

| Paquete | Instalado | Candidato | Publicación UTC | Cumple siete días en Madrid |
| --- | --- | --- | --- | --- |
| expo | 57.0.26 | 57.0.27 | 2026-10-06T12:10:58.345Z | 13/10/2026 14:10:58 |
| expo-auth-session | 57.0.13 | 57.0.14 | 2026-10-06T12:09:47.851Z | 13/10/2026 14:09:47 |
| expo-constants | 57.0.20 | 57.0.21 | 2026-10-06T12:10:52.271Z | 13/10/2026 14:10:52 |
| expo-linking | 57.0.11 | 57.0.12 | 2026-10-06T12:09:57.269Z | 13/10/2026 14:09:57 |
| expo-router | 57.0.24 | 57.0.25 | 2026-10-06T12:10:55.270Z | 13/10/2026 14:10:55 |

También están publicados el 6 de octubre los mínimos requeridos de CLI 57.0.28,
asset 57.0.19, config 57.0.10, modules-core 57.0.21, metro-config 57.0.13,
config-plugins 57.0.10, autolinking 57.0.14 y babel-preset-expo 57.0.14.
El último mínimo inspeccionado es config-plugins: cumple siete días el
13 de octubre a las 14:14:47 Europe/Madrid. Por tanto, antes de las 14:15
ni siquiera ese conjunto cumple la política. Esta fecha es un mínimo de
revisión, no una certificación del lockfile completo: los rangos pueden resolver
otras revisiones y hay dependencias transitivas más profundas.

La matriz de SDK 57 mantiene React Native 0.86 y React 19.2.3. Los peers de Router
57.0.25 exigen Linking ^57.0.12 y Constants ^57.0.21: conviene actualizar los
cinco parches de forma coordinada, no actualizar Router aislado.

## Generadores Android pendientes

Se leen sin ejecutar los tarballs oficiales de config-plugins 57.0.10 y
splash-screen 57.0.9. IntentFilters.js conserva `data-generated` y el generador
splash conserva `android:windowSplashScreenBehavior` en el estilo base. No hay
corrección acreditada de MissingPrefix/NewApi. ADR-0152 permanece vigente; no
se atribuye una solución al crash K2 ni se afirma haber repetido lint nativo.

## Alternativas y recomendación

Mantener la matriz madura cuesta poco y conserva la QA actual; deja el aviso CLI.
Actualizar ahora exigiría saltar la política aceptada y no corrige los generadores
Android inspeccionados. La recomendación es conservarla y revisar a partir del
13 de octubre, 14:15 Madrid: resolver todo el grafo con el control de antigüedad,
revisar peers/parches/overrides, fijar directos exactos, builds nativas limpias y
QA iOS/Android además de typecheck, web y gate común. No se necesita cambiar SDK.

ADR-0138 establece: «Compatibilidad, novedades y avisos no críticos no permiten
saltar esa espera». Esta consulta no identifica ni acredita una excepción crítica.
El usuario pidió revisar si eran instalables; no se interpreta como autorización
para saltar la política ni instalar versiones jóvenes.

## Fuentes y evidencia

- [Registro oficial de Expo](https://registry.npmjs.org/expo): versiones y fechas.
- [Auth Session](https://registry.npmjs.org/expo-auth-session),
  [Constants](https://registry.npmjs.org/expo-constants),
  [Linking](https://registry.npmjs.org/expo-linking),
  [Router](https://registry.npmjs.org/expo-router).
- [Config plugins](https://registry.npmjs.org/@expo%2fconfig-plugins) y
  [Modules core](https://registry.npmjs.org/expo-modules-core).
- [SDK oficial 57](https://expo.dev/sdk/57) y
  [procedimiento oficial de actualización](https://docs.expo.dev/workflow/upgrading-expo-sdk-walkthrough/).
- [ADR-0138](../adr/0138-wait-seven-days-except-critical-vulnerabilities.md),
  [ADR-0152](../adr/0152-correct-expo-generated-android-lint-errors.md).

Metadatos completos y fuentes seleccionadas en `/tmp/tm-expo-update-*20261011.json`
y `/tmp/tm-expo-*-IntentFilters.js` / `/tmp/tm-expo-*-withAndroidSplashStyles.js`.
`expo install --check` termina exit 1 con los cinco candidatos indicados.
No se instalan dependencias, se ejecutan tarballs ni se arrancan servicios.

Retrospectiva: una matriz recomendada por Expo puede ser publicada pero todavía
no madura para el proyecto. Revisar la fecha de sus generadores evita confundir
la madurez directa con la de herramientas que prebuild necesita.
