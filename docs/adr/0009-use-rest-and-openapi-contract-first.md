# ADR-0009: Usar REST con OpenAPI contract-first

- **Estado:** Aceptado
- **Fecha:** 2026-07-24
- **Decisor:** Usuario
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

El backend Go necesita exponer capacidades a un cliente universal ejecutado en
web, iOS y Android. Hay que decidir el estilo de comunicación, cuál será la
fuente de verdad del contrato y cómo evitar que cliente y servidor deriven.

## Contexto y restricciones

- El backend y la implementación de la API serán Go.
- La API HTTP es un adaptador de entrada del backend, no el backend completo.
- El cliente universal se desarrollará con React Native conforme a
  [ADR-0008](0008-use-a-universal-react-native-client.md).
- El cliente utilizará TypeScript para disponer de comprobación estática.
- Las aplicaciones instaladas pueden permanecer desactualizadas, por lo que el
  contrato debe evolucionar con compatibilidad deliberada.
- La lógica de negocio no puede depender del transporte, de OpenAPI ni del código
  generado.
- Las rutas y payloads concretos dependen todavía de decisiones funcionales.

## Criterios de decisión

1. contrato explícito y comprensible;
2. buena integración con web, iOS y Android;
3. detección temprana de incompatibilidades;
4. observabilidad y operación sencillas;
5. independencia entre transporte y dominio;
6. coste de adopción y mantenimiento proporcional a un equipo pequeño;
7. ecosistema adecuado para Go y TypeScript.

## Alternativas

### Alternativa A — REST con OpenAPI contract-first

Se diseña una descripción OpenAPI antes de implementar cada operación. El backend
Go implementa ese contrato y el cliente TypeScript se genera a partir de él.

- **Ventajas:** HTTP estándar y fácil de observar; documentación
  machine-readable; generación de cliente; mocks y pruebas de contrato; buena
  compatibilidad con navegador y aplicaciones.
- **Inconvenientes:** exige gobernar compatibilidad, errores, paginación y
  versionado; el contrato puede duplicar conceptos del código Go; la generación
  añade tooling.
- **Coste de adopción:** moderado; hay que diseñar, validar y generar el contrato.
- **Coste de mantenimiento:** bajo o moderado si OpenAPI permanece como fuente de
  verdad y CI detecta divergencias.
- **Riesgos:** diseñar endpoints que reflejen tablas, convertir DTOs en dominio o
  modificar código generado manualmente.

### Alternativa B — GraphQL schema-first

Un esquema GraphQL define tipos y operaciones; el backend implementa resolvers y
el cliente genera tipos y operaciones.

- **Ventajas:** consultas flexibles y contrato tipado; reduce over-fetching en
  vistas con datos heterogéneos.
- **Inconvenientes:** autorización por campo, caché, límites de complejidad,
  observabilidad y resolución N+1 requieren disciplina adicional.
- **Coste de adopción:** moderado o alto para el alcance inicial.
- **Coste de mantenimiento:** moderado; añade un modelo operativo específico.
- **Riesgos:** pagar complejidad antes de demostrar necesidades de consulta.

### Alternativa C — RPC con contrato

Un IDL define operaciones y mensajes, por ejemplo mediante gRPC o Connect.

- **Ventajas:** contratos fuertes, buena generación y comunicación eficiente.
- **Inconvenientes:** mayor fricción directa en navegador, semántica menos
  orientada a recursos públicos y tooling adicional.
- **Coste de adopción:** moderado.
- **Coste de mantenimiento:** moderado para varios targets.
- **Riesgos:** optimizar comunicación interna cuando el consumidor principal es
  un cliente público web/mobile.

### Alternativa D — REST sin descripción formal

Los handlers y documentación manual constituyen el contrato.

- **Ventajas:** mínimo tooling inicial.
- **Inconvenientes:** deriva entre documentación, backend y cliente; tipos
  duplicados; cambios incompatibles detectados tarde.
- **Coste de adopción:** bajo.
- **Coste de mantenimiento:** creciente con cada operación y consumidor.
- **Riesgos:** conocimiento implícito y pruebas insuficientes de compatibilidad.

## Comparación

REST con OpenAPI ofrece el mejor equilibrio para una API consumida por web, iOS y
Android: mantiene HTTP visible y operable, formaliza el límite y permite generar
el cliente sin introducir el modelo de ejecución de GraphQL o RPC.

GraphQL y RPC son alternativas válidas cuando existen necesidades demostradas,
pero añadirían complejidad sin evidencia actual. REST sin descripción formal es
inicialmente más corto, pero contradice la trazabilidad y calidad documental del
proyecto.

## Recomendación

**Opinión/recomendación:** alternativa A, REST pragmático con OpenAPI
contract-first y generación del cliente TypeScript.

La generación debe detenerse en el límite de transporte. No debe producir reglas
de negocio, casos de uso ni modelos del dominio.

## Decisión del usuario

Adoptar REST con OpenAPI contract-first.

- El backend y la API HTTP se implementarán en Go.
- La descripción OpenAPI será la fuente de verdad del contrato HTTP.
- El cliente que consume la API se generará en TypeScript.
- El cliente universal se escribirá en TypeScript para aprovechar tipado
  estático.
- No se generará lógica de negocio.

## Consecuencias

### Positivas

- Servidor y cliente se coordinan mediante un contrato versionable.
- La aplicación obtiene funciones y tipos TypeScript consistentes con OpenAPI.
- Se podrán generar documentación, mocks y datos para pruebas.
- CI podrá validar el contrato y detectar cambios incompatibles.
- El backend conserva Go como único lenguaje de ejecución del servidor.

### Negativas y deuda aceptada

- Habrá que seleccionar y mantener herramientas de lint, generación y
  compatibilidad.
- Un tipo TypeScript no valida datos en runtime; el servidor seguirá validando
  toda entrada y el cliente deberá tratar respuestas no confiables donde
  corresponda.
- El código generado puede producir APIs incómodas o tipos poco idiomáticos si el
  contrato está mal diseñado.
- La evolución deberá considerar clientes mobile que no se actualizan
  inmediatamente.

### Límites arquitectónicos

- OpenAPI describe DTOs de petición y respuesta, no entidades del dominio.
- Los handlers Go traducirán entre HTTP y casos de uso.
- Los casos de uso y el dominio no importarán paquetes HTTP ni tipos generados.
- El cliente generado no contendrá decisiones de autorización ni reglas de
  torneo.
- No se editará manualmente el código generado.
- La API verificará autenticación, autorización, validación e invariantes aunque
  el cliente tenga tipos estáticos.

## Validación

La futura implementación deberá demostrar que:

- una descripción OpenAPI válida existe antes de implementar una operación;
- el backend Go satisface el contrato mediante pruebas;
- el cliente TypeScript se reproduce desde una revisión concreta del contrato;
- regenerar el cliente no produce cambios sin registrar;
- CI detecta una descripción inválida y cambios incompatibles según la política
  que se acuerde;
- ningún paquete de dominio depende de DTOs OpenAPI o del transporte HTTP.

## Decisiones pendientes

- versión de OpenAPI soportada por todo el toolchain;
- estructura y ubicación del contrato;
- generador y configuración del cliente TypeScript;
- si se generan interfaces o DTOs de transporte para Go;
- framework o router HTTP de Go;
- formato de errores, paginación, filtros e idempotencia;
- política de compatibilidad y versionado;
- si el código generado se guarda en Git o se reproduce durante el build.

## Disparadores de revisión

- Las consultas reales requieren flexibilidad que REST no puede ofrecer sin
  proliferación significativa de endpoints.
- Aparecen comunicaciones internas de alto volumen con requisitos de RPC.
- La generación provoca más mantenimiento que seguridad de contrato.
- OpenAPI no puede expresar una necesidad real del protocolo sin extensiones
  frágiles.

## Documentación afectada

- [API.md](../../API.md)
- [ARCHITECTURE.md](../../ARCHITECTURE.md)
- [TECHNICAL_BASELINE.md](../../TECHNICAL_BASELINE.md)
- [SYSTEM_OPTIONS.md](../../SYSTEM_OPTIONS.md)
- [TESTING.md](../../TESTING.md)
- [DEVELOPMENT.md](../../DEVELOPMENT.md)
