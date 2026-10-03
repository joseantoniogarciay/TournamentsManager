# Observabilidad de `prod` con Helm

Este directorio contiene los valores revisables de los charts upstream que instalan observabilidad de terceros en el namespace `prod`. No es un chart propio: API y PostgreSQL siguen como manifiestos explícitos, conforme a ADR-0112.

## Perfil aceptado

| Señal              | Release Helm               | Chart fijado                              | Persistencia y retención inicial             |
| ------------------ | -------------------------- | ----------------------------------------- | -------------------------------------------- |
| Métricas y alertas | `observability-prometheus` | `prometheus-community/prometheus` 29.27.0 | Prometheus: 5 GiB solicitados, 24 h o 128 MB TSDB; Alertmanager: 1 GiB |
| Logs               | `observability-loki`       | `grafana-community/loki` 18.11.7          | Loki monolítico: 5 GiB solicitados; diagnóstico 7 días, seguridad 90 días                 |
| Trazas             | `observability-tempo`      | `grafana-community/tempo` 2.3.0           | Tempo monolítico: 5 GiB, 7 días              |
| Consulta           | `observability-grafana`    | `grafana-community/grafana` 13.0.1        | Grafana: 2 GiB                               |
| Recogida de logs   | `observability-alloy`      | `grafana/alloy` 1.12.1                    | Solo Pods de `prod` y lectura de métricas agregadas del disco    |

Los PVC local-path solicitan capacidad pero no imponen cuotas de disco. La retención efectiva y las alertas de espacio de ADR-0140 son controles independientes. No se habilitan HA, autoscaling, Prometheus Operator, scraping exhaustivo de K3s ni recolección general de métricas o trazas de Alloy. Solo se habilita su collector textfile para agregados de disco.

`prometheus-config.yaml` se aplica antes del release de Prometheus. Sustituye la
configuración por defecto del chart para evitar descubrimiento Kubernetes y
mantener scrapes estáticos de la API y del collector textfile de Alloy.
Así, Prometheus no necesita un token de ServiceAccount ni permisos de lectura
del clúster.

Alloy usa la API Kubernetes para leer los logs de Pods del namespace `prod`; no monta directorios de logs del host ni recoge `kube-system`. Sus permisos se restringen a `pods`, `pods/log` y `namespaces` de ese namespace. Solo monta en lectura el directorio de métricas agregadas descrito al final.

## Secret de Alertmanager

Antes de instalar, el operador crea en la VM el secreto `alertmanager-resend` con la clave `resend_alerts_sending_key`, usando el fichero local no versionado basado en `../secrets/alertmanager-resend-sending-key.example`. El chart solo monta ese fichero en `/etc/alertmanager-secrets`; la clave nunca entra en `values`, manifiestos renderizados ni logs.

El receptor de `prod` usa el remitente `FastTourney Alerts` y el prefijo `[PROD]`; no reutiliza el secreto de `dev`.

## Renderizado antes de aplicar

Los cinco comandos `helm template` deben ejecutarse con las versiones anteriores y sus valores correspondientes. La instalación no cambia Caddy, Cloudflare Tunnel, el Ingress público ni el `503` de `api.fasttourney.com`.

El procedimiento completo, incluida la instalación interactiva en la VM, la validación y rollback, está en [`docs/runbooks/k3s-observability.md`](../../../docs/runbooks/k3s-observability.md).


La política vigente es [ADR-0140](../../../docs/adr/0140-bound-production-telemetry-retention.md).
Alloy monta en solo lectura `/var/lib/fasttourney/observability-metrics`, generado
por el timer del host cada cinco minutos. No monta la raíz ni recolecta logs del
host. Loki conserva el histórico sin categoría hasta noventa días, y activa
compactor persistente para la purga por categoría.
