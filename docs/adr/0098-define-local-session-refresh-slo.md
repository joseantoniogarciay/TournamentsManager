# ADR-0098: Definir un SLO local para el refresh de sesión

- **Estado:** Aceptado
- **Fecha:** 2026-08-21
- **Decisor:** Usuario, mediante aceptación explícita
- **Propietario del análisis:** Asistente como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

La instrumentación correlacionada permite diagnosticar `POST /v1/sessions/refresh`, pero no expresa qué degradación requiere atención ni hace visible su tendencia en un único lugar.

## Contexto y restricciones

- ADR-0020 acepta la observabilidad local mínima y aplaza alertas y SLO hasta contar con un flujo crítico real.
- Refresh de sesión ya es ese flujo: evita una pérdida silenciosa de acceso y tiene métricas HTTP, trazas, logs y un runbook.
- El entorno es local y no configura Alertmanager, notificaciones externas, retención de producción ni compromisos para personas usuarias.
- Métricas, reglas y dashboard usan solo la plantilla de ruta y estado HTTP; no incorporan identificadores, tokens, cookies ni PII.

## Criterios de decisión

1. detectar indisponibilidad y degradación perceptible del refresh;
2. conservar una operación local pequeña y reversible;
3. basarse en métricas ya emitidas, sin SDK ni servicio nuevo;
4. expresar objetivos comprensibles y verificables.

## Alternativas

### Alternativa A — SLO local acotado al refresh de sesión

Disponibilidad móvil de 30 días del 99,5 % (respuesta distinta de `5xx`) y latencia p95 inferior a 500 ms. Prometheus calcula dos series de grabación y dos alertas locales; Grafana provisiona un dashboard de solo lectura.

- **Ventajas:** responde a un flujo crítico con señales existentes y enseña presupuesto de error sin añadir infraestructura.
- **Inconvenientes:** no entrega notificaciones fuera de las interfaces locales de Prometheus y Grafana.
- **Coste de adopción y mantenimiento:** bajo.
- **Riesgos:** un volumen local pequeño vuelve inestable la ventana de 30 días.

### Alternativa B — SLO general para toda la API

- **Ventajas:** cobertura amplia desde el inicio.
- **Inconvenientes:** mezcla rutas con expectativas y recuperaciones distintas; exigiría decidir varios objetivos y alertas sin evidencia.
- **Coste de adopción y mantenimiento:** medio.
- **Riesgos:** ruido y objetivos sin dueña operativa.

### No cambiar

Mantener métricas y runbook sin objetivos ni alerta.

- **Consecuencias:** el diagnóstico es posible después del incidente, pero no hay una señal explícita de degradación sostenida.

## Comparación

La alternativa A es la unidad mínima que satisface la pregunta operativa de ADR-0020. La B adelanta cobertura sin una necesidad medida; no cambiar mantiene la simplicidad pero deja sin cerrar el aprendizaje de objetivo y presupuesto.

## Recomendación

**Opinión/recomendación:** alternativa A, local y limitada a refresh. Un 99,5 % permite hasta un 0,5 % de `5xx`; el umbral de alerta del 7,2 % en cinco minutos equivale a consumir el presupuesto a 14,4 veces su ritmo sostenible.

## Decisión del usuario

**Aceptada:** alternativa A. El usuario confirmó el SLO de disponibilidad del 99,5 %, p95 inferior a 500 ms, dashboard local y alertas locales antes de pedir su implementación el 2026-08-21.

## Consecuencias

### Positivas

- Grafana muestra disponibilidad, presupuesto consumido, volumen, `5xx` y p95 del refresh en un único dashboard.
- Prometheus expone alertas locales para agotamiento rápido del presupuesto y latencia p95 sostenida.

### Negativas y deuda aceptada

- Las alertas no notifican fuera del entorno local mientras no exista una necesidad y autoridad para operar Alertmanager.
- No se infiere un SLO para otras rutas ni se configura una retención de producción.

## Validación

1. `promtool check config` y `promtool check rules` validan las reglas.
2. Prometheus carga las dos series y dos alertas.
3. Grafana aprovisiona el dashboard `SLO — Refresh de sesión`.
4. El runbook confirma un fallo controlado con las tres señales.

## Disparadores de revisión

- usuarios reales, una alerta que no sea accionable o una pérdida de sesión;
- necesidad de notificación remota, retención o SLO de otra ruta;
- una ventana de 30 días insuficiente por volumen o por acuerdo de servicio.

## Documentación afectada

- `docs/operations/OBSERVABILITY.md`
- `docs/runbooks/session-refresh-observability.md`
- `docs/governance/DECISIONS.md`
- `docs/project/LEARNING.md`
- `CHANGELOG.md`
