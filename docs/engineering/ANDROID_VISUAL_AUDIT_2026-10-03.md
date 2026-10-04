# Auditoría inicial Android — 2026-10-03

> Estado actualizado el 2026-10-04: recorrido inicial realizado en Android 14.
> Corregidas la apariencia de tabs y la oclusión del campo inferior por el teclado. No equivale
> a certificación completa de Android ni de los flujos autenticados.

## Alcance

El usuario solicita contrastar Android con la fase inicial comprobada en iOS:
arranque, Inicio, Cuenta, Ajustes, tema y apertura/cierre de Crear torneo.
Se aplican ADR-0015 (CNG), ADR-0019 (pruebas según riesgo), ADR-0054, ADR-0055,
ADR-0056 y ADR-0143. La revisión manual sobre una development build tiene menor
coste que incorporar una nueva herramienta de E2E para este preflight.
La corrección de apariencia aplica ADR-0054/0056 sin cambiar arquitectura,
dependencias ni negocio.

## Entorno y preparación

- Pixel API 34: Android 14, imagen arm64, emulador `emulator-5554`.
- Android Studio Panda 4, JDK incluido en su distribución, Android SDK existente.
- CNG: `APP_ENV=local expo prebuild --platform android --no-install` finalizó.
  Los archivos Android generados permanecen ignorados y no son fuente editada.
- Build: Gradle 9.3.1, `:app:assembleDebug`, arquitectura arm64-v8a y puerto
  Metro 8082. Resultado: `BUILD SUCCESSFUL`, 497 tareas ejecutadas en 15m 23s.
  Log temporal: `/tmp/tm-android-build.log`.
- `adb install -r` finalizó con `Success` para `app-debug.apk`.
  Paquete instalado: `com.fasttourney.app.local`.
- Metro escucha solo en `127.0.0.1:8082`. Se conectan los puertos 8082 y 8081
  mediante `adb reverse` al emulador; no se abre Metro en LAN.
- El emulador se abrió desde Device Manager. Se reabrió desde el proyecto para
  disponer de su pantalla dentro de Running Devices: la ventana independiente
  no resultó accesible con la herramienta de interacción.
- El primer arranque y compilación coinciden con alta carga del equipo. Se
  apagó iOS 27 y se canceló la sincronización Gradle duplicada del IDE; la
  compilación de terminal sigue siendo la referencia para su resultado.
- Prebuild avisa de `expo-system-ui` ausente para `userInterfaceStyle`. Es una
  observación de configuración; no prueba por sí sola un defecto visible.

## Evidencia y resultados

La sesión macOS se bloqueó durante la preparación. Tras desbloqueo manual del
usuario se recuperó la interacción restaurando la disposición acoplada de
Android Studio. El aviso «UI del sistema no responde» se cerró y no reapareció
en el recorrido observado. Los logs anteriores registraban un reinicio de
system_server por Watchdog antes de instalar FastTourney: no se atribuye a la app.

Metro se reinició tras la pausa. La development build cargó el bundle Android
(1748 módulos) desde localhost. El launcher abrió los ajustes de permiso de
superposición; se volvió con Atrás sin conceder ese permiso. Las herramientas
flotantes de Expo se distinguen de los controles del producto.

| Comprobación | Resultado observado |
| --- | --- |
| Arranque de FastTourney | Inicio visible en español; sin pantalla roja |
| Inicio y Cuenta sin sesión | Cards, campos, botones y etiquetas visibles con márgenes consistentes |
| Ajustes | Abre como ruta móvil con cierre circular; sin selector de idioma |
| Tema oscuro y claro | Contenido y controles cambian; barra de tabs corregida y verificada en ambos |
| Sistema | Selección disponible; comprobada con SO en claro, sin alternar el tema del SO |
| Atrás del sistema | Ajustes vuelve a Cuenta; Crear torneo vuelve a Inicio |
| Cierre circular | Ajustes vuelve a Cuenta |
| Crear torneo | Formulario visible en oscuro; no se crea ni publica un torneo |
| Segundo plano | Home de Android y reapertura desde icono; vuelve a Inicio en oscuro |
| Logs | Consulta de errores ReactNativeJS/AndroidRuntime sin entradas; alcance limitado a ese momento |

## Defecto corregido: barra clara con preferencia oscura

Antes del ajuste, la barra nativa de pestañas seguía clara mientras Inicio y
Cuenta usaban oscuro. La implementación instalada de Expo Router usa por defecto
colores dinámicos Material ligados al SO para fondo, etiquetas, iconos e
indicador. `unstable_nativeProps.colorScheme` no corrigió este comportamiento en
el Android probado.

`(tabs)/_layout.tsx` entrega esos colores desde los tokens del tema resuelto solo
en Android. Conserva Material y la configuración iOS existente. Fast Refresh
permitió comprobar el resultado con la misma build; no hubo cambios nativos.
Typecheck, lint del archivo y exportación web de 35 rutas pasan.

## Teclado sobre el campo inferior: corregido el 2026-10-04

Reproducción original: abrir Crear torneo y enfocar «Nombre de tu equipo» sin
haber desplazado el formulario. El teclado cubría parte del control y arrastrar
ocultaba el teclado. El problema estaba en la primitiva, que solo configuraba
insets iOS y no garantizaba visibilidad del foco.

`KeyboardAwareScrollView` mide el viewport y el campo enfocado, desplaza solo la
parte ocluida y reserva 20 px más el solapamiento que siga existiendo en Android.
No duplica la altura que ya consume el resize nativo. Reacciona al foco, teclado,
layout y tamaño de contenido; retirar el teclado elimina el espacio temporal.
También se adopta en búsqueda de administradores y transferencia, conservando
su `autoFocus`. Esas rutas autenticadas no se recorrieron visualmente.

Verificado en Pixel/API 34 con el bundle actualizado: el campo inferior queda
completo por encima de Gboard, el arrastre permite alcanzar la acción final sin
ocultar el teclado, el cambio al campo de torneo conserva el teclado y Back
restaura la altura normal. El blur sigue mostrando solo el error del campo
interactuado. No se envió ni guardó ningún torneo. Se contrastó el mismo formulario
en iPhone 18 Pro/iOS 27: control visible, entrada «Q», cambio de foco e Intro
ocultando el teclado con recuperación del layout. Typecheck, lint y exportación
web pasan.

## Límites y cobertura pendiente

Persistencia tras matar y arrancar de nuevo el proceso, tema del SO oscuro,
TalkBack, tamaño de letra ampliado, orientación, sesión autenticada, operaciones
API completas, pestaña Torneos, enlaces verificados, otras versiones/dispositivos
y release no se acreditan con este recorrido inicial.

## Retrospectiva técnica

Compilar, instalar, arrancar el sistema y recorrer el producto son evidencias
distintas. El recorrido posterior valida también el arranque y las pantallas descritas;
la instalación por sí sola no habría acreditado esas comprobaciones. Antes de atribuir un ANR al cliente, hay que identificar
el proceso afectado y reproducirlo dentro del recorrido del producto. No se han
cambiado dependencias ni políticas de bloqueo para intentar resolver el entorno.
