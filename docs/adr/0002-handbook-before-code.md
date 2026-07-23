# ADR-0002: Construir el handbook antes que el código

- **Estado:** Aceptado
- **Fecha:** 2026-07-23
- **Decisor:** Usuario, mediante solicitud explícita
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

Empezar por la implementación convertiría decisiones implícitas en restricciones
antes de acordar propósito, proceso, arquitectura y criterios de aprendizaje.

## Alternativas

### Implementar primero

Produce feedback ejecutable rápido, pero fuerza elecciones de dominio, estructura,
tooling y operación aún no analizadas.

### Documentar todo el sistema en detalle

Reduce incertidumbre aparente, pero crea diseño especulativo y sobreingeniería.

### Crear un handbook mínimo pero operativo

Define autoridad, decisiones, gates y plantillas sin inventar requisitos. Permite
que la documentación evolucione con evidencia.

## Decisión del usuario

Construir el Backend Engineering Handbook antes que el código y tratar
`PROJECT_MANIFESTO` como fuente de verdad.

## Consecuencias

- Fase 0 no contiene código de aplicación.
- Los documentos explicitan incógnitas en lugar de rellenarlas.
- El primer paso posterior es decidir alcance y caso de uso, no elegir librerías.
- El handbook se versionará y cambiará junto al proyecto.

## Validación

La Fase 0 termina cuando existe navegación, trazabilidad, plantillas, decisiones
iniciales y validación de enlaces.

## Disparadores de revisión

- La documentación no ayuda a decidir o operar.
- Se convierte en diseño especulativo no respaldado por requisitos.
- El coste de mantenimiento supera el aprendizaje producido.
