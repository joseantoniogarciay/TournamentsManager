# ADR-0113: Revocar sesiones propias ante eventos RISC de Google

- **Estado:** Aceptado
- **Fecha:** 2026-08-26
- **Decisor:** Usuario, mediante decisión explícita
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

FastTourney emite sesiones opacas propias después de validar Google. Si una
persona revoca el acceso de la aplicación o Google detecta que su cuenta fue
comprometida, esa sesión puede seguir activa hasta caducar o cerrarse
manualmente. Hay que sincronizar los eventos de seguridad de Google sin
convertir los tokens de Google en sesiones de producto.

## Contexto y restricciones

- ADR-0010, ADR-0050 y ADR-0062 separan identidad externa y sesión opaca.
- Google Cross-Account Protection entrega *Security Event Tokens* (SET) RISC
  firmados a un receptor HTTPS registrado para el mismo proyecto OAuth.
- Los eventos se usan solo para seguridad, antifraude y gestión de sesiones; no
  son analítica ni un canal de identidad de producto.
- La entrega puede repetirse. El endpoint no registra tokens, `sub`, emails,
  IDs de cuenta ni cuerpos de petición.
- No se almacenan refresh tokens de Google: el login actual solo usa ID tokens.

## Criterios

1. invalidar pronto las sesiones propias afectadas por una señal fiable;
2. verificar firma, emisor y audiencia antes de cambiar estado;
3. hacer idempotentes las entregas repetidas y no permitir que un evento antiguo
   cierre sesiones creadas después de la primera entrega;
4. mantener el límite hexagonal entre dominio, HTTP y PostgreSQL;
5. no introducir workers, colas ni tokens Google persistidos sin evidencia.

## Alternativas

### A — No integrar RISC

- **Ventajas:** no añade endpoint ni configuración en Google Cloud.
- **Inconvenientes:** una revocación o compromiso conocido por Google no alcanza
  las sesiones propias activas.
- **Coste de mantenimiento:** bajo.

### B — Receptor RISC validado y revocación idempotente

- **Ventajas:** sincroniza las señales de sesión/revocación relevantes y
  conserva la autenticación y autorización internas.
- **Inconvenientes:** requiere un endpoint público HTTPS y una cuenta de
  servicio solo para configurar el stream.
- **Coste de mantenimiento:** medio y acotado a una ruta, validación
  criptográfica, deduplicación y runbook.

### C — Consultar Google en cada petición autenticada

- **Ventajas:** no depende de entregas push.
- **Inconvenientes:** acopla cada solicitud a Google, añade latencia y no existe
  una consulta simple equivalente para este flujo basado en ID tokens.
- **Coste de mantenimiento:** alto.

## Comparación

A conserva la deuda que motivó la decisión. C transforma una señal asíncrona de
seguridad en una dependencia crítica de toda la API. B responde en el único
punto necesario y conserva la sesión opaca como autoridad de producto.

## Recomendación

**Recomendación:** B; es la solución mínima suficiente y el mecanismo publicado
por Google para proteger cuentas compartidas.

## Decisión del usuario

**Aceptada el 2026-08-26:** integrar Google Cross-Account Protection (RISC).

- La API expone un receptor HTTPS de infraestructura, fuera del contrato de
  producto, para los SET enviados por Google.
- Antes de producir efectos, valida la firma contra el JWKS indicado por la
  configuración RISC de Google, el emisor y una audiencia OAuth configurada.
- Para `sessions-revoked`, `tokens-revoked` y una cuenta deshabilitada por
  `hijacking`, revoca todas las sesiones opacas de la cuenta vinculada al
  `(issuer, sub)`. No elimina la identidad externa ni fusiona cuentas.
- La primera entrega de cada `jti` se recuerda durante un periodo acotado; las
  repeticiones no producen una transición ni cierran sesiones de un nuevo login.
- El endpoint devuelve `202` solo tras validar y aplicar o reconocer
  idempotentemente el evento; devuelve `400` ante SET inválido y `503` seguro
  ante un fallo técnico.
- La activación se completa registrando la URL pública en la RISC API con una
  cuenta de servicio mínima; su clave queda fuera de Git, imágenes y logs.

## Consecuencias

- Una revocación Google termina las sesiones FastTourney afectadas, sin borrar
  vínculos ni datos de torneos.
- Una cuenta con contraseña local puede volver a acceder con ese método; una
  cuenta solo Google necesita un login que Google acepte.
- Operamos una ruta adicional y debemos comprobar su disponibilidad y rechazos.
- RISC no garantiza eventos de Google Workspace ni reemplaza los controles
  propios de autenticación, reautenticación o fraude.

## Validación

1. Un SET válido revoca las sesiones de su cuenta y no las de otra.
2. Un SET repetido no revoca una sesión creada después del primero.
3. Firma, emisor, audiencia, estructura o tipo inválidos no cambian sesiones.
4. SET, `sub`, `jti`, audiencias y errores externos brutos no aparecen en logs,
   trazas o métricas.
5. La verificación de stream de Google recibe `202`.

## Disparadores de revisión

- Google cambia protocolo, JWKS, tipos de evento o condiciones de entrega.
- Se incorporan Apple u otros proveedores con señales equivalentes.
- El volumen justifica una cola o retención distinta de deduplicación.
- Un incidente exige bloquear también nuevos logins de una identidad afectada.

## Documentación afectada

- [Identidad](../engineering/IDENTITY.md)
- [Seguridad](../engineering/SECURITY.md)
- [Observabilidad](../operations/OBSERVABILITY.md)
- [Decisiones](../governance/DECISIONS.md)
- [Aprendizaje](../project/LEARNING.md)

## Fuentes técnicas

- [Google Cross-Account Protection (RISC)](https://developers.google.com/identity/protocols/risc?hl=es-419)
- [Google Sign in with Google Security Bundle](https://developers.google.com/identity/siwg/security-bundle)
