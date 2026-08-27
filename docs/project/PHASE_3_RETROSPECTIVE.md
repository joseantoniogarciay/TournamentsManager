# Retrospectiva técnica — Fase 3

- **Fecha:** 2026-08-21
- **Objetivo:** poder explicar el estado del backend y diagnosticar fallos con
  logs, métricas y trazas correlacionables, un dashboard y el runbook de una
  ruta crítica.
- **Participantes:** Usuario y Codex.

## Resultado frente al objetivo

La fase cumple su criterio de salida. `POST /v1/sessions/refresh` permite
recorrer una misma degradación desde la métrica agregada hasta la traza HTTP y
PostgreSQL y el log JSON correlacionado. Grafana provisiona las fuentes de
Prometheus, Loki, Tempo y Alertmanager, además del dashboard **SLO — Refresh de
sesión**. El runbook documenta diagnóstico, alertas, silencios, fallo controlado
y recuperación.

La prueba local detuvo PostgreSQL sin eliminar su volumen, generó respuestas
`500` seguras con `database.query_failed` y activó automáticamente
`SessionRefreshFailureRateCritical`. Alertmanager entregó el aviso a Mailpit y
su resolución tras recuperar la base de datos. En `dev`, una alerta sintética
marcada `test=true` recorrió Alertmanager, Resend SMTP, el alias de Cloudflare y
el buzón externo confirmado por el usuario.

Quedan fuera retención de producción, alta disponibilidad, guardias, perfiles,
muestreo avanzado, OpenTelemetry Collector y SLOs generales. Son deudas
deliberadas con disparadores registrados, no requisitos incumplidos de esta
fase.

## Decisiones

- **Funcionaron:** empezar por una pregunta operativa concreta evitó crear
  dashboards y alertas por completitud; Prometheus evalúa, Alertmanager agrupa y
  entrega, y Grafana concentra la consulta sin apropiarse de las reglas.
- **Coste inesperado:** duplicar el stack en `dev` suma seis contenedores y un
  segundo circuito de credenciales. Además, un secreto Docker es contenido
  literal: conservar los comentarios del archivo de ejemplo convirtió las tres
  líneas en una contraseña SMTP inválida, aunque la API transaccional funcionaba
  con su archivo de variables.
- **Revisar:** Collector o backend gestionado cuando haya varios procesos,
  destinos, filtrado central o retención insuficiente; guardias y receptores
  adicionales solo cuando exista una necesidad operativa real.
- **ADR ausentes:** ninguno para el alcance cerrado. ADR-0020, ADR-0098,
  ADR-0099 y ADR-0100 cubren stack, SLO y entrega local y externa.

## Aprendizaje

- Una señal aislada no demuestra observabilidad: el mismo fallo debe conservar
  contexto entre HTTP, PostgreSQL, logs y trazas.
- Un SLO convierte una métrica en una frontera de actuación; Alertmanager
  convierte la evaluación en una entrega agrupable y silenciable.
- Una prueba de entrega sintética y una prueba de regla real responden preguntas
  distintas. La fase necesitó ambas: validar el canal y validar el disparo.
- Los archivos de secretos no son archivos `.env`: deben contener exactamente
  el valor que consume el proceso y validarse por forma sin imprimirlo.

## Calidad profesional

- **Seguridad:** no se exportan tokens, cookies, SQL, argumentos, emails, IDs ni
  errores brutos. Las causas son cerradas; las interfaces operativas permanecen
  en loopback y Alertmanager usa una clave Resend exclusiva, restringida al
  subdominio remitente y fuera de Git.
- **Pruebas:** configuraciones y reglas se validan con `promtool` y `amtool`; la
  caída local demostró disparo y resolución reales, y la prueba de `dev` demostró
  entrega externa sin detener el servicio público.
- **Observabilidad:** logs, métricas y trazas comparten nombres estables y
  cardinalidad acotada; el inventario revisa las salidas relevantes de los
  endpoints del backend.
- **Operación y recuperación:** el runbook limita la caída a PostgreSQL, prohíbe
  borrar volúmenes y exige verificar la recuperación desde las tres señales.
- **Coste:** no se añadió SaaS de observabilidad ni recurso cloud. `dev` mantiene
  retención de 24 horas y Resend dentro de su cuota, a cambio de operar el stack
  en el Mac.
- **Documentación:** observabilidad, runbook, ADR, decisiones revisables,
  aprendizaje y changelog evolucionaron junto a cada corte.

## Complejidad

El stack local es mayor que el mínimo para ver logs, pero cada pieza ya responde
una pregunta demostrada: Prometheus mide y evalúa, Loki conserva logs, Tempo
conserva trazas, Grafana consulta y Alertmanager entrega. Añadir Collector,
Kubernetes, perfiles, HA o otro backend sin uno de sus disparadores sería
sobreingeniería.

La exportación directa y Compose siguen siendo suficientes mientras exista una
sola API y los datos de observabilidad sean cortos y desechables.

## Acciones

| Acción | Propietario | Disparador | Destino |
| --- | --- | --- | --- |
| Definir el laboratorio local de Kubernetes | Usuario/Codex | Resuelto: ADR-0101 acepta VM Linux + K3s; ADR-0109 autoriza iniciar Fase 4 | [ADR-0101](../adr/0101-use-linux-vm-k3s-and-ephemeral-eks-labs.md) y [ROADMAP.md](ROADMAP.md) |
| Revisar ruido, rebotes y cuota de alertas externas | Usuario/Codex | Primera alerta no accionable o problema de entrega | [DECISIONS_TO_REVISIT.md](../governance/DECISIONS_TO_REVISIT.md) |
| Revisar Collector, retención o backend gestionado | Usuario/Codex | Disparadores de ADR-0020 | ADR sucesor y [OBSERVABILITY.md](../operations/OBSERVABILITY.md) |

## Cierre

La Fase 3 queda cerrada. La observabilidad del backend permite detectar,
diagnosticar, entregar y resolver una degradación real sin exponer datos
sensibles. Su acción de Kubernetes se resolvió después mediante ADR-0101 (VM
Linux con K3s) y ADR-0109 (inicio sin bloquear por PostHog distribuible).
