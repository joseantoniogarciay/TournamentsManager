# ADR-0028: Usar HCP Terraform Free para el estado remoto inicial

- **Estado:** Aceptado
- **Fecha:** 2026-07-25
- **Decisor:** Usuario, mediante aceptación explícita de la alternativa A
- **Propietario del análisis:** Asistente como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

Antes del primer `terraform apply` sobre AWS, el estado debe ser remoto,
protegido, recuperable y bloqueable. Hay que elegir su backend y el bootstrap
mínimo sin adelantar la decisión de región y red AWS.

## Contexto y restricciones

- ADR-0025 adopta Terraform y ADR-0026 fija las futuras cuentas `management`,
  `nonprod` y `prod`, sin crear todavía ninguna cuenta o recurso AWS.
- ADR-0027 prohíbe gestionar recursos AWS reales con estado local y deja esta
  comparación explícitamente pendiente.
- El repositorio es público: el estado, los planes persistidos y los tokens no
  se versionan ni se publican.
- La decisión 18e —región y red AWS— sigue pendiente. Un backend S3 exigiría
  seleccionar una región y bootstrappear recursos AWS antes de esa decisión.
- El proyecto es individual y didáctico. La solución debe aportar recuperación
  y bloqueo sin convertir el backend de estado en una plataforma propia.
- Este ADR no crea una organización HCP, workspaces, tokens, cuentas AWS ni
  recursos facturables; solo decide qué se configurará al abrir la Fase 5.

## Criterios de decisión

1. bloquear operaciones concurrentes y conservar versiones recuperables del
   estado;
2. no requerir región, cuenta ni recursos AWS antes de decidir la red;
3. acceso mínimo y auditable, sin secretos en Git, planes publicados o logs;
4. coste y operación proporcionados a una persona y un servicio;
5. salida viable hacia S3 si cambian los límites, coste o requisitos de control.

## Alternativas

### Alternativa A — HCP Terraform Free con ejecución local

HCP Terraform almacena el estado en un workspace. Terraform se ejecuta
inicialmente desde la CLI local; HCP se usa para estado, historial y
coordinación, no para ejecutar workloads AWS de forma remota.

- **Ventajas:** no requiere región ni bootstrap AWS; ofrece estado remoto,
  historial y bloqueo gestionados; el plan gratuito cubre equipos pequeños
  hasta 500 recursos gestionados; reduce la operación inicial.
- **Inconvenientes:** incorpora dependencia de HashiCorp y sus límites o
  condiciones comerciales; se debe proteger un token de HCP para la
  integración CLI o CI futura.
- **Coste de adopción:** bajo; crear organización, workspace y configuración
  `cloud` cuando la Fase 5 esté autorizada.
- **Coste de mantenimiento:** bajo mientras se mantenga dentro del plan y el
  acceso se limite a los workspaces necesarios.
- **Riesgos:** confundir el backend remoto con autorización para aplicar AWS o
  habilitar ejecución remota y auto-apply sin controles. Se mitiga con ejecución
  local inicial, apply manual y roles AWS temporales ya acordados.

### Alternativa B — Backend S3 versionado con lockfile

Terraform guarda el estado en un bucket S3 privado, con versionado y
`use_lockfile = true`.

- **Ventajas:** control directo del dato en AWS; recuperación por versiones;
  permisos IAM detallados por objeto y sin SaaS adicional.
- **Inconvenientes:** necesita región, bucket, IAM, cifrado, protección pública
  y un bootstrap inicial; por tanto mezcla este gate con la decisión 18e y crea
  operación AWS antes de las cargas de producto.
- **Coste de adopción:** medio; bootstrap con un estado transitorio y migración,
  más políticas y procedimiento de recuperación.
- **Coste de mantenimiento:** bajo o medio; hay que mantener el bucket, sus
  políticas, ciclo de vida y acceso de cada cuenta.
- **Riesgos:** omitir versionado o locking, conceder lectura amplia del estado o
  dejar el bootstrap sin procedimiento reproducible.

### No cambiar — Estado local

No satisface ADR-0027 para infraestructura AWS real: no comparte estado, no
coordina escritores ni ofrece recuperación centralizada. No es una alternativa
válida antes del primer apply cloud.

## Comparación

Ambas alternativas remotas pueden bloquear y recuperar el estado si se
configuran correctamente. S3 ofrece mayor control operativo, pero requiere una
región y bootstrap AWS que todavía no están decididos. HCP Terraform Free
cierra el requisito actual sin adelantar esas decisiones. HashiCorp documenta
estado remoto, historial y cola de operaciones; su plan Free está limitado a
500 recursos gestionados. El backend S3 recomienda versionado y soporta locking
con `use_lockfile`; el locking con DynamoDB está deprecado.

## Recomendación

**Opinión/recomendación:** alternativa A. Para un único servicio y antes de
elegir región, HCP Terraform Free con ejecución local es la mínima solución que
aporta bloqueo y recuperación sin construir un backend AWS prematuro. S3 se
reconsiderará si aparece una necesidad de control directo o el plan deja de
encajar.

## Decisión del usuario

**Aceptada:** usar HCP Terraform Free como backend remoto inicial de Terraform.
Se configurará con workspaces separados por frontera de estado autorizada y
ejecución inicial local. Los `plan` y `apply` continuarán siendo explícitos;
ningún auto-apply, ejecución remota o recurso AWS queda autorizado por este ADR.

Al abrir la Fase 5 se creará la organización HCP y los workspaces mínimos, se
protegerá el token de HCP fuera de Git y se verificará la restauración de una
versión de estado. AWS seguirá autenticándose con los roles temporales definidos
en ADR-0026. La decisión de región y red continúa en el gate 18e.

## Consecuencias

### Positivas

- Se cierra el requisito de backend remoto sin crear recursos AWS ni elegir una
  región prematuramente.
- El estado dispone de historial y coordinación de operaciones antes de
  infraestructura real o colaboración.
- S3 permanece como ruta de salida documentada y no se introduce DynamoDB para
  locking.

### Negativas y deuda aceptada

- El estado depende inicialmente de un servicio de HashiCorp y de sus límites
  vigentes.
- Se aprende menos operación nativa de S3 al inicio; esa experiencia se aplaza
  hasta que exista una razón real para migrar.
- HCP añade una identidad y token que deben protegerse y revisarse.

## Validación

Antes del primer apply AWS se demostrará que:

- cada workspace usa HCP Terraform y no existe estado en Git;
- dos operaciones de escritura sobre el mismo workspace no pueden ejecutarse
  simultáneamente;
- se puede restaurar de forma controlada una versión previa del estado;
- el token de HCP no aparece en configuración versionada, logs ni planes;
- la ejecución permanece local y el acceso AWS usa credenciales temporales;
- un `plan` revisado no habilita `apply` automático.

## Disparadores de revisión

- superar o aproximarse al límite vigente de recursos del plan Free;
- cambio material de precio, límites, seguridad o disponibilidad de HCP
  Terraform;
- requisitos de residencia, control o recuperación que exijan S3;
- incorporación de colaboradores, CI/CD o ejecución remota con necesidades de
  gobierno no cubiertas;
- decisión de región AWS que haga preferible operar S3 directamente.

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

- [Planes y capacidades de HCP Terraform](https://developer.hashicorp.com/terraform/cloud-docs/overview)
- [Workspaces y estado de HCP Terraform](https://developer.hashicorp.com/terraform/cloud-docs/workspaces)
- [Operaciones remotas de HCP Terraform](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/run/remote-operations)
- [Backend S3 de Terraform](https://developer.hashicorp.com/terraform/language/backend/s3)
- [Versionado de Amazon S3](https://docs.aws.amazon.com/AmazonS3/latest/userguide/Versioning.html)
