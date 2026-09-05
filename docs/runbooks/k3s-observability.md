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
que Grafana muestra el dashboard SLO. Antes del gate público, provocar una
caída controlada de PostgreSQL, confirmar alerta `critical` y comprobar su
resolución al recuperar la dependencia.

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
