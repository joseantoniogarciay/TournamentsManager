# ADR-0089: Usar `fasttourney.com` y separar los hosts de enlaces de producción y desarrollo

- **Estado:** Aceptado
- **Fecha:** 2026-08-11
- **Decisor:** Usuario, mediante aceptación explícita
- **Propietario del análisis:** Asistente como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

Los enlaces de verificación y recuperación deben abrir la web y, cuando esté
instalada, la variante nativa correcta. El dominio propio debe permanecer estable
aunque cambie la IP doméstica y no debe permitir que la build de desarrollo se
asocie al host de producción.

## Contexto y restricciones

- Las apps usan `com.fasttourney.app` y `com.fasttourney.app.dev` (ADR-0015).
- Los enlaces HTTPS de identidad usan `/link/confirm` y `/link/password-reset`.
- Caddy será el borde HTTPS doméstico (ADR-0087) y DDNS actualizará la IP
  dinámica en Cloudflare.
- El backend exige que fuera de loopback `PUBLIC_BASE_URL` sea HTTPS y coincida
  con el origen de enlaces de la build móvil.

## Criterios de decisión

1. marca y dominio coherentes con los identificadores ya aceptados;
2. aislamiento entre desarrollo y producción;
3. rutas de enlaces simples, HTTPS y sin redirecciones para iOS y Android;
4. separación de la API sin ampliar la superficie de Universal/App Links.

## Alternativas

### Alternativa A — Dominio raíz para producción y subdominio `dev` para desarrollo

- **Ventajas:** URLs breves de producción; aislamiento explícito de builds,
  credenciales web y ficheros de asociación; facilita CORS y configuración por
  entorno.
- **Inconvenientes:** hay que servir los ficheros `.well-known` en dos hosts.
- **Coste de adopción y mantenimiento:** bajo.

### Alternativa B — Un único host para ambas builds

- **Ventajas:** un único conjunto de DNS y ficheros de asociación.
- **Inconvenientes:** mezcla las identidades de desarrollo y producción en el
  mismo límite de confianza.
- **Coste de adopción:** bajo; **mantenimiento:** medio por riesgo de confusión.

### Alternativa C — Subdominios independientes también para producción

- **Ventajas:** separación uniforme.
- **Inconvenientes:** añade una URL pública menos directa sin beneficio actual.
- **Coste de adopción y mantenimiento:** bajo o medio.

## Comparación

La alternativa A mantiene `fasttourney.com` como URL canónica de producto y
limita la asociación de cada host a su aplicación. Es más clara que mezclar
ambas builds y más simple que reservar otro subdominio de producción.

## Recomendación

**Opinión/recomendación:** alternativa A.

## Decisión del usuario

**Aceptada:** el dominio es `fasttourney.com`, registrado y gestionado con
Cloudflare DNS. Los hosts quedan fijados así:

| Finalidad | Host |
| --- | --- |
| web y enlaces de producción | `https://fasttourney.com` |
| web y enlaces de desarrollo | `https://dev.fasttourney.com` |
| API de producción | `https://api.fasttourney.com/v1` |
| API de desarrollo | `https://dev-api.fasttourney.com/v1` |

`www.fasttourney.com` redirigirá a `https://fasttourney.com`. Producción solo
publicará asociaciones para `com.fasttourney.app`; desarrollo solo para
`com.fasttourney.app.dev`. `api.fasttourney.com` no declara asociaciones de app.

La API de desarrollo queda separada de la de producción; no declara asociaciones
de app ni es un host de enlaces profundos. El nombre usa un único nivel de
subdominio para que Cloudflare Universal SSL gratuito pueda cubrirlo; el nombre
inicial `dev.api.fasttourney.com` requería un certificado adicional.

## Consecuencias

### Positivas

- Las asociaciones iOS y Android no cruzan entornos.
- `PUBLIC_BASE_URL` y `EXPO_PUBLIC_APP_LINK_URL` tienen un valor inequívoco por
  entorno público.
- El dominio puede conservarse al pasar entre el Mac y un laboratorio AWS.

### Negativas y deuda aceptada

- Dev y producción requieren DNS, HTTPS y `.well-known` independientes.
- El dominio no valida por sí solo derechos de marca; esa comprobación queda
  fuera del alcance técnico.

## Validación

- Cloudflare resuelve los tres hosts al destino previsto y DDNS actualiza la IP
  doméstica cuando cambie.
- Cada host de enlaces sirve sus dos ficheros `.well-known` con `200`, HTTPS,
  sin redirección ni autenticación.
- iOS valida el `appID` y Android la huella de firma de su variante, sin incluir
  la del otro entorno.
- Los emails de cada entorno contienen su propio origen y los enlaces abren la
  ruta `/link/*` correcta.

## Disparadores de revisión

- cambio de dominio, marca o registrador;
- una tercera variante de app;
- requisitos que justifiquen aislar también web y API de producción en otros
  hosts;
- incidente de asociación, CORS o credenciales entre entornos.

## Documentación afectada

- [infra/app-links](../../infra/app-links/README.md)
- [DEVELOPMENT.md](../engineering/DEVELOPMENT.md)
- [DEPLOYMENT.md](../operations/DEPLOYMENT.md)
- [SECURITY.md](../engineering/SECURITY.md)
- [DECISIONS.md](../governance/DECISIONS.md)
