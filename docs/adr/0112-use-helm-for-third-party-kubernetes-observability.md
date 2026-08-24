# ADR-0112: Usar Helm para la observabilidad de terceros en K3s

- **Estado:** Aceptado
- **Fecha:** 2026-08-24
- **Decisor:** Usuario, mediante decisión explícita
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

El runtime doméstico de `prod` incorporará Prometheus, Alertmanager, Loki,
Tempo, Grafana y un colector de logs. Expresar y mantener manualmente todos sus
recursos Kubernetes añadiría una cantidad elevada de manifiestos de terceros,
dependencias y actualizaciones. A la vez, la Fase 4 debe enseñar primero los
primitives de Kubernetes sin ocultarlos detrás de una capa de empaquetado.

## Contexto y restricciones

- El stack equivalente ya funciona en Compose con configuración versionada de
  métricas, alertas, logs, trazas y dashboards.
- La VM K3s es de un nodo, con 4 vCPU y 6 GB de RAM; no se pretende alta
  disponibilidad ni monitorización por completitud.
- La API, PostgreSQL, migraciones, backup, namespace y borde de entrada son
  recursos propios cuyo comportamiento se debe aprender y revisar de forma
  explícita.
- Los secretos siguen fuera de Git y no se introducen como texto plano en
  `values` de Helm.

## Criterios de decisión

1. aprender primero el modelo de Kubernetes visible;
2. reducir el mantenimiento de software de observabilidad de terceros;
3. conservar configuración, versiones y cambios revisables en Git;
4. no añadir un operador de Prometheus ni alta disponibilidad sin una necesidad
   demostrada;
5. mantener la aplicación y sus datos independientes de la herramienta de
   empaquetado.

## Alternativas

### A — Manifiestos propios para el core y Helm para observabilidad

Aplicar con `kubectl` manifiestos YAML explícitos para el core de `prod`.
Después, instalar los componentes de observabilidad de terceros mediante sus
charts Helm upstream, configurados por valores versionados y con versiones
fijadas. No se instala un operador inicialmente.

- **Ventajas:** el aprendizaje inicial conserva `Deployment`, `Service`,
  `StatefulSet`, PVC, `ConfigMap`, `Secret` e `Ingress` visibles; Helm reduce
  el trabajo repetitivo de los componentes externos.
- **Inconvenientes:** conviven dos mecanismos de entrega y hay que revisar los
  manifiestos renderizados antes de cada actualización.
- **Coste de adopción:** bajo para el core; medio al integrar la observabilidad.
- **Coste de mantenimiento:** bajo o medio; los valores y versiones se
  mantienen junto al repositorio.
- **Riesgos:** asumir que Helm resuelve por sí solo secretos, recursos o
  recuperación; esos controles siguen siendo decisiones de cada componente.

### B — Helm para toda la plataforma desde el inicio

Empaquetar también API, PostgreSQL y componentes propios en charts Helm antes de
practicar manifiestos directos.

- **Ventajas:** una interfaz de instalación uniforme.
- **Inconvenientes:** introduce plantillas y valores antes de comprender los
  recursos que generan; un chart propio añade mantenimiento sin capacidad nueva.
- **Coste de adopción y mantenimiento:** medio.
- **Riesgos:** convertir el ejercicio inicial de Kubernetes en depuración de
  plantillas Helm.

### C — Manifiestos propios para todo

Mantener también la observabilidad de terceros como YAML escrito y actualizado
directamente por el proyecto.

- **Ventajas:** máxima transparencia y ausencia de Helm.
- **Inconvenientes:** duplica el empaquetado, las dependencias y el esfuerzo de
  actualización ya resueltos por los mantenedores de cada componente.
- **Coste de adopción y mantenimiento:** medio o alto.
- **Riesgos:** configuraciones incompletas y upgrades manuales frágiles.

## Comparación

La alternativa B uniforma la herramienta, pero adelanta una abstracción que no
ayuda a aprender el core ni reduce el riesgo de PostgreSQL. La C conserva una
transparencia útil, pero su coste crece con seis componentes externos. La A
mantiene visibles las primitives de la aplicación y delega el empaquetado
repetitivo a los charts mantenidos por sus proveedores.

## Recomendación

**Recomendación:** A. Es la solución mínima suficiente: manifiestos explícitos
primero y Helm solo cuando la complejidad real del software externo lo justifica.

## Decisión del usuario

**Aceptada el 2026-08-24:** la entrega K3s será mixta. Primero se construirá el
core de `prod` con manifiestos YAML propios aplicados con `kubectl`, sin Helm ni
operador: host/control plane, namespace y RBAC mínimos, configuración y
secretos, API, PostgreSQL, backup y entrada. Después se adoptará Helm para
instalar y actualizar la observabilidad de terceros. No se crea un chart propio
que agrupe toda la plataforma ni se instala inicialmente un operador de
Prometheus.

La selección concreta de charts, versiones, valores, límites y persistencia se
hará en el diseño de ese módulo, antes de instalar software en la VM.

## Consecuencias

- El primer despliegue de API y PostgreSQL conserva manifiestos sencillos y
  directamente inspeccionables.
- La observabilidad usará valores Helm versionados, charts con versiones fijadas
  y revisión de los manifiestos renderizados antes de aplicar cambios.
- Reglas, dashboards y configuración siguen siendo artefactos versionados; Helm
  no autoriza guardar secretos en Git.
- Helm no sustituye a los controllers ni al kubelet de K3s: solo instala y
  actualiza recursos que aquellos reconcilian.

## Validación

1. Los módulos iniciales se pueden explicar, aplicar, inspeccionar y eliminar
   con `kubectl`, sin Helm.
2. Antes de instalar observabilidad, el diseño identifica cada chart, versión,
   valores, PVC, retención, requests/limits, secreto y procedimiento de
   actualización o rollback.
3. `helm template` permite revisar los recursos que se aplicarán, y una prueba
   controlada confirma que Prometheus descubre cada réplica de API.
4. Un rollback de chart no imprime secretos ni sustituye una restauración de los
   datos persistentes.

## Disparadores de revisión

- El mantenimiento de manifiestos propios de la API o PostgreSQL se vuelve
  repetitivo y justifica un empaquetado adicional.
- La monitorización de recursos del clúster o el descubrimiento de workloads
  exige un operador de Prometheus.
- Los recursos de la VM, la retención o la persistencia hacen inviable el stack
  elegido.
- Se adopta GitOps u otro mecanismo de entrega que cambie materialmente este
  límite.

## Documentación afectada

- [Decisiones](../governance/DECISIONS.md)
- [Roadmap](../project/ROADMAP.md)
- [Despliegue](../operations/DEPLOYMENT.md)
- [Decisiones a revisar](../governance/DECISIONS_TO_REVISIT.md)
- [Aprendizaje](../project/LEARNING.md)
- [Changelog](../../CHANGELOG.md)
