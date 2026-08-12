# Despliegue e infraestructura

> Estado: dirección objetivo definida; diseños concretos pendientes de fases.

## Progresión

1. un backend modular como unidad desplegable comprensible;
2. builds independientes del cliente universal para web, iOS y Android;
3. dependencias locales en Docker Compose;
4. servicio instrumentado y recuperable;
5. Kubernetes con k3d;
6. AWS mediante Terraform.

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

AWS se usará de forma efímera para aprender y validar Terraform, ECS/Fargate,
ALB, RDS, red, observabilidad y destrucción controlada, conforme a
[ADR-0088](../adr/0088-use-ephemeral-aws-learning-and-home-runtime.md). El
runtime habitual permanece en el Mac; AWS no se mantiene como producción
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

La topología de entrada y egress inicial sigue ADR-0029: el ALB será público y
terminará HTTPS; las tareas Fargate solo aceptarán tráfico desde él y PostgreSQL
permanecerá privado. No se creará NAT Gateway inicialmente. Antes del primer
`apply` se revisará y autorizará el coste completo del ALB, Fargate, base de
datos, IPv4, logs y transferencia.

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
simuladores y herramientas nativas. Para el runtime doméstico habitual se
validará el target `runtime` con configuración, red y datos separados de `dev`.

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
volumen PostgreSQL y proyecto Compose de dev no se comparten con el entorno
local ni con el futuro prod, conforme a ADR-0091.

ADR-0092 conserva dos despliegues recuperables de dev fuera de Git: cada uno
lleva el SHA, una imagen runtime etiquetada y una exportación web estática. Caddy
sirve el enlace simbólico de la versión activa y el rollback selecciona el SHA
anterior sin tocar PostgreSQL. GitHub Releases y tags no se crean por las
integraciones ordinarias de `develop`; se reservan para producción o hitos
distribuidos. Esto no equivale a backup: la base dev sigue sin dump ni
restauración probada mientras sus datos sean descartables.

Mailpit de dev solo se liga a `127.0.0.1:8026` y no tiene hostname público. Un
SMTP transaccional con el DNS de remitente configurado es requisito antes de
permitir que personas externas dependan de correos de verificación o recuperación.

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

Necesidad de Kubernetes, k3d, manifests o empaquetado, recursos, probes,
configuración, secretos y rollout.

### Fase 5

Terraform, cuenta AWS, identidad, bootstrap y estado remoto, red, cómputo,
datos, storage, observabilidad, CI/CD, coste y recuperación.
