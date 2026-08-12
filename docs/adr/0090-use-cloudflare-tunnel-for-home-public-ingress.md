# ADR-0090: Usar Cloudflare Tunnel como entrada pública doméstica

- **Estado:** Aceptado
- **Fecha:** 2026-08-11
- **Decisor:** Usuario, mediante aceptación explícita de Cloudflare Tunnel
- **Propietario del análisis:** Asistente como mentor técnico
- **Supera a:** ADR-0087
- **Superado por:** Ninguno

## Problema

El borde doméstico con puertos reenviados revela la IP residencial y permite que
el enlace WAN del Mac reciba tráfico directo, incluso si Cloudflare proxy está
activo. La IP es dinámica y el mantenimiento de DDNS y forwards añade trabajo.

## Contexto y restricciones

- El Mac ejecuta un conector `cloudflared` saludable de salida hacia Cloudflare.
- Los hosts públicos son `fasttourney.com`, `www.fasttourney.com`,
  `dev.fasttourney.com`, `api.fasttourney.com` y `dev-api.fasttourney.com`.
- El proyecto es personal; Cloudflare Access no se habilita por ahora. Sus
  usuarios no son usuarios de FastTourney.
- Caddy ya está instalado y aporta una configuración corta y revisable para
  enrutar por hostname en el Mac.

## Alternativas

### Alternativa A — Cloudflare Tunnel y Caddy local

- **Ventajas:** conexión únicamente saliente, sin puertos WAN ni DDNS; la IP de
  casa no se publica como origen DNS; TLS, proxy y mitigación HTTP se sitúan en
  Cloudflare.
- **Inconvenientes:** dependencia adicional de Cloudflare y de `cloudflared`.
- **Coste de mantenimiento:** bajo; revisar el conector y sus actualizaciones.

### Alternativa B — DNS proxied con puertos 80/443

- **Ventajas:** menos software local adicional.
- **Inconvenientes:** un atacante que conozca la IP puede atacar directamente el
  enlace residencial; conserva forwards y DDNS.
- **Coste de mantenimiento:** medio.

## Decisión del usuario

**Aceptada:** Cloudflare Tunnel será la entrada pública doméstica. `cloudflared`
conecta de forma saliente desde el Mac; Cloudflare termina TLS y enruta cada
hostname al Caddy local en `127.0.0.1:9080`. Caddy conserva el enrutamiento y los
proxies futuros, pero no escucha en LAN/WAN ni gestiona certificados públicos.

Tras validar cada ruta del túnel, se eliminarán los port forwards TCP 80/443 de
UniFi, el perfil DDNS y su token DNS. La API y PostgreSQL no se publican.

## Consecuencias

- La IP doméstica deja de ser el origen público configurado para FastTourney,
  pero un ataque directo contra una IP previamente conocida podría aún saturar
  el enlace del ISP; no existe una garantía de mitigación L3/L4 residencial.
- El plan Zero Trust Free usado para el túnel no convierte a visitantes de la
  aplicación en usuarios de Access ni activa cobro por sí mismo.
- El futuro uso efímero de AWS no queda condicionado por esta decisión.

## Validación

1. El conector `fasttourney-home` figura como Healthy.
2. Cada hostname responde por HTTPS a través del túnel y muestra el `503`
   previsto antes de publicar aplicaciones.
3. Caddy solo escucha en `127.0.0.1:9080`.
4. Tras validar, UniFi no reenvía TCP 80/443 y Cloudflare DNS no publica la IP
   residencial para estos hosts.

## Disparadores de revisión

- requisito de un servicio no HTTP, identidad Access o disponibilidad superior;
- incidente de Cloudflare Tunnel o migración estable fuera del Mac.

## Fuentes técnicas

- [Cloudflare Tunnel](https://developers.cloudflare.com/tunnel/)
- [Protección del origen](https://developers.cloudflare.com/fundamentals/security/protect-your-origin-server/)
