# ADR-0132: Organizar los adaptadores Go por capacidad

- **Estado:** Aceptado
- **Fecha:** 2026-09-20
- **Decisor:** Usuario, mediante aceptación explícita de la propuesta completa
- **Propietario del análisis:** Asistente como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

El backend conserva la dirección general de dependencias, pero dos ficheros de
los adaptadores concentraban aproximadamente el 37 % del Go de producción no
generado. El router HTTP mezclaba configuración y handlers de varias capacidades,
y un único fichero PostgreSQL reunía cuentas, acceso, sesiones, notificaciones y
torneos. Buscar un comportamiento exigía conocer su ubicación histórica.

## Contexto y restricciones

ADR-0001 exige Clean Architecture pragmática y ADR-0007 un monolito modular. El
manifiesto prohíbe añadir capas o abstracciones por simetría. La refactorización
no cambia contrato HTTP, persistencia, comportamiento ni unidad de despliegue.

## Criterios de decisión

1. navegación humana por capacidad;
2. dirección de dependencias verificable;
3. cambio reversible y sin riesgo funcional;
4. mínimo coste de mantenimiento;
5. ausencia de frameworks o paquetes artificiales.

## Alternativas

### Mantener los ficheros concentrados

- **Ventaja:** ningún cambio inmediato.
- **Inconveniente:** búsqueda y revisión empeoran con cada endpoint.
- **Mantenimiento:** creciente y dependiente de conocimiento tácito.

### Dividir ficheros dentro de los paquetes actuales

- **Ventajas:** mejora la localización sin alterar límites ni tipos; permite
  proteger imports con una prueba pequeña.
- **Inconvenientes:** aumenta el número de ficheros y exige nombres coherentes.
- **Mantenimiento:** bajo; Go compila el paquete como la misma unidad.

### Crear subpaquetes por cada capacidad y adaptador

- **Ventaja:** límites más rígidos.
- **Inconvenientes:** migración amplia, más API interna y riesgo de ciclos o
  interfaces ceremoniales.
- **Mantenimiento:** desproporcionado para el tamaño actual.

## Comparación

La división interna ofrece navegación y revisiones más acotadas sin pagar una
nueva topología. Mantener el estado actual no resuelve la evidencia medida y los
subpaquetes anticipan una necesidad que todavía no existe.

## Recomendación

**Opinión/recomendación:** dividir HTTP, PostgreSQL y sus pruebas por capacidad
dentro de sus paquetes actuales; usar una configuración explícita para construir
el handler; situar errores de negocio en su capacidad; y comprobar que negocio
no importa adaptadores ni HTTP importa PostgreSQL.

## Decisión del usuario

**Aceptada:** la propuesta completa el 2026-09-20.

## Consecuencias

### Positivas

- Los nombres de fichero indican dónde buscar cada comportamiento.
- El transporte deja de depender del adaptador PostgreSQL.
- La composición HTTP distingue configuración y dependencias.
- Una prueba automática conserva las reglas de imports.

### Negativas y deuda aceptada

- `AccountTournamentRepository` sigue siendo una estructura compartida para no
  duplicar pool y consultas; los ficheros separan responsabilidades sin forzar
  todavía varios tipos.
- El constructor corto de HTTP permanece como ayuda de pruebas mientras la
  composición real usa la configuración explícita.
- Los nombres históricos `league` solo se corregirán cuando el cambio sea local
  y no altere contratos o historia.

## Validación

- `go test ./...` incluye las reglas de arquitectura y comportamiento.
- `golangci-lint` y formato permanecen limpios.
- Ningún fichero de producción HTTP importa PostgreSQL.
- Los antiguos puntos de concentración quedan divididos por capacidad.

## Disparadores de revisión

- Un fichero de adaptador vuelve a mezclar varias capacidades.
- Los tipos compartidos impiden evolución o propiedad clara de datos.
- Aparecen ciclos entre capacidades o excepciones repetidas a las reglas.

## Documentación afectada

- [Arquitectura](../engineering/ARCHITECTURE.md)
- [Guía del backend](../../apps/backend/README.md)
- [Decisiones](../governance/DECISIONS.md)
- [Aprendizaje](../project/LEARNING.md)
- [Changelog](../../CHANGELOG.md)
