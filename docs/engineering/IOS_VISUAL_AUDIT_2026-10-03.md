# Auditoría nativa iOS — 2026-10-03

> Estado: incompatibilidad UIScene corregida; build Debug observada en iOS 27
> e iOS 18.5. En iOS 27 se validaron enlaces de esquema en frío y segundo plano,
> navegación inicial y persistencia del tema. Auditoría completa, sesión
> autenticada, Universal Links y distribución pendientes.

## Objetivo y decisiones vigentes

El usuario solicita comprobar funcionamiento y coherencia visual dentro de cada
SO, comenzando por iOS. ADR-0008 y ADR-0055 permiten conservar flujos compartidos
y adaptar navegación y presentación. ADR-0054, ADR-0096 y el sistema de diseño
definen tokens, primitivas y Figtree; ADR-0056 y ADR-0143 definen tema persistente
e idioma automático. No se propone un rediseño ni una nueva herramienta de E2E.

La revisión manual con evidencia tiene menor coste inicial que una suite extensa
de interfaz. Conforme a ADR-0019, la recomendación es identificar primero riesgos
reales y reservar automatización posterior para recorridos críticos repetibles.

## Evidencia del preflight

- `simctl` enumera iPhone 16 con iOS 18.5 y dispositivos con iOS 27.0.
- Se arrancó iPhone 18 Pro, identificador
  `1094CE55-E226-442A-A092-9EB662E897EA`, con iOS 27.0.
- El inventario de ese simulador no contiene identificadores `com.fasttourney`
  ni `host.exp.Exponent`. Falta instalar una app adecuada para el recorrido.
- No existe `apps/client/ios`; el proyecto mantiene generación nativa CNG.
- No se detectó Metro escuchando en 8082 durante el preflight.
- `expo install --check` terminó con código 1 y recomendó versiones distintas
  para Expo, módulos Expo y React Native. No se actualizaron dependencias.
  Este resultado es un aviso de compatibilidad; no demuestra un fallo de UI ni
  decide por sí mismo la actualización del stack.
- La herramienta de interfaz reportó el Mac bloqueado incluso después de que el
  usuario indicara haberlo desbloqueado y de reiniciar la conexión. No fue posible
  obtener el árbol accesible ni capturas del simulador mediante esa herramienta.

## Recorridos a ejecutar

Todos están **pendientes**. Esta tabla es una checklist, no resultados.

| Área                  | Comprobación                                                                                                                                    |
| --------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------- |
| Entorno               | Build y revisión de código identificadas, API de pruebas y estado de sesión conocidos; comprobar módulos nativos antes de diagnosticar estilos. |
| Inicio/Torneos/Cuenta | Tabs, reinicio de navegación, acceso a ajustes, último elemento visible por encima de la botonera y acción flotante.                            |
| Cabeceras y modales   | Atrás, cierre, títulos largos, margen frente a controles y retorno seguro con/sin historial.                                                    |
| Apariencia            | Claro, oscuro y sistema; persistencia al relanzar; cabeceras, tabs y transiciones con tema coherente.                                           |
| Formularios           | Teclado, foco, desplazamiento hasta acción final, validación por campo, controles táctiles y envío único.                                       |
| Cuenta                | Acceso, registro, ajustes, notificaciones y estados terminales de enlaces.                                                                      |
| Torneos               | Creación, liga, clasificación, eliminatorias y formato mixto; participantes, resultados, incidencias y acciones por permiso.                    |
| Diálogos              | Confirmación visible sobre el modal activo; backdrop y cierre; contraste y texto largo.                                                         |
| Estados               | Carga inicial, vacío, fallo de red y reintento; mensajes seguros y sin duplicación.                                                             |
| Accesibilidad         | Texto ampliado, etiquetas, estados de selección y VoiceOver; registrar explícitamente si no se prueba con lector real.                          |
| Variantes de iOS      | Contrastar iOS 27.0 con iOS 18.5 por las diferencias aceptadas de Liquid Glass y controles de cabecera.                                         |

Las pruebas autenticadas pueden crear cuentas ficticias en Yopmail o directamente
en base de datos, según autorización del usuario. Se recomienda una base local
aislada para fixtures, sin introducir datos reales. Un fixture directo en base de
datos no acredita registro, entrega de correo ni verificación end-to-end.

Cada hallazgo deberá registrar dispositivo/SO, build, entorno, tema, ruta,
precondiciones, reproducción, resultado esperado y observado y evidencia visual.
La splash distribuida requiere una build release; una development build no la
certifica. Una revisión de simulador tampoco acredita dispositivos físicos.

## Retrospectiva del preflight

El arranque de un runtime por CLI no acredita acceso a su interfaz ni que exista
una build instalada. Los checks de dependencias y la revisión visual aportan
evidencias distintas. La validación sigue pendiente: no hay hallazgos visuales
confirmados, cuentas creadas ni cambios de implementación derivados de esta fase.

## Reanudación: acceso al Mac y build local

Tras desbloquear de nuevo la sesión, la herramienta pudo consultar Finder y el
escritorio. A petición del usuario, se activó `caffeinate -di -t 14400` como
inhibición temporal de suspensión del sistema y pantalla por inactividad, durante
cuatro horas. `pmset -g assertions` confirmó ambas inhibiciones. No se cambiaron
preferencias permanentes ni políticas de seguridad; esto no evita un bloqueo
manual ni acredita el comportamiento de políticas externas.

- Expo generó `apps/client/ios` mediante CNG e instaló CocoaPods. Estos archivos
  permanecen ignorados y no se editaron manualmente.
- `expo run:ios --device <UUID>` pidió escoger un equipo Apple. Se canceló sin
  seleccionar ninguno y se compiló el workspace generado con `xcodebuild`, SDK
  `iphonesimulator`, destino iPhone 18 Pro y `CODE_SIGNING_ALLOWED=NO`.
- Resultado: `BUILD SUCCEEDED`, duración informada por Xcode 383,537 s.
- Build Debug local: `com.fasttourney.app.local`, instalada con `simctl install`
  en el dispositivo del preflight. El producto generado permanece en
  `/tmp/tm-ios-audit-build/Build/Products/Debug-iphonesimulator/FastTourneyLocal.app`;
  el log está en `/tmp/tm-ios-audit-build.log`.
- La API local respondió HTTP 200 en `http://127.0.0.1:8081/healthz`.
- Metro se preparó en localhost:8082, development client, con la variable temporal
  `EXPO_PUBLIC_API_BASE_URL=http://127.0.0.1:8081/v1`, sin editar `.env`.
  No se conectó la interfaz a esa API ni se crearon fixtures.
  Se detuvo Metro al confirmar el bloqueo de Device Hub; la inhibición temporal
  de suspensión conserva su vencimiento de cuatro horas.
- La apertura de Device Hub en Xcode 27 quedó en su proceso interno
  `DevicesSystemUpdater`. Su árbol accesible y captura muestran exclusivamente
  «Installing system components…» y un indicador de actividad. Persistía tras
  más de quince minutos; no se observó una ventana utilizable del simulador.
  El log de instalación registra `Package Authoring Error` para esos procesos;
  no se ha demostrado que ese aviso sea la causa del bloqueo.

La preparación reemplaza las observaciones iniciales de ausencia de build y
Metro, pero todos los recorridos de interfaz siguen pendientes. No se presenta
la compilación como una prueba de geometría, navegación, accesibilidad o API
end-to-end. La retrospectiva es conservar la distinción entre generar/compilar,
instalar y poder observar/interactuar con la app: son gates separados.

## Corrección prioritaria de iOS 27

El usuario autorizó primero corregir el requisito de iOS 27 porque puede influir
en el resto de la revisión (ADR-0144). Tras conceder permisos a la instalación,
Device Hub terminó y permitió interactuar con ambos simuladores.

### Reproducción y corrección

La build original volvió a SpringBoard en dos intentos de arranque en iOS 27.
El log de UIKit identifica la ausencia de ciclo de vida UIScene como causa. La
misma build alcanzó el development launcher en iOS 18.5; no se llegó entonces a
certificar la app completa.

El config plugin fuente declara una sola escena y asocia su ventana a
UIWindowScene. La integración de Expo y React Native permanece en AppDelegate,
recibiendo eventos y enlaces. La generación incorpora Swift al fichero ya
compilado por el proyecto, sin editar salidas de CNG ni actualizar dependencias.

La nueva build compiló (`BUILD SUCCEEDED`, 388,212 s). Se instaló en iPhone 18
Pro/iOS 27 e iPhone 16/iOS 18.5. En iOS 27 se observó el development launcher
estable y después la home del producto. Desapareció el rechazo nativo inicial.

La carga de React reveló un error distinto: HomeMetadata montaba Expo Head en
iOS, que exigía un origen Handoff. ADR-0120 limita esos metadatos SEO a web; se
añadió la condición de plataforma. Tras recargar, Inicio mostró su contenido,
acción Crear y preferencia de analítica sin ese render error.

### Límites de entorno y validación pendiente

- Metro `--localhost` escuchaba inicialmente solo en `::1`. Se reinició con
  `NODE_OPTIONS=--dns-result-order=ipv4first`; `lsof` confirmó exclusivamente
  `127.0.0.1:8082` y el launcher cargó el bundle iOS desde esa dirección.
- La build compilada con `CODE_SIGNING_ALLOWED=NO` mostró un error Keychain de
  entitlement ausente. Se recompiló con identidad ad hoc `-` sin equipo Apple
  (`BUILD SUCCEEDED`, 51,512 s). El artefacto resultante tenía entitlements vacíos.
  Se firmó el artefacto temporal de simulador con application-identifier y
  keychain-access-groups limitados a `com.fasttourney.app.local` y get-task-allow.
  Esto no modifica la configuración fuente ni una build distribuida. La lectura
  de Keychain con ese artefacto sigue pendiente de comprobación en ejecución.
  El artefacto con esos entitlements quedó instalado en ambos simuladores.
- El Mac volvió a bloquearse durante la verificación: la inhibición `-di`
  evitaba suspensión pero no ha garantizado evitar el bloqueo de sesión.
  No se cambiaron políticas de seguridad ni se intentó desbloquear la sesión.
- Falta comprobar la nueva escena en ejecución en iOS 18.5, enlaces de esquema
  en frío/caliente, Universal Links, retorno desde segundo plano, y una build
  release/dispositivo físico. No se declara la auditoría completa terminada.

Pruebas del transformador: idempotencia, conservación de otras integraciones y
rechazo de plantillas desconocidas; pasan y forman parte de `make verify` mediante
`make test-ios-scenes`. Typecheck, lint de los archivos modificados y exportación
web pasaron después de limitar HomeMetadata a web. No hubo cambios OpenAPI ni de
backend, ni se crearon cuentas o torneos.

Retrospectiva: corregir el primer fallo de arranque descubre errores de capas
posteriores. Una escena válida no demuestra restauración de sesión; una firma
ad hoc tampoco demuestra entitlements correctos. Registrar cada evidencia evita
confundir compilación, montaje de React y funcionamiento completo.

Checklist de cliente: CNG y plugin fuente; sin nuevos tokens, copy, dependencias
ni cambios de formularios o navegación JS; metadatos web conservados; ningún
endpoint tocado. Las comprobaciones nativas aún pendientes quedan explícitas.

## Validación tras restaurar la firma de Xcode

La firma manual temporal anterior impidió iniciar el proceso: RunningBoard
registró `Launchd job spawn failed` con código POSIX 163. No se identificó cuál
de los entitlements añadidos produjo el rechazo. Se retiró esa firma mediante
una reconstrucción incremental con la firma ad hoc que genera Xcode, sin cambiar
el equipo ni los entitlements fuente. Resultado: `BUILD SUCCEEDED` (11,043 s),
instalación en ambos simuladores y arranque estable. El `.xcent` de firma vacío
no representa todos los entitlements del simulador: Xcode genera por separado
`FastTourneyLocal.app-Simulated.xcent` y los incorpora durante el enlace. No se
requiere mantener el plist temporal añadido manualmente.

Metro se reinició en localhost con IPv4; la API local volvió a responder 200
al comprobarla con acceso de red permitido. No se modificó ningún servicio.

| Comprobación                            | Resultado observado                                                                                                                                   |
| --------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
| iOS 27: Inicio, Cuenta y Ajustes        | Cargan; sin el render error de Head ni el error previo de lectura Keychain en el log de Metro.                                                        |
| iOS 27: cambio a Oscuro                 | El modal y las pantallas reflejan el tema.                                                                                                            |
| iOS 27: enlace con app en segundo plano | Safari abre `fasttourney-local://link/password-reset`; la app muestra el estado seguro de enlace inválido al no incluir token.                        |
| iOS 27: cierre y enlace en frío         | Se cerró solo Fast Tourney desde App Switcher. El enlace abre el launcher; al seleccionar Metro se conserva la ruta de recuperación y el tema oscuro. |
| iOS 27: recuperación y creación         | Volver a la home funciona; Crear torneo abre el formulario con cabecera y cierre nativos. El cierre devuelve a tabs.                                  |
| iOS 18.5: misma build                   | Arranca, carga Inicio desde Metro y abre/cierra Crear torneo. No se observó rechazo UIKit.                                                            |
| Persistencia y limpieza de preferencias | Ajustes conserva Oscuro tras cerrar la app; después se restauró Sistema en iOS 27. Analítica opcional permanece desactivada.                          |

El arranque en frío de una development build necesita seleccionar Metro; este
resultado no acredita el arranque autónomo de una release. El enlace probado no
lleva token y no demuestra verificación, sustitución de sesión ni Universal
Links. La ausencia del error de lectura no prueba escritura o restauración de
credenciales SecureStore. No se crearon cuentas ni torneos en esta fase.

Retrospectiva de esta fase: la corrección de escenas queda comprobada en el
arranque de ambas versiones y en la entrega de enlaces de esquema al cliente.
Para simuladores, conservar la firma que genera Xcode es suficiente para esta
validación; añadir entitlements manualmente introdujo otro fallo y se revirtió.
La auditoría funcional y visual completa debe seguir con una cuenta ficticia y
recorridos autenticados, además de release/dispositivo físico y Android.


## Regresión de teclado compartido — 2026-10-04

Se cargó el bundle actualizado en la build instalada de iPhone 18 Pro/iOS 27.
En Crear torneo se activó explícitamente el teclado software desde Device Hub;
el teclado del Mac conectado no servía para comprobar oclusión. «Nombre de tu
equipo» quedó entero encima del teclado con margen. La tecla Q introdujo texto,
cambiar a «Nombre del torneo» mantuvo el teclado y pulsar Intro lo ocultó. El
formulario recuperó su disposición normal y mostró únicamente la validación del
campo vacío abandonado. No se envió el formulario.

La primitiva conserva los insets nativos iOS y añade desplazamiento al foco.
Esta comprobación no acredita iOS 18.5 con el nuevo comportamiento, dispositivos
físicos, teclado flotante iPad, orientación ni todas las rutas autenticadas.
