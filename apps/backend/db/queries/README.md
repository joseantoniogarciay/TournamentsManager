# Consultas sqlc

Cada archivo `.sql` contendrá consultas de una capacidad concreta, con la
anotación `-- name: Nombre :one|many|exec|execrows`. No se añaden consultas de
relleno: la primera se incorporará junto al caso de uso que la necesite.

`sqlc.yaml` analiza las migraciones como esquema y genera código Go tipado en
`internal/adapters/postgres/sqlc/`. Esa salida se versiona, se regenera con
`make sqlc-generate` y no se modifica manualmente.
