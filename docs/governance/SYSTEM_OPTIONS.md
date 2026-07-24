# Opciones para montar el sistema

> Estado: análisis y recomendaciones iniciales.
>
> Ninguna recomendación de este documento es una decisión aceptada. Cada gate
> importante terminará en ADR tras la elección explícita del usuario.
>
> El orden original de este documento ha sido reemplazado por
> [TECHNICAL_BASELINE.md](TECHNICAL_BASELINE.md). Las secciones de producto quedan
> como material aplazado.

## Silueta recomendada

**Silueta actual:** un monorepo, un backend modular en Go, PostgreSQL, una API HTTP
documentada, identidad basada en estándares y un cliente universal React Native
para web, iOS y Android. Docker Compose primero; observabilidad, Kubernetes y AWS
en las fases del manifiesto.

Esta silueta minimiza piezas desplegables sin cerrar la evolución.

## Orden de decisión original — superado

| Orden | Decisión | Por qué bloquea |
|---|---|---|
| 1 | MVP de torneo | Define el lenguaje, invariantes y vertical slice |
| 2 | Participantes, visibilidad e incorporación | Define autorización y datos |
| 3 | Identidad y sesiones | Afecta web, mobile, API y seguridad |
| 4 | Topología del repositorio y clientes | Define tooling y reutilización |
| 5 | Contrato API | Coordina clientes y backend |
| 6 | Persistencia y migraciones | Materializa el dominio |
| 7 | Entorno local | Permite comenzar implementación reproducible |
| 8 | Observabilidad y despliegue | Se diseña sobre un flujo real |

Este orden queda superado por
[ADR-0004](../adr/0004-technical-baseline-before-product-design.md). Primero se
confirma la base técnica; después se retoman los puntos de producto.

## 1. Alcance del primer torneo

### Alternativa A — Liga todos contra todos

- **Ventajas:** clasificación comprensible; obliga a aprender calendario, puntos y
  desempates.
- **Inconvenientes:** generación de jornadas y reglas de clasificación elevan el
  primer alcance.

### Alternativa B — Eliminatoria directa

- **Ventajas:** flujo y modelo más pequeños; resultado visual claro.
- **Inconvenientes:** byes, siembras y tamaños no potencia de dos siguen
  necesitando reglas.

### Alternativa C — Solo inscripción y publicación

- **Ventajas:** vertical slice mínimo para validar identidad, permisos y unión.
- **Inconvenientes:** todavía no prueba el motor competitivo.

**Recomendación:** empezar por inscripción y publicación, y escoger eliminatoria o
liga como segundo incremento. Así se aprende un eje complejo cada vez.

## 2. Modelo multi-deporte

### Alternativa A — Motor genérico desde el inicio

Configura participantes, puntuación, fases y desempates para cualquier deporte.
Tiene máxima ambición y máximo riesgo de diseñar abstracciones sin casos reales.

### Alternativa B — Fútbol completamente específico

Entrega rápido, pero puede introducir nombres y reglas difíciles de separar.

### Alternativa C — Fútbol primero con puntos de variación explícitos

El torneo conserva una identidad de deporte y conceptos neutrales solo donde ya
son naturales —torneo, participante, inscripción, fase y encuentro—. Las reglas
de fútbol permanecen concretas. La segunda disciplina aporta la evidencia para
extraer interfaces.

**Recomendación:** alternativa C. Diseñar para poder cambiar no significa construir
ahora un motor universal.

## 3. Identidad y autenticación

### Alternativa A — Identidad propia en Go

- **Aprendizaje:** máximo sobre contraseñas, sesiones, tokens y recuperación.
- **Control:** alto.
- **Coste/riesgo:** alto; una superficie crítica queda bajo mantenimiento propio.

### Alternativa B — Proveedor gestionado

Candidatos posteriores: Amazon Cognito, Auth0 u otro proveedor OIDC.

- **Ventajas:** registro, verificación, recuperación y controles maduros.
- **Inconvenientes:** coste, límites de personalización y dependencia.

Amazon Cognito actúa como proveedor OIDC para aplicaciones web y mobile y ofrece
registro y recuperación gestionados. [Documentación oficial de Cognito](https://docs.aws.amazon.com/cognito/latest/developerguide/cognito-user-pools.html).

### Alternativa C — Modelo híbrido

Un proveedor estándar gestiona credenciales y sesiones; TournamentsManager
mantiene un `User` interno y toda autorización de negocio. El backend depende del
contrato OIDC/JWT, no de objetos de dominio del proveedor.

**Recomendación:** alternativa C para producción. Permite aprender autenticación y
autorización sin almacenar contraseñas. Si se elige construir identidad propia
como ejercicio, debe ser una decisión consciente con threat model y revisión
específica.

Para web SPA y aplicaciones nativas, Authorization Code con PKCE es el flujo
recomendado por proveedores OIDC actuales. [Referencia de Auth0 sobre
PKCE](https://auth0.com/docs/api/authentication/authorization-code-flow-with-pkce/authorize-with-pkce).

La recuperación debe evitar enumeración de usuarios, usar tokens aleatorios,
expirables y de un solo uso, y no iniciar sesión automáticamente tras el cambio.
[OWASP Forgot Password Cheat
Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Forgot_Password_Cheat_Sheet.html).

## 4. Web y mobile

### Alternativa A — Aplicación universal Expo

Expo Router permite rutas compartidas para Android, iOS y web. Maximiza
reutilización y reduce proyectos.
[Documentación de Expo Router](https://docs.expo.dev/router/introduction/).

Riesgo: las páginas públicas de torneos, SEO y pantallas densas de administración
pueden necesitar una experiencia web distinta.

### Alternativa B — React web y Expo/React Native mobile separados

Comparten TypeScript, contratos, cliente API, validación, tokens de diseño y lógica
pura. La UI se comparte solo donde funcione de verdad.

React Native soporta módulos y archivos específicos por plataforma, incluyendo
separación entre web y native. [Documentación de código específico por
plataforma](https://reactnative.dev/docs/platform-specific-code.html).

### Alternativa C — Web/PWA primero, mobile después

Reduce el coste inicial y valida el producto, pero retrasa el aprendizaje y los
requisitos nativos.

**Decisión aceptada:** alternativa A. Web, iOS y Android serán un único producto
con paridad funcional y layouts adaptativos para móvil, tablet y escritorio. La
reutilización será el comportamiento por defecto, pero no justificará degradar
accesibilidad, rendimiento o usabilidad. Véase
[ADR-0008](../adr/0008-use-a-universal-react-native-client.md).

React recomienda iniciar aplicaciones nuevas con un framework, y Expo para
aplicaciones nativas. [Guía oficial de React](https://react.dev/learn/creating-a-react-app).
El framework universal y su estrategia de routing y rendering requieren comparar
calidad web, SEO, builds nativos, hosting y complejidad.

## 5. Topología del repositorio

### Alternativa A — Monorepo

Un Git contiene handbook, Go, web, mobile, paquetes compartidos e infraestructura.
Facilita cambios atómicos de contrato y documentación. Añade coordinación de
tooling.

Expo soporta oficialmente monorepos con workspaces, aunque advierte que no son
adecuados para todos los proyectos y añaden complejidad.
[Guía oficial de Expo](https://docs.expo.dev/guides/monorepos/).

### Alternativa B — Repositorios separados

Permite ciclos y permisos independientes, pero exige versionar contratos,
coordinar cambios y mantener pipelines separados desde el principio.

**Decisión aceptada:** monorepo mientras exista un equipo pequeño y un solo
producto. Véase [ADR-0005](../adr/0005-use-a-product-monorepo.md).

No añadir Nx o Turborepo hasta que el tiempo de tareas o la coordinación lo
justifique; los workspaces del package manager pueden ser suficientes.

Silueta candidata, todavía no creada ni aceptada como estructura:

```text
apps/
  api/       # Go
  client/    # Cliente universal web / iOS / Android
packages/
  contracts/ # tipos generados o esquemas compartidos
  api-client/
  design-tokens/
infra/
docs/
```

## 6. Backend

### Alternativa A — Monolito modular

Un proceso desplegable con módulos por capacidad. Transacciones y operación
sencillas; límites internos verificables.

### Alternativa B — Microservicios

Despliegue y escalado independientes, a cambio de red, consistencia distribuida,
observabilidad y operación mucho más complejas.

### Alternativa C — Funciones serverless

Operación gestionada y escalado por evento, pero aumenta la fragmentación y puede
acoplar ejecución, debugging y coste al proveedor.

**Decisión aceptada:** monolito modular en Go. Separar un servicio solo con
evidencia de autonomía, carga, seguridad o ciclo de despliegue. Véase
[ADR-0007](../adr/0007-use-a-modular-monolith-backend.md).

Módulos candidatos:

- identity profile;
- tournament catalog;
- participation;
- competition, cuando se implemente el formato;
- notification, inicialmente como adaptador.

Son límites de exploración, no paquetes aprobados.

## 7. API

### Alternativa A — HTTP REST + OpenAPI

Familiar para web/mobile, fácil de observar y compatible con generación de
clientes. Exige disciplina en recursos, errores y evolución.

### Alternativa B — GraphQL

Flexible para clientes, pero añade esquema, resolvers, autorización por campo,
cache y observabilidad específicas.

### Alternativa C — RPC

Contratos fuertes y eficiencia, con mayor fricción directa para navegador y APIs
públicas.

**Decisión aceptada:** REST pragmático con OpenAPI contract-first. El backend y la
API HTTP se implementan en Go; del contrato se generará el cliente TypeScript de
la aplicación universal. OpenAPI no generará dominio ni reglas de negocio. Véase
[ADR-0009](../adr/0009-use-rest-and-openapi-contract-first.md).

## 8. Participación y autorización

Modelos de entrada posibles:

- código compartido;
- invitación dirigida;
- enlace firmado;
- solicitud con aprobación;
- inscripción abierta.

**Recomendación:** decidir primero entre código e invitación. Toda acción debe
autorizarse por relación con el torneo, no solo por un rol global. “Organizador”
es un permiso contextual del torneo.

También deben decidirse visibilidad `public`, `unlisted` y `private`, porque
condiciona consultas, URLs, cache y privacidad.

## 9. Datos y consistencia

**Recomendación:** PostgreSQL como sistema de registro. No introducir Redis hasta
medir un problema. Las operaciones de crear y unirse deben ser transaccionales e
idempotentes donde pueda existir repetición desde mobile.

Decisiones pendientes:

- identificadores;
- estados del torneo;
- unicidad y capacidad;
- concurrencia de últimas plazas;
- borrado de cuenta y conservación histórica;
- zona horaria;
- migraciones;
- auditoría.

## 10. Mensajería y notificaciones

El registro y la recuperación requieren un canal, probablemente email.

Opciones:

- envío síncrono desde la petición: simple, pero frágil;
- tabla outbox más worker: fiable, con coste moderado;
- broker dedicado: potente, prematuro para el primer corte;
- proveedor de identidad que envíe sus propios mensajes.

**Recomendación:** si la identidad es gestionada, delegar sus correos al proveedor.
Para eventos de torneo, comenzar sin broker; introducir outbox cuando exista una
notificación que deba sobrevivir a fallos.

## Decisiones que no necesitamos aún

- Redis frente a Valkey;
- Kubernetes packaging;
- service mesh;
- event sourcing;
- CQRS;
- bus de eventos;
- motor multi-deporte por plugins;
- microfrontends;
- sincronización offline compleja;
- multi-región.

## Próximo decision gate

La siguiente decisión activa es la arquitectura de identidad, según
[TECHNICAL_BASELINE.md](TECHNICAL_BASELINE.md). Las decisiones de formato,
participantes, visibilidad e incorporación permanecen pausadas.
