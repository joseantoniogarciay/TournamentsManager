# ADR-0059: Centralizar la autenticación de sesión en el borde HTTP

- **Estado:** Aceptado
- **Fecha:** 2026-07-28
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico

## Problema

Las rutas autenticadas deben extraer, validar y responder homogéneamente ante una
sesión. Repetirlo en cada handler multiplica la frontera de seguridad.

## Alternativas

### A — Autenticación en cada handler

- **Ventajas:** proximidad al caso de uso.
- **Inconvenientes:** duplicación y riesgo de omitir comprobaciones.
- **Coste de mantenimiento:** creciente.

### B — Middleware de sesión para rutas protegidas

- **Ventajas:** una única extracción de token, validación y respuesta `401`; el
  handler recibe el ID de cuenta desde el contexto.
- **Inconvenientes:** exige aplicar el middleware exclusivamente a las rutas que
  lo necesitan.
- **Coste de mantenimiento:** bajo.

### C — Middleware que también autoriza por rol

- **Ventajas:** parece reducir código por recurso.
- **Inconvenientes:** mezcla identidad de sesión y reglas de negocio por liga.
- **Coste de mantenimiento:** alto.

## Recomendación

**Opinión/recomendación:** B. Centraliza una preocupación transversal sin
ocultar las reglas de autorización.

## Decisión del usuario

**Aceptada el 2026-07-28:** adoptar B.

- `RequireSession` acepta exactamente una credencial: cookie
  `__Host-tm_session` para web o Bearer para apps; ambas se rechazan.
- Valida sesión opaca, vencimiento, revocación y cuenta verificada; deposita
  exclusivamente el ID interno de cuenta en el contexto.
- Sesión ausente, inválida, vencida o revocada devuelve el mismo `401` seguro.
  Un fallo de infraestructura devuelve `500` sin detalles internos.
- El middleware no contiene autorización de liga ni reglas de negocio.
- CSRF queda separado y será obligatorio antes de la primera mutación
  autenticada mediante cookie; no afecta al `GET` actual ni a Bearer móvil.

## Consecuencias

- Las rutas `/me/*` reutilizan una única frontera de sesión.
- Los handlers se concentran en autorización de recurso y respuesta HTTP.
- No se añade un middleware genérico de roles, JWT ni estado global.

## Validación

- Sin token, con token inválido o con cookie y Bearer a la vez se responde `401`.
- Un token válido deja disponible el ID de cuenta solo al handler protegido.
- Un error de PostgreSQL no se confunde con una sesión inválida.
- Una ruta pública no usa ni requiere este middleware.

## Disparadores de revisión

- Nuevos transportes de sesión o autenticación entre servicios.
- La primera mutación web con cookie, que abre la decisión concreta de CSRF.

## Documentación afectada

- [Arquitectura](../engineering/ARCHITECTURE.md)
- [Seguridad](../engineering/SECURITY.md)
- [API](../engineering/API.md)
- [Decisiones](../governance/DECISIONS.md)
- [Aprendizaje](../project/LEARNING.md)
