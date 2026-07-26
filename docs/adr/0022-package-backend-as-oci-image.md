# ADR-0022: Empaquetar la API como imagen OCI

- **Estado:** Aceptado
- **Fecha:** 2026-07-25
- **Decisor:** Usuario, mediante aceptación explícita de la alternativa A
- **Propietario del análisis:** Asistente como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

El monolito modular de Go necesitará un artefacto repetible para ejecutarse
fuera de la estación de desarrollo. Hay que decidir cómo empaquetar la API sin
adelantar el runtime cloud, el registry ni el mecanismo de promoción.

## Contexto y restricciones

- El manifiesto define Docker y Docker Compose como dirección, pero no autoriza
  por sí solo Dockerfiles ni una plataforma concreta.
- ADR-0007 establece una única unidad de despliegue inicial: el backend.
- ADR-0018 conserva API y cliente en el host durante desarrollo; Compose opera
  únicamente dependencias locales. Esta decisión no lo modifica.
- ADR-0016 entrega la web inicial como artefacto client-side. Los paquetes iOS y
  Android tienen sus propios canales; no forman parte de esta imagen.
- ADR-0017 exige configuración externa y prohíbe secretos en Git, imágenes,
  logs y bundles.
- No existe todavía un vertical slice ni una API ejecutable. Por tanto, no se
  crean Dockerfiles, registros, workflows ni despliegues al aceptar este ADR.

## Criterios de decisión

1. producir el mismo artefacto para pruebas y futuros entornos;
2. mantener la portabilidad entre runtimes compatibles con OCI;
3. limitar el alcance a la API, sin contenerizar el cliente ni PostgreSQL;
4. conservar un bucle de desarrollo local rápido;
5. minimizar superficie de ataque, dependencias y deriva operativa;
6. permitir decidir después runtime, registry y promoción con evidencia.

## Alternativas

### Alternativa A — Imagen OCI/Docker para la API

Empaquetar la API Go como imagen compatible con OCI. Cuando exista código, la
imagen separará compilación y runtime mediante etapas de build; el resultado
incluirá solo lo necesario para ejecutar la API.

- **Ventajas:** artefacto reproducible y portable; separa build de ejecución;
  prepara health checks, despliegue y rollback sin elegir proveedor.
- **Inconvenientes:** requiere mantener Dockerfile, contexto de build, imágenes
  base y validación de seguridad.
- **Coste de adopción:** bajo o medio al crear la primera API.
- **Coste de mantenimiento:** bajo; sube si se multiplican imágenes o targets.
- **Riesgos:** confundir la imagen con una decisión de AWS, Kubernetes o Compose
  de desarrollo. Se mitiga delimitando explícitamente el alcance.

### Alternativa B — Desplegar un binario Go directamente

Compilar la API y copiar el binario al host de ejecución.

- **Ventajas:** menos ficheros y conceptos iniciales.
- **Inconvenientes:** el host asume más responsabilidades de runtime y aumenta
  la deriva entre entornos; no sigue la dirección Docker del manifiesto.
- **Coste de adopción:** bajo.
- **Coste de mantenimiento:** medio por configuración y actualizaciones del host.
- **Riesgos:** diferencias difíciles de reproducir entre la estación, CI y el
  servidor.

### Alternativa C — Buildpacks

Delegar la creación de la imagen a buildpacks, sin un Dockerfile propio.

- **Ventajas:** menos configuración inicial y convenciones incorporadas.
- **Inconvenientes:** oculta decisiones de empaquetado relevantes para el
  aprendizaje y añade una capa de tooling antes de necesitarla.
- **Coste de adopción:** medio.
- **Coste de mantenimiento:** medio por la plataforma de buildpacks y sus
  actualizaciones.
- **Riesgos:** dificultad para explicar o ajustar el contenido final de la
  imagen cuando aparezcan requisitos concretos.

### No cambiar

Continuar sin un artefacto de API definido.

- **Consecuencias:** se retrasan la paridad de artefacto y cualquier discusión
  verificable sobre runtime o promoción.

## Comparación

La A crea el límite operativo mínimo que necesita el monolito sin imponer su
plataforma de ejecución. La B es más simple a corto plazo, pero desplaza la
complejidad al host. La C automatiza decisiones que el proyecto quiere aprender
antes de abstraer. Las imágenes multi-stage permiten dejar herramientas de
compilación fuera del runtime final; los digests identifican de forma inmutable
el contenido que después podrá promoverse.

## Recomendación

**Opinión/recomendación:** alternativa A. Es la solución mínima suficiente:
una imagen OCI para la API, sin aplicar todavía una topología de producción ni
ampliar Docker Compose.

## Decisión del usuario

**Aceptada:** alternativa A. La API se empaquetará como imagen OCI/Docker cuando
exista una API ejecutable.

## Reglas de implementación

- La imagen abarcará solo la API Go; PostgreSQL, web, iOS y Android no se
  incorporan en ella.
- El Dockerfile futuro usará etapas separadas de build y runtime, base fijada y
  un contexto de build mínimo mediante `.dockerignore`.
- El runtime no incluirá el compilador, código fuente, secretos ni herramientas
  de desarrollo que no sean necesarias para ejecutar la API.
- El proceso se ejecutará con un usuario no privilegiado cuando la imagen y sus
  dependencias lo permitan.
- La configuración llegará por el contrato de entorno de ADR-0017; nunca se
  horneará en la imagen.
- Se validará la imagen mediante build, arranque y pruebas proporcionales cuando
  exista una ruta HTTP y health/readiness reales.
- Runtime, registry, versionado, firma, publicación y promoción siguen
  pendientes de decisiones posteriores.

## Consecuencias

### Positivas

- La API tendrá un artefacto consistente entre CI y despliegue futuro.
- El empaquetado no queda acoplado a AWS, Kubernetes ni a un VPS.
- La decisión conserva el desarrollo nativo de Go y Expo definido en ADR-0018.

### Negativas y deuda aceptada

- Habrá una pieza operativa más que actualizar y verificar.
- Aún no existe un entorno que ejecute ni publique la imagen.
- El cliente conserva estrategias de entrega separadas.

## Validación

Cuando exista la API, se demostrará que:

- la imagen se construye desde un clon limpio;
- inicia con configuración externa válida y falla de forma accionable ante una
  configuración o dependencia inválida;
- no contiene secretos ni herramientas de compilación en el runtime final;
- ejecuta con privilegios mínimos y expone señales de salud acordadas;
- el artefacto construido puede identificarse de forma inmutable antes de su
  futura promoción.

## Disparadores de revisión

- varios servicios o arquitecturas requieren una estrategia común de imágenes;
- un runtime objetivo no admite o no beneficia de imágenes OCI;
- el tamaño, tiempo de build o vulnerabilidades de la imagen exceden un
  presupuesto acordado;
- un requisito de seguridad exige firma, SBOM o controles adicionales;
- evidencia de entrega demuestra que el cliente web necesita un empaquetado
  distinto al artefacto estático previsto.

## Documentación afectada

- [TECHNICAL_BASELINE.md](../governance/TECHNICAL_BASELINE.md)
- [DEPLOYMENT.md](../operations/DEPLOYMENT.md)
- [DECISIONS.md](../governance/DECISIONS.md)
- [DECISIONS_TO_REVISIT.md](../governance/DECISIONS_TO_REVISIT.md)
- [LEARNING.md](../project/LEARNING.md)
- [README.md](../../README.md)
- [CHANGELOG.md](../../CHANGELOG.md)

## Fuentes técnicas

- [Docker: buenas prácticas de build](https://docs.docker.com/build/building/best-practices/)
- [Docker: multi-stage builds](https://docs.docker.com/build/building/multi-stage/)
- [Docker: image digests](https://docs.docker.com/dhi/explore/security-concepts/digests/)
