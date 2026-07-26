# ADR-0036: Permitir a administradores corregir resultados con historial

- **Estado:** Aceptado
- **Fecha:** 2026-07-26
- **Decisor:** Usuario
- **Propietario del análisis:** Codex
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

Los resultados registrados por administradores se aplican inmediatamente. Un
error de marcador debe poder corregirse sin paralizar la liga, pero la corrección
no debe eliminar la trazabilidad de lo que ocurrió.

## Contexto y restricciones

- Los administradores delegados son los únicos que gestionan resultados y sus
  registros se aplican de inmediato, según ADR-0034 y ADR-0035.
- El creador puede retirar administradores en cualquier momento.
- Aún no se han definido el formato exacto del marcador ni la clasificación.

## Criterios de decisión

1. corregir errores con poca fricción;
2. conservar responsabilidad y evidencia de cambios;
3. evitar concentrar tareas operativas en el creador;
4. no introducir revisión o arbitraje prematuros.

## Alternativas

### Alternativa A — Corrección directa con historial

Los administradores pueden reemplazar un resultado aplicado y el sistema conserva
quién hizo cada cambio, cuándo y los valores anterior y nuevo.

- **Ventajas:** corrección rápida y trazable; consistente con la delegación ya
  aceptada.
- **Inconvenientes:** un administrador puede realizar varias correcciones antes
  de que el creador actúe.
- **Coste de adopción:** moderado por el historial.
- **Coste de mantenimiento:** bajo a moderado.
- **Riesgos:** discusiones si falta una política posterior de revisión.

### Alternativa B — Solo el creador corrige

- **Ventajas:** control central de cambios posteriores.
- **Inconvenientes:** contradice la delegación operativa y ralentiza la solución
  de errores.
- **Coste de adopción:** bajo.
- **Coste de mantenimiento:** bajo.
- **Riesgos:** dependencia innecesaria del creador.

### Alternativa C — Corrección pendiente de aprobación

- **Ventajas:** revisión adicional antes de cambiar un resultado visible.
- **Inconvenientes:** añade estados y avisos equivalentes a los descartados para
  el registro inicial.
- **Coste de adopción:** moderado.
- **Coste de mantenimiento:** moderado.
- **Riesgos:** resultados obsoletos y procesos bloqueados.

### No cambiar

Un error aplicado no tendría una solución definida y la regla de aplicación
inmediata perdería seguridad operativa.

## Comparación

La alternativa B vuelve a centralizar una tarea que ya se delegó. La C repite la
complejidad de una aprobación que no se requiere en grupos cerrados. La A combina
agilidad con el mínimo control verificable: conservar el historial.

## Recomendación

**Recomendación:** alternativa A, corrección directa con historial.

## Decisión del usuario

**Aceptada el 2026-07-26:** los administradores delegados pueden corregir
directamente resultados ya aplicados. El sistema conserva un historial de quién
cambió el resultado, cuándo y los valores anterior y nuevo.

## Consecuencias

### Positivas

- Los errores se resuelven sin esperar al creador.
- La liga conserva una explicación verificable de cada corrección.

### Negativas y deuda aceptada

- No existe aún un mecanismo de disputa, restauración o revisión externa.
- El historial requiere una política futura de retención y presentación.

## Validación

- Un administrador puede corregir un resultado aplicado y la liga refleja el
  valor nuevo de inmediato.
- Cada corrección registra actor, instante, valor anterior y valor nuevo.
- Un usuario sin rol de administrador no puede registrar ni corregir resultados.
- Retirar la administración impide correcciones posteriores.

## Disparadores de revisión

- Disputas repetidas o correcciones maliciosas.
- Necesidad de restaurar una versión anterior.
- Requisitos de auditoría, arbitraje o aprobación formal.

## Documentación afectada

- [PRODUCT.md](../project/PRODUCT.md)
- [LEARNING.md](../project/LEARNING.md)
- [DECISIONS.md](../governance/DECISIONS.md)
