# Datos y persistencia

> Estado: PostgreSQL 18.4, pgx, sqlc y Goose activos; esquema incremental,
> identidades de mínimo privilegio y recuperación verificadas en `dev` y `prod`.

## Decisión vigente

[ADR-0011](../adr/0011-use-postgresql-pgx-sqlc-and-goose.md) establece la
dirección de persistencia. ADR-0072 aplazó Goose hasta existir datos que
conservar y ADR-0107 lo activó para las migraciones vigentes:

- PostgreSQL como sistema de registro relacional principal;
- `pgx` nativo para conexiones, pool y transacciones;
- SQL escrito por el equipo y código Go tipado generado mediante `sqlc`;
- migraciones SQL incrementales, inmutables y versionadas mediante `goose`.

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

Esquema inicial ── psql ──> PostgreSQL ── goose ──> versión vigente
```

`sqlc` es un generador, no un ORM ni un driver. Analiza el esquema y las
consultas y produce funciones, parámetros, resultados y escaneo tipados. `pgx`
es el driver que comunica Go con PostgreSQL. Goose evoluciona el esquema fuera
del arranque normal y registra su versión aplicada.

El [mapa entidad-relación](../diagrams/database-erd.md) ofrece una vista visual
de las tablas y claves foráneas vigentes. Es una ayuda de navegación; el esquema
SQL y sus migraciones continúan siendo la fuente de verdad ejecutable.

## PostgreSQL local

[ADR-0076](../adr/0076-run-the-local-api-in-compose-with-air.md) implementa
PostgreSQL 18.4 junto a API y Mailpit mediante Docker Compose, con volumen
nombrado, salud basada en `pg_isready` y puerto expuesto exclusivamente en
loopback. Expo y las migraciones se ejecutan desde el host. El procedimiento
operativo está en el [runbook local](../runbooks/local-postgresql.md).

## Identidades del entorno público de desarrollo

[ADR-0097](../adr/0097-separate-postgresql-runtime-and-migration-identities.md)
aplica mínimo privilegio a `tournaments-manager-dev`: el propietario del
esquema no puede iniciar sesión, la migración usa una identidad distinta y la
API usa una tercera con solo DML sobre las tablas y secuencias necesarias. El
bootstrap completo sobre una base vacía es explícito
(`make dev-public-bootstrap`), no forma parte del arranque ni del despliegue
ordinario de la API. Cada cambio futuro de esquema debe conservar los `GRANT`
de runtime y verificarlos antes de publicar.

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

- [ADR-0107](../adr/0107-activate-immutable-schema-migrations.md) activa Goose:
  `initial_schema.sql` se aplica solo a una base vacía y las migraciones SQL de
  `apps/backend/db/migrations/` se aplican después, en orden.
- Una migración aplicada en un entorno compartido es inmutable.
- Goose se ejecuta como paso explícito de despliegue, nunca como efecto
  secundario de iniciar la API.
- Toda migración se prueba desde una base vacía y sobre la versión anterior
  relevante.
- La presencia de una sección `Down` no garantiza un rollback seguro cuando hay
  pérdida o transformación de datos.
- Los cambios incompatibles requerirán una estrategia documentada de rollback,
  forward-fix o expand/contract.

## Límite con el dominio

La frontera usa tres clases de tipos con responsabilidades distintas:

- los paquetes de dominio y caso de uso definen sus propios `Input`, entidades y
  resultados para expresar reglas sin depender de PostgreSQL;
- `sqlc` genera `...Params` y `...Row` específicos de cada consulta; el adaptador
  PostgreSQL los usa para ejecutar SQL y realizar el escaneo tipado;
- los structs que reflejan una tabla completa solo se generan si alguna consulta
  los devuelve. `omit_unused_structs` evita mantener espejos de tablas que no
  participan en el código.

Los tipos generados por `sqlc` representan persistencia, no entidades de negocio.
El adaptador realiza mapeos explícitos entre ambos lados cuando el dominio
necesite tipos o invariantes propios. Que un `...Params` o `...Row` se utilice no
autoriza a propagarlo fuera del adaptador.

Las transacciones se definen desde el caso de uso. No se añade una abstracción
genérica de unit of work o repository antes de que proteja un límite real.

`product_suggestions` conserva el texto privado junto a su cuenta y fecha. La
base valida el cuerpo recortado entre 8 y 1.000 caracteres y elimina los registros
con la cuenta mediante `ON DELETE CASCADE`. Estados, respuestas y visibilidad no
forman parte de este esquema inicial; se añadirán solo con un caso de uso aceptado.

## Criterios de evolución

No quedan decisiones de persistencia bloqueantes para v1. Cada nuevo caso de
uso decide de forma local sus límites transaccionales, concurrencia,
idempotencia y bloqueos. Pool, timeouts y errores se configuran por entorno y se
revisan solo con evidencia operativa.

ADR-0108 protege `dev` con pgBackRest, WAL archivado, cifrado y restauración
aislada. `prod` usa su propio repositorio pgBackRest, réplica cifrada iniciada
desde el Mac y restauración aislada demostrada conforme a ADR-0114 y ADR-0115.
Los procedimientos viven en los runbooks de
[backup de dev](../runbooks/postgresql-backup-dev.md) y
[PostgreSQL de K3s](../runbooks/k3s-postgresql.md).

## Diseño del primer esquema

El [modelo inicial de datos](INITIAL_DATA_MODEL.md), aceptado en ADR-0045,
define las entidades, restricciones y transacciones del primer incremento.
[ADR-0047](../adr/0047-organize-initial-postgresql-schema-and-sqlc.md), ajustado
por ADR-0053, lo materializó en el esquema inicial. Las migraciones son ahora la
entrada de esquema para sqlc y para construir la base efímera; las consultas
viven en `db/queries` y la salida generada bajo el adaptador PostgreSQL se
versiona y no se edita manualmente.

## Cache

No se adopta caché en v1. Redis y Valkey son alternativas que solo se evaluarán
si aparece un problema medido que justifique ese coste. “Mejor rendimiento” sin
presupuesto de latencia, carga o perfil de consultas no es un requisito
suficiente.

Antes de añadir cache se documentará:

1. cuello de botella observado;
2. datos cacheados y propietario de la verdad;
3. estrategia de invalidación;
4. comportamiento ante fallo;
5. consistencia tolerada;
6. coste operativo;
7. Redis frente a Valkey y opción sin cache.

## Evidencia operativa actual

`make verify` comprueba la generación determinista y las migraciones se aplican
como paso explícito de despliegue. Los runbooks cubren backup y restauración
aislada en `dev` y `prod`. Un cambio destructivo sigue exigiendo documentar
rollback, forward-fix o expand/contract; el análisis de consultas y la
saturación de conexiones se investigan cuando exista evidencia, no como trabajo
abierto genérico.
