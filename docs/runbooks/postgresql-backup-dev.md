# Runbook: backup y restauración PostgreSQL de `dev`

> Estado: revalidado el 2026-10-03 con copia completa, incremental mediante
> LaunchAgent, WAL y restauración aislada con recuperación terminada.

## Alcance

Este runbook protege el clúster `tournaments-manager-dev`, no el entorno
`local`, la evidencia legal separada ni el runtime `prod` de K3s. Usa pgBackRest
2.59.1, copia física, archivado WAL y recuperación a un instante (PITR).

La copia base contiene el clúster completo. Los incrementales contienen cambios
físicos desde la copia anterior y el WAL permite reproducir cambios posteriores.
Una exportación `pg_dump` no puede sustituir este procedimiento porque no sirve
para reproducir WAL.

## Preparación única

1. Usa una carpeta privada de iCloud Drive distinta de `legal-audit-backups` y
   separada del repositorio de `prod`:

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
   Desktop puede escribir en ella. En Finder, selecciona únicamente `dev` y
   activa **Conservar en dispositivo** (Keep Downloaded). Descargar ahora o
   leer un archivo lo hace disponible en ese momento, pero no asegura su
   permanencia local. Un repositorio montado directamente en Docker necesita
   sus archivos disponibles, incluidos metadatos, manifiestos y WAL.
   [Apple documenta esta opción](https://support.apple.com/en-gb/guide/mac-help/mchl1a02d711/mac).
   Este ajuste conserva la arquitectura de ADR-0108 y ocupa el tamaño del
   repositorio; no cambia permisos, clave ni proveedor.

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

El comando requiere Python 3 y Docker. Lee la configuración del contenedor
activo de dev sin imprimir secretos, monta exclusivamente su repositorio en
solo lectura y crea un volumen temporal con nombre único. Restaura la etiqueta
indicada hasta alcanzar consistencia y termina la recuperación mediante
`--target-action=promote` exclusivamente en ese destino aislado. PostgreSQL
arranca sin red; se comprueba `pg_is_in_recovery() = false` y se comparan esquema,
cuentas, torneos, partidos, historial y aceptaciones legales con los agregados
de dev. El contenedor y el volumen temporal se eliminan al terminar, también
ante fallo. No monta `postgres-data` ni usa `postgres-restore-data`.

La comparación está pensada para una copia reciente y una ventana sin cambios
de negocio. Una copia histórica o escrituras concurrentes pueden dar diferencias
sin que el backup esté corrupto; hay que investigar antes de dar por superada la
verificación. Los conteos no equivalen a una comparación de todos los datos.
Esta prueba demuestra recuperación consistente de la copia elegida; no acredita
PITR a cualquier instante de un periodo anterior con fallos de archivado.

Para PITR, detén el destino aislado, restaura con `pgbackrest --type=time
--target='<instante UTC>' restore` y arráncalo allí. Nunca restaures sobre el
volumen de `dev` sin un incidente declarado y un plan de recuperación.

## Límites

`iCloud Drive` es una primera ubicación fuera del volumen PostgreSQL, no una
segunda ubicación independiente del Mac/cuenta ni almacenamiento inmutable. La
producción requiere revisar ese límite, la custodia/rotación de claves, RPO/RTO
y una prueba periódica registrada.

## Incidente de lectura iCloud — 2026-10-03

Docker devolvía `Input/output error` al leer `backup.info` y su copia. macOS
pudo leer ambos archivos; tras esa lectura Docker avanzó hasta `archive.info`.
Al leer también los metadatos de archivo desde macOS, `pgbackrest info` volvió
a `status: ok`, con cifrado AES-256-CBC. Finder mostraba `dev` como **Sin
descargar**. La evidencia apunta a disponibilidad local bajo demanda, no a una
clave incorrecta; no fue necesario recrear la stanza ni cambiar permisos.

- `pgbackrest check` aprobado y archivador nuevamente registrando éxitos;
  conserva 108 fallos históricos, que no se borraron.
- Completa `20261003-154848F` creada sin expiración automática durante el
  diagnóstico para conservar el histórico hasta verificar la recuperación.
- Incremental `20261003-154848F_20261003-155132I` creada mediante el LaunchAgent
  instalado: código de salida 0. Ambos calendarios están cargados.
- `make dev-public-backup-restore-verify` aprobado con esa incremental:
  `tournaments_manager_dev|f|19|4|10|57|17|3` (base, recuperación, esquema,
  cuentas, torneos, partidos, historial y aceptaciones legales).
- PostgreSQL y API siguieron activos; no hubo reinicio del servicio ni
  restauración sobre datos activos. La prueba no modifica producción.

La recuperación y el calendario se verificaron. Tras autorización explícita
posterior del usuario («Ok»), se activó **Conservar en dispositivo** únicamente
para `postgresql-backups/dev`. Finder confirmó la opción activa (`cmdUnpin`) e
completó la descarga: `dev`, `archive` y `backup` muestran «Se conservará en
el dispositivo», sin progreso de descarga pendiente. Después del ajuste, el repositorio
seguía en `status: ok`, con AES-256-CBC, y `pgbackrest check` volvió a pasar.
La conservación local evita depender de la descarga bajo demanda; no sustituye
la comprobación de backups y restauración ni garantiza por sí sola que iCloud
haya sincronizado la copia remota.

Retrospectiva: poder listar metadatos no demuestra que los bytes sean legibles
desde Docker, y un backup correcto necesita una restauración probada. El comando
anterior restauraba `/restore` en un contenedor sin el volumen que arrancaba el
servicio de verificación; ahora restauración y consulta comparten un único
volumen temporal. Una base que acepta conexiones todavía puede estar en pausa
de recuperación: se exige además comprobar que la haya terminado.
