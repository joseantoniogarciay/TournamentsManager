# ADR-0163: Contener recursos de la API y revalidar invitaciones tras el bloqueo

- **Estado:** Aceptado
- **Fecha:** 2026-10-12
- **Decisor:** Usuario, «Sí vamos a corregir todo eso», tras revisar los hallazgos

## Problema y evidencia

La revisión API_SECURITY_REVIEW_2026-10-12 reproduce inscripción después de
revocación concurrente, hashes con tokens inválidos, cuerpos sin límite y una
carrera de normalización. Las nuevas lecturas paginan la respuesta pero no el
trabajo previo. El gate técnico está cerrado.

## Alternativas, coste y recomendación

Aceptar los riesgos no cumple la corrección solicitada. Un WAF, Redis o una
entidad Serie añaden operación y no corrigen el snapshot de autorización.
Se adopta la solución mínima en el monolito: relectura tras bloqueo, límites
tempranos y transacciones existentes. Coste bajo en persistencia y moderado en
contrato, pruebas de límites y presupuestos; sin nuevas dependencias.

## Decisión y concreción técnica

- Invitación: resolver el torneo y bloquearlo; en una segunda sentencia verificar
  que el mismo hash sigue vigente. Rechazo seguro existente; inscripción/claim
  final sigue atómica con cuenta, estado, plaza y seguimiento.
- Recuperación: inspección barata antes de hash, manteniendo consumo final
  transaccional; errores técnicos no se convierten en enlace inválido.
- Dos operaciones HTTP costosas de contraseña simultáneas por proceso,
  compartidas por registro, login, recuperación, reautenticación y cambio local.
  Saturación rechazada inmediatamente con 429 y Retry-After; no cola sin límite.
- Confirmación de recuperación: 10/minuto por IP; cambios de credenciales y
  reautenticación: 10/minuto por cuenta, compartidos; repetición: 5/minuto por
  cuenta; Mis partidos: 60/minuto por cuenta; Ediciones: 60/minuto por IP.
- Contadores por proceso con máximo de 4096 claves por limitador. Capacidad
  agotada rechaza nuevas claves, sin evicción que permita saltarse el límite.
- Cuerpo global máximo de 64 KiB antes de decodificar; límites menores de
  callback, programación y repetición continúan vigentes. Exceso usa 400 seguro
  de validación, incluido chunked. Secretos de recuperación/verificación/ticket
  se acotan a 512 bytes, sin romper fixtures existentes; tokens OIDC mantienen
  su tamaño contractual y quedan sujetos al cuerpo máximo.
- Contexto de petición: 15 s; HTTP lectura: 10 s, escritura: 20 s, idle: 60 s,
  cabeceras: 5 s. PostgreSQL: statement_timeout 5 s y lock_timeout 2 s por
  conexión de runtime. Cancelar el contexto interrumpe esperas; un hash que ya
  se ejecuta termina bajo el máximo de concurrencia, no se simula preemption.
- Normalización copia Teams antes de modificarlo. Medir consultas de familia
  y partidos con volumen representativo; límites y deadlines contienen el
  trabajo sin truncar silenciosamente la familia ni cambiar la visibilidad.
- Nuevos 429 usan fallback seguro existente del cliente, no mensajes de negocio
  por estado ni detalles internos. Causa cerrada rate_limit.exceeded. Timeout
  conserva request.timeout; no se exportan inputs ni errores brutos.

Estos valores iniciales son concreción de la corrección autorizada, no un SLO ni
capacidad medida. Se documenta que limitadores y concurrencia son por proceso;
replicar la API requiere revisar el presupuesto agregado.

## Validación y cierre

Regresiones de revocación/rotación mientras se espera, doble claim, token inválido
sin hash, fallo de inspección/consumo, saturación y liberación del cupo, cuentas
independientes, exceso de cuerpo, deadline, SQL real, safe errors y races.
No se despliega ni se arranca local/dev u observabilidad. Retrospectiva y
evidencias se registran con el informe al cerrar la implementación.
