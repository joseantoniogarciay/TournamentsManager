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

| Área           | Resultado demostrable                             | Estado                                      |
| -------------- | ------------------------------------------------- | ------------------------------------------- |
| Arquitectura   | Explicar límites, dependencias y trade-offs       | En curso                                    |
| Go             | Construir y mantener un servicio idiomático       | No iniciado                                 |
| PostgreSQL     | Diseñar, migrar y operar datos con criterio       | No iniciado                                 |
| API            | Diseñar contratos evolutivos y observables        | No iniciado                                 |
| Testing        | Elegir pruebas por riesgo y velocidad de feedback | Fundamentos aceptados                       |
| Seguridad      | Modelar amenazas y aplicar controles verificables | No iniciado                                 |
| Contenedores   | Crear un entorno local reproducible               | No iniciado                                 |
| Observabilidad | Diagnosticar con logs, métricas y trazas          | No iniciado                                 |
| Kubernetes     | Desplegar, escalar y recuperar cargas             | No iniciado                                 |
| Terraform/AWS  | Aprovisionar y operar infraestructura             | Fundamentos IaC, cuentas y estado aceptados |

## Diario

### 2026-07-30 — Un recorrido vertical hace visibles los límites

- **Aprendido:** una ruta HTTP no es toda la API ni toda la lógica del backend.
  El handler traduce transporte; el servicio coordina el caso de uso; el
  repositorio adapta el puerto a PostgreSQL; `sqlc` genera la llamada Go a partir
  del SQL escrito por el equipo.
- **Evidencia:** el recorrido de `GET /v1/usernames/{username}/availability` en
  `apps/backend/README.md` enlaza `server.go`, `registration.Service`,
  `RegistrationRepository`, `db/queries/registrations.sql` y la salida generada
  por `sqlc`.
- **Coste aceptado:** se mantiene una guía local al backend, enlazada desde el
  README raíz y Arquitectura, en vez de repetir el detalle en todos los
  documentos normativos.

### 2026-07-29 — Una comprobación de disponibilidad mejora la UX, no la unicidad

- **Aprendido:** una respuesta de disponibilidad solo describe un instante: dos
  clientes pueden recibir «disponible» y competir después por el mismo nombre.
  Por ello PostgreSQL mantiene la restricción única y el alta es la única
  autoridad que decide definitivamente.
- **Evidencia:** `GET /v1/usernames/{username}/availability` solo acepta el
  formato del contrato, responde `Cache-Control: no-store` y el cliente espera
  400 ms desde la última escritura antes de llamarlo; una nueva entrada cancela
  la petición previa.
- **Coste aceptado:** el límite de 30 consultas por IP y minuto vive en memoria
  del proceso para mantener este corte simple. No protege de forma global entre
  réplicas; si se escala horizontalmente, se decidirá un control compartido en
  el borde o almacenamiento adecuado.

### 2026-07-29 — CORS es una política de frontera, no una cabecera por ruta

- **Aprendido:** CORS autoriza al JavaScript de un origen a leer una respuesta;
  no sustituye autenticación, autorización ni CSRF. Un `POST` JSON inicia además
  un preflight `OPTIONS` antes de la petición real.
- **Evidencia:** la API valida `CORS_ALLOWED_ORIGINS` al arrancar y aplica una
  allowlist exacta en el handler raíz. Devuelve el origen permitido, `Vary: Origin`
  y credenciales; responde el preflight solo para métodos y cabeceras
  que el cliente necesita.
- **Coste aceptado:** cada entorno debe declarar sus orígenes web. Se evita el
  arranque cómodo con `*` porque sería incompatible con las cookies de sesión y
  ocultaría qué frontends tienen permiso para consumir la API.

### 2026-07-29 — La experiencia de usuario y el diagnóstico técnico requieren señales distintas

- **Aprendido:** un replay explica la secuencia de interacción que precede a un
  error; una traza OpenTelemetry explica qué ocurrió en HTTP, lógica técnica y
  PostgreSQL. Ninguna de las dos señales sustituye a la otra, y se cruzan con un
  `request_id` opaco en vez de exponer identidad, payloads o PII.
- **Evidencia:** ADR-0060 aplaza PostHog Cloud hasta la primera beta distribuida,
  limita su alcance a experiencia de cliente y conserva el stack OpenTelemetry
  de ADR-0020 para backend. El proveedor no recibe autoridad de negocio.
- **Coste aceptado:** replay, analytics y error tracking quedan en un SaaS
  externo con región UE, gasto máximo de 0 € y revisión de privacidad previa;
  no se autoaloja una plataforma de producto sin volumen que lo justifique.

### 2026-07-29 — El espacio de una tab superpuesta es parte del scroll

- **Aprendido:** reservar el área de la barra de tabs como padding de la pantalla
  acorta visualmente toda la vista. La superficie debe alcanzar el borde inferior
  y el margen de seguridad debe pertenecer al contenedor desplazable.
- **Evidencia:** `Screen` admite omitir su inset inferior en rutas bajo
  `NativeTabs`; Inicio, Torneos y las rutas de Cuenta lo hacen. Los `ScrollView`
  de esos flujos mantienen su `paddingBottom` de `space[12]`, de modo que el
  último control se puede llevar por encima de la barra nativa.
- **Coste aceptado:** no se introduce una segunda abstracción de layout ni se
  fija una altura de tab bar, que variaría según plataforma y versión del sistema.

### 2026-07-29 — El feedback global también pertenece al área segura

- **Aprendido:** un aviso superpuesto debe calcular su posición desde el inset
  seguro, no desde un margen fijo; de otro modo, en un iPhone con notch puede
  quedar oculto detrás del sistema.
- **Evidencia:** `FeedbackProvider` posiciona el banner tras el inset superior,
  adopta el radio y margen lateral de una card y anima entrada y salida con el
  token compartido `motion.enterExit`. También identifica cada aviso activo,
  cancela su temporizador y sustituye de forma segura el anterior cuando llega
  un mensaje nuevo. El toque y un arrastre vertical hacia arriba descartan el
  aviso; los arrastres cortos se recuperan con `motion.feedback`.
- **Coste aceptado:** la transición se omite cuando el sistema solicita
  movimiento reducido; no se añade una librería ni una jerarquía nueva para un
  único aviso global.

### 2026-07-28 — El degradado conserva jerarquía cuando no colorea todo

- **Aprendido:** aplicar el degradado de marca a la acción primary filled crea
  una relación directa con el icono; limitarlo al borde de 1 px en la acción
  secondary conserva contraste y evita competir con la acción principal.
- **Evidencia:** `Button` parte de azul sólido y superpone el token
  `gradient.brandButton`, variante horizontal de `gradient.brand`: mantiene la
  dirección visual del icono cuadrado al concentrar el recorrido en el extremo
  derecho. El texto sobre esa marca usa un token blanco independiente del tema, no
  el color inverso de la superficie de la app.
- **Coste aceptado:** las acciones destructivas siguen usando rojo sólido, pues
  el degradado de marca no debe diluir su semántica de riesgo.

### 2026-07-29 — El degradado se ajusta desde el token, no desde cada botón

- **Aprendido:** adelantar la entrada del violeta al 35 % del recorrido aumenta
  su presencia sin cambiar el ángulo ni convertir toda la acción en una superficie
  violeta.
- **Evidencia:** `gradient.brand` y `gradient.brandButton` comparten las mismas
  paradas; los botones conservan una geometría horizontal específica.
- **Coste aceptado:** la barra de tabs nativa mantiene un tint azul sólido, porque
  su API no admite un degradado uniforme para icono y etiqueta.

### 2026-07-29 — Semántica de indicadores

- **Aprendido:** el color principal comunica una acción disponible. Un número de
  paso estático debe usar el color de texto y un borde, no un fondo de marca, para
  no adquirir una apariencia pulsable.
- **Evidencia:** los pasos orientativos de la home consumen `colors.text.primary`
  tanto para el número como para el borde circular, y se adaptan al tema claro u
  oscuro sin usar blanco o negro puros.

### 2026-07-29 — Un botón es una acción, no una caja

- **Aprendido:** un radio de 12 px encaja en campos, pero un botón de 44 px de
  alto necesita extremos semicirculares para expresar mejor que es una acción.
- **Evidencia:** la primitiva `Button` usa el token `radius.pill` en todas sus
  variantes; el botón secondary mantiene un borde azul de marca y deja
  transparente su interior para revelar la superficie que hay detrás.
- **Coste aceptado:** el contraste del borde y del texto se evalúa contra la
  superficie contenedora; por ello las pantallas siguen usando tokens de fondo
  semánticos y no colores arbitrarios.

### 2026-07-29 — Las acciones de proveedor caben en su identidad visual

- **Aprendido:** cuando una acción de proveedor se entiende por su marca, un
  botón solo con el icono reduce ruido visual sin perder accesibilidad.
- **Evidencia:** la acción de Google en Cuenta es circular, está centrada en su
  card, conserva un objetivo de 48 px y expone su estado y etiqueta localizada
  al lector de pantalla.
- **Coste aceptado:** el PNG oficial se versiona como asset local para no añadir
  una dependencia de iconos ni depender de red durante el renderizado.

### 2026-07-29 — La navegación persistente no se duplica como contenido

- **Aprendido:** una card que solo recuerda que existe una sección permanente no
  mejora la orientación. La home conserva contenido de bienvenida y la tab
  Torneos mantiene la responsabilidad de acceso a la biblioteca.
- **Evidencia:** se retiró el bloque informativo final de la home; Inicio mantiene
  la acción principal, el contexto y los pasos orientativos.

### 2026-07-28 — Dos apps no obligan a mantener dos proyectos Xcode

- **Aprendido:** Expo CNG permite resolver nombre, esquema, icono y futuras
  diferencias nativas desde una configuración por entorno. Así desarrollo y
  producción se pueden instalar como apps distintas sin convertir directorios
  generados en fuente versionada.
- **Evidencia:** `apps/client/app.config.ts` selecciona `Fast Tourney Dev` o
  `Fast Tourney` con `APP_ENV` y usa el icono común de 1024 × 1024.
- **Coste aceptado:** los bundle identifiers y perfiles de distribución esperan
  a disponer de un dominio controlado. Se evaluarán targets iOS separados solo
  si un config plugin no puede aislar una dependencia nativa necesaria.

### 2026-07-28 — Una preferencia local no es un permiso del sistema

- **Aprendido:** el tema es una preferencia de presentación y puede persistirse
  localmente para cualquier visitante; el permiso de notificaciones depende del
  sistema operativo, requiere una integración nativa y no se puede representar
  como un switch local que aparente concederlo.
- **Evidencia:** `PreferencesProvider` resuelve y persiste `system`, `light` y
  `dark` para todas las rutas, y la raíz entrega el tema correspondiente a React
  Navigation para que las transiciones nativas no usen temporalmente el tema
  contrario. El ajuste de notificaciones permanece deshabilitado y explica su
  límite, coherente con el alcance de `PRODUCT.md`.
- **Coste aceptado:** no se incorpora aún `expo-notifications`, configuración
  nativa ni almacenamiento de preferencias de entrega. Requerirán una decisión
  de producto y el diseño del ciclo de permiso antes de activarlos.

### 2026-07-28 — Un formulario visible no equivale a una sesión inventada

- **Aprendido:** una interfaz de acceso puede preparar el recorrido local y la
  ruta de alta sin afirmar que el cliente ya establece una sesión cuando el
  handler y el adaptador HTTP todavía no existen.
- **Evidencia:** Cuenta ofrece correo, contraseña y registro en una ruta
  profunda; las acciones comunican su indisponibilidad actual y Google se marca
  como próximo, conforme al contrato y ADR-0050.
- **Coste aceptado:** conectar el formulario exige implementar el adaptador de
  sesión y el flujo OIDC real; los controles actuales no envían credenciales.

### 2026-07-28 — El tipado no sustituye un bundle de Expo

- **Aprendido:** TypeScript no carga Metro ni resuelve el árbol de rutas en la
  misma forma que Expo. La exportación web detecta errores de bundling y router
  sin exigir un navegador ni un artefacto persistente.
- **Evidencia:** `make client-web-export` escribe temporalmente en `/tmp` y
  `make verify` lo ejecuta junto al resto de la puerta de calidad.
- **Coste aceptado:** no se añade todavía una prueba visual automatizada ni un
  artefacto de CI; se abrirán cuando exista un flujo crítico que los justifique.

### 2026-07-28 — Una tab es un límite de navegación, no un botón de pantalla

- **Aprendido:** Inicio, Torneos y Cuenta representan flujos de primer nivel;
  ubicarlos en la botonera evita duplicar accesos dentro de cada pantalla. Cuenta
  conserva un stack para que su evolución no altere el historial de Inicio.
- **Evidencia:** grupo `(tabs)` de Expo Router, con Inicio en la primera
  posición, y `NativeTabs` delegando el acabado Liquid Glass a iOS 26.
- **Coste aceptado:** Torneos y Cuenta solo muestran su contexto hasta que sus
  flujos autenticados y colecciones tengan datos reales.

### 2026-07-28 — La jerarquía también necesita límites espaciales

- **Aprendido:** el padding de pantalla protege el contenido de bordes y zonas
  del sistema; las cards delimitan bloques con significado. Usar ambos evita
  una pantalla plana sin convertir cada elemento en una tarjeta.
- **Evidencia:** `Screen` reserva 24 px de cabecera y `shared/ui/Card` añade
  20 px de margen exterior horizontal al agrupar los cuatro bloques de la home.
- **Coste aceptado:** no se añaden sombras ni variantes antes de que una lista
  o interacción demuestre necesitar un estado visual adicional.

### 2026-07-28 — Un token visual no localiza el producto

- **Aprendido:** los tokens hacen consistente la presentación, pero no eliminan
  la necesidad de catalogar el copy. Dejar textos en una ruta impide aplicar el
  locale detectado y hace que cada nueva traducción sea una búsqueda manual.
- **Evidencia:** extracción del copy de la home a un catálogo `es`, `en`, `it`
  y `fr`, con fallback explícito a inglés conforme a ADR-0056.
- **Coste aceptado:** el selector web persistente se implementará con el
  provider de preferencias compartido, antes de añadir una segunda pantalla.
- **Siguiente decisión:** incorporar ese provider de idioma y tema a la raíz
  del cliente sin duplicar estado por feature.

### 2026-07-28 — Un locale es una unidad de intercambio

- **Aprendido:** un archivo plano por locale permite que una plataforma de
  traducción importe y exporte el producto completo sin reconstruir textos
  dispersos. Los prefijos semánticos reúnen las claves comunes y de cada
  capacidad sin imponer objetos anidados al formato de intercambio.
- **Evidencia:** `shared/i18n/locales/{es,en,it,fr}.json` y la validación
  TypeScript de que cada catálogo tiene las claves del inglés.
- **Coste aceptado:** la preferencia persistente de web sigue pendiente del
  provider compartido; los JSON no implementan estado por sí mismos.

### 2026-07-28 — Una home inicial orienta sin inventar estado

- **Aprendido:** una home de invitado no necesita simular una sesión ni una
  biblioteca vacía: la siguiente acción principal y una explicación breve
  orientan sin presentar datos personalizados inexistentes.
- **Evidencia:** implementación de `/` sobre las primitivas Pulse y la regla de
  ADR-0057 que reserva «Administro» y «Guardados» para una sesión verificada.
- **Coste aceptado:** los destinos de crear torneo y cuenta continúan en cortes
  propios; sus acciones muestran una respuesta explícita en vez de aparentar
  que el flujo ya está disponible.
- **Siguiente decisión:** construir el flujo de borrador local y la pantalla de
  cuenta sobre las rutas canónicas acordadas.

### 2026-07-28 — Autenticar no es autorizar

- **Aprendido:** validar una cookie o Bearer identifica la cuenta y puede ser
  transversal; decidir qué puede hacer esa cuenta dentro de una liga requiere
  las reglas del caso de uso. Mezclar ambas cosas en middleware ocultaría la
  política de negocio.
- **Evidencia:** ADR-0059 y la ruta protegida `GET /me/leagues`.
- **Coste aceptado:** CSRF sigue como protección independiente para la primera
  mutación web autenticada; no se añade un token sin una operación que proteger.
- **Siguiente decisión:** diseñar esa protección CSRF junto con la primera
  mutación autenticada por cookie.

### 2026-07-28 — La home reutiliza una colección; no necesita una API propia

- **Aprendido:** la home puede pedir las primeras páginas de la misma colección
  autenticada que usa la biblioteca. Separar `administered` y `followed` en el
  servidor conserva la autorización y evita que el cliente cargue o clasifique
  relaciones ajenas.
- **Evidencia:** ADR-0058; tablas de administradores y seguidores con claves
  compuestas y `GET /me/leagues` paginado por UUIDv7.
- **Coste aceptado:** la home puede realizar dos lecturas pequeñas; no se añade
  agregación, caché global ni búsqueda antes de medir la necesidad.
- **Siguiente decisión:** implementar las mutaciones para seguir una liga y
  asignar o abandonar administración delegada antes de esperar datos de usuario.

### 2026-07-28 — Una colección personal no es una regla de autorización

- **Aprendido:** «Administro» y «Sigo» son vistas útiles de relaciones distintas
  con una liga. La primera reúne propiedad y administración delegada; la segunda
  expresa seguimiento. Ninguna clasificación en el cliente concede permisos.
- **Evidencia:** ADR-0034 y ADR-0057.
- **Coste aceptado:** hace falta una proyección autenticada adicional antes de
  mostrar colecciones reales; no se añade una caché global ni persistencia de
  navegación entre reinicios.
- **Siguiente decisión:** definir el recurso OpenAPI autenticado para las ligas
  relacionadas con la cuenta, con orden y paginación.

### 2026-07-27 — Una mitigación de tooling necesita fecha de caducidad

- **Aprendido:** con Go 1.26.5 y `golangci-lint` 2.12.2, incluir paquetes de
  prueba en el análisis puede terminar con `no go files to analyze`, aunque
  `go test ./...` los cargue y ejecute correctamente. No se ha confirmado un
  issue upstream que documente exactamente esta combinación.
- **Evidencia:** reproducción local y primera CI de la promoción de Fase 1;
  `make verify` pasa con `run.tests: false` y conserva `go test ./...` como
  comprobación de los tests.
- **Coste aceptado:** los linters no inspeccionan temporalmente `*_test.go`;
  las pruebas siguen compilándose y ejecutándose en local y CI.
- **Siguiente decisión:** al actualizar Go o `golangci-lint`, reactivar
  `run.tests: true`, ejecutar `make verify` y retirar la mitigación si pasa.

### 2026-07-27 — Una identidad externa no es una sesión

- **Aprendido:** Google acredita una identidad externa mediante un `subject`
  estable; TournamentsManager la vincula a una cuenta interna y emite después su
  propia sesión opaca. La contraseña no interviene en ese recorrido.
- **Evidencia:** ADR-0050 y la referencia de OpenID Connect de Google.
- **Coste aceptado:** el primer incremento debe configurar y validar Google,
  incluida la vinculación explícita de una cuenta local coincidente.
- **Siguiente decisión:** concretar el artefacto OIDC, nonce, CSRF y contrato
  HTTP antes de implementar el adaptador Google.

### 2026-07-26 — Renovar una sesión no exige JWT

- **Aprendido:** una credencial opaca puede rotarse silenciosamente y conservar
  revocación inmediata porque el servidor mantiene su estado; JWT aporta
  validación distribuida, no más comodidad intrínseca para la persona usuaria.
- **Evidencia:** ADR-0044 y comparación con sesiones de PostgreSQL.
- **Coste aceptado:** cada petición autenticada valida una sesión en PostgreSQL;
  no se adelantan servicios ni tokens de acceso distribuidos.
- **Siguiente decisión:** modelo de datos y contrato HTTP del registro,
  verificación, sesión y publicación.

### 2026-07-26 — Un vertical slice no necesita contener todo el dominio

- **Aprendido:** el primer recorrido útil debe cruzar identidad, autorización,
  datos y API, pero puede detenerse antes de las reglas deportivas avanzadas si
  estas no cambian la evidencia buscada.
- **Evidencia:** ADR-0043.
- **Coste aceptado:** inicio, resultados, bajas, cierre, cancelación y
  administración delegada esperan a un incremento posterior.
- **Siguiente decisión:** analizar sesiones, verificación local y modelo de
  datos antes de implementar el backend.

### 2026-07-26 — La persistencia es una propiedad del volumen, no del contenedor

- **Aprendido:** detener y recrear el contenedor no borra una base PostgreSQL si
  el volumen nombrado se conserva; borrar esos datos debe seguir siendo una
  operación explícita y confirmada.
- **Evidencia:** arranque saludable, reinicio y lectura posterior de una marca
  temporal mediante el runbook de PostgreSQL local.
- **Coste aceptado:** el reset solo es seguro cuando los datos locales son
  prescindibles; la confirmación explícita evita convertirlo en un efecto
  accidental de otro comando.
- **Siguiente decisión:** crear el esquema y las primeras migraciones al abrir
  el trabajo de backend de la Fase 2.

### 2026-07-26 — Un estado visible puede preceder a una notificación

- **Aprendido:** conservar una liga cancelada y mostrar su estado por enlace o
  en «ligas seguidas» comunica la verdad del recurso sin introducir aún entrega,
  preferencias y fallos de email o push.
- **Evidencia:** ADR-0042.
- **Coste aceptado:** las personas no reciben aviso proactivo ni explicación de
  la cancelación.
- **Siguiente decisión:** Gate 0B completado; preparar el entorno local antes de
  implementar el vertical slice.

### 2026-07-26 — La igualdad puede requerir reescribir un marcador

- **Aprendido:** conservar algunos resultados y otorgar otros tras una baja crea
  una desigualdad por el orden de calendario. Uniformar todos a `3-0` mantiene la
  liga, pero exige historial porque puede reemplazar datos reales.
- **Evidencia:** ADR-0041.
- **Coste aceptado:** todavía no hay motivos tipados de baja ni reincorporación.
- **Siguiente decisión:** completada en ADR-0042; Gate 0B cerrado.

### 2026-07-26 — Compartir no equivale a congelar

- **Aprendido:** publicar permite compartir y terminar de preparar una liga; el
  límite seguro para generar partidos y congelar equipos es el inicio. Así se
  evita replanificar una competición con resultados.
- **Evidencia:** ADR-0040, sucesor parcial de ADR-0032.
- **Coste aceptado:** una liga publicada aún no tiene partidos y puede quedarse
  sin iniciar hasta que exista una política de recordatorios o limpieza.
- **Siguiente decisión:** completada en ADR-0041; continúa motivo y comunicación
  de una cancelación.

### 2026-07-26 — Finalizar es una revisión, no un efecto colateral

- **Aprendido:** exigir todos los marcadores y una acción explícita del creador
  evita clasificaciones incompletas y deja una oportunidad final para detectar un
  error. Un `3-0` administrativo se representa como marcador ordinario para no
  crear incidencias que el dominio todavía no sabe interpretar.
- **Evidencia:** ADR-0039.
- **Coste aceptado:** se pierde el motivo de un resultado excepcional y no se
  puede corregir una liga finalizada en el primer corte.
- **Siguiente decisión:** cancelación completada en ADR-0040; continúa la baja
  de equipos.

### 2026-07-26 — Publicar prepara; iniciar habilita la competición

- **Aprendido:** separar publicación e inicio evita que un resultado transforme
  accidentalmente una liga que solo estaba lista para compartir. La transición
  explícita conserva al creador el control de cuándo empieza el juego.
- **Evidencia:** ADR-0038.
- **Coste aceptado:** no hay inicio automático, calendario ni cambios de reglas
  o equipos después de publicar.
- **Siguiente decisión:** completada en ADR-0039; continúan cancelación y bajas
  de equipos.

### 2026-07-26 — Un marcador útil no necesita modelar todas las incidencias

- **Aprendido:** dos goles enteros no negativos bastan para expresar victoria,
  derrota y empate en una liga con puntuación 3-1-0. Añadir excepciones sin sus
  reglas de clasificación produciría campos que el sistema no sabe interpretar.
- **Evidencia:** ADR-0037.
- **Coste aceptado:** no pueden registrarse aún penaltis, prórrogas,
  incomparecencias ni sanciones.
- **Siguiente decisión:** completada en ADR-0038; continúan finalización y bajas
  de equipos.

### 2026-07-26 — Corregir rápido no implica borrar el pasado

- **Aprendido:** permitir la corrección directa conserva el flujo operativo, pero
  el valor actual no explica por sí solo cómo llegó a existir. El historial de
  actor, instante y valores anterior/nuevo aporta trazabilidad sin una cola de
  aprobaciones.
- **Evidencia:** ADR-0036.
- **Coste aceptado:** todavía no existen restauración, disputa ni revisión por
  terceros.
- **Siguiente decisión:** completada en ADR-0037; continúa el momento válido de
  registro de resultados.

### 2026-07-26 — Delegar no exige una segunda cola de aprobación

- **Aprendido:** si el creador selecciona y puede revocar a quien administra
  resultados, una confirmación adicional suele añadir espera y estados pendientes
  sin aportar valor proporcional en una liga cerrada.
- **Evidencia:** ADR-0035.
- **Coste aceptado:** un marcador equivocado será visible hasta corregirse; la
  corrección y la auditoría se deciden después.
- **Siguiente decisión:** completada en ADR-0036; continúan datos y reglas de
  resultados.

### 2026-07-26 — Competir, seguir y administrar son relaciones distintas

- **Aprendido:** en una liga por equipos, el participante deportivo no tiene por
  qué ser una persona con cuenta. Separar equipos, seguidores y administradores
  evita convertir permisos de aplicación en reglas de competición.
- **Aprendido:** un `username` público permite buscar personas sin exponer su
  correo; mantenerlo inmutable al inicio reduce el modelo mientras se valida su
  uso.
- **Evidencia:** ADR-0034.
- **Coste aceptado:** la administración se concede sin aceptación y todavía no
  dispone de bloqueo, avisos ni límites frente a abuso.
- **Siguiente decisión:** completada en ADR-0035; continúa la corrección de
  resultados.

### 2026-07-27 — Un ID público identifica; la autorización protege mutaciones

- **Aprendido:** una URL puede construirse directamente con el ID de una liga
  visible cuando la lectura se permite a quien lo conozca. Ese ID no es un
  secreto ni concede permisos de edición, resultados o administración.
- **Aprendido:** separar identificador público y autorización evita introducir
  tokens de compartición cuando el producto no necesita restringir la lectura.
- **Evidencia:** ADR-0049, sucesor de ADR-0033.
- **Coste aceptado:** la proyección pública puede ser leída por quien conozca o
  adivine un ID visible; si eso deja de ser aceptable habrá que decidir una
  audiencia restringida o enlaces con capacidad.
- **Siguiente decisión:** implementación del primer vertical slice.

### 2026-07-26 — Publicar fija la estructura, no la visibilidad

- **Aprendido:** una liga no es solo un torneo con equipos: necesita reglas y
  emparejamientos para representar una competición. Generarlos al publicar fija
  una estructura reproducible sin tener que adelantar fechas, resultados o
  clasificación.
- **Aprendido:** el ciclo de vida controla qué cambios son válidos; la
  visibilidad controla quién puede descubrir o acceder al torneo. Son conceptos
  independientes y no deben mezclarse en un único estado.
- **Evidencia:** ADR-0032.
- **Coste aceptado:** una liga publicada no permite modificar equipos; las bajas,
  resultados y variantes de reglas exigirán decisiones posteriores.
- **Siguiente decisión:** completada en ADR-0033; continúa incorporación de
  participantes e invitaciones.

### 2026-07-26 — Un borrador no es todavía un torneo

- **Aprendido:** permitir preparar datos sin registro reduce fricción, pero un
  borrador solo local puede perderse cuando la verificación de email continúa en
  otro navegador o dispositivo.
- **Aprendido:** una cuenta pendiente es un estado temporal de alta, no una
  sesión ni una autorización. Asociar el borrador a esa cuenta evita tratarlo
  como dato anónimo en el servidor; verificar el correo es la frontera que
  habilita publicar el torneo.
- **Evidencia:** ADR-0031 y la aclaración de ADR-0010.
- **Coste aceptado:** definir expiración, purga y límites frente a abuso para
  cuentas y borradores pendientes antes de implementarlos.
- **Siguiente decisión:** completada en ADR-0032; continúa visibilidad y
  mecanismo de incorporación.

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

### 2026-07-26 — Diseñar datos y contrato juntos

- **Aprendido:** el estado temporal de una cuenta, el secreto de sesión y el
  recurso de negocio tienen ciclos de vida distintos; modelarlos como una sola
  tabla o token hace más fácil conceder permisos antes de verificarlos.
- **Aprendido:** OpenAPI describe el borde HTTP, mientras que restricciones y
  transacciones pertenecen al dominio y a la persistencia; ambos se revisan
  juntos para evitar contratos que no se pueden garantizar.
- **Evidencia:** ADR-0045, modelo lógico y contrato OpenAPI 3.1 del primer
  incremento, antes de migraciones o handlers.
- **Coste aceptado:** mantener alineados la especificación, las migraciones y
  las pruebas de contrato al evolucionar cada flujo.
- **Siguiente decisión:** tooling de lint, generación y verificación semántica
  de OpenAPI antes de implementar el adaptador HTTP.

### 2026-07-26 — Deriva de contrato como fallo verificable

- **Aprendido:** validar YAML no comprueba por sí solo reglas como declarar
  operaciones públicas o respuestas de error; un linter de OpenAPI convierte
  esas convenciones en una comprobación reproducible.
- **Aprendido:** el cliente generado debe vivir aislado del código manual, pues
  la regeneración puede limpiar su directorio completo.
- **Evidencia:** ADR-0046, Redocly CLI y Orval integrados en Make y pnpm.
- **Coste aceptado:** revisar los diffs de generación y mantener las versiones
  pineadas de las dos herramientas.
- **Siguiente decisión:** migraciones iniciales y consultas SQL del modelo
  aceptado.

### 2026-07-27 — Una migración también es un contrato

- **Aprendido:** el modelo lógico explica relaciones e invariantes; la migración
  demuestra cuáles puede proteger PostgreSQL mediante tipos, FKs, `CHECK` e
  índices únicos.
- **Aprendido:** sqlc no necesita una copia del esquema: analizar las migraciones
  conserva una sola fuente de verdad y evita generar consultas inexistentes.
- **Evidencia:** ADR-0047, primera migración Goose y configuración sqlc con
  `pgx/v5`.
- **Coste aceptado:** revisar cuidadosamente cada evolución de esquema y
  regenerar código SQL solo cuando exista una consulta de producto.
- **Siguiente decisión:** primera operación de persistencia y límites de su
  transacción antes de implementar registro.

### 2026-07-27 — Consolidar antes de compartir

- **Aprendido:** una migración local y no compartida puede consolidarse para que
  la primera instalación parta de un esquema inicial claro; una migración
  compartida se trata como inmutable y se evoluciona con otra migración.
- **Aprendido:** una cuenta pendiente puede tener todos sus datos de identidad;
  verificar correo acredita el canal, no completa un perfil incompleto.
- **Evidencia:** ADR-0048, migración inicial consolidada y reinicio comprobado
  de PostgreSQL local.
- **Coste aceptado:** límites de tasa y purga protegen la reserva temporal de
  usernames y el reenvío de verificación.
- **Siguiente decisión:** primera operación de persistencia y límites de su
  transacción antes de implementar registro.

### 2026-07-27 — Opcional no significa parcialmente válido

- **Aprendido:** una propiedad opcional en el objeto padre puede seguir remitiendo
  a un objeto con campos e invariantes obligatorios; omitir el objeto y enviar un
  objeto incompleto son entradas distintas.
- **Evidencia:** ADR-0052, `RegisterRequest` con `draft` opcional y
  `DraftInput` con `name` y un mínimo de dos equipos.
- **Coste aceptado:** el adaptador HTTP futuro debe comprobar estas reglas en
  runtime además del tipado generado para TypeScript.
- **Siguiente decisión:** primera operación de persistencia y límites de su
  transacción antes de implementar registro.

### 2026-07-29 — La splash nativa no es una pantalla de React

- **Aprendido:** la splash se muestra antes de ejecutar JavaScript; solo puede
  configurarse en la build nativa. El código de la app puede decidir cuándo
  ocultarla, pero no convertirla en un flujo de carga.
- **Evidencia:** `expo-splash-screen` configurado mediante el config plugin y
  retenido hasta hidratar la preferencia local de tema.
- **Coste aceptado:** la apariencia final se prueba en una build release, porque
  Expo Go y development builds no la reproducen fielmente.
- **Regla reutilizable:** no esperar red bajo la splash ni crear una falsa splash
  React salvo que un requisito de producto justifique explícitamente esa espera.

### 2026-07-30 — El cliente generado necesita un transporte de ejecución

- **Aprendido:** generar tipos y operaciones desde OpenAPI no elimina la
  configuración propia del cliente instalado; URL base, cancelación y futuras
  credenciales pertenecen a un transporte común, no a cada feature.
- **Evidencia:** Orval genera un parámetro `fetchFn`; `apiFetch` resuelve la URL
  base y el adaptador de registro invoca la operación generada.
- **Coste aceptado:** las features mantienen un adaptador pequeño que traduce el
  contrato a sus estados de interfaz, sin reconstruir URL, `fetch` ni DTOs.
- **Regla reutilizable:** para una operación del contrato, usar siempre la
  función OpenAPI generada a través del adaptador de la feature y `apiFetch`.

### 2026-07-30 — El fallback de errores debe separar transporte y negocio

- **Aprendido:** el tipo generado por Orval enumera las respuestas documentadas,
  pero no clasifica un rechazo de red ni una respuesta HTTP nueva. Solo el
  transporte común puede identificar con seguridad que no se recibió respuesta.
- **Evidencia:** `apiFetch` convierte esos rechazos en `APIConnectionError` y
  `shared/feedback/request-failure.ts` los traduce a las claves comunes de red o
  error seguro; el chequeo de username conserva su tratamiento específico de
  `200` y `429`.
- **Coste aceptado:** cada feature debe elegir explícitamente sus estados de
  negocio valiosos; los no previstos, cuerpos inválidos y `5xx` usan el fallback
  genérico sin exponer detalles del backend.

### 2026-07-30 — Un enlace universal acredita contexto, no hace un GET mutante

- **Aprendido:** un correo puede llevar a una ruta HTTPS compartida por web,
  iOS y Android, pero abrirlo mediante `GET` no debe activar una cuenta ni crear
  una sesión. El cliente conserva el token de un solo uso fuera de la URL y
  emite automáticamente el `POST` de confirmación.
- **Evidencia:** correo multipart con alternativa accesible, ruta
  `/link/confirm`, configuración declarativa CNG y plantillas de asociación
  bajo `infra/app-links/`.
- **Coste aceptado:** la asociación real exige un dominio, Team ID de Apple y
  huellas Android de la firma de distribución; se documentan como datos de
  despliegue, no se inventan en el repositorio.
- **Regla reutilizable:** un secreto en un enlace no se envía a analítica ni a
  recursos de terceros y se elimina del historial tan pronto como la aplicación
  lo recibe.

## Regla de evidencia

“Entendido” exige una explicación propia y una demostración. Un comando que
funciona o una respuesta del asistente no son evidencia suficiente por sí solos.
