# API

> Estado: principios iniciales; estilo y contrato pendientes del dominio.

## Objetivo

La API debe expresar casos de uso y contratos estables, no reflejar directamente
tablas ni detalles internos. Ningún endpoint se define antes de conocer actor,
objetivo, autorización, invariantes y respuesta de error.

## Decisiones previas

- consumidores: web, mobile y acceso público;
- primer contrato para listar/ver, crear y unirse a torneos;
- HTTP/REST, RPC u otra alternativa;
- formato y versionado del contrato;
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

## Definition of Done de un endpoint

- contrato y criterios de compatibilidad documentados;
- autorización y amenazas revisadas;
- validación e invariantes probadas;
- errores previsibles representados;
- métricas, logs y trazas definidos;
- límites y timeouts explícitos;
- documentación actualizada.

## Superficie candidata del primer vertical slice

Sin fijar todavía rutas ni payloads, la API necesitará capacidades para:

- listar torneos visibles para un invitado;
- consultar el detalle permitido;
- obtener la identidad y sesión actuales;
- crear un torneo como usuario autenticado;
- incorporarse a un torneo;
- consultar la relación del usuario con el torneo.

El diseño debe esperar a las decisiones de visibilidad, incorporación,
participantes e identidad descritas en [SYSTEM_OPTIONS.md](SYSTEM_OPTIONS.md).
