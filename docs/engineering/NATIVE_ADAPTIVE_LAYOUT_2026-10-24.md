# Revisión de layout nativo — 2026-10-24

## Alcance y evidencia de código

Petición del usuario: corregir las apps actuales para móvil/tablet, rotación y
Split View, especialmente las oclusiones laterales de iPhone Duo; mantener el
criterio en futuros desarrollos. Decisión: ADR-0166.

| Recorrido | Corrección compartida |
| --- | --- |
| Inicio, biblioteca, Mis partidos, cuenta, formularios, ficha y subrutas | Marco lateral de Screen basado en insets dinámicos; los márgenes de Card y listas se conservan |
| Acciones flotantes y controles absolutos dentro de Screen | Se sitúan dentro del marco seguro; el padding del padre por sí solo no protegería sus offsets |
| Cuadro y selector de rondas | En apps, el modo del cuadro se decide con ancho medido del contenedor, que ya excluye oclusiones |
| Popups, confirmaciones, edición y sugerencias | ModalDialog admite cuatro giros, proveedor de safe area propio y scroll para ventanas bajas |
| Avisos iOS/Android | Margen lateral habitual más inset lateral; Android descuenta ambos insets del ancho |
| Capacidades nativas iOS | Orientación default, tablet habilitada, pantalla completa no obligatoria |
| Capacidades nativas Android | MainActivity fullUser y redimensionable mediante CNG |

No se modifica navegación, copy, operaciones OpenAPI, reglas de negocio ni API.
No se añaden dependencias. La distribución web conserva sus reglas vigentes.

## Comprobaciones automáticas y nativas generadas

- Typecheck y lint ejecutados correctamente.
- Exportación web completada: 44 rutas estáticas.
- Regresiones CNG: dos pruebas del plugin Android y dos de UIScene correctas.
- Compatibilidad de dependencias: seis pruebas correctas, incluida limpieza de
  interfaz Swift de JSI con las dos sintaxis de extensiones privadas.
- Instalación congelada con políticas de antigüedad correctas; el grafo de
  versiones y peers original se conserva, solo cambia el hash del parche JSI.
- Info.plist generado: cuatro orientaciones en iPhone y iPad,
  UIRequiresFullScreen=false; proyecto Xcode TARGETED_DEVICE_FAMILY=1,2.
- AndroidManifest generado: fullUser y resizeableActivity=true; se conservan
  adjustResize, configChanges, integración de enlaces e intent filters.
- El gate make verify incorpora test-native-adaptive-window.

`expo install --check` avisa de ocho paquetes recomendados más recientes. No se
actualiza el SDK ni sus dependencias en esta corrección. Prebuild Android conserva
el aviso previo sobre expo-system-ui; no se añade ese módulo durante este trabajo.

## Incidencia de compilación y resolución

Dos intentos con Xcode/Swift 6.4 fallan al importar ExpoModulesJSI 57.1.1:
`cannot find type __ObjC::expo in scope`. Desactivar módulos Swift explícitos
solo en una invocación no lo resuelve y no se incorpora como ajuste al proyecto.
El script de la dependencia limpia extensiones privadas con `__ObjC.`; Swift 6.4
las emite con `__ObjC::`. Un parche pnpm amplía el filtro existente para ambas
sintaxis, manteniendo API y versiones. La regresión ejecuta el filtro instalado.
La build estándar termina correctamente en 287,993 s, con cero errores y un
aviso preexistente del script de Expo Dev Launcher; sin ajustes manuales de
interfaces generadas. Tras regenerar Pods, la instalación directa en ambos
simuladores termina correctamente. El primer arranque requirió reiniciar Duo
y servir Metro en LAN: localhost estaba escuchando solo en ::1 y la app
intentaba acceder por 127.0.0.1.

Logs locales de diagnóstico: /tmp/tm-native-adaptive-build.log y
/tmp/tm-native-adaptive-patched-build.log; el primer fallo está además en
apps/client/.expo/xcodebuild.log hasta que Expo lo reemplace.

## Ejecución visual

Ejecutado con development build local, sin operaciones de negocio ni envío de
formularios:

- iPhone Duo iOS 27.1: Inicio cerrado con barra lateral a la derecha; contenido
  separado de los controles del sistema. Crear torneo abierto, giro y valor de
  campo conservado. El borrador sintético se limpió al cerrar.
- Una ruta temporal de diagnóstico compone Screen, Card, TextField, ScrollView
  horizontal y ModalDialog reales con textos de catálogo. Se retira al terminar;
  no se incorpora una nueva ruta al producto.
- Duo abierto, cerrado y parcialmente abierto; se observan las dos oclusiones
  laterales y la superior al girar. En el popup largo se conserva el valor y,
  con ventana baja, el gesto de scroll permite alcanzar y activar Cerrar.
- iPad Pro 11 iPadOS 27.0: el mismo binario universal abre Inicio y las primitivas
  de diagnóstico a tamaño tablet, sin modo de compatibilidad de teléfono.

Mediciones estables del marco interior obtenidas mediante measureInWindow en la
ruta temporal (unidades de layout, no píxeles de la captura):

| Ventana | Insets | Marco de contenido |
| --- | --- | --- |
| 951 × 669 | right=84, top=0 | x=0, y=12, width=867 |
| 951 × 669 | left=84, top=0 | x=84, y=12, width=867 |
| 678 × 466 | right=84, top=0 | x=0, y=12, width=594 |
| 669 × 951 | left/right=0, top=82 | x=0, y=94, width=669 |
| 834 × 1210 (iPad) | left/right=0, top=32 | x=0, y=44, width=834 |

Las medidas durante la transición pueden preceder a la actualización de insets;
la tabla recoge el estado asentado, no una garantía de sincronía de cada frame.

La comprobación ampliada de ventanas reducidas en iPad se intentó con
Organizador visual, tras autorización explícita del usuario. Se verificó la
selección del ajuste, la apertura de la app y su formulario/diálogo. El gesto
de redimensionar no redujo la ventana de la app ni la de Ajustes. Al desplegar
un panel del sistema y volver apareció una superficie negra; abrir Ajustes por
simctl tampoco respondió hasta reiniciar el simulador. Se recuperó el entorno
y se verificó de nuevo Apps en ventanas como opción seleccionada. Estos hechos
no permiten certificar Split View ni atribuir la causa del bloqueo a la app.

Tras recuperar el simulador, el diálogo se comprobó con teclado software visible:
el campo y el botón Cerrar quedaron por encima del teclado; se editó el valor y
se activó Cerrar sin ocultar antes el teclado. El valor se conservó en la pantalla
subyacente. No se envió ninguna operación de negocio.

Como comprobación independiente, Resize Mode de Device Hub en Duo sí actualizó
las medidas a 375 × 678 (marco x=0, y=32, width=375, top=20), pero la captura
mostró una superficie negra durante esa modalidad. No se certifica como prueba
visual de multiventana ni se atribuye el origen a la app o al simulador sin una
reproducción aislada. Se abandonó esa modalidad y se restauró su estado apagado.
La prueba posterior en iPad también encontró un bloqueo del entorno; sigue
pendiente una reproducción aislada en iPadOS 27. La repetición siguiente sí
aporta una ejecución válida de Split View en iPadOS 18.5.

La ruta temporal se ha eliminado. Metro y los dos simuladores arrancados para
la sesión quedan apagados, conservando sus datos y las builds. La captura de
teclado se restituyó a off. No se inició backend ni observabilidad.

Pendiente además: QA visual Android y recorridos con torneos
cargados (clasificación/cuadro, enlaces y acciones de gestión). La inspección del
código y las regresiones no se presentan como ejecución de estos recorridos.

## Repetición de Split View solicitada por el usuario

Se repite el recorrido con el mismo binario universal. En iPadOS 27, Apps en
ventanas sigue activo; el gesto y el puntero simulado no consiguen reducir el
panel. Se restaura la entrada táctil. No se cambia de nuevo a Organizador visual.

Como contraste, se instala el mismo binario en el iPad Pro 11-inch (M4) con
iPadOS 18.5 disponible. El menú de tres puntos permite elegir Split View y
Safari como segunda app, sin navegar a ningún sitio. Se comprueba:

- Inicio real en panel vertical compartido; el texto se redistribuye y las tabs
  pasan a la parte inferior.
- La ruta temporal con primitivas reales mide un marco de 504 puntos en vertical.
- En horizontal se desplaza el divisor y se observan marcos de 600, 375 y 825
  puntos, todos con altura 798, x=0 e y=36 (top=24, bottom=25). Las medidas de
  useWindowDimensions pueden ir un evento por detrás; las anchuras anteriores
  son las medidas del marco, contrastadas con el cambio visible de panel.
- El texto sintético QA Split View se conserva al girar, estrechar y ampliar.
- ModalDialog queda contenido en el panel de 375 puntos; el texto se redistribuye
  y el control Cerrar permanece accesible. Con teclado software visible el campo
  queda por encima de él. El gesto sobre el contenido oculta el teclado en este
  recorrido; después se activa Cerrar. No se acredita cierre con teclado aún
  abierto para esta combinación estrecha.
- Inicio y Crear torneo reales abren en el panel ancho junto a Safari. El campo
  conserva el valor al salir del reparto hacia pantalla completa. Se limpia y se
  cierra el borrador sin iniciar sesión ni enviar operaciones de negocio.

No se observa una superficie negra de la app en esta ejecución de Split View.
No se añaden correcciones de producto a partir de esta repetición. Esto valida
los recorridos enumerados en iPadOS 18.5, sin extender el resultado a iPadOS 27,
Android ni cuadros y clasificaciones cargados. La ruta temporal vuelve a retirarse;
Metro y los simuladores de esta repetición quedan apagados.

## Checklist de cierre y límites

Se revisan reglas de cliente: tokens compartidos, localización sin copy nuevo,
44 px sin alterar, márgenes de Card, separación superior, reserva inferior del
contenido, navegación/tema y arquitectura de features. No se toca OpenAPI ni
feedback de negocio. La regla de no usar SafeAreaView de React Native se mantiene:
el diálogo emplea la implementación de react-native-safe-area-context.

La matriz completa incluye Duo cerrado/parcial/abierto, los giros admitidos,
tablet, Split View, teclado, scroll, extremos de tabla/cuadro y conservación de
estado. Solo se acreditan los recorridos enumerados como ejecutados; cualquier
combinación restante continúa pendiente. No se publica ni despliega producción.

## Retrospectiva técnica

Una configuración de orientación abierta habilita el recorrido, pero no resuelve
la geometría. El marco seguro común protege contenido estático, listas y controles
absolutos; el overlay necesita su propia ventana segura. Medir el contenedor
permite responder a ventanas pequeñas sin crear variantes por modelo. La QA visual
sigue siendo necesaria para detectar insets o comportamientos de UIKit que el
contrato de configuración no predice.


## Ampliación Duo con datos reales — 2026-10-31

Se reconstruye e instala la build universal con la firma habitual de Xcode.
Una primera build sin firma no permitió SecureStore por falta de entitlements;
se descarta para QA de sesión y se reconstruye con firma. La build final pasa.
Se arranca exclusivamente el entorno local API/PostgreSQL/Mailpit, sin
observabilidad, y se aplica su migración pendiente 23 con el migrator existente.
Se usan una cuenta y cuatro torneos sintéticos de ocho equipos: liga y cuadro
de fútbol, liga de voleibol y cuadro de tenis a cinco sets. No se usan datos ni
servicios de producción. No se cambia código de contrato o backend.

### Fallo reproducido y corrección

Tras girar una clasificación presentada con Duo parcialmente abierto, Pts quedó
bajo los controles laterales. useWindowDimensions ya indicaba 951 × 669, pero los
insets heredados de la ruta cubierta seguían en top=82/right=0. Un provider en la
escena de navegación visible actualiza a top=0/right=84/bottom=34 y delimita un
marco de 867 puntos. Se contrasta la corrección en la tabla real. El horizontal
opuesto informa left=84 y un marco de 594 puntos en ventana de 678.

El stack raíz aporta SafeAreaProvider por ruta nativa mediante screenLayout;
no se cambia su key al girar. La tab o ruta cubierta no es fuente de geometría
de una escena presentada. FullWindowOverlay mide igualmente su propia ventana
para colocar el banner con su margen habitual dentro de la zona segura.

### Recorridos ejecutados

- Inicio y acceso a cuenta; teclado software en Duo abierto, giro con email y
  contraseña conservados, e inicio de sesión mediante UI con la build firmada.
- Clasificación real de voleibol: ocho equipos, nombre largo, datos de nueve
  estadísticas y puntos fijos. Cerrado, parcialmente abierto y abierto, con
  protección lateral contrastada en ambos lados. En horizontal corto se alcanza
  la última fila; el offset final medido es 78. En ancho suficiente se ven todas
  las estadísticas hasta CT. En cerrado vertical se llega también a CT mediante
  gesto horizontal: offset final 170, cabecera y valores sincronizados, equipo y
  puntos fijos y protegidos. El diálogo de reglas abre y se cierra con Ok en
  poca altura. No se modifica la lógica de clasificación.
- Cuadro real de fútbol: desde cerrado se selecciona Final y se sigue el enlace
  al partido de origen de semifinales. Al abrir Duo se conserva Semifinales y se
  muestra la vista general, incluida la columna Final. Se abre Añadir resultado
  desde la semifinal con permisos reales del creador.
- Se introduce 2–1 en el editor, se pliega Duo y se recorren giros con teclado
  software y valores conservados. En un horizontal corto Guardar se activa con
  el teclado visible. En el opuesto se desplaza el contenido, lo que oculta el
  teclado por la política on-drag existente, y Guardar queda visible.
- Se detiene solo la API local antes del primer guardado: el banner muestra
  common_network_error sobre el diálogo, sin exponer el error técnico ni perder
  el marcador. Se restablece la API y se guarda desde UI. Una lectura pública
  independiente confirma semifinal completed, homeScore=2 y awayScore=1.
- Editor real de tenis a cinco sets: en Duo cerrado vertical se desplaza hasta
  ver Set 5 y Guardar completos. Diagnóstico temporal: contenido de 603 puntos,
  viewport de 471, offset final aproximado de 132. También se alcanzan Set 5 y
  Guardar completos en horizontal corto sin instrumentación. Al enfocar Set 5,
  el campo queda por encima del teclado software. No se guarda un resultado
  incompleto ni se modifica la validación del deporte.

### Límites que siguen abiertos

No se acredita Guardar en cada horizontal con teclado simultáneo para el
formulario de cinco sets, ni la edición de tenis persistida. Los controles presentes en accesibilidad no prueban alcance táctil.
La combinación de recarga de Metro, modal y giro llegó a mostrar una superficie
negra, recuperada tras relanzar y girar a vertical; no se ha aislado su origen.

Se mantiene pendiente QA visual Android y la ejecución de ventanas reducidas en
iPadOS 27. El Split View de 375/600/825 puntos ya descrito corresponde a iPadOS
18.5. Esta ampliación no certifica todas las acciones de gestión ni todas las
combinaciones de deporte, postura y teclado. No se declara una garantía completa.

Los listeners y medidas de diagnóstico se retiran. Se conservan las builds y los
datos sintéticos locales para reproducir los casos. Se cierra la sesión sintética desde UI. Metro, API, PostgreSQL, Mailpit y el
simulador Duo quedan apagados, conservando volúmenes y builds. No se inició
observabilidad.

Cierre automático: 10 regresiones de configuración/escena/dependencias pasan;
typecheck, lint y exportación web pasan con el código final. Se comprueban
formato y git diff --check. La checklist de cliente se revisa sin nuevo copy,
reglas de negocio, DTOs ni dependencias adicionales. La revisión amplía ADR-0166;
no cambia la decisión de producto ni publica producción.


### Intento de ejecución Android

Se encuentra el AVD Pixel_API_31 (arm64) existente y se intenta arrancar sin
crear un nuevo snapshot. QEMU 37.2.12 informa de hilos CPU y bucle principal
bloqueados, ADB permanece offline y el proceso termina con código 139 antes de
instalar la app. Se repite sin cargar snapshot y con GPU swiftshader: mismo
bloqueo y código 139. No se borra userdata ni se altera el AVD Taykus_Fold.

Este fallo de arranque del emulador impide QA visual Android en esta sesión;
no es una ejecución fallida de FastTourney. El manifest Android generado sí
contiene MainActivity fullUser, resizeableActivity=true y adjustResize, además
de sus integraciones previas. La regresión CNG pasa. Logs locales:
`/tmp/tm-android-adaptive-emulator-20261031.log` y
`/tmp/tm-android-adaptive-emulator-cold-20261031.log`.


La build Android local arm64 sí compila: `app:assembleDebug` termina BUILD
SUCCESSFUL con JDK 21.0.12 en 2 min 17 s tras preparar las dependencias nativas.
La primera ejecución con JBR 25 de Android Studio falló en configureCMake de
screens/worklets con el aviso de método restringido de java.lang.System; con
el JDK 21 ya instalado pasan ambas etapas, sin cambiar el stack ni dependencias.
El manifest empaquetado confirma fullUser, resizeableActivity=true y adjustResize
en MainActivity. Hay advertencias de deprecación de dependencias/Gradle; no se
presenta como una compilación sin avisos.

APK conservado en `apps/client/android/app/build/outputs/apk/debug/app-debug.apk`
(generado e ignorado por Git). Log:
`/tmp/tm-android-adaptive-jdk21-build-20261031.log`. No se instala porque ambos
arranques del emulador fallaron antes. ADB y los daemons Gradle de esta sesión se
apagan al terminar; no se ha reiniciado backend ni Metro para este intento.


## Corrección del emulador Android y nueva ejecución — 2026-10-31

A petición del usuario se investiga el arranque en lugar de mantenerlo como
bloqueo. Los .ips de macOS revelan MainLoopThread esperando dentro de CoreAudio
al crear la captura de entrada (AudioDeviceCreateIOProcID_mac_imp,
ca_init_voice_locked, AUD_open_in). Las CPU esperan; el detector de bloqueo
invoca después GenerateDumpAndDie y provoca el SIGSEGV. No es una caída de la app.

**Contraste y corrección:** el mismo AVD Pixel_API_31 arranca con -no-audio en
19,5 s. Se apaga y se guarda config.ini antes de modificar únicamente
hw.audioInput=yes a no. Se conserva hw.audioOutput=yes y el resto de datos. El
arranque habitual, sin forzar GPU software ni desactivar todo el audio, completa
en 14,4 s con el renderer Apple M2 Pro. También arranca desde Running Devices
de Android Studio. ADB llega a device y sys.boot_completed=1 en ambos casos.
La copia original está en /tmp/tm-Pixel_API_31-config-before-audio-fix.ini.
La corrección local del AVD queda aplicada por petición del usuario. No se cambian
preferencias de audio del host ni se toca Taykus_Fold.

**JDK del IDE:** al abrir el proyecto, Android Studio generó un criterio de JVM
25 que no se resolvía cambiando solo JAVA_HOME. Se genera criterio 21 con la tarea
oficial updateDaemonJvm y toolchainDownloadUrls vacío, para usar JDK local sin
nuevos repositorios. assembleDebug vuelve a pasar (34 s, arm64); la sincronización
del IDE pasa también (registro de las 15:03:32). Los archivos del IDE/nativos
son locales e ignorados; no se cambia stack ni código de producto.

**Recorridos Android ejecutados hasta aquí:** Inicio real, Crear torneo,
teclado software, texto sintético Qa conservado en giros 0/90/180/270, con campo
visible por encima del teclado y barra del sistema en ambos laterales. El giro
libre ya estaba activado y se conserva su ajuste. Se limpia y cierra el borrador
sin iniciar sesión ni crear un torneo. La clasificación de voleibol carga los
ocho equipos desde la API local; en vertical se desplaza hasta RR (CT en español),
con cabecera sincronizada y equipo/puntos fijos. En horizontal corto se alcanza
la última fila y los puntos quedan fuera de la barra lateral.


Esta ampliación resuelve el bloqueo de arranque Android descrito anteriormente.
La ejecución visual acreditada se limita a los recorridos enumerados: no se
certifican todavía diálogo de reglas, cuadro ni ventana dividida Android. Los
intentos sobre controles pequeños de cabecera en la vista de desarrollo no
aportaron una activación verificable; no se atribuyen al producto sin aislarlos.
Las pruebas iPadOS 27 siguen con el límite ya documentado.

Al cerrar se apagan Pixel, Android Studio (abierto para esta prueba), Metro, API,
PostgreSQL, Mailpit, ADB y el daemon Gradle. Se conservan volúmenes, APK y datos;
se verifica que hw.audioInput=no y hw.audioOutput=yes permanecen en el AVD. No se
inicia observabilidad ni se modifica código de producto en esta ampliación.
La build corregida pasa y sync del IDE pasa. Se retira el cambio de formato que
Metro regenera en expo-env.d.ts; git diff --check y formato vuelven a comprobarse.

Retrospectiva: examinar la pila anterior al detector de bloqueos y contrastar una
sola variable produjo una corrección menor que actualizar todo el emulador o
borrar el AVD. Las pruebas de ejecución siguen separadas del éxito de compilación.
