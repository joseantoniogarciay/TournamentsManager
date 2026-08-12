# Technical Baseline

> Estado: cerrado el 2026-07-25.
>
> Objetivo: confirmar la base técnica antes de diseñar el comportamiento detallado
> del producto. No autoriza scaffolding ni código por sí mismo.

## Reglas del gate

1. Se toma una decisión por vez, siguiendo sus dependencias.
2. Cada análisis incluye alternativas, ventajas, inconvenientes, mantenimiento y
   recomendación.
3. El usuario acepta la decisión.
4. Se registra el ADR.
5. Solo entonces se prepara el siguiente análisis.
6. Las versiones exactas se fijan al crear el entorno reproducible, con política
   de actualización, no como números aislados.
7. Una elección que necesite comportamiento de negocio se marca como bloqueada.

## Estado

| Orden | Decisión                                              | Estado    | Registro                                                                    |
| ----- | ----------------------------------------------------- | --------- | --------------------------------------------------------------------------- |
| 0     | Control de versiones: Git                             | Aceptada  | [ADR-0003](../adr/0003-use-git-for-version-control.md)                      |
| 0     | Flujo de ramas: `develop` como integración            | Aceptada  | [ADR-0013](../adr/0013-use-develop-as-integration-branch.md)                |
| 0     | Arquitectura clean/hexagonal pragmática               | Aceptada  | [ADR-0001](../adr/0001-pragmatic-clean-architecture.md)                     |
| 1     | Topología de repositorios: monorepo                   | Aceptada  | [ADR-0005](../adr/0005-use-a-product-monorepo.md)                           |
| 1     | GitHub público y secretos fuera de Git                | Aceptada  | [ADR-0006](../adr/0006-public-github-repository-security-boundary.md)       |
| 2     | Topología del backend: monolito modular               | Aceptada  | [ADR-0007](../adr/0007-use-a-modular-monolith-backend.md)                   |
| 3     | Estrategia web y mobile: cliente universal            | Aceptada  | [ADR-0008](../adr/0008-use-a-universal-react-native-client.md)              |
| 4     | Estilo y contrato de API: REST/OpenAPI contract-first | Aceptada  | [ADR-0009](../adr/0009-use-rest-and-openapi-contract-first.md)              |
| 5     | Identidad propia con Apple/Google federados           | Aceptada  | [ADR-0010](../adr/0010-own-identity-with-federated-login.md)                |
| 6     | Persistencia y acceso a datos                         | Aceptada  | [ADR-0011](../adr/0011-use-postgresql-pgx-sqlc-and-goose.md)                |
| 7     | Toolchain Go                                          | Aceptada  | [ADR-0012](../adr/0012-pin-go-toolchain-and-isolate-tools.md)               |
| 8     | Toolchain TypeScript                                  | Aceptada  | [ADR-0014](../adr/0014-use-node-pnpm-and-strict-typescript.md)              |
| 9     | Framework del cliente universal: Expo                 | Aceptada  | [ADR-0015](../adr/0015-use-expo-router-and-continuous-native-generation.md) |
| 10    | Routing y generación nativa: Expo Router + CNG        | Aceptada  | [ADR-0015](../adr/0015-use-expo-router-and-continuous-native-generation.md) |
| 11    | Rendering y adaptación por plataforma                 | Aceptada  | [ADR-0016](../adr/0016-use-client-side-web-rendering-initially.md)          |
| 12    | Configuración y secretos                              | Aceptada  | [ADR-0017](../adr/0017-use-env-contracts-github-environments-and-oidc.md)   |
| 13    | Entorno local                                         | Aceptada  | [ADR-0076](../adr/0076-run-the-local-api-in-compose-with-air.md)            |
| 14    | Estrategia de pruebas                                 | Aceptada  | [ADR-0019](../adr/0019-use-risk-based-layered-testing.md)                   |
| 15    | Observabilidad mínima                                 | Aceptada  | [ADR-0020](../adr/0020-use-minimal-correlated-observability.md)             |
| 16    | CI y política de calidad                              | Aceptada  | [ADR-0021](../adr/0021-use-advisory-ci-with-local-quality-gate.md)          |
| 17a   | Empaquetado de la API                                 | Aceptada  | [ADR-0022](../adr/0022-package-backend-as-oci-image.md)                    |
| 17b   | Runtime cloud de la API                               | Superada  | [ADR-0088](../adr/0088-use-ephemeral-aws-learning-and-home-runtime.md)      |
| 17c   | Registry y promoción de la API                        | Aceptada  | [ADR-0024](../adr/0024-use-ecr-and-digest-based-image-promotion.md)         |
| 18a   | Herramienta de IaC: Terraform                         | Aceptada  | [ADR-0025](../adr/0025-use-terraform-for-infrastructure-as-code.md)         |
| 18b   | Fundación AWS: cuentas e identidad                    | Aceptada  | [ADR-0026](../adr/0026-use-aws-organizations-and-temporary-identities.md)   |
| 18c   | Estado local antes de infraestructura cloud            | Aceptada  | [ADR-0027](../adr/0027-keep-local-state-until-first-cloud-apply.md)         |
| 18d   | Backend remoto y bootstrap de Terraform                | Aceptada | [ADR-0028](../adr/0028-use-hcp-terraform-free-for-remote-state.md)           |
| 18e1  | Entrada y egress inicial de la red AWS                | Aceptada | [ADR-0029](../adr/0029-use-public-alb-restricted-fargate-and-no-nat-initially.md) |
| 18e2  | Región, VPC, subredes y límites de coste              | Aceptada | [ADR-0030](../adr/0030-use-spain-region-and-two-az-cost-gated-network.md)   |
| 19    | Kubernetes local/cloud                                | Aplazada  | Fase 4 del manifiesto                                                       |

El gate técnico está cerrado: no quedan decisiones bloqueantes pendientes. Las
direcciones del manifiesto continúan sin autorizar recursos ni despliegues por sí
solas; los gates de producto y de cada fase siguen requiriendo sus decisiones,
validación y autorización de coste.

## Decisión 1 — Topología de repositorios — aceptada

### Problema

Backend, web, mobile, infraestructura, contratos y handbook evolucionarán a ritmos
distintos, pero pertenecen al mismo producto y inicialmente al mismo equipo. Hay
que decidir cuántos repositorios Git serán fuente de verdad.

### Criterios

En orden:

1. trazabilidad entre contrato, implementación y documentación;
2. simplicidad para aprender y operar;
3. reutilización controlada entre web y mobile;
4. independencia de builds y despliegues;
5. rendimiento y mantenibilidad del tooling;
6. posibilidad de separar componentes en el futuro.

### Alternativa A — Monorepo de producto

Un repositorio contiene handbook, backend, web, mobile e infraestructura.

**Ventajas**

- cambios atómicos de API, clientes, pruebas y documentación;
- un único historial, onboarding y política de calidad;
- compartir contratos no exige publicar paquetes desde el primer día;
- encaja con un equipo pequeño y una arquitectura inicialmente simple.

**Inconvenientes**

- varios ecosistemas conviven en el mismo árbol;
- CI debe detectar qué ha cambiado;
- puede tentar a compartir código o acoplar despliegues sin necesidad;
- si crece mucho, puede requerir tooling adicional.

**Mantenimiento**

Inicialmente bajo con Git, Go y workspaces del package manager. Nx, Turborepo u
otra capa se añadiría solo ante tiempos de build o coordinación medidos.

### Alternativa B — Un repositorio por aplicación

Repositorios separados para API, web, mobile e infraestructura.

**Ventajas**

- límites y pipelines independientes;
- permisos y releases aislados;
- menor mezcla de toolchains por repositorio.

**Inconvenientes**

- cambios de contrato coordinados entre varios historiales;
- paquetes compartidos necesitan versionado y distribución;
- más configuración, bots, permisos y mantenimiento;
- documentación transversal puede quedar desalineada.

**Mantenimiento**

Mayor desde el primer día: varios pipelines, dependencias entre repositorios y
estrategia de compatibilidad.

### Alternativa C — Híbrida

Un repositorio para backend/infra y otro para web/mobile, con handbook en uno de
ellos.

**Ventajas**

- separa Go/cloud del ecosistema TypeScript;
- web y mobile comparten packages con facilidad.

**Inconvenientes**

- el contrato cruza repositorios;
- la fuente de verdad documental resulta menos natural;
- mantiene buena parte de la coordinación de multirepo sin aislamiento total.

**Mantenimiento**

Intermedio; exige versionar el contrato o automatizar compatibilidad desde el
inicio.

### Recomendación

**Opinión:** alternativa A, monorepo de producto.

Es la solución con menor coste para un equipo pequeño, permite commits atómicos y
mantiene el handbook junto al sistema. Monorepo no significa un único build ni un
único despliegue: cada aplicación conservará dependencias, pruebas y artefactos
independientes.

La estructura concreta se decidirá después; aceptar monorepo no implica aceptar
Nx, Turborepo, pnpm ni una convención `apps/packages`.

### Decisión del usuario

**Aceptada:** alternativa A, monorepo de producto. Véase
[ADR-0005](../adr/0005-use-a-product-monorepo.md).

La publicación será un repositorio público en GitHub con los límites definidos en
[ADR-0006](../adr/0006-public-github-repository-security-boundary.md).

## Decisión 2 — Topología del backend — aceptada

### Problema

El backend debe ofrecer una base profesional para aprender dominio, persistencia,
seguridad, observabilidad y despliegue. Hay que decidir cuántas unidades de
ejecución y despliegue existirán inicialmente.

### Criterios

En orden:

1. simplicidad y capacidad de comprender el sistema completo;
2. límites internos claros y verificables;
3. facilidad para probar transacciones y reglas de negocio;
4. operabilidad local y en producción;
5. coste de infraestructura y mantenimiento;
6. capacidad de evolucionar con evidencia.

### Alternativa A — Monolito modular

Un servicio Go desplegable contiene módulos internos por capacidad. Los módulos
mantienen límites y dependencias explícitas, pero pueden compartir proceso y
PostgreSQL.

**Ventajas**

- una unidad de build, despliegue y observación;
- transacciones locales y debugging directo;
- entorno local sencillo;
- permite aprender arquitectura sin introducir red distribuida;
- los límites pueden extraerse más adelante.

**Inconvenientes**

- una regresión o despliegue afecta a toda la unidad;
- escalado conjunto;
- requiere disciplina para impedir acoplamiento entre módulos;
- puede degradar en un monolito desestructurado si no se verifican dependencias.

**Mantenimiento**

Bajo inicialmente. La complejidad principal es conservar límites internos y
pruebas de arquitectura.

### Alternativa B — Microservicios

Cada capacidad se ejecuta y despliega como servicio independiente.

**Ventajas**

- despliegue, escalado y propiedad independientes;
- aislamiento de fallos potencial;
- límites explícitos por red.

**Inconvenientes**

- consistencia distribuida, retries, idempotencia y compatibilidad de contratos;
- observabilidad y debugging entre servicios;
- varios pipelines, imágenes, configuraciones y runbooks;
- mayor coste local y cloud antes de conocer el dominio.

**Mantenimiento**

Alto desde el primer caso de uso. La red convierte fallos locales en fallos
parciales y obliga a operar una plataforma distribuida.

### Alternativa C — Funciones serverless

Casos de uso o eventos se despliegan como funciones administradas.

**Ventajas**

- infraestructura de ejecución gestionada;
- escalado por demanda;
- coste bajo en cargas esporádicas.

**Inconvenientes**

- fragmentación de código, configuración y observabilidad;
- límites de runtime, cold starts y debugging;
- mayor acoplamiento al proveedor;
- menor semejanza con el camino Docker/Kubernetes del manifiesto.

**Mantenimiento**

Intermedio o alto según número de funciones e integraciones. Reduce gestión de
servidores, no la complejidad del sistema.

### Recomendación

**Opinión:** alternativa A, monolito modular en Go.

Ofrece la mejor relación entre aprendizaje y complejidad. “Monolito” describe la
unidad de despliegue; “modular” exige límites internos. Empezar así no impide
extraer un worker o servicio cuando métricas, seguridad, carga o autonomía lo
justifiquen.

La aceptación no decidiría todavía paquetes, módulos de negocio, framework HTTP,
acceso a datos ni estructura de carpetas.

### Decisión del usuario

**Aceptada:** alternativa A, monolito modular en Go. Véase
[ADR-0007](../adr/0007-use-a-modular-monolith-backend.md).

## Decisión 3 — Estrategia web y mobile — aceptada

### Problema

El producto tendrá web y mobile. Hay que decidir si se construyen como una
aplicación universal, como aplicaciones especializadas que comparten piezas o de
forma secuencial empezando solo por web.

Esta decisión no selecciona todavía el framework universal concreto —Expo es
candidato—, navegación, rendering, estado ni design system.

### Criterios

En orden:

1. calidad de la experiencia pública web y de la experiencia nativa;
2. reutilización sin forzar abstracciones;
3. independencia de builds y despliegues;
4. mantenibilidad para un equipo pequeño;
5. compatibilidad con CI/CD del monorepo;
6. capacidad de evolucionar por plataforma.

### Alternativa A — Aplicación universal React Native

Una aplicación y un sistema de rutas sirven Android, iOS y web mediante React
Native for Web; Expo es el candidato natural.

**Ventajas**

- máxima reutilización inicial de pantallas y navegación;
- un único proyecto TypeScript;
- comportamiento y features coordinados entre plataformas.

**Inconvenientes**

- la web pública puede necesitar rendering, SEO y layouts diferentes;
- tablas, administración y accesibilidad web pueden exigir excepciones;
- cada diferencia de plataforma introduce condiciones o archivos específicos;
- compartir UI puede convertirse en objetivo en vez de consecuencia.

**Mantenimiento**

Bajo si las experiencias son realmente similares; creciente si web y native
divergen.

### Alternativa B — Aplicaciones especializadas con packages compartidos

Una aplicación React web y una aplicación React Native mobile tienen navegación,
entrypoint y UI propios. Comparten contratos, cliente API, validación, lógica pura
y tokens de diseño cuando aporten valor.

**Ventajas**

- web optimizable para contenido público, SEO y pantallas densas;
- mobile conserva navegación y experiencia nativas;
- builds, releases y despliegues independientes;
- reutilización deliberada, no obligatoria;
- encaja con el monorepo aceptado.

**Inconvenientes**

- dos aplicaciones y dos rutas de CI/CD;
- parte de la UI y de las pruebas se duplica;
- exige gobernar packages compartidos sin crear un “common” indiscriminado.

**Mantenimiento**

Moderado y explícito. Cada plataforma paga por sus diferencias reales.

### Alternativa C — Web/PWA primero y mobile después

Se construye la web responsive y se retrasa la aplicación nativa hasta estabilizar
el producto y el contrato API.

**Ventajas**

- menor trabajo inicial;
- validación temprana desde navegador;
- menos decisiones simultáneas.

**Inconvenientes**

- aplaza el aprendizaje mobile;
- el contrato y la UX pueden sesgarse hacia web;
- puede terminar manteniendo PWA y aplicación nativa además de la web.

**Mantenimiento**

Bajo al principio, con coste diferido cuando se incorpora mobile.

### Recomendación

**Opinión:** alternativa B, aplicaciones especializadas con packages compartidos.

Web y mobile compartirían aquello que es naturalmente portable: contrato,
cliente, validaciones y lógica TypeScript sin APIs de plataforma. La UI se
compartiría solo si una misma interacción funcionara bien en ambos medios.

Esta recomendación habría permitido escoger los frameworks web y mobile de forma
independiente. No fue la alternativa elegida.

### Decisión del usuario

**Aceptada:** alternativa A, cliente universal con React Native. Véase
[ADR-0008](../adr/0008-use-a-universal-react-native-client.md).

El requisito determinante es la paridad funcional: web, iOS y Android son el
mismo producto, incluido su uso en móvil y tablet. La interfaz será responsive y
adaptativa, no necesariamente idéntica píxel a píxel. Se permiten componentes
específicos cuando una plataforma los necesite, sin rebajar accesibilidad,
rendimiento o usabilidad para maximizar reutilización.

## Decisión 4 — Estilo y contrato de API — aceptada

### Problema

El backend Go debe ofrecer un límite estable al cliente universal. Hay que decidir
el estilo de comunicación, la fuente de verdad del contrato y cómo evitar deriva
entre servidor y cliente.

### Alternativas

- **REST + OpenAPI contract-first:** HTTP orientado a recursos, contrato formal y
  generación de cliente.
- **GraphQL schema-first:** consultas flexibles con un modelo operativo
  específico.
- **RPC con IDL:** contratos fuertes y generación, con mayor fricción para
  navegador y recursos públicos.
- **REST sin descripción formal:** menor tooling inicial y mayor riesgo de deriva.

### Recomendación

**Opinión:** REST pragmático con OpenAPI contract-first. Es la solución mínima que
formaliza el límite, funciona naturalmente con los tres targets y permite
generación sin introducir GraphQL o RPC antes de necesitarlos.

### Decisión del usuario

**Aceptada:** REST con OpenAPI contract-first y cliente TypeScript generado. El
backend y la implementación HTTP permanecen en Go. Véase
[ADR-0009](../adr/0009-use-rest-and-openapi-contract-first.md).

TypeScript se adopta para el cliente universal por su comprobación estática. La
versión, configuración del compilador, runtime, package manager y workspaces
siguen pendientes en la decisión de toolchain.

## Decisión 5 — Arquitectura de identidad — aceptada

### Problema

Hay que proporcionar email/password, verificación, recuperación y sesiones,
además de login con Apple y Google, manteniendo un usuario y una autorización de
negocio independientes de los proveedores.

### Alternativas

- identidad exclusivamente local;
- proveedor de identidad gestionado;
- identidad propia federada;
- solo identidades externas.

### Recomendación

**Opinión profesional:** un proveedor gestionado reduciría el riesgo operativo.

**Recomendación condicionada al aprendizaje:** identidad propia federada solo con
threat model, librerías establecidas, pruebas de abuso y revisión de seguridad.

### Decisión del usuario

**Aceptada:** el backend Go gestionará identidad, credenciales locales,
recuperación y sesiones. Apple y Google serán proveedores federados vinculables a
un mismo usuario interno. Véase
[ADR-0010](../adr/0010-own-identity-with-federated-login.md).

El backend extraerá el `subject` de credenciales verificadas y nunca vinculará
cuentas automáticamente solo por coincidencia de email.

## Decisión 6 — Persistencia y acceso a datos — aceptada

### Problema

Hay que elegir el sistema de registro, el acceso desde Go y la evolución del
esquema sin ocultar SQL ni acoplar el dominio a infraestructura.

### Alternativas

- `pgx` con SQL, escaneo y mapeo manual;
- `pgx` con SQL validado y código tipado generado por `sqlc`;
- ORM convencional como GORM;
- framework code-first como Ent.

Las migraciones incrementales se compararon mediante herramientas SQL sencillas,
migración automática de ORM y tooling declarativo más avanzado.

### Recomendación

**Opinión:** PostgreSQL con `pgx` nativo, `sqlc` y migraciones SQL mediante
`goose`. Conserva SQL visible, reduce código mecánico y evita introducir un ORM
completo.

### Decisión del usuario

**Aceptada:** alternativa B. PostgreSQL será el sistema de registro; `pgx`
proporcionará conexión, pool y transacciones; `sqlc` generará acceso Go tipado
desde SQL escrito por el equipo; `goose` aplicará migraciones SQL versionadas.
Véase [ADR-0011](../adr/0011-use-postgresql-pgx-sqlc-and-goose.md).

El código generado y las librerías quedan en el adaptador de persistencia. Las
migraciones no se ejecutarán como efecto secundario del arranque normal de la
API. Esquema, versiones, identificadores, organización y política de despliegue
siguen pendientes.

## Resultado del gate

Cuando todas las decisiones bloqueantes estén aceptadas tendremos:

- un mapa de aplicaciones y repositorios;
- contratos de integración;
- límites de identidad y datos;
- toolchains reproducibles;
- entorno local;
- estrategia de pruebas, seguridad y observabilidad;
- camino de CI, despliegue e infraestructura.

Después se reanudará el diseño de producto y se creará el scaffolding mínimo
compatible con ambos.
