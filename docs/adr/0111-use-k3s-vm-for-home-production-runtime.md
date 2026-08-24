# ADR-0111: Usar la VM K3s como runtime doméstico de producción

- **Estado:** Aceptado
- **Fecha:** 2026-08-23
- **Decisor:** Usuario, mediante decisión explícita
- **Propietario del análisis:** Codex como mentor técnico
- **Supera parcialmente a:** ADR-0091 y ADR-0101, solo en la ubicación del
  futuro runtime doméstico de `prod`
- **Superado por:** Ninguno

## Problema

El proyecto necesita inaugurar un entorno doméstico de producción con carga muy
baja. Compose cubre el desarrollo diario y el entorno público `dev`; la VM de
un nodo con K3s permite aprender y operar el modelo Kubernetes deseado, pero
ADR-0101 la reservaba solo para laboratorio.

## Contexto y restricciones

- La VM UTM usa Ubuntu Server 24.04 LTS ARM64 con 4 vCPU, 6 GB de RAM y disco
  dinámico de 30 GB (ADR-0110).
- Docker Desktop tiene un límite de 10 CPU y 8 GB de RAM para Compose.
- La carga esperada es de un único operador y casi ningún usuario externo.
- El Mac, la VM, la red residencial y el suministro eléctrico siguen siendo
  puntos únicos de fallo: no hay SLA ni alta disponibilidad.
- Cloudflare Tunnel y Caddy conservan la entrada pública doméstica; PostgreSQL,
  el API server de Kubernetes y las interfaces operativas no se exponen
  directamente (ADR-0090).
- Datos, configuración, secretos y observabilidad se aíslan de Compose `local`
  y `dev` (ADR-0091).
- La regla de ADR-0101 de detener la VM al activar un laboratorio EKS solo vale
  antes de publicar `prod`. Después, un laboratorio EKS requerirá una ventana de
  mantenimiento explícita o una decisión que aporte capacidad separada.

## Criterios de decisión

1. permitir producción de bajo tráfico sin coste cloud recurrente;
2. conservar el aprendizaje práctico de Kubernetes sobre un host Linux real;
3. mantener una separación estricta entre `local`, `dev` y `prod`;
4. no prometer disponibilidad o recuperación que un único Mac no pueda probar;
5. definir una puerta verificable antes de exponer datos de producción.

## Alternativas

### A — Producción doméstica en la VM K3s

- **Ventajas:** reutiliza la plataforma que se aprende; valida manifests,
  recursos, probes, secretos, persistencia y recuperación sobre un host Linux
  real; no añade gasto recurrente.
- **Inconvenientes:** el usuario opera Kubernetes y el host; un reinicio,
  suspensión o avería del Mac interrumpe el servicio.
- **Coste de adopción:** medio, por persistencia, backup, ingress y runbooks.
- **Coste de mantenimiento:** bajo o medio para una única VM y carga baja.

### B — Producción doméstica en un Compose `prod` separado

- **Ventajas:** menos componentes nuevos y continuidad con `dev`.
- **Inconvenientes:** deja K3s como laboratorio sin validar para el runtime que
  se desea aprender; duplica el recorrido operativo de producción.
- **Coste de adopción:** bajo.
- **Coste de mantenimiento:** bajo, pero con menor valor de aprendizaje.

### C — Producción permanente en AWS

- **Ventajas:** separa el servicio del Mac y permite servicios gestionados.
- **Inconvenientes:** introduce coste recurrente y decisiones de red, datos,
  identidad y recuperación aún no necesarias para la carga inicial.
- **Coste de adopción y mantenimiento:** medio o alto.

## Comparación

La alternativa B reduce el trabajo inicial, pero hace que el runtime de
producción no ejercite K3s y mantiene dos recorridos operativos. La C mejora el
aislamiento físico, pero su coste y decisiones pendientes no se justifican con
la carga actual. La A asume explícitamente el riesgo doméstico y lo compensa
con gates de persistencia, recuperación y rollback antes de publicar.

## Recomendación

**Recomendación:** A. Para una carga inicial casi nula es la solución mínima que
alinea el runtime público con el objetivo de aprendizaje de la Fase 4, sin
afirmar una disponibilidad que un único Mac no puede ofrecer.

## Decisión del usuario

**Aceptada el 2026-08-23:** `prod` se ejecutará en la VM doméstica de un nodo
con K3s. Compose conserva `tournaments-manager-local` para edición diaria y
`tournaments-manager-dev` para desarrollo público. El futuro Compose `prod`
de ADR-0091 no se implementará.

Esta decisión selecciona el runtime, no autoriza todavía publicar producción.
Antes de abrir `fasttourney.com` y `api.fasttourney.com` deben estar implantados
y verificados: persistencia aislada de PostgreSQL, pgBackRest con repositorio
cifrado fuera del volumen, copia completa semanal, incrementales diarios, WAL
archivado y restauración aislada probada, secretos fuera de Git, recursos y
probes, entrada Tunnel/Caddy hacia el ingress de K3s, despliegue y rollback
documentados, y observabilidad con alertas accionables. Hasta entonces los hosts
de producción permanecen en `503`.

`prod` reutiliza el patrón operativo aceptado para `dev` en ADR-0108. Conserva
su propio volumen, repositorio de backup y clave de cifrado; el repositorio vive
fuera del Mac mediante la misma ubicación doméstica sincronizada disponible. No
se finge independencia de cuenta o de proveedor: esa es una limitación aceptada
del hosting doméstico y un disparador de revisión, no una razón para degradar la
recuperación ya validada en `dev`.

Los laboratorios EKS siguen siendo efímeros, pero no detendrán implícitamente un
`prod` ya publicado. Su ejecución requerirá que el usuario autorice la ventana
de mantenimiento o una plataforma de laboratorio independiente.

## Consecuencias

- El runtime de producción y el laboratorio K3s pasan a ser el mismo entorno;
  los cambios dejan de ser meramente pedagógicos y exigen rollback y evidencia
  operativa.
- Los 4 vCPU y 6 GB son suficientes para el lanzamiento de baja carga, pero no
  son una garantía de capacidad; los recursos se revisan con métricas reales.
- K3s y Compose pueden coexistir, pero durante trabajo intensivo se prioriza
  apagar la VM o limitar las cargas para no agotar los 16 GB del Mac.
- La producción sigue siendo doméstica y de mejor esfuerzo. Una necesidad de
  SLA, alta disponibilidad, varios operadores o recuperación fuera del Mac
  obliga a reevaluar el runtime.

## Validación

1. La VM inicia K3s y expone un único nodo sano tras un reinicio controlado.
2. El namespace de `prod` contiene API, PostgreSQL y observabilidad con límites,
   requests y probes acordados; no reutiliza volúmenes ni secretos de Compose.
3. Cloudflare Tunnel alcanza solo el ingress de producción a través de Caddy y
   no hay puertos LAN/WAN ni bases de datos publicados.
4. `pgBackRest check`, una copia completa, un incremental y una restauración
   aislada recuperan PostgreSQL sin tocar datos activos, y el procedimiento queda
   registrado.
5. Un despliegue fallido conserva el servicio previo o permite volver a él sin
   exponer secretos ni modificar datos de forma no recuperable.

## Disparadores de revisión

- usuarios, disponibilidad o cumplimiento que requieran un SLA;
- necesidad de alta disponibilidad, varios nodos o mantenimiento excesivo;
- recursos de la VM insuficientes con evidencia de métricas;
- pérdida, suspensión o incidente del Mac que demuestre que el riesgo doméstico
  ya no es aceptable;
- coste o simplicidad de un runtime gestionado superiores al beneficio de K3s.

## Documentación afectada

- [Despliegue](../operations/DEPLOYMENT.md)
- [Roadmap](../project/ROADMAP.md)
- [Decisiones](../governance/DECISIONS.md)
- [Decisiones a revisar](../governance/DECISIONS_TO_REVISIT.md)
- [Aprendizaje](../project/LEARNING.md)
- [Changelog](../../CHANGELOG.md)
