# Diagnóstico acotado de la caída de DevLauncher — 2026-10-10

## Problema y alcance autorizado

El usuario elige diagnosticar la caída intermitente de DevLauncher tras cerrar
las tres muestras de QA. Fuente de cliente: develop 25172c3. Se prueba Pixel
API 34, app local ficticia y Metro/dev sin observabilidad. No se cambia código,
configuración CNG, dependencias ni política de publicación.

## Identidad del binario: diferencia encontrada

Antes de reproducir, el APK instalado tiene SHA256
8fd062230d0499149209f3acd1cfd91c36395628247fb1064c83ffa17f1abe8c.
Coincide exactamente con android-debug-before-release-20261008.apk del inventario
privado del 8 de octubre. El APK Debug local recompilado el 9 de octubre tiene
SHA256 7a371eea0216490f8397afacbc3e2d4a57d3db6a763a06b551800d1b31ba502b.
Se instala con adb install -r (Success), conservando datos, y se verifica que
el hash instalado coincide con este último.

La resolución efectiva del grafo actual entrega expo-dev-client 57.0.19 y
expo-dev-launcher 57.0.20, con Expo 57.0.26 / React Native 0.86.3. El histórico
registra compilación/instalación tras actualizar esa matriz. El bundle actual
no demuestra por sí solo que el emulador conserve ese binario: las muestras
anteriores del 10 de octubre acreditan sus gestos y bundle, pero no esta versión
nativa. No se invalidan los resultados observados ni se certifica la matriz
nativa actual con aquellas muestras.

El arranque con snapshot volvió a mostrar TalkBack activo y el APK antiguo.
Es compatible con restaurar estado anterior; no se demuestra el momento exacto
en que se sustituyó. Se añade contraste sin cargar ni guardar snapshots.

## Reproducción y controles

Se usa la URL exacta publicada por Metro en 8082. Un control frío con el APK
actual devuelve Status ok, MainActivity y 1240 ms; la captura muestra Inicio
autenticado. La carga inicial confirma 1797 módulos. No son métricas de rendimiento.

Cada caso: force-stop → am start → espera indicada desde el retorno del comando
→ font_scale 1.1 → espera 2 s → font_scale 1.0 → espera 2 s. Se guardan solo las
nuevas líneas AndroidRuntime de cada caso y una captura. finally restaura escala.

| Serie | TalkBack | Espera | Nuevos FATAL | Guard Expo |
| --- | --- | --- | --- | --- |
| Inicial 1 | Activo | 0,3 s | 1 | Sí |
| Inicial 2 | Activo | 1 s | 0 | No |
| Inicial 3 | Activo | 3 s | 0 | No |
| Control 1 | Apagado | 0,3 s | 1 | Sí |
| Control 2 | Apagado | 0,3 s | 0 | No |
| Control 3 | Apagado | 5 s | 0 | No |

La captura final del control de 5 s muestra Inicio autenticado. Estos seis casos
acreditan intermitencia y que TalkBack no es necesario para reproducirla; no
estiman su tasa general ni prueban que esperar un número fijo la elimine.
La captura de recuperación inmediata solo muestra splash, no una carga terminada.

## Hechos y diagnóstico

La traza nueva incluye ActivityRelaunchItem y handleRelaunchActivity, seguidos de
MainActivity.onCreate, ReactActivityDelegateWrapper y
DevLauncherAppLoader.createOnDelegateWillBeCreatedListener (línea 49). Lanza
IllegalArgumentException: App react context shouldn't be created before.

En la fuente debug fijada, ese listener exige currentReactContext == null al
crear el delegado. DevLauncherController registra el listener antes de launch y
lo retira después de completar la carga. El loader 57.0.20 es idéntico al 57.0.14
inspeccionado antes. El manifest generado no absorbe fontScale en configChanges.

**Diagnóstico:** la recreación durante la ventana de carga alcanza de nuevo el
listener de DevLauncher con contexto React ya existente, incumpliendo su
precondición de arranque. La traza y el código sostienen esta carrera de ciclo
de vida; no se instrumenta el orden de todas sus callbacks. No se atribuye a
los formularios ni a TalkBack. El control antiguo sin global-feedback ya
reproducía el mismo guard; no se repite esa compilación ni se presenta como
control sin adaptador de la nueva matriz. La fuente del cargador está en debug:
esta prueba no reproduce ni certifica el comportamiento de una build release.

## Alternativas y recomendación

- Esperar visualmente Inicio antes de cambiar ajustes del SO y, si hay caída,
  force-stop y reabrir la URL de Metro. Coste bajo y sin cambio de producto;
  reduce el solapamiento buscado, pero no es una corrección demostrada.
- Evaluar una corrección oficial madura en una fase de actualización. Coste
  habitual de revisión de grafo, recompilación y regresión; no se afirma que
  exista una versión corregida ni se instala otra dependencia en este diagnóstico.
- Parche local del loader: coste alto de mantenimiento y pruebas del ciclo de
  vida. Quitar el require ocultaría la precondición sin demostrar la limpieza
  del host/listeners; no se recomienda como arreglo rápido.

**Recomendación:** conservar la incidencia localizada y usar la primera pauta
operativa para QA. El usuario autorizó diagnóstico, no una nueva decisión de
parche o actualización. El fallo queda reproducido y diagnosticado, no corregido.

## Evidencia y retrospectiva

Evidencia privada en /tmp/tm-devlauncher-diagnosis-20261010: results.json,
no-talkback-results.json, case-*.log/png, no-talkback-case-*.log/png y
restored-settings.json. El informe de cierre enlaza este diagnóstico.

Separar identidad APK, bundle y estado del snapshot evita atribuir resultados
al grafo equivocado. Un control sin lector descarta su necesidad para reproducir
esta caída; comparar solo arranques correctos no cerraba el antecedente.
Restaurar estado y apagar dev conserva el coste acotado del diagnóstico.

## Cierre operativo verificado

Un arranque completo con -no-snapshot y -no-snapshot-save confirma el SHA256
7a371eea0216490f8397afacbc3e2d4a57d3db6a763a06b551800d1b31ba502b,
servicios accesibles null, accessibility_enabled=0, touch_exploration_enabled=0
y font_scale=1.0. Se conserva cold-boot-verification.json privado. No se borra
el snapshot antiguo: futuras pruebas deben arrancar sin cargarlo y verificar
la identidad. Pixel se apaga de nuevo; túneles retirados y Metro/dev detenidos,
conservando volúmenes. No se activa observabilidad ni se toca producción.
Solo cambian documentos; se restaura el salto final generado de expo-env.d.ts.
Se comprueban enlaces locales y git diff --check; no se repiten suites de código.
