# ADR-0150: Host Android local para el banner global

- **Estado:** Aceptado
- **Fecha:** 2026-10-07
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** Ninguno; concreta ADR-0054, ADR-0055 y ADR-0015.

## Problema y evidencia

El Modal raíz Android eleva el banner sobre los popups, pero captura sus toques
exteriores. El backdrop del editor no responde durante el aviso y sí después.
El alcance global y la vida independiente de las pantallas están decididos;
reducir el banner a Screen cambiaría ese comportamiento.

## Alternativas y mantenimiento

El [análisis previo](../engineering/ANDROID_GLOBAL_BANNER_HOST_ANALYSIS_2026-10-07.md)
compara un adaptador nativo local (coste medio y localizado), parchear internals
de React Native (coste alto en upgrades) y mantener la incidencia temporalmente.
Recomienda el adaptador, sin librerías adicionales ni permisos de overlay.

## Decisión explícita del usuario

El usuario responde **«Autorizar adaptador Android»** el 2026-10-07. Después
precisa que lo importante es no romper iOS y comprobar que su host no estuviera
ya roto. Se acepta añadir el adaptador Android local de Expo conservando el
contrato funcional. No se autoriza cambiar ese contrato ni introducir por
anticipación un adaptador iOS.

## Límites de implementación

- El provider raíz sigue siendo dueño del aviso, reemplazo, temporizador y cierre.
- Android usa una ventana hija de WindowManager no modal limitada al área del banner. Los toques
  exteriores alcanzan la ventana inferior; el banner conserva toque y gesto.
- Las ventanas de popup pueden aportar su anclaje nativo al host único; no poseen
  el aviso ni su temporizador. Cambiar o desmontar el anclaje no cancela el aviso.
- El módulo solo se enlaza en Android, vive en modules/ y sobrevive a CNG.
  No se editan los directorios nativos generados ni se parchea React Native.
- No se añade dependencia externa, permiso SYSTEM_ALERT_WINDOW ni acceso a
  ventanas de otras apps. Los textos y tokens llegan desde el cliente compartido.
- iOS conserva FullWindowOverlay y web conserva su host por Screen. Cualquier
  defecto detectado en iOS requiere evidencia propia y tratamiento separado.

## Validación y cierre

Build Android recompilada e instalada; popup, backdrop, navegación, reemplazo,
autocierre, toque, gesto, teclado, rotación y ciclo de Activity. iOS se verifica
antes y después: superposición, interacción inferior, supervivencia y descarte.
Check y exportación web obligatorios; se conserva evidencia y fixtures originales.
Una implementación parcial no equivale a cerrar el QA completo.

Estado de ejecución: módulo Android compilado, instalado y conectado. Pasan
backdrop, navegación (incluido Atrás y desmontaje de la ficha origen), toque, gesto,
reemplazo, autocierre y regreso desde segundo
plano en API 34. El control iOS antes/después pasa superposición, interacción
inferior, supervivencia y toque. Su arrastre previo se localiza en el cálculo de
PanResponder y pasa tras conservar el desplazamiento anterior al grant, sin
adaptador iOS ni cambio de umbral o alcance global. Teclado visible validado. La recreación desde la app cargada recupera la
ruta y la sesión; la recreación durante el arranque revela un fallo del
development launcher que se reproduce también con el adaptador desregistrado
y su host JavaScript desactivado. Se conserva como incidencia de desarrollo
separada, sin parchear Expo. Queda accesibilidad con lector de
pantalla antes del cierre total. La app fija
portrait; no se amplía su orientación para probar el host. Evidencia detallada en
PRODUCT_QA_REVIEW_2026-10-04.md.

## Consecuencias y revisión

Se asume Kotlin localizado y validación nativa en upgrades de Expo/Android. Se
revisará si el orden de ventanas exige internals de React Native o si no puede
conservarse el contrato con el host aislado. No se degrada el contrato para hacer
pasar una prueba. La prueba nativa debe preceder a afirmar resuelta la incidencia.
