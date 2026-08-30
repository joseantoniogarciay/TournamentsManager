# Runbook: backup y restauración PostgreSQL de `dev`

> Estado: validado el 2026-08-23 con copia completa, incremental, WAL y
> restauración aislada arrancada sin tocar el volumen activo.

## Alcance

Este runbook protege el clúster `tournaments-manager-dev`, no el entorno
`local`, la evidencia legal separada ni la futura producción. Usa pgBackRest
2.59.1, copia física, archivado WAL y recuperación a un instante (PITR).

La copia base contiene el clúster completo. Los incrementales contienen cambios
físicos desde la copia anterior y el WAL permite reproducir cambios posteriores.
Una exportación `pg_dump` no puede sustituir este procedimiento porque no sirve
para reproducir WAL.

## Preparación única

1. Usa una carpeta privada de iCloud Drive distinta de `legal-audit-backups` y
   separada del futuro repositorio de `prod`:

   ```sh
   POSTGRES_BACKUP_DESTINATION="$HOME/Library/Mobile Documents/com~apple~CloudDocs/FastTourney/postgresql-backups/dev"
   PGBACKREST_REPO1_CIPHER_PASS=<frase-aleatoria-larga-y-exclusiva>
   ```

   La clave no se guarda en iCloud, Git, logs ni en la clave del backup legal.
   Genera una mediante `openssl rand -base64 48` y guárdala en el gestor de
   secretos elegido para el Mac. No la cambies después de crear el repositorio:
   perderla impide restaurar.

2. Si el repositorio de `dev` ya existe en el nivel antiguo
   `postgresql-backups`, no cambies la variable mientras PostgreSQL está activo.
   Detén `dev`, mueve ese directorio completo a `postgresql-backups/dev`, cambia
   la variable y arranca `dev`. Comprueba después `make dev-public-backup-status`
   antes de retirar cualquier copia anterior. El nuevo sibling
   `postgresql-backups/prod` queda reservado para la réplica de K3s y nunca
   comparte clave ni archivos con `dev`.

3. Asegúrate de que iCloud Drive ha sincronizado la carpeta y de que Docker
   Desktop puede escribir en ella.

4. Inicializa y verifica. El comando crea la stanza, fuerza y comprueba el
   archivado de WAL, y toma la primera copia completa:

   ```sh
   make dev-public-backup-init
   make dev-public-backup-status
   ```

## Calendario y retención

- Domingo, 03:45: copia completa.
- Lunes a sábado, 03:45: incremental.
- `archive_timeout=3600`: el RPO objetivo inicial es una hora; iCloud puede
  añadir retraso de sincronización.
- Se conservan dos copias completas y sus incrementales/WAL: ventana efectiva
  aproximada de 7–14 días.

Instala manualmente los dos templates de `infra/home/launchd/` como
LaunchAgents del usuario que ejecuta Docker Desktop. Sustituye
`__LOG_DIRECTORY__` por una ruta privada no sincronizada y cárgalos con
`launchctl bootstrap gui/$(id -u) <ruta-del-plist>`. No ejecutes dos copias
concurrentes.

Para ejecutar manualmente una copia:

```sh
make dev-public-backup-full
make dev-public-backup-incremental
make dev-public-backup-status
```

## Restauración aislada y prueba

Identifica una etiqueta en `make dev-public-backup-status`, por ejemplo
`20260823-034500F_20260824-034500I`, y ejecuta:

```sh
make dev-public-backup-restore-verify BACKUP=20260823-034500F_20260824-034500I
```

El comando restaura en `postgres-restore-data`, inicia un PostgreSQL sin puerto
público, verifica una consulta y elimina solo el contenedor temporal. No toca
`postgres-data`.

Para PITR, detén el destino aislado, restaura con `pgbackrest --type=time
--target='<instante UTC>' restore` y arráncalo allí. Nunca restaures sobre el
volumen de `dev` sin un incidente declarado y un plan de recuperación.

## Límites

`iCloud Drive` es una primera ubicación fuera del volumen PostgreSQL, no una
segunda ubicación independiente del Mac/cuenta ni almacenamiento inmutable. La
producción requiere revisar ese límite, la custodia/rotación de claves, RPO/RTO
y una prueba periódica registrada.
