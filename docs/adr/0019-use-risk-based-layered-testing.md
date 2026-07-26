# ADR-0019: Usar pruebas por riesgo y capas

- **Estado:** Aceptado
- **Fecha:** 2026-07-25
- **Decisor:** Usuario, mediante aceptación explícita de la alternativa B
- **Propietario del análisis:** Asistente como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

El proyecto necesita evidencia automatizada de que conserva su comportamiento
sin convertir las pruebas en una segunda plataforma. Debe detectar invariantes,
transacciones, SQL, migraciones, contratos HTTP y flujos críticos con el coste y
la velocidad adecuados a cada riesgo.

## Contexto y restricciones

- ADR-0001 exige mantener el dominio independiente de infraestructura.
- ADR-0009 fija REST y OpenAPI contract-first; los contratos deben conservar
  compatibilidad verificable.
- ADR-0011 establece PostgreSQL, `pgx`, `sqlc` y `goose`; consultas,
  restricciones y transacciones requieren evidencia contra PostgreSQL real.
- ADR-0012 ya proporciona `go test`, `make test` y `make test-race`, pero no
  decide librerías, aislamiento de datos ni gates de calidad.
- ADR-0018 proporciona PostgreSQL local con Compose; la base de desarrollo no
  debe usarse para pruebas de integración.
- No existe todavía un vertical slice, API implementada ni cliente universal.
  Esta decisión no autoriza tests vacíos, librerías, una matriz de dispositivos,
  CI ni pruebas de carga.

## Criterios de decisión

1. detectar el comportamiento de mayor riesgo con evidencia fiel;
2. ofrecer feedback rápido y diagnosticable durante el desarrollo;
3. mantener datos y pruebas aislados y reproducibles;
4. minimizar dependencias y convenciones antes de que exista una necesidad real;
5. permitir que API, cliente y operación añadan evidencia específica al aparecer;
6. mantener un coste razonable para un equipo pequeño.

## Alternativas

### Alternativa A — Pruebas unitarias centradas en mocks

Usar principalmente dobles de puertos, `pgx` y HTTP; reservar la integración
real para comprobaciones manuales o muy puntuales.

- **Ventajas:** ejecución rápida y sin dependencia de PostgreSQL para la mayoría
  de pruebas.
- **Inconvenientes:** no demuestra SQL, restricciones, migraciones, aislamiento
  ni comportamiento real de transacciones; mantiene muchos dobles.
- **Coste de adopción:** bajo.
- **Coste de mantenimiento:** medio o alto al evolucionar consultas y puertos.
- **Riesgos:** confianza falsa y desacople entre mocks y PostgreSQL real.

### Alternativa B — Pruebas por riesgo y capas

Usar pruebas unitarias rápidas para dominio y casos de uso; integración contra
PostgreSQL real para persistencia; pruebas de contrato para fronteras externas;
y end-to-end mínimas para flujos críticos. Cada prueba se elige por el riesgo que
reduce, no por una cuota de cobertura.

- **Ventajas:** aplica dependencias reales donde su semántica importa y conserva
  un bucle rápido para reglas puras; localiza mejor los fallos que una suite E2E
  extensa.
- **Inconvenientes:** exige distinguir capas y administrar una base de pruebas
  efímera.
- **Coste de adopción:** medio y progresivo con el primer comportamiento real.
- **Coste de mantenimiento:** medio y predecible; cada capa mantiene solo la
  evidencia que le corresponde.
- **Riesgos:** convertir las capas en una pirámide rígida o añadir helpers y
  frameworks antes de tener repetición que los justifique.

### Alternativa C — Pruebas end-to-end como evidencia principal

Probar la mayoría del comportamiento arrancando el sistema completo y recorriendo
la interfaz o la API pública.

- **Ventajas:** evidencia cercana a la experiencia final y poco conocimiento de
  la estructura interna.
- **Inconvenientes:** ejecución lenta, diagnósticos ambiguos, setup costoso y
  casos límite difíciles de construir.
- **Coste de adopción:** medio o alto.
- **Coste de mantenimiento:** alto al crecer API, cliente y datos.
- **Riesgos:** suites inestables que retrasan el feedback y no cubren bien reglas
  o fallos de persistencia específicos.

### No cambiar

Conservar solo principios genéricos y elegir herramientas caso por caso sin una
estrategia común.

- **Consecuencias:** no define qué evidencia es suficiente para cada riesgo;
  favorece cobertura accidental, duplicación y decisiones inconsistentes.

## Comparación

La alternativa A optimiza velocidad aislando precisamente la infraestructura cuya
semántica PostgreSQL se ha decidido aprender y operar. La C da confianza en un
flujo completo, pero su coste y diagnóstico son desproporcionados como base.

La alternativa B reserva las pruebas reales para SQL, transacciones, migraciones
y contratos, y mantiene pruebas pequeñas para reglas puras. La biblioteca
estándar de Go ya cubre la base de pruebas unitarias y HTTP; agregar una librería
de assertions, mocks o contenedores no aporta valor hasta que haya una necesidad
repetida y medible.

## Recomendación

**Opinión/recomendación:** alternativa B. Es la solución mínima que relaciona
cada riesgo con evidencia adecuada, sin forzar una proporción de tests ni añadir
un framework de pruebas prematuro.

## Decisión del usuario

**Aceptada:** alternativa B, con estas reglas:

- dominio y casos de uso usarán pruebas unitarias rápidas con `testing` de Go;
- los dobles se limitarán a puertos con comportamiento externo relevante, nunca
  como sustitutos de PostgreSQL;
- consultas, restricciones, transacciones, migraciones y generación `sqlc` se
  validarán contra PostgreSQL real;
- la integración utilizará una base de pruebas efímera, creada y migrada desde
  vacío por ejecución, distinta de la base de desarrollo;
- handlers HTTP se probarán con `httptest` y los contratos con validación y
  generación determinista de OpenAPI;
- las pruebas end-to-end se limitarán a flujos críticos completos cuando exista
  una API y cliente funcionales;
- cobertura es una señal secundaria; no se fija un porcentaje global;
- cada corrección de bug añadirá una reproducción automatizada cuando sea viable;
- las herramientas de cliente, matriz de dispositivos, CI, carga, resiliencia y
  seguridad se decidirán cuando exista el riesgo y el componente correspondiente.

## Consecuencias

### Positivas

- Las reglas de negocio reciben feedback rápido sin acoplarse a infraestructura.
- La persistencia se valida con restricciones y semántica reales de PostgreSQL.
- Las migraciones y contratos tienen una ruta explícita de validación.
- Las suites completas permanecen pequeñas y centradas en el comportamiento
  crítico.

### Negativas y deuda aceptada

- Hay que preparar y limpiar una base de pruebas por ejecución.
- Algunos dobles y fixtures pequeños se escribirán localmente antes de justificar
  una dependencia específica.
- Las pruebas E2E, de cliente y de operación no estarán disponibles hasta que
  existan los componentes que las hacen significativas.

## Validación

Con el primer vertical slice se demostrará que:

- una invariante de dominio falla y pasa mediante una prueba unitaria;
- una migración alcanza la última versión desde una base de pruebas vacía y se
  repite sin cambios cuando ya está aplicada;
- una restricción, consulta o rollback de transacción se verifica contra
  PostgreSQL real;
- una incompatibilidad entre SQL y esquema falla en `sqlc` o compilación;
- un handler HTTP se prueba con `httptest` y respeta el contrato definido;
- un bug corregido queda reproducido de forma automatizada cuando sea viable;
- la base de desarrollo no recibe datos de las pruebas de integración.

## Disparadores de revisión

- los helpers locales de pruebas se duplican o resultan insuficientes de forma
  sostenida;
- la creación de la base efímera supera el presupuesto de feedback que se fije
  con evidencia;
- concurrencia, integración externa o despliegues exigen aislamiento adicional;
- aparece una regresión que las capas elegidas no detectan de forma razonable;
- el cliente universal necesita una matriz o herramientas específicas;
- CI, carga, seguridad u observabilidad fijan requisitos de evidencia distintos.

## Documentación afectada

- [TESTING.md](../engineering/TESTING.md)
- [TECHNICAL_BASELINE.md](../governance/TECHNICAL_BASELINE.md)
- [DECISIONS.md](../governance/DECISIONS.md)
- [DECISIONS_TO_REVISIT.md](../governance/DECISIONS_TO_REVISIT.md)
- [LEARNING.md](../project/LEARNING.md)
- [CHANGELOG.md](../../CHANGELOG.md)

## Fuentes técnicas

- [Go: paquete `testing`](https://go.dev/pkg/testing/)
- [Go: `net/http/httptest`](https://pkg.go.dev/net/http/httptest)
- [PostgreSQL: características de transacción](https://www.postgresql.org/docs/current/sql-set-transaction.html)
- [PostgreSQL: `ROLLBACK TO SAVEPOINT`](https://www.postgresql.org/docs/current/sql-rollback-to.html)
- [GitHub Actions: servicio PostgreSQL](https://docs.github.com/en/actions/tutorials/use-containerized-services/create-postgresql-service-containers)
