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

## Credencial del borde antes de publicar

ADR-0118 exige crear el Secret independiente `api-edge-proxy` desde una copia
privada de `infra/k3s/secrets/api-edge-proxy.env.example`. Generar un valor con
`openssl rand -hex 32`, guardarlo también en el gestor de secretos del Mac y
cargarlo como `FASTTOURNEY_API_EDGE_TOKEN` para el proceso Caddy. Crear el
Secret sin imprimirlo:

```sh
sudo /usr/local/bin/k3s kubectl -n prod create secret generic api-edge-proxy \
  --from-env-file=/ruta/privada/api-edge-proxy.env \
  --dry-run=client -o yaml | sudo /usr/local/bin/k3s kubectl apply -f -
```

La API usa `EDGE_PROXY_AUTH_TOKEN` desde ese Secret; Caddy sobrescribe el token
y `X-Client-IP` al activar su `reverse_proxy`. Si falta o difiere el token, la
API usa la IP inmediata y no acepta una cabecera falsificada. La rotación crea
el Secret nuevo, reinicia la API, recarga Caddy y prueba dos IP de ejemplo; el
rollback restaura ambos valores de la misma versión. No cambiar solo un lado.

**Evidencia de publicación, 2026-09-05:** Caddy se validó antes de recargarse;
el primer intento devolvió `502` mientras el servicio reiniciaba y se restauró
el `503`. El gate definitivo espera a que el listener loopback esté disponible.
Después, `api.fasttourney.com/healthz` devolvió `200` tanto por loopback como a
través de Cloudflare; las dos réplicas de API estaban `Running` y Prometheus
conservó `up{job="tournaments-manager-api"}=1`.
