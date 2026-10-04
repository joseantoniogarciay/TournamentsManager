# Runbook: PostgreSQL local con Docker Compose

> **Estado:** procedimiento implementado y validado con Docker Compose local.
>
> **Última prueba:** 2026-10-03 — Docker Desktop, migraciones hasta 17 sin
> reset, API con Air, correo, métricas y observabilidad local.

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

## Esquema inicial y migraciones

### Retirada local del 2026-10-03

Por petición del usuario, después de la validación se detuvieron Expo y Compose
y se eliminaron los contenedores, imágenes y volúmenes de
`tournaments-manager-local`, la caché de compilación creada en esta sesión,
`node_modules`, `.pnpm-store`, `.expo`, los proyectos nativos generados y los
temporales de pruebas. El código y los contratos locales se conservan. El perfil
público `tournaments-manager-dev` no se modificó.

Antes de retirar PostgreSQL se exportó una copia comprimida sin propietario ni
ACL y se verificó con `pg_restore --list`. Está en
`.config/local-backups/postgres-2026-10-03.dump`, ignorada por Git, con permisos
600. Contiene datos privados: no se publica ni se añade al repositorio.

Para recuperar esos datos en una nueva instancia local:

```bash
make db-up
docker compose --env-file infra/local/.env -f infra/local/compose.dev.yaml exec -T postgres sh -ec 'pg_restore -U "$POSTGRES_USER" -d "$POSTGRES_DB" --no-owner --no-acl --exit-on-error' < .config/local-backups/postgres-2026-10-03.dump
```

La restauración necesita una base vacía y ya incluye el esquema y el historial
Goose hasta la migración 17: no se aplica encima `db-schema-apply`. Para volver
a usar el cliente hay que reinstalar con `pnpm install`; Compose reconstruye o
descarga sus imágenes cuando se vuelva a arrancar.

En una base vacía se aplica primero el esquema base de
`apps/backend/db/schema/initial_schema.sql` y después las migraciones
inmutables de `apps/backend/db/migrations/`. El esquema sigue siendo la entrada
de `sqlc`; las migraciones no se reescriben tras aplicarse.

```bash
make db-schema-apply  # solo una base vacía
make dev-migrate
```

`dev-migrate` ejecuta Goose en un contenedor efímero del perfil local y guarda
su historial en PostgreSQL. Solo omite `SET ROLE` y los `GRANT` del perfil
público en una copia temporal: conserva el DDL y los archivos versionados. La
API local usa su administrador, mientras el perfil público mantiene separados
propietario, migrador y runtime. No aplica migraciones al arrancar la API.

Antes de actualizar una base existente, toma una copia con `pg_dump` y consulta
su historial Goose. Una base antigua modificada manualmente no permite asumir
que una migración está aplicada solo porque existe una tabla: hay que comparar
columnas, defaults, restricciones e índices antes de registrar la equivalencia.
El 2026-10-03 se verificaron y registraron así las migraciones 6 y 10 del volumen
local previo; las demás se ejecutaron normalmente, hasta la 17. No se reseteó
ni se sustituyó ningún volumen. Ese ajuste de historial es puntual y no forma
parte del script de migración.

En el entorno público de desarrollo, tras el bootstrap o ante una migración
pendiente, ejecuta explícitamente:

```bash
make dev-public-migrate
```

El comando crea un contenedor efímero con Goose y la credencial de migración;
la API no recibe esa credencial. Una contraseña de migración usada en la URL de
PostgreSQL debe usar caracteres seguros para URL o estar codificada.

El comando público usa exclusivamente `tournaments-manager-dev`; el local usa
`tournaments-manager-local`. Las migraciones preservan los datos. Un reset exige
aceptar explícitamente su pérdida y no es el procedimiento normal de actualización.

## Telemetría y espacio en disco

Los dos perfiles de desarrollo conservan señales técnicas de un día. La
rotación de Docker limita la copia de consola a dos archivos de 5 MB, con
compresión; su límite es por tamaño. La evidencia legal y los backups tienen
sus propios plazos y no forman parte de esta limpieza.

Con el entorno desmontado, `make dev-observability-clean` puede retirar solo
volúmenes técnicos sin contenedores y sin archivos de las últimas 24 horas;
`make dev-public-observability-clean` hace lo mismo para el perfil público.
Ambos conservan datos recientes, PostgreSQL y su backup. Consulta los detalles,
la purga asíncrona y los límites en [LOG_RETENTION.md](../operations/LOG_RETENTION.md).

## Parada, inspección y recuperación

```bash
make dev-down
make db-status
```

`dev-down` conserva el volumen y, por tanto, los datos. Para eliminar por completo
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

## Sesiones de pruebas bajo petición — ADR-0146

`make dev-up` levanta solo API, PostgreSQL y Mailpit. Para diagnóstico solicitado,
con la API ya activa, ejecuta `make dev-observability-up`; al terminar usa
`make dev-observability-down`. Cierra todas las pruebas con `make dev-down`.
No se reinician servicios con Docker Desktop y no se elimina ningún volumen al
apagar. Para público usa los equivalentes `dev-public-*`. Las tareas programadas
dev se suspenden mientras ese carril está apagado.
