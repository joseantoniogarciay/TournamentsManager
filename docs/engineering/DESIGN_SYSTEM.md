# Sistema de diseño

> Estado: base visual Pulse aceptada en ADR-0054.

## Alcance inicial

Los tokens viven en `packages/design-tokens` y no dependen de React, Expo ni de
la web. Las pantallas consumen nombres semánticos, nunca hexadecimales o píxeles
repetidos. Los textos de interfaz se incorporarán en un catálogo separado.

## Fundaciones

- **Color:** azul como acción primaria; violeta como acento; superficies claras;
  verde, ámbar y rojo reservados para estado y feedback.
- **Indicadores informativos:** un número de paso o un marcador no interactivo
  usa borde y texto primarios del tema; no emplea el color de acción para evitar
  que parezca un botón.
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
- **Navegación por tabs:** la tab activa usa el azul primario sólido. La barra
  nativa no admite un degradado como tint para icono y etiqueta.
- **Área inferior bajo tabs:** las rutas de una tab extienden su superficie hasta
  la barra nativa superpuesta. El margen para que el último control no quede
  oculto pertenece al `contentContainerStyle` del `ScrollView`, no al contenedor
  `Screen`; así el contenido puede desplazarse completamente por encima de la
  barra. Toda ruta bajo tabs usa el cálculo compartido
  `useTabContentBottomPadding`. En web, donde el safe-area inferior es cero pero
  la botonera estándar también se superpone, suma `space[10]` (40 px) al padding
  base de `space[12]`; en apps usa el inset seguro más ese mismo padding base.
- **Teclado y tabs:** en web la barra de tabs se ancla al borde inferior del
  viewport visual, también al aparecer el teclado. El padding inferior de un
  formulario web no toma el safe-area inset, porque puede variar al aparecer el
  teclado. Solo Safari recibe además una segunda medida del viewport al terminar
  de ocultar el teclado para descartar la altura intermedia. En iOS, los
  `ScrollView` de formularios ajustan su inset de teclado para que el campo
  activo pueda desplazarse por encima de la barra y el teclado.
- **Web en iPhone:** el viewport web usa `viewport-fit=cover` para que la
  superficie `canvas` alcance las zonas superior e inferior del navegador. Los
  insets existentes siguen reservando esas zonas al contenido; el documento web
  sincroniza su fondo y `theme-color` con el tema resuelto.
- **Botonera web:** web usa la barra inferior estándar de `Tabs`, no el fallback
  de `NativeTabs`. Conserva las tres rutas, iconos y colores semánticos; Cuenta
  ofrece en su cabecera el mismo acceso localizado a Ajustes que las apps.
- **Tipografía:** familia de sistema por plataforma; escala de 12 a 32 px y pesos
  de 400 a 700. Así se preserva legibilidad nativa sin dependencia externa.
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

| Componente        | Estados mínimos                                           | Regla de interacción                                                                                                                                                           |
| ----------------- | --------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Button            | primary, secondary, ghost, destructive, disabled, loading | `loading` deshabilita el control y reserva el ancho del texto para el loader.                                                                                                  |
| TextField         | default, focus, filled, error, disabled                   | El foco usa el borde azul primario del perímetro completo del campo; en web no se muestra un anillo interno adicional. El error aparece bajo el campo cuando el validador se ejecuta; no borra el valor ni el foco. Un campo de contraseña que ofrece visibilidad muestra ojo u ojo tachado según su estado y mantiene un objetivo táctil de 44 px. |
| Picker            | default, focus, selected, error, disabled                 | Abre un selector adaptado a plataforma y conserva etiqueta visible.                                                                                                            |
| Checkbox / Switch | default, selected, disabled, error                        | Objetivo táctil mínimo de 44 px.                                                                                                                                               |
| Card              | default, actionable, selected                             | Sirve para ligas, equipos y bloques de resumen, no como contenedor genérico indiscriminado.                                                                                    |
| Banner            | network-error, generic-error, success                     | Gestor global de aviso único: sustituye el actual, se coloca arriba del área segura como una card y tiene autocierre, toque o arrastre vertical hacia arriba para descartarlo. |
| LoadingTransition | active, mensaje localizado, movimiento reducido           | Capa opaca modal con mensaje y loader; bloquea interacción y solo entra o sale con `fade`. La duración mínima pertenece al flujo que la usa. |
| InlineMessage     | error, help, success                                      | Bajo el control asociado; texto claro y disponible para lector de pantalla.                                                                                                    |

`Card` está implementada en `shared/ui`: aplica superficie, borde, radio,
padding y margen exterior horizontal semánticos. La home la usa para separar
bloques de acción, explicación y pasos; no sustituye a los contenedores de
layout. El acceso persistente a la biblioteca de torneos vive en su tab y no se
duplica como una card informativa en la home.

El banner global conserva la separación lateral y el radio de una card, pero usa
un padding compacto de `space[3]` para no ocupar más altura de la necesaria. Se
coloca tras el inset seguro superior, con una separación adicional de 4 px
para no bajar innecesariamente desde el notch. Entra y sale con
`motion.enterExit`; si el sistema solicita movimiento reducido, aparece y se
descarta sin transición. El gestor mantiene un único aviso: al llegar uno nuevo,
cancela el temporizador y la salida del anterior y muestra únicamente el último.
Tocar la card o arrastrarla hacia arriba la descarta; un arrastre corto vuelve a
su posición para no cerrar el aviso por accidente.

Una acción externa que se represente solo con un icono de marca, como Google,
conserva un objetivo táctil de al menos 44 px, forma circular y `accessibilityLabel`
localizado. El asset se guarda localmente: no se descarga durante el uso de la app.
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

La validación de formato se ejecuta al abandonar un campo y al intentar enviar.
Como excepción acotada, `TextField` permite validarla al cambiar el texto cuando
el feedback inmediato ayuda a completar un requisito, como la longitud mínima
de una contraseña; el indicador complementario se muestra solo al cumplirlo.
Los requisitos que dependan del servidor se muestran cuando llegue la respuesta.
Un error por campo se asocia programáticamente a su control; el banner queda para
errores que no se pueden atribuir a un campo.

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
ejemplo, un límite de solicitudes); no se centralizan `status`, `type` ni copy
de negocio que todavía no se repitan.
