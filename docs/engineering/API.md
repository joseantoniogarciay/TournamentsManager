# API

> Estado: REST y OpenAPI contract-first aceptados; contrato del primer
> incremento diseñado en ADR-0045, sin implementación todavía.

## Relación entre API y backend

El backend es el sistema ejecutado en el servidor: dominio, casos de uso,
autorización, persistencia, procesos y adaptadores. La API REST es uno de sus
adaptadores de entrada. Ambos se implementarán en Go.

OpenAPI describe el contrato HTTP entre ese backend y sus consumidores. No
implementa el servidor ni contiene reglas de negocio.

```text
Cliente universal TypeScript
          │
          │ cliente generado / HTTP
          ▼
API REST y handlers Go
          │
          ▼
Casos de uso y dominio Go
          │
          ▼
Persistencia y servicios externos
```

## Objetivo

La API debe expresar casos de uso y contratos estables, no reflejar directamente
tablas ni detalles internos. Ningún endpoint se define antes de conocer actor,
objetivo, autorización, invariantes y respuesta de error.

## Decisión vigente

Se adopta REST pragmático con OpenAPI contract-first conforme a
[ADR-0009](../adr/0009-use-rest-and-openapi-contract-first.md).

- La descripción OpenAPI es la fuente de verdad del contrato HTTP.
- El backend Go implementa y valida el comportamiento.
- Del contrato se genera un cliente TypeScript para la aplicación universal.
- Los DTOs OpenAPI se traducen en el borde y no se convierten en modelos de
  dominio.
- La generación no alcanza casos de uso ni lógica de negocio.

## Flujo contract-first

1. Definir la operación y sus DTOs en OpenAPI.
2. Revisar semántica HTTP, autorización, errores y compatibilidad.
3. Validar el documento.
4. Generar o actualizar el cliente TypeScript.
5. Implementar el adaptador HTTP en Go.
6. Probar que implementación y cliente satisfacen el contrato.
7. Publicar el cambio junto con documentación y aprendizaje.

Contract-first significa que el contrato precede a la implementación del
endpoint; no que OpenAPI preceda a las decisiones de producto.

## Navegación frente a operación REST

Una ruta del cliente no equivale necesariamente a una operación de la API. Por
ejemplo, el enlace de vinculación aceptado en
[ADR-0010](../adr/0010-own-identity-with-federated-login.md) abre
`/auth/link/confirm?token=...` mediante `GET`, pero esa navegación no cambia
estado. La confirmación explícita realizará una operación `POST` del backend,
cuyo contrato se diseñará en OpenAPI antes de implementarse.

Esta separación mantiene `GET` como método seguro, permite deep linking con
fallback web y evita que una previsualización del enlace vincule una cuenta.

## Decisiones pendientes

- generación o escritura manual de tipos de transporte Go;
- mecanismos de protección CSRF para el transporte cookie;
- operaciones posteriores: edición, inicio, resultados, seguimiento y administración;
- idempotencia y concurrencia;
- paginación, filtros y ordenación;
- compatibilidad y retirada de versiones;
- límites, timeouts y protección de abuso.

## Baseline de calidad

Todo contrato futuro debe:

- ser explícito y validable;
- diferenciar errores del cliente, del dominio y de infraestructura;
- evitar filtrar información sensible;
- definir idempotencia donde una repetición pueda causar daño;
- exponer información suficiente para correlación y diagnóstico;
- contar con ejemplos y pruebas de contrato.
- poder regenerar el cliente TypeScript de forma determinista.
- no producir cambios de estado mediante `GET`.

## Definition of Done de un endpoint

- contrato y criterios de compatibilidad documentados;
- autorización y amenazas revisadas;
- validación e invariantes probadas;
- errores previsibles representados;
- métricas, logs y trazas definidos;
- límites y timeouts explícitos;
- documentación actualizada.
- cliente TypeScript regenerado y sin modificaciones manuales.

## Contrato del primer incremento

La fuente de verdad de diseño es
[`contracts/openapi/v1/openapi.yaml`](../../contracts/openapi/v1/openapi.yaml).
Usa OpenAPI 3.1, prefijo `/v1` y `application/problem+json` conforme a RFC 9457.
Incluye alta, reenvío y confirmación de verificación, login, sesión actual y
logout, consulta del borrador verificado, colección autenticada de ligas
relacionadas, publicación y lectura pública por ID. `GET /me/leagues` pagina por
UUIDv7 y filtra en el servidor las relaciones `administered` y `followed`; la
segunda excluye una liga ya administrada para que la UI no la duplique. Véase
[ADR-0058](../adr/0058-list-account-related-leagues-with-a-paginated-collection.md).
El alta exige identidad local; el borrador es opcional y, si se envía, debe
cumplir íntegramente las restricciones de `DraftInput`.

La entrega de sesión se declara explícitamente: `cookie` para web (cookie
`__Host-`) y `bearer` para móvil. El secreto solo aparece una vez en la respuesta
de transporte `bearer`; no se almacena ni se devuelve en consultas posteriores.

La consulta pública `GET /usernames/{username}/availability` devuelve únicamente
`available`, no reserva el valor y usa `Cache-Control: no-store`; la creación de
cuenta mantiene la comprobación única definitiva. El adaptador limita esta
consulta a 30 solicitudes por IP y minuto por proceso y entrega `429` con
`Retry-After` al rebasarlo. El cliente la inicia solo para formatos válidos,
después de 400 ms sin escritura, y cancela la solicitud anterior cuando el valor
cambia.

La API aplica CORS en su frontera HTTP completa, no en cada endpoint. Solo
acepta los orígenes exactos de `CORS_ALLOWED_ORIGINS`, rechaza los demás y
resuelve los preflight `OPTIONS` para `DELETE`, `GET`, `POST` y `PUT` con las
cabeceras `Authorization` y `Content-Type`. Devuelve el origen concreto, `Vary:
Origin` y permite credenciales para que el transporte web por cookie pueda
funcionar en el futuro; por ello no se admite el comodín `*`.

Las operaciones protegidas pasan por middleware de sesión: acepta cookie o
Bearer, nunca ambas credenciales a la vez, y deja el ID de cuenta en el contexto
interno. La autorización por liga permanece en el caso de uso. Una protección
CSRF independiente será requisito antes de una operación mutante por cookie.
Véase [ADR-0059](../adr/0059-centralize-session-authentication-at-the-http-boundary.md).

## Validación y generación

[ADR-0046](../adr/0046-lint-and-generate-openapi-with-redocly-and-orval.md)
fija Redocly CLI para validar el contrato y Orval para generar el cliente
TypeScript basado en Fetch. La salida versionada está en
`apps/client/src/api/generated/` y no se edita a mano. `pnpm run openapi:lint`,
`pnpm run openapi:generate` y `pnpm run openapi:generate:check` son las entradas
respectivas; Make las integra en `check` y `verify`.

## Exploración local con Scalar

Para explorar y probar manualmente el contrato sin publicar otra aplicación ni
añadir rutas al backend, ejecuta:

```bash
pnpm run openapi:ui
# o
make openapi-ui
```

Scalar se sirve solo en `http://127.0.0.1:8082` y lee directamente
`contracts/openapi/v1/openapi.yaml`. Admite OpenAPI 3.1 y no es una fuente
adicional del contrato ni forma parte del artefacto desplegable de la API.
Para probar operaciones cuando exista el servidor Go local, el contrato apunta
a `http://127.0.0.1:8080/v1`; el puerto `8082` queda reservado para Scalar.

## Superficie candidata del primer vertical slice

Sin fijar todavía rutas ni payloads, la API necesitará capacidades para:

- resolver el acceso visible para invitado o invitación;
- consultar el detalle permitido;
- obtener la identidad y sesión actuales;
- crear un torneo como usuario autenticado;
- incorporarse a un torneo;
- consultar la relación del usuario con el torneo.

El diseño debe esperar a las decisiones de visibilidad, incorporación,
participantes e identidad descritas en [SYSTEM_OPTIONS.md](../governance/SYSTEM_OPTIONS.md).
