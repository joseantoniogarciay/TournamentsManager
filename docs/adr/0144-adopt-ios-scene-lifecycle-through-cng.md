# ADR-0144: Adoptar el ciclo de vida UIScene mediante CNG

- **Estado:** Aceptado
- **Fecha:** 2026-10-03
- **Decisor:** Usuario
- **Supera a:** Ninguno; concreta ADR-0015 para compatibilidad de iOS

## Problema y evidencia

La build local compila con Xcode 27, pero iOS 27 la detiene antes de mostrar la
interfaz: el log de UIKit exige el ciclo de vida UIScene. La plantilla Expo
57.0.15 instalada crea UIWindow en AppDelegate y no declara un scene manifest.
La misma build alcanza el development launcher en iOS 18.5.

## Alternativas y coste

- Actualizar Expo: preferible cuando exista una corrección oficial verificada y
  madura; no se ha confirmado esa corrección y ADR-0138 exige siete días.
- Config plugin local: mantiene CNG y las versiones resueltas; coste limitado a
  revisar la adaptación de plantilla, enlaces y eventos al actualizar Expo.
- Editar Xcode generado: se pierde al regenerar y contradice ADR-0015.
- Posponer iOS 27: impide cumplir el requisito que el usuario prioriza.

## Recomendación y decisión del usuario

Se recomienda un config plugin local, una sola escena y delegación hacia los
puntos existentes de Expo/React Native. El usuario autoriza corregir primero el
requisito de iOS 27 porque puede influir en el resto de la auditoría. No autoriza
una excepción a la política de edad de dependencias ni un cambio de arquitectura
de negocio. Se conserva generación nativa reproducible y no se editan sus salidas.

## Implementación y consecuencias

El plugin declara UIApplicationSceneManifest e incorpora SceneDelegate al
AppDelegate Swift generado. La ventana nace asociada a UIWindowScene; se
conserva su referencia en AppDelegate para las integraciones actuales. Los
eventos de actividad y los enlaces de escena se entregan a los métodos existentes.
No se habilitan varias ventanas. Una plantilla desconocida debe fallar durante
prebuild en lugar de aplicar una transformación parcial silenciosa.

## Validación y revisión

Regeneración idempotente, compilación nativa y arranque en iOS 27 y 18.5; comprobar
entrada por enlace en frío/con app abierta y regreso desde segundo plano. La
auditoría y Desarrollo registran la evidencia y cualquier límite pendiente.
La build Debug arranca en iOS 27 y 18.5; iOS 27 conserva la ruta de esquema en
segundo plano y al arrancar en frío a través del launcher de desarrollo.
Universal Links, sesión autenticada y distribución permanecen pendientes.
Retirar el plugin cuando la plantilla oficial madura incorpore escenas; comparar
entonces enlaces y eventos antes de reemplazarlo. La compilación no acredita
por sí misma los flujos funcionales, la splash release ni dispositivos físicos.

## Fuentes

- [Apple TN3187](https://developer.apple.com/documentation/technotes/tn3187-migrating-to-the-uikit-scene-based-life-cycle)
- Plantilla generada y ExpoAppDelegate de las dependencias locales fijadas.
