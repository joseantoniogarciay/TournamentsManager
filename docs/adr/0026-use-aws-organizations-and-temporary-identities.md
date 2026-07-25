# ADR-0026: Usar AWS Organizations e identidades temporales

- **Estado:** Aceptado
- **Fecha:** 2026-07-25
- **Decisor:** Usuario, mediante aceptación explícita de la alternativa A
- **Propietario del análisis:** Asistente como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

La futura infraestructura necesita separar producción de los entornos no
productivos y limitar el acceso humano y de automatización sin depender de
credenciales AWS persistentes. Hay que decidir la fundación de cuentas e
identidades antes de diseñar el estado de Terraform y la red.

## Contexto y restricciones

- ADR-0025 acepta Terraform, pero aplaza cuenta, estado remoto, bootstrap y
  red.
- ADR-0017 exige GitHub Environments protegidos, OIDC y credenciales temporales
  cuando cloud esté disponible.
- ADR-0023 y ADR-0024 fijan ECS/Fargate y ECR para una futura Fase 5, sin
  autorizar recursos AWS ni gasto presente.
- El proyecto es individual y didáctico, pero tendrá futuros entornos `dev`,
  `staging` y `prod`.
- Esta decisión no fija correo raíz, región, backend de estado, políticas IAM
  detalladas, SCPs, red, workloads ni la creación de una cuenta.

## Criterios de decisión

1. aislar producción de errores y permisos de no-producción;
2. conservar una base simple para una persona y un servicio;
3. usar acceso humano federado y temporal con MFA;
4. permitir automatización por OIDC con mínimo privilegio por cuenta y entorno;
5. separar gobierno/facturación de las cargas de trabajo;
6. evitar una landing zone compleja antes de tener necesidades demostradas.

## Alternativas

### Alternativa A — AWS Organizations con gestión, no-producción y producción

Crear una organización con todas las capacidades y tres cuentas: una de gestión
y facturación sin cargas de trabajo, una `nonprod` para los futuros `dev` y
`staging`, y una `prod` para producción. El acceso humano usa una instancia de
organización de IAM Identity Center; GitHub Actions asume roles OIDC separados
por cuenta y environment.

- **Ventajas:** la cuenta es una frontera de acceso, coste y recursos; protege
  producción desde el inicio; centraliza el acceso humano; conserva roles y
  credenciales temporales para personas y automatización.
- **Inconvenientes:** requiere operar una organización, tres correos de cuenta,
  asignaciones de permisos y roles cruzados.
- **Coste de adopción:** medio, aplazado a la Fase 5.
- **Coste de mantenimiento:** bajo o medio para una única carga; crece con
  políticas, cuentas y colaboradores.
- **Riesgos:** aplicar permisos o guardrails complejos antes de haber observado
  las necesidades reales.

### Alternativa B — Una única cuenta con separación lógica por tags y nombres

Mantener `dev`, `staging` y `prod` dentro de una sola cuenta.

- **Ventajas:** menos pasos de inicio y una única superficie administrativa.
- **Inconvenientes:** no hay frontera de cuenta entre producción y
  no-producción; los errores de permisos, cuotas y facturación se comparten.
- **Coste de adopción:** bajo.
- **Coste de mantenimiento:** bajo inicialmente, pero aumenta al extraer
  producción o introducir controles diferenciados.
- **Riesgos:** convertir convenciones de nombres y tags en un sustituto frágil
  de aislamiento.

### Alternativa C — Landing zone completa desde el inicio

Adoptar Control Tower y cuentas separadas de auditoría, logs, seguridad,
servicios compartidos, no-producción y producción.

- **Ventajas:** fuerte separación de responsabilidades y una base cercana a un
  entorno empresarial maduro.
- **Inconvenientes:** más cuentas, controles, costes operativos y conceptos que
  no aportan valor aún a una persona y un servicio.
- **Coste de adopción:** alto.
- **Coste de mantenimiento:** alto.
- **Riesgos:** aprender la plataforma de gobierno antes de poder operar el
  producto y su infraestructura mínima.

### No cambiar

No decidir la fundación de cuentas mantiene bloqueadas la seguridad, el backend
de estado y el despliegue reproducible de la Fase 5.

## Comparación

La alternativa A aplica el aislamiento de cuenta recomendado para cargas de
producción sin adoptar una landing zone completa. La B reduce pasos inmediatos,
pero mezcla los entornos que deben limitar el alcance de incidentes. La C es
adecuada para organizaciones o portfolios mayores, no para el alcance actual.

## Recomendación

**Opinión/recomendación:** alternativa A. Tres cuentas es el mínimo que aísla
la producción, deja la cuenta de gestión libre de cargas y evita añadir cuentas
de seguridad o servicios compartidos sin evidencia de necesidad.

## Decisión del usuario

**Aceptada:** usar AWS Organizations con todas las capacidades y tres cuentas:
`management`, `nonprod` y `prod`. La cuenta `management` centraliza gobierno y
facturación y no alojará cargas de aplicación. `nonprod` alojará los futuros
entornos `dev` y `staging`; `prod` alojará exclusivamente producción.

El acceso humano se realizará mediante IAM Identity Center de organización,
roles temporales y MFA. El root user de cada cuenta se protegerá con MFA y se
usará solo para tareas excepcionales. No se crearán usuarios IAM ni access keys
persistentes para personas o GitHub Actions; la automatización futura asumirá
roles OIDC separados y de mínimo privilegio por cuenta y environment.

## Reglas de implementación

- La cuenta `management` no alojará ECS, ECR, bases de datos ni otros recursos
  de aplicación.
- Cada cuenta tendrá un correo raíz único, contactos alternativos y MFA antes
  de usarla para operaciones normales.
- Las personas recibirán permisos mediante grupos y permission sets de IAM
  Identity Center; no mediante usuarios IAM individuales.
- Cada rol OIDC limitará repositorio, rama o environment autorizado y los
  permisos a la cuenta de destino; el detalle se definirá al implementar CI/CD.
- No se crean todavía la organización, cuentas, Identity Center, roles, OIDC,
  permisos, SCPs ni recursos facturables.
- El estado remoto y su bootstrap se deciden en ADR posterior, sin asumir que
  residirán en `management`.

## Consecuencias

### Positivas

- Producción queda aislada de no-producción por una frontera nativa de AWS.
- Acceso humano y automatización usan credenciales temporales y auditables.
- La facturación queda centralizada sin ejecutar cargas en la cuenta de gestión.

### Negativas y deuda aceptada

- Tres cuentas introducen correos, roles, permisos, consolidación de facturas y
  coordinación de Terraform.
- No habrá de inicio cuentas dedicadas de auditoría, logs, seguridad ni
  servicios compartidos; se añadirán solo con un disparador medido.

## Validación

Al abrir la Fase 5 se demostrará que:

- `management`, `nonprod` y `prod` existen dentro de una organización con todas
  las capacidades;
- el acceso humano diario usa IAM Identity Center, MFA y roles temporales;
- no existen access keys activas para personas ni automatización;
- un rol OIDC de GitHub solo puede asumir el rol y la cuenta previstos;
- un recurso de `nonprod` no es accesible desde `prod` sin una autorización
  explícita y revisada;
- la cuenta de gestión no contiene cargas de aplicación.

## Disparadores de revisión

- varios servicios, equipos o requisitos regulatorios justifican cuentas de
  seguridad, logs o servicios compartidos;
- el coste o la gestión de tres cuentas no encajan con la carga real;
- se requiere separar `dev` de `staging` en cuentas distintas;
- un incidente o auditoría evidencia que los permission sets, roles o límites
  OIDC son insuficientes;
- AWS cambia materialmente Organizations o IAM Identity Center.

## Documentación afectada

- [TECHNICAL_BASELINE.md](../governance/TECHNICAL_BASELINE.md)
- [DECISIONS.md](../governance/DECISIONS.md)
- [DECISIONS_TO_REVISIT.md](../governance/DECISIONS_TO_REVISIT.md)
- [DEPLOYMENT.md](../operations/DEPLOYMENT.md)
- [SECURITY.md](../engineering/SECURITY.md)
- [LEARNING.md](../project/LEARNING.md)
- [README.md](../../README.md)
- [CHANGELOG.md](../../CHANGELOG.md)

## Fuentes técnicas

- [AWS Organizations: conceptos](https://docs.aws.amazon.com/organizations/latest/userguide/orgs_getting-started_concepts.html)
- [Guía AWS de entornos multi-cuenta](https://docs.aws.amazon.com/whitepapers/latest/organizing-your-aws-environment/organizing-your-aws-environment.html)
- [Buenas prácticas IAM](https://docs.aws.amazon.com/IAM/latest/UserGuide/best-practices.html)
- [IAM Identity Center](https://docs.aws.amazon.com/singlesignon/latest/userguide/what-is.html)
- [Usuario root de AWS](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_root-user.html)
