# ADR-0025: Usar Terraform para la infraestructura como código

- **Estado:** Aceptado
- **Fecha:** 2026-07-25
- **Decisor:** Usuario, mediante aceptación explícita de Terraform
- **Propietario del análisis:** Asistente como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

La Fase 5 requiere describir, revisar y reproducir la infraestructura AWS sin
acoplar la lógica de negocio al proveedor. Hay que elegir la herramienta de IaC
antes de decidir cuenta, identidad, estado remoto y red.

## Contexto y restricciones

- El manifiesto señala Terraform y AWS como dirección objetivo, pero ninguna
  herramienta de IaC estaba aceptada hasta esta decisión.
- ADR-0023 fija ECS con Fargate como runtime cloud futuro y ADR-0024 fija ECR
  privado; ambos aplazan la creación de recursos AWS a la Fase 5.
- ADR-0017 exige OIDC y credenciales temporales para cloud cuando estén
  disponibles; el estado, los planes y los logs no pueden contener secretos.
- Este ADR decide una herramienta, no una versión, backend de estado, módulo,
  cuenta AWS, región, red, proveedor ni recurso facturable.

## Criterios de decisión

1. configuraciones declarativas, revisables y con previsualización de cambios;
2. soporte sólido para AWS sin cerrar una evolución futura a otros proveedores;
3. coste de aprendizaje y mantenimiento proporcionado a un único servicio;
4. integración con control de versiones, revisión y credenciales temporales;
5. separación completa entre dominio Go e infraestructura.

## Alternativas

### Alternativa A — Terraform CLI con configuración HCL

Terraform describe el estado deseado en archivos versionables y separa los
pasos `plan` y `apply`.

- **Ventajas:** satisface la dirección del manifiesto; HCL es legible para
  infraestructura; su proveedor AWS no condiciona los módulos a un único
  proveedor; el plan hace visibles los cambios antes de aplicarlos.
- **Inconvenientes:** requiere decidir y proteger un estado remoto, bloqueo y
  bootstrap posteriores; introduce HCL y su ecosistema.
- **Coste de adopción:** bajo o medio, aplazado a la Fase 5.
- **Coste de mantenimiento:** bajo para pocos componentes si se mantienen
  módulos propios pequeños y versiones declaradas.
- **Riesgos:** convertir módulos genéricos o la abstracción multicloud en una
  plataforma prematura.

### Alternativa B — OpenTofu

OpenTofu ofrece un flujo y lenguaje muy similares, con un proyecto abierto
independiente.

- **Ventajas:** reduce la dependencia de una licencia o proveedor concreto y
  conserva un modelo declarativo conocido.
- **Inconvenientes:** añade una elección de compatibilidad, distribución y
  actualización que no aporta una necesidad actual frente a la dirección
  explícita del manifiesto.
- **Coste de adopción:** bajo o medio.
- **Coste de mantenimiento:** bajo o medio; exige vigilar compatibilidad de
  providers y módulos con el ecosistema Terraform.
- **Riesgos:** migrar por preferencia antes de que exista evidencia de coste o
  restricción real.

### Alternativa C — AWS CDK o CloudFormation

CDK define infraestructura en un lenguaje general y la sintetiza a
CloudFormation; CloudFormation define plantillas nativas de AWS.

- **Ventajas:** integración directa con el plano de control de AWS y sus
  mecanismos de despliegue.
- **Inconvenientes:** aumenta el acoplamiento a AWS; CDK suma generación y una
  capa de abstracción, mientras CloudFormation expone plantillas más verbosas.
- **Coste de adopción:** medio.
- **Coste de mantenimiento:** medio, especialmente al combinarlo con el resto
  de toolchains del monorepo.
- **Riesgos:** confundir la infraestructura con código de aplicación o crear
  constructs demasiado amplios.

### No cambiar

Seguir con recursos manuales o posponer IaC impediría revisar, reproducir y
destruir de forma fiable la futura infraestructura; no satisface la salida de
la Fase 5.

## Comparación

Terraform y OpenTofu cubren el modelo declarativo y la portabilidad deseada.
Terraform es la opción de menor sorpresa porque coincide con la dirección
documentada y conserva un flujo simple de configuración, plan y aplicación.
CDK y CloudFormation son viables para AWS, pero elevan el acoplamiento sin una
ventaja demostrada para el alcance actual.

## Recomendación

**Opinión/recomendación:** alternativa A. Terraform es la solución mínima
suficiente: hace la infraestructura revisable sin introducir un framework de
aplicación ni una abstracción multicloud prematura.

## Decisión del usuario

**Aceptada:** usar Terraform como herramienta de infraestructura como código.
Se usará para describir infraestructura AWS al abrir la Fase 5. Los módulos
serán propios, pequeños y orientados a recursos reales; no se crearán módulos
genéricos ni se introducirán SDKs AWS en el dominio.

## Reglas de implementación

- No se ejecuta `apply` ni se crean recursos AWS como consecuencia de este ADR.
- Las versiones exactas se fijarán junto al entorno reproducible y su política
  de actualización, no en este ADR.
- Antes del primer trabajo colaborativo se decidirán por ADR separado cuenta,
  identidad, bootstrap, estado remoto y bloqueo.
- Los planes, estados y logs se tratarán como datos sensibles y no se
  versionarán ni publicarán.
- La infraestructura vivirá bajo `infra/`; la lógica de negocio no importará
  Terraform, AWS SDK ni tipos de proveedor.

## Consecuencias

### Positivas

- La infraestructura futura tendrá cambios declarativos, revisables y
  reproducibles.
- AWS no filtra sus SDKs ni conceptos al dominio de la aplicación.
- El flujo `plan` antes de `apply` crea una evidencia verificable del cambio.

### Negativas y deuda aceptada

- El equipo debe aprender HCL, providers, módulos y operación segura del
  estado.
- Elegir Terraform no resuelve aún aislamiento de cuentas, IAM, red, costes,
  recuperación ni la seguridad del backend de estado.

## Validación

Al abrir la Fase 5 se demostrará que:

- `terraform fmt`, `validate` y `plan` se ejecutan sin secretos persistentes;
- un cambio revisado crea y destruye únicamente los recursos autorizados;
- el estado remoto y su bloqueo están protegidos antes del trabajo compartido;
- la destrucción, el rollback y el coste estimado están documentados antes de
  mantener recursos activos.

## Disparadores de revisión

- Terraform deja de satisfacer requisitos de licencia, seguridad, soporte o
  compatibilidad relevantes;
- la operación multicloud real justifica una estrategia distinta;
- el coste de mantener HCL o módulos propios supera una alternativa medida;
- un requisito AWS no puede gestionarse de forma segura y mantenible.

## Documentación afectada

- [TECHNICAL_BASELINE.md](../governance/TECHNICAL_BASELINE.md)
- [DECISIONS.md](../governance/DECISIONS.md)
- [DECISIONS_TO_REVISIT.md](../governance/DECISIONS_TO_REVISIT.md)
- [DEPLOYMENT.md](../operations/DEPLOYMENT.md)
- [LEARNING.md](../project/LEARNING.md)
- [README.md](../../README.md)
- [CHANGELOG.md](../../CHANGELOG.md)

## Fuentes técnicas

- [Introducción a Terraform](https://developer.hashicorp.com/terraform/intro)
- [Documentación de OpenTofu](https://opentofu.org/docs/)
- [AWS CloudFormation](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/Welcome.html)
- [AWS CDK](https://docs.aws.amazon.com/cdk/v2/guide/home.html)
