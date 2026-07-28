# ADR-0055: Usar arquitectura cliente feature-first y navegación adaptativa por plataforma

- **Estado:** Aceptado
- **Fecha:** 2026-07-28
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

El cliente universal debe organizar pantallas, formularios y llamadas HTTP sin
trasladar al frontend una arquitectura de backend ni duplicar la navegación entre
web, iOS y Android.

## Alternativas

### A — MVVM estricto por pantalla

- **Ventajas:** patrón conocido por Android y separación explícita de vista y
  estado.
- **Inconvenientes:** los ViewModels duplican hooks y estado de React; añade
  archivos y convenciones antes de que exista complejidad demostrada.
- **Coste de mantenimiento:** medio.

### B — Clean Architecture completa en el cliente

- **Ventajas:** límites muy explícitos y posible utilidad para un dominio offline
  complejo.
- **Inconvenientes:** casos de uso, repositorios e interfaces repetirían una API
  ya contract-first sin aportar valor inicial.
- **Coste de mantenimiento:** alto.

### C — Feature-first pragmática con flujo unidireccional

- **Ventajas:** cada capacidad conserva pantalla, componentes, validación y
  adaptador de API juntos; el código compartido se eleva solo tras repetirse.
- **Inconvenientes:** exige disciplina para no convertir `shared` en un cajón de
  sastre.
- **Coste de mantenimiento:** bajo.

## Decisión del usuario

**Aceptada el 2026-07-28:** adoptar la alternativa C.

- Expo Router expresa rutas y navegación; cada feature contiene sus pantallas,
  hooks, validación local y adaptación al cliente OpenAPI generado.
- La pantalla compone UI; los hooks coordinan estado, envío y feedback; las reglas
  de negocio y autorización permanecen en el backend.
- El estado es local por defecto. Contexto se reserva para sesión y banner global
  cuando existan; no se añaden Redux, Zustand, React Query ni otra librería de
  estado sin evidencia.
- Las rutas profundas conservan una URL canónica común. En web se presentan como
  página directa. En iOS y Android se presentan modalmente sobre la ruta actual.
- Cerrar una ruta profunda móvil vuelve a la pantalla previa si existe; en inicio
  en frío, donde no hay historial que restaurar, vuelve a `/`. En web, cerrar
  vuelve explícitamente a `/`.

## Consecuencias

- Se permite adaptar la presentación sin bifurcar el contrato de enlaces ni la
  lógica de la feature.
- Toda ruta profunda debe definir su estado de inicio en frío y su cierre seguro;
  no se asume una pantalla subyacente inexistente.
- Los tokens y primitivas de ADR-0054 se consumen desde `shared/ui`; no se copian
  estilos dentro de cada feature.

## Validación

- Una misma ruta se abre directamente en web y como modal en móvil.
- Cerrar el modal restaura la ruta previa cuando existe y `/` al abrir en frío.
- Una pantalla no invoca directamente el cliente HTTP generado.
- Dos features no duplican un componente compartido ya estabilizado.

## Disparadores de revisión

- Offline-first, sincronización compleja o edición colaborativa justifican una
  capa de datos o estado adicional.
- Una ruta no puede expresar la presentación requerida mediante Expo Router.
- El directorio `shared` crece sin criterios claros de reutilización.

## Documentación afectada

- [Arquitectura](../engineering/ARCHITECTURE.md)
- [Desarrollo](../engineering/DEVELOPMENT.md)
- [Sistema de diseño](../engineering/DESIGN_SYSTEM.md)
- [Decisiones](../governance/DECISIONS.md)
