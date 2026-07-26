# ADR-0035: Aplicar inmediatamente los resultados de administradores delegados

- **Estado:** Aceptado
- **Fecha:** 2026-07-26
- **Decisor:** Usuario
- **Propietario del análisis:** Codex
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

Una liga en curso necesita que alguien registre resultados. Hay que decidir si el
marcador que introduce un administrador delegado cambia de inmediato el estado
deportivo o si requiere una confirmación posterior del creador.

## Contexto y restricciones

- Solo los administradores delegados gestionan resultados; no cambian equipos,
  ciclo de vida ni administradores, según ADR-0034.
- El creador asigna directamente a esos administradores y puede retirarlos con
  efecto inmediato.
- La definición detallada de marcadores, correcciones, clasificación y empates
  sigue pendiente.

## Criterios de decisión

1. reflejar los resultados con poca fricción;
2. conservar una responsabilidad clara;
3. evitar estados intermedios y tareas de confirmación innecesarias;
4. permitir una futura corrección trazable;
5. mantener la solución adecuada a ligas cerradas de confianza.

## Alternativas

### Alternativa A — Aplicación inmediata

El resultado registrado por un administrador delegado actualiza la liga en el
mismo momento.

- **Ventajas:** flujo simple, resultado visible al instante y sin cola de tareas
  para el creador.
- **Inconvenientes:** un error se publica hasta que se corrija.
- **Coste de adopción:** bajo.
- **Coste de mantenimiento:** bajo; una futura auditoría añadirá datos, no un
  flujo de aprobación.
- **Riesgos:** delegar en una persona no fiable o cometer un error de entrada.

### Alternativa B — Propuesta pendiente de confirmación del creador

- **Ventajas:** segunda revisión antes de afectar a la liga.
- **Inconvenientes:** crea estados pendientes, avisos y retrasos; el creador ya
  eligió al administrador y puede revocarlo.
- **Coste de adopción:** moderado.
- **Coste de mantenimiento:** moderado por recordatorios y propuestas sin cerrar.
- **Riesgos:** resultados desactualizados y carga administrativa innecesaria.

### Alternativa C — Solo el creador registra resultados

- **Ventajas:** máxima centralización y control.
- **Inconvenientes:** contradice la delegación de resultados y concentra el
  trabajo en el creador.
- **Coste de adopción:** bajo.
- **Coste de mantenimiento:** bajo.
- **Riesgos:** inutilizar la función de administrador delegado.

### No cambiar

Sin esta regla, no se puede definir el efecto de registrar un resultado ni la
actualización de una liga seguida.

## Comparación

La alternativa B aporta un control adicional que no compensa su estado y
operación extra en grupos cerrados. La C no aprovecha la delegación aceptada. La
A encaja con una relación de confianza explícita y conserva la revocación como
salida inmediata.

## Recomendación

**Recomendación:** alternativa A, aplicación inmediata.

## Decisión del usuario

**Aceptada el 2026-07-26:** un resultado introducido por un administrador
delegado se aplica inmediatamente, sin confirmación del creador.

## Consecuencias

### Positivas

- Las personas que siguen la liga ven el resultado actualizado sin espera.
- No existen resultados pendientes ni una bandeja de confirmaciones.
- La delegación de resultados tiene utilidad práctica.

### Negativas y deuda aceptada

- Un marcador erróneo se hace visible hasta su corrección.
- Todavía deben definirse la corrección, auditoría y reglas de clasificación.

## Validación

- Un administrador autorizado registra un resultado y la liga refleja el cambio
  de inmediato.
- El creador no necesita ninguna acción de confirmación.
- Un usuario sin permisos de administrador no puede registrar resultados.
- Retirar a un administrador impide nuevos registros inmediatamente.

## Disparadores de revisión

- Errores frecuentes o disputas de resultados.
- Necesidad de arbitraje, doble confirmación o revisión externa.
- Requisitos de auditoría que exijan un historial inmutable más completo.

## Documentación afectada

- [PRODUCT.md](../project/PRODUCT.md)
- [LEARNING.md](../project/LEARNING.md)
- [DECISIONS.md](../governance/DECISIONS.md)
