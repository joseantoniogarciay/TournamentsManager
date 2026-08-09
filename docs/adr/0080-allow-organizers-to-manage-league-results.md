# ADR-0080: Permitir al organizador gestionar resultados de su liga

- **Estado:** Aceptado
- **Fecha:** 2026-08-09
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** ADR-0035, exclusivamente en quién puede registrar y corregir resultados
- **Superado por:** Ninguno

## Problema

La persona que crea una liga no puede registrar sus resultados, aunque es quien
controla su ciclo, equipos y administradores. La interfaz además mostraba un
estado «Pendiente» a personas sin permiso, mezclando información operativa con
lectura pública.

## Contexto y restricciones

- ADR-0034 mantiene al organizador como propietaria de la liga y a los
  administradores delegados como operadores de resultados.
- ADR-0035 limitaba los resultados a administradores delegados; esta decisión
  revisa solo esa limitación.
- Los resultados siguen requiriendo una liga `in_progress`, se aplican de
  inmediato y mantienen su historial interno.

## Alternativas

### A — Mantener solo administradores delegados

- **Ventajas:** separación estricta de responsabilidades.
- **Inconvenientes:** el organizador debe delegar incluso en una liga que opera
  personalmente.
- **Coste de mantenimiento:** bajo.

### B — Organizador y administradores delegados

- **Ventajas:** el organizador actúa como superadministrador sin crear un rol
  nuevo; conserva la delegación para repartir trabajo.
- **Inconvenientes:** el permiso de resultados deja de pertenecer solo al rol
  delegado.
- **Coste de mantenimiento:** bajo; una comprobación de autorización reúne las
  dos relaciones existentes.

### No cambiar

El organizador no puede completar por sí mismo una liga y la pantalla muestra
un estado operativo a personas que solo pueden leer.

## Comparación

La alternativa B satisface la gestión directa de ligas pequeñas sin añadir una
matriz de permisos, tablas ni estado de producto. A solo conserva una
restricción que ya no aporta valor.

## Recomendación

**Opinión/recomendación:** B, por ser el mínimo cambio coherente con la
propiedad ya aceptada.

## Decisión del usuario

**Aceptada el 2026-08-09:** el organizador y los administradores delegados
pueden registrar y corregir resultados. Una persona sin una de esas relaciones
solo ve el cruce y el resultado existente, sin controles ni estado «Pendiente».

## Consecuencias

### Positivas

- El organizador puede operar su propia liga sin delegación artificial.
- La lectura pública no comunica tareas a quien no puede realizarlas.

### Negativas y deuda aceptada

- Cualquier rol administrativo futuro debe definir explícitamente si hereda
  este permiso.

## Validación

- Organizador y administradora delegada pueden registrar y corregir resultados.
- Una cuenta ajena recibe `403` y no ve controles de resultado ni «Pendiente».
- La liga debe seguir en curso para cualquiera de las dos relaciones.

## Disparadores de revisión

- Se introducen más roles o permisos configurables por liga.
- Se exige aprobación de resultados o separación entre registrar y corregir.

## Documentación afectada

- [Producto](../project/PRODUCT.md)
- [API](../engineering/API.md)
- [OpenAPI v1](../../contracts/openapi/v1/openapi.yaml)
- [Decisiones](../governance/DECISIONS.md)
- [Aprendizaje](../project/LEARNING.md)
