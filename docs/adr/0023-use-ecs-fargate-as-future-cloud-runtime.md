# ADR-0023: Usar ECS con Fargate como runtime cloud futuro

- **Estado:** Superado por ADR-0088
- **Fecha:** 2026-07-25
- **Decisor:** Usuario, mediante aceptación explícita de la alternativa A
- **Propietario del análisis:** Asistente como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** ADR-0088

## Problema

Tras aceptar una imagen OCI para la API, hay que elegir la dirección del runtime
cloud que la ejecutará cuando el proyecto alcance esa fase. La elección debe
permitir aprender operación de contenedores sin adelantar Kubernetes ni incurrir
en coste mientras el producto y su vertical slice no existen.

## Contexto y restricciones

- ADR-0022 acepta una imagen OCI para la API, pero no decide dónde se ejecuta.
- El manifiesto sitúa AWS como cloud inicial y Kubernetes en una fase posterior.
- ADR-0018 establece que el desarrollo actual es local: PostgreSQL en Compose;
  API Go y cliente Expo en el host.
- ADR-0017 exige identidades temporales y mínimo privilegio para cloud futura.
- No existe un runtime, cuenta AWS, red, registry, API desplegable ni requisito
  de disponibilidad pública. Crear recursos ahora produciría coste sin evidencia
  de valor.
- La estrategia de promoción sigue pendiente; esta decisión solo selecciona el
  runtime cloud futuro.

## Criterios de decisión

1. no generar gasto ni recursos cloud antes de una necesidad demostrada;
2. ejecutar una imagen OCI sin gestionar servidores ni un clúster propio;
3. enseñar conceptos de servicio, tarea, red, identidad, salud y recuperación;
4. permitir avanzar a Kubernetes después, no como requisito inicial;
5. conservar la lógica de negocio independiente del proveedor;
6. encajar con AWS y Terraform como decisiones de la fase cloud.

## Alternativas

### Alternativa A — Amazon ECS con AWS Fargate

Usar Amazon ECS como orquestador de contenedores y Fargate como capacidad de
ejecución gestionada cuando comience la fase cloud. ECS declarará una tarea y un
servicio; Fargate proporcionará la capacidad sin instancias que administrar.

- **Ventajas:** ejecuta imágenes OCI sin servidores propios; un servicio puede
  reemplazar tareas fallidas; permite aprender límites operativos de AWS antes
  de Kubernetes.
- **Inconvenientes:** introduce conceptos de ECS, IAM, red y coste por recursos
  activos; la infraestructura requerirá diseño y Terraform posteriores.
- **Coste de adopción:** medio, aplazado hasta la fase cloud.
- **Coste de mantenimiento:** medio, menor que administrar hosts o Kubernetes.
- **Riesgos:** crear recursos permanentes demasiado pronto. Se mitiga prohibiendo
  provisión y gasto hasta que el gate de AWS esté abierto y exista un plan de
  coste, rollback y apagado.

### Alternativa B — VPS con Docker Compose

Alquilar una máquina virtual y ejecutar la imagen con Docker Compose.

- **Ventajas:** pocos conceptos iniciales y control directo del host.
- **Inconvenientes:** el equipo administra parches, acceso, capacidad, fallos,
  backups y recuperación del servidor; la paridad con el runtime cloud futuro
  se reduce.
- **Coste de adopción:** bajo o medio.
- **Coste de mantenimiento:** medio o alto por administración continua del host.
- **Riesgos:** convertir una solución temporal en plataforma operativa sin sus
  controles necesarios.

### Alternativa C — Kubernetes gestionado (EKS)

Ejecutar el backend en Amazon EKS.

- **Ventajas:** experiencia Kubernetes portable y rica en capacidades de
  orquestación.
- **Inconvenientes:** añade clúster, manifiestos, red, permisos y operación antes
  de haber aprendido el runtime gestionado más simple.
- **Coste de adopción:** alto.
- **Coste de mantenimiento:** alto.
- **Riesgos:** sobreingeniería contraria al roadmap, que aplaza Kubernetes a la
  Fase 4.

### Alternativa D — AWS App Runner

Usar el servicio web gestionado de App Runner para ejecutar la imagen.

- **Ventajas:** configuración inicial muy reducida.
- **Inconvenientes:** AWS ya no admite nuevos clientes desde el 31 de marzo de
  2026; no es una base viable para este proyecto.
- **Coste de adopción:** no aplicable para nuevos clientes.
- **Coste de mantenimiento:** no aplicable.
- **Riesgos:** depender de un servicio que no se puede adoptar actualmente.

### No cambiar

No fijar aún una dirección de runtime cloud.

- **Consecuencias:** se pospone el aprendizaje sobre el destino de la imagen y
  se arriesga a decidirlo de forma apresurada al llegar a la fase cloud.

## Comparación

La A ofrece un runtime de contenedores gestionado y compatible con la imagen
aceptada, sin administrar hosts ni adoptar Kubernetes. La B reduce conceptos
pero convierte al equipo en operador del servidor. La C anticipa una plataforma
que el roadmap reserva para después de un servicio operable. La D queda
descartada por la restricción actual de disponibilidad de AWS.

## Recomendación

**Opinión/recomendación:** alternativa A. ECS con Fargate es el siguiente nivel
de aprendizaje tras una imagen OCI: mantiene la infraestructura de cómputo
gestionada y deja Kubernetes como evolución posterior justificada.

## Decisión del usuario

**Aceptada:** usar Amazon ECS con AWS Fargate como runtime cloud futuro. Antes
de llegar a esa fase, todo el trabajo y las verificaciones permanecerán locales
y no se crearán recursos AWS ni gasto asociado. Kubernetes se evaluará después
de dominar ECS con Fargate, conforme a la Fase 4.

## Reglas de implementación

- No se crea ni configura ahora ninguna cuenta, servicio, imagen publicada,
  red, rol, registry o recurso facturable de AWS.
- El entorno actual conserva PostgreSQL en Compose y API/Expo en host, según
  ADR-0018.
- Cuando el gate de IaC y AWS se abra, Terraform describirá los recursos y se
  decidirán cuenta, región, red, ECR, IAM, secretos, observabilidad, límites de
  coste, apagado, rollback y recuperación.
- La lógica de negocio no importará SDKs ni tipos AWS; ECS/Fargate será un
  adaptador de infraestructura.
- Kubernetes/EKS no se introducirá hasta la Fase 4 y requerirá una decisión
  posterior con evidencia de aprendizaje y necesidad.
- La publicación, selección de registry, versionado, firma y promoción de la
  imagen siguen pendientes de ADR posterior.

## Consecuencias

### Positivas

- Existe una dirección cloud clara sin realizar gasto presente.
- Se aprende primero el ciclo de una imagen y un servicio gestionado antes de
  operar Kubernetes.
- La imagen OCI mantiene el artefacto separado del proveedor de ejecución.

### Negativas y deuda aceptada

- La experiencia real de ECS, IAM y red queda aplazada hasta la Fase 5.
- La dirección depende de AWS para infraestructura, aunque el dominio y la
  imagen permanecen portables.
- Aún falta decidir cómo se registra, publica y promueve la imagen.

## Validación

Al abrir la Fase 5, se demostrará que:

- Terraform crea y destruye de forma controlada la infraestructura necesaria;
- una tarea Fargate ejecuta la imagen OCI con configuración externa y mínimo
  privilegio;
- el servicio detecta una tarea no saludable y recupera el estado esperado;
- el coste estimado, mecanismo de apagado, rollback y recuperación están
  documentados antes de mantener recursos activos;
- no hay secretos en imagen, estado Terraform, planes ni logs.

## Disparadores de revisión

- el coste o las restricciones de ECS/Fargate no encajan con la carga real;
- la necesidad de ejecución on-premises o multicloud supera el coste de una
  adaptación específica;
- el aprendizaje o carga demuestran que Kubernetes aporta valor medible;
- un requisito de runtime no puede satisfacerse con Fargate;
- AWS cambia de forma material disponibilidad, precios o capacidades relevantes.

## Documentación afectada

- [TECHNICAL_BASELINE.md](../governance/TECHNICAL_BASELINE.md)
- [DEPLOYMENT.md](../operations/DEPLOYMENT.md)
- [DECISIONS.md](../governance/DECISIONS.md)
- [DECISIONS_TO_REVISIT.md](../governance/DECISIONS_TO_REVISIT.md)
- [LEARNING.md](../project/LEARNING.md)
- [README.md](../../README.md)
- [CHANGELOG.md](../../CHANGELOG.md)

## Fuentes técnicas

- [AWS ECS](https://docs.aws.amazon.com/AmazonECS/latest/APIReference/Welcome.html)
- [Servicios ECS](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/ecs_services.html)
- [Precios de AWS Fargate](https://aws.amazon.com/fargate/pricing/)
- [AWS App Runner: cambio de disponibilidad](https://docs.aws.amazon.com/apprunner/latest/api/API_StartDeployment.html)
- [Amazon EKS](https://docs.aws.amazon.com/eks/latest/userguide/)
