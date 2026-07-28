# Sistema de diseño

> Estado: base visual Pulse aceptada en ADR-0054.

## Alcance inicial

Los tokens viven en `packages/design-tokens` y no dependen de React, Expo ni de
la web. Las pantallas consumen nombres semánticos, nunca hexadecimales o píxeles
repetidos. Los textos de interfaz se incorporarán en un catálogo separado.

## Fundaciones

- **Color:** azul como acción primaria; violeta como acento; superficies claras;
  verde, ámbar y rojo reservados para estado y feedback.
- **Tipografía:** familia de sistema por plataforma; escala de 12 a 32 px y pesos
  de 400 a 700. Así se preserva legibilidad nativa sin dependencia externa.
- **Espaciado:** escala de 4 px; los layouts usan 16 px como separación base y
  24 px como padding de pantalla amplio.
- **Forma:** controles con radio 12 px, tarjetas con 16 px y pills con 999 px.
- **Movimiento:** 160 ms para feedback y 240 ms para entradas/salidas; se respeta
  la preferencia de movimiento reducido de la plataforma.

## Componentes a implementar

| Componente | Estados mínimos | Regla de interacción |
| --- | --- | --- |
| Button | primary, secondary, ghost, destructive, disabled, loading | `loading` deshabilita el control y reserva el ancho del texto para el loader. |
| TextField | default, focus, filled, error, disabled | El error aparece bajo el campo cuando el validador se ejecuta; no borra el valor ni el foco. |
| Picker | default, focus, selected, error, disabled | Abre un selector adaptado a plataforma y conserva etiqueta visible. |
| Checkbox / Switch | default, selected, disabled, error | Objetivo táctil mínimo de 44 px. |
| Card | default, actionable, selected | Sirve para ligas, equipos y bloques de resumen, no como contenedor genérico indiscriminado. |
| Banner | network-error, generic-error, success | Se coloca arriba, tiene autocierre, botón de cierre y gesto de descartar. |
| InlineMessage | error, help, success | Bajo el control asociado; texto claro y disponible para lector de pantalla. |

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
