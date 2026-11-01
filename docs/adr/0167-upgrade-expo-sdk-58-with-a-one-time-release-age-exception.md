# ADR-0167: Migrar a Expo SDK 58 con excepción puntual de maduración

- **Estado:** Aceptado
- **Fecha:** 2026-10-31
- **Decisor:** Usuario mediante decisión explícita en la conversación
- **Complementa:** ADR-0015, ADR-0138, ADR-0144 y ADR-0166.
- **Supera a:** ADR-0138 únicamente para las versiones exactas necesarias en esta migración.

## Problema, contexto y alternativas

El cliente usa Expo 57.0.26 y React Native 0.86.3. Expo SDK 58 incorpora
React Native 0.88, cambios de navegación y soporte del ciclo UIScene de iOS 27.
Actualizar primero los parches del SDK 57 tiene menor coste inmediato y fue la
recomendación inicial. Migrar directamente al SDK 58 evita dos rondas de upgrade,
pero exige revisar API TypeScript, plugins CNG, parches y builds nativas.
Esperar siete días a todo el árbol mantiene ADR-0138 sin excepciones, a costa
de aplazar los hotfixes que el usuario quiere incorporar ahora.

## Decisión explícita

El usuario decide: «Vamos de una al 58 a la 22 porque por esta vez al ser
hotfixes me fio mas de saltar a ella».

Se fija Expo en **58.0.22**, con el conjunto compatible que declara Expo para
ese SDK. Se autoriza por esta vez omitir la maduración de siete días para Expo 58.0.22.
El usuario amplía explícitamente la autorización: «Sí, incluyo sus dependencias
necesarias», incluidas Expo CLI, React Native y herramientas nativas (noxcturnal).
La decisión
no afirma que los hotfixes estén libres de riesgo ni que exista una vulnerabilidad
crítica. No cambia la política general ni autoriza futuras versiones jóvenes.

## Implementación y validación

Conservar lockfile congelado, edad mínima global y rechazo de fechas ausentes.
Registrar exclusiones exactas, revisar el diff y conservar los cambios locales
anteriores. No introducir comodines ni desactivar la política global. Retirar
las excepciones al cumplir siete días. Revisar los parches al cambiar de versión
y retirar los que upstream ya cubra. Mantener CNG y salida web estática. Adaptar la barra de estado de iOS a
UIViewControllerBasedStatusBarAppearance=true, requerido por iOS 27, conservando
el componente raíz y la preferencia de tema existentes.

Validar la matriz con Expo, typecheck, lint, regresiones de dependencias,
exportaciones y generación nativa. Reconstruir las apps para comprobar cambios
nativos; una exportación no acredita funcionamiento visual. Registrar resultados
y límites reales en DEVELOPMENT y LEARNING. No desplegar producción ni activar
observabilidad; al terminar cualquier sesión de pruebas, detener sus servicios.

## Fuentes

- [Expo SDK 58](https://expo.dev/changelog/sdk-58).
- [Actualizar Expo](https://docs.expo.dev/workflow/upgrading-expo-sdk-walkthrough/).
- [Registro de Expo 58.0.22](https://registry.npmjs.org/expo/58.0.22).

## Estado de ejecución

Excepción ampliada explícitamente por el usuario tras el rechazo inicial de la
revisión automática. Implementación completada y validada con Doctor, tipos,
regresiones, exportaciones, builds y arranque básico en iOS/Android. Android Lint
conserva los dos errores previos de ADR-0152; el QA visual completo queda fuera
de esa prueba básica. La sesión temporal autorizada se cierra. Véase el
[informe de migración](../engineering/EXPO_SDK_58_UPGRADE_2026-10-31.md).
