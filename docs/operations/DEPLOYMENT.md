# Despliegue e infraestructura

> Estado: Fase 4 completada en K3s doméstico; Fase 5 AWS cancelada por ADR-0128.

## Progresión

1. un backend modular como unidad desplegable comprensible;
2. builds independientes del cliente universal para web, iOS y Android;
3. dependencias locales en Docker Compose;
4. servicio instrumentado y recuperable;
5. Kubernetes con K3s en una VM Linux local;
6. sin laboratorio cloud previsto; la Fase AWS fue cancelada por ADR-0128.

Cada salto debe justificar qué capacidad añade y qué coste introduce.

## Paridad local-producción

“Parecerse” significa compartir contratos importantes:

- configuración externa;
- imágenes y artefactos equivalentes;
- migraciones;
- health/readiness semantics;
- límites y timeouts;
- señales observables;
- comportamiento ante dependencias no disponibles.

No significa ejecutar localmente todos los servicios gestionados ni reproducir la
topología completa.

## Reglas de infraestructura

- infraestructura reproducible y revisable;
- cambios pequeños con plan de verificación y rollback;
- secretos fuera de estado, planes y logs;
- recursos etiquetados, con propietario y coste visible;
- estado y locking de Terraform decididos antes del trabajo en equipo;
- ningún despliegue sin health checks y criterio de éxito;
- ningún backup sin restauración probada.

## Continuidad de producción

v1 usa el rollout y rollback verificado de K3s; no adopta blue/green. Se
evaluaría una instancia paralela detrás del Ingress y Caddy solo si aparece un
requisito de disponibilidad que el rollout actual no cumpla. ADR-0111 fija la
VM K3s como runtime y la retrospectiva de Fase 4 registra este límite como deuda
deliberada, no como trabajo pendiente.

Durante la conmutación ambas versiones deben ser compatibles con el mismo
esquema PostgreSQL. Los cambios destructivos o incompatibles exigirán una
estrategia explícita de migración (por ejemplo, expand/contract, forward-fix o
restauración), no solo volver a una imagen anterior.

## Límite de despliegue desde GitHub

El repositorio es público, pero el acceso operativo no. CI verifica en runners
alojados sin acceso al runtime doméstico. Los despliegues de `dev` y `prod` son
operaciones locales explícitas desde identidades dedicadas; no existe un runner
self-hosted permanente conectado al repositorio público.

La configuración y los secretos de despliegue siguen
[ADR-0017](../adr/0017-use-env-contracts-github-environments-and-oidc.md):
GitHub Environment protegido, secretos mínimos por entorno y OIDC para cloud
cuando esté disponible. Las credenciales persistentes se consideran una excepción
que debe documentarse.

## Límites de entrega del cliente universal

Un árbol de código compartido no implica un único artefacto ni una entrega
acoplada:

- web produce un artefacto desplegable en hosting web;
- iOS produce una aplicación firmada y distribuida por su canal;
- Android produce una aplicación firmada y distribuida por su canal;
- cada target puede ejecutar, publicar y revertir su pipeline de forma
  independiente;
- la compatibilidad con el contrato API debe verificarse antes de publicar,
  teniendo en cuenta que una aplicación instalada no se actualiza de inmediato.

La web se entrega con exportación estática de Expo, sin SSR general. La home `/`
es indexable y el artefacto contiene su HTML, `robots.txt`, sitemap y metadatos;
el artefacto publica además `/favicon.ico` y `/apple-touch-icon.png`; la home
enlaza ambos para navegador, buscadores y Safari. El borde deja rastreables esos
iconos y añade
`X-Robots-Tag: noindex, nofollow, noarchive` a las rutas de aplicación. Solo
`/league/{uuid}` se deriva a un proceso Go local que inyecta
metadatos sociales desde la API pública en el mismo shell estático. Más
superficies dinámicas o SSR completo se decidirán solo si Search Console,
previews o rendimiento aportan evidencia (ADR-0121 y ADR-0120).

La API se empaqueta como imagen OCI conforme a
[ADR-0022](../adr/0022-package-backend-as-oci-image.md). Esta imagen solo
contendrá el backend y mantendrá build y runtime separados. No decide todavía
la firma, el SBOM ni el escaneo reforzado.

## Referencia cloud histórica

ADR-0024 a ADR-0030, ADR-0088 y ADR-0101 conservan el análisis previo de ECR,
Terraform, IAM, red y laboratorios AWS/EKS. ADR-0128 cancela su ejecución en
este proyecto: no hay cuenta, backend remoto, `apply`, laboratorio ni gasto
cloud previsto. Solo una necesidad futura explícita reabriría ese análisis.

## Decisiones por fase

### Desarrollo local

Conforme a [ADR-0076](../adr/0076-run-the-local-api-in-compose-with-air.md),
Compose ejecuta API, PostgreSQL y Mailpit. La API selecciona el target `dev` del
Dockerfile y usa Air con el código montado; el target `runtime` valida el
artefacto sin compilador ni Air y es también la imagen desplegada en K3s para
`prod`. Expo se ejecuta en host por sus simuladores y herramientas nativas.

Una beta doméstica usa Cloudflare Tunnel como entrada HTTPS pública (ADR-0090).
El conector del Mac inicia una conexión saliente y alcanza Caddy solo por
loopback; no se reenvían puertos en UniFi ni se publica PostgreSQL o el puerto
interno de la API.

ADR-0089 fija `fasttourney.com` para producción, `dev.fasttourney.com` para
desarrollo, `api.fasttourney.com` para la API de producción y
`dev-api.fasttourney.com` para la API de desarrollo. Caddy sirve los ficheros
`.well-known` correctos en los dos primeros hosts; los hosts de API no participan
en Universal Links ni App Links.

El borde versionado vive en [`infra/home/Caddyfile`](../../infra/home/Caddyfile).
`tournaments-manager-dev` publica la API runtime solo por `127.0.0.1:8081` y
la web exportada de Expo se sirve estática en `dev.fasttourney.com`. Desde el
2026-09-05, `fasttourney.com` sirve el release web de producción; el cambio
revisable entre apertura y cierre sigue siendo importar `production_web` o
restaurar el `503` en Caddy. La configuración,
volumen PostgreSQL y proyecto Compose de dev no se comparten con `local` ni con
el namespace `prod` de K3s, conforme a ADR-0091 y ADR-0111.

El mismo release contiene la referencia pública de desarrollo en
`https://dev.fasttourney.com/api-docs/`. El script de despliegue copia la UI
Scalar y `openapi.yaml` junto a la exportación web; Caddy atiende esa ruta antes
del fallback de la SPA. Como usa el origen ya autorizado
`dev.fasttourney.com`, no cambia CORS de `dev-api` ni exige otro hostname.

ADR-0092 conserva dos despliegues recuperables de dev fuera de Git: cada uno
lleva el SHA, una imagen runtime etiquetada y una exportación web estática. Caddy
sirve el enlace simbólico de la versión activa y el rollback selecciona el SHA
anterior sin tocar PostgreSQL. GitHub Releases y tags no se crean por las
integraciones ordinarias de `develop`; se reservan para producción o hitos
distribuidos. ADR-0119 concreta que un hito usa un tag SemVer anotado sobre el
merge de `main`, una GitHub Release y un artefacto activo del mismo SHA. Esto no
equivale a backup. ADR-0108 añade para `dev` un repositorio
pgBackRest cifrado, copia base, incrementales y WAL archivado con restauración
aislada; véase el [runbook de backup PostgreSQL](../runbooks/postgresql-backup-dev.md).
`prod` tiene PostgreSQL con volumen y repositorio propios, pgBackRest cifrado,
WAL archivado, completa semanal, incrementales diarios y restauración aislada
verificada. Conforme a ADR-0114, el Mac inicia por SSH la réplica cifrada desde
la VM hacia su ubicación doméstica sincronizada; la ejecución programada y la
restauración desde esa réplica se demostraron el 2026-09-18. No se comparte una
carpeta UTM ni se entrega una clave privada del Mac a Kubernetes. Esa ubicación
sigue compartiendo Mac, cuenta y proveedor: no equivale a independencia ante su
pérdida.

Mailpit pertenece solo al entorno local y no tiene hostname público. El entorno
`dev` usa Resend por SMTP autenticado con STARTTLS; antes de invitar personas se
verifica `mail.fasttourney.com` y sus registros SPF, DKIM y DMARC. La clave de
solo envío vive fuera de Git en `infra/dev/api.docker.env`; véase ADR-0093.
`SUGGESTION_RECIPIENT` selecciona por entorno el buzón que recibe los avisos de
sugerencias privadas; no es un secreto ni se fija dentro del binario.

El mismo proyecto `dev` mantiene Prometheus, Alertmanager, Loki, Tempo, Promtail
y Grafana en red y volúmenes propios. Alertmanager usa una segunda clave Resend
de solo envío, montada desde un secreto local, para no compartir el radio de
revocación del correo transaccional. Las interfaces operativas se publican solo
en loopback y no forman parte de Cloudflare Tunnel; véase ADR-0100.

El script del artefacto web de producción declara, sin leer el `.env` local y
limpiando la caché de Metro:

```sh
EXPO_PUBLIC_API_BASE_URL=https://api.fasttourney.com/v1
EXPO_PUBLIC_APP_LINK_URL=https://fasttourney.com
```

La primera URL evita que la web pública contacte servicios locales; la segunda
hace que los enlaces de liga compartidos apunten al dominio público. La API y
la web de producción se abrieron el 2026-09-05 tras validar TLS, CORS y el
release; no supone autorizar los clientes ni asociaciones móviles pendientes.

Al activar el canal de apoyo de [ADR-0129](../adr/0129-accept-voluntary-developer-tips-with-platform-appropriate-payments.md),
`infra/home/secrets/production-web.env` añade los tres Payment Links públicos
de Stripe para 2 €, 5 € y 10 €. El script ya los propaga a la exportación web;
si alguno falta o no es HTTPS, la interfaz no ofrece propinas. No se añaden a
la configuración de builds nativas.

La preparación concreta de la web separa construir de activar:
`infra/home/stage-prod-web.sh` crea un release estático inmutable con SHA,
asociaciones móviles reales y configuración OAuth de producción externa a Git;
`activate-prod-web.sh` conmuta solo su enlace simbólico y
`rollback-prod-web.sh` vuelve a un SHA existente. El bloque Caddy
`production_web` conserva CSP, SPA y `/.well-known`; activarlo o volver al `503`
requiere una operación explícita y verificable. El
[runbook de publicación](../runbooks/production-web-publication.md) exige TLS
válido, Tunnel sano, CORS, correo, Google y recorridos controlados antes de esa
conmutación.

### Fase 4

VM Linux de un nodo con K3s como runtime doméstico de `prod`: manifests,
empaquetado, recursos, probes, configuración, secretos, persistencia, backup,
ingress, rollout y recuperación, conforme a ADR-0111.

ADR-0117 separa la administración de la VM de la identidad humana inicial:
`fasttourney-operator` entra por clave SSH dedicada en la red privada y recibe
`sudo` no interactivo para operar host, K3s y `kube-system`. La contraseña de
Ubuntu no se almacena en el Mac ni en Git; el bootstrap único permanece
interactivo. Véase el runbook de administración remota antes de depender de
operaciones no asistidas.

ADR-0112 concreta el orden de empaquetado: API, PostgreSQL y los recursos de
core se definen primero como manifiestos YAML propios aplicados con `kubectl`.
La observabilidad de terceros se añade después mediante Helm, con charts y
valores versionados; no se instala inicialmente un operador ni un chart propio
de toda la plataforma.

### Fase 5 — cancelada

Terraform, cuenta AWS, identidad, bootstrap y laboratorio EKS no se activarán
en este proyecto. Este material se conserva como referencia histórica y solo se
reabrirá mediante la decisión explícita y el análisis de coste de ADR-0128.
