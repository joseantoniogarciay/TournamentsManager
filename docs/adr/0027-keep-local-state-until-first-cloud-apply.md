# ADR-0027: Mantener estado local hasta el primer apply cloud

- **Estado:** Aceptado
- **Fecha:** 2026-07-25
- **Decisor:** Usuario, mediante aceptación explícita de estado local transitorio
- **Propietario del análisis:** Asistente como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

Terraform necesita estado para asociar la configuración declarada con los
recursos reales. Aún no existe infraestructura AWS ni trabajo colaborativo, pero
hay que evitar pagar u operar un backend remoto antes de necesitarlo.

## Contexto y restricciones

- ADR-0025 acepta Terraform y ADR-0026 fija cuentas AWS e identidades, pero no
  autoriza recursos cloud todavía.
- El repositorio es público y el estado puede contener valores sensibles; no se
  versiona ni se publica.
- El proyecto es individual durante esta etapa. Un backend remoto será necesario
  antes de compartir cambios o gestionar infraestructura AWS real.
- Esta decisión no selecciona HCP Terraform ni S3 como backend remoto futuro.

## Criterios de decisión

1. no crear gasto ni dependencia externa antes de una necesidad real;
2. no exponer estado ni valores sensibles en GitHub;
3. exigir bloqueo, recuperación y acceso controlado antes de gestionar AWS;
4. mantener abierta la comparación futura entre HCP Terraform Free y S3.

## Alternativas

### Alternativa A — Estado local transitorio y backend remoto posterior

Mantener el backend local mientras no haya recursos AWS ni colaboración. Antes
del primer `apply` que cree o cambie AWS, elegir y configurar un backend remoto.

- **Ventajas:** coste cero y ningún servicio adicional durante la etapa actual;
  evita decidir con hipótesis de uso y colaboración.
- **Inconvenientes:** no ofrece estado compartido, bloqueo remoto ni recuperación
  centralizada.
- **Coste de adopción:** bajo ahora; medio al abrir cloud.
- **Coste de mantenimiento:** nulo mientras no exista estado cloud; posterior
  según el backend elegido.
- **Riesgos:** usarlo indebidamente para una infraestructura real. Se mitiga
  prohibiendo cualquier `apply` AWS sin backend remoto aceptado.

### Alternativa B — S3 por cuenta desde ahora

Crear buckets S3 versionados, privados y con locking nativo para cada cuenta.

- **Ventajas:** backend AWS nativo, control directo del dato, versionado y coste
  proporcional al uso.
- **Inconvenientes:** crea recursos y coste AWS antes de necesitarlos; requiere
  bootstrap, IAM y operación de buckets.
- **Coste de adopción y mantenimiento:** medio.

### Alternativa C — HCP Terraform Free desde ahora

Usar el plan gratuito de HCP Terraform para estado remoto y ejecución.

- **Ventajas:** estado compartido, locking y ejecución remota sin crear backend
  AWS; el plan gratuito cubre equipos pequeños hasta 500 recursos gestionados.
- **Inconvenientes:** añade un SaaS y dependencia de HashiCorp; sus límites y
  condiciones pueden cambiar; no enseña la operación del backend AWS nativo.
- **Coste de adopción:** bajo o medio.
- **Coste de mantenimiento:** bajo mientras encaje en el plan gratuito.

### Alternativa D — GitHub como almacenamiento de estado

Guardar archivos `.tfstate` en Git, Git LFS, artifacts o secrets de GitHub.

- **Ventajas:** aparente reutilización de un servicio existente.
- **Inconvenientes:** no es backend Terraform, no proporciona locking ni
  actualización atómica; Git conserva el historial de valores sensibles y el
  repositorio es público.
- **Coste de adopción:** bajo, pero inseguro.
- **Coste de mantenimiento:** alto por recuperación manual y riesgo operativo.
- **Riesgos:** exposición de secretos y corrupción o divergencia de estado.

## Comparación

La A es proporcional al hecho de que todavía no hay recursos AWS. La B será
preferible si el aprendizaje y control del backend AWS nativo pesan más que su
operación. La C será preferible si se priorizan coste cero y capacidades remotas
gestionadas para un equipo pequeño. La D no es válida para estado Terraform.

## Recomendación

**Opinión/recomendación:** alternativa A ahora. Cuando sea necesario un backend
remoto, comparar HCP Terraform Free y S3 con el número de recursos, colaboración,
coste y necesidad de aprender la operación AWS real.

## Decisión del usuario

**Aceptada:** mantener estado local únicamente durante la etapa individual y
sin infraestructura AWS real. Antes del primer `terraform apply` que cree,
modifique o destruya recursos AWS, se decidirá mediante ADR entre HCP Terraform
Free y S3. GitHub no se usará como almacenamiento de estado.

## Reglas de implementación

- `.tfstate`, backups locales, planes y el directorio `.terraform/` permanecen
  ignorados por Git y fuera de logs o artefactos públicos.
- No se ejecutará `terraform apply` contra AWS hasta aceptar y configurar un
  backend remoto con locking y recuperación.
- El estado local solo puede utilizarse para ejercicios sin recursos AWS reales
  y permanece en un dispositivo con cifrado de disco habilitado.
- La siguiente decisión comparará HCP Terraform Free y S3, incluyendo límites
  vigentes, coste estimado, versionado, locking, IAM y recuperación.

## Consecuencias

### Positivas

- No se crea coste ni plataforma de estado antes de usar AWS.
- Se evita tomar una decisión por estimaciones de carga o colaboración.

### Negativas y deuda aceptada

- No existe colaboración ni recuperación centralizada durante esta etapa.
- El primer despliegue cloud queda condicionado a cerrar la decisión del backend
  remoto y verificar su bootstrap.

## Validación

Antes de gestionar AWS se demostrará que el backend remoto elegido:

- bloquea ejecuciones simultáneas;
- protege el acceso y el cifrado del estado;
- recupera una versión anterior tras un cambio accidental;
- no expone estado ni secretos en Git, logs o planes publicados.

## Disparadores de revisión

- primer `apply` sobre AWS;
- colaboración, CI/CD o necesidad de recuperación compartida;
- cambio de límites, precio o capacidades de HCP Terraform Free;
- coste o requisitos de control que hagan preferible S3.

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

- [Estado sensible en Terraform](https://developer.hashicorp.com/terraform/language/manage-sensitive-data)
- [Backend local](https://developer.hashicorp.com/terraform/language/backend/local)
- [Backend S3](https://developer.hashicorp.com/terraform/language/backend/s3)
- [HCP Terraform: planes y capacidades](https://developer.hashicorp.com/terraform/cloud-docs/overview)
- [Precios de Amazon S3](https://aws.amazon.com/s3/pricing/)
