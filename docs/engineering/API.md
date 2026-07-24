# API

> Estado: REST y OpenAPI contract-first aceptados; operaciones y tooling
> pendientes.

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

- consumidores: web, mobile y acceso público;
- primer contrato para listar/ver, crear y unirse a torneos;
- versión y estructura de OpenAPI;
- herramientas de lint, generación y compatibilidad;
- generación o escritura manual de tipos de transporte Go;
- identidad, autenticación y autorización;
- validación y modelo de errores;
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

## Superficie candidata del primer vertical slice

Sin fijar todavía rutas ni payloads, la API necesitará capacidades para:

- listar torneos visibles para un invitado;
- consultar el detalle permitido;
- obtener la identidad y sesión actuales;
- crear un torneo como usuario autenticado;
- incorporarse a un torneo;
- consultar la relación del usuario con el torneo.

El diseño debe esperar a las decisiones de visibilidad, incorporación,
participantes e identidad descritas en [SYSTEM_OPTIONS.md](../governance/SYSTEM_OPTIONS.md).
