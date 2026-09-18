# ADR-0124: Limitar a tres los torneos recientes de Inicio

- **Estado:** Aceptado
- **Fecha:** 2026-09-13
- **Decisor:** Usuario, mediante decisión explícita
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** ADR-0073, solo respecto al límite del resumen reciente
- **Superado por:** Ninguno

## Problema

El resumen de actividad reciente admite cinco torneos. Al añadir nuevas acciones
útiles debajo, cinco tarjetas desplazan demasiado el resto del contenido de
Inicio y hacen que el resumen compita con ellas.

## Contexto y restricciones

- ADR-0073 conserva la semántica, el orden y la deduplicación de la proyección.
- La colección es un resumen sin paginación, no una biblioteca.
- La sección Torneos sigue ofreciendo las colecciones completas por relación.
- El cliente no debe descargar cinco elementos para ocultar dos localmente.

## Criterios de decisión

1. Conservar visibles las relaciones con actividad más reciente.
2. Reservar espacio vertical para otras acciones de Inicio.
3. Mantener alineados producto, contrato y consulta.
4. Evitar configuración o paginación que el resumen no necesita.

## Alternativas

### A — Mantener cinco

- **Ventajas:** no requiere cambios y muestra más relaciones de inmediato.
- **Inconvenientes:** ocupa más altura y resta visibilidad al contenido siguiente.
- **Coste de mantenimiento:** bajo.

### B — Solicitar cinco y mostrar tres en el cliente

- **Ventajas:** cambio visual local y reversible.
- **Inconvenientes:** el contrato y el payload dejan de representar lo visible.
- **Coste de mantenimiento:** bajo, con una regla duplicada y ambigua.

### C — Limitar la proyección a tres en el servidor

- **Ventajas:** contrato, payload e interfaz comparten un único máximo; conserva
  el resumen compacto sin introducir paginación.
- **Inconvenientes:** dos relaciones menos quedan accesibles solo en Torneos.
- **Coste de adopción y mantenimiento:** bajo.

## Comparación

A prioriza densidad de actividad frente a la nueva jerarquía de Inicio. B reduce
la interfaz pero mantiene una petición innecesariamente mayor. C expresa el
producto con la mínima modificación coherente de extremo a extremo.

## Recomendación

**Opinión/recomendación:** alternativa C.

## Decisión del usuario

**Aceptada el 2026-09-13:** Inicio y `GET /v1/me/recent-tournaments` devuelven y
muestran como máximo tres torneos relacionados, conservando el orden, la
deduplicación y la precedencia de relación definidos en ADR-0073.

## Consecuencias

- El cuarto torneo y posteriores permanecen disponibles en la sección Torneos.
- La consulta PostgreSQL y `maxItems` de OpenAPI comparten el límite de tres.
- No se añade configuración, paginación ni recorte local en el cliente.

## Validación

- Una cuenta con cuatro o más relaciones recibe exactamente las tres de actividad
  más reciente.
- Una cuenta con menos de tres recibe todas sus relaciones disponibles.
- El cliente generado refleja un máximo de tres en el contrato.

## Disparadores de revisión

- La nueva composición de Inicio vuelve a dejar espacio suficiente para más
  actividad sin ocultar acciones principales.
- Los usuarios necesitan navegar actividad histórica desde el propio resumen.

## Documentación afectada

- [Producto](../project/PRODUCT.md)
- [Decisiones](../governance/DECISIONS.md)
- [Aprendizaje](../project/LEARNING.md)
- [Changelog](../../CHANGELOG.md)
