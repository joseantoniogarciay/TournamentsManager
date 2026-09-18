# ADR-0125: Persistir sugerencias privadas y avisar por correo

- **Estado:** Aceptado
- **Fecha:** 2026-09-13
- **Decisor:** Usuario, mediante decisión explícita
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

Las personas autenticadas necesitan comunicar qué echan en falta desde Inicio.
El responsable necesita recibir esas sugerencias ahora sin construir todavía un
CRM, pero el correo no puede ser la única copia si se quiere aprender de ellas y
evolucionarlas posteriormente.

## Contexto y restricciones

- PostgreSQL es el sistema de registro principal y Resend/Mailpit ya implementan
  el límite SMTP sin acoplar el dominio al proveedor.
- La primera versión es privada; no ofrece listado, publicación ni respuestas.
- El texto es contenido aportado por una persona y no entra en logs, trazas ni
  analítica.
- El correo es una dependencia falible y no debe revertir un guardado correcto.
- La política de privacidad debe explicar la nueva categoría y su entrega.

## Criterios de decisión

1. No perder una sugerencia por un fallo de correo.
2. Dar confirmación inequívoca y permitir reintentar sin perder el texto.
3. Limitar abuso con una regla comprensible.
4. Dejar un recurso mínimo ampliable sin modelar por adelantado un CRM social.
5. Mantener la lógica independiente de PostgreSQL y SMTP.

## Alternativas

### A — Enviar únicamente un correo

- **Ventajas:** implementación y consulta inmediatas.
- **Inconvenientes:** un fallo pierde el contenido; dificulta búsqueda, estados y
  evolución posterior; duplica la retención en un buzón personal.
- **Coste de mantenimiento:** bajo al principio, alto al migrar.

### B — Persistir y enviar un aviso secundario por correo

- **Ventajas:** PostgreSQL conserva el resultado; el correo avisa sin controlar
  el éxito; el mismo registro puede alimentar una futura herramienta interna.
- **Inconvenientes:** el aviso puede fallar aunque la sugerencia exista y deja una
  copia operativa en el buzón receptor.
- **Coste de adopción y mantenimiento:** bajo o medio.

### C — Construir ya un CRM y un tablón público

- **Ventajas:** gestión, respuestas y visibilidad desde el primer día.
- **Inconvenientes:** exige permisos, moderación, consentimiento de publicación,
  retirada, notificaciones y estados antes de validar el volumen real.
- **Coste de adopción y mantenimiento:** alto.

## Comparación

A satisface el aviso pero no la durabilidad. C anticipa múltiples capacidades sin
evidencia. B crea el mínimo registro fiable y mantiene el canal externo como un
adaptador sustituible.

## Recomendación

**Opinión/recomendación:** alternativa B.

## Decisión del usuario

**Aceptada el 2026-09-13:** Inicio muestra a las cuentas autenticadas una tarjeta
«¿Te falta algo?» debajo de Actividad reciente. Contiene un campo multilínea con
placeholder explicativo y una acción que se habilita cuando el texto, sin espacios
exteriores, tiene entre 8 y 1.000 caracteres.

`POST /v1/me/suggestions` persiste el texto, la cuenta y la fecha en PostgreSQL.
Cada cuenta puede enviar tres sugerencias por hora en la instancia API. Después
del commit se intenta enviar un aviso por el SMTP existente al destinatario
configurado por entorno, inicialmente `joseantoniogarciay@gmail.com`. Un fallo de
SMTP no revierte el registro ni cambia el `201`.

Al confirmar el guardado, el cliente vacía el campo y muestra un banner localizado
de agradecimiento. Ante cualquier fallo de guardado conserva el contenido; el
`429` recibe feedback específico y el resto usa los fallbacks comunes seguros.

## Consecuencias

- `product_suggestions` es privado y desaparece al eliminar su cuenta por la
  relación `ON DELETE CASCADE`.
- El aviso incluye username, identificador, fecha y texto; no incluye email ni ID
  de cuenta. Su copia operativa se elimina cuando deja de ser necesaria para
  revisar la sugerencia.
- El destinatario es configuración, no lógica de dominio ni un secreto embebido.
- El límite en memoria es suficiente para una instancia; varios réplicas exigirán
  un limitador compartido o aceptar una tolerancia proporcional.
- No se crean estados, panel administrador, respuestas ni visibilidad pública.

## Validación

- Siete caracteres no habilitan el envío; ocho sí; más de mil se recortan en el
  cliente y se rechazan en el servidor.
- Un `201` vacía el campo y muestra agradecimiento; un fallo conserva el texto.
- El cuarto envío de una cuenta durante la ventana devuelve `429` con
  `Retry-After`.
- Un fallo SMTP posterior al insert conserva la sugerencia y responde `201`.
- Logs, atributos de span y analítica no contienen texto, username, cuenta ni
  destinatario.

## Disparadores de revisión

- Se necesitan estados, búsqueda o respuestas internas.
- Se desea publicar una sugerencia o atribuirla mediante username.
- Hay varias réplicas de API o el abuso supera el limitador local.
- El buzón personal deja de ser adecuado por volumen, acceso o retención.

## Documentación afectada

- [Producto](../project/PRODUCT.md)
- [API](../engineering/API.md)
- [Base de datos](../engineering/DATABASE.md)
- [Seguridad](../engineering/SECURITY.md)
- [Observabilidad](../operations/OBSERVABILITY.md)
- [Despliegue](../operations/DEPLOYMENT.md)
- [Decisiones](../governance/DECISIONS.md)
- [Aprendizaje](../project/LEARNING.md)
- [Changelog](../../CHANGELOG.md)
