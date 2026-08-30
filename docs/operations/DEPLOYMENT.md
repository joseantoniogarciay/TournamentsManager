# Despliegue e infraestructura

> Estado: dirección objetivo definida; diseños concretos pendientes de fases.

## Progresión

1. un backend modular como unidad desplegable comprensible;
2. builds independientes del cliente universal para web, iOS y Android;
3. dependencias locales en Docker Compose;
4. servicio instrumentado y recuperable;
5. Kubernetes con K3s en una VM Linux local;
6. laboratorio EKS efímero en AWS mediante Terraform.

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

## Continuidad futura de producción

**Pendiente de diseño concreto antes de implementar.** Para evitar que un cambio
de versión interrumpa las peticiones, producción evaluará un despliegue
blue/green detrás del ingress K3s y Caddy: la instancia nueva arranca separada,
supera health checks y una validación funcional mínima, se conmuta el tráfico y
la instancia anterior permanece disponible durante un periodo breve de
observación. Un rollback posterior vuelve a arrancar y validar el SHA anterior
antes de conmutar otra vez. ADR-0111 fija la VM K3s como runtime, pero no decide
todavía este mecanismo.

Durante la conmutación ambas versiones deben ser compatibles con el mismo
esquema PostgreSQL. Los cambios destructivos o incompatibles exigirán una
estrategia explícita de migración (por ejemplo, expand/contract, forward-fix o
restauración), no solo volver a una imagen anterior.

## Límite de despliegue desde GitHub

El repositorio será público, pero el acceso operativo no. CI construirá y
verificará en infraestructura aislada. Los jobs de producción necesitarán un
environment protegido y una identidad temporal o dedicada de mínimo privilegio.

No se conectará el repositorio público a un runner permanente dentro del VPS. El
modelo concreto —push controlado desde runner alojado o pull de artefactos desde
el servidor— se decidirá antes del primer despliegue.

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

La web inicial se entregará como aplicación client-side conforme a
[ADR-0016](../adr/0016-use-client-side-web-rendering-initially.md). Static
rendering, SSR o una superficie web pública especializada se decidirán solo si
aparecen requisitos públicos de indexación, previews o rendimiento.

La API se empaquetará como imagen OCI conforme a
[ADR-0022](../adr/0022-package-backend-as-oci-image.md). Esta imagen solo
contendrá el backend y mantendrá build y runtime separados. No decide todavía
la firma, el SBOM ni el escaneo reforzado.

Cuando un laboratorio AWS lo necesite, registry y promoción seguirán
[ADR-0024](../adr/0024-use-ecr-and-digest-based-image-promotion.md): ECR privado,
tags inmutables y despliegue por digest. El Mac no necesita imitar un `staging`
permanente: desarrollo y release doméstico separan datos y configuración, y AWS
se destruye al acabar cada práctica.

AWS se usará de forma efímera para aprender y validar Terraform, EKS, red,
observabilidad y destrucción controlada, conforme a
[ADR-0088](../adr/0088-use-ephemeral-aws-learning-and-home-runtime.md) y
[ADR-0111](../adr/0111-use-k3s-vm-for-home-production-runtime.md). El runtime
doméstico de `prod` será K3s en la VM del Mac; AWS no se mantiene como producción
permanente sin una nueva decisión basada en usuarios, disponibilidad o coste.

La infraestructura de esa fase se describirá con Terraform conforme a
[ADR-0025](../adr/0025-use-terraform-for-infrastructure-as-code.md). Esta
elección no fija todavía una cuenta, backend de estado, bloqueo, región, red ni
proveedor configurado; esas decisiones se tomarán antes de cualquier `apply`.

La fundación AWS seguirá [ADR-0026](../adr/0026-use-aws-organizations-and-temporary-identities.md):
`management` centraliza gobierno y facturación sin cargas; `nonprod` alojará
los futuros `dev` y `staging`, y `prod` producción. El acceso humano y de
automatización será temporal; el backend de estado sigue pendiente.

Conforme a [ADR-0027](../adr/0027-keep-local-state-until-first-cloud-apply.md)
y [ADR-0028](../adr/0028-use-hcp-terraform-free-for-remote-state.md), el estado
es local solo mientras no haya infraestructura AWS real y HCP Terraform Free
será su backend remoto inicial. Los runs permanecerán inicialmente en la CLI
local; no hay auto-apply ni recursos AWS autorizados hasta abrir la Fase 5 y
verificar bloqueo, recuperación y acceso. Git no se usará para almacenar estado.

La topología de entrada y egress de EKS se decidirá antes del primer
laboratorio cloud. ADR-0029 queda superado parcialmente porque su diseño partía
de tareas Fargate. Antes del primer `apply` se revisará y autorizará el coste
completo de cómputo, red, datos, IPv4, logs y transferencia.

ADR-0030 fija la región futura en España (`eu-south-2`), una VPC `/16` no
solapada por cuenta y dos AZ con dos subredes públicas y dos privadas. Este mapa
no genera coste: el gasto seguirá bloqueado hasta presentar una estimación
completa y recibir autorización explícita del usuario.

## Decisiones por fase

### Desarrollo local

Conforme a [ADR-0076](../adr/0076-run-the-local-api-in-compose-with-air.md),
Compose ejecuta API, PostgreSQL y Mailpit. La API selecciona el target `dev` del
Dockerfile y usa Air con el código montado; el target `runtime` queda reservado
para validar el artefacto sin compilador ni Air. Expo se ejecuta en host por sus
simuladores y herramientas nativas. La imagen `runtime` se desplegará en K3s
para `prod`, con configuración, red y datos separados de `dev`.

Una beta doméstica usa Cloudflare Tunnel como entrada HTTPS pública (ADR-0090).
El conector del Mac inicia una conexión saliente y alcanza Caddy solo por
loopback; no se reenvían puertos en UniFi ni se publica PostgreSQL o el puerto
interno de la API.

ADR-0089 fija `fasttourney.com` para producción, `dev.fasttourney.com` para
desarrollo, `api.fasttourney.com` para la API de producción y
`dev-api.fasttourney.com` para la API de desarrollo. Caddy servirá los ficheros
`.well-known` correctos en los dos primeros hosts; los hosts de API no participan
en Universal Links ni App Links.

El borde versionado vive en [`infra/home/Caddyfile`](../../infra/home/Caddyfile).
`tournaments-manager-dev` publica la API runtime solo por `127.0.0.1:8081` y
la web exportada de Expo se sirve estática en `dev.fasttourney.com`; Caddy
conserva `503` para los hosts de producción no publicados. La configuración,
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
distribuidos. Esto no equivale a backup. ADR-0108 añade para `dev` un repositorio
pgBackRest cifrado, copia base, incrementales y WAL archivado con restauración
aislada; véase el [runbook de backup PostgreSQL](../runbooks/postgresql-backup-dev.md).
`prod` ya tiene PostgreSQL con volumen y repositorio propios, pgBackRest cifrado,
WAL archivado, una primera completa y restauración aislada local verificada.
Quedan por automatizar la completa semanal, los incrementales diarios y la
réplica verificable fuera de la VM. Conforme a ADR-0114, el Mac inicia por SSH
una copia del repositorio cifrado de
la VM hacia su ubicación doméstica sincronizada; no se comparte una carpeta UTM
ni se entrega una clave privada del Mac a Kubernetes. Esa ubicación sigue
compartiendo Mac, cuenta y proveedor: no equivale a independencia ante su
pérdida.

Mailpit pertenece solo al entorno local y no tiene hostname público. El entorno
`dev` usa Resend por SMTP autenticado con STARTTLS; antes de invitar personas se
verifica `mail.fasttourney.com` y sus registros SPF, DKIM y DMARC. La clave de
solo envío vive fuera de Git en `infra/dev/api.docker.env`; véase ADR-0093.

El mismo proyecto `dev` mantiene Prometheus, Alertmanager, Loki, Tempo, Promtail
y Grafana en red y volúmenes propios. Alertmanager usa una segunda clave Resend
de solo envío, montada desde un secreto local, para no compartir el radio de
revocación del correo transaccional. Las interfaces operativas se publican solo
en loopback y no forman parte de Cloudflare Tunnel; véase ADR-0100.

Antes de publicar el futuro artefacto web de producción, su script de exportación
debe declarar, sin leer el `.env` local y limpiando la caché de Metro:

```sh
EXPO_PUBLIC_API_BASE_URL=https://api.fasttourney.com/v1
EXPO_PUBLIC_APP_LINK_URL=https://fasttourney.com
```

La primera URL evita que la web pública contacte servicios locales; la segunda
hace que los enlaces de liga compartidos apunten al dominio público. La
publicación aún no está autorizada: `fasttourney.com` y `api.fasttourney.com`
permanecen deliberadamente en `503`.

### Fase 4

VM Linux de un nodo con K3s como runtime doméstico de `prod`: manifests,
empaquetado, recursos, probes, configuración, secretos, persistencia, backup,
ingress, rollout y recuperación, conforme a ADR-0111.

ADR-0112 concreta el orden de empaquetado: API, PostgreSQL y los recursos de
core se definen primero como manifiestos YAML propios aplicados con `kubectl`.
La observabilidad de terceros se añade después mediante Helm, con charts y
valores versionados; no se instala inicialmente un operador ni un chart propio
de toda la plataforma.

### Fase 5

Terraform, cuenta AWS, identidad, bootstrap y estado remoto, topología EKS,
datos, storage, observabilidad, CI/CD, coste y recuperación.
