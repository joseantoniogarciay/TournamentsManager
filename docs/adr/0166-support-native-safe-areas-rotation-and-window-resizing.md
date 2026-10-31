# ADR-0166: Respetar áreas seguras, orientaciones y ventanas redimensionables en apps

- **Estado:** Aceptado
- **Fecha:** 2026-10-24
- **Decisor:** Usuario
- **Supera a:** La restricción de orientación vertical de la configuración nativa.
- **Complementa:** ADR-0008, ADR-0054, ADR-0055 y ADR-0151.

## Problema y evidencia

La configuración fija portrait y no declara soporte de tablet iOS. Screen
protege arriba/abajo, pero omite los insets laterales. Los diálogos React Native
aceptan únicamente portrait por defecto y el cuadro decide su distribución por
el ancho de ventana, que puede superar el ancho de su contenedor.
En dispositivos como iPhone Duo las oclusiones pueden quedar a la derecha;
rotación y Split View no se resuelven suponiendo una barra superior fija.

## Alternativas y coste

- Corregir por pantalla y por modelo: coste creciente, reglas duplicadas y fallos
  al aparecer una postura o dispositivo nuevo.
- Primitivas compartidas con insets dinámicos y medición del contenedor: coste
  bajo, conserva el diseño existente y responde a rotación y redimensionado.
- Rediseñar toda la navegación y añadir layouts especiales por dispositivo:
  coste alto, innecesario sin evidencia de una necesidad concreta.

## Decisión explícita

**Aceptada mediante la petición del usuario del 2026-10-24:** optimizar las apps
móviles y tablet para todas las orientaciones y Split View, especialmente iPhone
Duo; sumar siempre el área segura disponible al margen de contenido y mantener
esta obligación para futuros desarrollos. Se adopta la segunda alternativa.
La web conserva sus reglas de ADR-0154.

Screen reserva en nativo los insets físicos izquierdo y derecho, sin añadir
margen de diseño: Card y los contenidos conservan sus 20 px. Los fondos llegan al
borde y un carrusel puede recorrer todo el espacio libre de oclusiones. No se
fijan insets por modelo ni se infiere su lado a partir de la orientación.
La navegación conserva su responsabilidad sobre la cabecera; el contenido bajo
ella no duplica el inset superior. Los overlays se protegen en su propia ventana.
Cada escena de navegación nativa tiene su propio SafeAreaProvider: el provider
de una ruta cubierta puede conservar insets anteriores a un giro. El banner iOS
mide dentro de FullWindowOverlay y suma el margen de diseño dentro de ese marco.
Los diálogos admiten las cuatro orientaciones y permiten desplazarse cuando la
altura disponible disminuye. El cuadro mide su contenedor para elegir entre
vista general y ronda. No se remonta una pantalla por rotación o cambio de ancho.

Expo CNG declara orientación default, supportsTablet y requireFullScreen=false.
Android concreta fullUser y resizeableActivity=true mediante un config plugin
CNG mínimo: admite cuatro giros respetando el bloqueo de rotación de la persona.
Los cambios nativos requieren regenerar, compilar e instalar; Fast Refresh no
valida capacidades del binario. Split View no implica habilitar varias escenas.

## Prerrequisito de compilación observado

Xcode con Swift 6.4 emite `extension __ObjC::expo...` en ExpoModulesJSI 57.1.1.
El script instalado ya elimina esas extensiones privadas para compatibilidad,
pero su filtro solo reconoce `__ObjC.`. Dos builds fallan al importar la interfaz,
también sin módulos Swift explícitos. Se amplía únicamente ese filtro mediante
parche pnpm versionado; no cambia API pública ni versiones. La regresión ejecuta
el filtro real con ambas sintaxis y comprueba que conserva declaraciones públicas.
Retirar el parche cuando el script upstream cubra ambas sintaxis y la build pase.

## Validación y límites

Typecheck, lint y exportación web; comprobar la configuración nativa generada.
Recorrer iPhone Duo cerrado/abierto, ambas orientaciones horizontales y verticales
cuando el SO las permita, tablet y Split View estrecho/ancho. Revisar formularios,
teclado, diálogos, tablas/cuadro, extremos de carruseles y botones flotantes;
conservar estado y acceso al último control. La evidencia de ejecución y lo
pendiente se registran por separado en el informe de revisión, sin certificar
una matriz completa a partir de capturas aisladas.

## Fuentes

- [Configuración Expo](https://docs.expo.dev/versions/latest/config/app/).
- [Modal React Native](https://reactnative.dev/docs/modal).
- [Activity Android](https://developer.android.com/guide/topics/manifest/activity-element).
- Implementación de Orientation de @expo/config-plugins 57.0.9 instalada.

## Documentación afectada

Checklist cliente, sistema de diseño, desarrollo, decisiones y aprendizaje.
