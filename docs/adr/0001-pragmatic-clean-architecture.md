# ADR-0001: Clean Architecture pragmática con principios hexagonales

- **Estado:** Aceptado
- **Fecha:** 2026-07-23
- **Decisor:** Usuario, mediante el manifiesto
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

Se necesita una arquitectura que permita aprender diseño backend, mantener la
lógica de negocio independiente de infraestructura y evitar capas o abstracciones
sin utilidad.

## Contexto y restricciones

El backend objetivo es Go y utilizará PostgreSQL y capacidades de infraestructura
local/cloud. El proyecto prioriza simplicidad, testabilidad, operabilidad y
portabilidad razonable.

## Alternativas

### Arquitectura por capas rígida

Familiar y predecible, pero puede forzar capas, mapeos e interfaces aunque no
protejan un límite real.

### Clean/hexagonal pragmática

Protege la dirección de dependencias y permite adaptadores en los bordes. Exige
criterio para no convertir cada detalle en puerto o interfaz.

### Diseño directo acoplado a frameworks e infraestructura

Minimiza estructura inicial, pero mezcla reglas de negocio con detalles externos,
dificulta pruebas relevantes y aumenta el coste de sustitución.

## Decisión del usuario

Adoptar Clean Architecture pragmática con principios hexagonales. Crear interfaces
solo cuando aporten desacoplamiento útil, faciliten pruebas relevantes o existan
varias implementaciones reales. No crear capas innecesarias.

## Consecuencias

### Positivas

- La lógica de negocio no depende de infraestructura.
- Los detalles externos pueden probarse en sus límites.
- La estructura puede crecer con necesidades reales.

### Negativas y riesgos

- “Pragmático” requiere revisión y puede interpretarse de forma inconsistente.
- El afán de pureza puede introducir mapeos e interfaces innecesarios.
- El acoplamiento puede esconderse si no se verifican dependencias.

## Validación

Antes del primer vertical slice se dibujarán dependencias. Las pruebas y revisión
comprobarán que el dominio y los casos de uso no importan adaptadores.

## Disparadores de revisión

- Las reglas generan más ceremonia que protección.
- Los límites del dominio exigen otra descomposición.
- Evidencia de acoplamiento o dificultad de pruebas.

## Documentación afectada

- [ARCHITECTURE.md](../../ARCHITECTURE.md)
- [STYLEGUIDE.md](../../STYLEGUIDE.md)
- [TESTING.md](../../TESTING.md)
