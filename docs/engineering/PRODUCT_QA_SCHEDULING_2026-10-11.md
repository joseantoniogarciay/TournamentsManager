# QA nativa de fecha, hora y lugar — 2026-10-11

Los recorridos funcionales revisados de [ADR-0158](../adr/0158-add-optional-tournament-and-match-scheduling.md)
pasan en iPhone 18 Pro / iOS 27.0 y Pixel API 34 / Android 14. No se encontraron
defectos de producto en esta muestra. Este informe completa la revisión web y
las pruebas de backend del incremento; no es una certificación de todos los
dispositivos ni de una distribución Release por tienda.

## Entorno y alcance

Development builds reconstruidas desde develop 8cac014, con ExpoStoreReview
incluido, bundles Hermes y Metro local. Las interacciones iOS se realizan con
Device Hub y accesibilidad; Android con ADB, autorizado explícitamente por el
usuario tras fallar el control de la ventana flotante de Android Studio.

La API local contiene exclusivamente un torneo ficticio, dos equipos y un
partido pendiente. Permite cambiar la relación personal e inyectar un HTTP 500.
Esto acredita la UI nativa y su adaptación HTTP, sin atribuir a la fixture la
persistencia ni autorización del servidor real. Estas últimas están cubiertas
por `make verify` y la suite de integración PostgreSQL previa, incluida la
migración 00021, cierre, bye, transferencia y retirada de permiso.

## Recorridos observados

| Caso | iOS | Android |
| --- | --- | --- |
| Creador abre campos opcionales sin exigirlos al torneo | Pasa | Pasa |
| 31/02/2026 se rechaza y conserva la entrada | Pasa | Pasa |
| Guardar 18/10/2026, 18:30 Europe/Madrid y Pista 2 | Pasa | Pasa |
| Volver desde el formulario abierto en la ficha y reapertura con valores | Pasa | Pasa |
| Lugar común Polideportivo Central con lugar propio Pista 2 | Pasa | Pasa |
| Quitar fecha/hora y vaciar lugar propio restablece la herencia | Pasa | Pasa |
| Administrador designado puede guardar cambios | Pasa | Pasa |
| Lector ve datos existentes sin acciones de edición | Pasa | Pasa |
| Lector sin datos no ve bloques vacíos ni acciones de edición | Pasa | Pasa |
| Enlace directo de edición como lector rechaza el formulario | Pasa | Pasa |
| Teclado de software y recuperación de acceso a los controles | Pasa | Pasa |
| HTTP 500 muestra fallback seguro y conserva campos; reintento recupera | No repetido en nativo | Pasa |

En iOS se pulsaron realmente las teclas `1`, `/` y `:`; el resto de entradas
se estableció mediante accesibilidad. El menú Device → Keyboard → Toggle
Software Keyboard permite revisar el teclado aunque haya uno de hardware.
El gesto sobre el formulario cierra el teclado y deja Guardar accesible.
En Android se observaron el teclado y los campos enfocados; las entradas se
inyectaron mediante ADB. Los resultados y la clasificación no se modificaron:
el único partido siguió pendiente.

La fixture inicialmente devolvía `venue: ""` al borrar el lugar común. El
cliente rechazó correctamente esa proyección inválida. Se corrigió la fixture
para omitir el campo, como hace el backend real, y se repitió el guardado vacío
con éxito. No se cambió código del producto para tolerar la fixture incorrecta.

## Evidencia

- iOS: [guardado](scheduling-qa-2026-10-11/ios-saved.png),
  [teclado](scheduling-qa-2026-10-11/ios-keyboard.png),
  [lector sin datos](scheduling-qa-2026-10-11/ios-reader-empty.png),
  [rechazo de edición](scheduling-qa-2026-10-11/ios-forbidden.png).
- Android: [guardado](scheduling-qa-2026-10-11/android-saved.png),
  [herencia](scheduling-qa-2026-10-11/android-inherited.png),
  [fallback de error](scheduling-qa-2026-10-11/android-error.png),
  [lector sin datos](scheduling-qa-2026-10-11/android-reader-empty.png),
  [rechazo de edición](scheduling-qa-2026-10-11/android-forbidden.png).

## Cierre y retrospectiva

No hay cambios de implementación en esta continuación de QA. Se conservan las
capturas y se cierran Metro, API ficticia, emulador Android y simulador iOS.
No se arrancó observabilidad ni se desplegaron servicios o migraciones.

La indexación de Android Studio llegó a consumir más de seis núcleos. Cerrar
el IDE y continuar por ADB, con un único emulador sin ventana y dos núcleos,
permitió completar la prueba. Conviene separar la compilación de la interacción
QA y cerrar cada dispositivo antes de abrir el siguiente. Los avisos transitorios
deben capturarse al producirse: esperar al volcado de accesibilidad puede perder
el banner, aunque los campos ya hayan recuperado su estado editable.
