# ADR-0127: Crear el torneo atómicamente con la sesión de acceso

- **Estado:** Aceptado
- **Fecha:** 2026-09-15
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** Ninguno

## Problema

Una persona puede preparar un torneo completo sin sesión y decidir después
iniciar sesión con contraseña o Google. Si el acceso crea primero la sesión y el
cliente crea después el torneo mediante otra petición, el segundo paso puede
fallar y dejar el resultado a medias.

## Contexto y restricciones

- ADR-0079 ya crea atómicamente cuenta y torneo cuando el borrador acompaña al
  alta.
- El borrador permanece local hasta cruzar una operación autenticable.
- Una cuenta pendiente que no obtiene sesión no puede crear un torneo mediante
  el login.
- La sesión propia, y no Google, sigue siendo la frontera de autorización.
- OpenAPI continúa siendo la fuente de verdad del contrato HTTP.

## Criterios

1. No confirmar un login si el torneo solicitado no se ha persistido.
2. Aplicar la misma garantía a contraseña y Google.
3. Borrar el borrador local solo tras una respuesta de éxito.
4. Mantener la lógica transaccional en PostgreSQL y fuera del cliente.
5. No introducir colas, sagas ni estados intermedios para una única base de datos.
6. Repetir la misma intención después de perder una respuesta no duplica el torneo.

## Alternativas

### A — Crear el torneo después del login desde el cliente

- **Ventajas:** reutiliza sin cambios los endpoints existentes.
- **Inconvenientes:** sesión y torneo pueden divergir si falla la segunda petición.
- **Coste de mantenimiento:** bajo, con una recuperación visible más compleja.

### B — Aceptar el borrador opcional en ambos endpoints de login

- **Ventajas:** sesión, torneo y equipos se confirman o revierten juntos en la
  misma transacción PostgreSQL; replica la garantía del alta.
- **Inconvenientes:** amplía los DTO y las transacciones de autenticación.
- **Coste de mantenimiento:** bajo; reutiliza validación y persistencia existentes.

### C — Coordinar login y creación mediante saga o cola

- **Ventajas:** permitiría separar almacenes o servicios en el futuro.
- **Inconvenientes:** añade estados compensatorios y consistencia eventual sin
  necesidad actual.
- **Coste de mantenimiento:** alto.

## Comparación y recomendación

La alternativa A no cumple la garantía solicitada y C es sobreingeniería para un
monolito con una sola base de datos. **Recomendación:** alternativa B.

## Decisión del usuario

**Aceptada el 2026-09-15:** los logins con contraseña y Google aceptan un
`draft` completo opcional. Cuando el login produce una sesión, el backend crea
sesión, torneo `published` y equipos en una única transacción. Un fallo revierte
todos esos efectos. El cliente conserva el borrador hasta recibir éxito y lo
elimina después.

**Ampliación aceptada el 2026-09-15:** cada borrador transferible conserva un
`draftId` UUID estable durante toda su vida local. PostgreSQL lo persiste como
identidad de origen del torneo y exige unicidad por cuenta organizadora. Un
reintento del mismo login puede crear una nueva sesión, pero reconoce el torneo
ya confirmado y no vuelve a insertar ni el torneo ni sus equipos. Un borrador
nuevo recibe siempre otro identificador.

Un login local de una cuenta pendiente conserva su respuesta `202`, renueva la
verificación y no transfiere el borrador porque no crea sesión.

## Consecuencias

- El contrato de sesión local y Google reutiliza `TournamentInput`.
- La persistencia de identidad coordina la transacción, pero las reglas del
  torneo continúan validadas antes de entrar en ella.
- No se añade un servicio distribuido ni una segunda operación compensatoria.
- Una respuesta perdida después del commit ya no duplica el torneo: el reintento
  usa el mismo `draftId` y la restricción única decide de forma concurrente.
- La sesión sigue siendo una credencial efímera: un reintento puede dejar una
  sesión previa no recibida, que caduca y puede revocarse como las demás.

## Validación

- Login local y Google con borrador válido crean exactamente una sesión, un
  torneo y sus equipos.
- Si falla cualquier inserción del torneo, no queda una sesión persistida.
- Login sin borrador conserva el comportamiento actual.
- Cuenta local pendiente no transfiere el borrador mediante login.
- El cliente no elimina el borrador ante `400`, `401`, `429`, `5xx` o fallo de red.
- Dos logins secuenciales o concurrentes con la misma cuenta y `draftId` dejan un
  único torneo y un único conjunto de equipos.
- Un borrador local antiguo sin identificador recibe uno y lo persiste antes de
  cruzar la frontera HTTP.

## Disparadores de revisión

- Se necesita devolver el torneo creado directamente en la respuesta de login.
- Sesiones y torneos dejan de compartir una transacción PostgreSQL.
- Se admite más de un borrador local simultáneo.

## Documentación afectada

- [Producto](../project/PRODUCT.md)
- [Identidad](../engineering/IDENTITY.md)
- [API](../engineering/API.md)
- [Observabilidad](../operations/OBSERVABILITY.md)
- [Decisiones](../governance/DECISIONS.md)
- [Aprendizaje](../project/LEARNING.md)
