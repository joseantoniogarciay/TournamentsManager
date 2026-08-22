# ADR-0088: Usar AWS de forma efímera para aprendizaje y el Mac como runtime doméstico habitual

- **Estado:** Superado parcialmente
- **Fecha:** 2026-08-11
- **Decisor:** Usuario, mediante aceptación explícita
- **Propietario del análisis:** Asistente como mentor técnico
- **Supera a:** ADR-0023
- **Superado por:** ADR-0101, en la ruta de orquestación cloud

## Problema

El proyecto necesita aprender el ciclo real de AWS sin convertir su infraestructura gestionada en un coste continuo desproporcionado para su volumen inicial.

## Contexto y restricciones

- ADR-0022 mantiene una imagen OCI portable para la API.
- ADR-0023 orientaba el runtime cloud futuro a ECS con Fargate, pero no creaba recursos ni autorizaba gasto.
- El Mac dispone de capacidad suficiente para desarrollo y un servicio doméstico de alcance pequeño; su red y alimentación son un único punto de fallo.
- AWS sigue siendo necesario como aprendizaje práctico de Terraform, EKS,
  identidad, red, observabilidad y destrucción controlada. La topología de
  datos, entrada y cómputo de EKS queda pendiente de decisión posterior.

## Criterios de decisión

1. no mantener coste cloud recurrente sin necesidad demostrada;
2. conservar un recorrido real y reproducible de AWS;
3. validar fuera de desarrollo el mismo artefacto OCI;
4. separar datos, puertos, configuración y volúmenes de desarrollo y del entorno doméstico de release;
5. documentar las limitaciones de disponibilidad del hosting doméstico.

## Alternativas

### Alternativa A — AWS efímero para aprendizaje y Mac como runtime habitual

- **Ventajas:** permite practicar AWS completo; elimina recursos facturables al terminar cada ejercicio; el coste habitual se limita al equipo y electricidad ya disponibles.
- **Inconvenientes:** no ofrece alta disponibilidad ni recuperación regional; exige operar actualizaciones, copias y red doméstica.
- **Coste de adopción:** medio por Terraform, runbooks y verificaciones.
- **Coste de mantenimiento:** bajo o medio en el Mac; variable y acotado en AWS al destruir cada laboratorio.
- **Riesgos:** olvidar recursos facturables en AWS o tratar el Mac como un servicio de disponibilidad garantizada.

### Alternativa B — AWS como producción permanente

- **Ventajas:** infraestructura gestionada, disponibilidad y aislamiento más profesionales.
- **Inconvenientes:** ALB, cómputo, base de datos, logs y red implican gasto continuo aunque el uso sea bajo.
- **Coste de adopción y mantenimiento:** medio o alto.

### Alternativa C — Solo Mac, sin AWS

- **Ventajas:** coste monetario adicional mínimo y operación directa.
- **Inconvenientes:** no se aprende el despliegue cloud acordado ni se valida la infraestructura declarativa en un proveedor real.
- **Coste de adopción:** bajo.

### No cambiar

- **Consecuencias:** AWS queda implícitamente como destino permanente y puede inducir un gasto que el alcance actual no justifica.

## Comparación

La alternativa A mantiene AWS como experiencia técnica verificable, pero evita convertir sus costes fijos en presupuesto operativo. La B solo se justifica al aparecer requisitos de disponibilidad, equipo o tráfico. La C ahorra más, pero elimina una parte explícita del objetivo de aprendizaje.

## Recomendación

**Opinión/recomendación:** alternativa A. Es el equilibrio mínimo suficiente entre aprendizaje cloud real y coste sostenible.

## Decisión del usuario

**Aceptada:** desarrollo y runtime doméstico habitual vivirán en el Mac. AWS se creará temporalmente para aprender y validar el despliegue; después de cada práctica se destruirán los recursos con Terraform. ECS/Fargate deja de ser el destino de producción permanente previsto en ADR-0023.

## Consecuencias

### Positivas

- La imagen OCI y la configuración externa siguen siendo portables entre Mac y AWS.
- AWS se practica con recursos reales y destrucción reproducible.
- Se evita una factura cloud continua sin tráfico que la justifique.

### Negativas y deuda aceptada

- El Mac comparte el riesgo de fallo de desarrollo y servicio doméstico.
- No hay SLA, alta disponibilidad, backup gestionado ni recuperación regional.
- La publicación doméstica usa Cloudflare Tunnel conforme a ADR-0090; dev y
  prod mantienen datos y runtime separados conforme a ADR-0091.

## Validación

- Dev y release doméstico usan proyectos, variables, puertos, volúmenes y bases de datos separados.
- La imagen `runtime` se arranca sin Air ni código montado.
- Un ejercicio AWS se crea mediante Terraform, verifica salud y logs, registra su coste estimado y termina con `terraform destroy` comprobado.
- No quedan recursos facturables tras la destrucción prevista.

## Disparadores de revisión

- usuarios, disponibilidad o cumplimiento que requieran un SLA;
- carga o seguridad que supere la capacidad de la conexión doméstica;
- coste doméstico o mantenimiento superior a un servicio gestionado;
- necesidad real de ejecución permanente en AWS.

## Documentación afectada

- [ADR-0023](0023-use-ecs-fargate-as-future-cloud-runtime.md)
- [DEPLOYMENT.md](../operations/DEPLOYMENT.md)
- [DECISIONS.md](../governance/DECISIONS.md)
- [DECISIONS_TO_REVISIT.md](../governance/DECISIONS_TO_REVISIT.md)
- [LEARNING.md](../project/LEARNING.md)
