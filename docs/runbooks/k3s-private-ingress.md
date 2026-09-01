# Ingress privado de la API en K3s

> Estado: aplicado y verificado el 2026-09-01.

## Límite

Este módulo crea el Ingress interno de `prod`. No modifica Caddy, Cloudflare
Tunnel ni el `503` de `api.fasttourney.com`; por tanto no publica producción.

## Aplicación y comprobación

En una sesión SSH interactiva de la VM:

```sh
sudo /usr/local/bin/k3s kubectl apply --dry-run=server -f infra/k3s/core/api-ingress.yaml
sudo /usr/local/bin/k3s kubectl apply -f infra/k3s/core/api-ingress.yaml
sudo /usr/local/bin/k3s kubectl -n prod get ingress api
```

Desde el Mac, la comprobación privada atraviesa el LoadBalancer de Traefik sin
usar Caddy ni Cloudflare:

```sh
curl --fail --resolve api.fasttourney.com:80:192.168.64.2 \
  http://api.fasttourney.com/healthz
```

El éxito demuestra `VM IP → LoadBalancer → Traefik → Ingress → Service → Pod`.
No pruebes aún el hostname público: debe seguir devolviendo `503`.

**Evidencia:** el API server aceptó el dry-run y creó `Ingress/api` con clase
`traefik`, host `api.fasttourney.com` y dirección privada `192.168.64.2`. Desde
el Mac, la resolución forzada de ese host a la IP privada devolvió `HTTP 200`
en `/healthz`, sin pasar por Caddy ni Cloudflare.
