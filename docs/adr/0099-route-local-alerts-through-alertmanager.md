# ADR-0099: Enrutar alertas locales mediante Alertmanager

- **Estado:** Aceptado
- **Fecha:** 2026-08-21
- **Decisor:** Usuario, mediante decisión explícita
- **Propietario del análisis:** Asistente como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

Las reglas SLO de Prometheus detectan degradación, pero no agrupan avisos, no admiten silencios ni entregan notificaciones repetidas mientras una alerta sigue activa. El equipo necesita ver ese estado también desde Grafana.

## Contexto y restricciones

- ADR-0098 define alertas locales para el refresh de sesión.
- Las reglas de Prometheus ya son configuración versionada y validada; no se migran a reglas gestionadas por Grafana.
- Mailpit es el receptor exclusivo local: no se añaden credenciales ni correo externo en este corte.
- Grafana permite consultar Alertmanager y crear silencios; las rutas y los receptores de Alertmanager Prometheus siguen siendo declarativos.

## Criterios de decisión

1. conservar reglas como código en Prometheus;
2. agrupar, silenciar y entregar alertas sin lógica propia;
3. aplicar cadencias distintas a `warning` y `critical`;
4. centralizar la consulta operativa en Grafana sin añadir un SaaS.

## Alternativas

### Alternativa A — Alertmanager Prometheus con datasource en Grafana

Prometheus envía alertas a Alertmanager; este agrupa por `alertname`, entrega correo local mediante Mailpit y se expone en Grafana como datasource.

- **Ventajas:** patrón nativo de Prometheus, reglas versionadas, silencios y agrupación; Grafana muestra reglas y alertas activas.
- **Inconvenientes:** rutas y receptores se editan en YAML, no desde Grafana.
- **Coste de adopción y mantenimiento:** bajo.
- **Riesgos:** la cadencia de `critical` puede producir ruido si se usa para síntomas no accionables.

### Alternativa B — Alertas gestionadas por Grafana

- **Ventajas:** reglas, contacto y política en una única interfaz.
- **Inconvenientes:** duplica o sustituye configuración Prometheus ya validada y aumenta el acoplamiento con Grafana.
- **Coste de adopción y mantenimiento:** medio.
- **Riesgos:** reglas menos visibles en revisión de infraestructura y migración prematura.

### No cambiar

Mantener alertas visibles solo en Prometheus.

- **Consecuencias:** no hay entrega, agrupación ni silencios centralizados.

## Comparación

La alternativa A añade solo el componente que resuelve entrega y control, sin mover la evaluación de reglas. La B concentra más interfaz, pero adelanta una migración sin necesidad. No cambiar deja incompleto el ciclo de alerta.

## Recomendación

**Opinión/recomendación:** alternativa A. Mantiene responsabilidades claras: Prometheus evalúa, Alertmanager decide la entrega y Grafana permite operar la información desde una interfaz común.

## Decisión del usuario

**Aceptada:** alternativa A. Las alertas `warning` se repiten cada cuatro horas; las `critical`, cada diez minutos. El primer aviso espera 30 segundos y las actualizaciones de grupo se evalúan cada cinco minutos. Se añaden dos ejemplos `critical`: API no monitorizable y tasa de `5xx` de refresh superior al 20 %.

## Consecuencias

### Positivas

- Los avisos se agrupan por nombre y admiten silencios desde Grafana.
- Mailpit permite verificar correo sin salida externa.
- La criticidad comunica la urgencia y determina la cadencia de repetición.

### Negativas y deuda aceptada

- Mailpit no entrega avisos fuera del equipo; el receptor externo requerirá una decisión separada, credenciales y una prueba de entrega.
- Grafana muestra reglas de Prometheus como solo lectura y no modifica la ruta o receptores de Alertmanager.

## Validación

1. `amtool check-config` y `promtool check config` validan ambos componentes.
2. Prometheus descubre Alertmanager y este acepta alertas.
3. Mailpit recibe un aviso de prueba `warning` y otro `critical`; la configuración entrega también sus resoluciones.
4. Grafana muestra las reglas gestionadas por Prometheus y las alertas de Alertmanager. Un silencio se crea desde su interfaz cuando una operación lo requiere.

## Disparadores de revisión

- primera alerta crítica no accionable o ruido repetido;
- necesidad de correo externo, on-call o varios receptores;
- más de un entorno o instancia que requiera alta disponibilidad.

## Documentación afectada

- `docs/operations/OBSERVABILITY.md`
- `docs/runbooks/session-refresh-observability.md`
- `docs/governance/DECISIONS.md`
- `docs/project/LEARNING.md`
- `CHANGELOG.md`
