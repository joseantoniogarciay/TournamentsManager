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

## Go — baseline para decidir en Fase 2

La convención inicial será seguir el toolchain y la biblioteca estándar de Go
antes de añadir herramientas o frameworks. Antes de codificar se documentarán:

- versión y política de actualización;
- formato, lint y análisis estático;
- nombres y límites de paquetes;
- errores y logging;
- dependencias y generación;
- estructura mínima del módulo.

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
