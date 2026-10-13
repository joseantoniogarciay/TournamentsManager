# Despliegue e infraestructura

Los símbolos y source maps de fiabilidad PostHog para releases de producción
siguen el [runbook cliente](../runbooks/client-error-tracking.md). El proyecto
255144 se reserva a prod; beta/local permanecen sin SDK. El staging web requiere
subida cuando hay clave pública y elimina mapas públicos. La reconstrucción
móvil y prueba real de entrega/simbolización permanecen pendientes.

> Estado: Fase 4 completada en K3s doméstico; Fase 5 AWS cancelada por ADR-0128.

## Progresión

1. un backend modular como unidad desplegable comprensible;
2. builds independientes del cliente universal para web, iOS y Android;
3. dependencias locales en Docker Compose;
4. servicio instrumentado y recuperable;
5. Kubernetes con K3s en una VM Linux local;
6. sin laboratorio cloud previsto; la Fase AWS fue cancelada por ADR-0128.

Cada salto debe justificar qué capacidad añade y qué coste introduce.

## Paridad local-producción

“Parecerse” significa compartir contratos importantes:

- configuración externa;
- imágenes y artefactos equivalentes;
- migraciones;
- health/readiness semantics;
- límites y timeouts;
- señales observables;
- comportamiento ante dependencias no disponibles.

No significa ejecutar localmente todos los servicios gestionados ni reproducir la
topología completa.

## Reglas de infraestructura

- infraestructura reproducible y revisable;
- cambios pequeños con plan de verificación y rollback;
- secretos fuera de estado, planes y logs;
- recursos etiquetados, con propietario y coste visible;
- estado y locking de Terraform decididos antes del trabajo en equipo;
- ningún despliegue sin health checks y criterio de éxito;
- ningún backup sin restauración probada.

## Continuidad de producción

v1 usa el rollout y rollback verificado de K3s; no adopta blue/green. Se
evaluaría una instancia paralela detrás del Ingress y Caddy solo si aparece un
requisito de disponibilidad que el rollout actual no cumpla. ADR-0111 fija la
VM K3s como runtime y la retrospectiva de Fase 4 registra este límite como deuda
deliberada, no como trabajo pendiente.

Durante la conmutación ambas versiones deben ser compatibles con el mismo
esquema PostgreSQL. Los cambios destructivos o incompatibles exigirán una
estrategia explícita de migración (por ejemplo, expand/contract, forward-fix o
restauración), no solo volver a una imagen anterior.

### Apagado ordenado de la API

ADR-0145 añade una espera `preStop.sleep.seconds: 5` ejecutada por kubelet, compatible
con el K3s 1.36.3 documentado y la imagen scratch. La API sigue atendiendo mientras
se propaga la retirada del endpoint. Después SIGTERM cierra el listener y drena
HTTP hasta diez segundos; si vence, se cierran las conexiones y se cancelan los
contextos pendientes. Se cierra PostgreSQL y se vacían trazas hasta cinco segundos.
K3s concede treinta segundos incluyendo preStop; Compose dev usa también treinta.
Los diez segundos restantes en K3s son margen, no un deadline propio de Pool.Close.
Una operación que ignore la cancelación puede alcanzar SIGKILL.

La espera es una hipótesis inicial, no una prueba de propagación de Traefik ni una
garantía de disponibilidad. Validar el rollout bajo tráfico según el
[runbook](../runbooks/k3s-api-image-import.md). No se añaden endpoints de drenaje,
shell ni init al runtime. Eliminaciones forzadas y fallos del host quedan fuera.

## Límite de despliegue desde GitHub

El repositorio es público, pero el acceso operativo no. CI verifica en runners
alojados sin acceso al runtime doméstico. Los despliegues de `dev` y `prod` son
operaciones locales explícitas desde identidades dedicadas; no existe un runner
self-hosted permanente conectado al repositorio público.

La configuración y los secretos de despliegue siguen
[ADR-0017](../adr/0017-use-env-contracts-github-environments-and-oidc.md):
GitHub Environment protegido, secretos mínimos por entorno y OIDC para cloud
cuando esté disponible. Las credenciales persistentes se consideran una excepción
que debe documentarse.

## Límites de entrega del cliente universal

Un árbol de código compartido no implica un único artefacto ni una entrega
acoplada:

- web produce un artefacto desplegable en hosting web;
- iOS produce una aplicación firmada y distribuida por su canal;
- Android produce una aplicación firmada y distribuida por su canal;
- cada target puede ejecutar, publicar y revertir su pipeline de forma
  independiente;
- la compatibilidad con el contrato API debe verificarse antes de publicar,
  teniendo en cuenta que una aplicación instalada no se actualiza de inmediato.

La web se entrega con exportación estática de Expo, sin SSR general. La home `/`
es indexable y el artefacto contiene su HTML, `robots.txt`, sitemap y metadatos;
el artefacto publica además `/favicon.ico` y `/apple-touch-icon.png`; la home
enlaza ambos para navegador, buscadores y Safari. El borde deja rastreables esos
iconos y añade
`X-Robots-Tag: noindex, nofollow, noarchive` a las rutas de aplicación. Solo
`/league/{uuid}` se deriva a un proceso Go local que inyecta
metadatos sociales desde la API pública en el mismo shell estático. Más
superficies dinámicas o SSR completo se decidirán solo si Search Console,
previews o rendimiento aportan evidencia (ADR-0121 y ADR-0120).

La API se empaqueta como imagen OCI conforme a
[ADR-0022](../adr/0022-package-backend-as-oci-image.md). Esta imagen solo
contendrá el backend y mantendrá build y runtime separados. No decide todavía
la firma, el SBOM ni el escaneo reforzado.

## Referencia cloud histórica

ADR-0024 a ADR-0030, ADR-0088 y ADR-0101 conservan el análisis previo de ECR,
Terraform, IAM, red y laboratorios AWS/EKS. ADR-0128 cancela su ejecución en
este proyecto: no hay cuenta, backend remoto, `apply`, laboratorio ni gasto
cloud previsto. Solo una necesidad futura explícita reabriría ese análisis.

## Decisiones por fase

### Desarrollo local

Conforme a [ADR-0076](../adr/0076-run-the-local-api-in-compose-with-air.md),
Compose ejecuta API, PostgreSQL y Mailpit. La API selecciona el target `dev` del
Dockerfile y usa Air con el código montado; el target `runtime` valida el
artefacto sin compilador ni Air y es también la imagen desplegada en K3s para
`prod`. Expo se ejecuta en host por sus simuladores y herramientas nativas.

ADR-0155 acepta mover la web productiva a Workers Static Assets, conservando
API y dev domésticos. El 2026-10-10 se prepara la implementación; la conmutación
todavía no está confirmada. Véase el [flujo Cloudflare](../../infra/cloudflare/README.md)
para candidatos trazables, previews, publicación y recuperación. Hasta verificar
las rutas activas en la cuenta, sigue vigente el hosting doméstico siguiente.

Una beta doméstica usa Cloudflare Tunnel como entrada HTTPS pública (ADR-0090).
El conector del Mac inicia una conexión saliente y alcanza Caddy solo por
loopback; no se reenvían puertos en UniFi ni se publica PostgreSQL o el puerto
interno de la API.

ADR-0089 fija `fasttourney.com` para producción, `dev.fasttourney.com` para
desarrollo, `api.fasttourney.com` para la API de producción y
`dev-api.fasttourney.com` para la API de desarrollo. Caddy sirve los ficheros
`.well-known` correctos en los dos primeros hosts; los hosts de API no participan
en Universal Links ni App Links.

El borde versionado vive en [`infra/home/Caddyfile`](../../infra/home/Caddyfile).
`tournaments-manager-dev` publica la API runtime solo por `127.0.0.1:8081` y
la web exportada de Expo se sirve estática en `dev.fasttourney.com`. Desde el
2026-09-05, `fasttourney.com` sirve el release web de producción; el cambio
revisable entre apertura y cierre sigue siendo importar `production_web` o
restaurar el `503` en Caddy. La configuración,
volumen PostgreSQL y proyecto Compose de dev no se comparten con `local` ni con
el namespace `prod` de K3s, conforme a ADR-0091 y ADR-0111.

El mismo release contiene la referencia pública de desarrollo en
`https://dev.fasttourney.com/api-docs/`. El script de despliegue copia la UI
Scalar y `openapi.yaml` junto a la exportación web; Caddy atiende esa ruta antes
del fallback de la SPA. Como usa el origen ya autorizado
`dev.fasttourney.com`, no cambia CORS de `dev-api` ni exige otro hostname.

ADR-0092 conserva dos despliegues recuperables de dev fuera de Git: cada uno
lleva el SHA, una imagen runtime etiquetada y una exportación web estática. Caddy
sirve el enlace simbólico de la versión activa y el rollback selecciona el SHA
anterior sin tocar PostgreSQL. GitHub Releases y tags no se crean por las
integraciones ordinarias de `develop`; se reservan para producción o hitos
distribuidos. ADR-0119 concreta que un hito usa un tag SemVer anotado sobre el
merge de `main`, una GitHub Release y un artefacto activo del mismo SHA. Esto no
equivale a backup. ADR-0108 añade para `dev` un repositorio
pgBackRest cifrado, copia base, incrementales y WAL archivado con restauración
aislada; véase el [runbook de backup PostgreSQL](../runbooks/postgresql-backup-dev.md).
`prod` tiene PostgreSQL con volumen y repositorio propios, pgBackRest cifrado,
WAL archivado, completa semanal, incrementales diarios y restauración aislada
verificada. Conforme a ADR-0114, el Mac inicia por SSH la réplica cifrada desde
la VM hacia su ubicación doméstica sincronizada; la ejecución programada y la
restauración desde esa réplica se demostraron el 2026-09-18. No se comparte una
carpeta UTM ni se entrega una clave privada del Mac a Kubernetes. Esa ubicación
sigue compartiendo Mac, cuenta y proveedor: no equivale a independencia ante su
pérdida.

Mailpit pertenece solo al entorno local y no tiene hostname público. El entorno
`dev` usa Resend por SMTP autenticado con STARTTLS; antes de invitar personas se
verifica `mail.fasttourney.com` y sus registros SPF, DKIM y DMARC. La clave de
solo envío vive fuera de Git en `infra/dev/api.docker.env`; véase ADR-0093.
`SUGGESTION_RECIPIENT` selecciona por entorno el buzón que recibe los avisos de
sugerencias privadas; no es un secreto ni se fija dentro del binario.

El mismo proyecto `dev` mantiene Prometheus, Alertmanager, Loki, Tempo, Promtail
y Grafana en red y volúmenes propios. Alertmanager usa una segunda clave Resend
de solo envío, montada desde un secreto local, para no compartir el radio de
revocación del correo transaccional. Las interfaces operativas se publican solo
en loopback y no forman parte de Cloudflare Tunnel; véase ADR-0100.

El script del artefacto web de producción declara, sin leer el `.env` local y
limpiando la caché de Metro:

```sh
EXPO_PUBLIC_API_BASE_URL=https://api.fasttourney.com/v1
EXPO_PUBLIC_APP_LINK_URL=https://fasttourney.com
```

La primera URL evita que la web pública contacte servicios locales; la segunda
hace que los enlaces de liga compartidos apunten al dominio público. La API y
la web de producción se abrieron el 2026-09-05 tras validar TLS, CORS y el
release; no supone autorizar los clientes ni asociaciones móviles pendientes.

Al activar el canal de apoyo de [ADR-0129](../adr/0129-accept-voluntary-developer-tips-with-platform-appropriate-payments.md),
`infra/home/secrets/production-web.env` añade los tres Payment Links públicos
de Stripe para 2 €, 5 € y 10 €. El script ya los propaga a la exportación web;
si alguno falta o no es HTTPS, la interfaz no ofrece propinas. No se añaden a
la configuración de builds nativas.

La preparación concreta de la web separa construir de activar:
`infra/home/stage-prod-web.sh` crea un release estático inmutable con SHA,
asociaciones móviles reales y configuración OAuth de producción externa a Git;
`activate-prod-web.sh` conmuta solo su enlace simbólico y
`rollback-prod-web.sh` vuelve a un SHA existente. El bloque Caddy
`production_web` conserva CSP, SPA y `/.well-known`; activarlo o volver al `503`
requiere una operación explícita y verificable. El
[runbook de publicación](../runbooks/production-web-publication.md) exige TLS
válido, Tunnel sano, CORS, correo, Google y recorridos controlados antes de esa
conmutación.

### Fase 4

VM Linux de un nodo con K3s como runtime doméstico de `prod`: manifests,
empaquetado, recursos, probes, configuración, secretos, persistencia, backup,
ingress, rollout y recuperación, conforme a ADR-0111.

ADR-0117 separa la administración de la VM de la identidad humana inicial:
`fasttourney-operator` entra por clave SSH dedicada en la red privada y recibe
`sudo` no interactivo para operar host, K3s y `kube-system`. La contraseña de
Ubuntu no se almacena en el Mac ni en Git; el bootstrap único permanece
interactivo. Véase el runbook de administración remota antes de depender de
operaciones no asistidas.

ADR-0112 concreta el orden de empaquetado: API, PostgreSQL y los recursos de
core se definen primero como manifiestos YAML propios aplicados con `kubectl`.
La observabilidad de terceros se añade después mediante Helm, con charts y
valores versionados; no se instala inicialmente un operador ni un chart propio
de toda la plataforma.

### Fase 5 — cancelada

Terraform, cuenta AWS, identidad, bootstrap y laboratorio EKS no se activarán
en este proyecto. Este material se conserva como referencia histórica y solo se
reabrirá mediante la decisión explícita y el análisis de coste de ADR-0128.

## Publicación de deportes e incidencias — 2026-10-03

El release v1.9.0 publica en producción el SHA
`69533ac6cae53bf7245248221cf8143c13329f40`. Dev conserva la revisión funcional
verificada `8fe7be1d7f998557051cdbe1bf72fcc2b50630ce`. CI de la propuesta,
develop y main terminó en verde. Goose aplicó 00015–00019 antes de cada API;
producción tiene dos réplicas listas, web y renderer del mismo SHA y el secreto
efímero de migración eliminado. Web/API de ambos entornos devuelven HTTPS 200;
CORS productivo devuelve 204 con su origen. Los diez torneos preexistentes de
producción siguen legibles y sus conteos de cuentas/torneos permanecen 3/10.
La pantalla productiva muestra los ocho deportes y bádminton 21/15 en móvil.

Dev verificó trece torneos sintéticos por HTTP y la lectura de abandono de
pádel en navegador; la cuenta y los torneos se eliminaron exclusivamente al
cerrar. No se enviaron correos ni se alteraron resultados reales de producción.
Se tomaron backup incremental cifrado de producción y copias lógicas privadas;
la restauración aislada y el ensayo de las cinco migraciones sobre datos de
ambos entornos terminaron correctamente. El fallo de lectura de pgBackRest
de dev en iCloud estaba pendiente al publicar: su copia lógica local probada
permitió continuar. Posteriormente, el 2026-10-03, se recuperó la lectura de
los metadatos desde macOS y se verificaron WAL, una copia completa nueva, la
incremental mediante LaunchAgent (salida 0) y su restauración aislada con
esquema 19 y agregados coincidentes. Véase el
[incidente y la prueba reproducible](../runbooks/postgresql-backup-dev.md).
Tras autorización explícita posterior del usuario, se activó Conservar en
dispositivo únicamente para dev. Finder confirmó dev, archive y backup
conservados en el dispositivo, sin descarga pendiente; repositorio cifrado y archivado WAL se volvieron a verificar
correctamente desde Docker.

Recuperación: web v1.8.3 `91bcbcd24e6c7bb066f831045456f5a31d1bb740`;
API anterior `f6a199ad42bd7e535db9905b132cd75d69486396`. Las migraciones son
forward-only. Una vez registradas incidencias, el cliente antiguo no las
interpreta; se prefiere corrección hacia delante, sin restaurar datos de forma
automática ni perder escrituras posteriores. No se distribuyen apps nativas;
el teclado nativo sigue pendiente en dispositivo. Los cambios locales de
retención/observabilidad no se incluyen en este release.

Retrospectiva: aislar desde la base remota preservó seguridad publicada y
trabajo pendiente. Probar migraciones sobre restauraciones reales evita
confundir una base de fixtures con compatibilidad de los datos existentes.

### Estabilización autorizada de los montajes de dev

Con autorización explícita posterior del usuario, el checkout de despliegue
se conserva en `.config/deployments/v1.9.0`, ignorado por Git. Se recrearon solo
Prometheus, Loki, Promtail, Tempo, Alertmanager y Grafana con `--no-deps`,
las mismas imágenes y sus volúmenes. Los seis servicios quedaron sanos y todos
sus montajes propios apuntan a esa ubicación estable; no se reinició API ni
PostgreSQL. Ambas APIs conservaron HTTPS 200. La revisión automática había
bloqueado el traslado inicial; la operación se completó tras esa autorización,
sin sobrescribir los archivos pendientes del workspace original.

### Promoción del bloque operativo de dev — 2026-10-03

Tras autorización del usuario, `b3a5e93` se publicó en `origin/develop` y pasó
[la CI completa](https://github.com/joseantoniogarciay/TournamentsManager/actions/runs/37138692151),
incluida integración PostgreSQL. Su configuración operativa se promovió desde
`.config/deployments/ops-b3a5e93`, ignorado por Git, conservando la copia anterior.
Se recrearon los ocho servicios para aplicar rotación y retención de ADR-0139,
con las mismas imágenes y volúmenes; la API conserva
`tournaments-manager-dev-api:git-8fe7be1d7f998557051cdbe1bf72fcc2b50630ce`.
No hubo rebuild de aplicación, migración ni cambio de web estática o renderer.

La retención efectiva es de 24 h en Loki y Tempo, y de un día o 128 MiB TSDB
para Prometheus; Loki completó un ciclo de retención. Los ocho servicios están
activos, Loki y Tempo listos y ambas APIs públicas y la web dev devuelven `200`.
Los agregados de PostgreSQL coinciden antes/después y `pgbackrest check` pasa.
La [revisión operativa](OPERATIONAL_BLOCK_REVIEW_2026-10-03.md) recoge los límites
y el manifiesto privado. No cambia el release v1.9.0 de producción.

## Promoción y retrospectiva — 2026-10-04

Se despliega la revisión limpia `11138a1100db853d04a2e388bae029c013f55843`,
conservada localmente en la rama `ops/api-graceful-shutdown`. Se aisló el cambio
sin incluir las modificaciones en curso del cliente. El wrapper existente
importó runtime/migrator ARM64 y verificó API y renderer del mismo SHA. Goose
confirmó versión 19 sin migraciones pendientes; el Secret efímero fue eliminado.
K3s aceptó el manifiesto en dry-run server-side y el Deployment aplicado confirma
30s de gracia, preStop.sleep 5s y dos réplicas listas con cero reinicios.

Se realizó un reinicio gradual controlado para verificar pods que ya tenían el
hook nuevo. Una primera medición con urllib recibió 403 del borde y se descartó:
no acreditaba acceso a la API. Se repitió con curl y GET /v1/sessions sin sesión,
que devuelve el 401 esperado. Durante 45,61 segundos se observaron 168 respuestas
401, cero errores de transporte y cero respuestas inesperadas, incluido el
rollout de verificación. Ambos pods salientes registraron inicio y finalización
del drenaje, sin timeout ni eventos FailedPreStopHook. La web HTTPS devuelve 200;
el renderer local publica el SHA desplegado.

La prueba es muestreo de una ruta sin autenticar, no una garantía de downtime
cero ni una prueba de transacciones largas bajo carga. Estas últimas están
cubiertas para HTTP por la prueba local con sockets; PostgreSQL bajo carga y un
fallo del host conservan sus límites documentados. No se publicó una nueva web,
un tag, un PR ni una GitHub Release. El siguiente merge debe incorporar esta
rama para mantener alineados código y producción.

**Aprendizaje:** una respuesta del borde no prueba que el backend atendiera la
petición. Validar primero ruta y respuesta esperada; el primer rollout instala
el nuevo hook, y para observarlo hay que retirar después un pod que ya lo tenga.

### Gate de asociaciones HTTPS para releases móviles

`FASTTOURNEY_REQUIRE_APP_LINKS=1` exige el par AASA/DAL al preparar la web dev o
prod. Se valida su identidad de entorno, cobertura AASA y formato de cada
huella antes de exportar; se copia dentro del release para conservar el rollback.
La propiedad de firma se comprueba contra las apps realmente distribuidas.
Caddy sirve los recursos como JSON fuera del fallback SPA. El [procedimiento](../../infra/app-links/README.md)
detalla carpetas, consultas HTTPS y validación en iOS/Android; no se considera
completado el lanzamiento móvil hasta superar las tres comprobaciones.

### Desarrollo y observabilidad bajo petición — ADR-0146

Local/dev permanecen apagados fuera de pruebas. Los comandos ordinarios de
arranque, despliegue y rollback no activan los seis servicios de observabilidad;
Compose los agrupa en `observability`, sin reinicio automático, y la API no
exporta trazas mientras estén apagados. Solo una petición explícita autoriza
`make dev-observability-up` o `make dev-public-observability-up`, con API ya activa.
Los comandos `*-observability-down` cierran el diagnóstico y `dev-down` /
`dev-public-down` cierran las pruebas conservando datos. Público suspende sus
LaunchAgents de renderer, purga y backups y los rehabilita al arrancar pruebas.
Producción/K3s conserva sus servicios y controles.

### Acceso Apple preparado — ADR-0147

La migración 00020 admite identidades Apple y evidencia legal de su alta, y crea
challenges cortos con state/nonce/prueba ligados al cliente. Se aplica antes de
promover la API y fuera de su arranque; no se ha desplegado en esta sesión.
Las plantillas API/cliente contienen placeholders inertes. La clave .p8 se monta
solo en el servidor como archivo de lectura; el bundle cliente recibe únicamente
el Services ID. El callback HTTPS por entorno es `/v1/apple-callback`.

No basta con rellenar Team ID: se registran Services ID/App ID, dominios, callback,
relay de email y Google OAuth móvil, y se validan builds reales. El [runbook](../runbooks/social-login.md)
conserva los gates, incluida revocación Apple al eliminar cuenta. La preparación
no activa Apple, no despliega dev/prod y no autoriza distribución.

## Preparación autorizada de v1.10.0 — 2026-10-08

El usuario autoriza actualizar dev y prod con el bloque acumulado. La revisión
de develop a651864 tiene CI completa aprobada, incluidas integraciones PostgreSQL,
y make verify local aprobado con 78 pruebas Node adicionales. El alcance es web
y API; no distribución móvil ni activación de proveedores OAuth pendientes.

Antes de promover se detecta la VM UTM de producción detenida: API pública 502
y web 200. Arrancar la instancia existente recupera K3s, dos réplicas API y
healthz 200 sin cambiar la versión instalada. Se conservan las versiones
anteriores de web, API y manifiestos para rollback. Las copias incrementales
pgBackRest y copias lógicas privadas preceden a la migración 20. El ensayo sobre
restauraciones aisladas verifica la nueva tabla, grants runtime y conservación
de cuentas, torneos, equipos y partidos. La migración es forward-only: un
rollback de aplicación no elimina identidades ni evidencia legal.

Dev se enciende para desplegar y probar y volverá a apagarse al cerrar, conforme
a ADR-0146; observabilidad dev permanece apagada. Producción conserva sus
servicios operativos. La release productiva sigue ADR-0119: merge no-ff a main,
tag anotado, GitHub Release y artefactos del SHA etiquetado.

### Cierre de v1.10.0 — 2026-10-08

Publicada [v1.10.0](https://github.com/joseantoniogarciay/TournamentsManager/releases/tag/v1.10.0)
sobre el merge no-ff `b002b4b99fc44a6ec0662b94ec962b8cce9cbea5` de main.
La [CI del SHA productivo](https://github.com/joseantoniogarciay/TournamentsManager/actions/runs/37818973693)
y la [CI de preparación en develop](https://github.com/joseantoniogarciay/TournamentsManager/actions/runs/37818687061)
terminan aprobadas. Develop avanza por fast-forward hasta el merge antes del
cierre documental. No se crean ramas temporales.

- Dev: API, web y renderer `08b7deb3a2def89a11fc8e53600a242bf424fb18`;
  migración Goose 20 y ocho comprobaciones HTTP/TLS/CORS/revisión correctas.
- Prod: API, migrator, web y renderer `b002b4b99fc44a6ec0662b94ec962b8cce9cbea5`;
  dos réplicas API listas, Goose 20 y nueve comprobaciones públicas correctas.
  Caddy versionado validado y recargado; Secret migrador efímero eliminado.
- Ambos: sesión y biblioteca anónimas devuelven 401, torneo inexistente 404,
  cuerpos JSON válidos. Comparación de recuentos y huellas completas de filas
  de cuentas, torneos, equipos y partidos coincide con las copias previas.
  La identidad runtime productiva conserva los grants de la tabla Apple.
- La preparación productiva ejecuta por primera vez la subida real de sourcemaps
  mediante PostHog CLI; no quedan `.map` en la web pública. Esto no acredita
  recepción de un error real ni simbolicación en el panel.

Backups pgBackRest previos: dev `20261004-014505F_20261008-174057I` y prod
`20261004-021506F_20261008-174058I`, ambos con estado correcto. Copias lógicas,
manifiestos anteriores, logs y comprobaciones permanecen privados en
`/private/tmp/tm-release-20261008`; no son un archivo duradero ni se versionan.
El rollback de aplicación conserva web `69533ac6cae53bf7245248221cf8143c13329f40`
y API `11138a1100db853d04a2e388bae029c013f55843`; no revierte la migración.

Al cerrar se ejecuta `make dev-public-down`: API, PostgreSQL y renderer dev
apagados, seis LaunchAgents descargados y volúmenes conservados. Observabilidad
dev no se inicia. Producción/K3s y su renderer permanecen activos. Dev queda
actualizado para el siguiente arranque autorizado, no permanentemente encendido.
Apple real, distribución móvil y los límites registrados de QA siguen pendientes.

Retrospectiva: verificar la cuenta activa de `gh` antes de atribuir el rechazo de
una Release a permisos ausentes. Había otra cuenta activa; usar la sesión ya
guardada de la propietaria resuelve la publicación sin ampliar scopes. Separar
salud pública, revisión desplegada, migración y conservación de datos permite
cerrar el deploy con evidencia concreta sin convertirlo en un QA universal.

### Suspensión productiva autorizada — 2026-10-08

Después del despliegue verificado, el usuario solicita reducir consumo y confirma
explícitamente «Apagar producción/K3s, aceptando que la API pública quede fuera
de servicio». Esta instrucción autoriza excepcionalmente el apagado de esta
sesión frente a la regla ordinaria de conservar producción activa de ADR-0146;
no modifica durabilidad ni borra recursos.

Se solicita apagado del SO invitado con `utmctl stop --request` para
fasttourney-k3s-lab (9494026F-7DD1-4FFC-9229-851A9391DBB5); UTM confirma stopped.
Se deshabilitan y descargan los LaunchAgents del Mac
com.fasttourney.prod-league-preview-renderer,
com.fasttourney.prod-postgresql-backup-full,
com.fasttourney.prod-postgresql-backup-incremental y
com.fasttourney.utm-autostart. Así no se reenciende la VM automáticamente ni se
siguen lanzando tareas contra una BD apagada. Se conservan plists, VM, discos,
volúmenes, releases y backups. Local/dev se apagan al cerrar el QA.

La API pública devuelve 502 esperado mientras está suspendida. La home estática
sigue respondiendo 200 mediante Caddy/Tunnel; no representa disponibilidad del
producto ni de sus previews de torneo. No se cambia el router ni el túnel.
Producción permanece apagada hasta una nueva petición explícita.

Para reactivar, con autorización: arrancar la VM existente con `utmctl start`
y esperar K3s/PostgreSQL y ambas réplicas API; comprobar healthz público. Después
rehabilitar con `launchctl enable` y cargar con `launchctl bootstrap` los tres
plists productivos anteriores del directorio ~/Library/LaunchAgents. Solo si se
quiere recuperar el arranque automático del Mac, rehabilitar también el plist
com.fasttourney.utm-autostart. Comprobar renderer 8091, revisión web/API
b002b4b99fc44a6ec0662b94ec962b8cce9cbea5 y programación de backups, sin redesplegar
ni migrar por el mero hecho de arrancar. La CI del cierre documental f1e59e0
termina aprobada antes de esta suspensión.

Retrospectiva: apagar producción es una decisión de disponibilidad, no solo de
consumo. Registrar consentimiento, estado público esperado y tareas suspendidas
hace reversible la operación sin confundir una API 502 deliberada con regresión.

### Recuperación productiva y hosting de borde — 2026-10-11

El usuario autoriza completar ADR-0155 y levantar K3s. Se recupera la VM
existente, PostgreSQL y ambas réplicas API con revisión `b002b4b…`, sin
redesplegar ni migrar. API pública 200 y CORS productivo 204. Se rehabilitan
renderer de rollback y los dos LaunchAgents de backup con programación original;
UTM autostart permanece deshabilitado. Dev/local siguen apagados; no se inicia
observabilidad dev. Los recursos de observabilidad de producción preexistentes
reanudarán junto con su cluster.

Web productiva conmutada a Workers Static Assets Free usando el mismo bundle.
Caddy/Tunnel web se conserva como origen de rollback; API y DNS no cambian.
Ver versiones activas, pruebas y retrospectiva en
[publicación web](../runbooks/production-web-publication.md). La home consume
una ejecución Worker para conservar su indexación; el resto de assets sigue
estático. Un certificado de interceptación del operador no se interpreta como
fallo de emisión Cloudflare: se contrasta la huella y la conexión.

### Publicación v1.11.0 — 2026-10-11

[Release v1.11.0](https://github.com/joseantoniogarciay/TournamentsManager/releases/tag/v1.11.0)
publicada sobre el merge no-ff `ec8112a383fadd2abb1bf0d1800406e01c5a761a`.
Incluye responsive desktop, accesibilidad/scroll, Expo 57.0.26 / RN 0.86.3,
Go 1.26.9 y hosting Cloudflare Free. No distribuye apps nativas en tiendas.

`make verify` local, integración PostgreSQL efímera y prueba race de apagado
pasan. La CI del candidato
[e2e1d57](https://github.com/joseantoniogarciay/TournamentsManager/actions/runs/38118422034)
pasa antes de promocionar. La [CI del SHA etiquetado](https://github.com/joseantoniogarciay/TournamentsManager/actions/runs/38118777031)
también termina aprobada. El merge conserva exactamente el árbol del candidato.
Web/infra, API, migrator y renderer de rollback tienen el SHA etiquetado; dos
réplicas API listas. Goose confirma versión 20 sin migraciones pendientes.
Backup cifrado `20261004-021506F_20261011-063119I` replicado y restaurado desde
la copia recibida, en un volumen aislado eliminado al terminar. Huellas completas
de cuentas/torneos/equipos/partidos coinciden antes/después (3/10/26/21 filas).
37 peticiones de healthz durante el rollout devuelven 200, sin fallos observados.
Secret migrador efímero eliminado; no se alteran PostgreSQL ni su durabilidad.

Nueve comprobaciones HTTP/TLS/CORS/manifiesto/renderer pasan. Biblioteca y
métodos de acceso anónimos devuelven 401 JSON; torneo ausente 404 JSON. Web
conserva 200 con metadatos para torneo público y 404 para UUID ausente, fallback
de subruta y rechazo de JS inexistente. Home indexable y preview noindex.
Navegador productivo: 1600 px muestra columnas 590/590 y tabs centradas 1200;
390 px vuelve a una columna, sin overflow de página. No se repite login
autenticado ni envío de correo; el QA previo acotado conserva sus límites.
Sourcemaps se suben al proyecto PostHog autorizado y se retiran del export; no
acredita por sí mismo recepción ni simbolicación de un error real.

Workers actuales y rollback en `infra/cloudflare/README.md`. API previa b002b4b
y manifiesto capturado conservados; origen Caddy anterior disponible para
recuperación. Evidencia operativa privada en `/private/tmp/tm-release-20261011`.
Dev/local siguen apagados. Develop avanza hasta el merge productivo; no se
crean ramas temporales. No se publican cambios de dev doméstico en esta sesión.

Retrospectiva: un solo SHA para los artefactos evita mezclar la migración de
hosting con bundles anteriores. Preparar y verificar antes de activar mantiene
rollback disponible; medir lecturas durante rollout y contrastar huellas aporta
evidencia de disponibilidad y conservación sin mutar fixtures productivos.

## 2026-10-11 — PostgreSQL, pgBackRest y Caddy

ADR-0157 autoriza PostgreSQL 18.6 y pgBackRest 2.59.3 en producción. Imagen
`tournaments-manager-postgresql:18.6-pgbackrest-2.59.3`,
checksum oficial del fuente pgBackRest verificado durante build. Se conserva
la imagen anterior y ambos PVC. Restauración previa y posterior satisfactoria
en emptyDir, repositorio de backups de solo lectura y sin montar datos activos.
El backup completo nuevo se crea con `--no-expire-auto`; no se expiran copias.
`pgbackrest check` pasa, PostgreSQL está Ready y API dentro de K3s devuelve 200.
Se exporta copia cifrada adicional fuera de la VM. Evidencia privada:
`/private/tmp/tm-infra-upgrade-20261011`.

Caddy 2.11.7 instalado desde Homebrew, configuración validada y servicio
`sh.brew.caddy` activo con binario nuevo. Durante el reinicio se observó 502
y «no route to host» desde launchd, aunque curl y K3s funcionaban. Volver al
binario y etiqueta anteriores tampoco recuperó inmediatamente el acceso.
Después de la comprobación del permiso de Red local por el usuario, la API
volvió a 200; no se atribuye causa definitiva solo por esa correlación.
La revisión automática bloqueó inicialmente reactivar la versión nueva;
tras recuperar 200 autorizó completar la transición. API pública, origen
y web comprobados con 2.11.7 activo. Se conserva binario 2.11.4 para rollback.

Retrospectiva: probar restauración antes de cambiar el único Pod permite
comprobar compatibilidad real. Validar configuración y binario de un proxy
no sustituye probar su conectividad desde el proceso gestionado por launchd.
Local/dev siguen apagados; observabilidad productiva existente no se modifica.

### Conciliación Git del hito operativo

Se promociona el bloque de infraestructura a main mediante merge explícito
y tag v1.11.1. Esta release registra PostgreSQL/pgBackRest y Caddy activos;
web y API conservan sus artefactos v1.11.0, sin reconstrucción ni despliegue
adicional. Las valoraciones nativas permanecen en develop, pendientes de
distribución. El despliegue operativo precedió al commit en esta sesión:
la conciliación corrige esa pérdida temporal de trazabilidad.

Retrospectiva: las siguientes actualizaciones deben guardar y promocionar
el bloque verificado antes de cambiar producción, conforme a ADR-0119.

## Publicación v1.12.0 — 2026-10-11

El usuario autoriza publicar web y backend, continuar pese al certificado ajeno
recibido en su conexión y subir los mapas fuente a PostHog UE 255144.
Tag v1.12.0, merge main `3a9d41086ce158daf145a2930c477e4fa213244f`.
Incluye repetición de torneos, elección de equipos libres en invitaciones, aviso
de plaza ocupada y fecha/hora/lugar opcionales (ADR-0158/0159/0160).
No distribuye builds nativas ni incorpora los parches jóvenes de Expo.

`make verify` local y CI de develop/main pasan, incluida integración PostgreSQL
en CI. Backup incremental cifrado replicado y restaurado en un volumen temporal
con repositorio solo lectura; Job de ensayo eliminado al terminar. Goose aplica
00021/00022 y confirma versión 22. API con dos réplicas listas en el SHA del tag;
secreto migrador efímero retirado. Huellas de datos anteriores coinciden antes y
después, excluyendo las columnas nuevas (cuentas/torneos/equipos/partidos: 3/10/26/21).

El artefacto definitivo conserva PostHog de fiabilidad; la CLI termina
correctamente, vincula los mapas al SHA y los retira del export. No demuestra
recepción ni simbolicación de un error real. Candidato previo sin PostHog
conservado fuera de publicación; nunca se activó en Cloudflare.

Cloudflare confirma ambos triggers y versiones activas:

- web: `e97c2239-325e-4b46-9076-6206b772cad0`;
- www: `97d3f0e0-0f4b-400b-b52b-574d007c43f5`.

Preview definitivo: TLS válido, deployment.json con SHA correcto y home/cuenta
noindex. Cinco checks del origen API pasan (health, 404, 401 de repetición y
lugar, CORS); renderer con SHA correcto. Origen web de rollback actualizado al
mismo SHA: cinco rutas correctas, cuenta noindex y ningún mapa público.
La verificación HTTPS de fasttourney.com/api desde la red del operador queda
pendiente: certificado autofirmado core1.netops.test con huella
D5:EB:30:1E:5A:00:4C:8B:6C:39:36:A5:DA:81:B1:E8:FC:F6:D7:7F,
coincidente con el incidente anterior. No se omite TLS ni cambia DNS/certificados.

Rollback: API y web v1.11.0 conservadas; Workers previos
`05f8a02a-d89b-4811-844e-4e97323c183e` y
`6c27d953-97e8-436c-b231-b9f50bb6e772`. Las migraciones aditivas permiten
volver al código anterior sin Down; evaluar pérdida antes de retirar columnas.
Evidencia privada: `/private/tmp/tm-release-v112-20261011`. Dev/local apagados.

Retrospectiva: distinguir fallos del proveedor, bloqueos de aprobación del
contenido enviado y limitaciones de la red. Publicación confirmada por el plano
de control y comprobación del origen no equivalen al recorrido HTTPS completo
desde todas las redes; conservar explícitamente ese límite.

### Arranque productivo solicitado — 2026-10-12

El usuario solicita levantar producción con UTM. Se arranca la VM existente
`fasttourney-k3s-lab`; UTM confirma `started`. La comprobación remota con la
configuración operativa existente confirma nodo `Ready`, API `2/2` y PostgreSQL
`1/1`. Tras la recuperación, API pública `/healthz` y web responden HTTP 200.
El renderer responde en `/-/ready` con revisión
`3a9d41086ce158daf145a2930c477e4fa213244f`, coincidente con la imagen API.
No se redespliega, migra ni publica el trabajo pendiente del workspace.

Renderer y LaunchAgents productivos de backup ya estaban habilitados y cargados:
se conservan sus calendarios (full domingo 04:15, incremental lunes a sábado
04:15). El último estado de salida histórico del incremental no es cero; esta
comprobación confirma programación, no ejecución exitosa de un nuevo backup.
UTM autostart permanece deshabilitado y dev/local permanecen apagados.

Retrospectiva: arrancar la VM y esperar readiness permite recuperar el release
existente sin convertir una solicitud de disponibilidad en una publicación.
Se verifica el servicio público después del estado del cluster, pues durante
el arranque el borde devuelve temporalmente 502.

## Publicación v1.13.0 — 2026-10-12

El usuario autoriza publicar las features y fixes pendientes con el proceso
completo. Tag anotado v1.13.0 y merge main
`ef4b1776866a17c96bd504e44e00a57698879606`. Incluye Mis partidos y Ediciones
(ADR-0161/0162), revelación de destinos, conservación de páginas, acceso a Cuenta,
selección accesible web y retorno de foco nativo. No distribuye binarios de tienda.

`make verify` local y CI develop/main pasan; integración PostgreSQL real en CI.
Se corrige únicamente el formato de dos JSON de evidencia detectado por el gate.
Dev se despliega desde el candidato limpio, pasa ocho checks públicos/origen de
lectura, autenticación, límites y CORS, y su UI abre Cuenta y Ediciones. El primer
smoke urllib recibe 403 en el borde mientras curl y origen responden 200; el mismo
smoke con user-agent ordinario pasa, sin omitir TLS ni cambiar protección del borde.

Backup incremental cifrado replicado y restaurado desde la copia recibida, montada
solo lectura, en volumen Docker aislado: `fasttourney_prod|f`; ensayo limpiado.
Goose confirma versión 22 sin nuevas migraciones. API 2/2 y renderer coinciden
con el SHA del tag, y el secreto migrador efímero queda eliminado. Huellas completas
de siete tablas coinciden antes/después; cuentas/torneos/equipos/partidos: 3/10/26/21.

Web conserva la configuración de fiabilidad y símbolos PostHog UE existentes:
preparación completada, mapas retirados del artefacto público. No demuestra recepción
ni simbolicación de un fallo real. Preview TLS válido y cinco checks pasan.
Cloudflare confirma triggers productivos y versiones:

- web: `9dc57060-fa07-40db-9100-fedd566c44fe`;
- www: `b8a591b7-7127-4114-888a-a43a32d7270b`.

Tras publicar pasan ocho checks API y once web por HTTPS: SHA, home indexable,
Mis partidos/Cuenta/legales/Ediciones noindex, torneo existente, inexistente 404,
asset/asociación inexistentes 404 y www 308 conservando query. Navegador confirma
Mis partidos, acceso a Cuenta y Ediciones con datos reales existentes, sin mutarlos.
Origen Caddy de rollback actualizado al mismo SHA. No se certifica login autenticado,
nuevo envío de correo ni locución/gestos pendientes de la matriz nativa.

El aviso de correo de registro se diagnostica en la versión anterior: la dirección
indicada ya tenía cuenta verificada, no se creó cuenta ni token nuevos. El alta
repetida devuelve 202 genérico sin envío por la política contra enumeración.
No demuestra un fallo SMTP ni exige cambiar autenticación. No se publica la dirección.

Rollback conservado: API/web v1.12.0, manifiesto y binario renderer anterior,
Workers web `e97c2239-325e-4b46-9076-6206b772cad0` y www
`97d3f0e0-0f4b-400b-b52b-574d007c43f5`. No requiere Down ni restaurar datos para
volver al código previo. Evidencia privada: `/private/tmp/tm-release-v113-20261012`.
Dev/local y sus tareas programadas apagados; producción permanece activa y UTM
autostart conserva su estado deshabilitado.

La revisión automática rechaza las notas iniciales de Release por detalle interno
sin destino acreditado. Se verifica repositorio público propio con permiso ADMIN
y se reduce el texto a producto, SHA, validación, límites y versión de rollback;
la creación de Release queda confirmada sin ampliar permisos ni omitir controles.

Retrospectiva: una respuesta genérica de alta repetida no prueba entrega de correo.
Separar el diagnóstico de identidad del despliegue evita introducir fixes sin causa.
Un release debe conservar la cadena tag–SHA–artefactos y sus límites de QA: las
capturas nativas no acreditan publicación en tiendas ni accesibilidad completa.

## Publicación web v1.13.1 — 2026-10-12

El usuario solicita desplegar la retirada del título de Mis partidos. Tag anotado
v1.13.1 sobre merge no-ff `9c9d01e5f26a46078265c77b17b65ff48f5791a4`;
[Release](https://github.com/joseantoniogarciay/TournamentsManager/releases/tag/v1.13.1).
Solo se publica web; API, renderer y PostgreSQL conservan v1.13.0 y esquema 22.
No se ejecutan migraciones ni se distribuyen binarios nativos.

`make verify` local pasa; integración PostgreSQL local omitida por ausencia de
base efímera. La [CI del candidato](https://github.com/joseantoniogarciay/TournamentsManager/actions/runs/38268949626)
y la [CI del tag](https://github.com/joseantoniogarciay/TournamentsManager/actions/runs/38269633582)
pasan con PostgreSQL real. El merge conserva exactamente el árbol del candidato.
Exportación productiva, preparación de símbolos y dry-run de ambos Workers pasan;
los mapas fuente se retiran del artefacto público. No acredita recepción ni
simbolicación de un error real.

Preview `f2cc92ee-669b-48f6-9b26-01eb689e13df`: diez checks HTTPS pasan.
Workers activos web `9a8afff4-913d-472a-abeb-25a2b089b96f` y www
`b1bbb26c-9dc2-408e-a4af-6ea2d2f93e06`. Diez checks públicos verifican SHA,
home indexable, colecciones/cuenta/legales noindex y 404 de torneo UUID v4,
asset y asociación ausentes. API health 200 y www 308 conservando query.
El navegador confirma entrada sin título adicional en Mis partidos sin sesión;
no se repite login autenticado ni QA nativa. El origen Caddy de recuperación
apunta a la misma web. No cambia DNS ni plan Free.

Rollback: web v1.13.0 con Worker `9dc57060-fa07-40db-9100-fedd566c44fe`,
www `b8a591b7-7127-4114-888a-a43a32d7270b` y artefacto local
`ef4b1776866a17c96bd504e44e00a57698879606`. No requiere restaurar datos.
Evidencia operativa en `/private/tmp/tm-v1131-*.log` y
`/private/tmp/tm-ui-header-*.log`. Develop vuelve a contener el merge productivo;
no se crean ramas temporales. Local/dev permanecen apagados y producción activa.

Retrospectiva: una corrección visual puede entregarse de forma independiente
sin reconstruir la API ni tocar persistencia. Conservar versiones diferentes
por componente y el manifiesto público evita atribuir un nuevo SHA al backend.

## Publicación de seguridad v1.13.2 — 2026-10-12

El usuario autoriza desplegar las correcciones de ADR-0163. Tag anotado
v1.13.2 y merge no-ff `28a9dfa51fed73437dcce09db9d73d486105d9ef`,
[Release](https://github.com/joseantoniogarciay/TournamentsManager/releases/tag/v1.13.2).
API, renderer, web Cloudflare y origen Caddy conservado coinciden con ese SHA.
No distribuye binarios nativos ni añade migraciones; Goose confirma esquema 22.

`make verify` local y la [CI del candidato](https://github.com/joseantoniogarciay/TournamentsManager/actions/runs/38293272215)
pasan; PostgreSQL real se verifica en CI y en la suite local separada con -race.
La [CI del merge productivo](https://github.com/joseantoniogarciay/TournamentsManager/actions/runs/38293939396)
también pasa.
La VM estaba detenida y el health público devolvía 502. Se arranca UTM sin
cambiar autostart; nodo Ready, PostgreSQL y API recuperados antes del rollout.

API queda 2/2 Ready sin reinicios nuevos; renderer readiness identifica el tag.
Job migrador completado y Secret efímero eliminado. Durante el rollout las
35 lecturas HTTPS de healthz devuelven 200, sin fallos de transporte ni 5xx.
No se conserva evidencia del log de drenaje de los pods antiguos ya eliminados;
la continuidad observada no sustituye esa evidencia ni prueba carga sostenida.
Ocho checks API públicos verifican lecturas, autenticación, validación y CORS.
Pruebas adicionales con tokens ficticios verifican rechazo 400 de cuerpo >64 KiB
y token >512 bytes; 22 confirmaciones inválidas dan veinte 400 y dos 429 con
Retry-After. Las cuotas son por proceso: dos réplicas pueden admitir veinte
solicitudes por minuto en total. Respuestas problem+json seguras y no-store.
Las huellas completas de siete tablas de negocio coinciden antes/después:
cuentas/torneos/equipos/partidos 3/10/26/21. No se mutan cuentas ni torneos reales.

Dry-run, preview HTTPS y once checks web públicos pasan: SHA, home indexable,
colecciones/cuenta/legales noindex, torneo existente, inexistente 404, asset y
asociación ausentes 404, www 308 conservando query. Versiones activas:

- web `a588241e-4707-4753-a9b3-e227d71c9a61`;
- www `953e1e7c-67ea-4e01-9dcd-e23ab08bf3a0`;
- preview `f205d3ea-dd2f-4536-b8a7-49889c7f7791`.

Rollback: API/renderer v1.13.0 `ef4b1776866a17c96bd504e44e00a57698879606`
con manifiesto y binario anteriores conservados; web/origen v1.13.1
`9c9d01e5f26a46078265c77b17b65ff48f5791a4`, Workers web
`9a8afff4-913d-472a-abeb-25a2b089b96f` y www
`b1bbb26c-9dc2-408e-a4af-6ea2d2f93e06`. No requiere Down ni restaurar datos.
Se conserva el backup cifrado de hoy restaurado en aislamiento antes de v1.13.0.
La revisión automática rechazó crear/exportar otra copia a iCloud por requerir
consentimiento específico para ese traslado sensible; no se ejecutó ni se rodeó
el rechazo. Esta promoción de código sin DDL usa el backup probado existente y
rollback de artefactos. No se afirma haber creado un backup nuevo.

Evidencia operativa privada: `/private/tmp/tm-release-v1132-20261012` y
`/private/tmp/tm-v1132-verify.log`. No se repite login real, envío de correo,
QA visual nativa ni prueba de saturación de Argon2 en producción. La concurrencia,
revocación, límites SQL y volumen se prueban en aislamiento. Local/dev y sus
LaunchAgents permanecen apagados; no se activa observabilidad dev ni cambia DNS,
plan Cloudflare o durabilidad PostgreSQL. Develop vuelve a contener el merge;
no se crean ramas temporales.

Retrospectiva: recuperar el runtime y desplegar un artefacto son comprobaciones
distintas. Probar límites con tokens ficticios acredita el contrato público sin
alterar identidades reales. Las cuotas en memoria deben documentarse por réplica;
el smoke no equivale a protección global contra ataques distribuidos.

## Publicación de sesiones activas v1.14.0 — 2026-10-12

Usuario autoriza con «Adelante». ADR-0164 aceptado: sesiones opacas, lista y cierre de otras familias; sin cambio a JWT. Tag anotado v1.14.0, merge main `791d087e6a92ee2b1a8b09064826360a8837a793`. Gates: make verify local; PostgreSQL 18.6 real con -race separado (49,919 s); CI develop 38299037389 y main 38299074673 pasan. QA web oscuro escritorio y teléfono, cuenta sintética, cierre selectivo y todas las demás. No validación visual nativa ni distribución de binarios.

Backup previo nuevo dentro de la VM `20261011-080242F_20261012-193318I`, pgBackRest cifrado y status ok. Restauración exacta en Job aislado, repositorio solo lectura, emptyDir y sin montar datos activos: completa, base fuera de recovery. No exportación iCloud.

Incidente de preparación: espacio libre inicial ~1 GiB. Docker build goose agotó disco del host y su filesystem quedó en solo lectura. UTM informó ENOSPC y la VM dejó de responder (API 502), antes de ejecutar migración o rollout. Se retiraron solo temporales QA y caché Go recreable de esta tarea. Docker no recuperó su motor mediante CLI stop/restart; se terminaron sus procesos locales bloqueados, sin borrar volúmenes. Se prepararon archivos de imagen verificables sin Docker: API linux/arm64 compilada en host con Go 1.26.9, tag limpio, flags de Dockerfile; certificados, comandos de auditoría y goose de un artefacto existente cuya fuente y dependencias no cambiaron, más nueva capa con todas las migraciones y los roles productivos. Hashes y procedencia privados en artifact-provenance.json; lectura de retorno y referencias comprobadas antes de transferir. No se oculta como un docker build exitoso.

Recuperación: UTM se cerró y reabrió tras el estado bloqueado; se arrancó la VM existente, PostgreSQL recuperó y K3s volvió a servir. API antigua 2/2 y healthz 200 antes de retomar publicación. Huellas de siete tablas de negocio idénticas antes/después de recuperación y migración; no fue necesario restaurar datos productivos.

Migración 00023 aplicada con Goose antes de la API, versión 23. Secret efímero retirado. API nueva 2/2 con la imagen del SHA etiquetado; renderer productivo verificado con el mismo SHA. Cloudflare web `528216d2-50bb-44c1-8999-5ff406594888` y www `2aa52d9f-d0a7-46fd-afaf-6a847ad1ec94`, mismo deployment.json. Preview HTTPS y doce checks web productivos pasan, incluida /account/sessions noindex; ocho checks API existentes y cuatro de sesiones (401 sin credenciales, preflight móvil 204) pasan. No se crearon ni revocaron sesiones reales para el smoke.

Rollback: web/www de v1.13.2 y su API `28a9dfa51fed73437dcce09db9d73d486105d9ef`; conservar columnas aditivas 00023. Release: https://github.com/joseantoniogarciay/TournamentsManager/releases/tag/v1.14.0. `develop` avanzado a `main`; sin rama temporal para borrar. Evidencia privada en `/private/tmp/tm-release-v114-20261012`.

Cierre: web/API disponibles, Job de restore aislado retirado, Metro y simulador apagados. Docker local queda detenido tras terminar el backend bloqueado; no se borraron volúmenes. No pudo confirmarse el borrado del contenedor desechable `tm-sessions-release-qa` porque el motor dejó de responder. Sus datos son sintéticos, con restart desactivado; retirarlo al recuperar Docker. Local/dev y sus tareas permanecen apagados; no se inició observabilidad de desarrollo. Binarios móviles y QA visual nativo pendientes. Disco host ~1,2 GiB libre: no realizar otro build Docker pesado sin recuperar margen.

Retrospectiva: el ~1 GiB inicial era insuficiente para reconstruir goose y no se debía usar para un build que comparte disco con producción. Un gate de espacio debe preceder builds pesados en ese host. La restauración previa y las huellas facilitaron recuperar y verificar sin alterar datos; la procedencia explícita del empaquetado alternativo evita declarar un build Docker que falló.
