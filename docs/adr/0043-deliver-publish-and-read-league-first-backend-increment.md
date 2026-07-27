# ADR-0043: Entregar publicación y consulta de liga como primer incremento backend

- **Estado:** Aceptado
- **Fecha:** 2026-07-26
- **Decisor:** Usuario, mediante aceptación explícita de la propuesta de Fase 2
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

La Fase 2 necesita producir valor verificable a través de dominio, identidad,
persistencia y API sin mezclar todo el ciclo deportivo ni introducir
infraestructura adicional antes de contar con evidencia.

## Contexto y restricciones

- ADR-0031 a ADR-0042 definen el producto inicial y sus reglas futuras.
- ADR-0007, ADR-0009 y ADR-0011 fijan monolito modular Go, REST contract-first y
  PostgreSQL con `pgx`, `sqlc` y `goose`.
- Fase 1 dejó PostgreSQL local, migraciones explícitas y el runbook validados.
- Resultados, bajas y cancelación están definidos, pero no son necesarios para
  demostrar la primera publicación y consulta de una liga.
- No se añaden Redis/Valkey, MinIO, observabilidad adicional, Kubernetes ni
  contenedores de aplicación en este incremento.

## Criterios de decisión

1. recorrer producto, autorización, datos, API y operación con el menor alcance;
2. preservar las reglas ya aceptadas sin adelantar capacidades aplazadas;
3. permitir pruebas de integración reales con PostgreSQL local;
4. aislar decisiones de implementación que aún necesitan análisis propio;
5. mantener bajo el coste de desarrollo y mantenimiento.

## Alternativas

### Alternativa A — Implementar todo el ciclo de liga desde el inicio

- **Ventajas:** cubre resultados, bajas, cierre y cancelación en una entrega.
- **Inconvenientes:** combina demasiadas transiciones, permisos y auditoría;
  hace más difícil localizar errores y validar el núcleo inicial.
- **Coste de adopción:** alto.
- **Coste de mantenimiento:** alto inicialmente.
- **Riesgos:** construir reglas avanzadas sobre identidad, API y persistencia aún
  no demostradas.

### Alternativa B — Publicar y consultar una liga como primer incremento

El recorrido incluye borrador local, alta/verificación/login local, transferencia
del borrador a una cuenta verificada, publicación de una liga con equipos y
consulta por organizador o enlace no listado.

- **Ventajas:** demuestra la ruta de extremo a extremo y mantiene un tamaño
  diagnosticable; deja las transiciones deportivas para un incremento posterior.
- **Inconvenientes:** el usuario todavía no puede iniciar la liga ni gestionar
  resultados.
- **Coste de adopción:** medio.
- **Coste de mantenimiento:** bajo o medio.
- **Riesgos:** requerir decidir mecanismos concretos de sesión y verificación
  antes de implementar autenticación.

### Alternativa C — Implementar solo salud y cuentas

- **Ventajas:** entrega técnica pequeña.
- **Inconvenientes:** no prueba reglas de negocio, publicación, autorización
  sobre una liga ni el contrato de lectura compartida.
- **Coste de adopción:** bajo.
- **Coste de mantenimiento:** bajo.
- **Riesgos:** falsa sensación de avance sin un recorrido de producto.

### No cambiar

Mantener el backend sin implementar deja sin evidencia las decisiones ya
aceptadas y bloquea el aprendizaje de datos, API y pruebas de integración.

## Comparación

La alternativa A concentra más funcionalidad de la que se necesita para validar
la base. La C reduce el riesgo técnico pero no demuestra valor de producto. La B
atraviesa los límites necesarios con reglas contenidas y permite ampliar el ciclo
de liga sin reescribir el recorrido inicial.

## Recomendación

**Opinión/recomendación:** alternativa B, por ser la mínima suficiente para
validar el backend real sin convertir Fase 2 en todo el producto.

## Decisión del usuario

**Aceptada el 2026-07-26:** implementar como primer incremento de Fase 2 el
recorrido hasta publicar y consultar una liga. Los resultados, inicio, cierre,
bajas, cancelación y administración delegada se implementarán después. Cada
decisión de implementación que alcance el umbral de ADR se analizará y aceptará
antes de escribir su código.

## Consecuencias

### Positivas

- El primer incremento prueba identidad local, autorización, PostgreSQL,
  migraciones, OpenAPI y lectura no listada.
- El dominio deportivo avanzado queda aislado de la primera evidencia técnica.
- Se puede validar que una cuenta solo gestiona sus propias ligas.

### Negativas y deuda aceptada

- El producto aún no permite jugar ni cerrar una liga publicada.
- Apple, Google y recuperación de contraseña no forman parte de este incremento.
- La implementación requiere abrir decisiones concretas de sesiones,
  verificación local, modelo de datos y errores HTTP.

## Validación

- Desde una base vacía, las migraciones crean el esquema y se repiten sin cambios.
- Una cuenta verificada publica una liga válida con equipos.
- Otra cuenta no puede modificar esa liga.
- El enlace no listado permite solo lectura de la liga publicada.
- Las pruebas unitarias e integración relevantes y `make verify` pasan.
- Contrato OpenAPI, handlers, dominio y documentación permanecen alineados.

## Disparadores de revisión

- El recorrido no produce suficiente evidencia de producto o seguridad.
- La verificación de email o sesiones exige infraestructura desproporcionada.
- Resultados o administración delegada se vuelven necesarios para validar el
  producto antes de cerrar el incremento.

## Documentación afectada

- [PRODUCT.md](../project/PRODUCT.md)
- [ROADMAP.md](../project/ROADMAP.md)
- [ARCHITECTURE.md](../engineering/ARCHITECTURE.md)
- [API.md](../engineering/API.md)
- [DATABASE.md](../engineering/DATABASE.md)
- [TESTING.md](../engineering/TESTING.md)
- [DECISIONS.md](../governance/DECISIONS.md)
- [DECISIONS_TO_REVISIT.md](../governance/DECISIONS_TO_REVISIT.md)
- [LEARNING.md](../project/LEARNING.md)
- [CHANGELOG.md](../../CHANGELOG.md)
