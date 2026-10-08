# Host global del banner Android — análisis técnico

> Estado: vía del adaptador autorizada por el usuario; ADR-0150 aceptado. Fecha: 2026-10-07.

El banner Android aparece sobre el editor de resultado, pero el Modal raíz
captura la pulsación del backdrop mientras el aviso está visible. La misma
pulsación cierra el editor después del aviso. El inventario de QA conserva las
capturas y el resultado original del fixture.

El contrato funcional permanece decidido: un único aviso global, por encima de
pantallas y popups, con temporizador y descarte independientes de la pantalla de
origen. En web el host sigue perteneciendo a Screen. ADR-0054 y ADR-0055 están
aceptados; el usuario ha reafirmado explícitamente esa diferencia entre plataformas.

## Evidencia técnica

La versión instalada de React Native 0.86.2 crea ComponentDialog con tema de
pantalla completa en ReactModalHostView.kt. transparent elimina DIM_BEHIND;
no introduce un área táctil limitada al banner. pointerEvents pertenece a la
jerarquía de vistas y no resuelve por sí solo el paso entre ventanas nativas.
FullWindowOverlay de react-native-screens 4.26.2 tiene implementación iOS;
no aporta ese host Android.

Android documenta PopupWindow.setTouchModal(false) para enviar los toques
exteriores a ventanas inferiores. Esto demuestra disponibilidad del mecanismo,
no que una integración concreta con Fabric, React y los popups actuales funcione.
Expo permite módulos locales y vistas nativas para registrar el token de ventana.

## Alternativas y coste

| Alternativa | Ventajas | Inconvenientes y mantenimiento |
| --- | --- | --- |
| Adaptador Android local de Expo para un host global no modal | Aísla la solución en infraestructura de UI; no modifica React Native; conserva provider, temporizador y contenido compartidos | Añade Kotlin y una build nativa; debe verificar orden de ventanas, área táctil, React/Fabric, insets, teclado, descarte y ciclo de Activity. Coste medio, localizado |
| Parche de React Native para un Modal con ventana limitada al banner | Reutiliza transporte de eventos y renderizado del Modal existente | Depende de internals de DialogRootViewGroup, medición y estado Fabric; debe revalidarse al actualizar React Native. Coste alto; no recomendado |
| Conservar temporalmente el host actual | Sin nueva infraestructura ni dependencia | Mantiene el bloqueo durante el aviso; la incidencia sigue abierta |

No se considera trasladar el host a Screen ni suprimir los gestos o el descarte:
esas alternativas cambiarían el comportamiento funcional aceptado. Tampoco se
pide SYSTEM_ALERT_WINDOW ni se coloca el banner sobre otras aplicaciones.

## Recomendación aceptada

Explorar el adaptador Android local, sin dependencias externas, después de la
aceptación explícita de ese nuevo coste nativo. Es una recomendación, no una
solución ya validada. El módulo limita la ventana al rectángulo del banner;
no basta con aplicar NOT_TOUCH_MODAL a una ventana que siga siendo fullscreen.
Debe mantener el aviso sobre una nueva ventana modal sin reiniciar su vida y
restaurar correctamente su host si se cierra la ventana inferior.

Antes de implementar, registrar la decisión técnica en un ADR aceptado. Si la
prueba nativa no conserva el contrato, retirar el experimento y mantener abierta
la incidencia; no adaptar el contrato a la limitación encontrada.

## Criterios de aceptación de la futura prueba

- Banner sobre ruta y sobre popup, sin doble renderizado visible.
- Backdrop, scroll, controles y navegación inferiores reciben sus toques.
- Cerrar popup, navegar o desmontar origen conserva el mismo aviso y temporizador.
- Toque y arrastre sobre el propio banner permiten descartarlo.
- Reemplazar un aviso conserva la política de uno único; no quedan ventanas huérfanas.
- Teclado, rotación, área segura, segundo plano y vuelta a foreground no rompen el host.
- Build Android instalada antes del QA; check y exportación web pasan.
- Sin cambios en los hosts iOS/web ni en reglas de negocio o mensajes HTTP.

## Fuentes

- [PopupWindow de Android](https://developer.android.com/reference/android/widget/PopupWindow#setTouchModal(boolean)).
- [Código nativo local con Expo](https://docs.expo.dev/workflow/customizing/).
- [Vistas con hijos en Expo Modules](https://docs.expo.dev/modules/module-api/#view-groups).
- Código fijado de React Native y react-native-screens en pnpm-lock.yaml.

Retrospectiva: preservar un contrato puede exigir cambiar el mecanismo nativo.
La existencia de una API adecuada no sustituye una prueba de integración, y ese
coste debe aceptarse antes de introducir infraestructura nueva en el cliente.


## Resultado de la prueba de mecanismos de ventana

La primera implementación PopupWindow pasa superposición, backdrop y descarte,
pero API 34 registra PopupWindow.this::dismiss en PopupDecorView.onAttachedToWindow
sin comprobar focusable. Se verifica en sources/android-34/android/widget/PopupWindow.java
del SDK instalado y con Atrás en el emulador. Restaurar el popup desde OnDismiss
conserva el aviso pero consume Atrás: no cumple el contrato de navegación inferior.

Se concreta el mismo adaptador autorizado con WindowManager.addView y una ventana
TYPE_APPLICATION_SUB_PANEL limitada al banner. FLAG_NOT_FOCUSABLE y
FLAG_NOT_TOUCH_MODAL dejan foco y toques exteriores a la ventana inferior. No se
registra un callback de Atrás, no se usa reflection ni internals de React Native.
La comprobación de integración debe repetirse sobre esa build; las capturas de
PopupWindow documentan el experimento anterior y no acreditan por sí solas la
implementación final.
