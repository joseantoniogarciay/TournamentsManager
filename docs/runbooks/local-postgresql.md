# Runbook: PostgreSQL local con Docker Compose

> **Estado:** procedimiento implementado y validado con Docker Compose local.
>
> **Última prueba:** 2026-07-26 — arranque, `healthcheck`, reinicio con volumen
> persistente, reset confirmado y migraciones vacías.

## Alcance

Este runbook opera exclusivamente PostgreSQL local. API Go y cliente Expo se
ejecutan desde el host; no hay contenedores de frontend ni de API. Véase
[ADR-0018](../adr/0018-use-compose-for-local-service-dependencies.md).

## Prerrequisitos

- Docker Desktop o Docker Engine con Docker Compose v2 y soporte de
  `docker compose up --wait`.
- Go 1.26.5 para ejecutar migraciones mediante la herramienta versionada.
- Una copia local de cada contrato:

```bash
cp infra/local/.env.example infra/local/.env
cp apps/backend/.env.example apps/backend/.env
```

Edita ambos archivos para que usuario, contraseña y base coincidan. No subas los
archivos `.env` a Git. La contraseña de ejemplo contiene caracteres seguros para
una URL; si se cambia por una contraseña con caracteres reservados, actualiza
`DATABASE_URL` con la codificación URL correspondiente.

## Arranque y verificación

```bash
make db-up
make db-status
```

`db-up` espera el `healthcheck`, que ejecuta `pg_isready`; si termina con éxito,
la base acepta conexiones. Para observar el arranque:

```bash
make db-logs
```

## Migraciones

Las migraciones viven en `apps/backend/db/migrations` y se aplican explícitamente:

```bash
make db-migrate
```

El comando usa `DATABASE_URL` de `apps/backend/.env` y ejecuta `goose`; no inicia
la API ni crea datos funcionales. Aún no hay migraciones porque esquema y primer
vertical slice no están implementados. En ese estado informa que se omite y
termina correctamente; al añadir la primera migración SQL, ejecutará `goose`.

## Parada, inspección y recuperación

```bash
make db-down
make db-status
```

`db-down` conserva el volumen y, por tanto, los datos. Para eliminar por completo
los datos locales, ejecuta:

```bash
make db-reset
```

El comando exige escribir `RESET` y elimina únicamente el volumen nombrado del
proyecto Compose. Durante la construcción inicial, ADR-0053 permite reescribir
el único esquema `00001_initial_schema.sql` antes de este reset. Después repite
`make db-up` y `make db-migrate`; no se conservan datos locales.

## Diagnóstico seguro

| Síntoma                     | Diagnóstico                                                                               | Mitigación                                                                                                                                                                       |
| --------------------------- | ----------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Falta `.env`                | `make db-up` informa la ruta ausente                                                      | Copiar el ejemplo correspondiente; no crear secretos en Git.                                                                                                                     |
| El puerto 5432 está ocupado | `make db-status` y revisar el proceso que lo usa                                          | Detener el proceso ajeno o cambiar ambos contratos de puerto/URL mediante un cambio documentado.                                                                                 |
| Salud no llega a `healthy`  | `make db-logs`                                                                            | Verificar usuario, base y contraseña solo en los `.env` locales; si el volumen contiene una inicialización previa incompatible, usar `db-reset` tras confirmar pérdida de datos. |
| Migración no conecta        | Verificar que `make db-status` muestra PostgreSQL saludable y que `DATABASE_URL` coincide | Corregir el contrato de backend; no modificar migraciones aplicadas para ocultar el fallo.                                                                                       |

## Límites

- No expongas PostgreSQL fuera de `127.0.0.1`.
- No añadas Redis/Valkey, MinIO, observabilidad ni contenedores de aplicación
  sin una decisión posterior.
- Este procedimiento no sustituye una restauración de backup ni valida el
  empaquetado de producción.
