# ADR-0101: Usar una VM Linux con K3s y laboratorios EKS efímeros

- **Estado:** Aceptado
- **Fecha:** 2026-08-22
- **Decisor:** Usuario, mediante decisión explícita
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** ADR-0029, solo en la selección de tareas Fargate como cómputo cloud
- **Superado por:** Ninguno

## Problema

La Fase 4 debe enseñar Kubernetes con una operación suficientemente fiel a un
futuro host doméstico, sin convertir el Mac ni AWS en una plataforma de alta
disponibilidad o de coste permanente. El recorrido cloud posterior debe validar
el mismo modelo Kubernetes sin ejecutar en paralelo dos runtimes que puedan
recibir tráfico o datos del mismo entorno.

## Contexto y restricciones

- La API dispone de una imagen OCI de runtime, configuración externa,
  `GET /healthz` y observabilidad, conforme a ADR-0022 y Fase 3.
- Compose mantiene el desarrollo diario y el runtime público `dev`; esta
  decisión no los sustituye.
- k3d ejecuta K3s dentro de Docker y es apropiado para desarrollo rápido, pero
  no permite practicar directamente la operación de un host Linux con K3s.
- Un K3s de un solo nodo es un clúster Kubernetes funcional, pero el Mac, la VM
  y su almacenamiento siguen siendo un único dominio de fallo. No se afirma ni
  se simula alta disponibilidad.
- ADR-0088 conserva AWS como laboratorio temporal y bloquea gasto hasta tener
  un plan, una estimación y autorización explícita. ADR-0029 queda parcialmente
  superado: su topología estaba ligada a ECS/Fargate y se rediseñará antes del
  primer laboratorio EKS.
- EKS será una prueba temporal y aislada: mientras esté activo, la VM K3s que
  represente el laboratorio se detendrá; no habrá dos despliegues activos del
  mismo entorno ni datos compartidos entre ambos.

## Criterios de decisión

1. practicar los primitives Kubernetes y la operación real de K3s sobre Linux;
2. conservar Compose como bucle de desarrollo de bajo coste;
3. mantener un camino comprensible de K3s local a Kubernetes gestionado en AWS;
4. no introducir una falsa alta disponibilidad ni coste cloud recurrente;
5. aislar configuración, secretos, tráfico y datos de cada laboratorio;
6. permitir destruir y reproducir cada entorno sin cambiar el dominio.

## Alternativas

### Alternativa A — VM Linux de un nodo con K3s y EKS temporal

Ejecutar K3s directamente en una VM Linux ligera local. En la Fase 5, crear un
laboratorio EKS con Terraform, validar un despliegue y destruirlo; durante esa
prueba la VM K3s permanece detenida.

- **Ventajas:** enseña K3s como servicio Linux, runtime, red, almacenamiento,
  upgrades y recuperación; conserva Deployments, Services y `kubectl` como
  modelo común para EKS; separa el laboratorio Kubernetes de Compose.
- **Inconvenientes:** añade una VM y sus copias, actualizaciones y diagnóstico;
  EKS requerirá decisiones específicas de red, identidad, datos, ingress,
  almacenamiento y coste.
- **Coste de adopción:** medio; la VM usa recursos locales y EKS se limita a
  ejercicios autorizados y temporales.
- **Coste de mantenimiento:** bajo o medio para una VM de un nodo; medio para
  cada laboratorio EKS, mitigado por destrucción con Terraform.
- **Riesgos:** tratar un nodo único como HA o mezclar datos entre laboratorios.
  Se mitiga declarando el límite y manteniendo datos y secretos aislados.

### Alternativa B — k3d local y ECS/Fargate temporal

Mantener K3s encapsulado en Docker para Fase 4 y usar el modelo propio de ECS
en AWS.

- **Ventajas:** menor fricción local y operación cloud inicial más simple.
- **Inconvenientes:** no practica K3s sobre un host Linux ni conserva el modelo
  Kubernetes al ir a AWS; obliga a aprender dos orquestadores distintos.
- **Coste de adopción:** bajo o medio.
- **Coste de mantenimiento:** bajo localmente; medio para el laboratorio ECS.
- **Riesgos:** confundir la facilidad de k3d con operación de K3s real.

### Alternativa C — K3s autogestionado sobre EC2

Instalar y operar K3s directamente en instancias EC2 en vez de usar EKS.

- **Ventajas:** máxima continuidad con la VM local y control del clúster.
- **Inconvenientes:** AWS no administra control plane, nodos, actualizaciones ni
  recuperación; duplica la responsabilidad de operar hosts durante el
  aprendizaje cloud.
- **Coste de adopción y mantenimiento:** medio o alto.
- **Riesgos:** convertir un laboratorio en un clúster cloud frágil y costoso.

### No cambiar

Mantener Kubernetes aplazado y la ruta cloud de ECS/Fargate.

- **Consecuencias:** se evita complejidad inmediata, pero no se cubre el
  objetivo de aprendizaje Kubernetes ya autorizado para la Fase 4.

## Comparación

La alternativa A añade una VM, pero enseña los límites que k3d oculta y reutiliza
el modelo Kubernetes en EKS. La B sigue siendo válida para ciclos rápidos, pero
no responde a la pregunta de operar K3s en un host. La C conserva Kubernetes en
AWS a costa de administrar también la plataforma; EKS acota esa responsabilidad.

## Recomendación

**Opinión/recomendación:** alternativa A. Es la mínima suficiente para aprender
operación Kubernetes local y validar posteriormente su variante gestionada en
AWS, sin asumir que un único Mac o nodo ofrece alta disponibilidad.

## Decisión del usuario

**Aceptada:** la Fase 4 usará una VM Linux ligera de un único nodo con K3s.
Compose conserva desarrollo y el runtime público `dev`. En Fase 5, EKS se
usará solo para pruebas temporales y autorizadas; al activarlo se detendrá la VM
K3s del laboratorio. Cada entorno tendrá datos, configuración y secretos
propios; no se replica ni se comparte estado entre ambos.

La distribución Linux, el virtualizador, el aprovisionamiento de la VM, la
topología EKS, cómputo, red, ingress, persistencia, secretos y coste siguen
pendientes de decisiones posteriores.

## Consecuencias

### Positivas

- El aprendizaje local reproduce un host K3s real sin comprar aún un mini-PC.
- El modelo de recursos Kubernetes se conserva al validar EKS.
- La destrucción de AWS mantiene el gasto acotado y evita una segunda plataforma
  activa.

### Negativas y deuda aceptada

- La VM comparte hardware, energía y red con el Mac; no hay HA.
- La documentación y automatización de la VM se convertirán en una pieza
  operativa que actualizar y respaldar.
- EKS no es una copia literal de K3s: las integraciones de AWS requieren diseño
  específico antes de crear recursos.

## Validación

- La VM se crea de forma reproducible, K3s inicia y expone un nodo sano.
- La API de runtime se despliega con configuración externa, recursos y probes;
  un fallo controlado demuestra recuperación.
- La VM puede detenerse y arrancarse sin compartir datos con Compose ni con AWS.
- Antes de EKS, Terraform presenta plan de creación y destrucción, coste,
  duración, rollback y aislamiento; el usuario lo autoriza explícitamente.
- El laboratorio EKS se verifica, se destruye y se comprueba la ausencia de
  recursos facturables previstos antes de reactivar la VM K3s.

## Disparadores de revisión

- necesidad de alta disponibilidad, varios nodos o disponibilidad garantizada;
- recursos del Mac insuficientes o mantenimiento mayor que un mini-PC dedicado;
- coste o complejidad de EKS incompatibles con el aprendizaje buscado;
- requisitos de cloud que justifiquen EKS permanente, ECS/Fargate u otro
  runtime;
- incidente que revele que detener la VM no evita una colisión de datos o tráfico.

## Documentación afectada

- [TECHNICAL_BASELINE.md](../governance/TECHNICAL_BASELINE.md)
- [DECISIONS.md](../governance/DECISIONS.md)
- [DEPLOYMENT.md](../operations/DEPLOYMENT.md)
- [DECISIONS_TO_REVISIT.md](../governance/DECISIONS_TO_REVISIT.md)
- [ROADMAP.md](../project/ROADMAP.md)
- [LEARNING.md](../project/LEARNING.md)

## Fuentes técnicas

- [K3s: clúster de un solo nodo](https://docs.k3s.io/quick-start)
- [K3s: arquitectura y alta disponibilidad](https://docs.k3s.io/architecture)
- [Amazon EKS](https://docs.aws.amazon.com/eks/latest/userguide/)
