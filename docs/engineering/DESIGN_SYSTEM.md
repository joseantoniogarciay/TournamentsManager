# Sistema de diseño

> Estado: base visual Pulse aceptada en ADR-0054.

## Alcance vigente

Los tokens viven en `packages/design-tokens` y no dependen de React, Expo ni de
la web. Las pantallas consumen nombres semánticos, nunca hexadecimales o píxeles
repetidos. Los textos de interfaz viven en los catálogos localizados de i18n.

## Fundaciones

- **Color:** azul como acción primaria; violeta como acento; superficies claras;
  verde, ámbar y rojo reservados para estado y feedback.
- **Indicadores informativos:** un número de paso o un marcador no interactivo
  usa borde y texto primarios del tema; no emplea el color de acción para evitar
  que parezca un botón.
- **Loaders:** los indicadores de progreso sobre una superficie normal usan el
  azul principal en claro y blanco en oscuro. Los que van sobre una acción
  primaria filled permanecen blancos en ambos temas para contrastar con el
  degradado de marca.
- **Degradado de marca:** una base azul `#155EEF` recibe una superposición violeta
  `#7F56D9` limitada a la esquina inferior derecha. Su eje es diagonal hacia esa
  esquina, como en el icono de la aplicación. El violeta entra desde el 35 % del
  recorrido: no tiñe todo el lateral. El icono cuadrado usa el token
  `gradient.brand` y los botones horizontales `gradient.brandButton`, ambos con
  los mismos puntos diagonalizados para que web, iOS y Android compongan la misma
  superficie. La cabecera y el CTA de los emails de verificación usan la misma
  paleta y paradas en CSS, con azul sólido como fallback para clientes sin
  soporte de degradados. Las acciones primarias filled llevan texto y loader
  blancos en ambos temas; las secundarias lo reservan para un borde de 1 px.
- **Splash nativa:** muestra únicamente la marca interior, sin el recuadro del
  icono de aplicación. La marca usa el mismo degradado azul–violeta sobre el
  canvas claro u oscuro, para conservar contraste sin duplicar un fondo dentro
  del fondo de la splash.
- **Navegación por tabs:** la tab activa usa el azul primario sólido. La barra
  nativa no admite un degradado como tint para icono y etiqueta. En iOS 26 o
  superior conserva el material nativo Liquid Glass; en versiones anteriores
  desactiva el blur y usa `surface.default`, igual que web y Android, para que
  la barra opaca mantenga contraste sobre el canvas.
- **Área inferior bajo tabs:** las rutas de una tab extienden su superficie hasta
  la barra nativa superpuesta. El margen para que el último control no quede
  oculto pertenece al `contentContainerStyle` del `ScrollView`, no al contenedor
  `Screen`; así el contenido puede desplazarse completamente por encima de la
  barra. Toda ruta bajo tabs usa el cálculo compartido
  `useTabContentBottomPadding`. En web, donde el safe-area inferior es cero pero
  la botonera estándar también se superpone, suma `space[10]` (40 px) al padding
  base de `space[12]`; en apps usa el inset seguro más ese mismo padding base.
- **Acción flotante contextual:** cuando una biblioteca necesita conservar una
  acción de creación visible, sitúa un botón circular de al menos 44 px sobre la
  botonera, con 20 px de separación lateral. El contenido desplazable reserva
  también su altura y separación para que el último elemento nunca quede bajo el
  botón. El icono conserva una etiqueta accesible localizada.
- **Teclado y tabs:** en web la barra de tabs se ancla al borde inferior del
  viewport visual, también al aparecer el teclado. El padding inferior de un
  formulario web no toma el safe-area inset, porque puede variar al aparecer el
  teclado. Solo Safari recibe además una segunda medida del viewport al terminar
  de ocultar el teclado para descartar la altura intermedia. Los formularios
  nativos usan `KeyboardAwareScrollView`: al enfocar un campo o aparecer el
  teclado, desplaza solo lo necesario para dejar el control completo visible
  con 20 px de separación. En iOS conserva el ajuste nativo de insets; en
  Android añade únicamente la oclusión restante tras el resize de la ventana.
  El espacio temporal desaparece al ocultar el teclado y el arrastre no lo
  descarta por defecto. `autoFocus` sigue siendo una elección de cada ruta;
  mantener visible el foco no exige abrir el teclado en todos los formularios.
- **Web en iPhone:** el viewport web usa `viewport-fit=cover` para que la
  superficie `canvas` alcance las zonas superior e inferior del navegador. Los
  insets existentes siguen reservando esas zonas al contenido; el documento web
  sincroniza su fondo y `theme-color` con el tema resuelto.
- **Primer frame web:** el HTML aplica antes de cargar React la preferencia de
  tema persistida o, cuando permanece en `system`, la preferencia del navegador.
  El fondo del documento, `color-scheme` y `theme-color` nacen ya con el tema
  resuelto para evitar un destello claro al recargar en oscuro. Como la
  exportación estática prerenderiza el árbol de React en claro y sus estilos son
  inline, ese árbol permanece oculto sobre el canvas correcto hasta que la
  hidratación confirma el tema resuelto; con JavaScript desactivado vuelve a ser
  visible. Un script mínimo y bloqueante del mismo origen ejecuta la lectura
  inicial sin relajar la CSP; la clave y los colores siguen procediendo de la
  fuente TypeScript mediante atributos del documento.
- **Botonera web:** web usa la barra inferior estándar de `Tabs`, no el fallback
  de `NativeTabs`. Conserva las tres rutas, iconos y colores semánticos; Cuenta
  ofrece en su cabecera el mismo acceso localizado a Ajustes que las apps.
- **Botonera Android:** `NativeTabs` recibe fondo, colores de etiquetas e iconos
  y superficie del indicador desde los tokens del tema resuelto. Los colores
  dinámicos Material por defecto siguen el tema del SO y no garantizan que una
  preferencia explícita de la app se aplique a la barra.
- **Controles de cabecera:** toda acción de navegación que no use Liquid Glass
  —web, Android e iOS anterior a 26— usa un objetivo circular de 44 px,
  superficie por defecto y borde semántico mediante `NavigationHeaderButton`.
  Web y Android lo separan 20 px del lateral. Web añade los 20 px completos;
  Android añade solo 4 px porque la toolbar nativa ya aporta 16 px por lateral.
  El margen del botón complementa ese inset, no lo duplica. En iOS anterior a 26 el botón no
  añade margen porque la cabecera ya aplica su inset nativo, alineándolo con el
  control de `Stack.Toolbar.Button` de iOS 26. iOS 26 o superior conserva
  `Stack.Toolbar.Button` para respetar Liquid Glass. Cuando una ruta muestra en
  cabecera el nombre de una entidad,
  este se centra, reserva 20 px frente a los controles laterales y puede ocupar
  dos líneas; no se impone un ancho fijo que lo trunque antes de agotar ese espacio.
- **Tipografía:** Figtree local en web, iOS y Android, con los pesos 400, 500,
  600 y 700 cargados antes de montar la interfaz. Los tokens seleccionan la
  familia real de cada peso, en vez de sintetizarlo con `fontWeight`. La escala
  permanece entre 12 y 32 px. Las etiquetas de la primitiva `Button` usan
  semibold (600) para reforzar su legibilidad sin cambiar tamaño ni altura. El
  marcador de un partido dentro de una fila usa la escala `title` en negrita:
  mantiene jerarquía frente a los equipos sin forzar el alto de la card ni
  recortar los glifos. `display` queda para títulos y resultados destacados
  fuera de una fila compacta. Véase ADR-0096.
- **Espaciado:** escala de 4 px; los layouts usan 16 px como separación base.
  Cada card reserva siempre 20 px de margen exterior horizontal y los layouts
  dejan 20 px entre cards, sin alterar su padding interno.
- **Forma:** los campos y demás controles compactos usan radio 12 px; las
  tarjetas usan 16 px. Los botones usan `radius.pill` (999 px), de modo que sus
  extremos son siempre semicirculares respecto de su altura. La variante
  secundaria conserva solo un borde azul de 1 px: su interior es transparente y
  deja ver la superficie de la pantalla que la contiene.
- **Movimiento:** 160 ms para feedback y 240 ms para entradas/salidas; se respeta
  la preferencia de movimiento reducido de la plataforma.

## Componentes a implementar

| Componente         | Estados mínimos                                            | Regla de interacción                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| ------------------ | ---------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Button             | primary, secondary, ghost, destructive, disabled, loading  | `loading` deshabilita el control y reserva el ancho del texto para el loader. `destructive` conserva la superficie transparente, borde y texto de error; no existe una variante destructiva rellena para mantener la misma jerarquía en menús y confirmaciones.                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| Selector de opción | default, selected, disabled                                | Las alternativas de una misma configuración comparten forma y estados aunque seleccionen dimensiones distintas. La opción elegida usa superficie primaria sólida y la no elegida, superficie y borde neutros; no se reutiliza la jerarquía visual de un botón de acción para representar selección.                                                                                                                                                                                                                                                                                                                                                                                                                          |
| TextField          | default, focus, filled, error, disabled                    | El foco usa el borde azul primario del perímetro completo del campo; en web no se muestra un anillo interno adicional. El error aparece bajo el campo cuando el validador se ejecuta; no borra el valor ni el foco. Un campo de contraseña que ofrece visibilidad muestra ojo u ojo tachado según su estado y mantiene un objetivo táctil de 44 px.                                                                                                                                                                                                                                                                                                                                                                          |
| Picker             | default, focus, selected, error, disabled                  | Abre un selector adaptado a plataforma y conserva etiqueta visible.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| Checkbox / Switch  | default, selected, disabled, error                         | Objetivo táctil mínimo de 44 px. Si su texto contiene documentos legales, sus enlaces se integran en ese texto y no duplican acciones independientes; pulsarlos abren la ruta legal sin modificar la selección.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| Card               | standard, compact, actionable, selected                    | Sirve para ligas, equipos y bloques de resumen, no como contenedor genérico indiscriminado. La densidad `compact` se reserva para filas de una sola línea con una acción de 44 px; conserva borde, radio y margen exterior.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| Banner             | network-error, generic-error, success                      | Gestor global de aviso único: sustituye el actual, se coloca arriba del área segura como una card y tiene autocierre, toque o arrastre vertical hacia arriba para descartarlo.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| ConfirmationDialog | visible, cerrado                                           | El estado y la API son compartidos, pero su host se monta en cada `Screen` activa. Así un diálogo de una ruta nativa se presenta sobre esa ruta y no detrás de su modal o de la tab bar. Usa el blur oscuro clásico de iOS, sin la tinta dinámica de Liquid Glass, para conservar un scrim neutro y legible. Su borde usa `border.default` en ambos temas para separar el diálogo sin crear una variante exclusiva de oscuro. Android añade una atenuación neutra leve si el blur no está disponible. En web, el scrim transiciona simultáneamente desde `blur(0px)` hasta el blur final; el `Modal` no aplica `fade` para que el filtro pueda muestrear la página durante toda la entrada. Tocar fuera equivale a cancelar. |
| LoadingTransition  | active, mensaje localizado, movimiento reducido            | Capa opaca modal con mensaje y loader; bloquea interacción y solo entra o sale con `fade`. Toda carga inicial que impida mostrar el contenido de una ruta la usa centrada sobre la pantalla; los indicadores locales se reservan para contenido parcial o acciones en curso.                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| RequestErrorCard   | error de red, genérico o no disponible; reintento o cierre | Estado terminal de una carga sin contenido. Recibe el mensaje seguro ya clasificado; reintenta cuando esa acción puede recuperar la carga y ofrece cierre cuando la feature conoce un estado terminal, como un recurso que devuelve `404`. Sustituye al banner para no duplicar el aviso en una pantalla vacía.                                                                                                                                                                                                                                                                                                                                                                                                              |
| InlineMessage      | error, help, success                                       | Bajo el control asociado; texto claro y disponible para lector de pantalla.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |

El deporte del torneo usa el «Selector de opción» al principio del formulario de
creación, antes de los campos que pueda condicionar. Sus valores son «Fútbol»,
«Baloncesto», «Balonmano», «Tenis» y «Pádel», conserva una selección visible y
no se vuelve editable tras publicar. Tenis presenta una segunda selección al
mejor de tres o cinco sets; pádel informa de que usa tres y ambos restringen el
formato a eliminatoria directa. El enum recibido gobierna el vocabulario
posterior: goles para fútbol y balonmano, puntos para baloncesto, ayuda inline
cuando un tanteo de baloncesto queda empatado, «lanzamientos de 7 metros» para
el desempate de balonmano y «participantes» para persona o pareja en deportes de
raqueta. El editor de raqueta presenta un par de campos por set, permite dejar
vacíos los sets que ya no se disputan y mantiene deshabilitada la acción hasta
que el resultado completo sea coherente. El resultado administrativo fijo se
explica en la confirmación de retirada, sin presentar un campo que sugiera que
puede configurarse.

`Card` está implementada en `shared/ui`: aplica superficie, borde, radio,
padding y margen exterior horizontal semánticos. La home la usa para separar
bloques de acción, explicación y pasos; no sustituye a los contenedores de
layout. En Inicio, los títulos de «Actividad reciente» y sugerencias permanecen
fuera de sus cards y dejan `space[4]` antes del primer contenido; los torneos
hermanos conservan entre sí `space[5]`, mientras las secciones principales se
separan con `space[6]`. El acceso persistente a la biblioteca de torneos vive en
su tab y no se duplica como una card informativa en la home.

En el cuadro de eliminatorias, la navegación entre partidos es secundaria frente
al resultado. El destino posterior se presenta como enlace sin borde, debajo de
la acción de resultado y alineado al final. El origen de cada plaza usa una
flecha atrás junto al equipo: el icono acompaña la altura visual del texto, pero
su objetivo táctil conserva 44 px y su etiqueta accesible nombra el partido de
origen completo. En vistas compactas, el selector de ronda permanece sticky y es
la única referencia visual repetida a la fase. En web grande desaparece esa
botonera: una cabecera sticky alineada con cada columna nombra simultáneamente las
fases visibles. Ninguno de los dos modos añade otro título dentro de cada tarjeta.
Al cambiar de ronda mediante el selector compacto, el desplazamiento vertical de
los partidos vuelve al inicio de esa sección: el primer partido queda bajo el
selector sticky y no hereda la posición de la ronda anterior.
Cada encuentro usa `bodyLarge` semibold con «Partido N»; la final muestra
únicamente «Final». Las etiquetas accesibles y los enlaces de navegación sí
conservan ronda y partido completos para no perder contexto. Los nombres de
clubes permanecen en `bodyLarge` regular y el marcador conserva la jerarquía
tipográfica principal.
El ganador se identifica con una corona sobre el degradado de marca, sin añadir
texto al nombre. Al navegar entre emparejamientos, la tarjeta destino se desplaza
hasta quedar completamente visible y sustituye el texto «Seleccionado» por un
borde con ese mismo degradado.
En web, cuando el conjunto de columnas desborda el viewport, una barra horizontal
sincronizada con el cuadro y sus cabeceras permanece fijada al borde inferior.
Permite recorrer todas las fases sin modificar la ronda seleccionada y se sitúa
por encima de una acción flotante cuando ambas coinciden. La barra pertenece al
viewport de pantalla como hermana del scroll vertical; no vive dentro de este,
porque un ancestro transformado puede cambiar el bloque de referencia de
`position: fixed`.

En «Liga + eliminatorias», liga, desempate y cuadro se presentan como fases
seleccionables separadas. «Desempate» solo aparece cuando existen partidos de
esa fase; una card breve explica el motivo, el requisito de ganador y la posible
repetición. Sus secciones nombran tanto el bloque independiente como el ciclo,
sin añadir esos resultados a la clasificación de liga. La acción que congela la
liga se llama «Cerrar liga y continuar», porque el sistema puede abrir un
desempate antes de poder crear el cuadro.

El banner global conserva la separación lateral y el radio de una card, pero usa
un padding compacto de `space[3]` para no ocupar más altura de la necesaria. Se
coloca tras el inset seguro superior, con una separación adicional de 4 px
para no bajar innecesariamente desde el notch. Entra y sale con
`motion.enterExit`; si el sistema solicita movimiento reducido, aparece y se
descarta sin transición. El gestor mantiene un único aviso: al llegar uno nuevo,
cancela el temporizador y la salida del anterior y muestra únicamente el último.
Tocar la card o arrastrarla hacia arriba la descarta; un arrastre corto vuelve a
su posición para no cerrar el aviso por accidente. En iOS y Android el host es
global y vive por encima de las rutas, para que sobreviva un reset de navegación.
En web se monta como overlay de la `Screen` activa con
`pointerEvents="box-none"`: solo la card recibe la interacción y el resto de
controles visibles sigue siendo operable. No usa
`Modal` web, porque su portal de viewport interfiere con las pulsaciones y puede
hacer que Safari iOS cambie el color de la zona segura superior. iOS y Android
mantienen el `Modal` nativo para presentar el aviso por encima de la navegación.

Una acción externa que se represente solo con un icono de marca, como Google,
conserva un objetivo táctil de al menos 44 px, forma circular y `accessibilityLabel`
localizado. El asset se guarda localmente: no se descarga durante el uso de la app.
En Cuenta, Apple y Google comparten una fila centrada, con Apple primero y Google
después en todas las plataformas. Ambos controles son círculos de 48 px separados
16 px, sin texto visible y con etiqueta accesible localizada. Apple usa un asset
local monocromo teñido con el color de texto; Google conserva su marca multicolor.
Los proveedores sin configuración permanecen visibles y deshabilitados.
El nonce requerido por un proveedor se precarga al enfocar la ruta que muestra
su acción, nunca al montar una tab que permanece oculta. Mientras se prepara,
el icono se sustituye por un loader sin bloquear el resto de la pantalla; un
fallo de esa precarga es silencioso y un toque posterior lo reintenta de forma
explícita. `InteractionBlocker` conserva
una capa transparente modal, accesible y reutilizable para futuros estados que
sí deban impedir la interacción de una ruta; no se usa cuando el proveedor
externo ya presenta y controla su propia interfaz.

La primera home usa las mismas primitivas de texto, botón y superficie: presenta
una única acción principal, un acceso secundario a cuenta y contenido orientativo
para una persona sin sesión. No simula colecciones personalizadas hasta que haya
sesión y datos autorizados que mostrar.
Su copy vive en los catálogos JSON planos de `shared/i18n/locales/`, con español,
inglés, italiano y francés; el idioma no soportado usa inglés como fallback. Las
claves semánticas (`common_cancel`, `home_create_tournament`) son estables para
permitir importar y exportar cada locale con una plataforma de traducción. La
detección y la selección de catálogo se centralizan en `shared/i18n/locale.ts`.

## Errores de formulario

El diálogo de resultados coloca campos y acción Guardar en el mismo contenido
desplazable, limitado al 85 % de la altura disponible. Los sets se presentan
como filas con dos columnas de tanteo y sus cabeceras una sola vez; cada campo
conserva una etiqueta accesible localizada con número de set o juego y lado.
Los campos flexibles permiten reducir su ancho con `minWidth: 0`, incluido el
`TextInput`, para que su tamaño intrínseco web no desborde columnas estrechas.
`ModalDialog` ofrece `avoidKeyboard` para formularios nativos: adapta el espacio
al teclado mediante la primitiva compartida, sin alterar los diálogos que no
solicitan ese comportamiento. En pantallas bajas se desplaza también Guardar;
no se superpone a los últimos campos.

La validación de formato se ejecuta al abandonar un campo y al intentar enviar.
Como excepción acotada, `TextField` permite validarla al cambiar el texto cuando
el feedback inmediato ayuda a completar un requisito, como la longitud mínima
de una contraseña; el indicador complementario se muestra solo al cumplirlo.
Un teclado o `inputMode` numérico solo facilita la entrada: un campo que acepte
exclusivamente enteros filtra también el teclado físico y el pegado en su valor
controlado, y normaliza sus límites contractuales al perder el foco.
Los requisitos que dependan del servidor se muestran cuando llegue la respuesta.
Un error por campo se asocia programáticamente a su control; el banner queda para
errores que no se pueden atribuir a un campo.
Si una operación nace dentro de `ModalDialog`, el diálogo conserva visible su
feedback mientras permanezca abierto: los rechazos atribuibles a un campo usan
su error inline y los fallos generales usan un aviso accesible dentro del propio
diálogo. No se envían al banner de la pantalla subyacente, que en web queda por
debajo del portal modal.

Una ruta terminal que ya explica el estado y ofrece la siguiente acción, como un
enlace de verificación inválido, no publica además el mismo error en el banner:
la card de la ruta es el único feedback. Así se evita duplicar el copy y ocupar
espacio antes de que la persona pueda leer la recuperación disponible.

Mensajes globales iniciales:

- Sin conectividad o petición no alcanzable: «No hemos podido conectarnos. Revisa
  tu conexión e inténtalo de nuevo.»
- Error no tipado o 5xx: «Estamos teniendo problemas. Lo sentimos, inténtalo más
  tarde.»

No se muestran cuerpos, trazas ni mensajes internos del backend.

La clasificación compartida solo distingue un rechazo de transporte marcado por
`apiFetch` de cualquier respuesta o fallo no tratado. Cada feature reconoce
antes los estados del contrato que cambian la recuperación de la persona (por
ejemplo, un límite de solicitudes o un `404` que hace inútil reintentar); no se
centralizan `status`, `type` ni copy de negocio que todavía no se repitan.

Tenis de mesa añade elección de 3, 5 o 7 juegos y reutiliza el vocabulario de
participantes. El editor etiqueta juegos y puntos, permite tanteos de hasta
cinco dígitos y conserva la validación del encuentro completo y el feedback seguro.

Voleibol reutiliza equipos, fases y editor de sets; muestra puntos y ayuda
25/15, con mejor de cinco fijo. La clasificación conserva equipo y puntos
fijos y desplaza nueve estadísticas (partidos, ganados, perdidos, sets,
cociente de sets, tantos y cociente de tantos). Los cocientes visibles usan
hasta tres decimales según locale; ∞ representa positivo/0. El orden procede
siempre de la proyección del backend.

La «i» abre el `ModalDialog` compartido con contenido desplazable y altura
máxima relativa: título, deporte, significado exacto de las abreviaturas,
tanteo, puntos, orden, igualdades y retirada. El formato mixto añade corte,
siembra y ciclos de desempate. Todo el copy se mantiene en es/en/it/fr; la
explicación de cada deporte corresponde a la regla vigente, sin fallback a
fútbol para los perfiles por sets.

### Selector de puntos de bádminton

Creación reutiliza `ConfigurationOption` para 21/15 puntos y muestra la regla
fija al mejor de tres. Los cuatro catálogos incluyen selector, resumen y ayudas
según perfil. El resultado reutiliza las filas compactas del diálogo por sets,
con puntos, juegos y etiquetas accesibles por lado. Guardar queda dentro del
scroll y se habilita solo con dos victorias válidas según el perfil persistido.

### Selector de incidencias de resultado

El `ModalDialog` de resultado reutiliza `ConfigurationOption` para
Marcador/Incidencia, motivo y participante afectado. Las opciones ajustan nombres
largos con `maxWidth: "100%"`, conservando tokens, semántica y objetivo táctil.
Abandono reutiliza los campos compactos de parcial y muestra la victoria antes
de Guardar, que permanece dentro del scroll. `IncidentSummary` comparte la
lectura en liga y cuadro, con parcial real y resultado administrativo etiquetado
cuando afecta a clasificación. Todo el copy vive en es/en/fr/it.

Revisión del JSX real en una ruta efímera sin envíos: 320×480 con nombre largo
y parcial de pádel, y 390×844 para baloncesto. La secuencia de dos sets
incompletos bloquea Guardar. La ruta se retira antes de exportar. El teclado
nativo requiere validación posterior en dispositivo; el viewport web no lo
acredita.

## Idioma automático

Según ADR-0143, el idioma se obtiene del sistema en móvil y del navegador en
web. No se ofrece selector de idioma. Se mantienen los catálogos es/en/it/fr
y el fallback inglés.


La decisión ADR-0149 retira temporalmente los controles de analítica de uso de
Inicio/Ajustes: no hay captura de producto activa. La fiabilidad mínima prod
no se presenta como una preferencia de uso; las preferencias antiguas no activan
SDK en beta/local ni eventos de producto.
