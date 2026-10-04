# Observabilidad de `prod` en K3s

> Estado: instalado y validado en la VM el 2026-09-05. Todos los componentes
> permanecen privados dentro de K3s.

## Límite

Este runbook instala Prometheus, Alertmanager, Loki, Tempo, Grafana y Alloy en
el namespace `prod` mediante charts upstream con versiones fijadas. No modifica
Caddy, Cloudflare Tunnel, el Ingress público ni el `503` de los hosts de
producción.

## Prerrequisitos

- K3s sano y el namespace `prod` existente.
- El operador SSH `fasttourney-operator`, con `sudo -n` limitado a la
  administración de K3s y Helm; el kubeconfig de K3s pertenece a `root`.
- Helm 3.18.6 instalado en la VM y accesible mediante
  `sudo env KUBECONFIG=/etc/rancher/k3s/k3s.yaml helm`.
- Un fichero local `alertmanager-resend-sending-key` de una sola línea y un fichero
  `grafana-admin.env` con `admin-user=admin` y una contraseña robusta.

Los ejemplos se ejecutan desde una copia temporal y privada del repositorio en
la VM (o desde su raíz, si existe allí). Los ficheros secretos no se copian al
repositorio.

**Evidencia 2026-09-05:** se descargó Helm 3.18.6 para Linux ARM64, se verificó
su SHA-256 oficial y se instaló en `/usr/local/bin/helm`. Quedaron desplegados
Prometheus/Alertmanager, Loki, Tempo, Grafana y Alloy; Prometheus obtiene
`up=1` de la API, Alloy abre streams de `prod` y Tempo recibió trazas OTLP HTTP.

## Preparación de secretos y dashboard

```sh
sudo /usr/local/bin/k3s kubectl -n prod create secret generic alertmanager-resend \
  --from-file=resend_alerts_sending_key=/ruta/privada/alertmanager-resend-sending-key \
  --dry-run=client -o yaml | sudo /usr/local/bin/k3s kubectl apply -f -

sudo /usr/local/bin/k3s kubectl -n prod create secret generic grafana-admin \
  --from-env-file=/ruta/privada/grafana-admin.env \
  --dry-run=client -o yaml | sudo /usr/local/bin/k3s kubectl apply -f -

sudo /usr/local/bin/k3s kubectl -n prod create configmap observability-grafana-dashboards \
  --from-file=session-refresh-slo.json=infra/observability/grafana/provisioning/dashboards/session-refresh-slo.json \
  --dry-run=client -o yaml | sudo /usr/local/bin/k3s kubectl apply -f -

sudo /usr/local/bin/k3s kubectl apply -f infra/k3s/observability/prometheus-config.yaml
```

Los comandos generan objetos sin imprimir sus valores. La sustitución de un
secreto no revoca una clave SMTP anterior: revocarla en Resend es una operación
separada si existiera exposición.

## Renderizado e instalación

```sh
sudo env KUBECONFIG=/etc/rancher/k3s/k3s.yaml helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
sudo env KUBECONFIG=/etc/rancher/k3s/k3s.yaml helm repo add grafana https://grafana.github.io/helm-charts
sudo env KUBECONFIG=/etc/rancher/k3s/k3s.yaml helm repo add grafana-community https://grafana-community.github.io/helm-charts
sudo env KUBECONFIG=/etc/rancher/k3s/k3s.yaml helm repo update

sudo env KUBECONFIG=/etc/rancher/k3s/k3s.yaml helm template observability-prometheus prometheus-community/prometheus --namespace prod --version 29.27.0 -f infra/k3s/observability/prometheus-values.yaml >/tmp/observability-prometheus.yaml
sudo env KUBECONFIG=/etc/rancher/k3s/k3s.yaml helm template observability-loki grafana-community/loki --namespace prod --version 18.11.7 -f infra/k3s/observability/loki-values.yaml >/tmp/observability-loki.yaml
sudo env KUBECONFIG=/etc/rancher/k3s/k3s.yaml helm template observability-tempo grafana-community/tempo --namespace prod --version 2.3.0 -f infra/k3s/observability/tempo-values.yaml >/tmp/observability-tempo.yaml
sudo env KUBECONFIG=/etc/rancher/k3s/k3s.yaml helm template observability-grafana grafana-community/grafana --namespace prod --version 13.0.1 -f infra/k3s/observability/grafana-values.yaml >/tmp/observability-grafana.yaml
sudo env KUBECONFIG=/etc/rancher/k3s/k3s.yaml helm template observability-alloy grafana/alloy --namespace prod --version 1.12.1 -f infra/k3s/observability/alloy-values.yaml >/tmp/observability-alloy.yaml
```

Revisar los cinco renderizados antes de aplicar: deben contener los PVC y
requests/limits esperados y no CRDs. Después instalar en este orden:

```sh
sudo env KUBECONFIG=/etc/rancher/k3s/k3s.yaml helm upgrade --install observability-prometheus prometheus-community/prometheus --namespace prod --version 29.27.0 -f infra/k3s/observability/prometheus-values.yaml --wait --timeout 5m
sudo env KUBECONFIG=/etc/rancher/k3s/k3s.yaml helm upgrade --install observability-loki grafana-community/loki --namespace prod --version 18.11.7 -f infra/k3s/observability/loki-values.yaml --wait --timeout 5m
sudo env KUBECONFIG=/etc/rancher/k3s/k3s.yaml helm upgrade --install observability-tempo grafana-community/tempo --namespace prod --version 2.3.0 -f infra/k3s/observability/tempo-values.yaml --wait --timeout 5m
sudo env KUBECONFIG=/etc/rancher/k3s/k3s.yaml helm upgrade --install observability-grafana grafana-community/grafana --namespace prod --version 13.0.1 -f infra/k3s/observability/grafana-values.yaml --wait --timeout 5m
sudo env KUBECONFIG=/etc/rancher/k3s/k3s.yaml helm upgrade --install observability-alloy grafana/alloy --namespace prod --version 1.12.1 -f infra/k3s/observability/alloy-values.yaml --wait --timeout 5m
sudo /usr/local/bin/k3s kubectl apply -f infra/k3s/core/api-config.yaml
sudo /usr/local/bin/k3s kubectl -n prod rollout restart deployment/api
sudo /usr/local/bin/k3s kubectl -n prod rollout status deployment/api --timeout=120s
```

## Verificación

```sh
sudo env KUBECONFIG=/etc/rancher/k3s/k3s.yaml helm list -n prod
sudo /usr/local/bin/k3s kubectl -n prod get pods,pvc
sudo /usr/local/bin/k3s kubectl -n prod get events --sort-by=.lastTimestamp
```

Comprobar que Prometheus obtiene el target `tournaments-manager-api`, que una
petición de refresh deja logs correlacionados en Loki y una traza en Tempo, y
que Grafana muestra el dashboard SLO.

### Auditoría de retención y almacenamiento

Auditoría y corrección del 2026-10-03 conforme a ADR-0140: siete días de
logs diagnósticos, noventa de seguridad y del histórico sin categoría. El
compactor debe estar habilitado; Tempo mantiene siete días y Prometheus un día
o 128 MB TSDB. Evidencia completa en `docs/operations/LOG_RETENTION.md`.

Consultar únicamente metadatos y configuración técnica; no imprimir Secrets,
valores Helm completos, cuerpos de logs ni el kubeconfig. Desde el Mac:

```sh
set -a
. infra/k3s/.env
set +a
ssh -o BatchMode=yes -o ConnectTimeout=10 \
  -i ~/.ssh/fasttourney_k3s_operator "$K3S_SSH_USER@$K3S_SSH_HOST" \
  'sudo -n /usr/local/bin/k3s kubectl -n prod get pods,pvc,svc;
   df -h /;
   sudo -n du -sh /var/lib/rancher/k3s/storage /var/log/pods /var/log/journal;
   sudo -n journalctl --disk-usage'
```

Una vez accesible el host, completar y registrar estas evidencias:

- Loki: configuración activa de retención y compactor, directorio persistente
  de marcadores, estado del borrado y antigüedad de los datos conservados.
- Tempo: retención efectiva de bloques de siete días, compactación y antigüedad
  de bloques completados.
- Prometheus: argumentos efectivos de 24 horas y 128 MB, antigüedad de bloques,
  espacio de TSDB, WAL y head. La retención por tamaño no es una cuota de disco.
- Host: rotación efectiva del kubelet y límites del journal; esas copias son
  independientes de Loki. Comprobar también espacio libre y crecimiento.
- Seguridad: comprobar categorías de autenticación, credenciales, RISC y límite
  de tasa, con noventa días conforme a ADR-0140; diagnóstico nuevo tiene siete
  días. El histórico mezclado sin categoría mantiene noventa días.

No borrar PVC ni cambiar reglas durante esta auditoría. Un PVC de 5 GiB no
demuestra por sí mismo que el almacenamiento imponga esa cuota. Si el host no
responde, registrar la comprobación como pendiente, nunca como conforme.

### Instalar los controles de espacio de ADR-0140

Copiar a un directorio temporal privado de la VM los archivos versionados:
`infra/k3s/host/observability-disk-metrics.py`,
`fasttourney-disk-metrics.service`, `fasttourney-disk-metrics.timer` y
`fasttourney-journald.conf`. Como administrador, instalar el Python en
`/usr/local/lib/fasttourney/observability-disk-metrics.py`, las unidades en
`/etc/systemd/system/` y el drop-in en
`/etc/systemd/journald.conf.d/fasttourney-retention.conf`. Crear primero
`/var/lib/fasttourney/observability-metrics` con modo 755. Archivos modo 644,
propiedad root; no contienen credenciales.

Validar las unidades con `systemd-analyze verify`, ejecutar `systemctl daemon-reload`,
arrancar `fasttourney-disk-metrics.service` y habilitar su timer. Reiniciar
`systemd-journald` para aplicar sus límites. No ejecutar un vacuum indiscriminado:
la propia política se encarga de sus archivos vencidos y rotados.

Renderizar primero Alloy mediante Helm y validar el `config.alloy` del ConfigMap
resultante con `/bin/alloy validate`: su plantilla interna debe sobrevivir a
`tpl` de Helm. Validar los valores sin renderizar no prueba ese comportamiento.
Aplicar Alloy antes de habilitar compactor y probar eventos sintéticos: diagnóstico,
autenticación, límite, salud fallida, salud correcta, RISC y JSON inválido.
La salud correcta debe descartarse; seguridad, RISC y JSON inválido se preservan.
Los casos están versionados en `production-retention-fixtures.json` y las pruebas
Prometheus en `production-storage-rules.test.yaml`; para estas últimas, extraer
el grupo `production-storage` del ConfigMap como `production-storage-rules.yml`
en el mismo directorio y ejecutar `promtool test rules`.
Usar imágenes ya instaladas con `imagePullPolicy: Never`, sin credenciales ni
ServiceAccount, y retirar los Pods de fixtures al terminar.

Comprobar el endpoint privado del collector:
`http://observability-alloy:12345/api/v0/component/prometheus.exporter.unix.disk/metrics`.
Debe emitir tamaños de los tres volúmenes, capacidad y espacio libre de raíz,
y `fasttourney_disk_check_timestamp_seconds`. Prometheus debe mostrar
`up{job="fasttourney-disk"}=1`, las reglas de espacio y de retención Loki cargadas y una muestra
con menos de quince minutos de antigüedad. No se comprueba el correo provocando
falta real de espacio: las condiciones de alerta se validan con fixtures.

Si falla el timer, el archivo anterior conserva su timestamp y se alerta por
medición vencida. Si el Service o exporter falla, se alerta por target caído o
métrica ausente. Un PVC local-path no impone cuota; atender las alertas antes de
que se agote la raíz y revisar aislamiento de almacenamiento si crece el volumen.

### Acceso privado a Grafana

Grafana es un `Service` `ClusterIP`: solo existe dentro de la red de K3s. Un
túnel SSH crea un puerto temporal en el Mac sin crear un Ingress, abrir un
puerto de la VM ni tocar Caddy o Cloudflare. Primero consultar el `clusterIP`
actual, porque no se debe fijar en la documentación:

```sh
ssh -o BatchMode=yes fasttourney-k3s \
  'sudo /usr/local/bin/k3s kubectl -n prod get service observability-grafana \
   -o jsonpath="{.spec.clusterIP}"; echo'

ssh -N -L 127.0.0.1:13000:<cluster-ip-de-grafana>:80 fasttourney-k3s
```

Con el segundo proceso abierto, acceder a `http://127.0.0.1:13000/login` y
autenticarse con el usuario `admin` y la contraseña del llavero local. Cerrar
el proceso SSH elimina el acceso. Nunca exponer esa contraseña ni recuperarla
desde Kubernetes para copiarla a la terminal.

### Prueba controlada de Alertmanager y Resend

Antes del gate público, enviar una alerta sintética, de duración corta, a la
API v2 privada de Alertmanager. Es la prueba mínima para el canal de correo:
no detiene PostgreSQL ni genera fallos reales de refresh. Las etiquetas deben
incluir `test="true"`, un `alertname` inequívoco y un identificador de ejecución
no sensible. La alerta debe llevar `endsAt` cercano para que su rollback sea
automático.

La confirmación no termina al aceptar el `POST`: comprobar que la alerta pasa
de activa a resuelta y que `alertmanager_notifications_total{integration="email"}`
aumenta dos veces —alerta y resolución— sin aumento de
`alertmanager_notifications_failed_total`. La resolución puede aparecer hasta
el `group_interval` configurado después de que cese la alerta. Los contadores
prueban que Resend aceptó SMTP; la entrega final en el buzón requiere una
comprobación independiente desde ese buzón.

**Evidencia 2026-09-05:** el dashboard provisionado `session-refresh-slo`
(`SLO — Refresh de sesión`) cargó con cinco paneles. Grafana confirmó
Prometheus, Loki y Tempo; su proxy obtuvo `200` de Alertmanager `/api/v2/status`.
El endpoint opcional de salud del plugin de Alertmanager devolvió
`Plugin unavailable`, por lo que no se usa como prueba de conectividad. La
alerta sintética `FastTourneyAlertDeliveryTest` fue aceptada, se resolvió por
`endsAt` sin tocar API ni PostgreSQL y produjo dos notificaciones SMTP sin
fallos: alerta y resolución. La persona operadora confirmó la llegada de ambos
correos al buzón receptor; así queda validado también el salto posterior a
Resend.

## Rollback

Si una actualización de chart falla, inspeccionar primero `helm status` y los
eventos; volver a la revisión Helm inmediatamente anterior solo si sus PVC y
configuración siguen siendo compatibles:

```sh
sudo env KUBECONFIG=/etc/rancher/k3s/k3s.yaml helm history observability-loki -n prod
sudo env KUBECONFIG=/etc/rancher/k3s/k3s.yaml helm rollback observability-loki <revision-anterior> -n prod --wait --timeout 5m
```

Repetir por release solo para la pieza fallida. Un rollback de Helm no restaura
ni sustituye datos de PVC y nunca revierte el esquema PostgreSQL. Conservar los
renderizados y el estado antes de desinstalar cualquier release persistente.
