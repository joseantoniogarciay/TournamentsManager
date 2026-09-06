# Observabilidad de `prod` con Helm

Este directorio contiene los valores revisables de los charts upstream que instalan observabilidad de terceros en el namespace `prod`. No es un chart propio: API y PostgreSQL siguen como manifiestos explícitos, conforme a ADR-0112.

## Perfil aceptado

| Señal              | Release Helm               | Chart fijado                              | Persistencia y retención inicial             |
| ------------------ | -------------------------- | ----------------------------------------- | -------------------------------------------- |
| Métricas y alertas | `observability-prometheus` | `prometheus-community/prometheus` 29.27.0 | Prometheus: 5 GiB, 24 h; Alertmanager: 1 GiB |
| Logs               | `observability-loki`       | `grafana-community/loki` 18.11.7          | Loki monolítico: 5 GiB, 24 h                 |
| Trazas             | `observability-tempo`      | `grafana-community/tempo` 2.3.0           | Tempo monolítico: 5 GiB, 7 días              |
| Consulta           | `observability-grafana`    | `grafana-community/grafana` 13.0.1        | Grafana: 2 GiB                               |
| Recogida de logs   | `observability-alloy`      | `grafana/alloy` 1.12.1                    | Sin dato persistente; solo Pods de `prod`    |

Los límites no pretenden reservar toda la VM: son una primera barrera frente a un consumo descontrolado. No se habilitan HA, autoscaling, Prometheus Operator, scraping exhaustivo de K3s ni las capacidades de métricas o trazas de Alloy.

`prometheus-config.yaml` se aplica antes del release de Prometheus. Sustituye la
configuración por defecto del chart para evitar descubrimiento Kubernetes y
mantener únicamente el scrape estático de `api.prod.svc.cluster.local:8080`.
Así, Prometheus no necesita un token de ServiceAccount ni permisos de lectura
del clúster.

Alloy usa la API Kubernetes para leer los logs de Pods del namespace `prod`; no monta directorios del host ni recoge `kube-system`. Sus permisos se restringen a `pods`, `pods/log` y `namespaces` de ese namespace.

## Secret de Alertmanager

Antes de instalar, el operador crea en la VM el secreto `alertmanager-resend` con la clave `resend_alerts_sending_key`, usando el fichero local no versionado basado en `../secrets/alertmanager-resend-sending-key.example`. El chart solo monta ese fichero en `/etc/alertmanager-secrets`; la clave nunca entra en `values`, manifiestos renderizados ni logs.

El receptor de `prod` usa el remitente `FastTourney Alerts` y el prefijo `[PROD]`; no reutiliza el secreto de `dev`.

## Renderizado antes de aplicar

Los cinco comandos `helm template` deben ejecutarse con las versiones anteriores y sus valores correspondientes. La instalación no cambia Caddy, Cloudflare Tunnel, el Ingress público ni el `503` de `api.fasttourney.com`.

El procedimiento completo, incluida la instalación interactiva en la VM, la validación y rollback, está en [`docs/runbooks/k3s-observability.md`](../../docs/runbooks/k3s-observability.md).
