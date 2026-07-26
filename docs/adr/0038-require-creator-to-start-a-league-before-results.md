# ADR-0038: Requerir que el creador inicie la liga antes de registrar resultados

- **Estado:** Aceptado
- **Fecha:** 2026-07-26
- **Decisor:** Usuario
- **Propietario del análisis:** Codex
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

Hay que definir quién inicia una liga y desde qué estado se pueden registrar
resultados para que `publicado` y `en_curso` tengan significado distinto.

## Contexto y restricciones

- Publicar genera partidos y fija composición (ADR-0032).
- Solo administradores delegados gestionan resultados (ADR-0034).
- Sus resultados y correcciones se aplican de inmediato (ADR-0035 y ADR-0036).

## Criterios de decisión

1. impedir resultados antes del comienzo real;
2. conservar en el creador las transiciones estructurales;
3. no exigir calendario ni tareas programadas.

## Alternativas

### Alternativa A — Inicio explícito por el creador

El creador ejecuta `publicado → en_curso`; solo en `en_curso` se registran o
corrigen resultados.

- **Ventajas:** frontera clara entre lista para compartir y competición activa.
- **Inconvenientes:** exige una acción adicional del creador.
- **Coste de adopción:** bajo.
- **Coste de mantenimiento:** bajo.
- **Riesgos:** una liga puede quedar publicada sin iniciar.

### Alternativa B — Resultados desde publicado

- **Ventajas:** menos acciones.
- **Inconvenientes:** elimina el significado operativo de `en_curso`.
- **Coste de adopción:** bajo.
- **Coste de mantenimiento:** bajo.
- **Riesgos:** resultados introducidos accidentalmente antes de empezar.

### Alternativa C — Inicio automático por fecha

- **Ventajas:** automatización cuando exista calendario.
- **Inconvenientes:** depende de fechas, zonas horarias y tareas fuera del
  alcance inicial.
- **Coste de adopción:** alto.
- **Coste de mantenimiento:** moderado.
- **Riesgos:** automatizar una capacidad aún no modelada.

### No cambiar

Permisos de resultado y significado de `en_curso` seguirían ambiguos.

## Comparación

La B simplifica una acción a cambio de una transición sin utilidad. La C requiere
capacidades todavía ausentes. La A aprovecha el ciclo aceptado y conserva un
inicio explícito y simple.

## Recomendación

**Recomendación:** alternativa A, inicio explícito por el creador.

## Decisión del usuario

**Aceptada el 2026-07-26:** el creador inicia una liga publicada para moverla a
`en_curso`. Solo en `en_curso` los administradores delegados pueden registrar o
corregir resultados.

Las modificaciones de equipos, vueltas u otras reglas después de publicar siguen
prohibidas en el primer corte. Una futura capacidad de cambio deberá preservar o
regenerar partidos sin invalidar resultados.

## Consecuencias

### Positivas

- `publicado` significa listo para compartir y `en_curso`, competición activa.
- El creador decide cuándo empieza la liga.

### Negativas y deuda aceptada

- Una liga no empieza sola y puede quedar publicada indefinidamente.
- No hay calendario, recordatorios ni cambios posteriores a publicar.

## Validación

- Solo el creador puede ejecutar `publicado → en_curso`.
- Registrar o corregir resultados fuera de `en_curso` se rechaza.
- Los administradores no pueden iniciar ni alterar el ciclo de vida.

## Disparadores de revisión

- Se introduce calendario y se necesita inicio automático.
- Se requieren cambios de equipos o número de vueltas tras publicar.
- Se acumulan ligas publicadas olvidadas.

## Documentación afectada

- [PRODUCT.md](../project/PRODUCT.md)
- [LEARNING.md](../project/LEARNING.md)
- [DECISIONS.md](../governance/DECISIONS.md)
