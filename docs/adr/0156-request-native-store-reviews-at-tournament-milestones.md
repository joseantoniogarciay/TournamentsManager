# ADR-0156: Solicitar valoraciones nativas en hitos de torneos

- **Estado:** Aceptado
- **Fecha:** 2026-10-11
- **Decisor:** Usuario, mediante decisión explícita en la conversación

## Problema, contexto y alternativas

Solicitar una valoración cuando la persona ya ha usado el producto. El usuario
propone cerrar el popup de torneo finalizado y la décima entrada al detalle de
cualquier torneo. El gate técnico está cerrado. Expo ofrece expo-store-review
para las APIs oficiales de Apple y Google; la tienda decide si muestra el diálogo
y no comunica si se ha enviado una valoración.

Un único intento por instalación tiene coste bajo pero pierde oportunidades
futuras. Dos disparadores con intervalo persistido tienen coste bajo y permiten
repetir sin insistencia. Sin intervalo se pueden encadenar peticiones. Un contador
en backend añadiría sincronización y datos sin necesidad para esta experiencia.

## Recomendación y decisión explícita

Recomendación: dos disparadores y control local compartido. El usuario acepta
«La opción 1 pero cada 60 dias»: cerrar el popup tras finalizar correctamente un
torneo, o entrar por décima vez a cualquier detalle, con 60 días entre intentos.

El contador es por instalación, suma entradas a la ficha raíz cargada con éxito
(incluidas vueltas desde otras rutas), sin contar rerenders, recargas de datos ni
torneos distintos como requisito. El hito de diez entradas se consume una vez;
las finalizaciones posteriores siguen siendo oportunidades tras 60 días.
Si el hito coincide con un diálogo abierto, espera a una superficie libre.
Todos los cierres del popup de éxito son equivalentes. Ver clasificación espera
al cierre del popup y a la solicitud antes de navegar.

Se persisten únicamente contador, hito pendiente y fecha del último intento,
sin IDs ni envío de analítica. Se guarda el intento antes de invocar la tienda,
incluso si esta no muestra el diálogo o falla. Fallos de almacenamiento impiden
la solicitud: no arriesgar peticiones repetidas. Web, builds Debug y entornos
local/dev no solicitan valoraciones. No se añade encuesta previa ni redirección
automática a la tienda. La política es independiente del adaptador nativo.

## Implementación y validación

expo-store-review 57.0.3 pertenece a la matriz Expo 57.0.26 y está publicado
desde 2026-09-11; no requiere excepción de edad. Comprobar límites de 60 días,
persistencia, concurrencia, cancelación de navegación, fallos seguros y décimo
acceso. Ejecutar gates cliente, exportación web y recompilar builds nativas.
La aparición real y el envío requieren distribución por tienda; un simulador o
build local no acreditan una valoración enviada. No publicar producción.

## Consecuencias y revisión

Reinstalar o borrar datos puede reiniciar el control local. Las cuotas de tienda
siguen aplicándose además del intervalo. Revisar tras experiencia distribuida,
cambios de las APIs o necesidad demostrada de sincronización entre dispositivos.

## Fuentes

- [Expo StoreReview](https://docs.expo.dev/versions/latest/sdk/storereview/).
- [Apple: solicitar valoraciones](https://developer.apple.com/documentation/storekit/requesting-app-store-reviews).
- [Google: momentos y cuotas](https://developer.android.com/guide/playcore/in-app-review).
- ADR-0015, ADR-0138 y ADR-0146.

## Resultado de implementación y cierre

Instalado expo-store-review 57.0.3 con Expo CLI; el lockfile añade solo ese
paquete y conserva el resto del grafo. Instalación congelada correcta. Las
bibliotecas Android review 2.0.1 y review-ktx 2.0.0 están publicadas desde 2022
en Maven Google; CocoaPods añade ExpoStoreReview desde la fuente vendorizada.

Siete pruebas de política y cinco regresiones de compatibilidad pasan, junto con
formato, lint y typecheck. Exportación web y bundles Hermes iOS/Android correctos.
Expo check offline confirma la matriz incluida; online conserva solo las cinco
revisiones jóvenes previas (expo, auth-session, constants, linking y router),
sin exclusions ni cambios para ocultarlas.

Android assembleDebug ARM64 termina BUILD SUCCESSFUL; iOS Debug ARM64 termina
BUILD SUCCEEDED. Ambas builds se instalan conservando datos en Pixel API 34 e
iPhone 17/iOS 27. No se solicita una reseña real ni se acredita el recorrido
visual de los disparadores con una distribución de tienda. Ese QA permanece
pendiente, además de los gates Android previos de ADR-0152 que este cambio no
resuelve. No se publica ni se promueve a main.

Los simuladores se arrancan solo para esta instalación y se apagan al cierre;
API, PostgreSQL y observabilidad no se arrancan. Evidencia en
/private/tmp/tm-store-review-{pods,ios-build,android-build,export}-20261011.log.

Retrospectiva: la cola y la persistencia previa evitan carreras y repetición
cuando la tienda no muestra el diálogo. La prueba de compilación e instalación
verifica integración nativa; las cuotas y el envío siguen dependiendo de tienda.
