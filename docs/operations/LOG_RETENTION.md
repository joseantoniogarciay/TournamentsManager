# Retención de registros

> Estado: corrección de producción aplicada conforme a ADR-0140, con siete
> días de diagnóstico y noventa de seguridad. Desarrollo conserva un día.
> Última revisión: 2026-10-03.

| Categoría               | Contenido permitido                                                        |                                            Plazo | Destino                                               |
| ----------------------- | -------------------------------------------------------------------------- | -----------------------------------------------: | ----------------------------------------------------- |
| Diagnóstico             | nivel, ruta, estado, latencia, IDs técnicos de correlación y causa cerrada |             24 horas en `local` y `dev`; 7 días en `prod` | Loki                                                  |
| Seguridad               | evento cerrado de autenticación o límite, sin email, IP, token ni cuerpo   | 24 horas en `dev`; 90 días en `prod` | Loki, streams separados por `retention_category` |
| Trazas                  | atributos técnicos sin PII ni secretos                                     | 24 horas en `local` y `dev`; 7 días en `prod` | Tempo                                                 |
| Métricas `dev`          | agregados sin identificadores                                              |                                         24 horas | Prometheus                                            |
| Evidencia de aceptación | versión, huella de términos, momento, canal y huella de email              |       cuenta activa + 5 años bloqueada tras baja | PostgreSQL y backup legal independiente               |

Los plazos son máximos; un incidente o requerimiento legal puede imponer un
bloqueo documentado. Las copias de la evidencia legal se cifran, se limitan a
personal autorizado y se purgan con una retención que nunca sea menor a la del
registro original. El backup legal no sustituye los planes de recuperación
PostgreSQL ya documentados para `dev` y `prod`.

## Auditoría inicial antes de corregir producción

La revisión del 2026-10-03 detecta que `infra/k3s/observability/loki-values.yaml`
declara `limits_config.retention_period: 24h`, pero no configura
`loki.compactor.retention_enabled`. Se comprobó el archivo `loki/values.yaml`
del chart upstream fijado, `loki-18.11.7`: su valor es `compactor: {}`.
Según la documentación oficial de Loki, sin activar la retención del compactor
los logs se conservan indefinidamente. Por tanto, declarar el plazo no permite
dar por implementada su purga. La configuración activa de la VM confirma
`compactor.retention_enabled: false` y `delete_request_store` vacío; la métrica
`loki_compactor_apply_retention_last_successful_run_timestamp_seconds` es cero.

### Evidencia del runtime, 2026-10-03, 11:57 CEST

Se abrió UTM y arrancó `fasttourney-k3s-lab` por petición del usuario. Tras
cargar la clave en el agente SSH se accedió como operador con sudo sin contraseña.
Se consultaron configuración técnica, métricas y metadatos de archivos; no se
leyeron cuerpos de logs, secretos ni datos de usuarios.

| Componente | Regla efectiva y evidencia | Ocupación | Resultado |
| --- | --- | ---: | --- |
| Loki | 24 h declaradas; compactor sin retención; 1.410 archivos de chunks, el más antiguo modificado el 2026-09-05 | 119 MB | No cumple la purga; existe almacenamiento anterior al plazo |
| Tempo | `compactor.compaction.block_retention: 168h`; sin reglas de retención en el archivo de overrides; 2 borrados y 162 marcados desde el arranque, cero errores de retención o compactación | 73 MB | Purga activa; los metadatos de bloques observados son del día actual |
| Prometheus | Flags `storage.tsdb.retention.time=1d` y `storage.tsdb.retention.size=4GiB`; 2 operaciones de retención y cero fallos de compactación | 2,1 MB | Límites activos, pero quedan 12 bloques del 24–25 de septiembre; confirmar retirada tras el siguiente ciclo |
| Consola Kubernetes | Kubelet `containerLogMaxSize=10Mi`, `containerLogMaxFiles=5`, revisión cada 10 s | 35 MB en `/var/log/pods` | Rotación por tamaño activa, sin TTL de un día |
| Journal del host | Sin overrides explícitos de `SystemMaxUse`, `SystemKeepFree` o `MaxRetentionSec` | 273,3 MB | Usa los valores predeterminados; no hay plazo explícito de conservación |

K3s y todos los servicios de observabilidad están activos. La raíz usa 11 de
27 GB (41 %), con 16 GB libres; los PVC ocupan en conjunto 389 MB. PostgreSQL
ocupa 113 MB y su repositorio de backup 31 MB. No se purgó ni modificó nada.
Las cantidades son una fotografía del arranque y no demuestran una tendencia
creciente. La fecha de modificación de un chunk es evidencia de archivo antiguo,
no una lectura de las fechas de sus eventos.

Prometheus recibe muestras actuales y el reloj de la VM es correcto. Los bloques
antiguos permanecen después de la parada de la VM. La limpieza de bloques
vencidos es asíncrona y puede tardar hasta dos horas según su documentación;
no se declara conforme hasta observar su retirada tras el arranque. No se
forzó una compactación ni se borraron archivos de TSDB.

Todos los PVC usan `local-path`. Su provisionador documenta que los límites de
capacidad no se imponen: solicitar 5 GiB no crea una cuota que proteja el disco
raíz. Es necesario controlar el espacio real, además de la retención.

Fuentes: [almacenamiento Prometheus](https://prometheus.io/docs/prometheus/latest/storage/),
[limitaciones de local-path-provisioner](https://github.com/rancher/local-path-provisioner#cons),
[retención de Loki](https://grafana.com/docs/loki/latest/operations/storage/retention/).

## Política de producción aceptada

[ADR-0140](../adr/0140-bound-production-telemetry-retention.md) sustituye los
plazos técnicos anteriores de ADR-0106: siete días de diagnóstico y noventa de
seguridad. Son plazos operativos justificados para investigar fallos e incidentes,
no obligaciones legales generales acreditadas para estos logs. Evidencia legal,
sus copias y PostHog mantienen su tratamiento independiente.

Alloy asigna la categoría `security` a eventos JSON de las operaciones de acceso,
identidad, credenciales, recuperación y RISC, así como rechazos con causas cerradas
de autenticación, autorización, credencial, reautenticación, sesión, CSRF o límite
de tasa. Los HTTP completados restantes son `diagnostic`. No se infieren causas
de negocio del estado HTTP. Los eventos no reconocidos de la API conservan
noventa días para no perder posibles eventos de seguridad. Solo se descartan
GET /healthz o GET /metrics completados con 2xx; se conservan sus fallos.

El histórico previo no tiene categoría y mezcla ambos tipos: se conserva con un
límite transitorio de noventa días. No se afirma que ese histórico de diagnóstico
ya cumpla siete días. Los nuevos streams sí tienen reglas separadas. La categoría
no es un control de autorización: la instancia Loki sigue siendo privada y sus
operadores pueden consultar ambos tipos.

Loki usa compactor con marcadores persistentes, intervalo de cinco minutos,
borrado asíncrono con demora de dos horas y diez workers. Tempo mantiene siete
días. Prometheus tiene un día o 128 MB de TSDB; WAL, head y compactación pueden
ocupar espacio adicional. Journal tiene siete días, 256 MB, archivos de 32 MB y
reserva de 1 GB libre; kubelet mantiene cinco archivos de 10 MiB por contenedor.

El host mide cada cinco minutos el espacio libre y el tamaño de Loki, Tempo y
Prometheus. Publica solo agregados en un archivo atómico; si falla una medición,
conserva la anterior y su timestamp. Alloy monta solo ese directorio en lectura,
activa únicamente el collector `textfile` y ofrece un Service ClusterIP privado.
Prometheus alerta por falta de espacio, tamaño anormal de telemetría o medición
inexistente, vencida o inaccesible. No se monta el disco raíz ni se añaden permisos
para leer Secrets.

Estas medidas acotan retención y detectan crecimiento; los umbrales de alerta
no son cuotas físicas. El almacenamiento local-path sigue compartiendo la raíz.
La aparición de una alerta exige intervención antes de agotarla; no se elimina
seguridad antes de plazo ni se presenta como garantizado un techo de disco.

### Verificación después de aplicar, 2026-10-03

- Loki 3.7.7 confirma compactor habilitado y reglas efectivas de siete y noventa
  días. Su primer ciclo se completó correctamente a las 12:38 CEST y se
  observaron 21 archivos de marcadores persistentes. La eliminación física es
  asíncrona y tiene una demora de dos horas; no se presenta como ya finalizada.
- Alloy 1.19.2 superó siete casos sintéticos de clasificación y filtrado. La
  primera prueba reveló que `tpl` de Helm evaluaba la plantilla interna de
  Alloy: se escapó y se validó el ConfigMap renderizado. Los Pods de prueba
  fueron retirados; no se desplegaron cambios funcionales de la API.
- Prometheus confirma `1d` y `128MiB`. El target de disco está `up`, las series
  llegan y las reglas de espacio están sanas e inactivas con capacidad normal.
  Las fixtures de `promtool test rules` prueban ausencia de falsos positivos,
  aviso, crítico y medición vencida o ausente. Los bloques del 24–25 de septiembre
  seguían presentes tras arrancar; deben retirarse en el siguiente ciclo de
  compactación y no se han borrado manualmente. Una alerta comprueba su edad
  continuamente, sin esperar a que crezca el volumen.
- Verificación final a las 12:47 CEST: los cuatro targets están `up`, las siete
  reglas tienen salud `ok` y la medición de disco tiene menos de tres minutos.
  La alerta de bloques Prometheus vencidos está `pending` durante el margen
  de tres horas; no se afirma que esos bloques ya hayan desaparecido. Loki
  completó otro ciclo a las 12:46 CEST. La API respondió `200` en /healthz.
- El timer está activo y renueva las mediciones. Journal aplica siete días,
  256 MB, reserva de 1 GB y archivos de 32 MB; bajó de 273,3 a 28,7 MB. La
  raíz está al 40 %, con unos 16 GB libres. `/var/log` ocupa 64 MB; syslog,
  auth.log y kern.log tienen rotación existente de cuatro archivos semanales
  comprimidos y timer activo (no comparten el plazo de logs de aplicación).
- La caché containerd ocupa 5,5 GB, separada de los registros. No se retiraron
  imágenes de producción, datos PostgreSQL, backups ni evidencia legal.

**Retrospectiva:** el plazo declarado no demostraba purga. Las fixtures detectaron
un fallo de clasificación antes de activarla; el cambio requiere validar categorías,
configuración activa, contadores, metadatos y medición del host por separado.

**Comprobación posterior, aproximadamente 18:39 CEST:** las siete alertas de
almacenamiento están sanas e inactivas. La antigüedad TSDB es de unas siete
horas y nueve minutos; la condición de bloques vencidos observada por la mañana
ya no está presente. Loki informa de un ciclo correcto hace 64 segundos y el
timer está activo. No se ha borrado ningún bloque manualmente. La evidencia y
los límites del cierre están en la [revisión operativa](OPERATIONAL_BLOCK_REVIEW_2026-10-03.md).

## Telemetría de desarrollo

[ADR-0139](../adr/0139-limit-development-telemetry-to-one-day.md) limita a un día
los logs, trazas y métricas de los dos perfiles Compose, `local` y `dev`.
No modifica la evidencia legal, sus copias, producción ni PostHog. La evidencia
legal mantiene la cuenta activa y cinco años bloqueada tras la baja conforme a
ADR-0106; no se trata como un log técnico desechable.

Loki aplica 24 horas con compactor habilitado, marcadores persistentes y borrado
asíncrono. Rechaza entradas de más de un día y limita las consultas a esa ventana.
Tempo conserva bloques de 24 horas. Prometheus conserva 24 horas o 128 MB de
retención TSDB, lo que se alcance antes; WAL, head y compactación pueden añadir
espacio transitorio, por lo que no es una cuota absoluta de disco.

Cada contenedor Compose rota su copia de consola en dos archivos de 5 MB,
comprimiendo el rotado. Este controlador Docker limita tamaño, no edad: una
línea antigua puede quedar en esa copia acotada hasta rotar o desmontar el
contenedor. La expiración de señales en Loki, Tempo y Prometheus se ejecuta en
ciclos y bloques; no implica borrar cada byte exactamente al cumplir 24 horas.

Promtail recoge exclusivamente la API de su propio proyecto Compose y descarta
solo el evento HTTP completado de un GET /healthz o GET /metrics con estado 2xx.
Conserva sus errores, JSON inválido, arranque y eventos de aplicación; los campos
extraídos para el filtro no se convierten en etiquetas de cardinalidad alta.

### Limpieza selectiva de histórico vencido

```bash
make dev-observability-clean
make dev-public-observability-clean
```

El script solo considera los volúmenes de Loki, Tempo, Prometheus y posiciones
de Promtail. Si existe un contenedor de ese servicio o algún archivo modificado
durante las últimas 24 horas, conserva el volumen completo. Lo inspecciona con
un montaje de solo lectura y no borra si falla la comprobación de edad. Nunca
usa una purga global. Reutiliza la imagen Promtail descargada; si falta, conserva
el volumen en vez de descargar una imagen solo para limpiar. Conserva PostgreSQL,
Grafana, Alertmanager y backups.
Con servicios activos, la retención automática hace la purga de señales vencidas.

El 2026-10-03 se retiraron los cuatro volúmenes técnicos antiguos de `dev`
tras verificar que no tenían datos recientes; unos 56 MB. Los de `local` ya
estaban retirados. La base de datos y sus copias no se modificaron. Las imágenes
y contenedores de prueba se retiran después de validar, sin dejar dev arrancado.

Fuentes: [retención de Loki](https://grafana.com/docs/loki/latest/operations/storage/retention/),
[rotación Docker](https://docs.docker.com/engine/logging/drivers/json-file/),
[almacenamiento Prometheus](https://prometheus.io/docs/prometheus/latest/storage/).

## Copia legal incremental

La purga de cuentas y evidencia vencida se ejecuta diariamente a las 03:15. El
backup se ejecuta a las 03:30, después de esa purga. Cada archivo
está cifrado antes de llegar a `iCloud Drive/FastTourney/legal-audit-backups` y
solo incluye filas nuevas o modificadas desde el último punto de control. Si no
hay cambios, no se sube un archivo vacío. La clave local no se sincroniza con
iCloud ni se incorpora a Git.

El estado local del backup relaciona únicamente identificadores técnicos con el
archivo y su fecha de retención. Cuando todos los registros de un archivo han
vencido, el mismo job diario elimina ese archivo cifrado. No guarda emails,
contraseñas, tokens ni contenido legal en claro.

La restauración se ejecuta de forma explícita sobre un destino aislado mediante
`legal-audit-restore`; no se descifra en la carpeta sincronizada ni se mezcla
con datos de la aplicación.
