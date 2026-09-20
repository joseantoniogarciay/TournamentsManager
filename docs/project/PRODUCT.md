# Producto

> Estado: producto v1 cerrado; capacidades posteriores se gestionan como
> incrementos independientes.
>
> Última actualización: 2026-09-19

## Visión

TournamentsManager permite descubrir, crear y gestionar torneos. Fútbol fue el
primer deporte con el que se validó el modelo y baloncesto es el primer perfil
adicional; los siguientes se incorporarán solo cuando sus reglas estén definidas.

## Hechos aceptados

- Existen targets web, iOS y Android en el cliente universal.
- Web, iOS y Android ofrecen el mismo producto con paridad funcional.
- La experiencia es responsive en navegadores y aplicaciones de móvil, tablet y
  escritorio, adaptando la presentación cuando corresponda.
- El producto se orienta a torneos entre amistades, clubes o grupos cerrados,
  con lectura pública por identificador y gestión autenticada.
- Una persona invitada tiene acceso a la home, al borrador local y a torneos
  visibles cuyo identificador conoce; eso no concede permisos de gestión.
- Preparar un borrador de torneo no exige una cuenta; persistirlo y publicar el
  torneo exige una cuenta verificada.
- La cuenta incluye registro, inicio de sesión y recuperación de contraseña.
- Una misma cuenta admite credenciales locales y login con Google. Apple no está
  adoptado en v1 y se revisa antes de distribuir el cliente iOS.
- El torneo declara un deporte inmutable elegido entre `football` y
  `basketball`; «Fútbol» engloba fútbol sala en esta iteración (ADR-0126).
- El recurso raíz es un torneo. La liga de fútbol existente se conserva como
  formato `league`; también existe la eliminatoria directa a partido
  único como formato `single_elimination`. Los formatos mixtos quedan fuera
  hasta acordar sus reglas. Véase ADR-0122.
- Un torneo contiene fases ordenadas. Cada torneo crea una sola fase, elegida
  entre liga a una o dos vueltas y eliminatoria directa a partido único. Una
  futura liga única o por grupos podrá clasificar equipos para un cuadro
  posterior cuando se acepten sus reglas. Véase ADR-0123.
- El creador conserva la propiedad, crea el torneo con su propio equipo y puede
  añadir equipos sin cuenta o invitar a otras personas para que inscriban el
  suyo antes del inicio; puede además asignar o retirar administradores
  delegados y transferir la propiedad bajo las reglas aceptadas (ADR-0130).
- La creación, equipos, inicio, resultados, retirada, cancelación, finalización,
  seguimiento, administración y transferencia están definidas en el contrato
  vigente.
- Se ha aceptado un cliente universal con React Native, Expo, Expo Router, CNG y
  rendering web client-side inicial; la home propia es la única superficie web
  indexable, sin convertir las rutas de aplicación en catálogo (ADR-0120).

## Actores y capacidades iniciales

### Invitado

- preparar un borrador local de torneo y sus equipos;
- ver únicamente las superficies que se definan como visibles sin cuenta;
- iniciar registro o login antes de persistir o publicar el torneo, o cuando
  intente una acción protegida.

Un invitado no es una cuenta con rol especial: es una persona sin sesión
autenticada.

### Cuenta pendiente de verificación

Es un registro temporal tras un alta local con email, contraseña y `username`.
No recibe una sesión de producto ni permisos de negocio. Un borrador completo
permanece local hasta enviar el alta; entonces crea un torneo publicado asociado a
la cuenta pendiente, que no puede administrar ni listar hasta verificarla.

### Usuario autenticado y verificado

- gestionar su sesión y perfil básico;
- usar su `username` público y único, que no puede cambiar en v1;
- crear un torneo;
- consultar los torneos con los que tiene relación.

## Home y biblioteca de torneos

La ruta `/` es la home. La botonera tiene «Inicio» como primera posición,
«Torneos» como segunda y «Cuenta» como tercera; Cuenta conserva su propio flujo.
En iOS 26 la barra usa el componente nativo con efecto Liquid Glass y se
superpone al contenido, que conserva margen inferior para permanecer accesible.
Sin sesión, la home explica el producto y ofrece explorar sin cuenta,
registrarse o iniciar sesión; explorar inicia o retoma el borrador local. La
misma ruta `/` es la única superficie indexable inicial en web, mientras rutas
de aplicación y ligas quedan fuera de buscadores (ADR-0120). Con sesión
verificada, la home también ofrece accesos rápidos a «Administro» y
«Guardados». Las ligas creadas por la cuenta y aquellas donde es administrador
delegado se consideran administradas. Las ligas seguidas se consideran
guardadas; cuando una liga cumple ambas relaciones, se muestra solo como
administrada.

Con sesión, Inicio muestra además hasta tres ligas relacionadas con actividad
reciente. La relación administrada prevalece sobre seguida si coinciden. Si no
hay ninguna, explica que ahí aparecerán las últimas ligas que tengan actividad.
Inicio vuelve a consultar esta proyección cada vez que recupera el foco, de modo
que una liga recién creada aparece al regresar; Inicio y Torneos permiten además
actualizar sus datos mediante pull-to-refresh para una sesión activa.

Debajo de Actividad reciente, una cuenta autenticada puede enviar una sugerencia
privada de entre 8 y 1.000 caracteres desde «¿Te falta algo?». El envío se activa
solo con una entrada válida. Al guardarse, el campo se vacía y un banner agradece
la aportación; un fallo conserva el texto para reintentar. Se admiten tres envíos
por cuenta y hora. La primera versión no publica sugerencias ni permite
respuestas: PostgreSQL conserva el registro y el correo solo avisa al responsable.

Justo debajo, la home autenticada web puede ofrecer «Apoya el desarrollo» con
propinas únicas visibles de 2 €, 5 € y 10 €. La sección solo aparece cuando los
tres Payment Links de Stripe están configurados; abre su checkout alojado y
declara que no hay contraprestación ni deducción fiscal. iOS y Android no la
renderizan ni reciben esos enlaces, conforme a [ADR-0129](../adr/0129-accept-voluntary-developer-tips-with-platform-appropriate-payments.md).
La interfaz de FastTourney se localiza, pero el nombre y la descripción del
catálogo de Stripe se mantienen en inglés para evitar productos y enlaces por
idioma.

La sección «Torneos» separa las colecciones completas en «Administro» y «Sigo».
Es una clasificación de navegación: las autorizaciones continúan verificándose
en el backend para cada liga y acción. La colección autenticada se define en
[ADR-0058](../adr/0058-list-account-related-leagues-with-a-paginated-collection.md)
antes de mostrar estas listas con datos reales.

En iOS y Android, cada sección de la botonera conserva su pila mientras la app
está activa. En web, son accesos directos a páginas con URL canónica e historial
del navegador. Véase [ADR-0057](../adr/0057-define-contextual-home-and-tournament-library.md).

### Organizador

Es inicialmente el usuario autenticado que creó el torneo, con permisos sobre él
y capacidad de crear equipos y gestionar resultados. El primer equipo es el
suyo y queda vinculado a su cuenta. Mientras el torneo no haya empezado puede
añadir otros equipos sin cuenta asociada o compartir una invitación para que
cada persona cree el suyo. Conserva la propiedad y es el único que puede asignar
o retirar administradores delegados. Puede transferir el torneo a otra cuenta
verificada mediante su `username`; la transferencia es inmediata y le retira
sus permisos administrativos.

### Administrador delegado

El creador lo asigna directamente mediante su `username`, sin aceptación previa.
El administrador puede abandonar la liga; el creador puede retirarlo con efecto
inmediato. Su único permiso operativo es gestionar resultados; la mecánica de
registro queda limitada al tanteo final local y visitante, ambos enteros no
negativos, y a los penaltis solo cuando una eliminatoria de fútbol empatada los
necesita.
Un resultado que registre se aplica de inmediato, sin confirmación del creador;
también puede corregirlo y el sistema conserva quién cambió qué y cuándo.

### Seguidor

Un usuario autenticado y verificado puede guardar una liga consultada mediante
enlace para recuperarla en «ligas seguidas». Seguir no concede permisos ni crea
participación deportiva. Inscribir un equipo mediante invitación crea también
este seguimiento para que el torneo aparezca inmediatamente en «Sigo».

### Participante

Es un equipo. Puede haberlo creado directamente la organizadora sin asociarlo a
una cuenta, o puede haberlo inscrito una cuenta verificada mediante invitación.
La cuenta vinculada representa a ese único equipo dentro del torneo, pero no se
convierte en administradora ni recibe permiso para gestionar resultados. No se
modelan jugadores ni varias personas por equipo.

## Explicación de la composición antes del inicio

Este incremento está aceptado e implementado conforme a ADR-0130; el texto
siguiente define cómo se explica en el cliente.

La creación no presenta una lista larga de equipos como requisito. Pide un único
campo obligatorio bajo el título «Tu equipo» y explica: «Empieza con el equipo
con el que participas. Podrás completar el torneo después». Si no existe un
borrador, el campo se prerrellena con el último nombre de equipo confirmado en
ese dispositivo y sigue siendo editable; un borrador conservado siempre tiene
precedencia.

Tras crear el torneo, la superficie de equipos presenta las dos opciones con el
siguiente mensaje base:

> Completa los equipos
>
> Puedes añadir equipos sin cuenta o compartir un enlace para que cada persona
> cree el suyo. Podrás usar las dos opciones hasta que comience el torneo.

Las acciones se nombran «Añadir equipo» y «Compartir invitación». La interfaz
evita usar solo «anónimo», porque podría entenderse como ocultación de identidad;
«sin cuenta» explica la diferencia real. Junto a la acción de iniciar se recuerda
que hacen falta al menos dos equipos y que, después de comenzar, la composición
queda cerrada.

Mientras el torneo no haya empezado, el detalle de la organizadora despliega
directamente la gestión de equipos en lugar de exigir que abra otra pantalla. La
acción «Iniciar torneo» permanece deshabilitada hasta alcanzar dos equipos, pero
la gestión no se compacta al cumplir ese mínimo: sigue visible para completar la
composición prevista. Una inscripción recibida por invitación aparece al volver
a cargar el torneo.

Quien abre la invitación ve el torneo al que se incorpora, un campo «Nombre de tu
equipo» y esta consecuencia antes de confirmar: «Tu equipo se añadirá al torneo
y lo encontrarás en Sigo. No recibirás permisos de administración». El campo se
prerrellena con el último nombre confirmado en ese dispositivo y sigue siendo
editable. Tanto una creación como una inscripción confirmadas actualizan ese
único valor local para el siguiente formulario de equipo.

## Flujos de identidad

### Registro

1. La persona proporciona correo electrónico, contraseña y `username` público.
2. Acepta las condiciones necesarias.
3. Se crea una cuenta pendiente de verificación y, si el borrador es completo,
   se transfiere junto al alta y queda asociado a esa cuenta.
4. Verifica la propiedad del correo mediante el canal enviado.
5. La cuenta se activa, puede iniciar sesión y puede publicar el torneo.

Si una cuenta pendiente inicia sesión con contraseña correcta, el sistema
invalida su enlace anterior y envía otro correo de verificación; no crea sesión
hasta que se complete esa verificación.

Un borrador local permite empezar sin sesión, pero no es requisito para crear la
cuenta. Si se transfiere al alta, se recupera tras verificar o iniciar sesión y
se descarta localmente después de la aceptación del registro. Una cuenta pendiente
no puede publicar ni realizar acciones protegidas. Véase
[ADR-0078](../adr/0078-transfer-local-drafts-with-registration.md).

### Login

1. La persona demuestra su identidad mediante contraseña o Google.
2. Si existe un borrador local completo, el cliente puede incluirlo en la misma operación.
3. El backend crea atómicamente la sesión y, cuando se envió, el torneo con sus equipos.
4. Repetir la operación con el mismo borrador y cuenta reutiliza el torneo ya creado.
5. El cliente descarta el borrador local solo al confirmar el éxito.
6. El backend autoriza cada acción posterior sobre recursos concretos.

Una cuenta local pendiente conserva el `202` sin sesión y no transfiere el
borrador mediante login. La garantía transaccional de acceso y torneo está
aceptada en [ADR-0127](../adr/0127-create-tournament-atomically-with-login-session.md).

Google es un método de acceso vinculado al mismo usuario interno. Añadir o cambiar
un email de contacto no reemplaza el vínculo con el proveedor.

Si una identidad Google nueva declara un email que ya pertenece a otra cuenta,
el acceso se deniega sin revelar el método existente ni enviar un enlace de
vinculación. La persona inicia sesión con su método habitual y, desde
`Cuenta > Seguridad`, puede añadir Google tras una reautenticación reciente. No
se fusionan cuentas ni se mueven identidades entre usuarios. Véanse
[ADR-0066](../adr/0066-deny-cross-account-identity-linking-and-merges.md) y
[ADR-0067](../adr/0067-allow-authenticated-same-account-access-method-linking.md).

### Recuperación de contraseña

“Recordar contraseña” se implementa como recuperación segura: nunca se recupera
la contraseña anterior. El backend emite por email un token temporal de un solo
uso y permite establecer una nueva, revocando las sesiones anteriores.

Mantener una sesión abierta (“recuérdame”) es una decisión diferente sobre
duración y renovación de sesiones.

## Primer incremento backend cerrado

El primer incremento atravesó producto, seguridad, datos, API y operación:

1. un invitado prepara localmente un torneo, elige fútbol o baloncesto y crea sus
   equipos;
2. una persona se registra, verifica su cuenta e inicia sesión sin perder el
   borrador;
3. el organizador autenticado persiste y publica el torneo con los datos mínimos;
4. el organizador consulta el estado actualizado del torneo y sus equipos;
5. una persona inicia sesión con Google y recibe la misma clase de sesión propia
   que con contraseña.

Los incrementos posteriores completaron registro y corrección de marcadores,
clasificación calculada en backend, retirada de equipos, administración
delegada, transferencia, notificaciones, seguridad de cuenta y los formatos y
deportes descritos en este documento.

El alcance está aceptado en [ADR-0043](../adr/0043-deliver-publish-and-read-league-first-backend-increment.md).
El Gate 0B está cerrado: el formato, los datos mínimos, el ciclo de vida, la
visibilidad, los participantes, la administración, los resultados, las bajas, la
cancelación y la frontera de identidad están definidos. Las capacidades aplazadas
no bloquean el primer vertical slice.

## Reglas deportivas soportadas

Las [ADR-0032](../adr/0032-define-minimum-football-league-data-and-lifecycle.md)
y [ADR-0040](../adr/0040-make-published-leagues-editable-until-start.md) definen
la estructura mínima que originó la liga de fútbol. ADR-0126 amplía ese núcleo:
el torneo elige `football` o `basketball` al crearse y conserva el deporte
durante todo su ciclo. Ambos admiten liga a una o dos vueltas y eliminatoria
directa a partido único; no incluyen fechas, horas, periodos ni prórrogas
desglosadas.

En fútbol, una liga puntúa 3-1-0. En dos vueltas prioriza la
mini-clasificación entre empatados y en una, diferencia de goles y goles a favor
generales. En baloncesto, la victoria suma 2, una derrota jugada 1 y una derrota
administrativa 0; no admite tanteo final empatado y desempata primero por la
mini-clasificación directa, seguida de diferencia de puntos y puntos anotados
generales. Una igualdad tras todos los criterios comparte posición. El backend
calcula la clasificación y la app solo presenta esa proyección (ADR-0081).
«Clasificación» no se muestra antes de iniciar el torneo; durante esa preparación,
el acceso disponible es «Equipos».

El ciclo persistido es `publicado → en_curso → finalizado`, con `cancelado` como
estado terminal desde `publicado` o `en_curso`. El borrador se prepara localmente
o queda asociado temporalmente a una cuenta pendiente; se descarta y no forma
parte de este ciclo.

Cuando todos los partidos están resueltos, solo la organizadora puede finalizar
explícitamente la liga. El backend conserva todos los equipos de la posición 1
como co-campeones y la app muestra el resultado final antes de llevar a la
clasificación. Una liga finalizada ya no admite marcadores ni correcciones.

Una liga visible es consultable sin sesión por su ID público. Mientras permanece
«Sin empezar», la organizadora puede modificar sus datos estructurales, añadir o
eliminar equipos sin cuenta y gestionar el enlace con el que otras cuentas
inscriben el suyo. Al iniciarla el creador elige una o dos vueltas, se exige un
mínimo de dos equipos, se validan los datos, se generan una sola vez los
emparejamientos y se congelan equipos y reglas. Solo entonces la organizadora y
los administradores delegados pueden registrar o corregir resultados. El creador
solo puede finalizarla cuando todos sus partidos tienen resultado. Si un equipo abandona en
`en_curso`, solo el creador puede declararlo: todos sus partidos, pendientes o
ya jugados, pasan a resultado administrativo fijo a favor del rival (`3-0` en
fútbol y `20-0` en baloncesto) y la liga continúa. El valor no es configurable.

## Cancelación

Solo el creador puede cancelar una liga desde `publicado` o `en_curso`. La
pantalla de detalle solicita confirmación explícita con el diálogo compartido
del sistema de diseño antes de ejecutar esta acción destructiva. No se exige
motivo; la liga conserva sus datos y su URL pública muestra estado `cancelado`.
Las personas que la siguen verán ese estado al volver a «ligas seguidas». Este
corte no envía email ni notificaciones push.

`cancelado` es terminal en el corte actual: no admite registrar ni corregir
resultados, ni finalizar la liga. Una futura restauración por el creador queda
fuera del alcance presente y requerirá una decisión e implementación posteriores.
Los lugares que ya muestran el estado —detalle y cajas de acceso a la liga—
reflejan «Liga cancelada» usando su presentación existente; no se añade una
etiqueta nueva.

## Visibilidad inicial

La [ADR-0049](../adr/0049-use-public-league-ids-for-read-only-access.md)
establece que una liga visible se puede consultar sin sesión mediante su ID
público. Esto incluye estados `publicado`, `en_curso`, `finalizado` y
`cancelado`. Conocer el ID concede solo lectura; no crea una relación de
participante ni permisos de administración o resultados.

Los borradores no son accesibles por ID. “Crear y publicar” es una comodidad de
interfaz que ejecuta la misma validación y transición que publicar un borrador.
La invitación de ADR-0130 es una capacidad separada: una cuenta verificada que
posee su secreto puede inscribir un equipo mientras el torneo no haya empezado.
No restringe la audiencia de lectura ni concede otra mutación.

## Fuera del alcance de v1

Salvo decisión posterior:

- pagos por funciones, suscripciones y premios; las propinas voluntarias sin
  contraprestación se rigen por ADR-0129;
- streaming o contenido multimedia;
- chat;
- marketplace;
- microservicios;
- motor genérico para todos los deportes;
- rankings globales;
- administración avanzada;
- calendario completo y arbitraje;
- notificaciones push;
- formatos mixtos con varias fases;
- email y push para avisar de asignaciones administrativas;
- invitaciones con aceptación, bloqueo y controles antiabuso para asignaciones;
- jugadores, plantillas y varias personas asociadas a un mismo equipo.

## Gate 0B

El primer vertical slice tiene definidos formato, ciclo de vida, visibilidad,
participantes, seguimiento, administración y resultados. Las mejoras aplazadas
se mantienen en «Fuera del primer alcance»; no bloquean el esquema ni los
contratos del primer corte.

# Seguridad de la cuenta

Una cuenta autenticada puede consultar su email, username y métodos de acceso,
añadir o cambiar una contraseña y vincular Google tras reautenticarse. Puede
retirar una contraseña o Google solo si acredita el otro método y la cuenta
conserva al menos uno; no puede fusionar cuentas. Cerrar sesión elimina el estado local de
inmediato e intenta la revocación remota sin bloquear la navegación. Puede
solicitar una baja lógica de 30 días: se invalidan de inmediato sus sesiones y
se retiran sus seguimientos y administraciones delegadas. Si organiza alguna
liga, debe cancelarla o transferirla antes. La transferencia directa se define
en ADR-0095.
