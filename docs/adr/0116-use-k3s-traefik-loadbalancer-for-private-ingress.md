# ADR-0116: Usar el Traefik y LoadBalancer incluidos en K3s para el Ingress privado

- **Estado:** Aceptado
- **Fecha:** 2026-09-01
- **Decisor:** Usuario, mediante decisión explícita
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

La API de `prod` necesita un punto HTTP estable dentro de la VM K3s para que
Caddy pueda alcanzarla sin exponer Pods, PostgreSQL ni puertos LAN/WAN.

## Contexto y restricciones

- Traefik ya es el `IngressClass` predeterminado de K3s y su Service
  `LoadBalancer` expone `192.168.64.2:80/443` en la red privada de la VM.
- La única réplica actual no obtiene alta disponibilidad ni reparto de carga;
  el LoadBalancer proporciona una frontera de red estable entre la IP de la VM
  y el Service de Traefik.
- Cloudflare Tunnel y Caddy siguen siendo el borde público aceptado por ADR-0090.
- `api.fasttourney.com` permanece en `503` hasta completar todos los gates de
  ADR-0111.

## Alternativas

### A — Traefik incluido y Service LoadBalancer de K3s

- **Ventajas:** no instala otro controlador, usa 80/443 estables y conserva
  Ingress, Service y Pod desacoplados.
- **Inconvenientes:** en un nodo no aporta balanceo ni alta disponibilidad.
- **Coste:** bajo; componentes ya operativos en K3s.

### B — NodePort directo

- **Ventajas:** elimina la capa ServiceLB.
- **Inconvenientes:** Caddy depende de un puerto alto explícito y no reduce una
  superficie real, pues sigue siendo un puerto de la VM.
- **Coste:** bajo, con peor convención operativa.

### C — Túnel SSH persistente

- **Ventajas:** aislamiento adicional.
- **Inconvenientes:** proceso extra, supervisión y fallos sin capacidad útil en
  una red de VM ya privada.
- **Coste:** medio; sobreingeniería para el caso actual.

## Decisión del usuario

**Aceptada el 2026-09-01:** usar Traefik incluido en K3s y su Service
LoadBalancer privado. Caddy alcanzará `192.168.64.2:80`; Traefik resolverá los
Ingress hacia Services internos. No se instala otro controlador ni se usa
NodePort explícito.

## Validación

1. El Ingress `api` usa la clase `traefik` y apunta al Service `api:http`.
2. Desde el Mac, `curl --resolve api.fasttourney.com:80:192.168.64.2
http://api.fasttourney.com/healthz` devuelve la salud de la API.
3. Caddy y Cloudflare conservan `503` hasta el gate de publicación.

## Disparadores de revisión

- Más nodos, alta disponibilidad, necesidad de TLS dentro de la VM o de reglas
  HTTP que Traefik no pueda mantener de forma simple.
