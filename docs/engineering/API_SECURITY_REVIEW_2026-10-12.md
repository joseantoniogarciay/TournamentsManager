# Revisión de seguridad de la API — 2026-10-12

## Resultado y alcance

**Estado actual:** correcciones implementadas en código según
[ADR-0163](../adr/0163-bound-api-resource-work-and-revalidate-invitations.md).
El [cierre de implementación](#cierre-de-implementación-autorizada) registra
pruebas y riesgos residuales. No se despliega ni se certifica producción.
La sección inicial conserva los hallazgos sobre el código previo.

**Resultado inicial: revisión abierta por hallazgos; sin certificación de seguridad.**
Se confirma un fallo concurrente de revocación de invitaciones y dos carencias
de contención de recursos. No se ha observado un bypass de sesión, una
inyección SQL o una filtración de errores internos en los recorridos probados.
Esto describe evidencia limitada, no demuestra ausencia de otras vulnerabilidades.

Código revisado: `develop`, `e724c0eb70803d3f6f931e3824f76c24f2ae3047`,
comparado con `v1.0.0`. Se leen manifiesto original y transcripción, handbook,
baseline cerrada, SECURITY.md, decisiones de identidad/autenticación/proxy y
ADR-0158 a ADR-0162. La superficie desde v1 incluye nuevos deportes y fases,
invitaciones, sugerencias, Apple, programación, repetición, equipos libres,
Mis partidos y Ediciones. La revisión prioriza fronteras HTTP, autorización,
persistencia concurrente, consumo de recursos y salidas seguras.

En la revisión inicial no se modificó implementación, contrato, ADR ni despliegue. No se consultan
datos de producción. Local/dev y observabilidad no se arrancan. Las pruebas
usan dobles, servidores de loopback del test y PostgreSQL 18.6 efímero, con
cuentas `example.test`, esquema inicial y las 22 migraciones. El rol de test
posee el esquema: estas pruebas no acreditan los privilegios del rol de runtime.

## Hallazgos confirmados

Las prioridades son una valoración de esta revisión, no una puntuación CVSS.

### F1 — P1: una inscripción pendiente puede usar una invitación ya revocada

Ubicación: `apps/backend/internal/adapters/postgres/tournament_invitation_repository.go`,
`JoinTeamInvitation`, consulta inicial con `FOR UPDATE OF tournaments`.

**Hecho reproducido:** una transacción bloquea el torneo; una inscripción con
el secreto vigente comienza y espera ese bloqueo. La primera transacción borra
la invitación y hace commit. La inscripción obtiene el bloqueo y consigue crear
equipo, vínculo de cuenta y seguimiento pese a la revocación ya confirmada.
No obtiene administración ni permiso para resultados.

La consulta une invitación y torneo bajo el snapshot de su sentencia. Bloquear
solo el torneo no obliga a releer la invitación después de la espera. La
comprobación secuencial de enlace revocado sí pasa y oculta este intercalado.
El mismo diseño requiere comprobar rotación concurrente; **esa variante no
se ha reproducido en esta sesión**. Es una frontera incorporada después de v1,
reutilizada por la reclamación de equipos de ADR-0160.

**Recomendación:** mantener la serialización por torneo y revalidar invitación,
hash y estado en una nueva sentencia después de adquirir el bloqueo, antes de
mutar. Conservar el rechazo seguro existente. Añadir regresiones deterministas
de revocación, rotación, inicio y reclamación mientras la petición espera.
Esta propuesta corrige la garantía aceptada; no requiere otra infraestructura.

Evidencia: [sonda PostgreSQL](evidence/api-security-2026-10-12/postgres_security_review_probe_test.go.txt)
y [ejecución con detector de carreras](evidence/api-security-2026-10-12/postgres-race-review.log).

### F2 — P1: tokens inválidos de recuperación desencadenan Argon2id sin contención

Ubicación: `internal/adapters/http/router.go`, ruta
`POST /v1/password-reset-confirmations`; `registration_sessions.go`,
`confirmPasswordReset`; `internal/registration/service.go`, `ResetPassword`.
Las rutas de esta sección son relativas a `apps/backend/`.

**Hecho reproducido:** doce peticiones de una misma IP con token inválido
alcanzan doce llamadas a `PasswordProtector.Hash`; todas reciben 409, ninguna 429. La sonda sustituye Argon2id por un contador: no ejecuta un ataque de carga.
En producción el protector usa Argon2id con 19 MiB y dos iteraciones por hash.
El repositorio comprueba el token después de ese trabajo. No hace falta cuenta,
sesión ni token válido para alcanzar la operación costosa. CORS no impide
llamadas de un cliente HTTP con Origin ausente.

**Impacto inferido:** peticiones concurrentes pueden consumir CPU y memoria de
la API. No se mide el umbral de agotamiento ni se afirma un incidente real.
La carencia ya existía en `v1.0.0`: no es una regresión introducida por Ediciones.

**Recomendación:** límite de intentos antes del hash y límite de concurrencia de
trabajo Argon2id. Evaluar comprobación barata previa del enlace manteniendo el
consumo final transaccional; una comprobación previa nunca sustituye ese consumo.
Revisar también reautenticación y cambio de contraseña. Los valores y respuestas
429 se deben acordar y reflejar en contrato, observabilidad y pruebas.

Evidencia: [sondas HTTP](evidence/api-security-2026-10-12/http_security_review_probe_test.go.txt).

### F3 — P2: cuerpos de identidad sin límite y ausencia de deadlines generales

Ubicación: `internal/adapters/http/registration_sessions.go`, `decodeBody`, y
handlers de identidad que lo llaman sin `MaxBytesReader`;
`cmd/api/main.go`, construcción de `http.Server`.

**Hecho reproducido:** un cuerpo JSON con un token de 2 MiB en confirmación de
recuperación se decodifica completo, alcanza el hash y acaba en 409. Solo se
valida que el token no esté vacío. Otras operaciones nuevas, como programación
y repetición, sí aplican límites de cuerpo propios; esa protección no cubre
automáticamente identidad.

El servidor configura `ReadHeaderTimeout=5s`, pero no `ReadTimeout`,
`WriteTimeout`, `IdleTimeout` ni un deadline general de petición. El pool no
añade un `statement_timeout` explícito. La configuración Caddy versionada no
añade límites de cuerpo ni de duración específicos para esta API. No se
inspecciona la configuración viva de Cloudflare, Traefik o PostgreSQL: puede
haber otras defensas operativas, pero esta revisión no puede acreditarlas.

**Impacto inferido:** asignaciones de memoria antes de validar y peticiones o
esperas técnicas prolongadas. El tamaño reproducido es moderado y local;
no se ejecutan ataques de conexiones lentas.

**Recomendación:** límite temprano de cuerpo, límites de longitud de secretos,
presupuesto temporal HTTP/DB y pruebas de excedente/cancelación. Mantener límites
por operación cuando sus tamaños legítimos difieran. No basta con un timeout
de escritura: el trabajo del handler necesita un contexto que expire.

## Riesgos de endurecimiento y límites de validación

### R1 — P2: paginación de Ediciones no acota el trabajo previo de la consulta

`db/queries/tournament_browsing.sql` recorre ancestros y todos los descendientes,
ordena la familia con `row_number()` y aplica el límite después. Mis partidos
también construye y ordena candidatos antes de seleccionar página. El router
no aplica limitador específico a estas lecturas ni a repetición. Ediciones es
pública por la política de enlace aceptada en ADR-0162.

**Hecho de código:** 50 elementos de respuesta no significan 50 filas procesadas.
**Riesgo inferido:** crecimiento de familias y consultas repetidas amplifican
carga. No se mide un DoS ni una familia de gran volumen en esta sesión. La
recursión usa `UNION`, no se presenta como un bucle infinito reproducido.

Recomendación: medir planes con volumen representativo, establecer presupuesto
temporal y de abuso; optimizar después con evidencia. Una entidad Serie, Redis
o caché distribuida tiene mayor coste y no queda justificada por esta revisión.
Descubrir otras ediciones desde una conocida es comportamiento aceptado, no
una fuga accidental ni una autorización para administrarlas.

### R2 — P2 de calidad: carrera sobre entrada compartida en la integración

La suite PostgreSQL con `-race` detecta escrituras concurrentes en
`tournaments.normalizedInput`, `creation.go:670`, desde
`TestIntegrationRepetitionIsIndependentIdempotentAndOwnerOnly`.
La normalización modifica el backing array de `Teams`; varias goroutines del
test reutilizan el mismo `RepeatInput`.

Esta carrera es real en el recorrido del test. **No se demuestra un exploit
HTTP:** cada handler decodifica una entrada propia. Recomiendo que la
normalización copie el slice antes de modificarlo, o que se explicite y pruebe
la propiedad exclusiva de la entrada; copiarlo tiene coste pequeño con 64 equipos.
La comprobación de concurrencia debe ejecutarse también con PostgreSQL real:
la suite general sin integración activa no ejecuta ese test.

## Controles y salidas revisados

| Frontera         | Evidencia y resultado                                                                                                                                                                          | Límite                                                         |
| ---------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------- |
| Sesión y CSRF    | Router aplica sesión y CSRF de cookie a programación/repetición/inscripción; rechaza cookie y Bearer simultáneos. Suites HTTP pasan.                                                           | No acredita configuración viva de cookies/proxy.               |
| Programación     | Bloqueo del torneo antes de comprobar propietario/delegado, partido dentro del mismo torneo y estado abierto; pruebas de lector, recurso cruzado, retirada de permiso, transferencia y cierre. | No se inyecta fallo en cada sentencia individual.              |
| Repetición       | Propiedad actual y completed bajo bloqueo; intención por origen/cuenta; nuevas relaciones independientes. Pruebas funcionales y carreras de reintento en DB.                                   | Carrera de entrada compartida detectada; sin limitador propio. |
| Reclamación      | Cuenta verificada por sesión, una plaza/un vínculo, invitación sin administración y capacidad máxima; pruebas de reclamación concurrente.                                                      | F1 invalida el cierre de revocación concurrente.               |
| Mis partidos     | Cuenta tomada del contexto; SQL filtra vínculos del equipo, no seguidores/delegados.                                                                                                           | Sin medición de carga representativa.                          |
| Ediciones        | Lectura por enlace aceptada, familia explícita; límites y UUID validados.                                                                                                                      | Descubribilidad aceptada; coste previo al LIMIT.               |
| Apple            | RS256, issuer/audience/exp, nonce/state, proof de cliente, callback de un uso, URL de retorno configurada, timeout de proveedor y JWKS acotado. Tests de proveedor/servicio/HTTP/DB.           | Sin login real en Apple durante esta revisión.                 |
| SQL y telemetría | SQL parametrizado; fallos simulados no exportan mensaje privado; causas cerradas y plantilla de ruta.                                                                                          | Se revisa código y tests, no eventos reales.                   |
| Tasa y proxy     | Limitadores existentes de login/alta/proveedores y pruebas de IP reenviada autenticada.                                                                                                        | Contadores en memoria por proceso; no cubren todas las rutas.  |

Los inventarios de [OBSERVABILITY.md](../operations/OBSERVABILITY.md) describen
éxito, validación, tasa, negocio y límites técnicos de los incrementos.
Se recorren 200/201, 400, 401/CSRF 403, permisos 403, ausencia 404, conflictos
409 y 500; para Apple también 202/303/429/503. Los fallos PostgreSQL de
conexión, consulta/restricción, deadline y cancelación tienen categorías seguras
probadas. Proveedor Apple cubre intercambio/JWKS y cancelación. Las rutas sin
limitador no generan un 429 propio. En repetición y programación, una
cancelación simulada puede escribir un 500 seguro; el cliente debe suprimir
feedback al abortar. Esta revisión no revalida visualmente el cliente ni
equipara simulación HTTP con inyección de fallos en cada frontera SQL.

## Verificaciones ejecutadas

- `go -C apps/backend test -race ./...`: pasa; PostgreSQL se omite sin activación.
  Primer intento restringido falló por prohibición de bind; repetición con
  servidores efímeros de test autorizados pasó.
- `govulncheck ./...`, también con `-show verbose`, usando Go 1.26.9:
  cero vulnerabilidades alcanzables y cero en paquetes importados. Un aviso
  de módulo: GO-2026-5932, `x/crypto/openpgp`, sin uso por la API y sin parche.
  El primer intento no pudo acceder a DNS; se repitió con acceso a la base oficial.
- Sondas HTTP: doce hashes con tokens inválidos y aceptación del cuerpo con
  token de 2 MiB. Dobles sin consumo Argon2id real ni red pública.
- PostgreSQL real con `-race -run Integration -count=1`: confirma F1 y detecta
  R2; **resultado global fallido** por esa carrera, no se declara verde.
- Suite funcional PostgreSQL sin `-race`, tras retirar sondas: 52 tests de
  integración pasan en 6,558 s. [Log funcional](evidence/api-security-2026-10-12/postgres-functional-review.log).

Las sondas se conservan como `.go.txt` para revisión y reproducción deliberada;
no son tests permanentes que fijen un comportamiento vulnerable como correcto.
Se pueden copiar temporalmente a su paquete original y ejecutar los nombres
`SecurityReview` exclusivamente contra una base efímera. La sonda PostgreSQL
reutiliza el helper de integración, que trunca cuentas; **nunca usar una base
de despliegue**. Los logs contienen únicamente fixtures y diagnósticos de test.

## Alternativas y recomendación para la siguiente fase

| Alternativa                                    | Ventajas                                                           | Coste y límites                                                                                              |
| ---------------------------------------------- | ------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------ |
| Mantener controles y aceptar hallazgos         | Ningún cambio técnico                                              | Mantiene F1 y exposición de recursos; requeriría aceptación explícita del riesgo.                            |
| Corregir F1 y contener recursos en el monolito | Reutiliza transacciones y límites existentes; regresiones pequeñas | Coste bajo para F1; moderado para acordar límites, contrato, telemetría y pruebas de identidad. Recomendada. |
| Añadir WAF/Redis/servicio externo de identidad | Puede aportar defensa operativa adicional                          | Mayor mantenimiento; no corrige por sí mismo el snapshot de F1 ni sustituye límites internos.                |

Recomiendo cerrar primero F1 con regresión, después acordar contención de hash,
cuerpos y deadlines, y finalmente medir las nuevas consultas. No se implementa
una nueva política importante durante esta revisión: la decisión final es del
usuario y cualquier ADR necesario debe quedar aceptado antes de implementarla.

## Retrospectiva técnica

Un bloqueo de fila no equivale a revalidar todas las entidades consultadas antes
de obtenerlo. Una suite funcional en verde y un escáner sin CVE alcanzables no
detectan por sí solos abuso de recursos o intercalados de revocación. El mayor
valor de esta revisión fue combinar casos negativos, un intercalado controlado
y PostgreSQL real con detector de carreras, manteniendo separadas evidencia,
riesgo inferido y alcance de producción no comprobado.

Fuentes primarias: [OWASP API4: consumo de recursos](https://api-security.owasp.org/editions/2023/en/0xa4-unrestricted-resource-consumption/),
[timeouts de http.Server](https://pkg.go.dev/net/http#Server),
[GO-2026-5932](https://pkg.go.dev/vuln/GO-2026-5932).

## Cierre de implementación autorizada

El usuario autoriza «Sí vamos a corregir todo eso». ADR-0163 queda aceptado antes
de implementar la política técnica. Se corrige en el checkout de develop, sin
publicación ni cambios de durabilidad, infraestructura o dependencias.

| Hallazgo | Corrección y comprobación | Riesgo residual |
| --- | --- | --- |
| F1, revocación concurrente | Relectura de hash vigente después de obtener el bloqueo; tests deterministas de revocación, rotación e inicio sin inscripción parcial. | La invitación sigue siendo una capacidad reenviable según política aceptada. |
| F2, hash con enlace inválido | Inspección previa sin Argon2id, consumo final atómico, 10 intentos/minuto/IP y cupo compartido de dos peticiones de contraseña; tests de cero hashes y saturación/liberación. | Un hash iniciado no se interrumpe; conserva su cupo hasta finalizar. |
| F3, cuerpo y tiempos | 64 KiB global antes de decode, límites menores conservados, secretos acotados; HTTP 15 s de contexto, lectura 10 s, escritura 20 s, idle 60 s; SQL 5 s/lock 2 s. | Runtime de producción no inspeccionado ni desplegado. |
| R1, coste de colecciones | 60/minuto/cuenta personal, 60/minuto/IP Ediciones, 5/minuto/cuenta repetición y presupuesto SQL; prueba local de 5001 ediciones y liga de 64 equipos / 2016 partidos. | LIMIT no hace constante el trabajo previo. La muestra no es un benchmark de producción ni cubre todos los tamaños/distribuciones. |
| R2, entrada compartida | Normalización copia Teams antes de modificarlo; suite PostgreSQL con detector de carreras pasa, incluido el reintento concurrente original. | Cada llamada conserva entrada propia; no se introduce estado global de negocio. |

Los limitadores admiten como máximo 4096 claves y evitan escanear toda la tabla
en cada intento. La cuota de credenciales se comparte por cuenta entre sesiones
y operaciones; cambiar el token no la reinicia. La saturación y la cuota tienen
429/Retry-After y causa segura rate_limit.exceeded, sin detalles o claves
exportadas. Nuevos cuerpos 400 y rechazos 429 se declaran en OpenAPI y se
regenera el cliente; el metadata x-resource-policy documenta los presupuestos.

Se preservan los fallos técnicos de inspección/consumo de recuperación y
consulta/emisión/consumo de tickets: no se convierten en enlaces o credenciales
inválidos. SQLSTATE 57014 y 55P03 usan request.timeout en raíz y límite técnico.
El refresh nativo solo guarda una respuesta 200; 400/429/500 preservan sesión
y producen error común seguro. Tests comprueban ausencia de escrituras y logout
con esos estados; el tratamiento de 401 no cambia.

### Evidencia de cierre

- [PostgreSQL con race detector](evidence/api-security-2026-10-12/fixed-postgres-race.log):
  55 tests de integración pasan, además de subcasos. Esquema y 22 migraciones
  en contenedor efímero PostgreSQL 18.6; no se cambia una base de despliegue.
- SQL real interrumpe pg_sleep(30) por statement_timeout y una espera por
  lock_timeout. Muestra de volumen: familia 5001 y liga de 2016 encuentros,
  páginas de 51 filas incluyendo fila de paginación adicional; tiempos concretos
  en el log, medidos una vez sobre Docker local.
- Suite Go completa con -race, lint Go y build; tests de HTTP, servicios y
  categorización técnica. La suite general no se presenta como prueba de
  integración: su PostgreSQL se omite y se acredita por el log separado.
- Contrato válido, generación OpenAPI repetida con hashes iguales y sqlc sin
  deriva. Typecheck, lint cliente, 70 tests de features/refresh y exportación web.
- Govulncheck mantiene cero vulnerabilidades alcanzables. El aviso de módulo
  openpgp no utilizado conserva el tratamiento y alcance del informe inicial.
- Once pruebas de seguridad operacional pasan. No se arranca observabilidad
  ni local/dev; se confirma eliminación del contenedor propio.

[Resumen reproducible de checks](evidence/api-security-2026-10-12/fixed-checks.md).
No se realiza QA visual nativa ni prueba en Apple real: el cambio de cliente
es de transporte y almacenamiento, probado con dobles y exportación web.
No se presenta la exportación como evidencia de simuladores o dispositivos.

### Retrospectiva del cierre

La corrección efectiva conserva el bloqueo pero cambia el momento de validar
la capacidad. Contener la operación costosa requiere cuota, concurrencia, cuerpo
y tiempo: un control aislado deja otros caminos abiertos. La política permanece
en HTTP/configuración de runtime y las reglas de negocio siguen independientes.
Se conserva la familia completa y se mide antes de añadir estructuras nuevas.
Los generadores que limpian carpetas se ejecutan antes de typecheck, en secuencia;
las pruebas de concurrencia se ejecutan con PostgreSQL activo para no ocultarlas
tras skips. Las correcciones están verificadas en código; desplegarlas y revisar
el runtime real constituyen una operación posterior que aquí no se ejecuta.
