# Mi perfil privado — implementación de ADR-0181

## Alcance

Cuenta incorpora Mi perfil como fila con navegación normal en web/iOS/Android.
Muestra los resultados vigentes de las participaciones vinculadas a la cuenta,
agrupados por deporte: disputados, victorias, empates cuando el formato/deporte
los permite, derrotas y porcentaje de victorias. Sin disputados, el porcentaje
es Sin datos. No son métricas de jugadores individuales ni auditoría de acciones
como administrador. Ser creador/delegado/seguidor no incorpora un equipo ajeno.

El historial incluye partidos resueltos, byes y resultados administrativos o por
incidencias. Identifica equipo propio, torneo, deporte y jornada/ronda; conserva
marcador, sets, penaltis e incidencia y permite consultar el partido del torneo.
Un registro no jugado indica expresamente su exclusión estadística.

Los totales abarcan todos los encuentros que satisfacen el filtro, no solo la
página visible. Filtros por deporte y fragmento literal de nombre del torneo,
aplicados conjuntamente en servidor. Los signos % y _ se buscan como caracteres,
no como comodines. Página de 20, máximo 50; cursores basados en resolución/ID.
Una corrección entre páginas puede alterar el orden: actualizar reconstruye las
páginas cargadas y la colección compartida elimina duplicados.

## Regla de cómputo

Solo state=completed y resultType=played cuentan. No_show, retirement,
administrative y bye no cuentan, aunque exista tanteo o resultado parcial.
Se usa el resultado oficial vigente: si una retirada sustituye un tanteo jugado
por uno administrativo, deja de contar en estos agregados. No se resucitan
versiones antiguas del registro interno de cambios ni se duplican partidos por
corregirlos. Los penaltis resuelven victorias en eliminatoria; el agregado de
sets permite contar resultados sin sumar puntos de deportes diferentes.

## Contrato y privacidad

GET /v1/me/profile toma la cuenta exclusivamente de la sesión, sin ruta ni
parámetro de cuenta ajena; un parámetro accountId se rechaza. No-store; proyección
sin IDs de cuenta. La ruta usa el mismo middleware y cuota de 60/minuto por cuenta
que Mis partidos, compartida entre sesiones. No precisa migración ni nueva tabla.
Consulta sqlc para historia y totales en la misma instantánea de sentencia.
El dominio valida filtros y límites y calcula porcentaje; la infraestructura
proyecta las filas del conjunto privado. Sin nueva biblioteca o almacén analítico.

La feature llama getPersonalProfile generado con authenticatedApiFetch. Valida
runtime de todos los registros, estado, lado propio, deportes, contadores y tasa.
400/429/5xx/cuerpo inválido usan el fallback seguro; transporte y cancelación
conservan la política común. No interpreta title/detail. La colección descarta
respuestas de otra identidad, conserva contenido confirmado ante refresh fallido
y ofrece reintento inicial y de paginación. Filtros se limpian al cambiar cuenta.

## Checklist de cliente

- Textos y accesibilidad en es/en/it/fr; claves semánticas, sin literales visibles.
- ContentSection para analítica/filtros; Card para fichas de encuentro; ListRow
  para el acceso. Márgenes de primitivas sin padding duplicado en Screen/scroll.
- Tokens, tema compartido, objetivo de 44 px y navegación de Cuenta existente.
- Screen bajo cabecera nativa y padding inferior compartido dentro del scroll.
- Sin selector de idioma, nueva dependencia, fetch directo ni DTO manual.
- Carga completa compartida, vacío real, error inicial/reintento y error parcial
  seguro; refresh nativo y acción web. Sin datos inventados ni banner de éxito.
- API y cliente generados desde OpenAPI; cancelación sin feedback y datos de
  otra sesión descartados. No modifica permisos ni operaciones deportivas.

## Validación y limitaciones

Tres integraciones PostgreSQL en contenedor desechable: aislamiento incluso para
administrador/seguidor y cursor ajeno, paginación con totales completos, búsqueda
literal y filtros, corrección y retirada, bye, penaltis, incomparecencia y ocho
deportes separados (incluidos sets). HTTP prueba estados y causas seguras, cuota,
autenticación, filtros y cancelación. Cliente prueba parsing, inconsistencias,
transporte autenticado, filtros y señal, junto a regresiones de colecciones.
Make verify incluye el nuevo test-personal-profile; typecheck/export web obligatorios.

QA visual real web/iOS/Android pendiente: el navegador está deshabilitado en Work.
No se acreditan con compilación la posición del scroll, texto ampliado ni lector.
Las matrices pendientes de ADR-0176/0178 tampoco se dan por cerradas. Esta entrega
no implementa Google Calendar (ADR-0180) ni completa la usabilidad de ADR-0179.
No activa producción ni observabilidad. PostgreSQL de pruebas se apaga al cierre.

## Retrospectiva

Reutilizar navegación, transporte y colección conserva los comportamientos de
sesión y recuperación. Mantener totales y página en una sola proyección evita
estadísticas parciales. El rol de administración y los marcadores administrativos
no deben convertirse accidentalmente en rendimiento deportivo personal.


### Resultado de cierre

`make verify` final termina correctamente: formato/lint, Go/TypeScript, suites de
cliente (incluido perfil), OpenAPI, exportación web con /account/profile,
generación OpenAPI/sqlc sin deriva, tidy, build y vulnerabilidades. La comparación
de generados usa como base los nuevos archivos preparados en el índice; después
se devuelve el índice al estado anterior sin alterar el código ni crear commits.
La integración genérica del gate se omite por no definir su URL; se ejecutan
separadamente las tres pruebas IntegrationPersonalProfile con PostgreSQL 18.6
sintético y esquema/migraciones reales. Las tres pasan. No se afirma haber
repetido toda la suite de integración PostgreSQL. Contenedor tm-profile-qa
apagado al terminar, con volumen y evidencia conservados. No hubo despliegue.


## Publicación v1.16.0

Publicado en web/API/renderer el 2026-11-02, SHA
0e56a27b97fd03069d4d6cf03f965dd69a94ec5d. Ambas CI con PostgreSQL completas
correctas, backup/restauración aislada y 27 comprobaciones HTTPS correctos,
incluido perfil anónimo 401 y nueve scripts con hash coincidente. QA visual y
recorrido autenticado real productivo siguen pendientes. No se distribuye móvil.
[Cierre y rollback](../operations/DEPLOYMENT.md#publicación-v1160--2026-11-02).


## Ajuste de controles — 2026-11-02 (pendiente de publicar)

Por petición del usuario, la actualización de apps usa solo pull to refresh,
también cuando no hay resultados. Web sitúa Actualizar a la derecha de la barra
de navegación, con indicador busy y protección frente a pulsaciones duplicadas.
Filtros ocupa esa barra en ambas plataformas: Stack.Toolbar nativo en iOS con
Liquid Glass y el control compartido en web/Android/iOS anterior. Web agrupa
Filtros y Actualizar con un único margen exterior y separación por tokens.

Filtros abre ModalDialog compartido, con cierre accesible, fondo común y contenido
desplazable limitado por la altura de ventana. El borrador del popup se inicializa
con los filtros aplicados; cerrar/backdrop no aplica cambios. Aplicar confirma y
cierra. Los filtros activos siguen visibles en el contenido del perfil. No cambia
la consulta, los totales, la paginación ni los mensajes HTTP. Textos reutilizados
de los cuatro catálogos; objetivos de 44 px. Native ScrollView dentro del diálogo
no aplica el fade canvas de las pantallas sobre su superficie distinta.

Typecheck, lint, exportación web y 20 regresiones de perfil/colecciones pasan.
QA visual web/iOS/Android pendiente por navegador deshabilitado. No se inicia
local/dev ni observabilidad ni se publica este ajuste con la release anterior.
