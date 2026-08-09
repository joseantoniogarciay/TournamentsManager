# ADR-0076: Ejecutar la API local en Compose con Air

- **Estado:** Aceptado
- **Fecha:** 2026-08-08
- **Decisor:** Usuario, mediante decisión explícita
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** ADR-0018
- **Superado por:** Ninguno

## Problema

El desarrollo diario debía ejercitar la misma topología Docker que se usará para
probar el artefacto de API en un futuro mini PC, sin perder la recompilación al
guardar código Go.

## Contexto y restricciones

- ADR-0022 ya acepta una imagen OCI de la API con build y runtime separados.
- ADR-0012 exige herramientas Go versionadas y prohíbe `@latest` en automatización.
- PostgreSQL y Mailpit ya son dependencias locales de Compose; el cliente Expo
  permanece en host por sus herramientas y simuladores nativos.
- Un entorno de desarrollo con Air no es una imagen de producción: monta código,
  contiene compilador y reinicia procesos.
- El mini PC no autoriza por sí solo una estrategia de producción, TLS, backups,
  secretos ni sustituye el runtime cloud futuro de ADR-0023.

## Criterios de decisión

1. compartir la red, nombres de servicios y contratos relevantes con la ejecución empaquetada;
2. recompilar la API automáticamente al editar Go;
3. no duplicar Dockerfiles ni incluir Air en runtime;
4. mantener configuración externa, datos persistentes y health checks;
5. conservar diagnóstico y recuperación explícitos.

## Alternativas

### Alternativa A — API Go en host y dependencias en Compose

- **Ventajas:** menor fricción de edición y depuración nativa.
- **Inconvenientes:** no practica la red ni el ciclo de la API contenida.
- **Coste de adopción y mantenimiento:** bajo.
- **Riesgo:** la imagen se valida tarde y los contratos `localhost` divergen.

### Alternativa B — Compose local con etapa `dev` y Air

- **Ventajas:** API, PostgreSQL y Mailpit comparten red Compose; la edición activa
  Air, y un mismo Dockerfile produce también el runtime mínimo.
- **Inconvenientes:** bind mounts, caché Go y configuración interna añaden piezas
  de diagnóstico; el bucle no es idéntico al runtime.
- **Coste de adopción:** medio.
- **Coste de mantenimiento:** bajo o medio, limitado a un Dockerfile, dos
  contratos de entorno y Compose local.
- **Riesgo:** confundir Air con producción. Se mitiga con targets separados.

### Alternativa C — Compose Watch sin Air

- **Ventajas:** usa capacidad nativa reciente de Docker Compose.
- **Inconvenientes:** sigue requiriendo definir la reconstrucción/reinicio de Go
  y no aporta una herramienta Go versionada para el ciclo de la API.
- **Coste de adopción y mantenimiento:** medio.

### No cambiar

- **Consecuencia:** se conserva la diferencia entre la API local en host y el
  artefacto que se quiere aprender a operar en Docker.

## Comparación

La alternativa B mantiene una única fuente para los targets `dev` y `runtime`:
la primera incluye Air y un volumen de código; la segunda copia solo el binario
estático y certificados. Por tanto iguala los contratos operativos útiles, sin
afirmar una igualdad falsa entre desarrollo y producción.

## Recomendación

**Opinión/recomendación:** alternativa B; era la mínima suficiente para el
objetivo de practicar Docker antes del mini PC.

## Decisión del usuario

**Aceptada:** alternativa B. Compose local ejecutará API, PostgreSQL y Mailpit;
Air se versiona como herramienta Go y solo se usa en la etapa `dev`. El cliente
Expo sigue en host. La etapa `runtime` no contiene Air, fuentes ni compilador.

## Consecuencias

### Positivas

- `make dev-up` entrega un bucle Docker reproducible con recarga automática.
- La API interna se conecta por `postgres` y `mailpit`, no por loopback.
- `make api-image-build` verifica el artefacto de runtime mínimo.

### Negativas y deuda aceptada

- Se mantienen dos contratos de API local: host y Compose, porque sus redes son
  distintas y expresarlo evita usar accidentalmente `127.0.0.1` dentro de Docker.
- La instalación de Air amplía el módulo de herramientas y se revisa como cualquier
  actualización versionada.
- La ejecución real del mini PC, TLS, backups, entrega y correo saliente requieren
  una decisión y runbook posteriores.

## Validación

- `make dev-init`, `make dev-up` y `GET /healthz` funcionan en un clon limpio.
- Una edición Go provoca una recompilación de Air sin recrear PostgreSQL.
- PostgreSQL y Mailpit llegan a `healthy` antes de iniciar la API.
- `make api-image-build` crea la etapa `runtime`; su historial no contiene la
  etapa `dev` y el proceso corre como UID no privilegiado.
- `make verify` y `docker compose ... config` terminan correctamente.

## Disparadores de revisión

- La recarga por bind mount es lenta o pierde eventos de manera reproducible.
- El tamaño, vulnerabilidades o tiempo de build exceden un presupuesto acordado.
- Se decide el despliegue real del mini PC, incluyendo TLS, correo, backups y
  arquitectura `amd64`/`arm64`.
- Kubernetes o el runtime cloud cambia materialmente la topología local.

## Documentación afectada

- [DEVELOPMENT.md](../engineering/DEVELOPMENT.md)
- [DEPLOYMENT.md](../operations/DEPLOYMENT.md)
- [Runbook local](../runbooks/local-postgresql.md)
- [DECISIONS.md](../governance/DECISIONS.md)
- [DECISIONS_TO_REVISIT.md](../governance/DECISIONS_TO_REVISIT.md)
- [LEARNING.md](../project/LEARNING.md)
- [CHANGELOG.md](../../CHANGELOG.md)

## Fuentes técnicas

- [Air](https://github.com/air-verse/air)
- [Docker multi-stage builds](https://docs.docker.com/build/building/multi-stage/)
- [Docker Compose Watch](https://docs.docker.com/compose/how-tos/file-watch/)
