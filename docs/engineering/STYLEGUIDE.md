# Guía de estilo

## Principio rector

El código y la documentación optimizan primero para comprensión. Una solución
breve no es simple si oculta decisiones, y una abstracción elegante no se acepta
si no resuelve una necesidad actual.

## Documentación

- Español para el handbook; nombres técnicos y de código en inglés cuando sean
  convenciones del ecosistema.
- Un documento empieza con propósito o estado cuando pueda confundirse con una
  decisión final.
- “Aceptado”, “propuesto”, “estándar”, “evidencia” y “opinión” se usan con
  significado explícito.
- Enlaces relativos y títulos descriptivos.
- Ejemplos ejecutables cuando exista código.
- No duplicar una regla normativa: enlazar su fuente.

## Go

El baseline de toolchain está aceptado en
[ADR-0012](../adr/0012-pin-go-toolchain-and-isolate-tools.md):

- Go 1.26.5 como toolchain exacto inicial;
- `goimports` como formateador y organizador de imports;
- golangci-lint v2 con lista explícita de reglas;
- `govulncheck` para vulnerabilidades conocidas;
- herramientas pineadas en un `go.tool.mod` separado;
- biblioteca estándar antes que frameworks o helpers adicionales.

Las excepciones `//nolint` nombran el linter y explican el motivo. No se desactiva
una categoría completa para resolver un caso local.

Continúan pendientes de su contexto de implementación:

- nombres y límites concretos de paquetes;
- política de errores y logging;
- estructura interna del monolito modular;
- reglas de código generado.

## TypeScript

El baseline está aceptado en
[ADR-0014](../adr/0014-use-node-pnpm-and-strict-typescript.md):

- TypeScript estricto desde el primer workspace;
- propiedades opcionales y accesos por índice tratados con precisión;
- overrides de clases marcados explícitamente;
- ESLint detecta problemas y Prettier decide el formato;
- las excepciones de lint deben ser locales y justificar el motivo;
- el código generado no se modifica manualmente.

El framework podrá ampliar esta configuración, pero no reducir silenciosamente
las garantías transversales.

## Reglas contra la sobreingeniería

Se exige justificación antes de introducir:

- una interfaz con una sola implementación sin límite útil;
- paquetes “common”, “utils” o “base” sin cohesión;
- repositorios genéricos;
- event buses, CQRS o microservicios;
- frameworks que sustituyan capacidades simples de la biblioteca estándar;
- abstracciones multi-cloud anticipadas.

La pregunta no es si el patrón es válido, sino qué problema actual resuelve y qué
coste permanente añade.
