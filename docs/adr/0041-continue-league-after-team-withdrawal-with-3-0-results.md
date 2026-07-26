# ADR-0041: Continuar la liga tras una baja con resultados 3-0

- **Estado:** Aceptado
- **Fecha:** 2026-07-26
- **Decisor:** Usuario
- **Propietario del análisis:** Codex
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

Una baja durante una liga en curso no debe perjudicar a quienes ya jugaron contra
el equipo retirado ni obligar a cancelar toda la competición.

## Contexto y restricciones

- Los resultados solo existen durante `en_curso` (ADR-0038).
- Los marcadores son simples y sus correcciones guardan historial (ADR-0036 y
  ADR-0037).
- El creador controla el ciclo; los administradores gestionan resultados
  individuales (ADR-0034 y ADR-0040).

## Criterios de decisión

1. mantener la liga operable;
2. tratar igual a todos los rivales del retirado;
3. conservar evidencia de los marcadores previos;
4. evitar clasificación parcial o replanificación compleja.

## Alternativas

### Alternativa A — Continuar con 3-0 uniforme

Solo el creador declara la baja. Los partidos del equipo retirado, pendientes o
ya jugados, se convierten en `3-0` a favor de su rival y guardan historial.

- **Ventajas:** la liga continúa y todos los rivales reciben el mismo trato.
- **Inconvenientes:** altera resultados que pudieron jugarse realmente.
- **Coste de adopción:** moderado por el ajuste masivo y el historial.
- **Coste de mantenimiento:** bajo a moderado.
- **Riesgos:** controversia sobre sustituir un resultado previo.

### Alternativa B — Cancelar la liga

- **Ventajas:** no reescribe marcadores.
- **Inconvenientes:** pierde el progreso de todos los demás equipos.
- **Coste de adopción:** bajo.
- **Coste de mantenimiento:** bajo.
- **Riesgos:** una respuesta desproporcionada ante una sola baja.

### Alternativa C — Respetar jugados y dar 3-0 solo en pendientes

- **Ventajas:** conserva los partidos realizados.
- **Inconvenientes:** trata distinto a los equipos según el orden de calendario.
- **Coste de adopción:** bajo.
- **Coste de mantenimiento:** bajo.
- **Riesgos:** desigualdad competitiva.

### No cambiar

La baja deja partidos y clasificación sin regla consistente.

## Comparación

La B cancela demasiado y la C crea la desigualdad que se quiere evitar. La A
uniforma todos los enfrentamientos y conserva trazabilidad de las sustituciones.

## Recomendación

**Recomendación:** alternativa A, continuar con `3-0` uniforme.

## Decisión del usuario

**Aceptada el 2026-07-26:** si un equipo abandona en `en_curso`, solo el creador
puede declararlo. La liga continúa y todos los partidos de ese equipo, pendientes
o ya jugados, quedan en `3-0` a favor de su rival. Los resultados previos se
conservan en el historial de cambios.

Antes de iniciar, el creador puede quitar un equipo de la liga publicada sin
generar partidos ni aplicar esta regla.

## Consecuencias

### Positivas

- Los demás equipos conservan el progreso de la liga.
- Todos los rivales del retirado reciben el mismo resultado.
- Las modificaciones son auditables.

### Negativas y deuda aceptada

- Un partido jugado puede dejar de reflejar su marcador real.
- No existe aún un motivo tipado de baja ni reincorporación.

## Validación

- Solo el creador declara una baja en `en_curso`.
- Todos y solo los partidos del retirado pasan a `3-0` para el rival.
- Los cambios conservan actor, instante y valores previos.
- La liga continúa y los demás partidos no se modifican.

## Disparadores de revisión

- Necesidad de distinguir abandono, expulsión, sanción o incomparecencia.
- Reglas que prefieran respetar resultados ya jugados.
- Necesidad de reincorporar equipos.

## Documentación afectada

- [PRODUCT.md](../project/PRODUCT.md)
- [LEARNING.md](../project/LEARNING.md)
- [DECISIONS.md](../governance/DECISIONS.md)
