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

El registry y la promoción siguen [ADR-0024](../adr/0024-use-ecr-and-digest-based-image-promotion.md):
la API usará ECR privado al abrir la Fase 5; los tags serán inmutables y ECS
desplegará por digest. `dev`, `staging` y `prod` compartirán el mismo artefacto
cuando se promueva, pero no necesariamente configuración, datos ni topología.
`staging` es un entorno de QA, no una rama permanente.

El runtime cloud futuro será Amazon ECS con Fargate conforme a
[ADR-0023](../adr/0023-use-ecs-fargate-as-future-cloud-runtime.md). Esta es una
dirección para la Fase 5, no autorización para crear recursos AWS: hasta entonces
el trabajo y las verificaciones permanecen locales y sin coste cloud.

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
simuladores y herramientas nativas. Esta paridad local no decide aún el despliegue
real del mini PC ni sustituye el runtime cloud futuro.

### Fase 4

Necesidad de Kubernetes, k3d, manifests o empaquetado, recursos, probes,
configuración, secretos y rollout.

### Fase 5

Terraform, cuenta AWS, identidad, bootstrap y estado remoto, red, cómputo,
datos, storage, observabilidad, CI/CD, coste y recuperación.
