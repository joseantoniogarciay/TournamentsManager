# ADR-0087: Usar Caddy como borde HTTPS de la beta doméstica

- **Estado:** Superado por ADR-0090
- **Fecha:** 2026-08-11
- **Decisor:** Usuario, mediante aceptación explícita de Caddy
- **Propietario del análisis:** Asistente como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** ADR-0090

## Problema

La beta doméstica se ejecutará detrás de una IP pública dinámica. Necesita una
entrada HTTPS estable que no exponga directamente la API ni PostgreSQL.

## Contexto y restricciones

- La WAN del Cloud Gateway Ultra tiene una IPv4 pública dinámica; se usará DDNS
  con un dominio antes de abrir el servicio a Internet.
- La API se empaqueta como imagen OCI (ADR-0022), pero el runtime cloud futuro
  sigue siendo ECS con Fargate detrás de un ALB (ADR-0023 y ADR-0029).
- Esta decisión solo cubre la beta doméstica y no crea recursos AWS ni modifica
  la dirección cloud aceptada.
- Los secretos y credenciales DNS quedan fuera de Git conforme a ADR-0017.

## Criterios de decisión

1. TLS automático con el menor mantenimiento posible;
2. una sola entrada pública HTTPS y API/BD no expuestas directamente;
3. configuración breve y revisable para un único servicio inicial;
4. sustitución sencilla al pasar al ALB de AWS.

## Alternativas

### Alternativa A — Caddy

- **Ventajas:** obtiene y renueva certificados automáticamente; configuración
  declarativa pequeña; reverse proxy integrado.
- **Inconvenientes:** añade un proceso y sus actualizaciones al host.
- **Coste de adopción y mantenimiento:** bajo.
- **Riesgos:** abrir puertos antes de tener dominio, proxy y firewall correctos.

### Alternativa B — Nginx con Certbot

- **Ventajas:** patrón muy conocido y flexible.
- **Inconvenientes:** separa la configuración del proxy y la renovación de
  certificados, con más piezas que operar.
- **Coste de adopción y mantenimiento:** medio.

### Alternativa C — Traefik

- **Ventajas:** integración dinámica con varios contenedores.
- **Inconvenientes:** capacidades y etiquetas adicionales sin valor probado
  para una única API.
- **Coste de adopción y mantenimiento:** medio.

### No cambiar

- **Consecuencias:** no hay una entrada HTTPS segura y estable para la beta.

## Comparación

Caddy cumple la entrada HTTPS única y automatiza el certificado sin introducir
el descubrimiento dinámico de Traefik ni el ciclo adicional de Certbot. Nginx
sería válido si aparecieran reglas complejas que justificasen su configuración.

## Recomendación

**Opinión/recomendación:** alternativa A, Caddy, como solución mínima
suficiente para la beta doméstica.

## Decisión del usuario

**Aceptada:** usar Caddy como reverse proxy HTTPS de la beta doméstica. Se
instala en el Mac anfitrión; su configuración pública queda bloqueada hasta
disponer de un dominio/DDNS, una API destino y una revisión de firewall.

### Aclaración de alcance — 2026-08-11

ADR-0088 sustituyó AWS permanente por uso efímero de aprendizaje. ADR-0089 fija
los hosts `fasttourney.com`, `dev.fasttourney.com` y `api.fasttourney.com`; esta
aclaración no cambia la decisión de usar Caddy como borde doméstico.

## Consecuencias

### Positivas

- El tráfico público llegará por HTTPS a un único borde.
- La API y PostgreSQL podrán permanecer en una red no pública.
- El dominio podrá conservarse al migrar después al ALB.

### Negativas y deuda aceptada

- El Mac, la conexión doméstica y Caddy son un único punto de fallo.
- Deben mantenerse actualizados el host, Caddy y sus reglas de red.

## Validación

- Caddy está instalado y su configuración se valida antes de iniciarse.
- El dominio DDNS resuelve la WAN actual y el certificado se emite para él.
- Solo TCP 80/443 se reenvían al borde; la API y PostgreSQL no son accesibles
  directamente desde Internet.
- Una prueba externa confirma HTTPS válido y que el proxy alcanza el health
  check de la API.

## Disparadores de revisión

- varios servicios o rutas requieren descubrimiento dinámico;
- un incidente, una necesidad de WAF o de alta disponibilidad;
- migración al ALB de AWS o abandono de la beta doméstica.

## Documentación afectada

- [DEPLOYMENT.md](../operations/DEPLOYMENT.md)
- [SECURITY.md](../engineering/SECURITY.md)
- [DECISIONS.md](../governance/DECISIONS.md)
- [DECISIONS_TO_REVISIT.md](../governance/DECISIONS_TO_REVISIT.md)
- [LEARNING.md](../project/LEARNING.md)

## Fuentes técnicas

- [Caddy: Automatic HTTPS](https://caddyserver.com/docs/automatic-https)
- [Caddy: reverse_proxy](https://caddyserver.com/docs/caddyfile/directives/reverse_proxy)
