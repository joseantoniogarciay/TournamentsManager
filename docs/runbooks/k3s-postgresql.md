# Runbook: PostgreSQL de `prod` en K3s

> Estado: aplicación inicial completada el 2026-08-30. La operación requiere
> una sesión SSH interactiva del operador para introducir la contraseña de
> `sudo`; los secretos no se imprimen ni se guardan en Git.

## Alcance

Este procedimiento prepara PostgreSQL 18 con pgBackRest para el namespace
`prod`. No publica puertos, no habilita ingress, no inicia migraciones ni retira
el `503` de producción.

## Modelo

- El `StatefulSet` conserva la identidad `postgresql-0`; un Pod recreado usa los
  mismos PVC, a diferencia de una réplica intercambiable de un `Deployment`.
- `data-postgresql-0` solicita `5Gi` para el clúster activo.
- `repository-postgresql-0` solicita `10Gi` para dos completas, incrementales y
  WAL. iCloud es una segunda copia asíncrona; no puede sustituir este PVC por
  `0Gi`, porque `archive-push` necesita persistir cada WAL antes de que exista
  una réplica en el Mac.
- Ambos PVC usan `local-path` y siguen dependiendo de la VM. `Retain` impide que
  Kubernetes los elimine automáticamente al borrar o escalar el StatefulSet,
  pero borrar un PVC manualmente continúa siendo destructivo.
- El Service `postgresql` solo resuelve dentro del clúster. La API usará
  `postgresql.prod.svc.cluster.local:5432`; la base no se publica en la LAN ni
  en Internet.

## Prerrequisitos

- La VM pasa `infra/k3s/scripts/verify-remote-host.sh --require-k3s`.
- El namespace `prod` existe y `local-path` es la StorageClass predeterminada.
- El operador tiene una sesión SSH abierta para introducir la contraseña de
  `sudo`; ni este runbook ni los manifests almacenan esa contraseña.
- Las contraseñas de administrador, migrador, runtime y cifrado se generan de
  forma aleatoria y exclusiva para `prod` fuera de Git.

## Artefacto PostgreSQL

La imagen incluye PostgreSQL 18.4 y pgBackRest 2.59.1. Se construye en el Mac y
se importa en el containerd de K3s con el tag:

```text
tournaments-manager-postgresql:18.4-pgbackrest-2.59.1
```

No se aplica el StatefulSet hasta confirmar en la VM que ese tag existe en
`sudo /usr/local/bin/k3s ctr images ls`. `imagePullPolicy: Never` evita que K3s
intente descargar una imagen desde un registry.

## Secret de runtime

En la VM, copia el contrato versionado a una ruta privada, edita sus valores y
crea el Secret sin imprimirlos:

```sh
cp /ruta/al/repositorio/infra/k3s/secrets/postgresql-runtime.env.example /tmp/postgresql-runtime.env
chmod 600 /tmp/postgresql-runtime.env
# Edita /tmp/postgresql-runtime.env con un editor local de la VM.
sudo /usr/local/bin/k3s kubectl create secret generic postgresql-runtime \
  --namespace prod \
  --from-env-file=/tmp/postgresql-runtime.env \
  --dry-run=client -o yaml | sudo /usr/local/bin/k3s kubectl apply -f -
```

La API recibirá solo su URL de runtime; la credencial de migración nunca entra
en el Secret de la API.

Antes de aplicar recursos persistentes, confirma que los cuatro campos del
Secret no conservan los placeholders del ejemplo. Una clave de cifrado
placeholder permite leer cualquier copia que se cree con ella; si se detecta
antes de datos reales, rota el Secret y recrea solo el PVC de repositorio, nunca
el PVC de datos por comodidad.

## Validación antes de persistir recursos

En la VM, este comando consulta al API server y no conserva objetos:

```sh
sudo /usr/local/bin/k3s kubectl apply --dry-run=server \
  -f /ruta/al/repositorio/infra/k3s/core/postgresql-config.yaml \
  -f /ruta/al/repositorio/infra/k3s/core/postgresql.yaml
```

Confirma que K3s acepta Services, StatefulSet, probes, recursos y PVC, pero no
prueba permisos de volumen, arranque de imagen ni recuperación.

## Aplicación y comprobación inicial

Solo tras revisar el dry-run y crear el Secret:

```sh
sudo /usr/local/bin/k3s kubectl apply -f /ruta/al/repositorio/infra/k3s/core/postgresql-config.yaml
sudo /usr/local/bin/k3s kubectl apply -f /ruta/al/repositorio/infra/k3s/core/postgresql.yaml
sudo /usr/local/bin/k3s kubectl -n prod rollout status statefulset/postgresql --timeout=120s
sudo /usr/local/bin/k3s kubectl -n prod get pods,services,pvc
```

Si el Pod no llega a `Ready`, inspecciona primero `kubectl -n prod describe pod
postgresql-0` y sus logs. No borres el StatefulSet ni los PVC para “probar de
nuevo”: conserva el diagnóstico y decide el rollback antes de tocar datos.

## Bootstrap de roles, esquema y Goose

ADR-0097 exige tres identidades: owner sin `LOGIN`, migrador y runtime. El Job
`postgresql-bootstrap.yaml` se usa una única vez sobre una base vacía. Sus SQL
entran mediante el ConfigMap `postgresql-bootstrap-sql`; crea roles, aplica el
esquema inicial y el grant base. Las migraciones históricas contienen los roles
de `dev` ya aplicados y son inmutables: el Job las renderiza exclusivamente en
`emptyDir` con los dos nombres `prod`, aplica esa copia temporal y registra las
versiones `0`, `2`, `3` y `4` en Goose.

Si el Job falla después del esquema inicial, no se relanza: usa
`postgresql-bootstrap-migrations.yaml`, que solo completa migraciones y el
registro de Goose. Verifica al final que runtime conecta, no puede crear tablas
y que las cuatro versiones aparecen aplicadas. Toma después una copia pgBackRest
incremental antes de desplegar la API.

## Backup y restauración aislada local

Inicializa una vez la stanza, valida el archivo WAL y crea una primera completa:

```sh
sudo /usr/local/bin/k3s kubectl -n prod exec postgresql-0 -- \
  pgbackrest --stanza=fasttourney-prod stanza-create
sudo /usr/local/bin/k3s kubectl -n prod exec postgresql-0 -- \
  pgbackrest --stanza=fasttourney-prod check
sudo /usr/local/bin/k3s kubectl -n prod exec postgresql-0 -- \
  pgbackrest --stanza=fasttourney-prod --type=full backup
```

La prueba aislada usa el manifiesto `postgresql-restore-verify.yaml`. Monta el
PVC del repositorio en solo lectura y restaura en `emptyDir`; no monta ni altera
`data-postgresql-0`, no escucha en la red y termina al comprobar que la base
restaurada no está en recuperación.

```sh
sudo /usr/local/bin/k3s kubectl -n prod delete job postgresql-restore-verify --ignore-not-found
sudo /usr/local/bin/k3s kubectl apply -f /ruta/al/repositorio/infra/k3s/core/postgresql-restore-verify.yaml
sudo /usr/local/bin/k3s kubectl -n prod wait --for=condition=complete job/postgresql-restore-verify --timeout=180s
sudo /usr/local/bin/k3s kubectl -n prod logs job/postgresql-restore-verify
```

La salida esperada acaba con `fasttourney_prod|f`: la base recuperada no está
en modo recuperación. Conserva el Job terminado como evidencia hasta registrar
el resultado; borrarlo solo elimina el Pod temporal y su `emptyDir`.

## Réplica en iCloud iniciada desde el Mac

ADR-0114 separa el privilegio remoto en tres wrappers de `root`: una completa,
un incremental y una exportación tar del repositorio. No aceptan argumentos ni
rutas del usuario SSH. Instálalos en la VM con propietario `root`, modo `0755`
y una regla `sudoers` limitada al operador; valida siempre el fichero con
`visudo -cf` antes de activarlo. El template sustituye `__K3S_OPERATOR__` por la
cuenta SSH real y no contiene secretos.

En el Mac, `infra/k3s/.env` conserva host, usuario y la ruta privada
`FastTourney/postgresql-backups/prod`; esa ruta no comparte archivos con `dev`.
`infra/k3s/scripts/backup-prod.sh full` o `incremental` inicia el backup por
SSH, recibe el repositorio en un directorio temporal hermano, comprueba sus
metadatos `backup.info` y `archive.info`, y solo entonces publica el directorio
`prod`. Un fallo previo conserva el destino publicado. La copia queda cifrada,
pero su restauración desde la réplica requiere recuperar la frase de cifrado del
gestor de contraseñas.

Las plantillas `launchd` de `infra/home/launchd/` programan la completa el
domingo a las 04:15 e incrementales de lunes a sábado a las 04:15, después del
calendario de `dev`. Antes de cargarlas, ejecuta manualmente una incremental y
confirma tanto `pgbackrest info` en la VM como la ruta publicada en iCloud.

`launchd` no puede ejecutar con seguridad artefactos desde `Desktop`, que macOS
protege mediante TCC. ADR-0115 usa un helper sandboxed que conserva bookmarks
de seguridad para el staging local y el directorio `postgresql-backups` elegido
en iCloud. El instalador conserva en Git la fuente, las plantillas y el
procedimiento, pero escribe el runtime en
`~/Library/Application Support/FastTourney/k3s`, y los logs privados no
sincronizados en `~/Library/Logs/FastTourney`:

```sh
bash infra/k3s/scripts/install-backup-launch-agents.sh
```

Después, una vez por Mac y sin pegar rutas ni secretos en la terminal, ejecuta
el helper y selecciona primero
`~/Library/Application Support/FastTourney/k3s/staging` y después
`FastTourney/postgresql-backups` de iCloud Drive:

Abre con doble clic `~/Library/Application Support/FastTourney/k3s/BackupPublisher.app`.
Si prefieres Terminal:

```sh
open ~/Library/Application\ Support/FastTourney/k3s/BackupPublisher.app --args configure
```

Los bookmarks viven dentro del contenedor privado del helper. Si se revocan,
caducan o cambia el directorio, el backup falla antes de sustituir la réplica y
se repite esta configuración. La sesión debe tener disponible `ssh-agent`.
Comprueba después que `launchd` los reconoce y ejecuta una incremental bajo el
agente antes de depender del calendario:

```sh
launchctl print "gui/$(id -u)/com.fasttourney.prod-postgresql-backup-incremental"
launchctl kickstart -k "gui/$(id -u)/com.fasttourney.prod-postgresql-backup-incremental"
tail -n 40 ~/Library/Logs/FastTourney/postgresql-prod-backup-incremental.log
```

El último gate de backup será restaurar desde la réplica ya recibida, no desde
el PVC local. En un terminal del Mac, lee la frase desde Contraseñas sin pegarla
en la conversación ni guardarla en un fichero y ejecuta:

```sh
read -rs PGBACKREST_REPO1_CIPHER_PASS
export PGBACKREST_REPO1_CIPHER_PASS
bash infra/k3s/scripts/restore-prod-replica-verify.sh
unset PGBACKREST_REPO1_CIPHER_PASS
```

Antes de restaurar, el verificador compara la frase enviada por SSH con el
Secret activo sin imprimir ninguno de ambos valores. Después monta
`FastTourney/postgresql-backups/prod` como solo lectura en un contenedor
temporal, restaura en un volumen temporal de Docker y comprueba
`fasttourney_prod|f`. No monta la VM, ningún PVC ni publica un puerto. Roles,
esquema y Goose ya están inicializados; la API y sus migraciones futuras siguen
siendo pasos explícitos.
