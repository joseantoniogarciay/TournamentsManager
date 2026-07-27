# Datos y persistencia

> Estado: PostgreSQL, pgx, sqlc y goose aceptados; PostgreSQL 18.4 en Compose
> local. El esquema y las políticas operativas de producción continúan pendientes.

## Decisión vigente

[ADR-0011](../adr/0011-use-postgresql-pgx-sqlc-and-goose.md) establece:

- PostgreSQL como sistema de registro relacional principal;
- `pgx` nativo para conexiones, pool y transacciones;
- SQL escrito por el equipo y código Go tipado generado mediante `sqlc`;
- migraciones SQL incrementales y versionadas mediante `goose`.

```text
Casos de uso y dominio
          │
          ▼
Puerto necesario de persistencia
          │
          ▼
Adaptador y mapeos
          │
          ├── código generado por sqlc
          ├── pgx / pgxpool
          └── PostgreSQL

Esquema versionado ── goose ──> PostgreSQL
```

`sqlc` es un generador, no un ORM ni un driver. Analiza el esquema y las
consultas y produce funciones, parámetros, resultados y escaneo tipados. `pgx`
es el driver que comunica Go con PostgreSQL. `goose` evoluciona el esquema fuera
del arranque normal de la API.

## PostgreSQL local

[ADR-0018](../adr/0018-use-compose-for-local-service-dependencies.md) implementa
PostgreSQL 18.4 mediante Docker Compose con volumen nombrado, salud basada en
`pg_isready` y puerto expuesto exclusivamente en loopback. API Go, Expo y las
migraciones se ejecutan desde el host. El procedimiento operativo está en el
[runbook local](../runbooks/local-postgresql.md).

## Principios

- El modelo de datos deriva del dominio y sus invariantes.
- La base de datos es un detalle externo respecto a la lógica de negocio, pero su
  semántica transaccional no debe ocultarse.
- Integridad y restricciones viven lo más cerca posible de los datos cuando
  PostgreSQL pueda garantizarlas.
- Toda evolución de esquema será reproducible, revisable y reversible o tendrá un
  plan explícito de recuperación.
- Backups solo cuentan cuando se prueba una restauración.
- El SQL generado por una herramienta no sustituye la revisión del equipo; en
  esta decisión el equipo escribe el SQL y `sqlc` genera Go.
- Una única base compartida no elimina la propiedad de tablas por módulo.
- El código generado no entra en el dominio ni se modifica manualmente.
- No se crean repositorios genéricos ni una interfaz por tabla.

## Política de migraciones

- Mientras solo exista PostgreSQL local y los datos sean descartables,
  [ADR-0053](../adr/0053-keep-a-single-resettable-local-initial-schema.md)
  establece una excepción temporal: `00001_initial_schema.sql` es el único
  esquema inicial y se reescribe tras un reset explícito. No se añaden pasos
  incrementales durante esta etapa.
- Las migraciones iniciales se escriben en SQL.
- Una migración aplicada en un entorno compartido es inmutable.
- `goose` se ejecuta como paso explícito de desarrollo o despliegue, no como
  efecto secundario de iniciar la API.
- Toda migración se prueba desde una base vacía y sobre la versión anterior
  relevante.
- La presencia de una sección `Down` no garantiza un rollback seguro cuando hay
  pérdida o transformación de datos.
- Los cambios incompatibles requerirán una estrategia documentada de rollback,
  forward-fix o expand/contract.

## Límite con el dominio

Los tipos de filas generados por `sqlc` representan persistencia, no entidades de
negocio. El adaptador realiza mapeos explícitos cuando el dominio necesite tipos
o invariantes propios.

Las transacciones se definen desde el caso de uso. No se añade una abstracción
genérica de unit of work o repository antes de que proteja un límite real.

## Decisiones pendientes tras el esquema inicial

- límites de consistencia y transacciones de cada caso de uso;
- concurrencia, idempotencia y bloqueos;
- configuración del pool, timeouts y errores;
- datos de desarrollo y pruebas;
- backup, restore, retención y datos sensibles.

## Diseño del primer esquema

El [modelo inicial de datos](INITIAL_DATA_MODEL.md), aceptado en ADR-0045,
define las entidades, restricciones y transacciones del primer incremento.
[ADR-0047](../adr/0047-organize-initial-postgresql-schema-and-sqlc.md), ajustado
por ADR-0053, lo materializa en el único esquema inicial, sin crear aún consultas de negocio ni
adaptadores. `db/migrations` es la fuente de esquema para Goose y sqlc; las
consultas futuras viven en `db/queries` y la salida generada bajo el adaptador
PostgreSQL se versiona y no se edita manualmente.

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

El handbook deberá incluir generación determinista, migración, rollback o
forward-fix, backup, restauración, análisis de consultas y respuesta ante
saturación de conexiones.
