# Registro de aprendizaje

## Método

Para cada capacidad se sigue el ciclo:

1. explicar el fundamento con palabras propias;
2. comparar alternativas;
3. construir la versión mínima;
4. observar su comportamiento;
5. provocar y diagnosticar un fallo;
6. documentar lo aprendido y los límites;
7. revisarlo en la retrospectiva de fase.

## Mapa de competencias

| Área           | Resultado demostrable                             | Estado      |
| -------------- | ------------------------------------------------- | ----------- |
| Arquitectura   | Explicar límites, dependencias y trade-offs       | En curso    |
| Go             | Construir y mantener un servicio idiomático       | No iniciado |
| PostgreSQL     | Diseñar, migrar y operar datos con criterio       | No iniciado |
| API            | Diseñar contratos evolutivos y observables        | No iniciado |
| Testing        | Elegir pruebas por riesgo y velocidad de feedback | Fundamentos aceptados |
| Seguridad      | Modelar amenazas y aplicar controles verificables | No iniciado |
| Contenedores   | Crear un entorno local reproducible               | No iniciado |
| Observabilidad | Diagnosticar con logs, métricas y trazas          | No iniciado |
| Kubernetes     | Desplegar, escalar y recuperar cargas             | No iniciado |
| Terraform/AWS  | Aprovisionar y operar infraestructura             | Fundamentos IaC, cuentas y estado aceptados |

## Diario

### 2026-07-25 — El mapa de red no es una factura

- **Aprendido:** una región es la zona geográfica de AWS; una AZ es una ubicación
  aislada dentro de esa región; una VPC es la red privada que las abarca. Las
  subredes son sus porciones de direcciones internas, no máquinas ni IP públicas.
- **Aprendido:** reservar dos AZ permite que ALB y RDS usen una topología válida,
  pero no cobra por sí mismo. El coste empieza al ejecutar ALB, tareas, base de
  datos, IPv4, logs o transferencia.
- **Evidencia:** ADR-0030; España tiene tres AZ y los servicios base elegidos.
- **Coste aceptado:** no se autoriza gasto recurrente sin estimación completa y
  aprobación explícita de importe y duración.
- **Siguiente decisión:** Gate 0B, formato y participantes del primer vertical
  slice de producto.

### 2026-07-25 — Público no significa abierto directamente

- **Aprendido:** publicar una API no exige que cualquiera pueda conectar con
  ella directamente. El ALB es la puerta pública; la API solo acepta las
  conexiones que proceden de esa puerta mediante una regla de security group.
- **Aprendido:** NAT protege la salida de recursos privados, pero no es gratis.
  Para el alcance inicial se prescinde de ella y se mantiene la base de datos
  privada; si el egress privado aporta valor real, la decisión se reabrirá.
- **Evidencia:** ADR-0029; AWS recomienda restringir los targets a aceptar
  tráfico exclusivamente desde el security group del ALB.
- **Coste aceptado:** ALB, Fargate, IPv4 pública, logs y datos costarán cuando
  exista despliegue; el presupuesto se revisará antes de crear recursos.
- **Siguiente decisión:** región, CIDR, subredes, AZ y presupuesto AWS.

### 2026-07-25 — El backend de estado no debe adelantar la red cloud

- **Aprendido:** un backend remoto no solo guarda un archivo: coordina quién
  puede modificar el estado y conserva evidencia para recuperarlo. El estado
  puede ser remoto aunque `plan` y `apply` se ejecuten inicialmente desde la
  CLI local.
- **Aprendido:** HCP Terraform Free resuelve bloqueo e historial sin decidir
  aún región ni crear un bucket AWS; por eso evita que el backend de estado
  fuerce prematuramente la decisión de red. S3 sigue siendo la alternativa de
  salida si cambian los límites, el coste o el control requerido.
- **Evidencia:** ADR-0028; HCP Terraform Free limita el plan a 500 recursos
  gestionados y el backend S3 requiere versionado y locking explícito.
- **Coste aceptado:** depender inicialmente de HCP y proteger un token de
  acceso, aplazando la operación nativa del backend S3.
- **Siguiente decisión:** región y red AWS.

### 2026-07-25 — El estado no es código ni debe vivir en Git

- **Aprendido:** Git registra la intención de infraestructura; el estado de
  Terraform registra la asociación entre esa intención y recursos reales. Un
  valor marcado como sensible puede seguir existiendo en el estado, por lo que
  no se publica ni se versiona.
- **Aprendido:** estado local no cuesta nada, pero no aporta locking remoto ni
  recuperación compartida. Es suficiente sin AWS real; un backend remoto será
  obligatorio antes del primer apply cloud.
- **Evidencia:** ADR-0027; la comparación posterior se cerró en ADR-0028 con
  HCP Terraform Free como backend remoto inicial.
- **Coste aceptado:** retrasar el primer despliegue cloud hasta decidir y
  verificar el backend remoto.
- **Siguiente decisión:** completada en ADR-0028; continúa región y red AWS.

### 2026-07-25 — Una cuenta AWS es una frontera, no un usuario

- **Aprendido:** una cuenta AWS aísla recursos, permisos y facturación. Separar
  `nonprod` de `prod` limita el radio de un error sin requerir una landing zone
  completa desde el inicio.
- **Aprendido:** el acceso humano usa federación, MFA y roles temporales; una
  cuenta root no es una identidad diaria, y GitHub Actions obtiene credenciales
  temporales mediante OIDC en lugar de conservar access keys.
- **Evidencia:** ADR-0026; AWS Organizations organiza las tres cuentas e IAM
  Identity Center centraliza el acceso humano.
- **Coste aceptado:** gestionar correos raíz, permisos, roles y facturación de
  tres cuentas antes de desplegar la primera carga.
- **Siguiente decisión:** estado y bootstrap de Terraform.

### 2026-07-25 — IaC describe el resultado, no una receta manual

- **Aprendido:** Terraform expresa el estado deseado de infraestructura y su
  flujo separa `plan` —previsualizar cambios— de `apply` —ejecutarlos tras
  aprobación—. El estado es necesario para comparar la declaración con los
  recursos reales y requiere un diseño de seguridad independiente.
- **Aprendido:** seleccionar Terraform no crea una cuenta AWS, una VPC ni un
  recurso facturable; cuenta, identidad, backend de estado y red siguen siendo
  decisiones separadas y ordenadas.
- **Evidencia:** ADR-0025; Terraform queda como IaC declarativa para la Fase 5,
  sin SDKs AWS ni tipos de proveedor en la lógica de negocio.
- **Coste aceptado:** aprender HCL y operar correctamente providers, módulos y
  estado remoto cuando la Fase 5 lo requiera.
- **Siguiente decisión:** fundación AWS: cuenta e identidad.

### 2026-07-25 — Promover un artifact no es promover una rama

- **Aprendido:** GitHub conserva código fuente, ECR conserva imágenes OCI y un
  digest identifica exactamente el artifact que ECS puede ejecutar. El digest
  validado en staging debe ser el mismo que llega a producción; no se reconstruye
  durante la promoción.
- **Aprendido:** `dev`, `staging` y `prod` son entornos, no ramas. `develop`
  puede seguir integrando cambios en dev mientras una `release/*` temporal
  conserva en staging el subconjunto que QA y negocio quieren publicar.
- **Evidencia:** ADR-0024; ECR admite tags inmutables y ECS/Fargate puede
  descargar imágenes privadas mediante el rol de ejecución.
- **Coste aceptado:** operar entornos aislados y componer releases selectivas
  añade recursos, conflictos potenciales y sincronización; se activa solo cuando
  haya equipo o una necesidad real, no durante la práctica individual actual.
- **Siguiente decisión:** IaC y AWS.

### 2026-07-25 — Runtime gestionado antes que Kubernetes

- **Aprendido:** ECS organiza servicios de contenedores y Fargate aporta la
  capacidad gestionada para ejecutarlos; no hay que mantener las máquinas que
  hay por debajo.
- **Aprendido:** escoger un destino cloud futuro no obliga a desplegar hoy. La
  práctica actual continúa por completo en local y no debe crear gasto AWS.
- **Evidencia:** ADR-0023; un servicio ECS puede reemplazar tareas que fallan y
  Fargate factura por los recursos consumidos mientras se ejecutan.
- **Coste aceptado:** la red, IAM, Terraform, registro, costes y recuperación
  real se aprenderán en la Fase 5, antes de mantener recursos activos.
- **Siguiente decisión:** registry y promoción de la imagen.

### 2026-07-25 — Empaquetar no es elegir dónde desplegar

- **Aprendido:** una imagen OCI es una caja reproducible para ejecutar la API;
  no es AWS, Kubernetes, un VPS ni un contenedor que deba incluir toda la
  aplicación y sus dependencias.
- **Aprendido:** separar build y runtime evita transportar al entorno final el
  compilador y herramientas de desarrollo; el mismo artefacto podrá identificarse
  de forma inmutable antes de promoverse.
- **Evidencia:** ADR-0022; Docker recomienda las imágenes multi-stage para
  separar compilación y ejecución.
- **Coste aceptado:** habrá que mantener y comprobar la imagen cuando exista una
  API, sin adelantar registry, runtime, firma ni pipeline de despliegue.
- **Siguiente decisión:** runtime y promoción de la API.

### 2026-07-25 — CI como contraste, no como burocracia

- **Aprendido:** la utilidad de CI en un proyecto individual es repetir el
  contrato de calidad en un entorno limpio, no obligar a una persona a revisar
  su propio cambio mediante una PR.
- **Aprendido:** un check obligatorio puede bloquear una promoción cuando el
  recurso externo no está disponible; por eso la puerta vigente es `make verify`
  local y CI aporta una segunda evidencia visible.
- **Evidencia:** ADR-0021; GitHub Actions en runners estándar para repositorio
  público, sin secretos, permisos de escritura ni triggers privilegiados.
- **Coste aceptado:** un resultado rojo no bloquea técnicamente `main` mientras
  no haya colaboración; la disciplina de promoción sigue siendo manual.
- **Siguiente decisión:** contenedores y despliegue.

### 2026-07-25 — Señales correlacionadas, no prints aislados

- **Aprendido:** un log describe un evento discreto; una traza reúne los spans
  de una operación; una métrica agrega medidas en el tiempo. Son señales
  complementarias, no alternativas.
- **Aprendido:** la instrumentación automática debe cubrir límites técnicos;
  los spans manuales se reservan para operaciones significativas que no se vean
  automáticamente, nunca uno por función.
- **Evidencia:** ADR-0020; OpenTelemetry, Prometheus, Grafana, Loki y Tempo,
  con correlación por contexto y sin OpenTelemetry Collector inicialmente.
- **Coste aceptado:** operar cuatro servicios locales y retrasar Collector,
  alertas y SLO hasta que exista una necesidad demostrable.
- **Siguiente decisión:** CI y política de calidad.

### 2026-07-25 — Evidencia proporcional al riesgo

- **Aprendido:** un test rápido no equivale a evidencia suficiente cuando el
  riesgo vive en SQL, restricciones o transacciones; esos comportamientos se
  validan con PostgreSQL real.
- **Aprendido:** una suite end-to-end completa no sustituye las pruebas pequeñas:
  cuesta más diagnosticarla y debe reservarse a recorridos críticos.
- **Evidencia:** ADR-0019, con pruebas unitarias estándar, integración en una
  base efímera, contratos HTTP y E2E mínimos.
- **Coste aceptado:** preparar una base de pruebas aislada y posponer librerías
  adicionales hasta que exista una necesidad repetida.
- **Siguiente decisión:** observabilidad mínima.

### 2026-07-25 — Dependencias contenidas, aplicaciones nativas

- **Aprendido:** la paridad local con producción consiste en conservar contratos
  importantes —configuración externa, PostgreSQL real, salud, volumen y
  migraciones—, no en contenerizar todo desde el primer día.
- **Aprendido:** Expo para web, iOS y Android conserva mejor su bucle local al
  ejecutarse en el host, donde dispone de sus herramientas y simuladores nativos.
- **Evidencia:** ADR-0018; PostgreSQL queda delimitado como primera dependencia
  de Docker Compose.
- **Coste aceptado:** la API futura debe diagnosticar PostgreSQL no disponible y
  la imagen final de API se validará en una decisión posterior de despliegue.
- **Siguiente decisión:** estrategia de pruebas.

### 2026-07-24 — Configuración pública, secretos y entornos

- **Aprendido:** no toda configuración es secreta, pero toda configuración debe
  tener propietario, contrato y destino.
- **Aprendido:** cualquier valor `EXPO_PUBLIC_*` acaba accesible para quien use
  el cliente, así que nunca es secreto.
- **Aprendido:** OIDC permite que CI obtenga credenciales cloud temporales sin
  guardar access keys largas en GitHub.
- **Evidencia:** ADR-0017 y reglas actualizadas en seguridad, desarrollo y
  despliegue.
- **Coste aceptado:** los `.env` locales se crean fuera de Git y la validación de
  configuración debe implementarse en cada runtime.
- **Siguiente decisión:** completada en ADR-0018.

### 2026-07-24 — Rendering simple para producto privado

- **Aprendido:** si el producto inicial vive en torneos privados, SEO, static
  rendering y SSR no son necesidades de base sino posibles evoluciones.
- **Aprendido:** compartir cliente no significa mezclar diferencias de plataforma
  dentro de la lógica; se aíslan en componentes o archivos específicos.
- **Evidencia:** ADR-0016 y documentación ajustada a privacidad inicial y acceso
  invitado limitado.
- **Coste aceptado:** la web inicial no optimiza indexación ni previews sociales
  por torneo.
- **Siguiente decisión:** completada en ADR-0017.

### 2026-07-24 — Rutas universales y nativo generado

- **Aprendido:** Expo Router asigna una URL a cada pantalla; esa misma ruta puede
  abrirse en web o mediante deep link nativo.
- **Aprendido:** CNG no elimina el código nativo de una app: desplaza su fuente
  de verdad a configuración y plugins reproducibles.
- **Evidencia:** ADR-0015, rutas y directorios nativos futuros documentados e
  ignorados por Git.
- **Coste aceptado:** usar convenciones Expo y no tocar directorios generados.
- **Siguiente decisión:** completada en ADR-0016.

### 2026-07-24 — Toolchain TypeScript reproducible

- **Aprendido:** `Current` recibe primero las novedades de Node; LTS prioriza una
  ventana estable y prolongada, adecuada para desarrollo y CI.
- **Aprendido:** una versión `latest` no es automáticamente la mejor elección;
  TypeScript debe permanecer dentro del rango compatible del linter.
- **Aprendido:** TypeScript 6 rechaza una lista `files` vacía; el check raíz
  valida configuración real mientras todavía no existen workspaces TypeScript.
- **Evidencia:** ADR-0014, runtime y package manager pineados, workspaces, lockfile
  y checks compartidos.
- **Coste aceptado:** pnpm, ESLint y Prettier añaden piezas separadas a cambio de
  dependencias explícitas y responsabilidades claras.
- **Siguiente decisión:** framework del cliente universal.

### 2026-07-24 — Flujo de integración proporcional al equipo

- **Aprendido:** una rama `develop` conserva `main` como hito estable, pero puede
  acumular cambios no publicables y bloquear promociones completas.
- **Evidencia:** ADR-0013 y guía de contribución con reglas para sincronización,
  promoción y excepciones.
- **Coste aceptado:** mantener dos ramas de larga vida mientras el trabajo sea
  principalmente individual.
- **Siguiente decisión:** toolchain TypeScript.

### 2026-07-24 — Identidad canónica del repositorio

- **Aprendido:** la ruta de un módulo Go público forma parte de su identidad y
  debe coincidir con el propietario y nombre reales del remoto.
- **Evidencia:** cuenta de GitHub autenticada y nombre
  `joseantoniogarciay/TournamentsManager` comprobado antes del primer push.
- **Coste evitado:** no publicar una ruta provisional que después obligue a
  cambiar imports y consumidores.
- **Siguiente decisión:** toolchain TypeScript.

### 2026-07-24 — Toolchain Go reproducible

- **Aprendido:** `go tool` ejecuta herramientas declaradas; `tool` las registra y
  `-modfile` selecciona un grafo alternativo.
- **Aprendido:** `go.tool.mod` no es un nombre mágico; Make y el wrapper de VS
  Code encapsulan su selección explícita.
- **Evidencia:** ADR-0012, módulos separados, Makefile y formato al guardar con
  `goimports` pineado.
- **Coste aceptado:** mantener dos pares `go.mod`/`go.sum` y ejecutar
  `tidy-check` para ambos.
- **Siguiente decisión:** toolchain TypeScript.

### 2026-07-24 — Persistencia SQL-first tipada

- **Aprendido:** `pgx` es el driver PostgreSQL, `sqlc` genera código Go desde SQL
  y `goose` versiona la evolución del esquema; resuelven problemas diferentes.
- **Aprendido:** el código generado evita trabajo mecánico y detecta deriva, pero
  no diseña consultas, transacciones, índices ni modelos de dominio.
- **Evidencia:** ADR-0011 y reglas documentadas para separar filas, adaptadores y
  dominio.
- **Coste aceptado:** mantener SQL, generación determinista, mapeos explícitos y
  una política operativa de migraciones.
- **Siguiente decisión:** toolchain Go.

### 2026-07-24 — Identidad propia federada

- **Aprendido:** el `subject` identifica una cuenta dentro del proveedor; el
  backend debe extraerlo de un token verificado y mapearlo a un usuario interno.
- **Aprendido:** una URL navegable no debe convertir un `GET` en una operación
  sensible; la apertura presenta el estado y una confirmación explícita mediante
  `POST` consume el intento antes de volver a la home.
- **Evidencia:** ADR-0010 y flujos documentados para cambio de email y vinculación
  con prueba fresca.
- **Coste aceptado:** operar credenciales, sesiones, recuperación, OAuth/OIDC y
  controles de abuso.
- **Siguiente decisión:** completada en ADR-0011.

### 2026-07-24 — Navegación del handbook

- **Aprendido:** mantener todos los documentos en la raíz aumenta visibilidad al
  principio, pero deja de escalar cuando oculta las unidades del monorepo.
- **Evidencia:** handbook agrupado por proyecto, gobierno, ingeniería y
  operaciones, con índices y enlaces validados.
- **Coste aceptado:** mantener rutas e índices como parte de cualquier movimiento
  documental.
- **Siguiente decisión:** arquitectura de identidad.

### 2026-07-24 — Contrato API y cliente generado

- **Aprendido:** la API HTTP es un adaptador del backend, no todo el backend;
  OpenAPI coordina el servidor Go y el consumidor TypeScript.
- **Evidencia:** ADR-0009 con REST contract-first, generación del cliente y
  límites respecto al dominio.
- **Coste aceptado:** mantener lint, generación y compatibilidad, especialmente
  con aplicaciones instaladas que se actualizan más tarde.
- **Siguiente decisión:** arquitectura de identidad.

### 2026-07-24 — Estrategia universal de cliente

- **Aprendido:** compartir producto y comportamiento no obliga a usar una
  presentación idéntica; responsive y adaptativo resuelven tamaños y capacidades
  de entrada diferentes.
- **Evidencia:** ADR-0008 con paridad funcional, límites de plataforma y
  disparadores de revisión.
- **Coste aceptado:** aislar excepciones web/native y validar por separado SEO,
  accesibilidad, rendimiento y releases.
- **Siguiente decisión:** estilo y contrato de API.

### 2026-07-24 — Topología técnica inicial

- **Aprendido:** monorepo no significa un único pipeline, y monolito no significa
  ausencia de límites.
- **Evidencia:** ADR de monorepo, publicación segura y monolito modular.
- **Coste aceptado:** disciplina para mantener módulos y pipelines independientes.
- **Siguiente decisión:** estrategia web/mobile y frontera de reutilización.

### 2026-07-23 — Fundación documental

- **Aprendido:** una arquitectura profesional comienza por explicitar autoridad,
  proceso, estados y criterios de salida.
- **Evidencia:** manifiesto transcrito, ADR iniciales, mapa del handbook y
  plantillas operativas.
- **Incertidumbre:** requisitos del producto y decisiones de implementación.
- **Siguiente experimento:** definir el alcance y primer caso de uso antes del
  entorno o del backend.

## Regla de evidencia

“Entendido” exige una explicación propia y una demostración. Un comando que
funciona o una respuesta del asistente no son evidencia suficiente por sí solos.
