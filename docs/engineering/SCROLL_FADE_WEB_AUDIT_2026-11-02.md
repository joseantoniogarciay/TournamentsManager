# Auditoría del fade y eventos de scroll web — 2026-11-02

## Problema y causa

La foto de la ficha de torneo muestra el borde superior sin fade. En la versión
instalada de React Native Web, `ScrollViewBase.normalizeScrollEvent` no incluye
`contentInset`, aunque el tipo universal lo declara. Leer `.top` lanzaba una
excepción antes de actualizar bordes y antes de llamar al consumidor. Por tanto,
el defecto afectaba potencialmente a todos los scrolls verticales compartidos.

## Recorrido de código

Se revisan los 26 archivos de pantallas/features con scroll: inicio, biblioteca,
cuenta/acceso/registro/contraseña/recuperación/sesiones, ajustes, notificaciones,
legales, crear torneo, unirse, ficha, equipos, administradores/alta/transferencia,
clasificación, ediciones, encuentros personales, programación y repetición,
además del cuadro compartido.

- Los scrolls de pantalla importan la primitiva compartida.
- `KeyboardAwareScrollView` utiliza esa misma primitiva y solo lee
  `contentOffset.y` en su callback.
- `SectionList` de la ficha utiliza `renderFadingScroll`, de identidad estable.
- Los callbacks de ficha y cuadros leen `contentOffset.x/y`, presente en web.
- El scroll horizontal animado de clasificación recibe Animated.event solo en
  nativo y usa exclusivamente contentOffset.x; no lee contentInset.
- No hay otras lecturas de contentInset, velocity, zoomScale,
  responderIgnoreScroll o targetContentOffset en el cliente.
- El backdrop desplazable de ModalDialog conserva el ScrollView nativo sin
  callback de fade: no comparte el defecto ni requiere fondo canvas opaco.

La corrección considera cero el inset ausente y conserva los insets nativos.
No se copian correcciones por pantalla ni se modifica la lógica de negocio.

## Validación y límites

`make test-scroll-fade` ejecuta cinco pruebas del componente: evento web sin
insets, evento nativo, extremos/proximidad, contenido corto, inset ajustado,
callback original, integración con el wrapper de formularios y paso directo
horizontal con ref. Se integra en `make verify`. Typecheck, lint y exportación
web correctos; gate completo local correcto con PostgreSQL omitido sin DB local.

Esto acredita revisión de código y regresión ejecutable; no es un recorrido
visual de 26 pantallas en Safari. Computer Use devuelve «Local browser use is
disabled in Work settings». QA visual Safari/iOS/Android pendiente. No se
arrancan local/dev ni observabilidad.

## Retrospectiva

Una primitiva compartida propaga tanto la solución como los fallos. El tipo de
React Native no garantiza la forma del evento web: probar ese límite evita que
una excepción visual interrumpa también el comportamiento de la feature.
