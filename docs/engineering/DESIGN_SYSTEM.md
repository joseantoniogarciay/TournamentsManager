# Sistema de diseño

> Estado: base visual Pulse aceptada en ADR-0054.

## Alcance inicial

Los tokens viven en `packages/design-tokens` y no dependen de React, Expo ni de
la web. Las pantallas consumen nombres semánticos, nunca hexadecimales o píxeles
repetidos. Los textos de interfaz se incorporarán en un catálogo separado.

## Fundaciones

- **Color:** azul como acción primaria; violeta como acento; superficies claras;
  verde, ámbar y rojo reservados para estado y feedback.
- **Degradado de marca:** azul `#155EEF` a violeta `#7F56D9`; las acciones
  primarias filled lo usan como superficie y las secundarias lo reservan para un
  borde de 1 px, manteniendo la superficie interior neutra.
- **Tipografía:** familia de sistema por plataforma; escala de 12 a 32 px y pesos
  de 400 a 700. Así se preserva legibilidad nativa sin dependencia externa.
- **Espaciado:** escala de 4 px; los layouts usan 16 px como separación base.
  Cada card reserva siempre 20 px de margen exterior horizontal y los layouts
  dejan 20 px entre cards, sin alterar su padding interno.
- **Forma:** controles con radio 12 px, tarjetas con 16 px y pills con 999 px.
- **Movimiento:** 160 ms para feedback y 240 ms para entradas/salidas; se respeta
  la preferencia de movimiento reducido de la plataforma.

## Componentes a implementar

| Componente        | Estados mínimos                                           | Regla de interacción                                                                         |
| ----------------- | --------------------------------------------------------- | -------------------------------------------------------------------------------------------- |
| Button            | primary, secondary, ghost, destructive, disabled, loading | `loading` deshabilita el control y reserva el ancho del texto para el loader.                |
| TextField         | default, focus, filled, error, disabled                   | El error aparece bajo el campo cuando el validador se ejecuta; no borra el valor ni el foco. |
| Picker            | default, focus, selected, error, disabled                 | Abre un selector adaptado a plataforma y conserva etiqueta visible.                          |
| Checkbox / Switch | default, selected, disabled, error                        | Objetivo táctil mínimo de 44 px.                                                             |
| Card              | default, actionable, selected                             | Sirve para ligas, equipos y bloques de resumen, no como contenedor genérico indiscriminado.  |
| Banner            | network-error, generic-error, success                     | Se coloca arriba, tiene autocierre, botón de cierre y gesto de descartar.                    |
| InlineMessage     | error, help, success                                      | Bajo el control asociado; texto claro y disponible para lector de pantalla.                  |

`Card` está implementada en `shared/ui`: aplica superficie, borde, radio,
padding y margen exterior horizontal semánticos. La home la usa para separar
bloques de acción, explicación, pasos y contexto; no sustituye a los
contenedores de layout.

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
Los requisitos que dependan del servidor se muestran cuando llegue la respuesta.
Un error por campo se asocia programáticamente a su control; el banner queda para
errores que no se pueden atribuir a un campo.

Mensajes globales iniciales:

- Sin conectividad o petición no alcanzable: «No hemos podido conectarnos. Revisa
  tu conexión e inténtalo de nuevo.»
- Error no tipado o 5xx: «Estamos teniendo problemas. Lo sentimos, inténtalo más
  tarde.»

No se muestran cuerpos, trazas ni mensajes internos del backend.
