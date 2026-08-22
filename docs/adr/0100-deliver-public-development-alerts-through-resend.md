# ADR-0100: Entregar las alertas de desarrollo público mediante Resend

- **Estado:** Aceptado
- **Fecha:** 2026-08-21
- **Decisor:** Usuario, mediante decisión explícita
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

El SLO de refresh ya alerta en local mediante Mailpit, pero `dev.fasttourney.com`
necesita conservar la misma capacidad de detección y avisar fuera del Mac. Mailpit
es deliberadamente local y no puede entregar ese aviso.

## Contexto y restricciones

- ADR-0091 separa `tournaments-manager-dev` de local y prohíbe publicar Mailpit.
- ADR-0093 ya acepta Resend mediante SMTP autenticado con STARTTLS para el correo
  transaccional de `dev`.
- ADR-0099 mantiene Prometheus como evaluador y Alertmanager como responsable de
  agrupar, silenciar y entregar alertas; su receptor externo requería una decisión
  separada.
- Grafana, Prometheus y Alertmanager siguen siendo interfaces operativas locales:
  se ligan solo a loopback y no se añaden rutas en Caddy o Cloudflare Tunnel.
- Las alertas y sus etiquetas no incluyen secretos, tokens ni PII.

## Alternativas

### Alternativa A — Reutilizar la clave SMTP transaccional de Resend

- **Ventajas:** un secreto menos que crear y rotar.
- **Inconvenientes:** revocar o filtrar la clave de alertas interrumpe verificaciones
  y recuperaciones; no separa responsabilidades operativas.
- **Coste de mantenimiento:** bajo, con radio de impacto mayor.

### Alternativa B — Resend SMTP con una clave exclusiva de Alertmanager

Alertmanager usa `smtp.resend.com:587`, el usuario `resend` y STARTTLS obligatorio.
La contraseña se monta desde un archivo secreto local fuera de Git; el correo
transaccional conserva su clave en `api.docker.env`.

- **Ventajas:** mismo protocolo y proveedor ya aceptados, pero revocación,
  trazabilidad y rotación independientes; no añade SDK ni servicio.
- **Inconvenientes:** hay una segunda clave que custodiar y revisar.
- **Coste de mantenimiento:** bajo.

### No cambiar

Mantener las alertas solo en Grafana y Mailpit local.

- **Consecuencias:** `dev` puede degradarse sin aviso fuera de la interfaz local.

## Recomendación

**Opinión/recomendación:** alternativa B. Es la separación mínima que limita el
impacto de una credencial sin introducir otro proveedor ni acoplar el dominio a
una API propietaria.

## Decisión del usuario

**Aceptada el 2026-08-21:** `tournaments-manager-dev` ejecuta el mismo stack de
Prometheus, Alertmanager, Loki, Tempo, Promtail y Grafana que local. Alertmanager
entrega avisos `warning` y `critical` por SMTP de Resend con STARTTLS, con una clave
*Sending access* exclusiva. El remitente visible es `FastTourney Dev Alerts`
`<alerts@mail.fasttourney.com>`, el asunto empieza por `[DEV]` y el receptor
inicial es `alerts@fasttourney.com`. El correo transaccional visible es
`FastTourney Dev <no-reply@mail.fasttourney.com>` y sus asuntos también empiezan
por `[DEV]`.

## Consecuencias

### Positivas

- El SLO de refresh conserva reglas, agrupación, silencios, dashboard y rutas en
  ambos entornos.
- Revocar la clave de alertas no afecta al envío de enlaces de identidad.
- Las interfaces de observabilidad no se exponen a visitantes de `dev`.

### Negativas y deuda aceptada

- El aviso depende de Internet, Resend y de la cuota de la cuenta.
- El stack añade volúmenes y seis contenedores al Mac; conserva retención corta de
  24 horas y no aporta alta disponibilidad.

## Validación

1. `docker compose ... config` resuelve el secreto sin mostrar su valor.
2. Prometheus descubre API y Alertmanager; Grafana en `127.0.0.1:3001` muestra
   reglas y alertas del entorno `dev`.
3. Una alerta de prueba llega a `alerts@fasttourney.com` y llega su resolución.
4. Parar PostgreSQL provoca la señal segura, trazas y logs correlacionados; al
   recuperarlo la API vuelve a servir sin exponer detalles internos.
5. Ni la clave de Alertmanager ni la transaccional aparecen en Git, imágenes o
   logs.

## Disparadores de revisión

- Alertas no accionables, ruido repetido, rebotes o cuota insuficiente.
- Necesidad de varios receptores, guardias, escalado o un entorno adicional.
- Migración a un backend gestionado, Collector o runtime fuera del Mac.

## Documentación afectada

- `infra/dev/compose.yaml`
- `infra/dev/README.md`
- `docs/operations/OBSERVABILITY.md`
- `docs/runbooks/session-refresh-observability.md`
- `docs/governance/DECISIONS.md`
- `docs/governance/DECISIONS_TO_REVISIT.md`
- `docs/project/LEARNING.md`
- `CHANGELOG.md`
