# Retrospectiva técnica — Fase 2

- **Fecha:** 2026-08-21
- **Objetivo:** entregar un primer vertical slice de negocio en Go que
  atravesara dominio, identidad, persistencia, API y pruebas sin acoplar la
  lógica de negocio a infraestructura.
- **Participantes:** Usuario y Codex.

## Resultado frente al objetivo

La fase cumple y supera su criterio de salida. El incremento mínimo aceptado en
ADR-0043 quedó demostrado: desde una base vacía, una cuenta puede registrarse,
verificarse, establecer una sesión, transferir un borrador local, publicar una
liga con equipos y consultarla mediante sus relaciones autenticadas o su ID
público no listado. La autorización impide que otra cuenta la administre.

El recorrido mantiene la dirección de dependencias acordada. Los casos de uso
Go declaran puertos concretos; HTTP, PostgreSQL, Google y SMTP son adaptadores;
`cmd/api` compone las dependencias. OpenAPI 3.1 es la fuente del contrato HTTP,
el cliente TypeScript se genera con Orval y PostgreSQL ejecuta SQL escrito por
el equipo mediante `pgx` y código tipado de `sqlc`.

Tras cumplir el primer incremento, Fase 2 continuó con capacidades ya decididas:
login Google, recuperación y gestión de métodos de acceso, rotación de sesiones,
inicio y ciclo completo de liga, resultados con historial, clasificación,
co-campeones, bajas, cancelación, administración delegada, transferencia y
notificaciones internas. Estas ampliaciones aportan producto y pruebas, pero no
se reinterpretan como requisitos necesarios para haber cerrado el slice
original.

Quedan fuera Apple federado, almacenamiento de objetos, cache, migraciones sobre
datos no descartables, backup/restauración de producción, pruebas de carga y un
SLA de producción. Tienen decisiones o disparadores propios y no invalidan el
cierre.

## Decisiones

- **Funcionaron:** el monolito modular permitió transacciones locales y límites
  por capacidad sin introducir red distribuida; REST/OpenAPI contract-first
  mantuvo contrato, backend y cliente alineados; PostgreSQL conservó invariantes
  y concurrencia donde la semántica real importaba.
- **Coste inesperado:** el alcance siguió creciendo después de cumplir el slice y
  la retrospectiva se pospuso. El resultado es más completo, pero se perdió una
  frontera temprana para revisar aprendizaje y deuda. La documentación de estado
  quedó desfasada respecto a la implementación.
- **Revisar:** activar Goose cuando exista un esquema compartido que deba
  evolucionar sin reset; revisar límites del monolito solo ante evidencia de
  autonomía, escala o seguridad; ampliar protección de abuso y pruebas de carga
  cuando haya tráfico representativo.
- **ADR ausentes:** ninguno para el criterio de salida. El cierre documental de
  fase era la pieza pendiente; las ampliaciones relevantes cuentan con ADR
  aceptados hasta ADR-0095 y sus decisiones relacionadas.

## Aprendizaje

- Un vertical slice no es una capa horizontal: necesita atravesar actor,
  autorización, regla, transacción, contrato y respuesta observable.
- Clean/hexagonal aporta valor cuando el puerto protege un límite real. No se
  necesita un repositorio genérico, una interfaz por tabla ni una capa simétrica.
- Contract-first evita deriva solo si lint, generación y diff del cliente forman
  parte de la puerta de calidad.
- Los mocks no demuestran restricciones, bloqueos ni atomicidad de PostgreSQL;
  las transacciones de mayor riesgo necesitan integración con una base real.
- Cerrar una fase al alcanzar su salida no impide continuar el producto: crea
  un punto explícito para distinguir evidencia obtenida de alcance posterior.

## Calidad profesional

- **Seguridad:** Argon2id protege contraseñas; sesiones y tickets son opacos,
  rotatorios, revocables y de un uso cuando corresponde. CORS, CSRF, límites de
  tasa, reautenticación y errores seguros se resuelven en sus fronteras sin
  revelar existencia de cuentas, credenciales ni detalles internos.
- **Pruebas:** casos de uso y handlers usan `testing` y `httptest`; Google se
  valida con JWT/JWKS locales; PostgreSQL real cubre transacciones de creación,
  resultados, clasificación, cierre, concurrencia, identidad y transferencia.
  `make verify` comprueba además formato, lint, generación, build y
  vulnerabilidades.
- **Observabilidad:** el criterio inicial definió errores y correlación sin
  invadir el dominio; la instrumentación completa y su validación pertenecen al
  cierre posterior de Fase 3.
- **Operación y recuperación:** el esquema se aplica explícitamente desde cero,
  la API no migra al arrancar y `dev` separa propietario, migrador y runtime. El
  reset sigue siendo válido solo mientras los datos afectados sean descartables.
- **Coste:** no se introdujeron microservicios, Redis/Valkey, almacenamiento de
  objetos ni Kubernetes para construir el dominio. Go, PostgreSQL y un proceso
  desplegable fueron suficientes.
- **Documentación:** producto, API, datos, seguridad, pruebas, ADR y aprendizaje
  evolucionaron junto a cada capacidad; el estado de fase y algunas listas de
  pendientes necesitaron esta reconciliación final.

## Complejidad

El monolito modular y los puertos específicos demostraron valor. La composición
manual en `cmd/api` sigue siendo legible y no justifica un contenedor de
inyección. `pgx` y `sqlc` preservan SQL explícito sin un ORM; la excepción de
esquema inicial evita operar migraciones artificiales mientras los datos sean
reseteables.

La principal complejidad acumulada está en la amplitud funcional del backend y
su adaptador HTTP central. Antes de dividir procesos o crear abstracciones se
debe medir acoplamiento, tiempos de cambio, carga o aislamiento; el tamaño por sí
solo no autoriza microservicios.

## Acciones

| Acción | Propietario | Disparador | Destino |
| --- | --- | --- | --- |
| Activar migraciones incrementales con Goose | Usuario/Codex | Primer dato compartido que no pueda resetearse | ADR-0072, [DATABASE.md](../engineering/DATABASE.md) y runbook |
| Probar backup y restauración | Usuario/Codex | Antes de tratar datos de producción como durables | [DATABASE.md](../engineering/DATABASE.md) y runbook nuevo |
| Revisar arquitectura modular | Usuario/Codex | Dependencia cruzada, despliegue autónomo o cuello medido | ADR-0007 y [ARCHITECTURE.md](../engineering/ARCHITECTURE.md) |
| Cerrar cada incremento en cuanto cumpla su salida | Usuario/Codex | Siguiente criterio de fase o hito | [ROADMAP.md](ROADMAP.md) y playbook de retrospectiva |

## Cierre

La Fase 2 queda cerrada. El backend demuestra negocio, autorización,
persistencia, contratos y pruebas de extremo a extremo con una arquitectura
pragmática y una única unidad de despliegue. Las ampliaciones posteriores no
cambian el criterio original; aportan evidencia adicional sobre la misma base.
