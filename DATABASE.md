# Datos y persistencia

> Dirección objetivo: PostgreSQL. Diseño y herramientas pendientes de decisión.

## Principios

- El modelo de datos deriva del dominio y sus invariantes.
- La base de datos es un detalle externo respecto a la lógica de negocio, pero su
  semántica transaccional no debe ocultarse.
- Integridad y restricciones viven lo más cerca posible de los datos cuando
  PostgreSQL pueda garantizarlas.
- Toda evolución de esquema será reproducible, revisable y reversible o tendrá un
  plan explícito de recuperación.
- Backups solo cuentan cuando se prueba una restauración.

## Decisiones necesarias antes del primer esquema

- límites de consistencia y transacciones;
- identificadores y estrategia temporal;
- herramienta y política de migraciones;
- consultas: SQL directo, generador u otra alternativa;
- concurrencia, idempotencia y bloqueos;
- datos de desarrollo y pruebas;
- backup, restore, retención y datos sensibles.

## Cache

Redis aparece como candidato y Valkey debe evaluarse cuando exista un problema
medido que justifique cache. “Mejor rendimiento” sin presupuesto de latencia,
carga o perfil de consultas no es un requisito suficiente.

Antes de añadir cache se documentará:

1. cuello de botella observado;
2. datos cacheados y propietario de la verdad;
3. estrategia de invalidación;
4. comportamiento ante fallo;
5. consistencia tolerada;
6. coste operativo;
7. Redis frente a Valkey y opción sin cache.

## Evidencia operativa futura

El handbook deberá incluir migración, rollback/forward-fix, backup, restauración,
análisis de consultas y respuesta ante saturación de conexiones.
