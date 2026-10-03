# ADR-0139: Limitar la telemetría de desarrollo a un día

- **Estado:** Aceptado
- **Fecha:** 2026-10-03
- **Decisor:** Usuario, mediante decisión explícita
- **Supera a:** ADR-0106 solo en los plazos de telemetría técnica de desarrollo
- **Complementa:** ADR-0020 y ADR-0100

## Problema, evidencia y restricciones

El usuario pide que montar dev ocupe poco disco, conservar solo un día de logs
y eliminar datos técnicos sobrantes. Ha aclarado que la evidencia legal y sus
copias mantienen sus plazos propios. El gate técnico está cerrado.

Loki tenía 720 horas configuradas sin activar el compactor de retención;
Tempo conservaba siete días; Prometheus ya conservaba 24 horas. Solo Loki tenía
rotación de su salida Docker. Los perfiles Compose local y público reutilizan
estas configuraciones; producción usa valores Helm separados.

## Alternativas y coste

- **Retención y rotación en las piezas existentes:** mantiene el diagnóstico,
  coste bajo y sin servicios nuevos. La purga es asíncrona, no un borrado de
  cada byte exactamente al cumplir 24 horas.
- **Añadir un job que borre archivos de Docker diariamente:** introduce permisos
  sobre el daemon y riesgo de interferencia con los escritores; coste y
  mantenimiento mayores, sin necesidad para controlar el espacio.
- **Mantener la configuración:** evita cambios, pero permite crecer al disco y
  no satisface el plazo decidido.

## Recomendación y decisión del usuario

**Recomendación:** retención efectiva de 24 horas en Loki y Tempo, límite
adicional de tamaño en Prometheus y rotación comprimida de Docker. Filtrar solo
chequeos HTTP técnicos correctos y mantener errores y eventos de negocio.

**Decisión aceptada:** telemetría técnica de `local` y `dev` durante 24 horas;
limpieza de su histórico desechable. La evidencia legal en PostgreSQL y sus
backups mantiene los plazos de ADR-0106. No se cambia producción ni se afirma
que una retención técnica sustituya a un registro legal. La analítica alojada
en PostHog tiene un control separado; esta configuración no modifica ese servicio.

## Implementación

- Loki: 24 horas, rechazo de entradas anteriores a ese plazo y consultas de
  hasta un día; compactor habilitado con marcadores persistentes y borrado
  asíncrono. No basta con cambiar `retention_period`.
- Tempo: bloques de 24 horas. Prometheus: 24 horas y límite de 128 MB sobre
  su retención TSDB, sin presentar ese valor como un límite absoluto de disco
  (WAL, head y compactación añaden espacio transitorio).
- Docker: por contenedor, dos archivos de 5 MB y compresión de los rotados.
  La copia de consola tiene límite de tamaño, no TTL: Docker no ofrece una
  opción de edad en este controlador. Desmontar elimina esa copia.
- Promtail: recoger únicamente la API del perfil correspondiente y descartar
  logs de éxito GET /healthz o GET /metrics. No descartar fallos de esas rutas,
  JSON inválido, arranque ni eventos de aplicación.
- La limpieza inspecciona en lectura los volúmenes de Loki, Tempo, Prometheus
  y posiciones de Promtail: solo retira un volumen sin contenedores y sin ningún
  archivo modificado durante las últimas 24 horas. Ante datos recientes o error
  conserva el volumen completo. Conserva PostgreSQL, Grafana, Alertmanager, evidencia y backups. No usa un
  `docker volume prune` global ni levanta el entorno al terminar.

## Validación y límites

Validar Compose sin imprimir valores resueltos, verificar la configuración con
las imágenes fijadas y probar el filtrado con éxito, fallo y evento de negocio.
Comprobar el arranque de Loki y su rechazo de eventos antiguos. Retirar los
contenedores e imágenes temporales de validación después de las pruebas.

24 horas es la ventana de conservación de las señales; la eliminación física
se ejecuta por ciclos y bloques. La copia Docker se limita por tamaño. No se
promete una cuota total de disco ni expiración física exacta para todo archivo.
Si se requiere un límite duro de disco, hará falta un volumen con cuota.

## Fuentes

- [Loki: retención y compactor](https://grafana.com/docs/loki/latest/operations/storage/retention/)
- [Docker: json-file y rotación](https://docs.docker.com/engine/logging/drivers/json-file/)
- [Prometheus: almacenamiento](https://prometheus.io/docs/prometheus/latest/storage/)
- [Tempo: configuración](https://grafana.com/docs/tempo/latest/configuration/)
