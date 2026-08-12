# ADR-0093: Usar Resend por SMTP para el correo del desarrollo público

- **Estado:** Aceptado
- **Fecha:** 2026-08-12
- **Decisor:** Usuario, mediante decisión explícita
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

`dev.fasttourney.com` necesita entregar correos de verificación y recuperación a personas externas. Mailpit captura correo únicamente en el Mac y ADR-0091 prohíbe exponerlo; no puede satisfacer ese recorrido.

## Contexto y restricciones

- ADR-0044 mantiene Mailpit como herramienta de diagnóstico local; no es un proveedor de producción.
- ADR-0091 exige SMTP transaccional y DNS de remitente antes de invitar personas que dependan de esos correos en `dev` público.
- El puerto `registration.Mailer` impide que los casos de uso dependan de SMTP o de un proveedor concreto.
- El plan gratuito de Resend dispone actualmente de 3.000 emails mensuales y 100 diarios; no admite sobreconsumo, por lo que el envío se detiene al alcanzar el límite. Esta cifra debe verificarse antes de cada cambio de plan.

## Criterios de decisión

1. entregar correo real sin coste variable inesperado en el volumen de `dev`;
2. preservar Mailpit y el ciclo local sin credenciales externas;
3. proteger la credencial en tránsito y limitar su alcance;
4. evitar una dependencia de SDK y una abstracción nueva innecesaria;
5. poder sustituir el proveedor sin modificar el dominio.

## Alternativas

### Alternativa A — Resend mediante SMTP autenticado con STARTTLS

- **Ventajas:** plan gratuito con corte sin cargos automáticos; DNS, entrega y panel gestionados; protocolo SMTP estándar; adapta el mailer existente con cambios acotados.
- **Inconvenientes:** alta de cuenta, dominio y DNS; límite diario de 100; la credencial API se usa como contraseña SMTP.
- **Coste de adopción:** bajo o medio.
- **Coste de mantenimiento:** bajo; vigilar cuota, DNS y rebotes en el panel.
- **Riesgos:** mala entregabilidad por DNS incompleto o agotamiento de cuota; se mitigan con dominio verificado, SPF/DKIM/DMARC y alertas operativas.

### Alternativa B — API HTTP y SDK de Resend

- **Ventajas:** API específica y respuestas estructuradas del proveedor.
- **Inconvenientes:** dependencia de API/SDK, otra forma de mensajes y más código sin necesidad demostrada.
- **Coste de adopción:** medio.
- **Coste de mantenimiento:** medio por actualización y acoplamiento.
- **Riesgos:** migración más costosa a otro proveedor.

### Alternativa C — Otro proveedor SMTP transaccional

- **Ventajas:** puede ofrecer otros precios, regiones o capacidades.
- **Inconvenientes:** no mejora el flujo actual y exige repetir evaluación, verificación DNS y operación.
- **Coste de adopción:** bajo o medio.
- **Coste de mantenimiento:** bajo o medio.
- **Riesgos:** elegir por una cuota temporal sin necesidad operativa.

### No cambiar

`dev` no puede invitar personas externas a flujos que requieran correo.

## Comparación

La alternativa A cubre la necesidad con el protocolo ya usado por el adaptador sin propagar una API propietaria hacia el dominio. B aporta más superficie sin un caso que la justifique. C permanece viable porque SMTP conserva la portabilidad, pero no presenta ventaja presente sobre A.

## Recomendación

**Opinión/recomendación:** alternativa A, usando `smtp.resend.com:587` con STARTTLS obligatorio y una API key de permiso *Sending access* restringida al dominio remitente.

## Decisión del usuario

**Aceptada el 2026-08-12:** Resend entrega el correo transaccional de `tournaments-manager-dev` mediante SMTP autenticado con STARTTLS. Mailpit permanece exclusivamente en `tournaments-manager-local`.

Antes de activar invitaciones externas se verificará el subdominio remitente `mail.fasttourney.com` en Resend y se publicarán los registros SPF, DKIM y DMARC que indique el proveedor. La clave de solo envío se guardará únicamente en `infra/dev/api.docker.env`, fuera de Git; no se registrará en logs ni se compartirá con local o producción.

## Consecuencias

### Positivas

- Las personas externas pueden recibir los enlaces de identidad de `dev`.
- La lógica de registro conserva el mismo puerto de salida y Mailpit sigue siendo inspectable para el desarrollo diario.
- El límite gratuito no genera un cobro automático.

### Negativas y deuda aceptada

- `dev` depende de Resend, de Internet y de DNS para enviar correo.
- Al agotarse 100 emails diarios o 3.000 mensuales, el envío falla hasta que se renueve la cuota o el usuario cambie de plan explícitamente.
- Se debe operar una clave y monitorizar la entregabilidad.

## Validación

1. La API de `dev` arranca solo con usuario y contraseña SMTP configurados en pareja; el transporte exige STARTTLS antes de autenticar.
2. El dominio remitente aparece verificado en Resend con SPF, DKIM y DMARC.
3. Una alta y una recuperación en `dev` entregan enlaces HTTPS en las cuatro localizaciones; Mailpit local continúa capturando los mismos flujos.
4. La clave no figura en Git, imágenes, logs ni salidas de pruebas.
5. Se comprueba el uso y se prueba el comportamiento seguro al rechazar el proveedor una entrega.

## Disparadores de revisión

- Se acercan de forma sostenida los límites de 100 diarios o 3.000 mensuales.
- Entregabilidad, rebotes o soporte muestran que el proveedor no es suficiente.
- Producción necesita otro aislamiento, región, SLA o volumen.
- Se requiere recepción de correo, campañas o automatizaciones.

## Documentación afectada

- [DECISIONS.md](../governance/DECISIONS.md)
- [DECISIONS_TO_REVISIT.md](../governance/DECISIONS_TO_REVISIT.md)
- [DEPLOYMENT.md](../operations/DEPLOYMENT.md)
- [infra/dev/README.md](../../infra/dev/README.md)
- [LEARNING.md](../project/LEARNING.md)
- [CHANGELOG.md](../../CHANGELOG.md)

## Fuentes técnicas

- [Resend: precios](https://resend.com/pricing?product=transactional)
- [Resend: cuotas y límites](https://resend.com/docs/knowledge-base/account-quotas-and-limits)
- [Resend: envío por SMTP](https://resend.com/docs/send-with-smtp)
