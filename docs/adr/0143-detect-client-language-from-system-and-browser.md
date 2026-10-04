# ADR-0143: Detectar el idioma desde el sistema y el navegador

- **Estado:** Aceptado
- **Fecha:** 2026-10-03
- **Decisor:** Usuario
- **Supera parcialmente a:** ADR-0056, solo en la selección de idioma web.

## Problema y contexto

La auditoría visual señalaba como pendiente el selector web que exigía ADR-0056.
El usuario ha aclarado que el idioma debe depender del SO en móvil y del navegador
en web, y ha excluido el selector de las correcciones autorizadas.

## Alternativas y coste

- Mantener un selector web persistente: permite una preferencia propia, pero
  añade estado, almacenamiento y sincronización a la detección existente.
- Detectar automáticamente en ambas plataformas: conserva el comportamiento
  actual y tiene menor coste de mantenimiento.

## Recomendación y decisión del usuario

El usuario acepta la detección automática: SO en iOS/Android y navegador en web.
No se ofrece selector de idioma. Se mantienen es/en/it/fr y fallback inglés.
Las preferencias de tema de ADR-0056 no cambian.

## Consecuencias y validación

No hace falta modificar locale.ts: ya consume expo-localization y selecciona
el primer idioma detectado si está soportado, o inglés en caso contrario.
La exportación estática conserva inglés hasta la resolución en cliente.
La auditoría deja de tratar la ausencia del selector como defecto.
Esta fase no certifica todas las traducciones ni cambios de idioma en caliente.

## Disparadores de revisión

Una necesidad explícita de idioma independiente del dispositivo o navegador.

## Documentación afectada

ADR-0056, reglas del cliente, auditoría visual y aprendizaje.
