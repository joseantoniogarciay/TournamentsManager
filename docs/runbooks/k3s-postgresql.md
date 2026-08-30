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

El siguiente gate es aplicar ADR-0114: automatizar las copias y la réplica
atómica hacia `FastTourney/postgresql-backups/prod`, y repetir esta verificación
desde la réplica recibida. Roles separados, esquema y Goose siguen siendo un
paso explícito.
