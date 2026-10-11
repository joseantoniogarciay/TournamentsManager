# Diagnóstico de Android Lint — 2026-10-09

Estado: contraste de herramientas; no modifica el gate ni selecciona otro motor.
Fuente del cliente: develop 8f836c02dcba27f99e3e6c73811261aa95f0bc26.

## Problema y evidencia

El lint completo no llega al informe de la app: Worklets aborta en K2 UAST con
Cannot find a KaModule for the VirtualFile. El catálogo instalado de React
Native 0.86.2 declara AGP 8.12.0 y Kotlin 2.1.20. No inferir la versión de AGP
por las versiones de jars que haya en la caché de Android Studio.

Worklets 0.10.3 aplica fix-prefab.gradle.kts y generate-stub-pch.gradle.kts desde
su build.gradle.kts. La propia dependencia desactiva lintVital y cita un fallo
del analizador; eso no permite afirmar que nuestro lint completo haya pasado.
No se añade ninguna desactivación, baseline o supresión al proyecto.

## Contrastes reproducibles

Todos los comandos usan el JDK de Android Studio y --no-daemon --max-workers=2.
No se arranca API, PostgreSQL, Metro, emulador Android ni observabilidad.

1. JAVA_TOOL_OPTIONS=-Dlint.use.fir.uast=false no elimina el fallo KaModule.
   La traza conserva K2: no demuestra haber cambiado el motor del worker.
2. :react-native-worklets:lintAnalyzeDebug con
   -Pandroid.experimental.lint.version=9.0.1 y --rerun-tasks reemplaza el fallo
   por findFirCompiledSymbol only works on compiled declarations. Exit 1.
   Es un contraste por línea de comandos, sin actualizar AGP ni persistir
   configuración. No se adopta lint 9.0.1 como corrección.
3. La propiedad de AGP -Pandroid.lint.useK2Uast=false sí permite completar
   :react-native-worklets:lintAnalyzeDebug --rerun-tasks: exit 0, 48 tareas
   ejecutadas. No equivale a validar K2 ni a una actualización del producto.
4. Se ejecuta además :app:lintDebug :global-feedback:lintDebug con esa misma
   propiedad temporal para obtener un inventario complementario. Su resultado
   se registra en el informe de QA al finalizar.

Intentar dependencies --configuration lintClassPath no es un inventario válido:
Gradle informa que esa configuración no existe. No usar esa ejecución fallida
para atribuir una versión del motor a un módulo.

## Alternativas, coste y recomendación

- Conservar K2 y esperar una versión publicada que resuelva el caso: mantiene
  el criterio vigente; coste bajo de configuración, pero el gate sigue abierto.
- Actualizar solo lint mediante la propiedad soportada por Google: menos coste
  que migrar AGP/Kotlin; requiere una versión corregida, fecha de publicación,
  siete días de maduración y validación de toda la cadena. El ensayo 9.0.1 falla.
- Seleccionar K1 permanentemente: evita este crash en el contraste de Worklets,
  pero cambia el motor y sus capacidades. Un informe K1 no acredita K2. No se
  adopta silenciosamente como forma de cerrar el gate.
- Migrar AGP completo o parchear Worklets: mayor alcance y mantenimiento sin
  evidencia de que sea necesario; no es la solución mínima para este diagnóstico.

Recomendación: conservar el gate abierto y utilizar K1 como evidencia adicional,
registrando las incidencias que realmente alcance. No instalar una corrección
sin publicación identificada ni saltar ADR-0138 por compatibilidad.

## Fuentes primarias y límites

- [Google: ejecutar lint más reciente sin migrar AGP](https://googlesamples.github.io/android-custom-lint-rules/usage/newer-lint.md.html).
- [Google: AGP 9.0, incidencias corregidas](https://developer.android.com/build/releases/agp-9-0-0-release-notes).
- [Google: corrección adicional de scripts incluidos, commit del 7 de octubre](https://android.googlesource.com/platform/tools/base/+/44330ad8b855e38930c70ca44f65e781dd95d1ab).

El último cambio explica una clasificación incorrecta de scripts externos por
K2 y retira una exclusión temporal. Es contexto coherente con nuestra traza,
no una prueba de que una versión publicada disponible cierre nuestro caso.
No confundir fecha de autor, integración y publicación de un artefacto Maven.

## Evidencia privada y retrospectiva

Logs android-worklets-lint-{k1-diagnostic,901-diagnostic,k1-gradle}-20261009.log,
android-worklets-lint-classpath-20261009.log y android-full-lint-k1-20261009.log,
en /private/tmp/tm-product-qa-20261004. No versionar trazas completas ni caches.

Una opción de JVM no prueba qué motor eligió un worker. Comparar la traza y el
resultado de la propiedad real de AGP distingue un contraste efectivo de un
intento sin efecto. Una corrección de la primera excepción tampoco implica
resolver toda la cadena: contrastar siempre el informe final y sus límites.
