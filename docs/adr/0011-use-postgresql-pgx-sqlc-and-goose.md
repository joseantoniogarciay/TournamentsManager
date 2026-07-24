# ADR-0011: Usar PostgreSQL con pgx, sqlc y goose

- **Estado:** Aceptado
- **Fecha:** 2026-07-24
- **Decisor:** Usuario, mediante elección explícita de la alternativa B
- **Propietario del análisis:** Asistente como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

El backend Go necesita persistir identidad y, más adelante, torneos manteniendo
integridad, transacciones y consultas observables. Hay que decidir el motor
principal, cómo acceder a él desde Go y cómo evolucionar el esquema sin ocultar
la semántica de PostgreSQL ni acoplar el dominio a infraestructura.

## Contexto y restricciones

- El manifiesto establece PostgreSQL como stack objetivo y exige que local se
  parezca a producción.
- El backend será un monolito modular en Go.
- La arquitectura clean/hexagonal pragmática impide que el dominio dependa de la
  base de datos o sus librerías.
- El proyecto prioriza aprender SQL, transacciones, restricciones, concurrencia y
  operación.
- Una única base puede ser compartida inicialmente, pero los módulos conservarán
  propiedad explícita de sus datos.
- No existe un requisito de portabilidad entre motores SQL.
- No se añadirá cache sin un problema medido.

## Criterios de decisión

1. hacer visible y comprensible el comportamiento de PostgreSQL;
2. conservar tipado y feedback temprano en Go;
3. mantener el dominio independiente de persistencia;
4. reducir código mecánico sin ocultar consultas;
5. permitir transacciones y características propias de PostgreSQL;
6. hacer reproducibles y revisables los cambios de esquema;
7. mantener un coste razonable para un equipo pequeño.

## Alternativas

### Alternativa A — pgx con SQL y mapeo manual

El adaptador utiliza `pgx` nativo y escribe manualmente consultas, structs,
llamadas a `Scan` y mapeos.

- **Ventajas:** control completo, pocas herramientas y aprendizaje directo.
- **Inconvenientes:** código repetitivo y riesgo de desalinear columnas, tipos y
  orden de escaneo.
- **Coste de adopción:** bajo.
- **Coste de mantenimiento:** medio o alto al crecer el número de consultas.
- **Riesgos:** duplicación y errores mecánicos detectados tarde.

### Alternativa B — pgx con SQL tipado por sqlc y migraciones goose

El equipo escribe SQL; `sqlc` valida consultas contra el esquema y genera el
código Go de acceso. `pgx` proporciona la interfaz PostgreSQL nativa y `goose`
aplica migraciones SQL incrementales.

- **Ventajas:** SQL explícito, código Go tipado, menos boilerplate y migraciones
  revisables.
- **Inconvenientes:** añade dos herramientas y un ciclo de generación; las
  consultas muy dinámicas pueden requerir variantes o código específico.
- **Coste de adopción:** medio.
- **Coste de mantenimiento:** medio y predecible si la generación es
  determinista.
- **Riesgos:** filtrar tipos generados al dominio, tratar generación como magia o
  multiplicar consultas específicas sin criterio.

### Alternativa C — ORM convencional

GORM u otro ORM representa tablas y asociaciones mediante estructuras Go y
construye gran parte del SQL.

- **Ventajas:** productividad alta para CRUD, asociaciones y consultas comunes.
- **Inconvenientes:** SQL y carga de relaciones menos visibles; convenciones,
  hooks y comportamiento del ORM se convierten en otra superficie que aprender.
- **Coste de adopción:** bajo o medio.
- **Coste de mantenimiento:** medio; aumenta al escapar a SQL específico.
- **Riesgos:** N+1, carga excesiva y acoplamiento entre modelo ORM y dominio.

### Alternativa D — Framework code-first

Ent y su ecosistema describen el esquema en Go y generan entidades, relaciones,
consultas y migraciones.

- **Ventajas:** API tipada y potente para modelos con muchas relaciones.
- **Inconvenientes:** DSL, generación y framework más amplios antes de conocer el
  dominio.
- **Coste de adopción:** alto.
- **Coste de mantenimiento:** medio o alto.
- **Riesgos:** sobreingeniería y dependencia fuerte del modelo del framework.

## Comparación

La alternativa A maximiza visibilidad, pero conserva trabajo mecánico que no
aporta aprendizaje proporcional cuando aumentan las consultas. Las alternativas
C y D aceleran CRUD, pero esconden parte del SQL que este proyecto quiere
comprender y añaden convenciones propias.

La alternativa B no elimina la complejidad de datos: obliga a escribir consultas,
decidir transacciones y analizar planes. Su ventaja es trasladar desajustes entre
SQL y Go hacia generación o compilación sin introducir un ORM completo.

## Recomendación

**Opinión/recomendación:** alternativa B: PostgreSQL, interfaz nativa de `pgx`,
`sqlc` y migraciones SQL versionadas mediante `goose`.

Es la solución mínima que combina aprendizaje explícito de SQL con reducción de
boilerplate. `database/sql` con `pgx/stdlib` se reservará para una dependencia
real que exija `*sql.DB`; la portabilidad teórica del driver no justifica perder
la interfaz nativa cuando solo se ha elegido PostgreSQL.

## Decisión del usuario

**Aceptada:** utilizar:

- PostgreSQL como sistema de registro relacional principal;
- `pgx` mediante su interfaz nativa para conexiones, pool y transacciones;
- `sqlc` para validar SQL y generar código Go tipado;
- `goose` para migraciones SQL incrementales y versionadas.

La elección no autoriza todavía versiones exactas, estructura de paquetes,
esquema, consultas ni parámetros del pool.

## Reglas de implementación

- El dominio y los casos de uso no importarán `pgx`, `sqlc`, `goose` ni tipos
  PostgreSQL.
- SQL, tipos generados y mapeos permanecerán en adaptadores de persistencia.
- El equipo seguirá siendo autor de las consultas; el código generado no se
  editará manualmente.
- No se creará un repositorio genérico ni una interfaz por cada tabla.
- Las transacciones se definirán desde invariantes del caso de uso.
- PostgreSQL aplicará restricciones de integridad cuando pueda expresar la
  invariante sin sustituir las reglas del dominio.
- Las migraciones iniciales serán SQL; no se usarán migraciones Go sin una
  necesidad documentada.
- `goose` se ejecutará como paso separado del arranque normal de la API.
- No se usarán migraciones automáticas de ORM.
- Una migración aplicada en un entorno compartido no se reescribe; se añade otra.
- Un rollback destructivo no se considera seguro solo porque exista `Down`; se
  documentará rollback, forward-fix o evolución expand/contract según el cambio.
- Cada módulo será propietario de sus tablas aunque comparta la instancia.
- Redis o Valkey no forman parte de esta decisión.

## Consecuencias

### Positivas

- Las consultas y capacidades específicas de PostgreSQL permanecen visibles.
- Muchos errores de columnas y tipos se detectan antes de ejecutar la aplicación.
- Se reduce el código repetitivo de consultas sin introducir un ORM completo.
- El esquema y sus cambios quedan versionados, revisables y reproducibles.
- El driver, la generación y las migraciones respetan el límite de
  infraestructura.

### Negativas y deuda aceptada

- El equipo debe aprender SQL, PostgreSQL, `pgx`, configuración de `sqlc` y
  operación de `goose`.
- La generación añade una herramienta que debe fijarse y automatizarse.
- Las consultas con filtros altamente dinámicos pueden resultar verbosas.
- Existirán mapeos entre filas generadas y modelos de dominio.
- Las migraciones de producción exigirán políticas operativas adicionales.

## Validación

Antes del primer vertical slice deberá demostrarse:

- generación determinista de `sqlc` y diff limpio sin cambios de entrada;
- migración desde una base PostgreSQL vacía hasta la última versión;
- detección de una consulta incompatible con el esquema;
- prueba de integración de consultas y restricciones contra PostgreSQL real;
- una transacción de caso de uso con rollback ante fallo;
- repetición segura del comando de migración cuando no hay cambios;
- ausencia de imports de persistencia en dominio y casos de uso;
- procedimiento documentado de avance, fallo y recuperación de una migración.

## Decisiones pendientes

- versiones exactas y forma reproducible de instalar las herramientas;
- organización de esquema, migraciones, consultas y código generado;
- si el código generado se versiona o se verifica solo mediante CI;
- identificadores, nullability, timestamps y convenciones SQL;
- límites transaccionales e idempotencia por caso de uso;
- configuración del pool, timeouts y tratamiento de errores PostgreSQL;
- datos de pruebas, semillas, backup y restauración;
- política de migración durante despliegues y cambios incompatibles;
- observabilidad y análisis de consultas.

## Disparadores de revisión

- `sqlc` no puede representar consultas necesarias de forma mantenible.
- Las consultas dinámicas producen duplicación o complejidad sostenida.
- Una dependencia crítica exige `database/sql`.
- Aparece un requisito real de soportar otro motor.
- El volumen de cambios requiere generación, lint o detección de drift más
  avanzada que `goose`.
- Un incidente revela fallos en migraciones, transacciones o aislamiento.

## Documentación afectada

- [DATABASE.md](../engineering/DATABASE.md)
- [ARCHITECTURE.md](../engineering/ARCHITECTURE.md)
- [DEVELOPMENT.md](../engineering/DEVELOPMENT.md)
- [TESTING.md](../engineering/TESTING.md)
- [TECHNICAL_BASELINE.md](../governance/TECHNICAL_BASELINE.md)
- [SYSTEM_OPTIONS.md](../governance/SYSTEM_OPTIONS.md)
- [DECISIONS_TO_REVISIT.md](../governance/DECISIONS_TO_REVISIT.md)

## Fuentes técnicas

- [PostgreSQL: definición de datos y restricciones](https://www.postgresql.org/docs/current/ddl.html)
- [pgx: driver y toolkit PostgreSQL para Go](https://github.com/jackc/pgx)
- [sqlc: generación de Go tipado desde SQL](https://docs.sqlc.dev/en/stable/)
- [sqlc: uso de transacciones](https://docs.sqlc.dev/en/latest/howto/transactions.html)
- [goose: migraciones SQL y Go](https://github.com/pressly/goose)
- [GORM: guía oficial](https://gorm.io/docs/)
- [Ent: generación de código](https://entgo.io/docs/code-gen/)
