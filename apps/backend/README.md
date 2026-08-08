# Backend Go

Este directorio contiene el backend desplegable del producto. La API HTTP es un
adaptador de entrada: no contiene por sí sola toda la lógica del backend. Las
reglas de negocio no dependen de HTTP, PostgreSQL ni `sqlc`.

La dirección arquitectónica y las decisiones que la sostienen están en la
[guía de arquitectura](../../docs/engineering/ARCHITECTURE.md),
[API](../../docs/engineering/API.md) y
[ADR-0011](../../docs/adr/0011-use-postgresql-pgx-sqlc-and-goose.md).

## Mapa rápido

```text
cmd/api/                         composición y arranque de la aplicación
internal/adapters/http/          rutas HTTP, middleware y handlers
internal/<capacidad>/            casos de uso, modelos y puertos de negocio
internal/adapters/postgres/      implementación PostgreSQL de los puertos
db/queries/                      SQL escrito por el equipo
internal/adapters/postgres/sqlc/ código Go generado; no se edita a mano
db/schema/                       esquema PostgreSQL inicial que sqlc usa como referencia
```

`cmd/api/main.go` construye las dependencias: abre el pool PostgreSQL, crea los
repositorios y los entrega a los servicios; después entrega esos servicios al
adaptador HTTP. Esa composición es el punto donde se unen las capas.

## Recorrido real: disponibilidad de username

La operación pública actual es `GET /v1/usernames/{username}/availability`.
Su contrato HTTP se define en OpenAPI; el cliente TypeScript se genera desde ese
contrato. El flujo de una petición válida es:

```text
Cliente
  → GET /v1/usernames/{username}/availability
  → internal/adapters/http/server.go: NewHandler registra la ruta
  → usernameAvailability valida el formato y aplica el límite de tasa
  → registration.Service.UsernameAvailable
  → registration.Repository.IsUsernameAvailable (puerto de negocio)
  → postgres.RegistrationRepository.IsUsernameAvailable
  → r.queries.IsUsernameAvailable
  → SQL de db/queries/registrations.sql
  → PostgreSQL
  → {"available": true|false}
```

El handler traduce HTTP a la entrada del caso de uso y convierte el resultado en
JSON. `registration.Service` coordina la capacidad sin importar paquetes HTTP,
`pgx` ni `sqlc`. La implementación PostgreSQL conoce esos detalles y satisface
la interfaz `registration.Repository`.

### Qué significa `s` y qué significa `r`

En Go, el identificador situado entre `func` y el nombre del método es el
**receptor**: la instancia sobre la que se invoca el método.

```go
func (s Service) UsernameAvailable(ctx context.Context, username string) (bool, error) {
	return s.repository.IsUsernameAvailable(ctx, username)
}

func (r RegistrationRepository) IsUsernameAvailable(ctx context.Context, username string) (bool, error) {
	return r.queries.IsUsernameAvailable(ctx, username)
}
```

- `s` es un `Service`; su campo `repository` tiene el puerto que necesita el
  caso de uso.
- `r` es un `RegistrationRepository`; su campo `queries` es un
  `*sqlc.Queries`.

El repositorio se construye con `sqlc.New(pool)`, de modo que `r.queries` usa el
pool PostgreSQL creado durante el arranque. Así el caso de uso no depende de una
base de datos concreta y el detalle de ejecución queda en el adaptador.

## De dónde sale `r.queries.IsUsernameAvailable`

El equipo declara la consulta en
[`db/queries/registrations.sql`](db/queries/registrations.sql):

```sql
-- name: IsUsernameAvailable :one
SELECT NOT EXISTS (
    SELECT 1 FROM accounts WHERE username = $1
) AS available;
```

`-- name: ...` es un comentario SQL al que `sqlc` da significado adicional:
`IsUsernameAvailable` será el nombre del método Go y `:one` indica que la
consulta devuelve una fila. Al ejecutar `make sqlc-generate`, `sqlc` analiza
las consultas contra el esquema de `db/schema/` y genera, entre otros,
`internal/adapters/postgres/sqlc/registrations.sql.go`:

```go
func (q *Queries) IsUsernameAvailable(ctx context.Context, username string) (bool, error)
```

Los archivos bajo `internal/adapters/postgres/sqlc/` están versionados para que
el cambio sea revisable, pero son salida generada: se modifican cambiando el SQL
o el esquema y regenerando, nunca a mano.

La respuesta de disponibilidad es solo informativa. Dos personas pueden verla
como disponible a la vez; la restricción única de PostgreSQL y la operación de
alta conservan la autoridad final sobre la unicidad.

## Configuración y arranque local

`internal/config/config.go` lee las variables reales con `os.Getenv`.
`internal/config/config_test.go` no inicia la API ni carga un `.env`: prueba esa
validación inyectando valores controlados.

Para el arranque local, `make dev-init` crea los contratos locales y `make dev-up`
inicia la API en Docker Compose junto a PostgreSQL y Mailpit. Air recompila la
API al guardar. `infra/local/api.docker.env` usa los nombres de servicio Docker,
mientras `apps/backend/.env` queda disponible para un arranque puntual en host.
Consulta
[Desarrollo](../../docs/engineering/DEVELOPMENT.md) y el
[runbook PostgreSQL local](../../docs/runbooks/local-postgresql.md) para los
comandos y el contrato completo de variables.

## Pruebas

Go ejecuta todos los ficheros que terminan en `_test.go`; el prefijo agrupa por
responsabilidad, no es una ubicación global obligatoria. Por ejemplo:

- `internal/adapters/http/server_test.go`: comportamiento HTTP;
- `internal/registration/service_test.go`: reglas del caso de uso;
- adaptadores PostgreSQL: consultas, restricciones y sus mapeos.

La estrategia por capas y riesgo está documentada en
[TESTING.md](../../docs/engineering/TESTING.md).
