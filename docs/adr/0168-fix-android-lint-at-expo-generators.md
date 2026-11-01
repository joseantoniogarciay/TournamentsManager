# ADR-0168: Corregir Android Lint en los generadores fijados de Expo

- **Estado:** Aceptado
- **Fecha:** 2026-10-31
- **Decisor:** Usuario mediante petición de resolver los errores ahora
- **Supera a:** ADR-0152 en la decisión de esperar correcciones oficiales.
- **Complementa:** ADR-0015 y ADR-0167.

## Problema y evidencia

SDK 58 compila y arranca, pero Android Lint informa MissingPrefix por la marca
sin namespace data-generated de los filtros de enlaces y NewApi por
android:windowSplashScreenBehavior en un estilo aplicable desde API 24.
Los generadores fijados son @expo/config-plugins 58.0.11 y expo-splash-screen
58.0.12. El motor lint ya alcanza el informe sin el bloqueo anterior.

## Alternativas y coste

Esperar upstream conserva dos errores y evita mantener parches. Era la decisión
de ADR-0152. Dos parches exactos corrigen su origen y conservan CNG, con el coste
de revisar y retirar cada parche en upgrades. Un plugin adicional después de
los generadores duplicaría sus reglas de propiedad e introduciría otro orden de
ejecución que mantener. Editar android/ directamente pierde la corrección al
regenerar; una supresión o subir minSdk no corrige la compatibilidad.

## Recomendación y decisión del usuario

Se recomienda corregir los dos generadores con el mecanismo de patchedDependencies
ya existente. La petición del usuario «Pero no puedes solucionar los errores de
android lint?» solicita resolverlos ahora y sustituye la espera anterior. Se
adopta la corrección mínima del origen, sin nuevas versiones, permisos ni cambios
de minSdk, enlaces o diseño de splash. No se ocultan diagnósticos de lint.

## Implementación y mantenimiento

La propiedad de filtros se serializa como comentario XML reservado
`@generated expo-intent-filter`. El lector de Manifest restaura el indicador
interno que Expo ya usa y el escritor lo convierte a comentario: el XML final
no contiene data-generated, el merger no interpreta la marca y el APK no la
conserva. Los manifiestos antiguos con el atributo original también se migran.
Se mantienen filtros manuales y regeneración idempotente. Se descarta la marca
arbitraria tools:expoGenerated porque el merger solo admite instrucciones tools
conocidas; no se reutiliza una instrucción de merge o supresión como marcador.

El estilo base de splash conserva los tres atributos compatibles. El estilo
completo de values-v33 añade icon_preferred desde API 33; comparte referencias
al logo y colores para conservar los overlays claro/oscuro. La escritura conserva
otros estilos del mismo archivo y es idempotente. No se modifican directorios
nativos generados a mano.

Registrar hashes de los parches en el lockfile, ejecutar regresiones del generador,
prebuild repetido, build y lint de app/módulo local. Comprobar el manifest final y
los recursos compilados, además de typecheck y exportación web. La validación de
recursos por API no sustituye un recorrido visual de dispositivos no ejecutados.
Retirar los parches cuando upstream cubra estos casos y las comprobaciones pasen.
No se arrancan API local/dev ni observabilidad ni se publica producción.

## Fuentes

- [Atributos tools de Android](https://developer.android.com/studio/write/tool-attributes).
- [Recursos alternativos Android](https://developer.android.com/guide/topics/resources/providing-resources).
- Generadores instalados y lint-results-debug.xml de la revisión SDK 58.


## Cierre y evidencia

Build Debug arm64 y lint de app/módulo local correctos: cero errores, con 46 y
6 avisos respectivamente. Las 122 pruebas del cliente, typecheck, ESLint,
exportación web e instalación congelada pasan. Dos regeneraciones reales son
idénticas; aapt2 verifica el manifest y las variantes de recursos en el APK.
No se declara QA visual de API 24/34 sin ejecutarlo. Véase
[el informe](../engineering/ANDROID_LINT_FIX_2026-10-31.md).
