# Despliegue e infraestructura

> Estado: dirección objetivo definida; diseños concretos pendientes de fases.

## Progresión

1. proceso local comprensible;
2. dependencias en Docker Compose;
3. servicio instrumentado y recuperable;
4. Kubernetes con k3d;
5. AWS mediante Terraform.

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

## Decisiones por fase

### Fase 1

Docker Compose, configuración, datos, ciclo de vida y comandos.

### Fase 4

Necesidad de Kubernetes, k3d, manifests o empaquetado, recursos, probes,
configuración, secretos y rollout.

### Fase 5

Cuenta AWS, identidad, red, cómputo, datos, storage, observabilidad, estado de
Terraform, CI/CD, coste y recuperación.
