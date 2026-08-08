# Runbook: PostgreSQL local con Docker Compose

> **Estado:** procedimiento implementado y validado con Docker Compose local.
>
> **Última prueba:** 2026-07-26 — arranque, `healthcheck`, reinicio con volumen
> persistente, reset confirmado y esquema inicial reescribible.

## Alcance

Este runbook opera el entorno local: PostgreSQL, Mailpit y API con Air en
Compose. Expo sigue en host; no hay contenedor de frontend. Véase
[ADR-0076](../adr/0076-run-the-local-api-in-compose-with-air.md).

## Prerrequisitos

- Docker Desktop o Docker Engine con Docker Compose v2 y soporte de
  `docker compose up --wait`.
- Cliente `psql` dentro del contenedor PostgreSQL para aplicar el esquema inicial.
- Una copia local de cada contrato:

```bash
cp infra/local/.env.example infra/local/.env
cp infra/local/api.docker.env.example infra/local/api.docker.env
```

Edita los contratos para que usuario, contraseña y base coincidan. No subas los
archivos `.env` a Git. La contraseña de ejemplo contiene caracteres seguros para
una URL; si se cambia por una contraseña con caracteres reservados, actualiza
ambos `DATABASE_URL` con la codificación URL correspondiente. En
`api.docker.env`, los hosts deben permanecer como `postgres` y `mailpit`.

## Arranque y verificación

```bash
make dev-up
```

`dev-up` espera a que API, PostgreSQL y Mailpit estén disponibles. Air compila la
API tras cada guardado Go. Comprueba `http://127.0.0.1:8080/healthz`; Mailpit se
abre en `http://127.0.0.1:8025`. Para observar el arranque:

```bash
make dev-logs
```

## Esquema inicial

Durante la primera versión no hay migraciones activas. El único esquema
reescribible vive en `apps/backend/db/schema/initial_schema.sql` y se aplica
explícitamente:

```bash
make db-schema-apply
```

El comando usa el PostgreSQL local de Compose; no inicia la API ni crea datos
funcionales. Tras un cambio incompatible, elimina antes el volumen con
`make db-reset`, vuelve a arrancar PostgreSQL y aplica de nuevo el esquema.

## Parada, inspección y recuperación

```bash
make dev-down
make db-status
```

los datos locales, ejecuta:
`dev-down` conserva el volumen y, por tanto, los datos. Para eliminar por completo
los datos locales, ejecuta:
los datos locales, ejecuta:

```bash
make db-reset
```

El comando exige escribir `RESET` y elimina únicamente el volumen nombrado del
proyecto Compose. Durante la construcción inicial, ADR-0053 permite reescribir
el único esquema `initial_schema.sql` antes de este reset. Después repite
`make db-up` y `make db-schema-apply`; no se conservan datos locales.

## Diagnóstico seguro

| Síntoma                     | Diagnóstico                                                                               | Mitigación                                                                                                                                                                       |
| --------------------------- | ----------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Falta `.env`                | `make dev-up` informa la ruta ausente                                                     | Copiar el ejemplo correspondiente; no crear secretos en Git.                                                                                                                     |
| El puerto 5432 está ocupado | `make db-status` y revisar el proceso que lo usa                                          | Detener el proceso ajeno o cambiar ambos contratos de puerto/URL mediante un cambio documentado.                                                                                 |
| Salud no llega a `healthy`  | `make dev-logs`                                                                           | Verificar usuario, base y contraseña solo en los `.env` locales; si el volumen contiene una inicialización previa incompatible, usar `db-reset` tras confirmar pérdida de datos. |
| Esquema no conecta          | Verificar que `make db-status` muestra PostgreSQL saludable                              | Corregir el contrato local y reaplicar el esquema tras resetear si el cambio es incompatible.                                                                                   |

## Límites

- No expongas PostgreSQL fuera de `127.0.0.1`.
- No añadas Redis/Valkey, MinIO u observabilidad sin una decisión posterior.
- `make api-image-build` valida el empaquetado mínimo, pero este procedimiento no
  sustituye una restauración de backup ni decide el despliegue real.
