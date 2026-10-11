# ADR-0160: Elegir equipos libres o crear otros mediante invitación

- **Estado:** Aceptado
- **Fecha:** 2026-10-11
- **Decisor:** Usuario, mediante «En repetir o añadiendo debería de darse la opción de elegir equipos no vinculados o crear nuevo si hay hueco»
- **Supera parcialmente:** ADR-0130 y ADR-0159, en el alcance de las inscripciones mediante invitación

## Problema y contexto

La renovación de inscripciones de una edición repetida revela una necesidad
común: los equipos que la organizadora añade sin cuenta también deben poder
ser elegidos por participantes invitados. El nombre libre sin lista obliga a
adivinar qué equipos están disponibles. El torneo publicado sigue permitiendo
componer equipos; iniciado deja de admitir inscripciones.

## Alternativas, mantenimiento y recomendación

- Solo crear nombres nuevos: coste nulo, impide ocupar equipos preparados.
- Permitir elegir libres solo en repeticiones: coste moderado y dos reglas de
  inscripción distintas sin necesidad.
- Elegir un equipo sin cuenta o crear otro en cualquier invitación: misma regla
  transaccional, coste moderado inicial y mantenimiento bajo. Recomendación.

## Decisión explícita y consecuencias

El usuario acepta la tercera alternativa para repetir e incorporar participantes.
La invitación muestra nombres disponibles (sin identidades de cuenta). Pulsar
uno completa el campo de nombre; escribir otro conserva el alta actual si hay
menos de 64 equipos. Ocupar un equipo existente no consume una plaza adicional,
también con 64 equipos. Se rechaza un nombre ocupado o una cuenta ya inscrita;
la organizadora ya tiene su equipo y no obtiene un segundo por esta vía.

La disponibilidad es una fotografía: inscripción, inicio y revocación verifican
el estado bajo el mismo lock de torneo. Quien recibe el enlace puede ocupar un
equipo libre; el nombre no demuestra identidad ni participación anterior.
La organizadora controla a quién comparte o revoca el enlace. La elección crea
la relación de cuenta y seguimiento, sin otorgar administración ni arrastrar
permisos de otra edición. Sin equipos libres ni capacidad, la invitación deja
de estar disponible. No se añade un directorio global ni un catálogo entre torneos.

## Validación

Elegir sin aumentar el número de equipos, nombre ocupado, cuenta duplicada,
nombre nuevo con capacidad, lista libre actualizada, capacidad 64, carreras de
reclamación, inicio/revocación concurrentes y conservación de equipos originales.
HTTP e inspección no exponen IDs de cuenta; errores y cancelación son seguros.
Localización y UI deben mostrar opciones y conservar nombre libre.

## Disparadores de revisión

Reclamaciones que necesiten consentimiento o prueba de identidad, equipos
persistentes entre torneos o inscripciones de varias personas por equipo.

## Documentación afectada

PRODUCT.md, API.md, DATABASE.md, DESIGN_SYSTEM.md, OBSERVABILITY.md,
DECISIONS.md, LEARNING.md y CHANGELOG.md.

## Concreción aceptada: plaza ocupada

El usuario confirma que una opción visible puede quedar ocupada antes de enviar.
Se responde 409 con el tipo RFC 9457 `urn:tournaments-manager:problem:team-occupied`.
El cliente muestra «Esta plaza ya ha sido ocupada. Elige otro equipo disponible.»,
conserva el nombre y vuelve a inspeccionar los equipos libres. Otros conflictos
no usan ese mensaje; cuerpos inválidos o tipos desconocidos conservan el fallback.
