# ADR-0070: Configurar la liga al iniciarla

- **Estado:** Aceptado
- **Fecha:** 2026-08-01
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** ADR-0032 y ADR-0040, exclusivamente en el momento y opciones de configuración de competición
- **Superado por:** Ninguno

## Problema

Una liga recién creada debe poder compartirse y completar sus equipos sin que
el organizador tenga que fijar todavía el calendario ni decidir si se juega a
una o dos vueltas.

## Contexto y restricciones

- `published` representa una liga creada pero sin empezar; la interfaz la
  denomina «Sin empezar».
- Solo el creador puede iniciarla y los equipos quedan congelados después.
- El inicio genera los partidos una única vez y habilita resultados.

## Criterios de decisión

1. Expresar claramente cuándo empieza la competición.
2. Evitar regenerar partidos o reinterpretar resultados.
3. Ofrecer configuración útil sin diseñar reglas deportivas especulativas.

## Alternativas

### A — Configurar al crear

- **Ventajas:** una única acción de creación.
- **Inconvenientes:** obliga a cerrar decisiones mientras aún se prepara la
  liga y confunde creación con inicio.
- **Coste de mantenimiento:** bajo.

### B — Configurar al iniciar

- **Ventajas:** separa la preparación de la competición y fija el calendario
  en su límite natural.
- **Inconvenientes:** el inicio necesita un formulario adicional.
- **Coste de mantenimiento:** bajo.

### C — Cambiar configuración con la liga en curso

- **Ventajas:** flexibilidad máxima.
- **Inconvenientes:** exige regeneración, reconciliación de resultados y reglas
  de clasificación no necesarias.
- **Coste de mantenimiento:** alto.

## Comparación

B conserva la liga editable antes de competir y evita la complejidad de C.

## Recomendación

**Opinión/recomendación:** alternativa B.

## Decisión del usuario

**Aceptada el 2026-08-01:** la liga se crea en `published` («Sin empezar»). Al
iniciarla, el creador elige una o dos vueltas; la puntuación permanece 3-1-0 en
este corte. El servidor valida la configuración, la guarda, genera los partidos
atómicamente y cambia el estado a `in_progress`. Después no se modifican equipos
ni configuración.

## Consecuencias

### Positivas

- La creación es ligera y una liga puede prepararse antes de jugarse.
- Calendario, configuración y comienzo tienen una frontera única y comprobable.

### Negativas y deuda aceptada

- No hay fechas, horarios, desempates ni reglas de puntuación configurables.

## Validación

- Una liga publicada no contiene partidos.
- Empezar con una o dos vueltas crea todos los emparejamientos una vez.
- Un segundo inicio y cualquier cambio posterior son rechazados.

## Disparadores de revisión

- Necesidad de más reglas antes del inicio.
- Formatos distintos de liga.

## Documentación afectada

- [Producto](../project/PRODUCT.md)
- [API](../engineering/API.md)
- [Modelo inicial de datos](../engineering/INITIAL_DATA_MODEL.md)
- [Decisiones](../governance/DECISIONS.md)
- [Aprendizaje](../project/LEARNING.md)
