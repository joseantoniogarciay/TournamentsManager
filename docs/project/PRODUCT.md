# Producto

> Estado: Gate 0B cerrado; primer vertical slice definido.
>
> Última actualización: 2026-07-26

## Visión

TournamentsManager permitirá descubrir, crear y gestionar torneos. El fútbol será
el primer deporte para aprender y validar el modelo, pero el producto aspira a
incorporar otros deportes sin reescribir su núcleo.

## Hechos aceptados

- Existirán aplicaciones web y mobile.
- Web, iOS y Android ofrecerán el mismo producto con paridad funcional.
- La experiencia será responsive en navegadores y aplicaciones de móvil, tablet y
  escritorio, adaptando la presentación cuando corresponda.
- El producto inicial se orienta a torneos privados entre amistades, clubes o
  grupos cerrados.
- Una persona invitada tendrá acceso limitado; la visibilidad pública de torneos
  se decidirá más adelante.
- Preparar un borrador de torneo no exigirá una cuenta; persistirlo y publicar el
  torneo exigirá una cuenta verificada.
- La cuenta incluirá registro, inicio de sesión y recuperación de contraseña.
- Una misma cuenta admitirá credenciales locales y login con Google; Apple se
  incorporará en un incremento posterior.
- El fútbol es el deporte inicial.
- El formato inicial será una liga de fútbol. Eliminatorias, formatos mixtos y
  otros deportes quedan fuera del primer corte, sin impedir evaluarlos después.
- El creador del torneo será inicialmente su único organizador y creará los
  equipos. Delegar administración se decidirá para una iteración posterior.
- Las acciones detalladas de creación y gestión de un torneo se definirán de forma
  incremental.
- Se ha aceptado un cliente universal con React Native, Expo, Expo Router, CNG y
  rendering web client-side inicial.

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
permanece local hasta enviar el alta; entonces crea una liga publicada asociada a
la cuenta pendiente, que no puede administrar ni listar hasta verificarla.

### Usuario autenticado y verificado

- gestionar su sesión y perfil básico;
- usar su `username` público y único, que no podrá cambiar inicialmente;
- crear un torneo;
- consultar los torneos con los que tiene relación.

## Home y biblioteca de torneos

La ruta `/` es la home. La botonera tiene «Inicio» como primera posición,
«Torneos» como segunda y «Cuenta» como tercera; Cuenta conserva su propio flujo.
En iOS 26 la barra usa el componente nativo con efecto Liquid Glass y se
superpone al contenido, que conserva margen inferior para permanecer accesible.
Sin sesión, crear inicia o retoma el borrador local; con
sesión verificada, la home también ofrece accesos rápidos a «Administro» y
«Guardados». Las ligas creadas por la cuenta y aquellas donde es administrador
delegado se consideran administradas. Las ligas seguidas se consideran
guardadas; cuando una liga cumple ambas relaciones, se muestra solo como
administrada.

Con sesión, Inicio muestra además hasta cinco ligas relacionadas con actividad
reciente. La relación administrada prevalece sobre seguida si coinciden. Si no
hay ninguna, explica que ahí aparecerán las últimas ligas que tengan actividad.
Inicio y Torneos se actualizan mediante pull-to-refresh para una sesión activa.

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
y capacidad de crear sus equipos y gestionar resultados. Conserva la propiedad y
es el único que puede asignar o retirar administradores delegados.

### Administrador delegado

El creador lo asigna directamente mediante su `username`, sin aceptación previa.
El administrador puede abandonar la liga; el creador puede retirarlo con efecto
inmediato. Su único permiso operativo será gestionar resultados; la mecánica de
registro queda limitada a goles locales y visitantes, ambos enteros no negativos.
Un resultado que registre se aplica de inmediato, sin confirmación del creador;
también puede corregirlo y el sistema conserva quién cambió qué y cuándo.

### Seguidor

Un usuario autenticado y verificado puede guardar una liga consultada mediante
enlace para recuperarla en «ligas seguidas». Seguir no concede permisos ni crea
participación deportiva.

### Participante

Es exclusivamente un equipo creado por el organizador. Las personas no se unen a
equipos ni a la competición en el primer corte.

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
2. El cliente obtiene una sesión apropiada para web o mobile.
3. El backend autoriza cada acción sobre recursos concretos.

Google es un método de acceso vinculado al mismo usuario interno. Añadir o cambiar
un email de contacto no reemplaza el vínculo con el proveedor.

Si el primer login con Google coincide con una cuenta local todavía no vinculada, el
acceso queda pendiente. Se envía un enlace de un solo uso al correo verificado y
no se crea una sesión hasta completar la vinculación.

El enlace abrirá la aplicación instalada mediante deep linking o, en su defecto,
la web. Mostrará una confirmación explícita antes de vincular la cuenta. Tras
validarlo se establece la nueva sesión; si el dispositivo tenía otra sesión
activa, el cliente cambia automáticamente a la cuenta recién validada y termina
en la home. Un enlace inválido o caducado muestra el error y la recuperación
posible.

### Recuperación de contraseña

“Recordar contraseña” se interpreta como recuperación segura: nunca se recupera
la contraseña anterior. Se emitirá un token o código temporal mediante un canal
verificado y se permitirá establecer una nueva.

Mantener una sesión abierta (“recuérdame”) es una decisión diferente sobre
duración y renovación de sesiones.

## Primer incremento backend aceptado

El primer incremento debe ser pequeño y atravesar producto, seguridad, datos, API
y operación:

1. un invitado prepara localmente un torneo de fútbol de liga y sus equipos;
2. una persona se registra, verifica su cuenta e inicia sesión sin perder el
   borrador;
3. el organizador autenticado persiste y publica el torneo con los datos mínimos;
4. el organizador consulta el estado actualizado del torneo y sus equipos;
5. una persona inicia sesión con Google y recibe la misma clase de sesión propia
   que con contraseña.

El corte siguiente implementa el registro y la corrección inmediata de
marcadores por administradores delegados en ligas en curso. La clasificación se
calcula en backend y se expone como una proyección de lectura; la retirada de
equipos conserva su entrega posterior.

El alcance está aceptado en [ADR-0043](../adr/0043-deliver-publish-and-read-league-first-backend-increment.md).
El Gate 0B está cerrado: el formato, los datos mínimos, el ciclo de vida, la
visibilidad, los participantes, la administración, los resultados, las bajas, la
cancelación y la frontera de identidad están definidos. Las capacidades aplazadas
no bloquean el primer vertical slice.

## Liga de fútbol inicial

Las [ADR-0032](../adr/0032-define-minimum-football-league-data-and-lifecycle.md)
y [ADR-0040](../adr/0040-make-published-leagues-editable-until-start.md) definen
la estructura mínima de una liga: nombre, fútbol, formato liga, organizador,
estado, equipos y partidos generados al iniciar. La configuración inicial permite
una o dos vueltas con puntuación 3-1-0; no incluye fechas, horas ni marcadores
especiales. Tras iniciarla, el backend calcula la clasificación: en dos vueltas
prioriza la mini-clasificación entre empatados y en una, diferencia de goles y
goles a favor generales; una igualdad que persiste comparte posición. La app
solo presenta esta proyección (ADR-0081).

El ciclo persistido es `publicado → en_curso → finalizado`, con `cancelado` como
estado terminal desde `publicado` o `en_curso`. El borrador se prepara localmente
o queda asociado temporalmente a una cuenta pendiente; se descarta y no forma
parte de este ciclo.

Cuando todos los partidos están resueltos, solo la organizadora puede finalizar
explícitamente la liga. El backend conserva todos los equipos de la posición 1
como co-campeones y la app muestra el resultado final antes de llevar a la
clasificación. Una liga finalizada ya no admite marcadores ni correcciones.

Una liga visible es consultable sin sesión por su ID público y el creador puede modificar sus equipos y
datos estructurales. En interfaz, `publicado` se muestra como «Sin empezar».
Al iniciarla el creador elige una o dos vueltas, se validan los datos, se generan
una sola vez los emparejamientos y se congelan equipos y reglas. Solo entonces los
organizador y los administradores delegados pueden registrar o corregir resultados. El creador solo puede
finalizarla cuando todos sus partidos tienen resultado. Si un equipo abandona en
`en_curso`, solo el creador puede declararlo: todos sus partidos, pendientes o
ya jugados, pasan a `3-0` a favor del rival y la liga continúa.

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
Las invitaciones y una audiencia restringida siguen fuera de este corte.

## Fuera del primer alcance

Salvo decisión posterior:

- pagos y premios;
- streaming o contenido multimedia;
- chat;
- marketplace;
- microservicios;
- motor genérico para todos los deportes;
- rankings globales;
- administración avanzada;
- calendario completo y arbitraje;
- notificaciones push;
- eliminatorias y formatos mixtos;
- email y push para avisar de asignaciones administrativas;
- invitaciones con aceptación, bloqueo y controles antiabuso para asignaciones;
- jugadores y membresías de personas en equipos.

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
liga, debe cancelarla o transferirla antes; esas operaciones se decidirán en un
corte posterior. Véase ADR-0074.
