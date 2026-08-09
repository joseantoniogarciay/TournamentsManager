# ADR-0083: Materializar resultados oficiales por cuenta al finalizar una liga

- **Estado:** Aceptado
- **Fecha:** 2026-08-09
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

Un futuro perfil debe mostrar participaciones, podios y estadísticas fiables sin
aceptar cifras enviadas por cliente ni recalcular todo el historial en cada lectura.

## Contexto y restricciones

- ADR-0039 hace final e inmutable una liga con resultados completos.
- ADR-0081 calcula clasificación en backend; ADR-0082 permite co-campeones.
- Una cuenta pertenece a un único equipo dentro de una misma liga.
- La vinculación de cuentas a equipos antes del inicio se decide e implementa
  posteriormente.

## Alternativas

### A — Contadores globales editables en `accounts`

- **Ventajas:** perfil de lectura inmediata.
- **Inconvenientes:** no conserva el origen de cada cifra ni una reparación trazable.
- **Coste de mantenimiento:** medio.
- **Riesgos:** doble aplicación y desincronización.

### B — Calcular el historial al consultar el perfil

- **Ventajas:** no persiste derivados.
- **Inconvenientes:** coste creciente y no fija explícitamente el resultado oficial.
- **Coste de mantenimiento:** creciente.
- **Riesgos:** consultas costosas.

### C — Resultado oficial inmutable por cuenta y liga

- **Ventajas:** trazabilidad por torneo y agregación eficiente del perfil.
- **Inconvenientes:** tabla derivada y cierre atómico adicionales.
- **Coste de mantenimiento:** bajo a moderado.
- **Riesgos:** proteger unicidad y transacción.

## Comparación

La A optimiza prematuramente y la B traslada el coste a cada lectura. La C fija
una vez el resultado oficial y conserva su origen.

## Recomendación

**Opinión/recomendación:** alternativa C.

## Decisión del usuario

**Aceptada el 2026-08-09:** al finalizar una liga, el backend crea en la misma
transacción un resultado oficial inmutable por cada cuenta vinculada a un equipo.
La fila conserva liga, cuenta, equipo, posición y estadísticas derivadas. El
perfil suma estas filas; no hay contadores editables en `accounts`. Una cuenta
solo puede estar en un equipo de la liga. Las cuentas de todos los equipos en
posición 1 reciben oro en un co-campeonato.

## Consecuencias

- El cierre oficializa estadísticas personales una única vez.
- Las membresías deben congelarse antes de iniciar.
- Cada medalla conserva liga y equipo de origen.

## Validación

- Una cuenta genera como máximo un resultado por liga.
- El cierre solo persiste valores derivados por backend.
- Las cuentas de co-campeones reciben oro.

## Disparadores de revisión

- Correcciones después de finalizar.
- Una cuenta en más de un equipo de la misma liga.
- Métricas no derivables de resultados y membresías.

## Documentación afectada

- [Producto](../project/PRODUCT.md)
- [Modelo inicial de datos](../engineering/INITIAL_DATA_MODEL.md)
- [API](../engineering/API.md)
- [Decisiones](../governance/DECISIONS.md)
- [Aprendizaje](../project/LEARNING.md)
