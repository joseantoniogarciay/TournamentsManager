# ADR-0024: Usar ECR privado y promoción por digest con releases selectivas

- **Estado:** Aceptado
- **Fecha:** 2026-07-25
- **Decisor:** Usuario, mediante aceptación explícita de la alternativa A y del
  flujo de releases selectivas
- **Propietario del análisis:** Asistente como mentor técnico
- **Supera a:** Ninguno; aclara la excepción futura a ADR-0013
- **Superado por:** Ninguno

## Problema

La API se empaquetará como imagen OCI y se ejecutará en ECS con Fargate en la
Fase 5. Hay que decidir dónde se guarda el artefacto y cómo se promueve entre
entornos sin reconstruirlo. El proceso debe permitir que desarrollo continúe
mientras QA valida un subconjunto de funcionalidades seleccionado por negocio.

## Contexto y restricciones

- ADR-0022 acepta una imagen OCI para la API, pero dejó registry, publicación y
  promoción pendientes.
- ADR-0023 fija ECS con Fargate como runtime cloud futuro.
- ADR-0017 fija GitHub Environments protegidos y OIDC para credenciales cloud
  temporales; el repositorio es público.
- ADR-0021 mantiene CI como evidencia informativa y `make verify` local como
  puerta vigente mientras el trabajo sea individual.
- ADR-0013 mantiene hoy `develop` como integración diaria y `main` como estado
  estable. El proyecto actual es individual y didáctico: no necesita todavía
  ramas de feature ni entornos cloud.
- En un equipo, `develop` puede contener funcionalidades A, B y C, mientras
  negocio necesita publicar solo A+B. Desplegar el último estado de `develop`
  a producción obligaría a publicar también C o a bloquear el trabajo.
- DNS y TLS no identifican un entorno por sí mismos. Los futuros subdominios
  `dev.tournamentsmanager.es` y `staging.tournamentsmanager.es` son rutas de
  acceso; cada entorno debe conservar configuración, secretos y datos aislados
  de producción.
- Esta decisión no autoriza crear hoy ECR, AWS, GitHub Environments, dominios,
  workflows de publicación ni ningún recurso facturable.

## Criterios de decisión

1. ejecutar en producción exactamente el artefacto validado por QA;
2. evitar credenciales persistentes de un registry externo en ECS/Fargate;
3. preservar trazabilidad entre commit, imagen, digest, despliegue y rollback;
4. permitir un entorno `dev` que siga avanzando mientras `staging` se estabiliza;
5. seleccionar releases sin adoptar proceso innecesario durante el trabajo
   individual actual;
6. limitar el coste operativo y mantener portabilidad OCI.

## Alternativas

### Alternativa A — ECR privado, tags inmutables y promoción por digest

Publicar la imagen de la API en un único repositorio privado de Amazon ECR.
Cada build recibe un tag trazable `git-<SHA-completo>` y queda identificado por
su digest `sha256`. Los despliegues referencian `repositorio@sha256:...`, nunca
tags mutables como `latest`, `dev`, `staging` o `production`.

Cuando exista equipo o selección de entregas, una rama temporal `release/*`
nace desde `main`. Incorpora solo las ramas de feature seleccionadas y se
despliega en `staging`. Tras la aceptación de QA, `main` y producción reciben
el mismo digest. `develop` sigue desplegando su integración en `dev`.

- **Ventajas:** encaja de forma nativa con ECS/Fargate e IAM; no requiere
  secretos de registry externos; rollback y auditoría usan un digest concreto;
  evita detener integración mientras QA valida; separa publicación de despliegue.
- **Inconvenientes:** ECR vincula la operación inicial a AWS; las releases
  selectivas requieren ramas de feature limpias, orden de integración y
  sincronizar correcciones de release hacia `develop`.
- **Coste de adopción:** medio, aplazado a Fase 5.
- **Coste de mantenimiento:** bajo o medio para un servicio; aumenta si las
  ramas de feature viven demasiado o si los feature flags no se retiran.

### Alternativa B — GHCR privado y promoción por digest

Guardar las imágenes en GitHub Container Registry y hacer que ECS descargue
desde allí.

- **Ventajas:** publicación muy cercana al repositorio y a GitHub Actions.
- **Inconvenientes:** Fargate necesita credenciales del registry externo en
  Secrets Manager y permisos adicionales; añade una frontera GitHub-AWS que ECR
  evita.
- **Coste de adopción y mantenimiento:** medio.

### Alternativa C — Un repositorio ECR por entorno o tags mutables por entorno

Copiar la imagen entre repositorios `dev`, `staging` y `prod`, o reutilizar tags
mutables para señalar el destino actual.

- **Ventajas:** la lectura inicial parece directa.
- **Inconvenientes:** duplicación de políticas y permisos, o pérdida de
  trazabilidad si un tag cambia de contenido. No demuestra que QA y producción
  ejecutaron el mismo artefacto.
- **Coste de mantenimiento:** medio o alto.

### Alternativa D — Promover siempre el último estado de `develop`

Usar un único entorno no productivo y publicar todo `develop` al llegar a
producción.

- **Ventajas:** flujo de ramas mínimo.
- **Inconvenientes:** obliga a publicar funcionalidades no seleccionadas o a
  congelar despliegues y trabajo de integración durante QA.
- **Coste de mantenimiento:** bajo al inicio, alto cuando haya trabajo paralelo.

## Comparación

La alternativa A mantiene una sola copia registral del artefacto, deja que ECS
la descargue con su rol de ejecución y usa el digest como identidad verificable.
La B sirve para otros runtimes, pero añade autenticación externa a Fargate. La C
confunde aislamiento de entornos con copias de artefactos. La D es adecuada para
un proyecto individual, pero no resuelve la selección de entregas en un equipo.

## Recomendación

**Opinión/recomendación:** alternativa A. Es la solución mínima que permite
evolucionar de la práctica individual a un equipo sin convertir cada cambio en
una nueva plataforma de entrega. Las ramas `release/*` son excepcionales y de
vida corta; no se crea una rama permanente `staging`, porque `staging` es un
entorno, no una línea de desarrollo.

## Decisión del usuario

**Aceptada:** usar un repositorio privado de Amazon ECR para la imagen OCI de
la API, con tags inmutables trazables y despliegues por digest. Los entornos
futuros serán `dev`, `staging` y `prod`; sus subdominios se configurarán junto
con la infraestructura de Fase 5.

Mientras el proyecto sea individual y didáctico, se conserva el flujo vigente
de ADR-0013: trabajo en `develop` y promoción de bloques completos a `main`.
Cuando haya equipo, trabajo paralelo o necesidad de seleccionar funcionalidades
por entrega, se usarán ramas de feature y una `release/<versión>` temporal para
componer el lote validado en `staging`. No se mantendrá una rama `staging`
permanente.

## Reglas de implementación

- En Fase 5, Terraform creará un repositorio ECR privado por servicio, con
  inmutabilidad de tags y una política de ciclo de vida. La retención exacta se
  fijará con frecuencia de publicación y coste observados.
- GitHub Actions publicará mediante OIDC y permisos mínimos. Construir y subir
  una imagen es publicación; desplegar un digest existente es promoción.
- Toda imagen publicada llevará el tag inmutable `git-<SHA-completo>` y etiquetas
  OCI de fuente. El digest será la referencia de despliegue y rollback.
- `develop` podrá publicar candidatos en `dev`; `staging` y `prod` usarán solo
  despliegues explícitos aprobados por el GitHub Environment correspondiente.
- Una `release/*` nace desde `main` e integra únicamente ramas de feature
  seleccionadas; no se usa `cherry-pick` como flujo ordinario.
- Las ramas de feature seleccionables deben contener solo su cambio y no absorber
  después el estado completo de `develop`, para no arrastrar funcionalidades no
  seleccionadas a la release.
- Durante QA, `release/*` acepta solo correcciones de estabilización. Al aprobar,
  se integra en `main`; a continuación `main` se sincroniza hacia `develop` antes
  de borrar las ramas de release y de feature ya publicadas.
- Un feature flag puede sustituir una release selectiva cuando el código debe
  integrarse pronto pero su comportamiento aún no debe quedar visible. Los flags
  son temporales y se retiran tras cumplir su objetivo.
- No se crean ahora recursos AWS, dominios, certificados, entornos cloud,
  workflows privilegiados ni costes asociados.

## Consecuencias

### Positivas

- QA y producción ejecutarán la misma imagen, verificable por digest.
- `dev` puede seguir recibiendo integración mientras `staging` conserva un
  candidato estable.
- ECR evita almacenar credenciales de un registry externo para que Fargate
  descargue la imagen.
- Un rollback selecciona un digest previamente desplegado, sin recompilar.

### Negativas y deuda aceptada

- Dev, staging y producción requieren recursos, datos, configuración y
  observabilidad separados cuando se abra Fase 5; el coste relevante no está en
  los subdominios ni en TLS, sino en operar esos recursos.
- La composición selectiva de releases duplica integración y puede introducir
  conflictos entre ramas; se activa solo con un disparador real.
- Firma, SBOM, escaneo reforzado, auto-despliegue, topología DNS/TLS, cuentas,
  red, bases de datos y retención exacta siguen pendientes de decisiones futuras.

## Validación

Al abrir la Fase 5, se demostrará que:

- Terraform crea y destruye ECR y sus políticas de forma controlada;
- un workflow con OIDC publica una imagen sin secretos AWS persistentes;
- ECR rechaza sobrescribir un tag ya publicado;
- ECS/Fargate en staging y producción referencia el mismo digest aprobado;
- se puede volver a un digest anterior documentado;
- `develop` sigue desplegando en dev mientras una release se valida en staging;
- una corrección de release aparece posteriormente en `develop`.

## Disparadores de revisión

- ECR, ECS/Fargate o su coste dejan de encajar con las necesidades reales;
- se necesitan varios servicios, varias cuentas o requisitos de aislamiento que
  justifiquen cambiar la topología del registry;
- el ritmo de entrega permite liberar directamente desde una rama siempre sana
  con feature flags y elimina el valor de `release/*`;
- las ramas de release o feature generan conflictos y demoras recurrentes;
- se exige firma, SBOM, procedencia verificable, escaneo o compliance adicional.

## Documentación afectada

- [ADR-0013](0013-use-develop-as-integration-branch.md)
- [TECHNICAL_BASELINE.md](../governance/TECHNICAL_BASELINE.md)
- [DECISIONS.md](../governance/DECISIONS.md)
- [DECISIONS_TO_REVISIT.md](../governance/DECISIONS_TO_REVISIT.md)
- [DEPLOYMENT.md](../operations/DEPLOYMENT.md)
- [LEARNING.md](../project/LEARNING.md)
- [CONTRIBUTING.md](../../CONTRIBUTING.md)
- [README.md](../../README.md)
- [CHANGELOG.md](../../CHANGELOG.md)

## Fuentes técnicas

- [Amazon ECR con Amazon ECS](https://docs.aws.amazon.com/AmazonECR/latest/userguide/ECR_on_ECS.html)
- [Inmutabilidad de tags en ECR](https://docs.aws.amazon.com/AmazonECR/latest/userguide/image-tag-mutability.html)
- [Políticas de ciclo de vida de ECR](https://docs.aws.amazon.com/AmazonECR/latest/userguide/LifecyclePolicies.html)
- [Rol de ejecución de ECS](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task_execution_IAM_role.html)
- [GitHub Environments](https://docs.github.com/en/actions/concepts/workflows-and-actions/deployment-environments)
- [Ramas de release](https://trunkbaseddevelopment.com/branch-for-release/)
