# ADR-0082: Persistir co-campeones al finalizar una liga

- **Estado:** Aceptado
- **Fecha:** 2026-08-09
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

Al cerrar una liga con todos los partidos resueltos, la aplicación debe conservar
su resultado deportivo para futuras vistas de perfil sin inventar una ganadora
cuando los criterios de clasificación mantienen un empate en el primer puesto.

## Contexto y restricciones

- ADR-0039 exige cierre explícito con todos los resultados completos.
- ADR-0081 calcula la clasificación en backend y establece que una igualdad
  persistente comparte posición.
- Los equipos no representan todavía membresías o cuentas de personas.

## Alternativas

### A — Un único campeón arbitrario

- **Ventajas:** modelo de lectura mínimo.
- **Inconvenientes:** contradice una posición compartida y falsea el resultado.
- **Coste de mantenimiento:** bajo, con coste de confianza alto.

### B — Co-campeones derivados de la primera posición

- **Ventajas:** conserva exactamente el resultado calculado y permite mostrar
  oro compartido sin un flujo adicional.
- **Inconvenientes:** un perfil futuro deberá representar múltiples campeones.
- **Coste de mantenimiento:** bajo.

### C — Partido de desempate o elección manual

- **Ventajas:** produce una sola ganadora.
- **Inconvenientes:** añade calendario, autoridad y reglas nuevas.
- **Coste de mantenimiento:** alto.

## Recomendación

**Opinión/recomendación:** alternativa B, la mínima coherente con ADR-0081.

## Decisión del usuario

**Aceptada el 2026-08-09:** al finalizar, el backend persiste todos los equipos
que ocupan la posición 1 como co-campeones. No se acepta una ganadora enviada
por cliente, no hay selección manual y el popup final muestra oro compartido
cuando corresponda.

## Consecuencias

- El cierre calcula, persiste y devuelve la clasificación final en una operación
  de backend.
- Un perfil futuro deriva campeonatos de esta relación persistida.
- Plata y bronce se muestran solo cuando existan las posiciones correspondientes.

## Validación

- El cierre rechaza una liga con partidos pendientes.
- Una liga con primera posición compartida persiste todos sus co-campeones.
- Un cliente no puede indicar ni alterar el resultado final.

## Disparadores de revisión

- Se decide un desempate adicional para producir una sola ganadora.
- Los equipos se vinculan a cuentas o participantes individuales.

## Documentación afectada

- [Producto](../project/PRODUCT.md)
- [API](../engineering/API.md)
- [Decisiones](../governance/DECISIONS.md)
- [Aprendizaje](../project/LEARNING.md)
