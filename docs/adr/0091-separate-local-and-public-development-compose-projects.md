# ADR-0091: Separar los proyectos Compose local y de desarrollo público

- **Estado:** Aceptado
- **Fecha:** 2026-08-11
- **Decisor:** Usuario, mediante decisión explícita
- **Propietario del análisis:** Asistente como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

El ciclo local con Air y Mailpit es apropiado para edición diaria, pero no debe
ser el mismo estado, código vivo ni base de datos que usan personas externas en
`dev.fasttourney.com`.

## Alternativas

### Alternativa A — Reutilizar el Compose local para publicar dev

- **Ventajas:** ninguna nueva configuración ni volumen.
- **Inconvenientes:** cada guardado con Air altera el entorno público; datos y
  correo de pruebas se mezclan con el trabajo diario.
- **Coste de mantenimiento:** bajo, con riesgo operativo alto.

### Alternativa B — Proyecto Compose runtime independiente para dev público

- **Ventajas:** datos, puertos y configuración aislados; usa la misma imagen
  `runtime` mínima que se validará para release; no publica PostgreSQL ni Mailpit.
- **Inconvenientes:** un segundo volumen PostgreSQL y contratos de entorno.
- **Coste de mantenimiento:** bajo o medio.

## Decisión del usuario

**Aceptada:** mantener tres contextos: `tournaments-manager-local` para trabajo
diario con Air, `tournaments-manager-dev` para desarrollo público, y un futuro
`tournaments-manager-prod` para release doméstico. Los tres reutilizan los
nombres de servicio internos `api` y `postgres`; el nombre del proyecto Compose
aporta el namespace y aislamiento.

`tournaments-manager-dev` usa el target `runtime`, PostgreSQL y volumen propios,
y publica solo la API en `127.0.0.1:8081` para Caddy. La web se exporta estática
y Caddy la sirve. Mailpit es exclusivamente diagnóstico local en `127.0.0.1`;
un SMTP transaccional real es requisito previo para invitar usuarios que deban
recibir correo.

## Consecuencias

- Las pruebas locales no modifican datos ni despliegue de dev público.
- Dev público no obtiene hot reload ni acceso público a Docker, PostgreSQL o
  Mailpit.
- Producción repetirá el patrón con proyecto, configuración y datos propios.

## Validación

1. Los tres proyectos pueden coexistir sin puertos ni volúmenes compartidos.
2. Caddy alcanza únicamente `127.0.0.1:8081` para `dev-api`.
3. `dev.fasttourney.com` sirve una exportación estática con base API HTTPS.
4. Mailpit no responde desde un hostname público.

## Disparadores de revisión

- envío real de correos, backup/restauración, despliegue automático, varios
  colaboradores o requisitos de disponibilidad.
