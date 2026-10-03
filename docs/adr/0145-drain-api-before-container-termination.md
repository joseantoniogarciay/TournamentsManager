# ADR-0145: Drenar la API antes de terminar el contenedor

- **Estado:** Aceptado
- **Fecha:** 2026-10-03
- **Decisor:** Usuario, mediante «Apliquemos mejora entonces»
- **Concreta:** ADR-0022 y ADR-0111

## Problema, contexto y criterios

La API captura SIGTERM y drena HTTP hasta diez segundos, pero K3s solo concede
quince segundos y no espera la propagación de la retirada del endpoint. El cierre
de telemetría podía consumir otros diez segundos después del pool. La imagen
scratch ejecuta /api directamente: no necesita tini para recibir señales ni
contiene un ejecutable sleep. Se busca reducir cortes durante rollout sin añadir
infraestructura, dependencias ni lógica de negocio.

## Alternativas y coste

- Mantener el cierre actual: coste nulo, conserva la carrera con el ingress y el
  presupuesto insuficiente.
- Espera nativa preStop y presupuesto coherente: coste bajo; añade cinco segundos
  a la terminación y requiere medir la propagación real.
- Endpoint de drenaje explícito: coste mayor por estado, contrato y protección de
  una nueva superficie operativa; no hay evidencia que lo justifique ahora.

## Recomendación y decisión

El usuario acepta mejorar el mecanismo existente tras revisar sus límites.
Se concreta con preStop.sleep de cinco segundos, terminationGracePeriodSeconds
30, HTTP hasta diez segundos y telemetría hasta cinco. La API sigue atendiendo
mientras kubelet espera; SIGTERM inicia Shutdown. Si el drenaje vence, Server.Close
cierra las conexiones restantes y cancela sus contextos antes del cierre del pool.
Compose dev dispone también de treinta segundos, sin preStop de Kubernetes.
El hook nativo está disponible en el K3s 1.36.3 documentado por el proyecto.

## Consecuencias y validación

Orden normal: retirada del endpoint y espera, SIGTERM, HTTP, PostgreSQL,
telemetría. Quedan diez segundos nominales de margen en K3s para cierre del pool
y planificación. Pool.Close no tiene deadline: un adaptador que ignore la
cancelación todavía puede consumir el margen y acabar en SIGKILL. El runtime
mantiene el límite final; no se promete disponibilidad continua.

Pruebas con servidor y sockets reales verifican respuesta activa completa,
rechazo de nuevas conexiones y cancelación forzada al vencer el drenaje.
La aceptación del manifiesto y una prueba bajo tráfico a través de Traefik se
verifican al promover una revisión limpia según el runbook. Una comprobación
local no acredita la propagación de endpoints ni el estado desplegado.

Revisar cinco segundos si aparecen errores durante rollout; ajustar con medidas,
no ampliando indefinidamente la espera. Revisar el presupuesto si cambian los
plazos de HTTP, telemetría, PostgreSQL o el runtime. Un fallo del host, eliminación
forzada o SIGKILL no disfruta de este mecanismo.

## Evidencia de cierre local

El 2026-10-03 pasan `go test -race ./cmd/api`, `make format-check-go lint-go test
build` (lint sin incidencias), Prettier para los dos YAML y `git diff --check`.
El YAML de K3s se parsea localmente y confirma 30s/5s. Las pruebas de persistencia
que requieren una base aislada no se han ejecutado; no cambian SQL ni adaptadores.
La validación server-side no se pudo realizar: el alias SSH fasttourney-k3s no
está configurado en este host. No se ha desplegado: el checkout contiene cambios
previos y el procedimiento de promoción exige una revisión limpia explícita.
La prueba de tráfico en el cluster permanece pendiente de esa promoción.

## Documentación y fuentes

Deployment, runbook K3s, LEARNING, CHANGELOG e índice de decisiones.

- [Kubernetes: terminación de pods](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-termination)
- [Kubernetes: hooks y presupuesto compartido](https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks/)
- [Go: Server.Shutdown](https://pkg.go.dev/net/http#Server.Shutdown)
