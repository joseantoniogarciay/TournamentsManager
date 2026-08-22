# Registro de aprendizaje

## 2026-08-21 — Un vertical slice necesita una frontera de cierre

Fase 2 demostró pronto el recorrido acordado de identidad, sesión, publicación y
lectura de liga, pero el cierre formal se pospuso mientras el backend incorporaba
el ciclo deportivo, administración y recuperación. El producto ganó capacidad,
pero la documentación de estado perdió el punto exacto en el que ya existía
evidencia suficiente.

**Regla reutilizable:** cerrar la retrospectiva cuando se cumple el criterio de
salida y registrar las ampliaciones como incrementos posteriores. Una fase
cerrada no impide seguir construyendo; separa lo aprendido de lo que se decidió
añadir después.

## 2026-08-21 — Una alerta se valida por disparo, entrega y resolución

Una alerta sintética demuestra que el canal de notificación funciona, pero no
que la expresión de Prometheus alcance su umbral. La caída controlada de
PostgreSQL demostró en local el recorrido completo de
`SessionRefreshFailureRateCritical`: respuestas `5xx`, estado activo, correo en
Mailpit y resolución tras recuperar la dependencia. En `dev`, una alerta marcada
`test=true` aisló la entrega mediante Resend, Cloudflare y el buzón final sin
interrumpir el servicio público.

**Regla reutilizable:** verificar por separado la detección y el transporte, y
cerrar siempre la prueba confirmando la recuperación. Una alerta activa sin
entrega o una entrega sintética sin regla evaluada dejan preguntas diferentes
sin responder.

## 2026-08-21 — Un archivo de secreto no es un archivo de variables

Docker monta el contenido literal de un secreto. Los comentarios conservados
desde un archivo de ejemplo pasaron a formar parte de la contraseña SMTP de
Alertmanager y Resend rechazó la autenticación, mientras la API funcionaba
porque su archivo `.env` sí interpreta comentarios y asignaciones.

**Regla reutilizable:** un secreto de archivo contiene solo el valor esperado
por el proceso. Su preflight comprueba presencia, número de líneas y forma sin
mostrarlo; la documentación distingue expresamente secreto literal de contrato
`.env`.

## 2026-08-21 — La entrega de una alerta no debe compartir la credencial del producto

Alertmanager y la API pueden usar el mismo SMTP de Resend, pero tienen radios de
impacto distintos: uno comunica degradación operativa y el otro entrega enlaces
de identidad. Una clave *Sending access* exclusiva permite revocar o rotar el
canal de alertas sin impedir verificaciones ni recuperación de cuentas. STARTTLS
en el puerto 587 cifra la conexión antes de enviar esa clave.

La paridad útil entre local y `dev` es compartir reglas, dashboard y señales, no
abrir Grafana al público ni mezclar volúmenes o logs. Los servicios conservan
nombres internos iguales para reutilizar su configuración, mientras Promtail
filtra el namespace Compose para que cada entorno observe solo su propia API.

**Retrospectiva técnica:** duplicar el stack suma seis contenedores y operación,
pero evita una abstracción o SaaS nuevo y conserva la práctica de diagnóstico
correlacionado. La siguiente revisión debe basarse en ruido, cuota o necesidad de
guardias, no en añadir alertas por cobertura aparente. Véase ADR-0100.

## 2026-08-21 — La excepción de Expo cubre el cierre exacto que el resolvedor necesita

`expo install --check` identifica los paquetes directos desalineados, pero su
actualización puede requerir además parches transitivos jóvenes. La excepción de
ADR-0077 se forma con las versiones exactas que pnpm rechaza durante esa misma
resolución, no con prefijos de paquetes: así se recupera la matriz compatible sin
abrir una vía permanente para futuras publicaciones. Los directos quedan fijados
y React Native conserva la versión que corresponde al SDK de Expo.

## 2026-08-16 — Un control creado dinámicamente debe continuar la tarea

Al añadir un campo repetible, el foco pasa al nuevo control una vez montado. La
acción mantiene así la continuidad con el teclado en web, iOS y Android, y evita
que una persona escriba por error en el campo anterior o deje el nuevo equipo
vacío. `TextField` reenvía la referencia nativa en vez de que cada formulario
recree su implementación.

## 2026-08-16 — Parametrizar reduce la probabilidad; el privilegio limita el impacto

Una consulta parametrizada impide que una entrada se interprete como SQL, pero
no convierte en inocua una credencial comprometida. Separar propietario sin
login, migrador y runtime deja a la API con el conjunto de operaciones que usa
realmente y convierte los permisos de una tabla nueva en parte visible del
cambio de esquema. No hace falta un rol por módulo mientras el backend siga
siendo un único proceso; ese aislamiento sería coste sin límite de despliegue
que lo aprovechase. Véase ADR-0097.

## 2026-08-16 — El peso de una acción se define en su primitiva

La legibilidad de una acción textual mejora al usar semibold sin aumentar su
tamaño ni cambiar el objetivo táctil. Aplicarlo en `Button` propaga el peso real
Figtree 600 a todas sus variantes y evita ajustes tipográficos repetidos en cada
pantalla.

## 2026-08-13 — El marcador de una fila necesita jerarquía contenida

La escala `display` (32 px) es adecuada para un título o un resultado aislado,
pero comprime visualmente una fila de partido y puede dejar los glifos demasiado
cerca de su borde. Un marcador dentro de una `Card` usa `title` en negrita:
preserva la lectura inmediata sin introducir un token, variante o componente
específico antes de que el patrón se repita.

## 2026-08-13 — Un modal transversal no pertenece al stack de una tab

Una ruta puede conservar su URL bajo `/account` y, aun así, presentarse desde el
stack raíz. Ajustes y Notificaciones se elevan para que web, iOS y Android
compartan una salida explícita con X hacia Cuenta. iOS y Android las presentan
modalmente; web usa la página estable del stack porque el modal experimental de
Expo Router puede reconstruir y sustituir la ruta activa al redimensionar.

## 2026-08-13 — Una tab web recargada no conserva una pila que no existe

Tras recargar una URL profunda, el navegador recupera la ruta pero no la pila
interna de su tab. Pulsar esa misma tab debe reemplazar explícitamente su URL por
la raíz; confiar solo en `popToTop` no cambia nada cuando no hay entradas que
descartar. Android e iOS delegan el mismo gesto repetido en `NativeTabs`.

## 2026-08-12 — SMTP conserva la portabilidad si el dominio recibe un puerto

El proveedor de correo resuelve entrega, reputación y DNS; el caso de uso solo
necesita pedir que se entregue un enlace. Mantener SMTP en el adaptador evita
que una API propietaria determine el modelo de identidad. En el entorno público
la autenticación ocurre únicamente después de STARTTLS; Mailpit no se elimina,
porque sigue resolviendo una necesidad distinta: inspección local sin secretos.
El plan gratuito de Resend corta el envío al llegar a 100 correos diarios o
3.000 mensuales, así que es un límite operativo que se monitoriza, no capacidad
garantizada. Véase ADR-0093.

## 2026-08-12 — Una purga no debe retener identidad por preservar historia

Eliminar una cuenta puede chocar con el historial compartido de una liga. La
clave foránea de los cambios de resultado usa `ON DELETE SET NULL`: el resultado
y su cronología continúan disponibles, pero la persona eliminada deja de ser
identificable. El job se ejecuta fuera de la API mediante `launchd`, por lo que
un reinicio de servidor no borra ni duplica su planificación.

## 2026-08-12 — Descripción e indexación son controles independientes

La meta description describe una página si un buscador decide mostrarla, pero no
controla si puede indexarla. Un entorno público de desarrollo conserva
metadatos útiles y envía `X-Robots-Tag: noindex, nofollow, noarchive` desde el
borde para no aparecer en buscadores. Esto no limita el acceso ni sustituye
controles de autenticación.

## 2026-08-11 — El borde puede validarse sin publicar una aplicación incompleta

Caddy puede obtener el certificado y ejercer el redirect canónico antes de que
exista una web o API apta para Internet. Una respuesta `503` explícita evita que
un puerto abierto revele por accidente un servicio interno o plantillas de
asociación sin firmas reales. El proxy solo se añade junto con su artefacto y
configuración verificables.

## 2026-08-11 — Un host de enlaces es una frontera de confianza

Universal Links y App Links no se conceden a una marca, sino a la combinación de
host HTTPS, identificador de aplicación y firma. Separar `fasttourney.com` de
`dev.fasttourney.com` evita que una build de desarrollo herede la asociación y
las credenciales web de producción. Cada host publica únicamente sus propios
ficheros `.well-known`. Véase ADR-0089.

## 2026-08-11 — Un entorno de aprendizaje no tiene que estar encendido siempre

Un entorno AWS enseña sus propiedades reales solo si se crea con IaC, se verifica
y se destruye. Mantenerlo encendido sin tráfico convierte el aprendizaje en coste
fijo. El Mac conserva dos ejecuciones aisladas —desarrollo y release doméstico—,
pero esa separación no le concede alta disponibilidad: comparten equipo, red y
alimentación. Véase ADR-0088.

## 2026-08-11 — Un túnel evita abrir el router, pero no vuelve residencial al ISP

Cloudflare Tunnel abre una conexión saliente desde el Mac y permite que
Cloudflare termine HTTPS sin publicar la IP doméstica como origen ni mantener
forwards/DDNS. Caddy puede seguir en loopback para enrutar por hostname. Esto
reduce mucho la exposición de origen; no impide que una IP residencial ya
conocida pueda recibir tráfico directo que sature su enlace. Véase ADR-0090.

## 2026-08-11 — Un dev público no es el bucle local de Air

Un proyecto Compose aporta namespace a sus contenedores, red y volúmenes, por lo
que `tournaments-manager-dev` puede reutilizar los nombres internos `api` y
`postgres` sin mezclarse con `tournaments-manager-local`. Para que los cambios
diarios no alteren a usuarios externos, el primero usa la imagen `runtime` y web
exportada estática; Air queda en local. Véase ADR-0091.

## 2026-08-11 — DDNS y HTTPS resuelven problemas distintos

Una IP doméstica dinámica exige que el DNS actualice el nombre público, pero no
cifra ni enruta el tráfico. Caddy recibe HTTPS para ese nombre y lo reenvía a la
API interna; en la alternativa inicial el router dirigía TCP 80/443 hacia Caddy.
La decisión actual usa Cloudflare Tunnel (ADR-0090), que elimina esos forwards y
el DDNS, sin publicar PostgreSQL ni el puerto interno de la API.

## 2026-08-10 — Los patrones visuales repetidos necesitan una regla verificable

Una ruta modal no debe recrear a ojo una `X` ni dejar su contenido a ras del
dispositivo: en iOS reutiliza el control nativo de la barra y en web/Android el
botón circular de 44 px con los tokens de superficie y borde. Una `Card` define
su margen horizontal de 20 px, pero una pantalla de superficie plana no debe
envolverse en una solo para obtenerlo: su contenedor desplazable aplica
`paddingHorizontal: space[5]`. Si un diseño elimina la etiqueta visible de un
campo, el placeholder y la etiqueta accesible siguen siendo textos localizados
distintos. Estas reglas se han incorporado a la checklist obligatoria de
`apps/client/AGENTS.md` para que una corrección puntual se convierta en una
comprobación previa reutilizable.

Un `409` solo recibe copy específico cuando el contrato de la operación cambia
la recuperación: la autoasignación de la organizadora es uno de esos casos. La
pantalla mantiene temporalmente ese resultado de búsqueda, a petición de
producto, para validar que la persona recibe el banner antes de decidir filtrarlo
preventivamente.

El fallback genérico no se considera cubierto porque una pantalla llame a
`show()` dentro de un `catch`: un host fuera del stack queda por debajo de un
`fullScreenModal` nativo. El estado del feedback sigue siendo global, pero su
superficie se presenta desde la `Screen` activa en una capa transparente sobre
la escena. Así se coloca siempre tras el área segura global —también sobre una
cabecera— y se descarta para recuperar cualquier control cubierto. La
comprobación manual de un fallo no mapeado al terminar el loader completa la
evidencia en la plataforma que se esté corrigiendo.

## 2026-08-10 — Una mutación ya contiene la actualización de sus vistas

Cuando el contrato devuelve la liga completa tras una mutación, repetir una
consulta para actualizar otra pantalla desperdicia red y permite que las vistas
diverjan. Un almacén reactivo por `leagueId` conserva esa proyección canónica y
notifica únicamente a quienes la presentan. No sustituye una futura sincronía
entre dispositivos: esa actualización seguirá necesitando refresh, revalidación
o tiempo real. Véase ADR-0085.

## 2026-08-10 — Un popup es una superficie compartida, no un modal local

La confirmación de cerrar sesión fija la presentación de los popups: fondo con
blur, oscurecimiento adicional en Android, diálogo centrado y cierre accesible
desde el backdrop. `ModalDialog` conserva ese comportamiento para mensajes
informativos de una sola acción y confirmaciones; cada pantalla aporta solo su
contenido y sus botones. Así una corrección visual o de accesibilidad alcanza
todas las variantes sin copiar un `Modal` ni sus capas.

Los menús de acciones también usan esta superficie en web, iOS y Android. La
cabecera puede conservar el control de apertura nativo de cada plataforma, pero
no sustituye el contenido por un menú de sistema que diverja en estilo o acciones.

## 2026-08-10 — La ayuda contextual no compite con el estado de la pantalla

Las reglas de clasificación se consultan bajo demanda desde la acción de
información: antes de iniciar la liga la pantalla conserva solo el aviso de que
la clasificación todavía no está disponible. Una vez iniciada, muestra la
proyección y las reglas. La condición usa el estado de contrato `published`, no
si la lista de posiciones está vacía. En la barra nativa, el símbolo sigue la
convención de plataforma: `info` sin círculo en iOS y el icono de información
con círculo en Android.

## 2026-08-09 — Un resumen autenticado no repite el onboarding

La home cambia de trabajo al existir sesión: pasa de explicar cómo empezar a
resumir los torneos de la cuenta. Mantener las cards introductorias después de
autenticarse diluye la actividad reciente y repite información que ya no ayuda.
La biblioteca conserva su acción de creación como botón flotante sobre la
botonera; su espacio se reserva en el scroll para mantener accesible el último
torneo. La creación se presenta como modal a pantalla completa en apps y como
página directa en web; su barra ofrece una salida explícita, mientras el
formulario no expone detalles internos sobre cómo persiste el borrador.
Cuando hace falta autenticarse para completar la creación, Cuenta se apila como
otra modal del flujo y, tras iniciar sesión, se vuelve a Crear liga en vez de
cambiar de tab.

## 2026-08-09 — Administrar y crear son relaciones distintas

La colección «Administro» agrupa tanto al creador como a una cuenta delegada.
La UI no debe inferir la propiedad desde esa pestaña: usa la relación explícita
`organizer` para identificar al creador y mantiene el estado de la liga en un
único helper localizado, evitando que una misma transición se describa de forma
distinta según la pantalla.

## 2026-08-09 — La cuenta crea la frontera remota del borrador

Un borrador puede ser local antes de que exista identidad y transferirse de forma
atómica al crear una cuenta. La verificación no crea la liga: habilita la sesión
y la consulta autenticada. Si el borrador ya es una liga válida, persiste como
`published`, sin crear un estado temporal adicional. Véanse ADR-0078 y ADR-0079.

## 2026-08-08 — Paridad útil no significa llevar Air a producción

Un Dockerfile multi-stage puede compartir la resolución de módulos y producir
dos targets con responsabilidades distintas. `dev` incluye Air y observa el
código montado para reducir el ciclo de edición; `runtime` recibe únicamente el
binario estático, certificados y un UID no privilegiado. Compose practica la red
real entre API, PostgreSQL y Mailpit, pero Air, bind mounts y el compilador no
son parte del artefacto que se ejecutará fuera de desarrollo. Separar ambos
targets evita tanto la divergencia completa como copiar herramientas de desarrollo
en producción. Véase ADR-0076.

## 2026-08-09 — La maduración de dependencias necesita una salida de compatibilidad acotada

Una edad mínima reduce exposición para actualizaciones ordinarias, pero no debe
mantener una matriz de runtime que su propio proveedor declara incompatible. La
salida segura no es excluir una familia completa de paquetes: una excepción
exacta y versionada para el conjunto solicitado por Expo mantiene trazabilidad,
permite reconstruir el binario nativo y deja el resto de la política intacto.
La build limpia es imprescindible: typecheck y lockfile no prueban la
compatibilidad entre JavaScript y los componentes nativos montados por iOS.

Véase ADR-0077.

## 2026-08-04 — El lockfile reproduce; la maduración reduce exposición futura

Un lockfile con integridad evita que una instalación ordinaria cambie la
resolución ya revisada, pero no decide qué hacer al incorporar una dependencia
nueva. pnpm puede aplicar una edad mínima de publicación tanto a dependencias
directas como transitivas. Configurar siete días, modo estricto y rechazo de
metadatos sin fecha convierte la política en un fallo visible en vez de una
convención fácil de omitir. No es una garantía de que un paquete antiguo sea
seguro; las actualizaciones siguen requiriendo revisión del diff y respuesta a
advisories.

## 2026-08-04 — Una baja de cuenta no puede dejar una propiedad huérfana

La baja lógica separa retirar de inmediato acceso y relaciones personales de
aplazar la purga física durante una ventana de recuperación. La propiedad de una
liga es distinta: no se borra ni transfiere implícitamente. Si la cuenta aún
organiza una liga, el backend conserva la operación sin cambios y devuelve un
conflicto de negocio que la feature explica con un banner. Así se evita convertir
una eliminación de cuenta en una decisión oculta sobre datos compartidos.

## 2026-08-04 — Expo Go requiere un Metro vivo y seleccionado explícitamente

En esta sesión, un Metro iniciado con `nohup ... &` desde una ejecución efímera
termina sin dejar un servidor utilizable. La vía fiable es mantener
`pnpm --filter @tournaments-manager/client exec expo start --lan` en una
terminal interactiva. Como la presencia de `expo-dev-client` selecciona por
defecto una development build, hay que pulsar `s`, esperar la URL
`exp://<IP-LAN>:8082` que imprime Metro y abrir esa URL exacta en Expo Go. La
URL `com.fasttourney...://expo-development-client` no sirve para Expo Go.

## 2026-08-02 — Preparar no es competir

## 2026-08-02 — Un 401 protegido invalida la sesión, no la feature visible

Una operación protegida que recibe `401` ya no puede recuperar por sí sola: la
credencial local dejó de representar una sesión autorizada. El transporte
protegido delega la invalidación en un único coordinador, que borra los secretos,
resetea la navegación y publica un aviso localizado una sola vez. Las
operaciones públicas conservan su propio `401` —por ejemplo, credenciales de
login incorrectas— y no activan este cierre global.

## 2026-08-02 — Recuperar el foco no equivale a actualizar datos

Las tabs mantienen su contexto durante la sesión; volver a enfocarlas no debe
convertirse en sondeo de red. Cada colección carga una vez por cuenta y conserva
el resultado al alternar tabs. Pull-to-refresh es la acción explícita que vuelve
a consultar la API, mientras un nuevo login reinicia esa marca de carga.

## 2026-08-02 — Actividad reciente no es fecha de creación

Una colección ordenada por UUIDv7 o fecha de creación permite encontrar ligas
nuevas, pero no informa de cambios posteriores. Inicio necesita una proyección
propia, limitada a cinco relaciones y ordenada por una marca del dominio que se
actualiza dentro de cada mutación relevante. Seguir una liga cambia la relación
de una cuenta, no la actividad de la liga; mezclar ambas cosas haría que el
resumen priorizara acciones personales en vez de cambios útiles del torneo.

## 2026-08-02 — La colección principal elige su contexto, no su volumen

Las pestañas «Administro» y «Sigo» comparten una única biblioteca visible para
evitar dos bloques que compiten por atención. Al cargar la colección, se
selecciona «Administro» si contiene alguna liga; solo cuando está vacía se abre
«Sigo», incluso si esta última tiene más elementos. Así la prioridad expresa la
relación de administración acordada y no una regla implícita basada en conteos.
Cada liga conserva su propia card y un estado vacío se centra en el área libre
bajo el selector, para que no parezca otro elemento de la colección.

## 2026-08-02 — Un 401 de login es recuperable, no un fallo genérico

El `401` documentado por `POST /v1/sessions` confirma que el servidor rechazó
las credenciales sin revelar cuál falló. La feature de autenticación lo trata
como un caso recuperable y muestra un mensaje localizado para revisar los datos
introducidos; red, `5xx`, cuerpos inválidos y estados no previstos mantienen el
fallback común seguro. Así no se generaliza el significado de `401` fuera de
esta operación ni se expone el detalle del backend.

## 2026-08-02 — Una mutación iniciada por un efecto necesita un cerrojo síncrono

Un estado como `isConfirming` no evita que React ejecute dos veces un efecto
antes del siguiente render, algo que el entorno de desarrollo hace
intencionadamente. La confirmación de registro conserva el token ya iniciado
en un `ref` antes de llamar a la API: así emite un único `POST` por token y no
puede mostrar el `409` de su propia repetición tras haber creado la sesión.

## 2026-08-02 — El destino tras iniciar sesión pertenece al recorrido que lo inició

Un reemplazo de sesión no siempre implica ir a Inicio. La verificación mediante
enlace y el restablecimiento conservan su regreso aceptado a `/`, mientras que
un login local o con Google iniciado desde Cuenta vuelve a `/account`. El
proveedor de sesión recibe ese destino de forma explícita y la navegación común
lo aplica en las tres plataformas, sin duplicar condiciones por web, iOS o
Android.

## 2026-08-02 — Un blur no se debe tapar con la capa que debía complementar

El material nativo de `BlurView` ya aporta una tinta visual. Superponer además
el lienzo del tema al 68 % lo hacía casi opaco: se perdía el desenfoque y el
color de fondo dominaba la pantalla. Además, un `Modal` de React Native vive en
otra ventana y su `BlurView` no puede muestrear los píxeles de la ruta previa.
El estado de confirmaciones se comparte, pero su host se monta en la `Screen`
activa: así el `Modal` pertenece a la ruta presentada y no queda detrás de un
`fullScreenModal` ni de la tab bar. En un scrim de pantalla completa, el blur
oscuro clásico de iOS es más predecible que los materiales dinámicos: estos
últimos incorporan una tinta del sistema que puede dominar la superficie.
Android conserva una atenuación neutra y leve como respaldo. Así el contexto se
percibe sin competir con el diálogo.

La separación de un diálogo no debe depender de que un tema tenga menos
contraste. Un borde con el token semántico `border.default` conserva la misma
intención en claro y oscuro, y deja que cada tema resuelva el valor adecuado sin
crear un parche visual exclusivo para la variante oscura.

## 2026-08-02 — El fondo de una confirmación es una salida explícita

El área exterior del diálogo no es decorativa: tocarla ejecuta la misma acción
que «Cancelar». La primitiva la expone además como un botón accesible con esa
misma etiqueta, para que la interacción visual y la semántica no diverjan.

## 2026-08-02 — El cierre de un popup OAuth web no pertenece al arranque nativo

`maybeCompleteAuthSession()` resuelve el popup de OAuth al volver a una página
web. En iOS y Android la sesión es nativa; invocar ese puente al recargar el
bundle no aporta valor y puede interferir con el navegador del sistema. Por eso
la operación queda limitada a web, mientras los flujos nativos conservan su
redirect URI propio.

## 2026-08-02 — Un enlace recibido siempre se normaliza antes de navegar

Expo puede entregar un enlace a una app ya viva también durante un reload del
bundle. Si una URL HTTPS se pasa sin normalizar al router, este puede delegarla
en Safari como destino externo. `+native-intent` la convierte siempre en una
ruta interna; solo el enlace de arranque se difiere hasta que Inicio está
montado.

Las URL protocol-relative (`//dominio/ruta`) son otro formato externo: empiezan
por `/`, pero Expo Router las traduce a `https://…`. Por eso el normalizador las
interpreta como URL antes de aceptar una cadena como ruta interna.

Un borrador local reduce fricción sin convertir datos temporales en recursos del
servidor. Una liga creada ya tiene organizador y ciclo de vida; mostrarla como
«Sin empezar» evita confundirla con aquel borrador. Generar el calendario al
iniciar permite elegir ida o vuelta en el límite donde equipos y reglas se
congelan. La deuda aceptada es no recuperar un borrador desde otro dispositivo.
Véanse ADR-0069 y ADR-0070.

## 2026-08-02 — Una biblioteca clasifica relaciones, no permisos

«Administro» y «Sigo» son colecciones distintas que el backend ya autoriza y
pagina. El cliente las presenta sin inferir permisos desde la UI; al abrir una
liga, la operación concreta vuelve a comprobar su autorización. Véanse
ADR-0057 y ADR-0058.

## 2026-08-02 — La integración necesita la misma base que falla en producción

Una restricción SQL que impedía ida y vuelta no apareció en el caso de uso con
dobles; la detectó la primera prueba contra PostgreSQL. CI levanta ahora una
base efímera para repetir esa evidencia sin apuntar a datos locales. Véase
ADR-0071.

## 2026-08-01 — El feedback inmediato debe resolver una necesidad concreta

La validación al escribir no se aplica por defecto: puede distraer y revelar un
error antes de que la persona termine un campo. `TextField` admite el disparador
`change` para requisitos progresivos, como los ocho caracteres de una contraseña:
mientras se escribe muestra el mínimo pendiente y, al cumplirlo, lo sustituye
por el indicador de fuerza. Los demás campos conservan la validación al perder
el foco y al enviar.

## 2026-08-01 — La interacción de un campo no es la del formulario

La validación de formato se muestra al abandonar el campo correspondiente, sin
revelar errores en campos que la persona todavía no ha visitado. El intento de
envío sí muestra todos los errores pendientes. Por ello cada control conserva
su propia marca de interacción y el formulario una marca distinta de intento de
envío; usar una única marca global en ambos eventos produce feedback prematuro.

## 2026-08-01 — OAuth tras una reautenticación

Un popup OAuth web debe abrirse desde un gesto vigente. Por eso, después de
validar una contraseña se presenta una acción explícita «Continuar con Google»:
el challenge se crea solo tras la reautenticación y el popup no depende de que
el navegador conserve el gesto a través de una petición de red. El modal se
descarta al terminar y Cuenta recupera el estado desde la API.

## Método

Para cada capacidad se sigue el ciclo:

1. explicar el fundamento con palabras propias;
2. comparar alternativas;
3. construir la versión mínima;
4. observar su comportamiento;
5. provocar y diagnosticar un fallo;
6. documentar lo aprendido y los límites;
7. revisarlo en la retrospectiva de fase.

## Mapa de competencias

| Área           | Resultado demostrable                             | Estado                                      |
| -------------- | ------------------------------------------------- | ------------------------------------------- |
| Arquitectura   | Explicar límites, dependencias y trade-offs       | En curso                                    |
| Go             | Construir y mantener un servicio idiomático       | No iniciado                                 |
| PostgreSQL     | Diseñar, migrar y operar datos con criterio       | No iniciado                                 |
| API            | Diseñar contratos evolutivos y observables        | No iniciado                                 |
| Testing        | Elegir pruebas por riesgo y velocidad de feedback | Fundamentos aceptados                       |
| Seguridad      | Modelar amenazas y aplicar controles verificables | No iniciado                                 |
| Contenedores   | Crear un entorno local reproducible               | No iniciado                                 |
| Observabilidad | Diagnosticar con logs, métricas y trazas          | No iniciado                                 |
| Kubernetes     | Desplegar, escalar y recuperar cargas             | No iniciado                                 |
| Terraform/AWS  | Aprovisionar y operar infraestructura             | Fundamentos IaC, cuentas y estado aceptados |

## Diario

### 2026-08-01 — El foco compartido pertenece al perímetro del control

- **Aprendido:** un campo compuesto por un contenedor y un `TextInput` debe
  representar el foco en el contenedor, que es la caja que la persona percibe
  como control, no en el input interno.
- **Evidencia:** React Native Web conserva el `outline` del elemento HTML si no
  se desactiva explícitamente; al mismo tiempo, el estado de foco de la primitiva
  aplica `border.focus` al contenedor en web, iOS y Android.
- **Coste aceptado:** se conserva un pequeño estado local en `TextField` para
  delegar la apariencia al componente compartido y evitar CSS específico por ruta.

### 2026-08-01 — Un degradado universal debe describir una geometría común

- **Aprendido:** los puntos de inicio y final de un degradado no tienen idéntica
  semántica en CSS y en los renderizadores nativos. Un punto desplazado puede
  cambiar solo el ángulo en web y desplazar realmente el recorrido en iOS y
  Android.
- **Evidencia:** `expo-linear-gradient` convierte web a un `linear-gradient()`
  CSS, mientras que iOS y Android dibujan el vector entre los puntos. Los tokens
  de icono y botón usan ahora el mismo vector de esquina superior izquierda a
  inferior derecha y las mismas paradas; el email replica la paleta azul
  `#155EEF` y violeta `#7F56D9`, con azul sólido como fallback.
- **Coste aceptado:** el backend no importa tokens TypeScript. La pequeña
  duplicación de los valores CSS queda documentada y cubierta por la prueba del
  email, evitando acoplar Go a la infraestructura del cliente.

### 2026-08-01 — El nonce de una sesión externa debe venir de nuestra API

- **Aprendido:** el cliente no genera ni interpreta la prueba de identidad de
  Google. Solicita un challenge de un solo uso, transmite su nonce al proveedor
  y entrega únicamente el ID token resultante a la API.
- **Evidencia:** la feature `federated-google` usa las operaciones OpenAPI con
  `apiFetch`; `expo-auth-session` recibe el nonce como parámetro de autorización
  y el backend valida después issuer, audience, expiración y nonce.
- **Coste aceptado:** la creación de un cliente Android necesita el SHA-1 del
  certificado con que se firme la build. No se sustituye por la huella de iOS ni
  se publica una app Android sin esa restricción.

### 2026-08-01 — Un popup de identidad exige preparar su prueba antes del gesto

- **Aprendido:** en web, esperar una petición antes de abrir la autenticación
  puede hacer que el navegador bloquee el popup. El challenge se prepara al
  llegar a Cuenta; si no está disponible, el toque solo reintenta esa
  preparación y el siguiente abre Google.
- **Evidencia:** `useGoogleAuthentication` separa `isPreparing` de la
  autenticación ya abierta. La pantalla reemplaza el icono por progreso durante
  ambos estados; Google controla su propia superficie, por lo que Cuenta no
  añade una barrera de interacción redundante.
- **Coste aceptado:** tras un fallo o una caducidad poco frecuentes puede hacer
  falta un segundo toque. Se evita abrir ventanas vacías o automatizar un popup
  fuera del gesto, que sería menos fiable y menos accesible.

### 2026-08-01 — Bloquear depende del compromiso de la operación

- **Aprendido:** un loader no basta para proteger una operación propia ya
  enviada. Registro usa `InteractionBlocker` desde el `POST` hasta su respuesta
  para impedir navegación, edición o nuevos envíos durante ese intervalo.
- **Evidencia:** la validación local y las consultas de disponibilidad no
  bloquean la ruta; el bloqueo comienza exclusivamente con `isSubmitting`.
- **Coste aceptado:** la barrera es transparente y anuncia progreso al lector
  de pantalla. No se aplica a una precarga reversible de Google, donde la
  persona puede cambiar libremente a otro método de acceso.

### 2026-08-02 — El formato condiciona el feedback de disponibilidad

- **Aprendido:** el username del alta valida en cada cambio porque solo los
  valores que cumplen la regex pueden consultar disponibilidad. Si deja de
  cumplirla, se muestra de inmediato el mismo error de formato que aparecería
  al perder foco; al corregirlo, el campo recupera su estado de comprobación y,
  después, su disponibilidad.
- **Evidencia:** la ruta `account/register` usa `validationTrigger="change"`
  exclusivamente en el campo username. `useUsernameAvailability` continúa
  cancelando la consulta previa y vuelve a comprobar únicamente entradas
  válidas tras 400 ms.
- **Coste aceptado:** el feedback de formato aparece desde la primera edición
  del campo. Es una excepción acotada al patrón de disponibilidad; email y los
  demás campos conservan validación al abandonar el control.

### 2026-08-02 — La sesión web se restaura preguntando al backend

- **Aprendido:** una cookie `HttpOnly` no puede convertirse en estado visible
  leyendo almacenamiento local. Al arrancar la web consulta `GET /v1/sessions`:
  una sesión válida devuelve su identidad y vigencia; `401` solo representa una
  visita anónima y no muestra feedback.
- **Evidencia:** el middleware conserva la credencial únicamente en el contexto
  de la petición y el repositorio vuelve a validar su hash antes de devolver la
  proyección `CurrentSession`. El provider web restaura solo un resultado `200`,
  por lo que una comprobación iniciada antes de confirmar un enlace no borra la
  nueva identidad.
- **Coste aceptado:** web hace una consulta protegida al inicio. Es necesaria
  para recuperar una cookie que JavaScript no puede leer y no cambia la
  estrategia móvil, que restaura su estado desde Keychain/Keystore.

### 2026-07-30 — Un recorrido vertical hace visibles los límites

- **Aprendido:** una ruta HTTP no es toda la API ni toda la lógica del backend.
  El handler traduce transporte; el servicio coordina el caso de uso; el
  repositorio adapta el puerto a PostgreSQL; `sqlc` genera la llamada Go a partir
  del SQL escrito por el equipo.
- **Evidencia:** el recorrido de `GET /v1/usernames/{username}/availability` en
  `apps/backend/README.md` enlaza `server.go`, `registration.Service`,
  `RegistrationRepository`, `db/queries/registrations.sql` y la salida generada
  por `sqlc`.
- **Coste aceptado:** se mantiene una guía local al backend, enlazada desde el
  README raíz y Arquitectura, en vez de repetir el detalle en todos los
  documentos normativos.

### 2026-07-29 — Una comprobación de disponibilidad mejora la UX, no la unicidad

- **Aprendido:** una respuesta de disponibilidad solo describe un instante: dos
  clientes pueden recibir «disponible» y competir después por el mismo nombre.
  Por ello PostgreSQL mantiene la restricción única y el alta es la única
  autoridad que decide definitivamente.
- **Evidencia:** `GET /v1/usernames/{username}/availability` solo acepta el
  formato del contrato, responde `Cache-Control: no-store` y el cliente espera
  400 ms desde la última escritura antes de llamarlo; una nueva entrada cancela
  la petición previa.
- **Coste aceptado:** el límite de 30 consultas por IP y minuto vive en memoria
  del proceso para mantener este corte simple. No protege de forma global entre
  réplicas; si se escala horizontalmente, se decidirá un control compartido en
  el borde o almacenamiento adecuado.

### 2026-07-29 — CORS es una política de frontera, no una cabecera por ruta

- **Aprendido:** CORS autoriza al JavaScript de un origen a leer una respuesta;
  no sustituye autenticación, autorización ni CSRF. Un `POST` JSON inicia además
  un preflight `OPTIONS` antes de la petición real.
- **Evidencia:** la API valida `CORS_ALLOWED_ORIGINS` al arrancar y aplica una
  allowlist exacta en el handler raíz. Devuelve el origen permitido, `Vary: Origin`
  y credenciales; responde el preflight solo para métodos y cabeceras
  que el cliente necesita.
- **Coste aceptado:** cada entorno debe declarar sus orígenes web. Se evita el
  arranque cómodo con `*` porque sería incompatible con las cookies de sesión y
  ocultaría qué frontends tienen permiso para consumir la API.

### 2026-08-08 — CORS y CSRF deben compartir la misma confianza explícita

- **Aprendido:** permitir un origen en CORS no hace que una protección CSRF lo
  acepte automáticamente. Una mutación web autenticada por cookie necesita que
  ambos controles conozcan el mismo origen confiable.
- **Evidencia:** el borrado de cuenta desde `http://localhost:8081` alcanzaba
  la API, pero recibía `403` de la protección CSRF pese a estar en
  `CORS_ALLOWED_ORIGINS`. La API registra ahora cada origen ya validado también
  como origen confiable del control CSRF y una prueba HTTP cubre ese flujo.
- **Regla reutilizable:** CORS y CSRF conservan responsabilidades distintas;
  cuando se permite una web con cookies, su allowlist validada debe alimentar
  explícitamente ambos controles, sin convertir CORS en sustituto de CSRF.

### 2026-08-08 — Invalidar una sesión también reinicia el estado de navegación

- **Aprendido:** eliminar los secretos y el usuario en memoria no basta si el
  contenedor de tabs conserva rutas profundas o vistas ya montadas de la sesión
  anterior.
- **Evidencia:** tras programar la eliminación, Inicio quedaba seleccionado
  pero la ruta de datos de acceso podía conservarse hasta recargar. El reset
  vacía la pila activa solo cuando puede cerrarse y reemplaza la ruta por el
  destino anónimo; las tabs se remontan con la revisión de sesión que ya se
  incrementa al cerrar sesión, caducar una credencial o borrar la cuenta.
- **Regla reutilizable:** toda transición a sesión anónima reinicia a la vez la
  identidad, las rutas profundas y los flujos de navegación dependientes de
  sesión; no debe depender de una recarga del navegador.

### 2026-07-29 — La experiencia de usuario y el diagnóstico técnico requieren señales distintas

- **Aprendido:** un replay explica la secuencia de interacción que precede a un
  error; una traza OpenTelemetry explica qué ocurrió en HTTP, lógica técnica y
  PostgreSQL. Ninguna de las dos señales sustituye a la otra, y se cruzan con un
  `request_id` opaco en vez de exponer identidad, payloads o PII.
- **Evidencia:** ADR-0060 aplaza PostHog Cloud hasta la primera beta distribuida,
  limita su alcance a experiencia de cliente y conserva el stack OpenTelemetry
  de ADR-0020 para backend. El proveedor no recibe autoridad de negocio.
- **Coste aceptado:** replay, analytics y error tracking quedan en un SaaS
  externo con región UE, gasto máximo de 0 € y revisión de privacidad previa;
  no se autoaloja una plataforma de producto sin volumen que lo justifique.

### 2026-07-29 — El espacio de una tab superpuesta es parte del scroll

- **Aprendido:** reservar el área de la barra de tabs como padding de la pantalla
  acorta visualmente toda la vista. La superficie debe alcanzar el borde inferior
  y el margen de seguridad debe pertenecer al contenedor desplazable.
- **Evidencia:** `Screen` admite omitir su inset inferior en rutas bajo
  `NativeTabs`; Inicio, Torneos y las rutas de Cuenta lo hacen. Los `ScrollView`
  de esos flujos usan `useTabContentBottomPadding`, de modo que el último
  control se puede llevar por encima de la barra nativa. En web, el cálculo
  añade `space[10]` (40 px) porque su safe-area inferior es cero y la botonera
  estándar permanece superpuesta; se aplica a cualquier ruta bajo tabs, no solo
  a Inicio.
- **Coste aceptado:** no se introduce una segunda abstracción de layout ni se
  fija una altura de tab bar, que variaría según plataforma y versión del sistema.

### 2026-07-29 — El feedback global también pertenece al área segura

- **Aprendido:** un aviso superpuesto debe calcular su posición desde el inset
  seguro, no desde un margen fijo; de otro modo, en un iPhone con notch puede
  quedar oculto detrás del sistema.
- **Evidencia:** `FeedbackProvider` posiciona el banner tras el inset superior,
  adopta el radio y margen lateral de una card y anima entrada y salida con el
  token compartido `motion.enterExit`. También identifica cada aviso activo,
  cancela su temporizador y sustituye de forma segura el anterior cuando llega
  un mensaje nuevo. El toque y un arrastre vertical hacia arriba descartan el
  aviso; los arrastres cortos se recuperan con `motion.feedback`.
- **Coste aceptado:** la transición se omite cuando el sistema solicita
  movimiento reducido; no se añade una librería ni una jerarquía nueva para un
  único aviso global.

### 2026-07-28 — El degradado conserva jerarquía cuando no colorea todo

- **Aprendido:** aplicar el degradado de marca a la acción primary filled crea
  una relación directa con el icono; limitarlo al borde de 1 px en la acción
  secondary conserva contraste y evita competir con la acción principal.
- **Evidencia:** `Button` parte de azul sólido y superpone el token
  `gradient.brandButton`. El texto sobre esa marca usa un token blanco
  independiente del tema, no el color inverso de la superficie de la app.
- **Coste aceptado:** las acciones destructivas siguen usando rojo sólido, pues
  el degradado de marca no debe diluir su semántica de riesgo.

### 2026-07-29 — El degradado se ajusta desde el token, no desde cada botón

- **Aprendido:** adelantar la entrada del violeta al 35 % del recorrido aumenta
  su presencia sin cambiar el ángulo ni convertir toda la acción en una superficie
  violeta.
- **Evidencia:** `gradient.brand` y `gradient.brandButton` comparten las mismas
  paradas y una geometría diagonal completa, común a web, iOS y Android.
- **Coste aceptado:** la barra de tabs nativa mantiene un tint azul sólido, porque
  su API no admite un degradado uniforme para icono y etiqueta.

### 2026-07-29 — Semántica de indicadores

- **Aprendido:** el color principal comunica una acción disponible. Un número de
  paso estático debe usar el color de texto y un borde, no un fondo de marca, para
  no adquirir una apariencia pulsable.
- **Evidencia:** los pasos orientativos de la home consumen `colors.text.primary`
  tanto para el número como para el borde circular, y se adaptan al tema claro u
  oscuro sin usar blanco o negro puros.

### 2026-07-29 — Un botón es una acción, no una caja

- **Aprendido:** un radio de 12 px encaja en campos, pero un botón de 44 px de
  alto necesita extremos semicirculares para expresar mejor que es una acción.
- **Evidencia:** la primitiva `Button` usa el token `radius.pill` en todas sus
  variantes; el botón secondary mantiene un borde azul de marca y deja
  transparente su interior para revelar la superficie que hay detrás.
- **Coste aceptado:** el contraste del borde y del texto se evalúa contra la
  superficie contenedora; por ello las pantallas siguen usando tokens de fondo
  semánticos y no colores arbitrarios.

### 2026-07-29 — Las acciones de proveedor caben en su identidad visual

- **Aprendido:** cuando una acción de proveedor se entiende por su marca, un
  botón solo con el icono reduce ruido visual sin perder accesibilidad.
- **Evidencia:** la acción de Google en Cuenta es circular, está centrada en su
  card, conserva un objetivo de 48 px y expone su estado y etiqueta localizada
  al lector de pantalla.
- **Coste aceptado:** el PNG oficial se versiona como asset local para no añadir
  una dependencia de iconos ni depender de red durante el renderizado.

### 2026-07-29 — La navegación persistente no se duplica como contenido

- **Aprendido:** una card que solo recuerda que existe una sección permanente no
  mejora la orientación. La home conserva contenido de bienvenida y la tab
  Torneos mantiene la responsabilidad de acceso a la biblioteca.
- **Evidencia:** se retiró el bloque informativo final de la home; Inicio mantiene
  la acción principal, el contexto y los pasos orientativos.

### 2026-07-28 — Dos apps no obligan a mantener dos proyectos Xcode

- **Aprendido:** Expo CNG permite resolver nombre, esquema, icono y futuras
  diferencias nativas desde una configuración por entorno. Así desarrollo y
  producción se pueden instalar como apps distintas sin convertir directorios
  generados en fuente versionada.
- **Evidencia:** `apps/client/app.config.ts` selecciona `Fast Tourney Dev` o
  `Fast Tourney` con `APP_ENV` y usa el icono común de 1024 × 1024.
- **Coste aceptado:** los bundle identifiers y perfiles de distribución esperan
  a disponer de un dominio controlado. Se evaluarán targets iOS separados solo
  si un config plugin no puede aislar una dependencia nativa necesaria.

### 2026-07-28 — Una preferencia local no es un permiso del sistema

- **Aprendido:** el tema es una preferencia de presentación y puede persistirse
  localmente para cualquier visitante; el permiso de notificaciones depende del
  sistema operativo, requiere una integración nativa y no se puede representar
  como un switch local que aparente concederlo.
- **Evidencia:** `PreferencesProvider` resuelve y persiste `system`, `light` y
  `dark` para todas las rutas, y la raíz entrega el tema correspondiente a React
  Navigation para que las transiciones nativas no usen temporalmente el tema
  contrario. El ajuste de notificaciones permanece deshabilitado y explica su
  límite, coherente con el alcance de `PRODUCT.md`.
- **Coste aceptado:** no se incorpora aún `expo-notifications`, configuración
  nativa ni almacenamiento de preferencias de entrega. Requerirán una decisión
  de producto y el diseño del ciclo de permiso antes de activarlos.

### 2026-07-28 — Un formulario visible no equivale a una sesión inventada

- **Aprendido:** una interfaz de acceso puede preparar el recorrido local y la
  ruta de alta sin afirmar que el cliente ya establece una sesión cuando el
  handler y el adaptador HTTP todavía no existen.
- **Evidencia:** Cuenta ofrece correo, contraseña y registro en una ruta
  profunda; las acciones comunican su indisponibilidad actual y Google se marca
  como próximo, conforme al contrato y ADR-0050.
- **Coste aceptado:** conectar el formulario exige implementar el adaptador de
  sesión y el flujo OIDC real; los controles actuales no envían credenciales.

### 2026-07-28 — El tipado no sustituye un bundle de Expo

- **Aprendido:** TypeScript no carga Metro ni resuelve el árbol de rutas en la
  misma forma que Expo. La exportación web detecta errores de bundling y router
  sin exigir un navegador ni un artefacto persistente.
- **Evidencia:** `make client-web-export` escribe temporalmente en `/tmp` y
  `make verify` lo ejecuta junto al resto de la puerta de calidad.
- **Coste aceptado:** no se añade todavía una prueba visual automatizada ni un
  artefacto de CI; se abrirán cuando exista un flujo crítico que los justifique.

### 2026-07-28 — Una tab es un límite de navegación, no un botón de pantalla

- **Aprendido:** Inicio, Torneos y Cuenta representan flujos de primer nivel;
  ubicarlos en la botonera evita duplicar accesos dentro de cada pantalla. Cuenta
  conserva un stack para que su evolución no altere el historial de Inicio.
- **Evidencia:** grupo `(tabs)` de Expo Router, con Inicio en la primera
  posición, y `NativeTabs` delegando el acabado Liquid Glass a iOS 26. En web,
  el cierre de una ruta de Cuenta reemplaza explícitamente `/account`: tras una
  recarga, el historial del navegador puede pertenecer a otra tab.
- **Coste aceptado:** Torneos y Cuenta solo muestran su contexto hasta que sus
  flujos autenticados y colecciones tengan datos reales.

### 2026-07-28 — La jerarquía también necesita límites espaciales

- **Aprendido:** el padding de pantalla protege el contenido de bordes y zonas
  del sistema; las cards delimitan bloques con significado. Usar ambos evita
  una pantalla plana sin convertir cada elemento en una tarjeta.
- **Evidencia:** `Screen` reserva 24 px de cabecera y `shared/ui/Card` añade
  20 px de margen exterior horizontal al agrupar los cuatro bloques de la home.
- **Coste aceptado:** no se añaden sombras ni variantes antes de que una lista
  o interacción demuestre necesitar un estado visual adicional.

### 2026-07-28 — Un token visual no localiza el producto

- **Aprendido:** los tokens hacen consistente la presentación, pero no eliminan
  la necesidad de catalogar el copy. Dejar textos en una ruta impide aplicar el
  locale detectado y hace que cada nueva traducción sea una búsqueda manual.
- **Evidencia:** extracción del copy de la home a un catálogo `es`, `en`, `it`
  y `fr`, con fallback explícito a inglés conforme a ADR-0056.
- **Coste aceptado:** el selector web persistente se implementará con el
  provider de preferencias compartido, antes de añadir una segunda pantalla.
- **Siguiente decisión:** incorporar ese provider de idioma y tema a la raíz
  del cliente sin duplicar estado por feature.

### 2026-07-28 — Un locale es una unidad de intercambio

- **Aprendido:** un archivo plano por locale permite que una plataforma de
  traducción importe y exporte el producto completo sin reconstruir textos
  dispersos. Los prefijos semánticos reúnen las claves comunes y de cada
  capacidad sin imponer objetos anidados al formato de intercambio.
- **Evidencia:** `shared/i18n/locales/{es,en,it,fr}.json` y la validación
  TypeScript de que cada catálogo tiene las claves del inglés.
- **Coste aceptado:** la preferencia persistente de web sigue pendiente del
  provider compartido; los JSON no implementan estado por sí mismos.

### 2026-07-28 — Una home inicial orienta sin inventar estado

- **Aprendido:** una home de invitado no necesita simular una sesión ni una
  biblioteca vacía: la siguiente acción principal y una explicación breve
  orientan sin presentar datos personalizados inexistentes.
- **Evidencia:** implementación de `/` sobre las primitivas Pulse y la regla de
  ADR-0057 que reserva «Administro» y «Guardados» para una sesión verificada.
- **Coste aceptado:** los destinos de crear torneo y cuenta continúan en cortes
  propios; sus acciones muestran una respuesta explícita en vez de aparentar
  que el flujo ya está disponible.
- **Siguiente decisión:** construir el flujo de borrador local y la pantalla de
  cuenta sobre las rutas canónicas acordadas.

### 2026-07-28 — Autenticar no es autorizar

- **Aprendido:** validar una cookie o Bearer identifica la cuenta y puede ser
  transversal; decidir qué puede hacer esa cuenta dentro de una liga requiere
  las reglas del caso de uso. Mezclar ambas cosas en middleware ocultaría la
  política de negocio.
- **Evidencia:** ADR-0059 y la ruta protegida `GET /me/leagues`.
- **Coste aceptado:** CSRF sigue como protección independiente para la primera
  mutación web autenticada; no se añade un token sin una operación que proteger.
- **Siguiente decisión:** diseñar esa protección CSRF junto con la primera
  mutación autenticada por cookie.

### 2026-07-28 — La home reutiliza una colección; no necesita una API propia

- **Aprendido:** la home puede pedir las primeras páginas de la misma colección
  autenticada que usa la biblioteca. Separar `administered` y `followed` en el
  servidor conserva la autorización y evita que el cliente cargue o clasifique
  relaciones ajenas.
- **Evidencia:** ADR-0058; tablas de administradores y seguidores con claves
  compuestas y `GET /me/leagues` paginado por UUIDv7.
- **Coste aceptado:** la home puede realizar dos lecturas pequeñas; no se añade
  agregación, caché global ni búsqueda antes de medir la necesidad.
- **Siguiente decisión:** implementar las mutaciones para seguir una liga y
  asignar o abandonar administración delegada antes de esperar datos de usuario.

### 2026-07-28 — Una colección personal no es una regla de autorización

- **Aprendido:** «Administro» y «Sigo» son vistas útiles de relaciones distintas
  con una liga. La primera reúne propiedad y administración delegada; la segunda
  expresa seguimiento. Ninguna clasificación en el cliente concede permisos.
- **Evidencia:** ADR-0034 y ADR-0057.
- **Coste aceptado:** hace falta una proyección autenticada adicional antes de
  mostrar colecciones reales; no se añade una caché global ni persistencia de
  navegación entre reinicios.
- **Siguiente decisión:** definir el recurso OpenAPI autenticado para las ligas
  relacionadas con la cuenta, con orden y paginación.

### 2026-07-27 — Una mitigación de tooling necesita fecha de caducidad

- **Aprendido:** con Go 1.26.5 y `golangci-lint` 2.12.2, incluir paquetes de
  prueba en el análisis puede terminar con `no go files to analyze`, aunque
  `go test ./...` los cargue y ejecute correctamente. No se ha confirmado un
  issue upstream que documente exactamente esta combinación.
- **Evidencia:** reproducción local y primera CI de la promoción de Fase 1;
  `make verify` pasa con `run.tests: false` y conserva `go test ./...` como
  comprobación de los tests.
- **Coste aceptado:** los linters no inspeccionan temporalmente `*_test.go`;
  las pruebas siguen compilándose y ejecutándose en local y CI.
- **Siguiente decisión:** al actualizar Go o `golangci-lint`, reactivar
  `run.tests: true`, ejecutar `make verify` y retirar la mitigación si pasa.

### 2026-07-27 — Una identidad externa no es una sesión

- **Aprendido:** Google acredita una identidad externa mediante un `subject`
  estable; TournamentsManager la vincula a una cuenta interna y emite después su
  propia sesión opaca. La contraseña no interviene en ese recorrido.
- **Evidencia:** ADR-0050 y la referencia de OpenID Connect de Google.
- **Coste aceptado:** el primer incremento debe configurar y validar Google,
  incluida la vinculación explícita de una cuenta local coincidente.
- **Siguiente decisión:** concretar el artefacto OIDC, nonce, CSRF y contrato
  HTTP antes de implementar el adaptador Google.

### 2026-07-26 — Renovar una sesión no exige JWT

- **Aprendido:** una credencial opaca puede rotarse silenciosamente y conservar
  revocación inmediata porque el servidor mantiene su estado; JWT aporta
  validación distribuida, no más comodidad intrínseca para la persona usuaria.
- **Evidencia:** ADR-0044 y comparación con sesiones de PostgreSQL.
- **Coste aceptado:** cada petición autenticada valida una sesión en PostgreSQL;
  no se adelantan servicios ni tokens de acceso distribuidos.
- **Siguiente decisión:** modelo de datos y contrato HTTP del registro,
  verificación, sesión y publicación.

### 2026-07-26 — Un vertical slice no necesita contener todo el dominio

- **Aprendido:** el primer recorrido útil debe cruzar identidad, autorización,
  datos y API, pero puede detenerse antes de las reglas deportivas avanzadas si
  estas no cambian la evidencia buscada.
- **Evidencia:** ADR-0043.
- **Coste aceptado:** inicio, resultados, bajas, cierre, cancelación y
  administración delegada esperan a un incremento posterior.
- **Siguiente decisión:** analizar sesiones, verificación local y modelo de
  datos antes de implementar el backend.

### 2026-07-26 — La persistencia es una propiedad del volumen, no del contenedor

- **Aprendido:** detener y recrear el contenedor no borra una base PostgreSQL si
  el volumen nombrado se conserva; borrar esos datos debe seguir siendo una
  operación explícita y confirmada.
- **Evidencia:** arranque saludable, reinicio y lectura posterior de una marca
  temporal mediante el runbook de PostgreSQL local.
- **Coste aceptado:** el reset solo es seguro cuando los datos locales son
  prescindibles; la confirmación explícita evita convertirlo en un efecto
  accidental de otro comando.
- **Siguiente decisión:** crear el esquema y las primeras migraciones al abrir
  el trabajo de backend de la Fase 2.

### 2026-07-26 — Un estado visible puede preceder a una notificación

- **Aprendido:** conservar una liga cancelada y mostrar su estado por enlace o
  en «ligas seguidas» comunica la verdad del recurso sin introducir aún entrega,
  preferencias y fallos de email o push.
- **Evidencia:** ADR-0042.
- **Coste aceptado:** las personas no reciben aviso proactivo ni explicación de
  la cancelación.
- **Siguiente decisión:** Gate 0B completado; preparar el entorno local antes de
  implementar el vertical slice.

### 2026-07-26 — La igualdad puede requerir reescribir un marcador

- **Aprendido:** conservar algunos resultados y otorgar otros tras una baja crea
  una desigualdad por el orden de calendario. Uniformar todos a `3-0` mantiene la
  liga, pero exige historial porque puede reemplazar datos reales.
- **Evidencia:** ADR-0041.
- **Coste aceptado:** todavía no hay motivos tipados de baja ni reincorporación.
- **Siguiente decisión:** completada en ADR-0042; Gate 0B cerrado.

### 2026-07-26 — Compartir no equivale a congelar

- **Aprendido:** publicar permite compartir y terminar de preparar una liga; el
  límite seguro para generar partidos y congelar equipos es el inicio. Así se
  evita replanificar una competición con resultados.
- **Evidencia:** ADR-0040, sucesor parcial de ADR-0032.
- **Coste aceptado:** una liga publicada aún no tiene partidos y puede quedarse
  sin iniciar hasta que exista una política de recordatorios o limpieza.
- **Siguiente decisión:** completada en ADR-0041; continúa motivo y comunicación
  de una cancelación.

### 2026-07-26 — Finalizar es una revisión, no un efecto colateral

- **Aprendido:** exigir todos los marcadores y una acción explícita del creador
  evita clasificaciones incompletas y deja una oportunidad final para detectar un
  error. Un `3-0` administrativo se representa como marcador ordinario para no
  crear incidencias que el dominio todavía no sabe interpretar.
- **Evidencia:** ADR-0039.
- **Coste aceptado:** se pierde el motivo de un resultado excepcional y no se
  puede corregir una liga finalizada en el primer corte.
- **Siguiente decisión:** cancelación completada en ADR-0040; continúa la baja
  de equipos.

### 2026-07-26 — Publicar prepara; iniciar habilita la competición

- **Aprendido:** separar publicación e inicio evita que un resultado transforme
  accidentalmente una liga que solo estaba lista para compartir. La transición
  explícita conserva al creador el control de cuándo empieza el juego.
- **Evidencia:** ADR-0038.
- **Coste aceptado:** no hay inicio automático, calendario ni cambios de reglas
  o equipos después de publicar.
- **Siguiente decisión:** completada en ADR-0039; continúan cancelación y bajas
  de equipos.

### 2026-07-26 — Un marcador útil no necesita modelar todas las incidencias

- **Aprendido:** dos goles enteros no negativos bastan para expresar victoria,
  derrota y empate en una liga con puntuación 3-1-0. Añadir excepciones sin sus
  reglas de clasificación produciría campos que el sistema no sabe interpretar.
- **Evidencia:** ADR-0037.
- **Coste aceptado:** no pueden registrarse aún penaltis, prórrogas,
  incomparecencias ni sanciones.
- **Siguiente decisión:** completada en ADR-0038; continúan finalización y bajas
  de equipos.

### 2026-07-26 — Corregir rápido no implica borrar el pasado

- **Aprendido:** permitir la corrección directa conserva el flujo operativo, pero
  el valor actual no explica por sí solo cómo llegó a existir. El historial de
  actor, instante y valores anterior/nuevo aporta trazabilidad sin una cola de
  aprobaciones.
- **Evidencia:** ADR-0036.
- **Coste aceptado:** todavía no existen restauración, disputa ni revisión por
  terceros.
- **Siguiente decisión:** completada en ADR-0037; continúa el momento válido de
  registro de resultados.

### 2026-07-26 — Delegar no exige una segunda cola de aprobación

- **Aprendido:** si el creador selecciona y puede revocar a quien administra
  resultados, una confirmación adicional suele añadir espera y estados pendientes
  sin aportar valor proporcional en una liga cerrada.
- **Evidencia:** ADR-0035.
- **Coste aceptado:** un marcador equivocado será visible hasta corregirse; la
  corrección y la auditoría se deciden después.
- **Siguiente decisión:** completada en ADR-0036; continúan datos y reglas de
  resultados.

### 2026-07-26 — Competir, seguir y administrar son relaciones distintas

- **Aprendido:** en una liga por equipos, el participante deportivo no tiene por
  qué ser una persona con cuenta. Separar equipos, seguidores y administradores
  evita convertir permisos de aplicación en reglas de competición.
- **Aprendido:** un `username` público permite buscar personas sin exponer su
  correo; mantenerlo inmutable al inicio reduce el modelo mientras se valida su
  uso.
- **Evidencia:** ADR-0034.
- **Coste aceptado:** la administración se concede sin aceptación y todavía no
  dispone de bloqueo, avisos ni límites frente a abuso.
- **Siguiente decisión:** completada en ADR-0035; continúa la corrección de
  resultados.

### 2026-07-27 — Un ID público identifica; la autorización protege mutaciones

- **Aprendido:** una URL puede construirse directamente con el ID de una liga
  visible cuando la lectura se permite a quien lo conozca. Ese ID no es un
  secreto ni concede permisos de edición, resultados o administración.
- **Aprendido:** separar identificador público y autorización evita introducir
  tokens de compartición cuando el producto no necesita restringir la lectura.
- **Evidencia:** ADR-0049, sucesor de ADR-0033.
- **Coste aceptado:** la proyección pública puede ser leída por quien conozca o
  adivine un ID visible; si eso deja de ser aceptable habrá que decidir una
  audiencia restringida o enlaces con capacidad.
- **Siguiente decisión:** implementación del primer vertical slice.

### 2026-07-26 — Publicar fija la estructura, no la visibilidad

- **Aprendido:** una liga no es solo un torneo con equipos: necesita reglas y
  emparejamientos para representar una competición. Generarlos al publicar fija
  una estructura reproducible sin tener que adelantar fechas, resultados o
  clasificación.
- **Aprendido:** el ciclo de vida controla qué cambios son válidos; la
  visibilidad controla quién puede descubrir o acceder al torneo. Son conceptos
  independientes y no deben mezclarse en un único estado.
- **Evidencia:** ADR-0032.
- **Coste aceptado:** una liga publicada no permite modificar equipos; las bajas,
  resultados y variantes de reglas exigirán decisiones posteriores.
- **Siguiente decisión:** completada en ADR-0033; continúa incorporación de
  participantes e invitaciones.

### 2026-07-26 — Un borrador no es todavía un torneo

- **Aprendido:** permitir preparar datos sin registro reduce fricción, pero un
  borrador solo local puede perderse cuando la verificación de email continúa en
  otro navegador o dispositivo.
- **Aprendido:** una cuenta pendiente es un estado temporal de alta, no una
  sesión ni una autorización. Asociar el borrador a esa cuenta evita tratarlo
  como dato anónimo en el servidor; verificar el correo es la frontera que
  habilita publicar el torneo.
- **Evidencia:** ADR-0031 y la aclaración de ADR-0010.
- **Coste aceptado:** definir expiración, purga y límites frente a abuso para
  cuentas y borradores pendientes antes de implementarlos.
- **Siguiente decisión:** completada en ADR-0032; continúa visibilidad y
  mecanismo de incorporación.

### 2026-07-25 — El mapa de red no es una factura

- **Aprendido:** una región es la zona geográfica de AWS; una AZ es una ubicación
  aislada dentro de esa región; una VPC es la red privada que las abarca. Las
  subredes son sus porciones de direcciones internas, no máquinas ni IP públicas.
- **Aprendido:** reservar dos AZ permite que ALB y RDS usen una topología válida,
  pero no cobra por sí mismo. El coste empieza al ejecutar ALB, tareas, base de
  datos, IPv4, logs o transferencia.
- **Evidencia:** ADR-0030; España tiene tres AZ y los servicios base elegidos.
- **Coste aceptado:** no se autoriza gasto recurrente sin estimación completa y
  aprobación explícita de importe y duración.
- **Siguiente decisión:** Gate 0B, formato y participantes del primer vertical
  slice de producto.

### 2026-07-25 — Público no significa abierto directamente

- **Aprendido:** publicar una API no exige que cualquiera pueda conectar con
  ella directamente. El ALB es la puerta pública; la API solo acepta las
  conexiones que proceden de esa puerta mediante una regla de security group.
- **Aprendido:** NAT protege la salida de recursos privados, pero no es gratis.
  Para el alcance inicial se prescinde de ella y se mantiene la base de datos
  privada; si el egress privado aporta valor real, la decisión se reabrirá.
- **Evidencia:** ADR-0029; AWS recomienda restringir los targets a aceptar
  tráfico exclusivamente desde el security group del ALB.
- **Coste aceptado:** ALB, Fargate, IPv4 pública, logs y datos costarán cuando
  exista despliegue; el presupuesto se revisará antes de crear recursos.
- **Siguiente decisión:** región, CIDR, subredes, AZ y presupuesto AWS.

### 2026-07-25 — El backend de estado no debe adelantar la red cloud

- **Aprendido:** un backend remoto no solo guarda un archivo: coordina quién
  puede modificar el estado y conserva evidencia para recuperarlo. El estado
  puede ser remoto aunque `plan` y `apply` se ejecuten inicialmente desde la
  CLI local.
- **Aprendido:** HCP Terraform Free resuelve bloqueo e historial sin decidir
  aún región ni crear un bucket AWS; por eso evita que el backend de estado
  fuerce prematuramente la decisión de red. S3 sigue siendo la alternativa de
  salida si cambian los límites, el coste o el control requerido.
- **Evidencia:** ADR-0028; HCP Terraform Free limita el plan a 500 recursos
  gestionados y el backend S3 requiere versionado y locking explícito.
- **Coste aceptado:** depender inicialmente de HCP y proteger un token de
  acceso, aplazando la operación nativa del backend S3.
- **Siguiente decisión:** región y red AWS.

### 2026-07-25 — El estado no es código ni debe vivir en Git

- **Aprendido:** Git registra la intención de infraestructura; el estado de
  Terraform registra la asociación entre esa intención y recursos reales. Un
  valor marcado como sensible puede seguir existiendo en el estado, por lo que
  no se publica ni se versiona.
- **Aprendido:** estado local no cuesta nada, pero no aporta locking remoto ni
  recuperación compartida. Es suficiente sin AWS real; un backend remoto será
  obligatorio antes del primer apply cloud.
- **Evidencia:** ADR-0027; la comparación posterior se cerró en ADR-0028 con
  HCP Terraform Free como backend remoto inicial.
- **Coste aceptado:** retrasar el primer despliegue cloud hasta decidir y
  verificar el backend remoto.
- **Siguiente decisión:** completada en ADR-0028; continúa región y red AWS.

### 2026-07-25 — Una cuenta AWS es una frontera, no un usuario

- **Aprendido:** una cuenta AWS aísla recursos, permisos y facturación. Separar
  `nonprod` de `prod` limita el radio de un error sin requerir una landing zone
  completa desde el inicio.
- **Aprendido:** el acceso humano usa federación, MFA y roles temporales; una
  cuenta root no es una identidad diaria, y GitHub Actions obtiene credenciales
  temporales mediante OIDC en lugar de conservar access keys.
- **Evidencia:** ADR-0026; AWS Organizations organiza las tres cuentas e IAM
  Identity Center centraliza el acceso humano.
- **Coste aceptado:** gestionar correos raíz, permisos, roles y facturación de
  tres cuentas antes de desplegar la primera carga.
- **Siguiente decisión:** estado y bootstrap de Terraform.

### 2026-07-25 — IaC describe el resultado, no una receta manual

- **Aprendido:** Terraform expresa el estado deseado de infraestructura y su
  flujo separa `plan` —previsualizar cambios— de `apply` —ejecutarlos tras
  aprobación—. El estado es necesario para comparar la declaración con los
  recursos reales y requiere un diseño de seguridad independiente.
- **Aprendido:** seleccionar Terraform no crea una cuenta AWS, una VPC ni un
  recurso facturable; cuenta, identidad, backend de estado y red siguen siendo
  decisiones separadas y ordenadas.
- **Evidencia:** ADR-0025; Terraform queda como IaC declarativa para la Fase 5,
  sin SDKs AWS ni tipos de proveedor en la lógica de negocio.
- **Coste aceptado:** aprender HCL y operar correctamente providers, módulos y
  estado remoto cuando la Fase 5 lo requiera.
- **Siguiente decisión:** fundación AWS: cuenta e identidad.

### 2026-07-25 — Promover un artifact no es promover una rama

- **Aprendido:** GitHub conserva código fuente, ECR conserva imágenes OCI y un
  digest identifica exactamente el artifact que ECS puede ejecutar. El digest
  validado en staging debe ser el mismo que llega a producción; no se reconstruye
  durante la promoción.
- **Aprendido:** `dev`, `staging` y `prod` son entornos, no ramas. `develop`
  puede seguir integrando cambios en dev mientras una `release/*` temporal
  conserva en staging el subconjunto que QA y negocio quieren publicar.
- **Evidencia:** ADR-0024; ECR admite tags inmutables y ECS/Fargate puede
  descargar imágenes privadas mediante el rol de ejecución.
- **Coste aceptado:** operar entornos aislados y componer releases selectivas
  añade recursos, conflictos potenciales y sincronización; se activa solo cuando
  haya equipo o una necesidad real, no durante la práctica individual actual.
- **Siguiente decisión:** IaC y AWS.

### 2026-07-25 — Runtime gestionado antes que Kubernetes

- **Aprendido:** ECS organiza servicios de contenedores y Fargate aporta la
  capacidad gestionada para ejecutarlos; no hay que mantener las máquinas que
  hay por debajo.
- **Aprendido:** escoger un destino cloud futuro no obliga a desplegar hoy. La
  práctica actual continúa por completo en local y no debe crear gasto AWS.
- **Evidencia:** ADR-0023; un servicio ECS puede reemplazar tareas que fallan y
  Fargate factura por los recursos consumidos mientras se ejecutan.
- **Coste aceptado:** la red, IAM, Terraform, registro, costes y recuperación
  real se aprenderán en la Fase 5, antes de mantener recursos activos.
- **Siguiente decisión:** registry y promoción de la imagen.

### 2026-07-25 — Empaquetar no es elegir dónde desplegar

- **Aprendido:** una imagen OCI es una caja reproducible para ejecutar la API;
  no es AWS, Kubernetes, un VPS ni un contenedor que deba incluir toda la
  aplicación y sus dependencias.
- **Aprendido:** separar build y runtime evita transportar al entorno final el
  compilador y herramientas de desarrollo; el mismo artefacto podrá identificarse
  de forma inmutable antes de promoverse.
- **Evidencia:** ADR-0022; Docker recomienda las imágenes multi-stage para
  separar compilación y ejecución.
- **Coste aceptado:** habrá que mantener y comprobar la imagen cuando exista una
  API, sin adelantar registry, runtime, firma ni pipeline de despliegue.
- **Siguiente decisión:** runtime y promoción de la API.

### 2026-07-25 — CI como contraste, no como burocracia

- **Aprendido:** la utilidad de CI en un proyecto individual es repetir el
  contrato de calidad en un entorno limpio, no obligar a una persona a revisar
  su propio cambio mediante una PR.
- **Aprendido:** un check obligatorio puede bloquear una promoción cuando el
  recurso externo no está disponible; por eso la puerta vigente es `make verify`
  local y CI aporta una segunda evidencia visible.
- **Evidencia:** ADR-0021; GitHub Actions en runners estándar para repositorio
  público, sin secretos, permisos de escritura ni triggers privilegiados.
- **Coste aceptado:** un resultado rojo no bloquea técnicamente `main` mientras
  no haya colaboración; la disciplina de promoción sigue siendo manual.
- **Siguiente decisión:** contenedores y despliegue.

### 2026-07-25 — Señales correlacionadas, no prints aislados

- **Aprendido:** un log describe un evento discreto; una traza reúne los spans
  de una operación; una métrica agrega medidas en el tiempo. Son señales
  complementarias, no alternativas.
- **Aprendido:** la instrumentación automática debe cubrir límites técnicos;
  los spans manuales se reservan para operaciones significativas que no se vean
  automáticamente, nunca uno por función.
- **Evidencia:** ADR-0020; OpenTelemetry, Prometheus, Grafana, Loki y Tempo,
  con correlación por contexto y sin OpenTelemetry Collector inicialmente.
- **Coste aceptado:** operar cuatro servicios locales y retrasar Collector,
  alertas y SLO hasta que exista una necesidad demostrable.
- **Siguiente decisión:** CI y política de calidad.

### 2026-07-25 — Evidencia proporcional al riesgo

- **Aprendido:** un test rápido no equivale a evidencia suficiente cuando el
  riesgo vive en SQL, restricciones o transacciones; esos comportamientos se
  validan con PostgreSQL real.
- **Aprendido:** una suite end-to-end completa no sustituye las pruebas pequeñas:
  cuesta más diagnosticarla y debe reservarse a recorridos críticos.
- **Evidencia:** ADR-0019, con pruebas unitarias estándar, integración en una
  base efímera, contratos HTTP y E2E mínimos.
- **Coste aceptado:** preparar una base de pruebas aislada y posponer librerías
  adicionales hasta que exista una necesidad repetida.
- **Siguiente decisión:** observabilidad mínima.

### 2026-07-25 — Dependencias contenidas, aplicaciones nativas

- **Aprendido:** la paridad local con producción consiste en conservar contratos
  importantes —configuración externa, PostgreSQL real, salud, volumen y
  migraciones—, no en contenerizar todo desde el primer día.
- **Aprendido:** Expo para web, iOS y Android conserva mejor su bucle local al
  ejecutarse en el host, donde dispone de sus herramientas y simuladores nativos.
- **Evidencia:** ADR-0018; PostgreSQL queda delimitado como primera dependencia
  de Docker Compose.
- **Coste aceptado:** la API futura debe diagnosticar PostgreSQL no disponible y
  la imagen final de API se validará en una decisión posterior de despliegue.
- **Siguiente decisión:** estrategia de pruebas.

### 2026-07-24 — Configuración pública, secretos y entornos

- **Aprendido:** no toda configuración es secreta, pero toda configuración debe
  tener propietario, contrato y destino.
- **Aprendido:** cualquier valor `EXPO_PUBLIC_*` acaba accesible para quien use
  el cliente, así que nunca es secreto.
- **Aprendido:** OIDC permite que CI obtenga credenciales cloud temporales sin
  guardar access keys largas en GitHub.
- **Evidencia:** ADR-0017 y reglas actualizadas en seguridad, desarrollo y
  despliegue.
- **Coste aceptado:** los `.env` locales se crean fuera de Git y la validación de
  configuración debe implementarse en cada runtime.
- **Siguiente decisión:** completada en ADR-0018.

### 2026-07-24 — Rendering simple para producto privado

- **Aprendido:** si el producto inicial vive en torneos privados, SEO, static
  rendering y SSR no son necesidades de base sino posibles evoluciones.
- **Aprendido:** compartir cliente no significa mezclar diferencias de plataforma
  dentro de la lógica; se aíslan en componentes o archivos específicos.
- **Evidencia:** ADR-0016 y documentación ajustada a privacidad inicial y acceso
  invitado limitado.
- **Coste aceptado:** la web inicial no optimiza indexación ni previews sociales
  por torneo.
- **Siguiente decisión:** completada en ADR-0017.

### 2026-07-24 — Rutas universales y nativo generado

- **Aprendido:** Expo Router asigna una URL a cada pantalla; esa misma ruta puede
  abrirse en web o mediante deep link nativo.
- **Aprendido:** CNG no elimina el código nativo de una app: desplaza su fuente
  de verdad a configuración y plugins reproducibles.
- **Evidencia:** ADR-0015, rutas y directorios nativos futuros documentados e
  ignorados por Git.
- **Coste aceptado:** usar convenciones Expo y no tocar directorios generados.
- **Siguiente decisión:** completada en ADR-0016.

### 2026-07-24 — Toolchain TypeScript reproducible

- **Aprendido:** `Current` recibe primero las novedades de Node; LTS prioriza una
  ventana estable y prolongada, adecuada para desarrollo y CI.
- **Aprendido:** una versión `latest` no es automáticamente la mejor elección;
  TypeScript debe permanecer dentro del rango compatible del linter.
- **Aprendido:** TypeScript 6 rechaza una lista `files` vacía; el check raíz
  valida configuración real mientras todavía no existen workspaces TypeScript.
- **Evidencia:** ADR-0014, runtime y package manager pineados, workspaces, lockfile
  y checks compartidos.
- **Coste aceptado:** pnpm, ESLint y Prettier añaden piezas separadas a cambio de
  dependencias explícitas y responsabilidades claras.
- **Siguiente decisión:** framework del cliente universal.

### 2026-07-24 — Flujo de integración proporcional al equipo

- **Aprendido:** una rama `develop` conserva `main` como hito estable, pero puede
  acumular cambios no publicables y bloquear promociones completas.
- **Evidencia:** ADR-0013 y guía de contribución con reglas para sincronización,
  promoción y excepciones.
- **Coste aceptado:** mantener dos ramas de larga vida mientras el trabajo sea
  principalmente individual.
- **Siguiente decisión:** toolchain TypeScript.

### 2026-07-24 — Identidad canónica del repositorio

- **Aprendido:** la ruta de un módulo Go público forma parte de su identidad y
  debe coincidir con el propietario y nombre reales del remoto.
- **Evidencia:** cuenta de GitHub autenticada y nombre
  `joseantoniogarciay/TournamentsManager` comprobado antes del primer push.
- **Coste evitado:** no publicar una ruta provisional que después obligue a
  cambiar imports y consumidores.
- **Siguiente decisión:** toolchain TypeScript.

### 2026-07-24 — Toolchain Go reproducible

- **Aprendido:** `go tool` ejecuta herramientas declaradas; `tool` las registra y
  `-modfile` selecciona un grafo alternativo.
- **Aprendido:** `go.tool.mod` no es un nombre mágico; Make y el wrapper de VS
  Code encapsulan su selección explícita.
- **Evidencia:** ADR-0012, módulos separados, Makefile y formato al guardar con
  `goimports` pineado.
- **Coste aceptado:** mantener dos pares `go.mod`/`go.sum` y ejecutar
  `tidy-check` para ambos.
- **Siguiente decisión:** toolchain TypeScript.

### 2026-07-24 — Persistencia SQL-first tipada

- **Aprendido:** `pgx` es el driver PostgreSQL, `sqlc` genera código Go desde SQL
  y `goose` versiona la evolución del esquema; resuelven problemas diferentes.
- **Aprendido:** el código generado evita trabajo mecánico y detecta deriva, pero
  no diseña consultas, transacciones, índices ni modelos de dominio.
- **Evidencia:** ADR-0011 y reglas documentadas para separar filas, adaptadores y
  dominio.
- **Coste aceptado:** mantener SQL, generación determinista, mapeos explícitos y
  una política operativa de migraciones.
- **Siguiente decisión:** toolchain Go.

### 2026-07-24 — Identidad propia federada

- **Aprendido:** el `subject` identifica una cuenta dentro del proveedor; el
  backend debe extraerlo de un token verificado y mapearlo a un usuario interno.
- **Aprendido:** una URL navegable no debe convertir un `GET` en una operación
  sensible; la apertura presenta el estado y una confirmación explícita mediante
  `POST` consume el intento antes de volver a la home.
- **Evidencia:** ADR-0010 y flujos documentados para cambio de email y vinculación
  con prueba fresca.
- **Coste aceptado:** operar credenciales, sesiones, recuperación, OAuth/OIDC y
  controles de abuso.
- **Siguiente decisión:** completada en ADR-0011.

### 2026-07-24 — Navegación del handbook

- **Aprendido:** mantener todos los documentos en la raíz aumenta visibilidad al
  principio, pero deja de escalar cuando oculta las unidades del monorepo.
- **Evidencia:** handbook agrupado por proyecto, gobierno, ingeniería y
  operaciones, con índices y enlaces validados.
- **Coste aceptado:** mantener rutas e índices como parte de cualquier movimiento
  documental.
- **Siguiente decisión:** arquitectura de identidad.

### 2026-07-24 — Contrato API y cliente generado

- **Aprendido:** la API HTTP es un adaptador del backend, no todo el backend;
  OpenAPI coordina el servidor Go y el consumidor TypeScript.
- **Evidencia:** ADR-0009 con REST contract-first, generación del cliente y
  límites respecto al dominio.
- **Coste aceptado:** mantener lint, generación y compatibilidad, especialmente
  con aplicaciones instaladas que se actualizan más tarde.
- **Siguiente decisión:** arquitectura de identidad.

### 2026-07-24 — Estrategia universal de cliente

- **Aprendido:** compartir producto y comportamiento no obliga a usar una
  presentación idéntica; responsive y adaptativo resuelven tamaños y capacidades
  de entrada diferentes.
- **Evidencia:** ADR-0008 con paridad funcional, límites de plataforma y
  disparadores de revisión.
- **Coste aceptado:** aislar excepciones web/native y validar por separado SEO,
  accesibilidad, rendimiento y releases.
- **Siguiente decisión:** estilo y contrato de API.

### 2026-07-24 — Topología técnica inicial

- **Aprendido:** monorepo no significa un único pipeline, y monolito no significa
  ausencia de límites.
- **Evidencia:** ADR de monorepo, publicación segura y monolito modular.
- **Coste aceptado:** disciplina para mantener módulos y pipelines independientes.
- **Siguiente decisión:** estrategia web/mobile y frontera de reutilización.

### 2026-07-23 — Fundación documental

- **Aprendido:** una arquitectura profesional comienza por explicitar autoridad,
  proceso, estados y criterios de salida.
- **Evidencia:** manifiesto transcrito, ADR iniciales, mapa del handbook y
  plantillas operativas.
- **Incertidumbre:** requisitos del producto y decisiones de implementación.
- **Siguiente experimento:** definir el alcance y primer caso de uso antes del
  entorno o del backend.

### 2026-07-26 — Diseñar datos y contrato juntos

- **Aprendido:** el estado temporal de una cuenta, el secreto de sesión y el
  recurso de negocio tienen ciclos de vida distintos; modelarlos como una sola
  tabla o token hace más fácil conceder permisos antes de verificarlos.
- **Aprendido:** OpenAPI describe el borde HTTP, mientras que restricciones y
  transacciones pertenecen al dominio y a la persistencia; ambos se revisan
  juntos para evitar contratos que no se pueden garantizar.
- **Evidencia:** ADR-0045, modelo lógico y contrato OpenAPI 3.1 del primer
  incremento, antes de migraciones o handlers.
- **Coste aceptado:** mantener alineados la especificación, las migraciones y
  las pruebas de contrato al evolucionar cada flujo.
- **Siguiente decisión:** tooling de lint, generación y verificación semántica
  de OpenAPI antes de implementar el adaptador HTTP.

### 2026-07-26 — Deriva de contrato como fallo verificable

- **Aprendido:** validar YAML no comprueba por sí solo reglas como declarar
  operaciones públicas o respuestas de error; un linter de OpenAPI convierte
  esas convenciones en una comprobación reproducible.
- **Aprendido:** el cliente generado debe vivir aislado del código manual, pues
  la regeneración puede limpiar su directorio completo.
- **Evidencia:** ADR-0046, Redocly CLI y Orval integrados en Make y pnpm.
- **Coste aceptado:** revisar los diffs de generación y mantener las versiones
  pineadas de las dos herramientas.
- **Siguiente decisión:** migraciones iniciales y consultas SQL del modelo
  aceptado.

### 2026-07-27 — Una migración también es un contrato

- **Aprendido:** el modelo lógico explica relaciones e invariantes; la migración
  demuestra cuáles puede proteger PostgreSQL mediante tipos, FKs, `CHECK` e
  índices únicos.
- **Aprendido:** sqlc no necesita una copia del esquema: analizar las migraciones
  conserva una sola fuente de verdad y evita generar consultas inexistentes.
- **Evidencia:** ADR-0047, primera migración Goose y configuración sqlc con
  `pgx/v5`.
- **Coste aceptado:** revisar cuidadosamente cada evolución de esquema y
  regenerar código SQL solo cuando exista una consulta de producto.
- **Siguiente decisión:** primera operación de persistencia y límites de su
  transacción antes de implementar registro.

### 2026-07-27 — Consolidar antes de compartir

- **Aprendido:** una migración local y no compartida puede consolidarse para que
  la primera instalación parta de un esquema inicial claro; una migración
  compartida se trata como inmutable y se evoluciona con otra migración.
- **Aprendido:** una cuenta pendiente puede tener todos sus datos de identidad;
  verificar correo acredita el canal, no completa un perfil incompleto.
- **Evidencia:** ADR-0048, migración inicial consolidada y reinicio comprobado
  de PostgreSQL local.
- **Coste aceptado:** límites de tasa y purga protegen la reserva temporal de
  usernames y el reenvío de verificación.
- **Siguiente decisión:** primera operación de persistencia y límites de su
  transacción antes de implementar registro.

### 2026-07-27 — Opcional no significa parcialmente válido

- **Aprendido:** una propiedad opcional en el objeto padre puede seguir remitiendo
  a un objeto con campos e invariantes obligatorios; omitir el objeto y enviar un
  objeto incompleto son entradas distintas.
- **Evidencia:** ADR-0052, `RegisterRequest` con `draft` opcional y
  `DraftInput` con `name` y un mínimo de dos equipos.
- **Coste aceptado:** el adaptador HTTP futuro debe comprobar estas reglas en
  runtime además del tipado generado para TypeScript.
- **Siguiente decisión:** primera operación de persistencia y límites de su
  transacción antes de implementar registro.

### 2026-07-29 — La splash nativa no es una pantalla de React

- **Aprendido:** la splash se muestra antes de ejecutar JavaScript; solo puede
  configurarse en la build nativa. El código de la app puede decidir cuándo
  ocultarla, pero no convertirla en un flujo de carga.
- **Evidencia:** `expo-splash-screen` configurado mediante el config plugin y
  retenido hasta hidratar la preferencia local de tema.
- **Coste aceptado:** la apariencia final se prueba en una build release, porque
  Expo Go y development builds no la reproducen fielmente.
- **Regla reutilizable:** no esperar red bajo la splash ni crear una falsa splash
  React salvo que un requisito de producto justifique explícitamente esa espera.

### 2026-08-14 — La splash no reutiliza el recuadro del icono de aplicación

- **Aprendido:** el icono de aplicación cuadrado funciona como identificador en
  el sistema, pero repetido sobre la splash introduce un segundo fondo y hace la
  marca más pesada.
- **Evidencia:** `fast-tourney-splash-mark.png` conserva solo la marca interior
  transparente con el degradado azul–violeta y se aplica sobre los canvases
  claro y oscuro de `expo-splash-screen`.
- **Regla reutilizable:** exportar una marca transparente específica para la
  splash; verificarla en ambos temas antes de cambiar el recurso nativo.

### 2026-07-30 — El cliente generado necesita un transporte de ejecución

- **Aprendido:** generar tipos y operaciones desde OpenAPI no elimina la
  configuración propia del cliente instalado; URL base, cancelación y futuras
  credenciales pertenecen a un transporte común, no a cada feature.
- **Evidencia:** Orval genera un parámetro `fetchFn`; `apiFetch` resuelve la URL
  base y el adaptador de registro invoca la operación generada.
- **Coste aceptado:** las features mantienen un adaptador pequeño que traduce el
  contrato a sus estados de interfaz, sin reconstruir URL, `fetch` ni DTOs.
- **Regla reutilizable:** para una operación del contrato, usar siempre la
  función OpenAPI generada a través del adaptador de la feature y `apiFetch`.

### 2026-07-30 — El fallback de errores debe separar transporte y negocio

- **Aprendido:** el tipo generado por Orval enumera las respuestas documentadas,
  pero no clasifica un rechazo de red ni una respuesta HTTP nueva. Solo el
  transporte común puede identificar con seguridad que no se recibió respuesta.
- **Evidencia:** `apiFetch` convierte esos rechazos en `APIConnectionError` y
  `shared/feedback/request-failure.ts` los traduce a las claves comunes de red o
  error seguro; el chequeo de username conserva su tratamiento específico de
  `200` y `429`.
- **Coste aceptado:** cada feature debe elegir explícitamente sus estados de
  negocio valiosos; los no previstos, cuerpos inválidos y `5xx` usan el fallback
  genérico sin exponer detalles del backend.

### 2026-07-30 — Un enlace universal acredita contexto, no hace un GET mutante

- **Aprendido:** un correo puede llevar a una ruta HTTPS compartida por web,
  iOS y Android, pero abrirlo mediante `GET` no debe activar una cuenta ni crear
  una sesión. El cliente conserva el token de un solo uso fuera de la URL y
  emite automáticamente el `POST` de confirmación.
- **Evidencia:** correo multipart con alternativa accesible, ruta
  `/link/confirm`, configuración declarativa CNG y plantillas de asociación
  bajo `infra/app-links/`.
- **Coste aceptado:** la asociación real exige un dominio, Team ID de Apple y
  huellas Android de la firma de distribución; se documentan como datos de
  despliegue, no se inventan en el repositorio.
- **Regla reutilizable:** un secreto en un enlace no se envía a analítica ni a
  recursos de terceros y se elimina del historial tan pronto como la aplicación
  lo recibe.

### 2026-07-31 — Reemplazar una sesión exige resetear navegación, no cerrar una modal

- **Aprendido:** `dismissAll()` equivale a un `popToTop` del stack más cercano;
  no representa un reset de aplicación y falla si el deep link llegó en frío sin
  una pila que pueda responder a esa acción.
- **Evidencia:** ADR-0061; la transición de sesión mantiene una capa global
  durante la confirmación y el árbol raíz de navegación se reconstruye una vez
  obtenida la nueva sesión.
- **Coste aceptado:** la transición necesita comprobar tanto enlaces abiertos en
  frío como enlaces recibidos sobre modales y tabs ya activas.
- **Regla reutilizable:** navegar a una raíz no debe disparar datos de todas las
  secciones; cada raíz carga solo al ser visible y con la identidad vigente.

### 2026-07-31 — Un deep link de arranque necesita una raíz estable antes de ejecutarse

- **Aprendido:** el router puede recibir el enlace que despierta una app nativa
  antes de que Inicio haya llegado a pantalla. Ejecutar su ruta de inmediato
  compite con el montaje y puede dejar transiciones o estado de navegación
  incoherentes.
- **Evidencia:** `+native-intent` conserva únicamente en memoria el enlace de
  arranque, fuerza la raíz `/` y Inicio lo entrega al router en el siguiente
  frame; un enlace recibido por una app viva no se desvía.
- **Coste aceptado:** hay un frame adicional antes de ejecutar un enlace de
  arranque, a cambio de una raíz de navegación estable y sin persistir tokens.
- **Regla reutilizable:** el tratamiento de enlaces de arranque pertenece al
  borde de navegación, no a cada flujo de negocio que pueda abrirse desde uno.

### 2026-07-31 — Restaurar una sesión móvil no autoriza al cliente

- **Aprendido:** Keychain/Keystore puede conservar los dos secretos opacos, la
  identidad y sus expiraciones como una sola unidad de arranque; ni el perfil ni
  las fechas sustituyen la validación de cada operación en el backend.
- **Evidencia:** ADR-0062; `apiFetch` comparte una única renovación cuando queda
  menos de una hora de access y el `SessionProvider` hidrata la identidad local.
- **Coste aceptado:** un refresh inválido resetea la sesión mediante el
  coordinador; un rechazo de red la conserva para permitir reintento posterior.
- **Regla reutilizable:** almacenar tokens relacionados por separado permite
  restauraciones parciales; persistirlos juntos reduce ese estado imposible.

### 2026-07-31 — Una transición de identidad puede tener una duración visual mínima

- **Aprendido:** una confirmación de deep link puede completarse más rápido que
  la animación que comunica el cambio de identidad. Mantener una capa opaca con
  texto y loader durante al menos dos segundos permite revisar su feedback sin
  retrasar la petición ni modificar el resultado de verificación.
- **Evidencia:** la ruta `link/confirm` inicia la transición antes del `POST`
  generado y, tras una respuesta correcta, espera solo el tiempo restante hasta
  dos segundos antes de sustituir la sesión y resetear la navegación.
- **Coste aceptado:** el usuario tarda hasta dos segundos adicionales en llegar
  a Inicio después de una verificación satisfactoria; los fallos no se retrasan.
- **Regla reutilizable:** cuando una espera exista solo para observar feedback,
  se mide desde el inicio de la transición y nunca se antepone a la operación;
  esa capa solo usa fade in y fade out, sin desplazar el contenido.

### 2026-07-31 — La capa de carga es una primitiva, no una regla de sesión

- **Aprendido:** el patrón de bloqueo con mensaje, loader y fade puede aparecer
  fuera de identidad; acoplarlo a `SessionProvider` dificultaría reutilizarlo.
- **Evidencia:** `LoadingTransition` vive en `shared/ui`, recibe estado y copy
  localizado y SessionProvider solo le aporta el estado de transición actual.
- **Coste aceptado:** cada flujo conserva su duración mínima y su resultado; la
  primitiva se limita a presentación, accesibilidad y movimiento reducido.
- **Regla reutilizable:** extraer una primitiva compartida cuando el aspecto y
  la interacción se repiten, sin trasladarle decisiones de negocio.

### 2026-07-31 — El contrato puede distinguir fallos que el backend aún agrupa

- **Aprendido:** `409` para un enlace ya consumido y `410` para uno caducado
  ofrecen recuperaciones diferentes y el cliente debe poder expresarlas sin
  mostrar el detalle RFC 9457.
- **Evidencia:** el adaptador de registro traduce ambos estados declarados a un
  error de feature; la ruta muestra copy localizado y la previsualización de
  desarrollo reproduce el caso `409` sin HTTP.
- **Coste aceptado:** el backend actual devuelve `409` para todos los tokens
  inválidos; el copy de caducidad queda preparado hasta que la persistencia
  clasifique y entregue `410` de forma verificable.
- **Regla reutilizable:** una feature mapea solo los estados declarados que
  cambian la recuperación; los demás continúan por el fallback seguro.

### 2026-07-31 — El idioma de un email es una preferencia de cuenta, no un ajuste aislado

- **Aprendido:** persistir el locale efectivo al crear la cuenta permite
  localizar el email de verificación y reutilizar la misma preferencia en
  entregas futuras sin introducir una configuración exclusiva de correo.
- **Evidencia:** ADR-0063 limita los valores a `es`, `en`, `it` y `fr`, que son
  los locales ya aceptados para el cliente; el backend valida el dato antes de
  guardarlo y seleccionar una plantilla.
- **Coste aceptado:** un cambio posterior de idioma en el sistema o navegador
  no actualiza todavía esa preferencia; una futura sincronización autenticada
  se evaluará solo si existe necesidad demostrada.
- **Regla reutilizable:** las preferencias de presentación que trascienden una
  entrega concreta pertenecen a la cuenta y se validan en el límite del backend.

### 2026-07-31 — El gestor de contraseñas decide el guardado de credenciales

- **Aprendido:** `new-password` permite que iOS, Android y la web propongan una
  contraseña fuerte y creen o actualicen la entrada del gestor sin que el
  producto manipule el llavero.
- **Evidencia:** ADR-0064 separa la semántica del campo de la política del
  backend: sugerencia de 15 o más, mínimo manual de 8 y validación final en API.
- **Coste aceptado:** el medidor es orientativo; no reemplaza controles de
  servidor ni afirma que una contraseña sea invulnerable.
- **Regla reutilizable:** declarar la intención semántica correcta del campo y
  dejar que el proveedor de credenciales conserve el secreto.

### 2026-08-01 — La visibilidad de una contraseña es un control, no un carácter

- **Aprendido:** un carácter Unicode no representa de manera consistente una
  acción de mostrar u ocultar entre plataformas. El control debe expresar su
  estado con los símbolos nativos de ojo y ojo tachado, además de conservar un
  área táctil de 44 px.
- **Evidencia:** `TextField` usa `expo-symbols`: SF Symbols en iOS y Material
  Symbols en Android y web. La prop compartida recibe el estado visible para
  que icono y etiqueta accesible cambien juntos.
- **Coste aceptado:** se añade una dependencia pequeña incluida en Expo Go, en
  vez de incorporar una biblioteca de iconos completa o assets duplicados.
- **Regla reutilizable:** los iconos que describen un estado interactivo se
  derivan del mismo estado que la acción y su etiqueta accesible.

### 2026-08-01 — Una tab nativa no debe minimizarse si invalida su inset

- **Aprendido:** la minimización de `NativeTabs` de iOS 26 altera el espacio
  inferior que usa el scroll. Si no se recompone al cerrar el teclado hasta un
  nuevo scroll, deja una separación visual incorrecta.
- **Evidencia:** el problema desaparece al fijar `minimizeBehavior="never"`;
  la barra mantiene una altura estable mientras el teclado entra y sale.
- **Coste aceptado:** se renuncia a ocultar la tab bar al desplazarse. Es menor
  que mantener un layout que puede quedarse desincronizado.
- **Regla reutilizable:** una animación del contenedor de navegación solo se
  conserva si mantiene correctos los insets tras cambios del teclado.

### 2026-08-01 — La precarga de un proveedor no debe producir feedback fuera de su ruta

- **Aprendido:** una tab puede montarse sin estar visible. Solicitar el nonce de
  Google en el montaje convierte un fallo de preparación en un banner global al
  iniciar en Inicio, donde no hay acción de Google.
- **Evidencia:** la carga pasa al foco de Cuenta y sus fallos automáticos no se
  publican como error; el botón sigue reintentando la carga y comunica los
  fallos de una acción iniciada explícitamente.
- **Coste aceptado:** la primera llegada a Cuenta puede mostrar brevemente el
  loader del icono. Evita peticiones y feedback ajenos a la ruta visible.
- **Regla reutilizable:** una precarga solo puede emitir feedback si la persona
  ve y entiende la acción que la originó.

### 2026-08-01 — El teclado cambia el viewport web y el inset nativo de forma distinta

- **Aprendido:** el teclado virtual reduce el viewport visual de la web, mientras
  que iOS necesita ajustar el inset del `ScrollView` para que su contenido siga
  siendo desplazable. Dejar que la tab bar participe en el flujo normal puede
  situarla por encima del borde visible en web.
- **Evidencia:** la barra web se fija al borde inferior del viewport visual y
  los formularios no suman el safe-area inset inferior dinámico en web; la
  primitiva de formularios habilita `automaticallyAdjustKeyboardInsets` en iOS.
  Al recuperar altura, solo Safari recibe una segunda medida de
  `visualViewport` tras 250 ms para no conservar su valor intermedio.
- **Coste aceptado:** se usa un selector CSS web moderno con fallback al layout
  actual en navegadores sin `:has`; no se introduce una librería de teclado.
- **Regla reutilizable:** tratar por separado el anclaje de navegación web y el
  desplazamiento del contenido nativo al aparecer el teclado.

### 2026-08-01 — Una ruta terminal no duplica su motivo en el banner global

- **Aprendido:** cuando una ruta ya presenta el motivo de un fallo y la acción
  de recuperación, repetir el texto en un banner superior reduce el espacio
  útil y divide la atención sin añadir información.
- **Evidencia:** `link/confirm` representa los enlaces ausentes, caducados,
  consumidos y los fallos seguros no tipados dentro de una `Card`; deja de
  publicar el mismo resultado mediante `FeedbackProvider`.
- **Coste aceptado:** el aviso no persiste al salir de la ruta, lo cual es
  correcto porque el motivo y su acción solo son relevantes mientras esa ruta
  está visible.
- **Regla reutilizable:** una pantalla terminal con recuperación propia usa
  feedback en el contenido; el banner global se reserva para avisos que no
  tengan una ruta visible que los explique.

## Regla de evidencia

“Entendido” exige una explicación propia y una demostración. Un comando que
funciona o una respuesta del asistente no son evidencia suficiente por sí solos.

# 2026-08-01 — Reautenticación no equivale a sesión

Un token de sesión demuestra que el cliente mantiene una sesión; no debe ser
suficiente por sí solo para cambiar autenticadores persistentes. Un ticket
breve, ligado a la sesión y consumido dentro de la transacción conserva ese
límite sin crear una segunda sesión ni un estado de interfaz global.

### 2026-08-02 — Iniciar sesión no autoriza persistir una contraseña

- **Aprendido:** el formulario de login transmite la contraseña una sola vez al
  endpoint público; el cliente solo conserva los secretos de sesión que emite
  el backend y únicamente en el almacenamiento seguro nativo.
- **Evidencia:** `POST /v1/sessions` compara Argon2id en el caso de uso, crea
  hashes de access y refresh para PostgreSQL y devuelve los secretos solo para
  el transporte Bearer. La web recibe una cookie `HttpOnly`.
- **Coste aceptado:** el backend consulta PostgreSQL para autenticar y mantiene
  dos secretos opacos para móvil; a cambio no se añade JWT ni una contraseña
  recuperable en el cliente.
- **Regla reutilizable:** una credencial de login no se guarda ni se reutiliza
  como sesión; solo el backend decide y emite la credencial de sesión.

### 2026-08-02 — Los callbacks de foco solo dependen de identidades estables

- **Aprendido:** un callback entregado a `useFocusEffect` se vuelve a ejecutar
  mientras la ruta tiene el foco cada vez que cambia alguna dependencia por
  identidad. Si ese efecto actualiza estado, una función recreada en cada render
  puede convertir una carga normal en un ciclo de actualizaciones.
- **Evidencia:** `getTranslator` devolvía un closure nuevo en cada render de la
  pestaña de torneos; la carga modificaba el estado y React Navigation volvía a
  activar el efecto. Los traductores ahora se conservan por locale y el valor
  del contexto de feedback se memoriza a partir de `show`.
- **Coste aceptado:** se mantienen cuatro funciones de traducción de módulo,
  un coste constante y despreciable que evita añadir estado global de idioma.
- **Regla reutilizable:** toda dependencia de `useFocusEffect` que no cambie
  por un dato de producto debe tener una identidad estable.

### 2026-08-04 — Una tab resume y una ruta profunda detalla

- **Aprendido:** la pestaña Cuenta puede mostrar solo la identidad visible y un
  acceso a los datos sensibles, dejando email y métodos de autenticación en una
  ruta propia con historial y cabecera nativos.
- **Evidencia:** la ruta `/account/access` obtiene los métodos mediante el
  adaptador `account-access`; Cuenta no duplica esa carga ni su presentación.
- **Coste aceptado:** se añade una transición antes de cambiar una contraseña o
  vincular Google, a cambio de una pantalla de Cuenta más clara y extensible.
- **Regla reutilizable:** usar una ruta profunda para el detalle cuando una
  pantalla de tab solo necesita resumir una sección de gestión.

### 2026-08-04 — Un selector dependiente de red espera la primera respuesta

- **Aprendido:** mostrar una selección antes de conocer las colecciones que la
  determinan puede cambiarla de inmediato y producir una transición visual
  confusa.
- **Evidencia:** la biblioteca de torneos conserva un estado explícito de primera
  carga y solo presenta «Administro» y «Sigo» después de recibir ambas
  colecciones. La selección inicial prioriza «Administro» si tiene elementos o
  si ambas colecciones están vacías; solo usa «Sigo» cuando es la única con
  elementos.
- **Coste aceptado:** se añade un booleano local, en lugar de una caché o store
  compartida, porque el estado solo coordina el primer render de esta pantalla.
- **Regla reutilizable:** un control cuyo valor inicial depende de varias
  respuestas no se muestra hasta tener todos los datos que definen esa decisión.

### 2026-08-09 — Las transiciones del ciclo se serializan sobre la entidad

- **Aprendido:** dos acciones válidas por separado, como iniciar y cancelar una
  liga, pueden competir sobre el mismo estado si se ejecutan al mismo tiempo.
- **Evidencia:** el adaptador PostgreSQL bloquea la fila de `leagues` con `FOR
UPDATE`, comprueba la organizadora y el estado dentro de la misma transacción,
  y solo después actualiza a `in_progress` o `cancelled`.
- **Coste aceptado:** una transición concurrente espera brevemente el bloqueo;
  es menor que introducir colas, eventos o control de versiones optimista en
  este primer ciclo.
- **Regla reutilizable:** una transición exclusiva de una entidad se autoriza y
  valida dentro de la transacción que modifica su estado.

### 2026-08-09 — El último deeplink reemplaza la intención pendiente

- **Aprendido:** una ruta de confirmación no puede conservar un token previo
  cuando recibe un enlace nuevo; de hacerlo, puede confirmar o mostrar el
  resultado de una intención que la persona ya sustituyó.
- **Evidencia:** la confirmación conserva el intento activo con un
  `AbortController`, cancela el transporte anterior al recibir otro token y
  solo permite que el intento vigente cambie sesión, feedback o navegación.
- **Coste aceptado:** se mantiene un estado local mínimo de coordinación en la
  ruta, sin persistir secretos ni crear un store global adicional.
- **Regla reutilizable:** en una ruta que recibe acciones de un solo uso por
  deeplink, el último enlace es la autoridad y las cancelaciones intencionadas
  no producen feedback.

### 2026-08-09 — Un estado de contrato no es copy de interfaz

- **Aprendido:** los valores de estado que devuelve la API son identificadores
  estables para el código, no textos que deba leer una persona.
- **Evidencia:** una traducción compartida presenta los cuatro estados de liga
  en la home, la biblioteca y el detalle, evitando que una de esas vistas
  exponga `in_progress`.
- **Regla reutilizable:** toda enumeración del contrato que llegue a una UI se
  traduce mediante una clave semántica compartida antes de renderizarse.

### 2026-08-09 — La corrección rápida necesita una huella mínima

- **Aprendido:** aplicar marcadores de inmediato no exige una capa de eventos o
  aprobación, pero sí conservar quién reemplazó qué para poder explicar un
  cambio.
- **Evidencia:** la transacción bloquea liga y partido, actualiza el marcador y
  añade una fila con el valor previo, el nuevo, la administradora y el instante.
- **Coste aceptado:** el historial permanece interno; no se adelantan una UI de
  restauración ni reglas de disputa.
- **Regla reutilizable:** para una mutación corregible, una tabla de cambios
  acotada suele ser suficiente antes de introducir event sourcing.

### 2026-08-09 — El detalle conserva la salida y separa la edición del resumen

- **Aprendido:** una ruta profunda necesita una salida explícita incluso cuando
  se abre sin historial; la cabecera puede volver a la ruta previa y usar la
  home como fallback seguro.
- **Evidencia:** el detalle de liga se presenta como modal en apps y página en
  web, con un título truncado a dos líneas y acciones agrupadas en el menú
  nativo de la barra. Cada partido muestra primero equipos y marcador; los
  campos de edición y su acción quedan en un bloque separado.
- **Coste aceptado:** se añade un menú de cabecera adaptado a plataforma, sin
  introducir una librería de menús ni duplicar la lógica de acciones.
- **Regla reutilizable:** en una vista de gestión, separar lectura, entrada y
  confirmación reduce el error de interacción sin crear nuevos estados de
  producto.

### 2026-08-09 — La jornada es contexto persistente de una lista de partidos

- **Aprendido:** repetir la jornada dentro de cada partido desperdicia espacio
  y se pierde el contexto al desplazarse por una lista larga.
- **Evidencia:** el detalle agrupa los partidos por jornada y usa cabeceras de
  sección persistentes; la jornada actual permanece visible hasta que la
  siguiente toma su lugar.
- **Coste aceptado:** la vista construye secciones locales a partir de los
  partidos ya recibidos, sin solicitar una nueva proyección de API.
- **Regla reutilizable:** cuando una colección tiene grupos secuenciales, una
  cabecera sticky conserva el contexto con menos ruido que duplicar su etiqueta
  en cada fila.

### 2026-08-09 — Una mutación densa se edita fuera de la tarjeta de resumen

- **Aprendido:** la tarjeta de un partido debe permitir escanear el cruce y su
  marcador; los campos solo son necesarios durante la acción de registrar o
  corregir el resultado.
- **Evidencia:** el detalle abre un popup con los dos marcadores y bloquea esa
  interacción mientras se guarda. Al recibir la proyección actualizada, cierra
  el popup y sustituye la tarjeta sin navegación adicional.
- **Coste aceptado:** el estado del popup y el marcador en edición es local a la
  ruta, sin introducir una store ni una capa de formularios nueva.
- **Regla reutilizable:** separar lectura de edición mejora la densidad de una
  colección y reduce las acciones accidentales en móvil.

### 2026-08-09 — El nombre JSON forma parte del contrato, no de la estructura interna

- **Aprendido:** un campo correctamente leído y almacenado puede seguir siendo
  invisible para el cliente si el serializador usa un nombre distinto al de
  OpenAPI.
- **Evidencia:** la jornada se persistía como `round_number`, pero la respuesta
  emitía `roundNumber` mientras el contrato declaraba `round`; el cliente
  generado recibía por tanto `undefined`.
- **Coste aceptado:** una prueba HTTP adicional verifica el cuerpo JSON de una
  mutación de resultado, además de las pruebas del dominio y persistencia.
- **Regla reutilizable:** las pruebas de borde HTTP deben afirmar los nombres y
  la forma del contrato público, no solo que las estructuras internas contengan
  los valores esperados.

### 2026-08-09 — La propiedad conserva las capacidades operativas delegables

- **Aprendido:** delegar una tarea no debe impedir que la persona propietaria
  pueda realizarla; hacerlo fuerza una asignación artificial incluso cuando
  opera una liga pequeña por sí misma.
- **Evidencia:** la autorización de resultados acepta la organizadora o una
  administradora delegada, mientras que una cuenta ajena sigue siendo rechazada.
- **Coste aceptado:** la comprobación consulta ambas relaciones ya existentes,
  sin crear un rol adicional ni duplicar a la organizadora en la tabla de
  delegaciones.
- **Regla reutilizable:** al introducir delegación directa, conserva de forma
  explícita el permiso en la propiedad salvo que una decisión establezca una
  separación de deberes deliberada.

### 2026-08-09 — La composición se modifica antes de generar el calendario

- **Aprendido:** añadir un inscrito no debe exigir recrear una liga mientras la
  competición todavía no ha empezado.
- **Evidencia:** la nueva operación añade un único equipo, solo para la
  organizadora y únicamente en `published`; el servidor bloquea la liga antes
  de calcular la siguiente posición y rechaza duplicados, más de 64 equipos o
  cualquier estado posterior.
- **Coste aceptado:** se incorpora una mutación específica y un popup local en
  la vista de equipos. La eliminación conserva al menos dos equipos; no se
  implementa edición de la composición en curso ni regeneración de partidos.
- **Regla reutilizable:** cuando una transición genera derivados persistidos,
  concentra el último cambio estructural justo antes de esa transición y hazlo
  cumplir también en el servidor.

### 2026-08-09 — Una clasificación es una proyección, no un formulario

- **Aprendido:** puntos, posiciones y futuras victorias de torneo son derivados
  de resultados persistidos; una interfaz solo puede mostrarlos, nunca
  declararlos.
- **Evidencia:** el caso de uso calcula la tabla tras leer o mutar una liga y el
  contrato entrega filas con estadísticas y posición. La pantalla no contiene
  un comparador ni suma puntos; explica las reglas configuradas. Las pruebas
  cubren grupos empatados de dos, tres y cuatro equipos.
- **Coste aceptado:** se mantiene una función de dominio con pruebas para una y
  dos vueltas, en vez de guardar posiciones que deberían reconciliarse tras una
  corrección de marcador. Antes de valorar materializarla se medirá el caso de
  64 equipos a dos vueltas y 4.032 partidos, tanto al leer como al corregir un
  marcador.
- **Regla reutilizable:** cuando una lectura afectará a logros, historial o
  perfil, deriva su valor de la fuente de verdad del backend antes de presentar
  cualquier métrica al cliente.

### 2026-08-09 — Un límite de dominio necesita feedback antes del rechazo

- **Aprendido:** el servidor mantiene el máximo de 64 equipos, pero la persona
  necesita conocer el motivo antes de intentar una operación que será rechazada.
- **Evidencia:** el mismo límite compartido impide añadir un campo al borrador o
  abrir el diálogo de una liga publicada y muestra un banner localizado.
- **Coste aceptado:** el cliente anticipa el límite visible, sin sustituir la
  comprobación transaccional del backend ante cambios simultáneos.
- **Regla reutilizable:** replica en la interfaz una restricción estable para
  orientar la acción, pero conserva siempre la validación definitiva en el
  servidor.

### 2026-08-09 — El cierre deportivo debe ser atómico y explícito

- **Aprendido:** que el último marcador esté registrado no basta para cerrar una
  liga: la organizadora conserva una revisión final y el servidor debe volver a
  comprobar el estado bajo el mismo bloqueo que persiste el resultado oficial.
- **Evidencia:** `POST /leagues/{leagueId}/complete` rechaza pendientes y
  transiciones inválidas; calcula y guarda todas las posiciones 1 en la misma
  transacción antes de fijar `completed`.
- **Regla reutilizable:** cuando una transición congela derivados de negocio,
  no los calcules fuera de la transacción ni aceptes que el cliente los envíe.

### 2026-08-09 — Dos peticiones simultáneas no son dos transiciones válidas

- **Aprendido:** repetir una orden de cierre por un doble toque o reintento de
  red no debe duplicar la transición ni sus derivados oficiales.
- **Evidencia:** la integración lanza dos cierres a la vez: el bloqueo de fila
  permite que uno complete la liga y obliga al otro a releer `completed` y
  devolver conflicto. La tabla de campeones queda con una sola fila para una
  liga con ganadora única.
- **Regla reutilizable:** protege las transiciones de estado con un bloqueo de
  la fuente de verdad, vuelve a validar dentro de la transacción y prueba el
  resultado de peticiones competidoras contra una base real.

### 2026-08-09 — Un conflicto conocido debe recuperar el estado, no parecer un fallo

- **Aprendido:** un `409` de cierre simultáneo informa de que otra petición ya
  alcanzó el resultado deseado; tratarlo como error genérico hace que la
  interfaz conserve un estado obsoleto.
- **Evidencia:** la pantalla de liga reconoce solo el `409` del contrato de
  finalización, cierra su confirmación, vuelve a leer la liga y muestra un
  mensaje localizado. Los demás rechazos conservan el fallback seguro común.
- **Regla reutilizable:** mapea cada estado de negocio declarado que habilite
  una recuperación concreta en la feature dueña de la acción; nunca expongas
  el cuerpo de error del backend.

### 2026-08-09 — Un diálogo que confirma una pantalla nativa pertenece a esa pantalla

- **Aprendido:** en iOS una confirmación global puede quedar detrás de una ruta
  presentada por navegación nativa, aunque su estado React se actualice.
- **Evidencia:** la acción de finalizar era visible y recibía el toque, pero el
  diálogo compartido no se presentaba sobre `fullScreenModal`. La confirmación
  se movió a un `Modal` de la propia ruta y se validó en iPhone Simulator.
- **Regla reutilizable:** cuando una mutación crítica se confirma dentro de una
  ruta nativa, su modal debe compartir el host de esa ruta o verificarse en el
  simulador antes de dar el flujo por utilizable.

### 2026-08-10 — Si el destino muestra el resultado, no hace falta un éxito transitorio

- **Aprendido:** después de añadir una administradora, la lista de
  administradoras es una confirmación más clara y persistente que un banner de
  éxito que desaparece.
- **Evidencia:** el formulario cierra sobre el listado y este vuelve a consultar
  el contrato al recibir foco. Antes de buscar, el formulario carga la misma
  colección y filtra sus usernames junto al de la sesión; así no propone de
  nuevo una cuenta ya delegada ni a la propia organizadora. La misma pantalla
  permite retirar una cuenta con la confirmación y la fila compacta ya
  establecidas para equipos.
- **Coste aceptado:** se añaden las operaciones protegidas de listar y retirar
  administradoras al contrato, en lugar de inferir la colección desde la
  relación de la sesión o conservar una copia global en el cliente.
- **Regla reutilizable:** tras una mutación que conduce a una pantalla capaz de
  representar el nuevo estado, vuelve a leer esa proyección y evita un aviso de
  éxito redundante; reserva el banner para recuperaciones, errores o resultados
  que no queden visibles en la ruta destino.

### 2026-08-10 — Una tabla móvil fija su identidad y desplaza solo el detalle

- **Aprendido:** en una clasificación, posición, equipo y puntos son la identidad
  de cada fila; las estadísticas complementan la comparación, pero no deben
  ocultar esos tres datos al desplazar horizontalmente.
- **Evidencia:** la tabla mantiene los dos extremos fijos y desplaza como un único
  bloque `PJ`, `PG`, `PE`, `PP`, `GF`, `GC` y `DG`. El indicador y la ayuda aparecen
  solo cuando ese bloque no cabe en el viewport.
- **Coste aceptado:** se duplican las filas visuales de las tres columnas para
  sincronizar su altura, sin añadir una librería de tabla ni estado compartido.
- **Regla reutilizable:** para datos comparables y densos en móvil, fija la
  identidad de la fila y su métrica principal; reserva el desplazamiento para
  las columnas secundarias y no muestres affordances de scroll si no hay desborde.

### 2026-08-12 — La información de privacidad debe coincidir con el ciclo técnico real

- **Aprendido:** una política no puede prometer recuperación de cuenta si el
  producto solo ofrece una espera antes de su eliminación definitiva.
- **Evidencia:** la baja queda pendiente 30 días y una tarea externa elimina la
  cuenta en lotes; el historial de resultados conserva el dato deportivo y
  anonimiza a su autora.
- **Regla reutilizable:** publica la política antes de recoger datos y enlázala
  desde el acceso, el registro y los ajustes; describe únicamente proveedores y
  tratamientos que estén realmente activos.

### 2026-08-12 — La navegación debe conservar el mismo affordance sin ocultar la plataforma

- **Aprendido:** un control de cabecera sin superficie ni margen lateral parece
  accidental frente a las acciones circulares de las rutas profundas.
- **Evidencia:** web y Android comparten `NavigationHeaderButton`, con objetivo
  de 44 px, borde, superficie y 20 px respecto al lateral; iOS usa su toolbar
  nativa equivalente.
- **Regla reutilizable:** comparte el patrón visual de navegación entre web y
  Android, pero deja que iOS use sus controles de barra cuando la plataforma ya
  ofrece una presentación y un área táctil correctas.

### 2026-08-13 — El blur web debe entrar con el scrim, no después del diálogo

- **Aprendido:** `animationType="fade"` reduce la opacidad del ancestro del
  `Modal`, aislando el grupo de composición y evitando que `backdrop-filter`
  muestree la página mientras entra el diálogo. El blur aparece entonces de
  golpe al final de la animación.
- **Regla reutilizable:** en web, el `Modal` no anima su opacidad. El scrim se
  monta transparente con `blur(0px)` y, tras dos frames de animación, transiciona
  durante 160 ms al oscurecimiento y blur finales. iOS y Android conservan su
  tratamiento nativo.

### 2026-08-12 — Una biblioteca bajo tabs necesita un viewport desplazable explícito

- **Aprendido:** una regla web que localiza la tab bar por
  `[role="tablist"]` no puede asumir que solo existe la navegación principal;
  un control segmentado accesible comparte correctamente ese mismo rol.
- **Evidencia:** el selector global que fijaba cualquier padre de un `tablist`
  convertía toda la biblioteca de torneos en `position: fixed`. La regla ahora
  identifica además la ruta estable de la tab «Torneos», mientras la biblioteca
  mantiene su selector fuera del `ScrollView` y conserva la reserva inferior
  común de la botonera.
- **Regla reutilizable:** en una ruta bajo tabs, el contenedor desplazable recibe
  explícitamente el alto flexible y la reserva inferior de
  `useTabContentBottomPadding`. Una regla global que dependa de semántica ARIA
  debe acotarse a la estructura que pretende modificar; los indicadores de
  navegación de filas se comparten dentro de un contenedor de 24 px, con icono
  al 80 % para que no domine el contenido, no como glifos de texto con tamaño
  accidental.

### 2026-08-12 — Una exportación pública debe aislar configuración y caché locales

- **Aprendido:** Expo incorpora `EXPO_PUBLIC_*` en el bundle y puede reutilizar
  transformaciones de Metro; una exportación pública puede conservar por error
  la URL de API local aunque el comando le proporcione una URL distinta.
- **Evidencia:** el bundle de `dev.fasttourney.com` contenía
  `http://localhost:8080/v1`, lo que activaba el permiso de acceso local de
  Brave. El script de publicación desactiva la carga de `.env`, establece la
  URL HTTPS de la API y del enlace compartido, y limpia la caché antes de
  exportar.
- **Regla reutilizable:** todo script que exporte un artefacto público con
  configuración distinta de la local debe declarar sus variables, impedir
  overrides de `.env` y reconstruir las transformaciones que las incorporan.

### 2026-08-12 — HSTS pertenece a la respuesta HTTPS pública, no al salto interno

- **Aprendido:** Cloudflare termina TLS para la persona visitante aunque el
  túnel llegue por HTTP de loopback a Caddy. Por ello Caddy puede emitir HSTS y
  Cloudflare lo entrega en la respuesta HTTPS pública.
- **Decisión aplicada:** `max-age=31536000`, sin `includeSubDomains` ni
  `preload`; esas extensiones se reservarán hasta poder garantizar HTTPS para
  todos los subdominios presentes y futuros.
- **Regla reutilizable:** validar una cabecera de seguridad tanto en el router
  local como por HTTPS en el hostname público después de recargar el borde.

### 2026-08-12 — Una CSP se despliega primero observando el artefacto real

- **Aprendido:** una política de contenido demasiado genérica puede impedir que
  el bundle web cargue aunque la aplicación compile correctamente. El modo
  `Content-Security-Policy-Report-Only` permite descubrir esos desajustes sin
  interrumpir a las personas usuarias.
- **Evidencia:** inicio, acceso, registro, recuperación, privacidad y enlaces
  de verificación de `dev.fasttourney.com` no registraron violaciones durante
  la observación. La política obligatoria permite el origen propio, la API de
  desarrollo y los endpoints de Google previstos para identidad federada.
- **Regla reutilizable:** añadir un origen a CSP solo tras una ruta funcional
  que lo requiera y comprobar el mismo flujo con la política en modo obligatorio.

### 2026-08-12 — Las cabeceras complementan una CSP, no la duplican

- **Aprendido:** CSP restringe el contenido que una página puede cargar, pero
  no evita que el navegador adivine tipos MIME, filtre rutas completas como
  referencia o habilite capacidades del dispositivo que la web no usa.
- **Decisión aplicada:** el borde añade `nosniff`, una política de referrer
  conservadora y desactiva mediante `Permissions-Policy` capacidades no
  requeridas. `frame-ancestors 'none'` de CSP cubre el anti-embebido, por lo
  que no se añade el encabezado legado `X-Frame-Options`.

### 2026-08-12 — La IP reenviada es un límite de confianza, no un dato del cliente

- **Aprendido:** detrás de Cloudflare Tunnel, Caddy y Docker, el socket de la
  API identifica al proxy interno, no a la persona visitante. Sin una IP
  reenviada fiable, los límites por IP agrupan a todas las personas o se pueden
  falsear si se confía en una cabecera de cualquiera.
- **Decisión aplicada:** Caddy copia `CF-Connecting-IP` en `X-Client-IP`; la
  API solo la usa cuando el peer pertenece a `TRUSTED_PROXY_CIDRS`. Registro se
  limita a cinco solicitudes por minuto e IP; acceso y recuperación conservan
  sus límites existentes usando la misma identidad resuelta.
- **Regla reutilizable:** cada proxy adicional exige revisar la red de confianza
  y probar que una cabecera falsa desde un peer no confiable no cambia la IP
  aplicada por la API.

### 2026-08-12 — Git conserva fuentes; el rollback necesita artefactos locales

- **Aprendido:** un commit permite reconstruir una versión, pero no restaura de
  forma inmediata la combinación exacta de imagen y web que servía el Mac.
- **Decisión aplicada:** dev conserva solo el SHA actual y el anterior, junto
  con una imagen runtime, una exportación estática y un manifiesto sin secretos.
  Caddy selecciona la web mediante un enlace simbólico atómico.
- **Regla reutilizable:** no almacenar artefactos, imágenes o backups en Git;
  asociarlos a un SHA y distinguir rollback de código de restauración de datos.

### 2026-08-12 — Un rollback sin corte requiere dos versiones compatibles

- **Aprendido:** compilar y superar CI no garantiza que una versión funcione
  ante tráfico y configuración reales; una retirada rápida limita el impacto.
- **Dirección a decidir antes de producción:** blue/green detrás de Caddy:
  validar una segunda instancia antes de dirigirle tráfico y conservar la
  anterior mientras se observa la nueva.
- **Regla reutilizable:** una conmutación de procesos no revierte la base de
  datos. Ambas versiones deben coexistir sobre el mismo esquema o el cambio
  necesita una estrategia explícita de evolución y recuperación.

### 2026-08-13 — El contrato HTTP incluye los nombres y el formato de sus campos

- **Aprendido:** que los valores del dominio sean correctos no basta si el
  codificador JSON expone los nombres por defecto de Go o una fecha que no
  cumple RFC 3339. El cliente generado consume los nombres definidos por
  OpenAPI, no los nombres internos del struct.
- **Evidencia:** las notificaciones llegaban como `ID` y `CreatedAt` y la fecha
  procedente de PostgreSQL no era portable para `Date`; React recibía claves
  ausentes y avisaba de elementos sin `key`.
- **Regla reutilizable:** los DTO HTTP declaran explícitamente sus etiquetas
  JSON y normalizan los tiempos a RFC 3339. Una prueba del handler verifica el
  cuerpo serializado cuando una ruta tiene un cliente generado.

### 2026-08-13 — El puerto de la web local es parte de su contrato de enlaces

- **Aprendido:** si Expo cambia automáticamente de puerto porque `8082` está
  ocupado, los enlaces de correo apuntan a otro origen y dejan de abrir la web
  local correcta.
- **Regla reutilizable:** la web local se inicia explícitamente en `8082`; se
  libera ese puerto antes de arrancarla y CORS solo autoriza sus orígenes
  `localhost` y `127.0.0.1` en ese puerto.

### 2026-08-13 — La documentación pública debe viajar con el release que describe

- **Aprendido:** una referencia de API desplegada como proceso local separado
  puede describir un contrato distinto al de la web o API pública actuales.
- **Regla reutilizable:** el despliegue copia la UI de referencia y el OpenAPI
  junto al artefacto web versionado; Caddy sirve esa ruta antes del fallback de
  la SPA para mantenerla consistente y recuperable con el mismo SHA.

### 2026-08-13 — El callback OAuth web debe cargar el cierre antes que su ruta

- **Aprendido:** el callback se recibe en una ventana auxiliar. Si
  `maybeCompleteAuthSession()` vive solo en una pantalla diferida, el popup puede
  montar otra ruta y quedarse abierto como una segunda aplicación.
- **Regla reutilizable:** el layout raíz web completa el popup antes de montar
  Expo Router y el redirect URI es explícito hacia la ruta del recorrido que lo
  inició. La continuación que necesita datos de cuenta se presenta con
  `ModalDialog`, no como una card insertada en la pantalla subyacente.

### 2026-08-13 — Retirar un acceso acredita el que permanece

- **Aprendido:** permitir eliminar el único método de una cuenta convierte una
  sesión válida en un callejón sin salida. Acreditar el método que permanece
  protege además contra la retirada de una recuperación por quien solo controla
  el método que quiere eliminar.
- **Regla reutilizable:** cada ticket de reautenticación se vincula
  criptográficamente a una finalidad y es de un uso; el backend comprueba en la
  misma mutación que sigue existiendo otro método de acceso.

### 2026-08-13 — Una baja deportiva conserva su evidencia, no elimina al equipo

- **Aprendido:** borrar el equipo durante una competición rompería las claves de
  los partidos y ocultaría la causa de una clasificación modificada.
- **Decisión aplicada:** la baja marca al equipo, sustituye en una única
  transacción todos sus partidos por 3-0 para el rival y conserva una entrada de
  historial por cada marcador previo o pendiente.
- **Regla reutilizable:** las mutaciones masivas que cambian resultados deben
  devolver la proyección canónica completa; así el cliente actualiza todas las
  vistas sin reconstruir una clasificación local.

### 2026-08-14 — Una carga terminal necesita una recuperación dentro de la ruta

- **Aprendido:** un banner no basta cuando el error impide montar todo el
  contenido: se desvanece mientras la pantalla permanece aparentemente cargando.
- **Regla reutilizable:** `RequestErrorCard` presenta el mensaje seguro ya
  clasificado y un reintento localizado cuando este puede recuperar la carga.
  Si la feature conoce un `404` terminal, ofrece cierre y conserva el contexto
  de navegación: en móvil descarta el modal y en inicio en frío usa el fallback
  seguro. Las rutas que sí tienen contenido conservan el banner global para
  evitar crear dos patrones de error equivalentes. Mientras esa carga está en
  curso, `LoadingTransition` centra el progreso sobre la ruta; nunca se usa una
  card de «Cargando».

### 2026-08-14 — Un parche de seguridad alcanzable se promociona sin espera

- **Aprendido:** una política de maduración protege frente a regresiones, pero
  no debe retrasar una corrección de seguridad que el análisis alcanza desde el
  código del producto.
- **Regla reutilizable:** un parche crítico autorizado actualiza toolchain,
  imagen de compilación y documentación en el mismo cambio; `make verify` debe
  confirmar el parche antes de promoverlo.

### 2026-08-14 — La transferencia conserva la competición, no la autoridad anterior

- **Aprendido:** cambiar la propiedad de una competición compartida no exige
  reconstruir partidos, resultados ni administraciones; exige serializar y
  autorizar el cambio en la misma transacción que reemplaza a la organizadora.
- **Regla reutilizable:** una transferencia inmediata elimina expresamente los
  privilegios de la persona anterior y limpia la delegación de la destinataria
  si ya existía. Cuando la acción cierra un modal, `showAfterNavigation()`
  espera al host de la `Screen` destino antes de presentar su confirmación.

### 2026-08-14 — La identidad tipográfica requiere los pesos reales

- **Aprendido:** un token que apunta a la fuente del sistema no garantiza la
  misma familia entre web, iOS y Android. Además, aplicar `fontWeight` sobre un
  archivo regular puede sintetizar un peso y cambiar métricas o legibilidad.
- **Regla reutilizable:** una fuente compartida se carga localmente antes del
  primer render y cada variante tipográfica selecciona su asset real mediante
  tokens; cabeceras y navegación se revisan porque no necesariamente pasan por
  la primitiva de texto.

### 2026-08-16 — Un botón destructivo de contorno debe conservar contraste semántico

- **Aprendido:** el texto blanco sobre una superficie transparente se pierde en
  el tema claro, aunque el borde use un color de estado.
- **Regla reutilizable:** `destructive` combina borde y texto de error sobre la
  superficie existente, tanto en menús como en confirmaciones; no se mantiene
  una variante destructiva rellena que compita visualmente con ella.

### 2026-08-16 — El material nativo depende de la versión del sistema

- **Aprendido:** una barra de navegación nativa puede aplicar transparencia o
  material aunque el sistema no ofrezca el tratamiento visual para el que fue
  diseñada, y entonces pierde contraste contra el contenido.
- **Regla reutilizable:** Liquid Glass se conserva solo en iOS 26 o superior.
  En versiones anteriores la barra fija un fondo opaco con el token
  `surface.default`, igual que las implementaciones web y Android.

### 2026-08-16 — El inset de la cabecera también forma parte del control

- **Aprendido:** un botón circular compartido no debe sumar un margen propio en
  iOS anterior a 26: React Navigation ya reserva el inset lateral de la barra.
- **Regla reutilizable:** en iOS anterior a 26, `NavigationHeaderButton` usa el
  mismo objetivo de 44 px, borde y superficie que web y Android, pero deja la
  alineación lateral a la cabecera. Así coincide con los controles de toolbar
  de iOS 26 y el patrón cubre también acciones como Notificaciones.

### 2026-08-16 — El loader expresa contraste, no jerarquía de texto

- **Aprendido:** reutilizar `text.primary` para un indicador de progreso lo
  convierte en negro sobre claro y no garantiza su contraste en oscuro.
- **Regla reutilizable:** el token semántico `indicator.default` usa el azul
  principal sobre superficies claras y blanco sobre superficies oscuras; un
  loader dentro de una acción primaria filled sigue usando blanco en ambos
  temas porque su contexto es el degradado de marca.

### 2026-08-16 — Rotar credenciales web exige distinguir persistencia y entrega

- **Aprendido:** una sesión válida en el servidor no restaura una web si su
  cookie era de sesión y Safari descartó el proceso. Access y refresh deben
  persistir con sus vencimientos propios, permanecer `HttpOnly` y rotarse juntos.
- **Regla reutilizable:** el cliente web comparte una sola renovación tras un
  `401` de access y repite la petición una vez. No se refresca durante unload:
  si se pierde la respuesta de una rotación estricta, se exige login antes que
  aceptar una ventana de gracia que rebaje la detección de reuso.

### 2026-08-17 — Una señal solo es útil si permite seguir el mismo fallo

- **Aprendido:** logs, métricas y trazas no sustituyen una investigación si
  cada una usa identificadores y dimensiones distintos. Un primer flujo pequeño
  —el refresh de sesión— permite demostrar la correlación HTTP → PostgreSQL
  antes de diseñar paneles o alertas generales.
- **Regla reutilizable:** los logs registran plantilla de ruta, estado y
  `trace_id`; Prometheus usa solamente etiquetas acotadas; Tempo no almacena
  SQL, argumentos, cookies, tokens ni query strings. La instrumentación usa una
  copia saneada de la URL para telemetría y entrega la URL original al handler,
  de modo que las consultas siguen funcionando sin exponer sus valores.

### 2026-08-20 — Un span debe nombrar el trabajo, no la librería

- **Aprendido:** `HTTP server` y `postgresql.query` prueban que hubo actividad,
  pero obligan a expandir atributos para entender una traza y ocultan el coste
  de las operaciones de CPU entre llamadas a base de datos.
- **Regla reutilizable:** el nombre visible se deriva de una ruta plantilla o
  de un identificador estático de sqlc, nunca de datos de la petición o SQL. Un
  decorador de infraestructura mide `auth.password.verify` sin introducir SDKs
  de observabilidad en el caso de uso de registro.

### 2026-08-20 — Los límites locales y externos se decoran desde infraestructura

- **Aprendido:** registrar una cuenta concentra trabajo Argon2id, una
  transacción PostgreSQL y una entrega SMTP. Los tres son límites que pueden
  explicar latencia o fallos; la validación y la generación local de tokens no
  añaden una señal equivalente.
- **Regla reutilizable:** el caso de uso conserva puertos de aplicación y los
  adaptadores técnicos los decoran con spans sin atributos sensibles. Las
  consultas manuales reciben un identificador estático, igual que las
  generadas por sqlc. Una operación SMTP distingue solo el propósito técnico
  —verificación o restablecimiento—, nunca la persona destinataria. El puerto
  de Argon2id cubre tanto el alta como cualquier cambio de credencial.

### 2026-08-20 — Una consulta atómica conserva un único límite observable

- **Aprendido:** verificar un registro consume el token, confirma la cuenta y
  crea la sesión dentro de `VerifyRegistrationAndCreateSession`. Aunque el SQL
  use varios CTE, PostgreSQL lo ejecuta como una operación del adaptador.
- **Regla reutilizable:** no se fragmenta una operación atómica en spans que no
  existen en el proceso. SHA-256 y la generación de secretos son coste local
  despreciable frente a los límites HTTP y PostgreSQL; medirlos añadiría ruido.

### 2026-08-20 — Las consultas manuales también necesitan un nombre estable

- **Aprendido:** `sqlc` conserva su identificador estático en el SQL generado,
  pero una consulta escrita directamente con `pgx` no lo adquiere por sí sola.
- **Regla reutilizable:** la anotación `-- name:` precede a una consulta manual
  y permite nombrar su span sin exportar SQL, argumentos ni términos de
  búsqueda.

### 2026-08-20 — El error operativo se clasifica; no se copia

- **Aprendido:** un error de driver o SMTP puede contener detalle interno o
  datos proporcionados por una persona. `RecordError` lo convertiría en un
  evento de traza compartido con Grafana.
- **Regla reutilizable:** los límites técnicos registran una causa cerrada y
  segura en `tournaments_manager.failure.reason`, mantienen un resumen fijo y
  no exportan el error bruto. La causa de negocio se añade solo en la feature
  que puede usarla para diagnosticar o recuperar.

### 2026-08-20 — Una ruta se revisa por sus salidas, no solo por su recorrido feliz

- **Aprendido:** una traza nombrada no explica por qué una persona recibió un
  rechazo esperado. Validación, límite de tasa y reglas de negocio son salidas
  distintas de un mismo endpoint, aunque no sean fallos de infraestructura.
- **Regla reutilizable:** el span HTTP raíz recibe una causa cerrada y segura
  cuando esa salida aporte diagnóstico; no se crea un span por rama ni se
  registran valores introducidos. La feature mantiene la decisión de las
  causas de negocio, igual que mantiene su feedback de recuperación.

### 2026-08-21 — Un SLO convierte señales en una decisión operativa

- **Aprendido:** una métrica de latencia o errores no indica por sí sola cuándo actuar. Un objetivo separa la degradación tolerable del consumo de un margen acordado; por eso disponibilidad y p95 se expresan por separado.
- **Regla reutilizable:** empezar por un flujo crítico y por métricas ya agregadas. Un dashboard local y alertas visibles en Prometheus bastan para aprender el ciclo; Alertmanager y notificaciones se incorporan solo al aparecer una necesidad operativa explícita (ADR-0099). Los SLOs globales siguen requiriendo evidencia posterior.

### 2026-08-21 — Evaluar una alerta no es entregarla

- **Aprendido:** Prometheus determina si una condición está activa, pero no es el lugar para decidir agrupación, silencios o la frecuencia de recordatorios.
- **Regla reutilizable:** Alertmanager recibe alertas evaluadas, las agrupa y entrega según la severidad. Grafana puede centralizar la consulta y los silencios sin convertirse en la autoridad de las reglas Prometheus.

### 2026-08-21 — Una colección remota se valida elemento a elemento

- **Aprendido:** los tipos generados desde OpenAPI desaparecen en runtime; un
  miembro malformado de una respuesta JSON no debe inutilizar los elementos
  válidos que la acompañan.
- **Regla reutilizable:** el adaptador de cada feature valida el contenedor de
  una colección y transforma individualmente sus miembros. Rechaza el
  contenedor imposible como respuesta no esperada y descarta únicamente los
  miembros inválidos antes de entregarlos a la interfaz. No se modifica el
  cliente generado ni se introduce una validación global de reglas de negocio.

### 2026-08-22 — El consentimiento puede prepararse sin adelantar la telemetría

- **Aprendido:** una preferencia de analítica no necesita activar el proveedor
  que regulará en el futuro. Un provider local compartido puede partir de
  `false`, persistir la elección por plataforma y permitir que Inicio y Ajustes
  expresen el mismo control sin empezar a capturar datos.
- **Regla reutilizable:** la primera aceptación puede retirar una invitación
  contextual de la home y confirmar dónde se revoca; Ajustes conserva la misma
  card y switch para cualquier cambio posterior. La preferencia local no es
  evidencia de consentimiento para un proveedor hasta que ese proveedor se
  integre y se revise su configuración efectiva.
