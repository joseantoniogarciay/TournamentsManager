# Desarrollo

> Estado: política inicial; herramientas concretas pendientes de Fase 1.

## Principios

- Git conserva la trazabilidad entre decisión, cambio, prueba y documentación.
- El entorno local debe parecerse a producción en comportamiento, no replicar todo
  su coste ni complejidad.
- Cada automatización debe poder explicarse y tener una ruta de diagnóstico.
- Los comandos frecuentes tendrán una única entrada documentada cuando exista
  código.
- Configuración, secretos y datos de ejemplo se tratarán de forma explícita.
- La documentación cambia en el mismo conjunto de cambios que el comportamiento.
- OpenAPI es la fuente editable del contrato HTTP; el cliente TypeScript generado
  no se modifica manualmente.
- El equipo escribe las consultas SQL; el código de acceso generado por `sqlc`
  no se modifica manualmente.
- Las migraciones `goose` se ejecutan explícitamente y no al arrancar la API.
- Toda generación debe ser reproducible mediante un comando versionado y producir
  un diff limpio cuando las entradas no cambian.

## Flujo de trabajo

1. Formular el problema y comprobar si requiere decisión.
2. Actualizar o crear ADR si supera el umbral.
3. Definir criterios de aceptación y riesgos.
4. Hacer el cambio mínimo que produzca aprendizaje o valor.
5. Ejecutar comprobaciones automáticas y manuales relevantes.
6. Actualizar handbook, changelog y troubleshooting.
7. Registrar el aprendizaje cuando cambie el modelo mental.

## Definition of Done

Un cambio está terminado cuando:

- cumple criterios de aceptación;
- no contradice un ADR aceptado;
- incluye pruebas proporcionales al riesgo;
- no introduce una abstracción sin necesidad demostrable;
- permite observar y diagnosticar su comportamiento;
- actualiza la documentación afectada;
- no contiene secretos ni dependencias no justificadas.
- conserva alineados contrato OpenAPI, implementación Go y cliente TypeScript.

## Entorno local pendiente

Fase 1 decidirá y documentará:

- versiones de Go, PostgreSQL y herramientas;
- Docker Compose y ciclo de vida de servicios;
- variables de entorno y secretos locales;
- comandos de `sqlc`, migraciones `goose` y datos semilla;
- health checks;
- comandos de lint, test, build y cleanup;
- soporte de plataforma.

Hasta entonces no deben inventarse comandos ni prerequisitos.
