# ADR-0151: Adaptar los títulos de navegación al espacio de la barra iOS

- **Estado:** Aceptado
- **Fecha:** 2026-10-09
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** Ninguno; concreta la adaptación accesible de ADR-0054 y ADR-0055.
- **Superado por:** Ninguno

## Problema

La ficha de torneo permite dos líneas en un título React Native dentro de la
barra UIKit. La evidencia privada ios-release-text-size-11-20261008.png muestra
recorte vertical; el mismo fallo se observó a Text Size 7. Limitar el ancho
conserva separados los controles, pero no aumenta la altura de la barra.

## Contexto y restricciones

Conservar Figtree, escalado de texto y los botones nativos de Stack.Toolbar
en iOS 26 o superior. La regla vigente exige título centrado y separación
de 20 px frente a controles. El cambio debe permitir leer el nombre completo
sin reducir el tamaño solicitado por la persona. Android y web ya tienen su
propia disposición; no se cambia API, navegación ni lógica del torneo.

## Criterios de decisión

Lectura completa, adaptación solo cuando resulte necesaria, controles nativos,
escalado intacto, regla compartida y coste de mantenimiento mínimo.

## Alternativas

### A — Cabecera adaptable debajo de la barra cuando el título no quepa

- Ventajas: conserva controles nativos y escalado; el nombre puede ocupar la
  altura necesaria sin competir con Cerrar o Acciones.
- Inconvenientes: el nombre cambia de ubicación con texto ampliado y consume
  más espacio vertical.
- Adopción: componente mínimo compartido y ajuste en ficha de torneo.
- Mantenimiento: bajo; comprobar tamaño normal, ampliado, idiomas y scroll.
- Riesgos: evitar duplicar el nombre, saltos de layout y omitirlo a lectores.

### B — Sustituir toda la barra por una cabecera React Native

- Ventajas: control completo de altura y disposición.
- Inconvenientes: sustituye controles y comportamiento nativos aceptados.
- Adopción y mantenimiento: mayores; insets, gestos, transiciones y botones
  tendrían que mantenerse y probarse por plataforma.

### No cambiar

El nombre continúa recortado. Reducir o desactivar el escalado no cumple el
objetivo accesible y no se recomienda como arreglo.

## Comparación y recomendación

**Opinión/recomendación:** A, por conservar la navegación aceptada y resolver
la necesidad de espacio con un layout que pueda crecer. La activación debe
basarse en el espacio disponible y el tamaño de texto, sin un ancho fijo ni
una reducción del escalado. 

## Decisión del usuario

**Aceptada el 2026-10-09:** alternativa A. El usuario precisa que solo debe
activarse cuando no haya otra opción: se mide la altura real del título con
el ancho reservado y se compara con la altura disponible de la barra. No se
activa por un umbral arbitrario de fontScale.

## Consecuencias

Se concreta una excepción de posición del título en iOS con texto ampliado.
La barra mantiene sus acciones; la cabecera accesible muestra una única copia
del nombre completo y conserva los tokens y márgenes compartidos.

## Validación

Simulador iOS firmado: título completo a tamaños normal, 7 y 11; controles
separados y operables; nombre largo en es/en/it/fr; scroll, cambios de ancho
y rotación cuando la orientación se admita, y lector cuando el canal permita
acreditarlo. Typecheck, exportación web y gate común.
No declarar certificada la matriz de accesibilidad por una captura.

## Disparadores de revisión

Un cambio de UIKit o react-native-screens permite crecimiento nativo, aparece
un salto al cambiar tamaño, o el título deja de estar disponible al navegar.

## Fuentes y documentación afectada

- [Native Stack](https://reactnavigation.org/docs/native-stack-navigator/): un
  header completamente personalizado pierde capacidades de la barra nativa.
- RNSScreenStackHeaderConfig.mm de la dependencia instalada asigna el título
  personalizado a UINavigationItem.titleView. La captura prueba el recorte;
  esa asignación aporta contexto, no demuestra por sí sola toda su causa.
- DESIGN_SYSTEM.md, PRODUCT_QA_REVIEW_2026-10-04.md y LEARNING.md.

**Aclaración de alcance del usuario — 2026-10-09:** la regla se comparte con
las demás pantallas que muestran título de navegación, además de las entidades.
Screen ofrece un título adaptable optativo; no inventa títulos donde la barra
no los muestra. El componente solo cambia de posición tras medir que no cabe.
