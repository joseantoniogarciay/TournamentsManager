# ADR-0045: Diseñar el modelo inicial de datos y el contrato OpenAPI

- **Estado:** Superado parcialmente
- **Fecha:** 2026-07-26
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** ADR-0048, exclusivamente en la composición del alta y verificación

## Problema

El incremento aceptado necesita un modelo persistente y un contrato HTTP antes de
crear migraciones o handlers. Sin ellos, registro, verificación, sesión y
publicación podrían discrepar en sus estados, autorizaciones o reglas de
seguridad.

## Contexto y restricciones

- ADR-0031 conserva un borrador solo después del alta, asociado a una cuenta
  pendiente; ADR-0043 delimita el primer incremento.
- ADR-0033 exige lectura no listada sin sesión; ADR-0034 exige `username` al
  activar la cuenta.
- ADR-0044 fija Argon2id, sesiones opacas, verificación de 24 horas, inactividad
  de 7 días y duración absoluta de 30 días.
- PostgreSQL es el registro de verdad y OpenAPI contract-first es obligatorio.
- Este ADR diseña; no autoriza migraciones, adaptadores HTTP ni generación de
  cliente.

## Alternativas

### A — Identificadores UUIDv7, secretos opacos y recursos REST versionados

Los recursos persistentes usan UUIDv7. Las credenciales y enlaces compartibles
son secretos aleatorios independientes, almacenados solo como huella. El
contrato es OpenAPI 3.1 bajo `/v1`, con errores RFC 9457.

- **Ventajas:** orden temporal aproximado sin exponer secuencias; secretos
  revocables y no reutilizables; contrato estable y explícito.
- **Inconvenientes:** UUIDv7 requiere generación consciente y las huellas no se
  pueden usar para recuperar el secreto.
- **Coste de mantenimiento:** bajo; no introduce un servicio ni abstracción.

### B — Enteros secuenciales y tokens almacenados en claro

- **Ventajas:** SQL y depuración inicial más simples.
- **Inconvenientes:** facilita enumeración si se expone por error y una fuga de
  base de datos permite reutilizar sesiones, verificaciones o enlaces.
- **Coste de mantenimiento:** alto por el riesgo de seguridad y migraciones de
  corrección posteriores.

### C — JWT y un único endpoint RPC de flujo

- **Ventajas:** menor número aparente de recursos HTTP.
- **Inconvenientes:** contradice ADR-0044, mezcla casos de uso y dificulta la
  revocación inmediata y la evolución del contrato.
- **Coste de mantenimiento:** medio o alto.

## Decisión del usuario

**Aceptada el 2026-07-26:** alternativa A, con estas concreciones:

- UUIDv7 para identificadores internos y expuestos de recursos; no se usan como
  secreto de compartición ni de autenticación.
- Tokens de sesión, verificación y enlace no listado son aleatorios, opacos, de
  al menos 128 bits y se persisten exclusivamente como una huella SHA-256 con
  un contexto por tipo de token.
- Una cuenta pendiente y su borrador caducan siete días después de su creación;
  una purga explícita los elimina. La verificación vence a las 24 horas y puede
  reenviarse sin revelar si el correo existe.
- OpenAPI 3.1 se publica en `contracts/openapi/v1/openapi.yaml`; `/v1` es el
  prefijo de compatibilidad, RFC 9457 el formato de error y las mutaciones usan
  `POST`, `PUT` o `DELETE`, nunca `GET`.
- La verificación recibe el `username` y activa la cuenta en la misma
  transacción que consume el token y crea la sesión.

## Consecuencias

- La publicación transforma atómicamente el borrador de una cuenta verificada
  en liga, equipos, partidos y enlace no listado.
- La API no expone hashes, contraseñas, tokens, email normalizado ni IDs de
  sesión.
- La limitación de tasa, entrega SMTP, transacciones, índices concretos y
  migraciones se validarán en implementación sin alterar este contrato.

## Validación

- El documento OpenAPI es válido y describe todos los flujos del incremento.
- El modelo lógico impide publicar desde una cuenta pendiente y consultar un
  borrador por enlace.
- Una segunda verificación o el uso de una sesión revocada falla sin crear otra
  sesión ni modificar datos.
- Las pruebas futuras cubren las restricciones y la atomicidad de publicación.

## Disparadores de revisión

- Requisito de claves públicas ordenables por otro sistema, particionado o
  compatibilidad que haga inadecuado UUIDv7.
- Retención legal, abuso o soporte que requiera otro plazo de purga.
- Un consumidor externo que necesite una versión de API distinta.

## Documentación afectada

- [Modelo inicial](../engineering/INITIAL_DATA_MODEL.md)
- [API](../engineering/API.md)
- [Datos y persistencia](../engineering/DATABASE.md)
- [Identidad y acceso](../engineering/IDENTITY.md)
- [Aprendizaje](../project/LEARNING.md)
- [OpenAPI v1](../../contracts/openapi/v1/openapi.yaml)
