# Implementación y validación de Mis partidos y Ediciones

Fecha: 2026-10-11. Decisiones: ADR-0161 y ADR-0162 aceptados.

## Resultado

Cuarta tab, consulta personal paginada, apertura de fase/ronda/partido y lectura
paginada de familia de ediciones implementadas. Se reutilizan las relaciones,
perfiles deportivos, transporte, controles y tokens existentes. No se añaden
migración, dependencia, caché persistente ni permisos nuevos.

## Evidencia

- `make check`: formato Go/TS, lint, tipos, OpenAPI y suite Go correctos.
- Suite Node completa: 109 pruebas correctas; seguridad operacional: 11 pruebas.
- Suite PostgreSQL de integración completa sobre PostgreSQL 18.4 efímero con
  esquema y migraciones: correcta. Participación frente a delegación/seguimiento,
  retiro, cancelación, byes/rivales desconocidos, programación/lugar heredado,
  resultados, cursores, cadena/ramas, orden y mismo nombre independiente.
- Adaptadores: generado + apiFetch, rechazo de cuerpos/estados desconocidos,
  señal de cancelación y feedback seguro. Hook real: fallo de refresh conserva
  datos, paginación/reintento sin duplicados, bloqueo de doble petición y
  respuestas de cuenta/vista antigua descartadas tras abortar.
- HTTP: éxito, validación, autenticación, no encontrado, conexión/restricción/
  consulta fallida, timeout y cancelación real sin cuerpo; causas seguras.
- Generación OpenAPI y sqlc repetida: hashes idénticos antes/después. La
  comprobación usa idempotencia; no se presenta `make verify` como ejecutado
  íntegramente porque su comprobación Git de deriva incluye cambios sin commit.
- Build Go y comprobaciones tidy correctos. govulncheck: cero vulnerabilidades
  alcanzables; informa una vulnerabilidad de módulo sin llamadas afectadas.
- Exportaciones web, iOS y Android correctas. Esto verifica bundles, no builds
  de tienda ni gestos en dispositivos.

## QA visual y funcional web

Chrome con viewport 390 × 844 y API interceptada con datos ficticios sobre la
exportación real; no se acredita como conexión al backend desplegado.

Se comprueban tarjeta personal, jornada 25 en lista virtualizada de 30 jornadas
con equipos y controles visibles, regreso a Mis partidos, vacío por vista,
ediciones propias con campeón y marca actual, errores iniciales y reintento
en ambas colecciones, ausencia de información privada del backend, vacío
global, destino a Final en cuadro y ausencia del botón Ediciones cuando el
indicador es falso. Inspección visual de tarjetas, navegación y márgenes.

Se reproduce React #418 de hidratación del idioma español también en la
pantalla Inicio existente (render estático inglés frente a idioma del navegador).
No se atribuye al incremento y no se modifica la política de idioma de ADR-0143.
No aparecen otros errores JavaScript en estos recorridos.

## Checklist de cierre del cliente

Textos y accesibilidad en es/en/it/fr; tokens/primitivas compartidas; controles
44 px; cards y bloques con margen exterior correcto; safe areas y reserva de
tab bar; cabecera y presentación adaptadas; API detrás de adaptadores; carga,
vacíos, reintentos y cancelaciones; documentación y aprendizaje actualizados.

Las revisiones iOS y Android posteriores se detallan abajo. Pendiente de QA
ampliada en dispositivos: pull to refresh en iOS, lector de pantalla, texto
ampliado, temas y movimiento reducido.
También queda la matriz visual ampliada de grupos/desempates e inicio en frío.
Las decisiones están implementadas; esta validación pendiente se declara por
separado y no se acredita con exportaciones o la maqueta.

## Retrospectiva técnica

La existencia global debe venir del backend: una vista vacía no demuestra que
no existan partidos. La relación de repetición permite recorrer una familia sin
un modelo adicional. Localizar texto en el DOM no certifica visibilidad: se
mide el partido y la cabecera sticky para situar todo el contenido. Una ruta
absoluta de exportación evita comprobar por error un bundle anterior.

Los recursos de prueba se cierran al terminar; no se despliega ni se activa
observabilidad ni se modifican servicios de producción.


## Revisión iOS en simulador — 2026-10-11

iPhone 18 Pro, iOS 27.0, development build local existente conectada a Metro
con código actual y API Go local real. PostgreSQL local 18.6; volumen existente
conservado, migraciones aplicadas hasta 22; únicamente API, PostgreSQL y Mailpit.
Observabilidad apagada. Datos ficticios creados para esta revisión, sin alterar
cuentas o torneos anteriores.

| Recorrido | Resultado observado |
| --- | --- |
| Mis partidos sin sesión | Card de invitado y acción de acceso, con tab bar nativa |
| Iniciar sesión desde esa card | Detectado destino incorrecto a Datos de acceso; corregido a /account y formulario habitual verificado |
| Lista personal | Equipo propio, fecha/lugar heredado, partido programado primero, márgenes y controles correctos |
| Partido pendiente programado | Abre Jornada 4, con equipos, programación y controles completamente visibles |
| Regreso desde partido | Conserva tab personal y vista elegida; probado también desde Resultados |
| Resultados con API local interrumpida | Error común de conexión y Reintentar; al restablecer recupera resultados de la primera edición en la misma vista |
| Ediciones desde ficha | Cabecera nativa, orden reciente primero, estado y campeón propio; sin tab bar |
| Abrir ficha histórica y sus ediciones | Conserva sus resultados y cambia Edición consultada al miembro correcto de la misma familia |
| Ediciones con API interrumpida | Error inicial con Reintentar; recupera la misma familia y marca actual |
| Fallo de actualización de Ediciones al regresar | Conserva ambas tarjetas y muestra feedback seguro |
| Cambiar cuenta | No conserva partidos de la cuenta anterior |
| Pendientes vacío con resultados existentes | Mensaje por vista; Resultados sigue mostrando encuentros de su equipo |
| Vacío global | Mensaje global en Pendientes y Resultados; la acción Torneos abre la tab correspondiente |

La automatización de Device Hub permite pulsar y leer controles, pero sus
arrastres y desplazamientos no produjeron movimiento observable ni nueva
petición de refresh. No se acredita el gesto pull to refresh como aprobado ni
como fallo de producto; se solicita comprobación manual. La operación de
actualización y conservación de datos sí se comprueba mediante regreso/focus.

Quedan pendientes el gesto, la matriz visual ampliada, temas, texto aumentado,
VoiceOver y arranque en frío. El Mac se bloqueó durante la revisión; tras el
desbloqueo del usuario se retomó y comprobó el vacío global. Android todavía
no se ha abierto: se revisa una app cada vez según la petición del usuario.

`expo install --check` propone cinco parches posteriores del SDK 57; se usa la
development build instalada, sin añadir módulos ni modificar dependencias en
esta revisión. Typecheck, tests de browsing y exportación web pasan tras la
corrección del destino de acceso. No se certifica una build de tienda.

Al cerrar esta sesión se detienen Metro y los servicios locales de Compose;
no quedan servicios de esta sesión ejecutándose y se conservan los volúmenes
y fixtures. La comprobación manual del gesto requiere reanudar la sesión de
pruebas; permanece pendiente y no se marca iOS como completamente aprobado.

## Revisión Android en emulador — 2026-10-11

Pixel API 34, Android 14, development build local existente conectada al bundle
actual de Metro. API real en `10.0.2.2:8080`, PostgreSQL y Mailpit locales,
sin observabilidad. Se reutilizan las cuentas y torneos ficticios de iOS.
El usuario autoriza expresamente ADB para pulsaciones, gestos y lectura tras
comprobar que el control de Android Studio dirigía los clics al editor.

| Recorrido | Resultado observado |
| --- | --- |
| Invitado y acceso | Cuarta tab nativa, card y formulario habitual de Cuenta correctos |
| Lista personal | Programado primero, equipos, jornada, fecha y lugar heredado correctos |
| Pull to refresh de Mis partidos | Gesto real produce una nueva consulta de API y conserva la lista |
| Partido programado | Jornada 4 y tarjeta completa visibles; comprobado antes y después de la corrección, también con cuenta sin permisos de edición |
| Regreso | Atrás desde partido conserva Pendientes; cierre circular conserva Resultados |
| Resultados con API interrumpida | Error común y Reintentar; restablecer API recupera los tres resultados en la misma vista |
| Fallo de refresh personal | Mantiene resultados y muestra banner de conexión seguro |
| Ediciones desde ficha histórica | Orden reciente primero, marca Edición consultada, estado y campeón propios, cierre circular y sin tab bar |
| Abrir otra edición | Ficha de Edición 2 en curso, con su lugar y partidos propios |
| Pull to refresh de Ediciones | Nueva consulta de la misma familia tras el gesto |
| Error inicial de Ediciones | Error común con Reintentar; recupera familia y edición consultada correctas |
| Fallo de refresh de Ediciones | Conserva ambas tarjetas y muestra banner seguro |
| Cambio de cuenta y vacío global | No conserva encuentros de la cuenta anterior; mensaje global en ambas vistas y botón lleva a Torneos |
| Pendientes vacío con resultados | Mensaje por vista; Resultados muestra los tres encuentros de Halcones |

Se detecta que la alineación automática no termina cuando un partido final
queda visible pero no puede alcanzar la posición superior deseada. Se limita
el offset al recorrido real del contenido y se conserva la vista nativa del
contenedor medido con `collapsable={false}`. Después de recargar por completo
el cliente se vuelve a abrir el resultado, se comprueba su visibilidad y se
desplaza hasta el resumen sin que el ajuste lo devuelva al partido. El enlace
al pendiente también permite desplazar después. La regresión de este ajuste
compartido en iOS queda pendiente; la regresión web de jornada 25 pasa.

El primer arranque quedó en blanco; un reinicio de la app local permitió
continuar. Se observa el aviso de actualización de estado anterior al montaje
durante ese arranque; no se atribuye a esta feature ni se da por resuelto con
la revisión. Las rutas se observan después de completar su carga; una captura
de la transición no se considera evidencia del estado final.

Tras el ajuste pasan typecheck, lint, los cinco tests de browsing, exportación
web y la regresión web de navegación, vacíos, errores/reintentos, cuadro y ficha
sin otras ediciones. No se certifica build de tienda. Quedan pendientes
TalkBack, texto ampliado, temas, movimiento reducido y la matriz nativa ampliada
de grupos, desempates, cuadros y paginación. La ficha independiente sin botón
Ediciones se verifica en web, no se acredita aquí como recorrido Android.

Retrospectiva: comprobar la tarjeta visible no basta; el deep link debe liberar
el desplazamiento posterior, incluido el caso donde el final de lista limita
la alineación. Un gesto de actualización se contrasta con una nueva consulta,
y cada error se valida en la colección que la persona estaba consultando.

Al cerrar la revisión se detienen Metro, el emulador Pixel, ADB y los servicios
locales de Compose; la consulta de servicios activos queda vacía. Se conservan
volúmenes y fixtures, incluida la participación histórica ficticia usada para
el vacío por vista. No se despliega ni se modifican servicios de producción.

## Revisión iOS del ajuste de alineación — 2026-10-11

Se reabre una sesión local sin observabilidad en iPhone 18 Pro / iOS 27.0 y se
recarga completamente el bundle corregido, con API local en 127.0.0.1:8080.
La cuenta ficticia histórica muestra el vacío de Pendientes y sus tres
Resultados. Abrir Águilas contra Halcones de Jornada 3 sitúa toda la tarjeta
visible al final de la lista y permite observar su resaltado temporal.

El arrastre de Device Hub no produce movimiento ni una nueva consulta personal
(contador antes y después: 2). Se deja el detalle preparado y se pide al usuario
comprobar desplazamiento hasta el resumen y, después de cerrar, pull to refresh
en Mis partidos. Hasta recibir ese resultado no se acredita ninguna de esas
dos comprobaciones como aprobada. El posicionamiento y resaltado sí se observan
con el código corregido; no se modifica código en esta revisión.

El usuario pospone esas dos comprobaciones manuales para una revisión futura.
Quedan pendientes, sin bloquear el cierre de esta sesión ni acreditarlas como
aprobadas. Se apagan API/dependencias locales conservando los volúmenes; la
consulta de servicios activos queda vacía y no hay listener Metro en 8082.
Observabilidad apagada. La comprobación futura requiere reanudar las pruebas.

## Continuación iOS por casos — 2026-10-11

El usuario elige iOS primero y pide continuar sin su intervención al no estar
cerca del equipo. Se recarga el código actual con API local real y sin
observabilidad. Pendientes muestra el vacío por vista y Resultados recupera los
tres encuentros históricos; abrir Águilas–Halcones de Jornada 3 deja la tarjeta
completa visible.

**Caso 1 pendiente de comprobación manual futura:** desde ese partido, deslizar
hasta el resumen superior y confirmar que la pantalla permanece donde se deja,
sin volver sola a Jornada 3. No se acredita como aprobado. También siguen
pendientes los gestos pull to refresh de Mis partidos y Ediciones en iOS cuando
no se consiga producir y contrastar una nueva consulta mediante automatización.

Preparación: Metro se inicia priorizando IPv4 en localhost para coincidir con
la development build. Se corrige el prefijo `/v1` omitido inicialmente en la URL
de API de esta sesión; esos 404 son un error de preparación, no un defecto de
producto. La recarga inicial requiere recuperar el launcher tras reiniciar el
simulador; no se acredita este recorrido como una prueba de arranque Release.

### Casos adicionales observados en esta continuación

| Caso | Resultado y alcance |
| --- | --- |
| Ediciones desde el histórico | Orden reciente primero, campeón Águilas y Edición 1 marcada; árbol accesible con botón seleccionado |
| Tema oscuro | Ediciones, ficha activa y Mis partidos usan superficie, texto y cabecera oscuros; tarjetas y controles legibles en la muestra |
| Abrir Edición 2 y consultar su familia | Conserva estado en curso y Pabellón de pruebas; misma familia, ahora con Edición 2 seleccionada |
| Regreso por la navegación anidada | Vuelve por Ediciones y ficha histórica hasta Mis partidos con Resultados seleccionado |
| Texto aumentado un paso | Ediciones y Mis partidos conservan títulos, tarjetas y controles legibles |
| Texto aumentado cuatro pasos desde el inicial | Mis partidos conserva tabs y tarjetas visibles legibles; el vacío de Pendientes se adapta en varias líneas; Ediciones conserva título, nombres, estados y campeón sin recorte observado |
| Restauración | Cuatro decrementos y tema claro restauran la apariencia inicial, comprobada visualmente |

No se equipara la inspección del árbol accesible con una prueba de VoiceOver.
No se acredita alcance a todas las tarjetas de una lista larga mediante gesto,
ni todos los tamaños de accesibilidad, ni toda transición animada. Continúan
pendientes VoiceOver, movimiento reducido, enlaces en frío y matriz nativa de
grupos/desempates/cuadros/paginación, además de los gestos manuales anteriores.
No se modifica código ni dependencias de producto.

Retrospectiva: las preferencias del SO se verifican sobre pantalla y cabecera
nativas, y se restauran al cerrar. Un tamaño aumentado observado no certifica
los extremos ni el desplazamiento; esa cobertura se registra por separado.

## Continuación Android: tema, escalado y movimiento reducido

Pixel API 34 / Android 14; iOS apagado antes de arrancar Android. API local
real, mismos datos ficticios y bundle actual de Metro; observabilidad apagada.
Se reutiliza la autorización expresa de ADB de la revisión anterior.

| Caso | Resultado observado |
| --- | --- |
| Tema oscuro, tamaño normal | Resultados, tarjetas y tab bar adaptados y legibles |
| Texto al 200%, oscuro | Tabs y vacío de Pendientes legibles; tarjetas de Resultados legibles y último resultado completo alcanzable mediante gesto |
| Abrir último resultado visible al 200% | Abre su jornada; un gesto hacia el resumen desplaza la ficha sin devolverla al destino |
| Ediciones al 200%, oscuro | Cabecera, edición consultada, nombres, estados y campeón completos; abre Edición 2 con sus datos propios |
| Animaciones desactivadas, 200%, enlace con bundle activo | Partido programado de Jornada 4 completo, fecha/zona/lugar legibles; permite desplazar después |
| Enlace desde app cerrada | Abre el launcher de Expo; no acredita enlace en frío de producto ni Release |
| TalkBack | Servicio y exploración táctil activos confirmados; entradas ADB, toque/gesto de hardware y atajo no producen foco observable. No se acredita navegación ni locución como aprobadas |

Al 200% la tab nativa abrevia visualmente «Mis partidos», aunque el árbol
accesible conserva el nombre completo. Cambiar font_scale recrea la actividad
y vuelve a Inicio; se reabre la colección para cada revisión. No se atribuye
ese reinicio a una pérdida de datos de la feature.

El primer enlace de launcher con app activa falla con
`App react context shouldn't be created before`; force-stop y apertura del
enlace de development client permiten recuperar la app. Se registra como
incidencia observada de development build, sin diagnóstico definitivo ni
certificación de Release. No se modifican módulos ni dependencias.

Evidencia: [Resultados oscuros](evidence/tournament-browsing-2026-10-11/android-dark-results.png),
[vacío al 200%](evidence/tournament-browsing-2026-10-11/android-dark-large-empty.png),
[último resultado al 200%](evidence/tournament-browsing-2026-10-11/android-dark-large-results-bottom.png),
[Ediciones al 200%](evidence/tournament-browsing-2026-10-11/android-dark-large-editions.png),
[destino con animaciones desactivadas](evidence/tournament-browsing-2026-10-11/android-large-reduced-motion-match.png).

Se restauran font_scale=1.0, modo claro, preferencias accesibles anteriores y
valores originales de animación. Emulador y ADB apagados antes de pasar a web.
TalkBack efectivo, arranque sin launcher y matriz nativa ampliada de formatos y
paginación siguen pendientes. El estado final no es «todo Android aprobado».

Retrospectiva: un comando de toque que activa una tarjeta no demuestra foco de
TalkBack; hay que verificar el foco y la activación del lector por separado.
La navegación en frío de una development build puede detenerse en su launcher.

## Continuación web y corrección accesible

Navegador integrado, API local real y misma cuenta ficticia; después de apagar
Android. Ediciones permite navegación con Tab y Enter, marca la edición
consultada y abre la edición más reciente con sus datos propios. La carga
directa de Edición 2 con matchId muestra completo el partido de Jornada 4,
fecha y sede; desplazar hacia jornadas anteriores no devuelve al destino.
Mis partidos recupera tres resultados, abre Jornada 3 y conserva Resultados
al regresar. Pendientes presenta el vacío correspondiente.

Se detecta que `accessibilityState.selected` no genera `aria-selected` en el
adaptador web instalado. Se sustituye solo esa propiedad por `aria-selected`,
alias universal que React Native traduce al mismo estado nativo. Tras recarga
completa, DOM confirma Pendientes=false y Resultados=true; abrir un resultado
y volver conserva Resultados=true y el árbol accesible lo marca seleccionado.
No se cambian tokens, tamaños, textos localizados ni operaciones HTTP.

Validación de la corrección: typecheck del repositorio, ESLint sin warnings y
Prettier del archivo, y exportación web Expo completados correctamente.
La revisión nativa anterior precede al cambio; la compatibilidad del alias se
contrasta con tipos y código del React Native instalado, sin acreditar una
nueva sesión nativa posterior al cambio.

Preparación: usar página localhost y API 127.0.0.1 produce login 200 seguido de
401. Con ambos en 127.0.0.1 la sesión se conserva; se registra la configuración
de hosts como causa probable, sin atribuir un defecto a la feature.

Retrospectiva: la selección visual y el foco no prueban la selección semántica.
Comprobar el atributo en el DOM y después de volver detecta diferencias entre
adaptadores sin añadir dependencias ni abstraer un control de dos opciones.
Siguen abiertos los casos manuales y límites nativos enumerados arriba;
esta sesión no certifica toda la matriz de sistemas operativos ni Release.

Cierre: iOS y Android apagados, pestaña temporal cerrada, Metro detenido y
`make dev-down` completado. API local sin conexión y puerto 8082 sin listener;
volúmenes y evidencia conservados. Observabilidad no se activó durante el QA.

## Matriz ampliada y límites de consulta — 2026-10-12

Se reanuda iOS por separado, con bundle actual y API local real. Se crean
fixtures ficticios bajo la cuenta QA histórica: grupos de ocho equipos,
desempate tras liga de tres equipos empatada, cuadro de ocho equipos y liga
de 24 equipos. El cuadro se resuelve y finaliza mediante la API para producir
un destino a Final y una familia de 21 ramas más la raíz. Estos datos amplían
las colecciones anteriores; la cuenta ya no presenta Pendientes vacío.

| Comprobación API real | Resultado |
| --- | --- |
| Pendientes con tamaño predeterminado | 20 + 8 filas; 28 IDs únicos; mismo orden que consulta con límite 50 |
| Liga de 24 equipos | Los 23 partidos reales del equipo vinculado están en la colección paginada |
| Resultados | 8 filas únicas; mismo orden con límite 50 |
| Familia de ediciones | 20 + 2 filas; 22 IDs únicos; más reciente primero; raíz con campeón Matriz 1 |
| Lectura desde dos ramas distintas | Misma familia completa y mismo orden que desde la raíz |
| Sin sesión, vista inválida, límite mayor de 50, cursor inválido y torneo inexistente | 401/400/404 esperados, application/problem+json, sin detail/trazas/SQL |

Se descubre `limit=0` aceptado como 200, contrario al mínimo 1 de OpenAPI.
El adaptador distingue límite explícito vacío/cero de omisión; ambas
colecciones rechazan los primeros con 400 y conservan el valor predeterminado
al omitir el parámetro. Regresión HTTP de omisión, 1/50 y vacío/0/-1/51,
con causa validation.rejected en el span raíz. API local confirma la corrección.
No se cambia el contrato ni se crea mensaje específico de negocio en cliente.
Suite Go completa correcta; la primera ejecución queda limitada por el sandbox
al abrir puertos httptest y se repite con acceso autorizado a puertos locales.

Evidencia: [resultados y destinos ficticios de API](evidence/tournament-browsing-2026-10-11/matrix-api-2026-10-12.json).

La app iOS carga el bundle y muestra los fixtures en Actividad reciente.
El Mac se bloquea durante la preparación y el usuario lo desbloquea. Device Hub
no expone los elementos individuales de la tab bar y los clics por coordenadas
fallan incluso tras traer la ventana al frente y centrarla. Safari expone la
hoja de apertura, pero no su botón Abrir; el enlace canónico a Mis partidos
queda esperando ese botón. No se acredita aún apertura nativa de los destinos,
paginación por botón/gesto ni regresión nativa de aria-selected en esta sesión.
Se solicita únicamente destrabar ese aviso; los gestos manuales pospuestos
siguen pendientes y no se sustituyen por pruebas de API.

Retrospectiva: el límite ausente y el cero explícito son casos diferentes del
contrato. Una familia de ramas real valida el cursor y la independencia de su
punto de entrada; comprobar sus páginas en API no certifica accesibilidad ni
desplazamiento de la interfaz nativa.

## Android: destinos y regreso tras paginación — 2026-10-12

iOS apagado antes de iniciar Pixel API 34. Development build, API real local,
texto normal (font_scale=1.0), apariencia oscura observada; no se cambian tema
ni escala en este bloque. Bundle recargado por completo para la regresión.

| Caso | Resultado observado |
| --- | --- |
| Selección accesible tras aria-selected | Pendientes y Resultados seleccionados en el árbol nativo según la vista elegida |
| Mis partidos, Cargar más | Gesto alcanza el control; carga página adicional y permite llegar a grupos/desempate |
| Grupo 1 · Jornada 3 | Desde la tarjeta personal abre Matriz 1–Matriz 4 completo, con controles visibles |
| Desempate | Selecciona la fase de desempate y Ciclo 1; Matriz 1–Matriz 3 completo y resaltado |
| Cuadro, resultado de Ronda 3 | Selecciona Final; Matriz 1–Matriz 3, marcador 2–0 y campeón completos y visibles |
| Ediciones, fallo de página adicional | Conserva tarjetas, muestra mensaje común de conexión y Reintentar alcanzable por gesto |
| Ediciones, reintento | Recupera Rama 1 y raíz; raíz consultada marcada y campeón Matriz 1 |
| Regreso desde Rama 1 | Conserva la segunda página de Ediciones y su posición; raíz sigue marcada |

Se detecta que volver desde un partido de la segunda página sustituye las 28
filas por las primeras 20 y desplaza la posición. Se corrige useCollection:
refresh vuelve a consultar desde el inicio hasta cubrir el tramo previamente
cargado o agotar las páginas. La sustitución es atómica; un error intermedio
conserva todos los datos anteriores. No se fusionan filas antiguas con datos
actualizados ni se introduce caché. Las peticiones continúan por adaptadores
generados y apiFetch, con AbortSignal y fallback seguro.

Tras recarga completa, volver desde Grupo 1 · Jornada 3 conserva la segunda
página y exactamente la misma posición observada. Regresión compartida también
confirmada al volver de Rama 1 en Ediciones. Fast Refresh no bastaba para
validar el cambio sobre la instancia anterior.

Validación: siete tests de browsing correctos, incluyendo sustitución de filas
ya ausentes, fallo intermedio y cancelación por nueva identidad; typecheck,
ESLint sin warnings y exportación web correctos. Checklist cliente: sin cambios
de texto, tokens, layout, controles, rutas ni permisos; estado local de feature,
adaptador generado + apiFetch, cancelación y feedback seguro conservados.
La corrección compartida queda pendiente de comprobación visual iOS por el
bloqueo de control documentado; exportación no acredita esa regresión.

En el primer arranque Android se observa pantalla en blanco y aviso de estado
antes del montaje. Tras force-stop, reapertura y espera de carga la interfaz
queda disponible; el aviso no se da por resuelto ni se atribuye a browsing.

Evidencia: [grupos](evidence/tournament-browsing-2026-10-11/android-matrix-groups.png),
[regreso anterior](evidence/tournament-browsing-2026-10-11/android-matrix-return-before.png),
[regreso corregido](evidence/tournament-browsing-2026-10-11/android-matrix-return-second-page.png),
[desempate](evidence/tournament-browsing-2026-10-11/android-matrix-tiebreak.png),
[final](evidence/tournament-browsing-2026-10-11/android-matrix-final.png),
[fallo de página](evidence/tournament-browsing-2026-10-11/android-matrix-editions-page-error.png),
[segunda página](evidence/tournament-browsing-2026-10-11/android-matrix-editions-second-page.png),
[regreso a Ediciones](evidence/tournament-browsing-2026-10-11/android-matrix-editions-return.png).

Retrospectiva: una navegación correcta con una sola página no cubre el regreso
con páginas adicionales. La reconstrucción del tramo cargado añade consultas
proporcionales a ese tramo y mantiene datos actuales sin otra capa de estado.

Cierre de este bloque: iOS apagado por menú, Android y ADB apagados, Metro
detenido y make dev-down completado; Compose sin servicios activos y puerto
8082 sin listener. Fixtures y volúmenes conservados, observabilidad apagada.
El aviso Abrir de Safari no requiere ya intervención en esta sesión cerrada;
se retomará el control iOS junto con su regresión pendiente en una nueva sesión.

## Sesión manual guiada — 2026-10-12

El usuario solicita primero QA manual y después reintentar automatización.
Se reanudan API local y Metro sin observabilidad; iOS arrancado, Android apagado.
Se prepara el enlace al development client actual en Safari y se solicita pulsar
Abrir porque la hoja no expone el botón al control accesible. Esperando resultado
de preparación; ningún caso manual nuevo se acredita todavía. Primer caso
previsto: desplazamiento al resumen después de abrir un partido, sin retorno
automático al destino. Servicios activos solo durante esta sesión de pruebas.


## Continuación autónoma y recuperación iOS — 2026-10-12

El usuario pide continuar sin intervención manual y recuperar automatización.
La solicitud anterior de pulsar Abrir deja de ser requisito para continuar.

Web, API real, sesión existente, viewport 1280 × 720:
- Mis partidos: 28 tarjetas tras paginar; Grupo 1 · Jornada 3 muestra
  Matriz 1–Matriz 4 y controles completos. Volver conserva 28 tarjetas,
  sin Cargar más y Pendientes con aria-selected=true.
- Ediciones: 22 entradas; Rama 1 muestra Rama Azul/Rama Roja y estado propio.
  Volver conserva 22 entradas, sin Cargar más; raíz marcada Edición consultada,
  finalizada y campeón Matriz 1.

iOS: Show in Device Hub abre la ventana completa; Capture Keyboard y toque
dentro de la pantalla permiten Return. En esta ventana funcionan coordenadas,
incluido Abrir de Safari. Cliente nativo actual cargado. Inspector muestra
Light, texto 3, Reduce Motion y VoiceOver apagados; no se cambian preferencias.

- Pendientes y Resultados seleccionados en el árbol accesible según la vista:
  corrección aria-selected revalidada en iOS.
- Gesto alcanza Cargar más; segunda página disponible, 28 tarjetas.
- Grupo 1 · Jornada 3: partido y controles completos. Arrastre desde margen
  derecho vuelve al resumen y Jornada 1 sin retorno automático al destino:
  ese caso pendiente de desplazamiento queda aprobado.
- Volver conserva 28 tarjetas; posición exacta no acreditada por falta de
  comparación visual antes/después.
- Desempate seleccionado, Ciclo 1, Matriz 1–Matriz 3 resaltado y completo.
- Final seleccionada, Matriz 1–Matriz 3, marcador 2–0 y ganador completos.
- Ediciones: 22 entradas tras paginar, Rama 1 con equipos propios;
  volver conserva 22 entradas y raíz seleccionada.

Recarga por gesto sigue pendiente: API local interrumpida, lista llevada al
inicio y arrastre adicional sin petición/spinner/feedback verificable.
No se considera aprobado ni fallo de producto. El Mac se bloquea antes de
comprobar refresco al recuperar foco; no se solicita intervención al usuario.
Lectura/navegación real VoiceOver/TalkBack y apertura en frío de Release
siguen sin certificar.

Siete tests automatizados de browsing correctos, incluidos refresco atómico,
fallo intermedio y cancelación por identidad. Sin cambios de implementación
en este bloque. Capturas inspeccionadas en sesión, sin nuevas imágenes guardadas.

Retrospectiva: la vista completa recupera toques y gestos bloqueados en la
compacta. Todas las filas en el árbol no prueban posición exacta ni lector.

Cierre: API restaurada tras interrupción y después make dev-down completado;
Metro detenido, iPhone 18 Pro de esta sesión apagado mediante su identificador.
Volúmenes y fixtures conservados; observabilidad no iniciada.


## Reanudación tras desbloqueo: foco y VoiceOver — 2026-10-12

API y Metro reanudados sin observabilidad, solo iPhone 18 Pro activo.
Mis partidos parte de la primera página (20 tarjetas). API local interrumpida:
arrastre y scroll ascendente no disparan recarga verificable; pull-to-refresh
permanece pendiente. Cambiar a Inicio y volver sí muestra el mensaje común de
conexión, conserva las 20 tarjetas y selección Pendientes. API restaurada:
volver a la pantalla recupera la consulta y contenido sin mensaje nuevo.

VoiceOver activado desde inspector, aviso de gestos aceptado. Toque en
Pendientes y Control+Option+Right mueve foco visible a Resultados; otro avance
lo mueve a una tarjeta. Control+Option+Space no produjo activación verificable.
Doble toque en Resultados cambia la vista y el estado seleccionado.
Toque en primera tarjeta de cuadro muestra el marco de foco; doble toque abre
la Final con Matriz 1–Matriz 3, 2–0 y ganador en el árbol accesible.
Se acredita ese recorrido de foco y activación, no locución audible,
orden completo de todas las páginas ni retorno automático del foco.
VoiceOver restaurado apagado y confirmado en inspector.

Se inicia build Release local de simulador, esquema FastTourneyLocal,
API local explícita, sin firma ni publicación; resultado pendiente.

### Resultado de Release y cierre de la reanudación

La compilación sin firma terminó correctamente, pero al abrir el enlace solo
mostró el error común de petición. No se acredita su causa: el acceso a la
sesión nativa era una hipótesis de preparación, no un fallo demostrado de la
feature. Se recompiló el mismo código con firma ad hoc de simulador
(`CODE_SIGNING_ALLOWED=YES CODE_SIGN_IDENTITY=- CODE_SIGN_STYLE=Manual`).
Ambas compilaciones terminaron con `BUILD SUCCEEDED`.

Con la app firmada instalada y Metro detenido, Safari abrió desde la app
cerrada el enlace de la Final del fixture Cuadro. Se observaron el título
QA matriz · Cuadro, estado Torneo finalizado, Final seleccionada, Matriz 1
con corona y marcador 2 frente a Matriz 3 con 0. No apareció Dev Launcher.
Se aprueba apertura en frío de ese enlace en Release local de simulador;
no acredita universal links, dispositivo físico ni distribución de tienda.
La sesión existente también carga Inicio, Cuenta y Mis partidos.

En Release se interrumpió otra vez la API local y se probaron arrastres sobre
la tarjeta y el margen, además de scroll ascendente. No hubo spinner ni
feedback que acreditase invocación de RefreshControl. Pull-to-refresh iOS
sigue pendiente; los intentos no prueban un fallo del producto. La API se
restauró antes del cierre. VoiceOver mantiene la aprobación parcial descrita
arriba; locución, orden completo y retorno de foco siguen pendientes.

La preparación CLI reproducible, sin añadir un framework de E2E, es:

```sh
cd apps/client/ios
APP_ENV=local EXPO_PUBLIC_API_BASE_URL=http://127.0.0.1:8080/v1 \
  xcodebuild -workspace FastTourneyLocal.xcworkspace \
  -scheme FastTourneyLocal -configuration Release -sdk iphonesimulator \
  -destination 'id=1094CE55-E226-442A-A092-9EB662E897EA' \
  -derivedDataPath /tmp/tm-qa-release-derived \
  CODE_SIGNING_ALLOWED=YES CODE_SIGN_IDENTITY=- CODE_SIGN_STYLE=Manual build
xcrun simctl install 1094CE55-E226-442A-A092-9EB662E897EA \
  /tmp/tm-qa-release-derived/Build/Products/Release-iphonesimulator/FastTourneyLocal.app
```

El identificador corresponde al simulador usado en esta sesión y debe cambiarse
si se usa otro. La preparación no arranca API, Metro ni observabilidad. La
apertura y los gestos se verifican por UI; compilar o instalar no los aprueba.

Retrospectiva: separar preparación nativa de evidencia funcional evitó tocar
la feature ante el error de la build sin firma. Una Release con firma ad hoc
permite comprobar el enlace sin depender de Metro. La automatización actual
cubre compilación, instalación y regresiones de colección; la herramienta UI
todavía no acredita el refresco por gesto ni la locución del lector.

Cierre verificado: `make dev-down` completado, simulador de esta sesión
apagado y sin listeners en 8082/8098. Volúmenes y fixtures conservados;
observabilidad no iniciada. `git diff --check` sin errores.

## Android Release y automatización de enlaces — 2026-10-12

Pixel API 34 / Android 14, arm64-v8a. iOS permanece apagado. API, PostgreSQL
y Mailpit locales arrancados solo para QA, sin observabilidad ni Metro.
Se reutilizan sesión y fixtures; no se crean otros torneos ni se modifican datos.

`:app:assembleRelease` termina `BUILD SUCCESSFUL`. APK local no depurable
(confirmado mediante rechazo de `run-as`), firmada con la clave de pruebas
existente, no con credenciales de publicación. Primer enlace desde force-stop:
`LaunchState: COLD`, ruta Torneo sin launcher y fallback común de conexión.
La variante Release no declara la excepción HTTP que usa Debug para la API local.

Se prepara un overlay optativo de QA, separado de la configuración del producto
y del árbol Android generado. `base-config` mantiene HTTP bloqueado y solo
permite el host del emulador `10.0.2.2`, sin subdominios. El init script rechaza
otra applicationId distinta de `com.fasttourney.app.local`. Con este overlay,
Release carga el enlace y conserva la sesión. No se publica esta APK ni se
amplía la política de red de las variantes dev/prod.

### Smoke automatizado

`tests/android-release-links-smoke.py` usa ADB autorizado, biblioteca estándar
Python y fixtures existentes. Requiere un emulador, APK no depurable y lectores
apagados; registra el modo de entrada sin restringirlo. Por cada caso:
force-stop, enlace por intent, verificación COLD,
comparación exacta de textos/descripciones de nodos que intersectan la pantalla,
captura PNG, XML y resultado del lanzamiento. Falla con código distinto de cero
si no alcanza el destino esperado. No arranca servicios ni crea datos.

| Caso | Resultado |
| --- | --- |
| Grupos | PASS COLD; Grupo 1 · Jornada 3, Matriz 1–Matriz 4, partido resaltado y controles visibles |
| Desempate | PASS COLD; Desempate 1 · Ciclo 1, Matriz 1–Matriz 3 resaltado y controles visibles |
| Final | PASS COLD; Final, Matriz 1–Matriz 3, 2–0 y Ganador |

Las primeras expectativas del smoke usaban Jornada 3/Ciclo 1 como textos
independientes; el inspector mostró encabezados compuestos y se corrigieron
las expectativas. Los tres casos pasan con los textos exactos. Capturas
inspeccionadas además de las aserciones. Esto aprueba esquemas locales en
Release de emulador; no certifica app links HTTPS ni distribución de tienda.

Evidencia: [grupos](evidence/tournament-browsing-2026-10-11/android-release-groups.png),
[desempate](evidence/tournament-browsing-2026-10-11/android-release-tiebreak.png),
[final](evidence/tournament-browsing-2026-10-11/android-release-bracket.png),
[casos reutilizables](evidence/tournament-browsing-2026-10-11/android-release-cases.json).

Preparación reproducible desde la raíz, con el SDK/JDK existentes y el emulador
ya arrancado. La primera compilación necesita API local disponible para la
prueba posterior, no para generar el bundle:

```sh
export TM_QA_ANDROID_OVERLAY="$PWD/docs/engineering/evidence/tournament-browsing-2026-10-11/android-release-overlay"
JAVA_HOME='/Applications/Android Studio.app/Contents/jbr/Contents/Home' \
ANDROID_HOME="$HOME/Library/Android/sdk" APP_ENV=local \
EXPO_PUBLIC_API_BASE_URL=http://10.0.2.2:8080/v1 \
  apps/client/android/gradlew -p apps/client/android :app:assembleRelease \
  -PreactNativeArchitectures=arm64-v8a -I "$TM_QA_ANDROID_OVERLAY/init.gradle"
"$HOME/Library/Android/sdk/platform-tools/adb" -s emulator-5554 install -r \
  apps/client/android/app/build/outputs/apk/release/app-release.apk
"$HOME/Library/Android/sdk/platform-tools/adb" -s emulator-5554 shell input touchscreen tap 1 1
python3 tests/android-release-links-smoke.py \
  docs/engineering/evidence/tournament-browsing-2026-10-11/android-release-cases.json \
  --adb "$HOME/Library/Android/sdk/platform-tools/adb" --serial emulator-5554 \
  --evidence /tmp/tm-qa-android-release-smoke
```

Serial, arquitectura y rutas SDK/JDK corresponden a esta sesión. Los casos
requieren esos fixtures conservados y sesión de QA; en otra base se deben pasar
otros casos con sus IDs y expectativas. El toque en la esquina de la barra de
estado restaura el modo táctil nativo tras usar teclado; el smoke registra ese
modo como contexto de prueba. El smoke no instala, cambia preferencias
ni realiza login. No sustituye la revisión visual ni los lectores de pantalla.

### TalkBack parcial

TalkBack 14.2 instalado. Activación directa en Settings.Secure enlazó el servicio,
pero inicialmente no activó exploración táctil. Se detuvo y reactivó desde la
interfaz de Ajustes; después dumpsys confirmó exploración táctil y manejo del
doble toque activos. No basta comprobar `enabled_accessibility_services`.

Alt+Right produjo marco verde en Volver de Ediciones. Los avances repetidos y
Alt+Shift+Right no acreditaron recorrido completo; Alt+Enter tampoco acreditó
activar el cierre. Tab enfocó Rama 21 con marco verde y foco de teclado;
Enter abrió su ficha y el marco del lector apareció en el cierre. Se acredita
ese recorrido de teclado con TalkBack activo, no navegación íntegra mediante
gestos del lector, locución ni retorno automático a la tarjeta. Toques ADB
activan controles directamente y no se contabilizan como exploración táctil.

Evidencia: [Volver enfocado](evidence/tournament-browsing-2026-10-11/android-release-talkback-back-focus.png),
[tarjeta enfocada](evidence/tournament-browsing-2026-10-11/android-release-talkback-card-focus.png),
[destino](evidence/tournament-browsing-2026-10-11/android-release-talkback-card-open.png).
TalkBack y exploración táctil restaurados apagados antes del smoke.

Retrospectiva: una Release sin Metro descubre diferencias reales de preparación
de red. El overlay optativo permite usar fixtures locales sin cambiar política
de producto. Una aserción de texto independiente falla si la pantalla agrupa el
encabezado: usar expectativas del destino observado, no fragmentos inventados.
Foco del lector y foco de teclado pueden diferir; documentar ambos evita dar
por aprobada una activación accesible a partir de una pulsación normal.

### Caso corregido y regresión de preparación del enlace

La primera versión del caso de desempate enlazaba el ID
`01a12cfb-462a-7009-9189-663f0a1328dd` (Matriz 1–Matriz 2), pero exigía
Matriz 1–Matriz 3. Algunas pantallas mostraban ambos encuentros y pasaban;
otras mostraban solo el destino correcto del enlace y eran rechazadas.
Ese falso fallo se atribuyó provisionalmente al desplazamiento y al modo de
teclado, sin comprobar primero la correspondencia entre ID y equipos.

La revisión del fixture y del resaltado confirmó el error del test. El caso
ahora usa `01a12cfb-462a-7739-bbfc-1a3f4f02f2e7`, el ID real de Matriz 1–Matriz 3.
Esos rechazos iniciales no demuestran un fallo del producto ni un problema
causado por teclado. Se conserva la evidencia del
caso inválido, etiquetada expresamente:
[pantalla](evidence/tournament-browsing-2026-10-11/android-release-invalid-case-tiebreak.png),
[árbol](evidence/tournament-browsing-2026-10-11/android-release-invalid-case-tiebreak.xml).

Con el ID correcto, la segunda repetición completa sí dejó el partido objetivo
fuera de pantalla con `TouchMode: 1`. Esta es evidencia independiente y válida
de la regresión: [pantalla anterior](evidence/tournament-browsing-2026-10-11/android-release-before-ready-tiebreak.png),
[árbol anterior](evidence/tournament-browsing-2026-10-11/android-release-before-ready-tiebreak.xml).

La alineación de listas se invocaba desde eventos de layout antes de que el
efecto de enlace resolviera `relationship` y seleccionara el partido. Podía
marcarse terminada sobre una lista sin controles de administración; al aparecer
los controles, la altura y posición cambiaban sin repetir esa alineación.
`revealLinkedMatch` y `alignLinkedMatch` ahora exigen que `linkHandled` corresponda
al matchId actual. Ese marcador ya se asigna después de cargar torneo y relación.
Se reutiliza la señal existente, sin temporizadores nuevos ni otra store.

El ajuste especulativo de medidas, los logs diagnósticos y la restricción del
smoke a modo táctil se retiraron. La Release final solo incorpora la guarda de
preparación del enlace. Los siete tests de browsing, typecheck, ESLint, formato
de la ruta y exportación web pasan. Checklist cliente: textos, tokens, layout,
semántica accesible, rutas, cierre, permisos, adaptadores OpenAPI y feedback
conservados; solo cambia cuándo puede comenzar la alineación existente.
La comprobación nativa iOS de esta guarda compartida queda pendiente.

Retrospectiva adicional: comprobar en el fixture los equipos del ID solicitado
antes de investigar renderizado. Que otro partido aparezca en la misma pantalla
puede ocultar una expectativa incorrecta. El marco de resaltado observado en
la captura permite contrastar el enlace real con la aserción de contenido.
Además, el layout definitivo depende de la relación de cuenta, no solo del
torneo. Una ruta profunda no debe dar por terminada su alineación antes de que
su propio efecto de enlace confirme que esos datos están listos.

Resultado final: Release recompilada con la guarda de preparación, caso de
desempate corregido y tres repeticiones completas de grupos/desempate/final:
**9 PASS**, todos con arranque COLD. La captura final confirma Matriz 1–Matriz 3
resaltado y controles completos. [Resultados](evidence/tournament-browsing-2026-10-11/android-release-touch-results.txt).
Se aprueban esos enlaces de Release local Android; TalkBack conserva la
validación parcial descrita y la guarda compartida requiere revalidación iOS.

Cierre final: TalkBack apagado (`enabled_accessibility_services=null`,
`touch_exploration_enabled=0`), emulador detenido, `make dev-down` completado
y ADB cerrado mediante kill-server. Sin listeners de API/Metro en 8080/8082;
Android Studio puede reabrir su servidor ADB de gestión. Volúmenes y fixtures
conservados, observabilidad no iniciada. `git diff --check` sin errores.

### Revalidación iOS de la preparación del enlace (2026-10-12)

La Release local con la guarda compartida se compiló e instaló con firma ad hoc
en iPhone 18 Pro, iOS 27.0. API local disponible y Metro apagado. Todos los
enlaces se introdujeron en Safari mediante CUA y se aceptó «Abrir» en la UI.
La primera apertura ocurrió después de reiniciar el simulador; entre las demás
se terminó el proceso de `com.fasttourney.app.local` con éxito antes de abrir
el siguiente enlace. No se inyectaron enlaces ni gestos con simctl.

| Destino | Repeticiones en frío | Resultado observado |
| --- | --- | --- |
| Grupos, Matriz 1–Matriz 4, Grupo 1 · Jornada 3 | 1 | Partido y sus dos controles completos visibles, sin desplazar manualmente. |
| Desempate, ID `01a12cfb-462a-7739-bbfc-1a3f4f02f2e7` | 2 | Matriz 1–Matriz 3 completamente visible con «Fecha, hora y lugar» y «Añadir resultado» en ambas aperturas. |
| Cuadro, Final | 1 | Final seleccionada, Matriz 1–Matriz 3, marcador 2–0 y ganador Matriz 1 visibles. |

**4 PASS visuales mediante CUA**, con los permisos de creador ya cargados.
La guarda compartida queda revalidada en iOS para estos fixtures. Las capturas
y árboles observados constan en la conversación; no se generaron archivos PNG
de esta sesión. El resaltado temporal ya había expirado en las capturas finales:
la aprobación corresponde a selección de fase/ronda y visibilidad del destino,
no a una nueva comprobación del marco temporal de resaltado.

Device Hub quedó inicialmente en «Connecting display» aunque el dispositivo
estaba Booted y la instalación terminaba correctamente. Cambiar de dispositivo
y de vista no lo resolvió; reiniciar únicamente este simulador recuperó la
pantalla. No se atribuye ese bloqueo de la herramienta al producto.

Retrospectiva: la señal de relación resuelta funciona en ambos clientes nativos
para los casos probados. La presencia de elementos en el árbol accesible iOS
no demuestra que estén en pantalla: estas aprobaciones también contrastaron
la captura. Se conservan pendientes el gesto de refresco iOS y la validación
completa de gestos, locución y retorno de foco con lectores de pantalla.

Cierre de sesión: `make dev-down` completado, simulador apagado y sin listeners
en 8080, 8082 ni 8098. Volúmenes y fixtures conservados; observabilidad no
iniciada. `git diff --check` sin errores.

### Revisión ampliada de lectores solicitada por el usuario (2026-10-12)

Se revisan consecutivamente las Release locales, primero iOS y después Android,
con las sesiones ficticias existentes. No se modifica código en este bloque.

VoiceOver activo en iOS: Control+Option+Right mueve el marco desde Pendientes a
Resultados, primera tarjeta Página 1–Página 2 y segunda tarjeta Página 1–Página 3,
en ese orden. Doble toque sobre la segunda tarjeta abre QA matriz · Paginación.
La ficha queda en torno a Jornada 7, con el marco del lector encima de contenido
recortado; el partido solicitado (Jornada 22) no está visible. No se ha aislado
si la causa es el lector o el recorrido de alineación de esta liga grande.
El cierre recibe foco, pero doble toque y Control+Option+Space no cierran la
ficha desde Device Hub; una observación devuelve AXError.cannotComplete.
La activación del cierre y retorno del foco iOS siguen sin certificar por esta
limitación de interacción. VoiceOver restaurado apagado y confirmado en UI.

TalkBack se activa desde Ajustes y se acepta su permiso esperado en el emulador
de QA. dumpsys confirma exploración táctil, manejo del doble toque y gestos de
varios dedos habilitados. No se usa UIAutomator durante el recorrido activo.
Alt+Right enfoca el título, pero repeticiones y Alt+Shift+Right no acreditan
avance. Tab recorre Pendientes y Resultados; el siguiente paso salta a Inicio.
No se equipara ese recorrido de teclado con el orden del lector.

Tras un arranque COLD de Mis partidos, dos pulsaciones de flecha abajo enfocan
Pendientes y la primera tarjeta Página 1–Página 2. Enter abre su ficha con el
marco en el cierre. La ficha muestra Jornada 7, sin el partido de Jornada 23
en pantalla. Enter sobre el cierre no acredita activación. Atrás del sistema
sí vuelve a Mis partidos, pero el marco reaparece en Pendientes, no en la tarjeta.
Se acredita apertura por teclado y retorno mediante Atrás, **no restauración
del foco a la tarjeta ni aprobación completa de TalkBack**.

Evidencia Android: [tarjeta enfocada](evidence/tournament-browsing-2026-10-11/android-reader-card-focus.png),
[liga abierta](evidence/tournament-browsing-2026-10-11/android-reader-large-league-open.png),
[retorno en Pendientes](evidence/tournament-browsing-2026-10-11/android-reader-return-pending-focus.png).
La evidencia visual iOS consta en las capturas de la conversación.

Pendientes concretos descubiertos: investigar el destino fuera de pantalla en
la liga de 24 equipos con lector activo, restaurar/verificar el foco de la tarjeta
al regresar y comprobar el cierre con gestos del lector en una superficie que
los transmita de forma fiable. Las herramientas disponibles no proporcionan
audio del dispositivo: la locución no se ha escuchado ni aprobado.

Retrospectiva: separar foco accesible de foco de teclado y esperar la pantalla
estable antes de capturar. Una captura inmediata tras una pulsación puede
mostrar aún el destino anterior. Los casos pequeños que pasaron enlaces en frío
no cubren por sí solos la liga de 24 equipos ni la interacción con lectores.

Cierre: VoiceOver apagado en UI; TalkBack restaurado a servicios habilitados
`null` y exploración táctil `0`. Simuladores detenidos, `make dev-down` completado
y ADB cerrado. Sin listeners API/Metro; volúmenes y evidencia conservados,
sin observabilidad. `git diff --check` correcto.

### Liga grande sin lectores: diagnóstico y corrección (2026-10-12)

El fallo se reproduce sin TalkBack: Release local Android, arranque COLD, enlace
a Página 1–Página 2 en Jornada 23 de la liga de 24 equipos/276 encuentros.
Queda en Jornada 9 y el smoke falla. [Captura anterior](evidence/tournament-browsing-2026-10-11/android-large-league-before.png).
No es un fallo exclusivo del lector.

Diagnóstico temporal de Release: el destino usa el índice plano 308, pero los
20 intentos se agotan cuando el mayor índice medido alcanza 115. Se programan
varios intentos con la misma ventana medida. Reiniciar el presupuesto al avanzar
no basta por sí solo: quedan comandos de scroll simultáneos. Los logs diagnósticos
se retiraron del código final; no se añadió observabilidad.

Corrección mínima en la ficha, compatible con ADR-0161: agrupar solicitudes
mediante una sola operación pendiente por fotograma; renovar el presupuesto
cuando avanza highestMeasuredFrameIndex; iniciar la revelación al preparar el
partido destacado. Este último paso se detecta en iOS: si el layout llega antes
de la preparación del enlace, la lista puede quedar en Jornada 1 sin otro evento
que inicie el recorrido. Se conservan alturas dinámicas, virtualización, scroll
sin animación y límite ante ausencia de progreso. No se fijan alturas de tarjetas
ni se añade getItemLayout ficticio o dependencia.

Una Release Android intermedia pasa los cinco enlaces (J22/J23, grupos, desempate
y final), antes de añadir el inicio explícito descubierto en iOS. Esa evidencia
no aprueba el binario final. Una apertura anterior devolvió timeout/LaunchState
UNKNOWN y se descartó; también se conservan como fallos las pruebas del ajuste
parcial, sin atribuirlas a éxito.

Release iOS con inicio explícito, firma ad hoc, VoiceOver apagado y Metro apagado: **2 PASS
visuales**, Jornada 23 Página 1–Página 2 y Jornada 22 Página 1–Página 3. Cada
enlace se introduce desde Safari mediante CUA; app terminada entre ambos.
Ambas tarjetas y sus controles «Fecha, hora y lugar»/«Añadir resultado» completos
en pantalla, sin desplazamiento manual. Capturas en la conversación. Se acredita
la visibilidad del destino; no se vuelve a certificar el resaltado temporal.

El smoke opt-in admite --timeout (35 s predeterminado, máximo 120 s). Esta matriz
de liga grande usa 60 s por caso, sin cambiar el valor predeterminado ni fingir
una garantía de latencia. [Casos reproducibles](evidence/tournament-browsing-2026-10-11/android-large-league-cases.json).
El script comprueba Release no depurable, lector apagado, arranque COLD y contenido
con intersección de pantalla; guarda launch/XML/PNG. No certifica gestos o locución.

Preflight y cierre de cliente: se revisan ADR-0054/55/56/57 y documentos de diseño,
arquitectura/desarrollo, más ADR-0161. No cambia UI, copy, catálogos, tokens,
permisos, navegación o contrato HTTP. La modificación toca únicamente el
recorrido de revelación; las operaciones conservan adaptador generado y apiFetch.
Typecheck y ESLint de ruta pasan; las siete pruebas de browsing pasan.

Retrospectiva: una guarda de preparación debe disponer de un disparador posterior;
contar callbacks no equivale a contar intentos sin progreso. Validar una lista
virtualizada larga con permisos cargados y capturas del viewport, separadamente
de grupos pequeños y de lectores. Siguen pendientes el retorno de foco a la
tarjeta, la activación fiable del cierre y la QA completa de lectores/paginación.

La regresión Android del inicio explícito arroja 4 PASS y 1 FAIL (desempate):
la alineación se puede iniciar antes de disponer de dimensiones nativas válidas
de la tarjeta. La guarda final exige ancho y alto no nulos antes de calcular
y confirmar su offset; este binario requiere repetir la matriz. No se conserva
una aprobación de una build anterior como aprobación del código final.

Descartar medidas vacías no bastó por sí solo: la repetición todavía dio 4 PASS
y 1 FAIL. La condición final no confirma alineación por convergencia del offset
únicamente; exige además que la tarjeta esté en el viewport (hasta la altura
visible disponible si una tarjeta lo supera). Un tamaño de contenido anterior
puede limitar el offset sin que el destino esté revelado.

Android final con esta condición: **5 PASS COLD** (desempate primero, J23, J22,
grupos y final). La captura del desempate confirma Matriz 1–Matriz 3 resaltado
y sus dos controles completos. [Resultados finales](evidence/tournament-browsing-2026-10-11/android-large-league-results.txt)
y [evidencia final](evidence/tournament-browsing-2026-10-11/android-large-league-final/tiebreak.png).
Typecheck, lint de ruta, formato y exportación web final pasan.

Revalidación del binario iOS final con la condición de viewport: **2 PASS COLD**,
Jornada 23 Página 1–Página 2 y desempate Matriz 1–Matriz 3, con ambos controles
completos en cada destino. Build final BUILD SUCCEEDED, instalada tras completar
la compilación, esquema introducido en Safari mediante CUA y app terminada
entre recorridos. VoiceOver apagado confirmado en inspector. Las capturas
constan en la conversación. El bloqueo temporal del Mac se resolvió tras el
desbloqueo del usuario y se completó el desempate en esa misma build final.
Esta revalidación acredita visibilidad; no amplía la QA de lectores ni certifica
el resaltado temporal en iOS.

La matriz automatizada Android final conserva cinco PASS. Comprobaciones finales
de código: typecheck y ESLint de la ruta, formato de ruta/casos, sintaxis Python
y exportación web sin errores. git diff --check correcto. No se ha modificado
ningún endpoint ni el contrato OpenAPI durante este ajuste.

Cierre de sesión: make dev-down completado, sin listeners en 8080/8082/8098;
simulador iOS apagado y servidor ADB cerrado. Volúmenes y evidencia conservados.
No se activó observabilidad ni se intervino producción.

### Retorno de foco y cierre accesible (2026-10-12)

Problema confirmado antes del ajuste: regresar desde una tarjeta no restauraba
su foco accesible. Se conserva únicamente el identificador de origen en la
pantalla de Mis partidos, con referencias a las tarjetas montadas. Al reaparecer
la ruta raíz tras la transición nativa, si Mis partidos sigue enfocada y el
lector está activo, se envía el evento de foco a la tarjeta de origen. Cambiar
cuenta o segmento descarta el origen; una tarjeta ausente no se sustituye por
otra. Sin lector no se fuerza foco. No hay persistencia, estado global ni nueva
dependencia. Se reutiliza AccessibilityInfo.sendAccessibilityEvent de RN 0.86.3.
La X cargada usa common_close, igual que durante carga, en ambas cabeceras.

Alternativa descartada: gestor global/persistido de foco. Su mayor coste de
mantenimiento no aporta valor a este recorrido; el estado local basta dentro
del alcance accesible aceptado en ADR-0161 y autorizado por el usuario.

Release iOS final, VoiceOver activo mediante inspector de Device Hub:
**2 PASS de cierre y retorno de foco**, desde Página 1–Página 3 (J22) y después
Página 1–Página 2 (J23). Toque enfoca la tarjeta, doble toque la abre; toque en
la X y doble toque cierran. Tras estabilizar la transición, las capturas muestran
el marco de VoiceOver en la tarjeta de origen, no en Pendientes. No se interviene
manualmente en el foco de retorno. Capturas en la conversación; VoiceOver
restaurado apagado y simulador iOS detenido antes de pasar a Android.

**FAIL separado con lector activo:** las mismas aperturas dejan el destino
lejano fuera del viewport (J22/J23), pese a los PASS en frío sin lector de la
fase anterior. No se aprueba la revelación con VoiceOver ni la locución. La
causa requiere aislar la presentación de la ruta, la medición y el foco del
lector; estos PASS de cierre/retorno no corrigen ni ocultan ese fallo.

Android Release final, Pixel API 34, TalkBack 14.2: **PASS retorno a segunda
tarjeta tras Atrás** y **PASS retorno a primera tarjeta tras salir de su ficha**.
Se confirma servicio enlazado, exploración táctil y manejo de doble toque.
Durante el recorrido activo no se usa UIAutomator. Capturas con marco verde:
[segunda tarjeta](evidence/tournament-browsing-2026-10-11/android-focus-return-back.png)
y [primera tarjeta](evidence/tournament-browsing-2026-10-11/android-focus-return-first.png).
Las flechas/Enter pueden abrir mediante foco de teclado; no se equiparan al orden
completo del lector. La prueba de Enter con marco accesible en la X abrió Fecha,
hora y lugar: el foco de teclado y el foco del lector eran distintos. Se salió
sin guardar ni modificar datos. Alt+Enter mediante input keycombination no
activó la tarjeta; no se acredita el atajo ni doble toque de cierre Android.
El toque inyectado en la X sí cierra con el servicio activo, pero tampoco prueba
el gesto del lector. La superficie Running Devices de Android Studio no expuso
el emulador iniciado por CLI. No se atribuye esta limitación a un fallo probado
del handler de cierre, que conserva el patrón compartido.

Revelación Android con TalkBack: una apertura de J22 muestra ambos controles
del destino, pero otras dejan la tarjeta recortada. **No se aprueba estabilidad
del destino con lector**. Se conserva la captura de fallo y la de destino visible
por separado: [recortado](evidence/tournament-browsing-2026-10-11/android-focus-open-settled.png),
[visible](evidence/tournament-browsing-2026-10-11/android-focus-open-again.png).

Preflight: manifiesto DOCX/MD, README, docs/README, gate técnico cerrado,
ADR-0054/55/56/57/0161/0162 y diseño/arquitectura/desarrollo revisados. Checklist:
catálogos existentes common_close en es/en/it/fr; tokens, objetivo táctil,
Toolbar iOS y cabecera compartida sin cambiar; no nuevos textos visibles,
permisos, dependencias ni estilos. HTTP/OpenAPI intactos. Typecheck, ESLint de
ambos archivos, formato y exportación web pasan con el código final. Ambas
Release compilan y se instalan después de completar la compilación.

Retrospectiva: restaurar en la aparición nativa de la pantalla de origen evita
que el reset del modal sustituya ese foco. El marco del lector y el foco de
teclado deben verificarse separadamente: Enter puede activar otro control.
Un PASS de cierre/retorno no acredita revelación, gesto completo ni audio.
Pendientes: aislar la revelación con lectores activos, gesto de cierre TalkBack,
locución y revisión de lectores sobre paginación/errores.

Referencias primarias contrastadas con las versiones instaladas:
[evento de foco de React Native](https://reactnative.dev/docs/0.81/accessibilityinfo#sendaccessibilityevent)
y [atajos y keymaps de TalkBack](https://support.google.com/accessibility/android/answer/6110948?hl=en-EN).
No se asume que el keymap mejorado de versiones posteriores exista en TalkBack 14.2.

Cierre: VoiceOver apagado confirmado en UI; TalkBack vuelve a servicios null y
exploración táctil 0. Simuladores y ADB detenidos; make dev-down completado y sin
listeners en 8080/8082/8098. Evidencia y volúmenes conservados, sin observabilidad.


### Diagnóstico adicional de revelación con VoiceOver (2026-10-12)

Se reabre la sesión local de QA iOS, sin observabilidad. Las pruebas comparan
la misma Release y la misma tarjeta de Mis partidos: Jornada 23 aparece completa
sin VoiceOver, pero con el lector vuelve al inicio. La instrumentación temporal
se consulta en el registro unificado del simulador, subsistema
`com.facebook.react.log`; la salida de `simctl launch --console` no contenía
estos mensajes. Solo se registran flags, índices y medidas, sin datos de usuario.
[Evidencia numérica](evidence/tournament-browsing-2026-10-11/ios-reader-reveal-diagnostic-2026-10-12.log).

Hecho observado: en una apertura fallida el índice solicitado es 308; el máximo
medido progresa hasta 196. El scroll pasa por 33.502 px, el contenido cambia de
34.248 a 21.523,67 px y el offset vuelve a cero sin montar el destino. No se
observa una alineación completada. La causa de ese retorno no queda demostrada.

Se prueban y descartan por falta de estabilidad: esperar solo a `transitionEnd`
(J22 pasa dos aperturas, J23 falla), enviar foco al resumen tras alinearlo
(J23 falla), reanudar en cada scroll (J23 pasa, J22 falla) y coordinar un único
reintento de 120 ms también después de un scroll sin error (J22 falla). Ninguna
de esas variantes se conserva en el código. Sus PASS parciales no cierran el
fallo. Los logs temporales también se retiran.

En React Native 0.86.3 instalado, `VirtualizedSectionList.scrollToLocation` suma
`itemIndex` y las secciones anteriores; `_subExtractor` reserva el índice cero
a la cabecera y resta uno antes de resolver la primera fila. Se contrasta
`itemIndex + 1` como candidato aislado, sin alterar presupuestos ni agregar
temporizadores. Resultado provisional: J22 visible con ambos controles y
VoiceOver activo; doble toque de cierre y retorno de foco a la segunda tarjeta
confirmados. J23 pendiente: el Mac se bloquea después de seleccionar la primera
tarjeta y justo al activarla. También quedan pendientes las regresiones en frío
y Android del candidato. No se declara corregido el fallo todavía.

El controlador de UI devuelve en ocasiones `AXError.cannotComplete` sobre la
ficha larga. Se comprueba Mac desbloqueado y respuesta de controles del
simulador, se intenta ventana compacta y se reinicia solo la app. No se
equipara ese error del controlador con un fallo demostrado del cierre de la
app; tampoco se aprueba una interacción sin observar su resultado.


Cierre de la comprobación iOS del candidato de índice: **5 PASS visuales** con
la Release compilada en `tm-reader-row-index-build.log`: J22 desde Mis partidos
con VoiceOver, J23 desde Mis partidos con VoiceOver, J23 en frío sin lector,
J23 en frío con lector ya activo antes de abrir desde Safari y desempate
Matriz 1–Matriz 3 en frío sin lector. Todos muestran completos ambos controles
del destino. La apertura de J23 interrumpida por bloqueo se ejecuta después
de desbloquear, desde la primera tarjeta enfocada; no se cuenta el intento
bloqueado como prueba. J22 confirma además doble toque de cierre y marco de
retorno sobre la segunda tarjeta. J23 conserva limitación del controlador al
intentar cerrar la ficha larga. Capturas visibles en la conversación.

El cambio funcional frente a la fase previa es solo `itemIndex + 1` en
`scrollToLocation`; no queda lógica nueva de foco ni temporizadores. Typecheck,
ESLint, formato y exportación web pasan. No cambia contrato, endpoints, copy,
tokens ni estilo. VoiceOver restaurado apagado y simulador iOS detenido antes
de comenzar los recorridos Android. No se aprueba audio, foco del lector en
el partido, orden completo del lector ni estabilidad fuera de estos casos.


Android del candidato universal `itemIndex + 1`: **4 PASS / 1 FAIL**, todos
COLD, con lectores apagados. J23, grupos, desempate y final pasan; J22 se queda
en Jornada 10 y agota el timeout de 60 s.
[Evidencia del candidato universal](evidence/tournament-browsing-2026-10-11/android-row-index-final/).
Ese resultado impide entregarlo como ajuste universal. Se limita el desplazamiento
de índice a iOS y se conserva el índice anterior en Android/web. Nueva regresión
Android pendiente; no se oculta ni convierte el fallo intermedio en un PASS.


Código final: `itemIndex + (Platform.OS === "ios" ? 1 : 0)`. La rama iOS
conserva el mismo índice probado en las cinco aperturas; Android/web vuelven
a su índice anterior. Las dos Release compilan, y typecheck, ESLint, formato
y exportación web pasan con la condición final. La Release iOS con condición
compila sin reactivar su simulador durante la fase Android; los cinco recorridos
iOS corresponden al candidato de índice cuyo valor en iOS es idéntico.

Android Release con condición final: **5 PASS COLD sin lector**: J23, J22,
grupos, desempate y final.
[Resultados](evidence/tournament-browsing-2026-10-11/android-ios-only-final/results.txt)
y [evidencia](evidence/tournament-browsing-2026-10-11/android-ios-only-final/).
No se atribuyen estos resultados a la variante universal, cuyo FAIL se conserva.

TalkBack activo y exploración táctil comprobada: **2 PASS visuales COLD**
mediante intents locales de J22/J23. Ambos controles de cada destino se ven
completos. [J22](evidence/tournament-browsing-2026-10-11/android-ios-only-talkback-j22.png),
[J23](evidence/tournament-browsing-2026-10-11/android-ios-only-talkback-j23.png).
El foco verde queda en Cerrar: no se aprueba foco del lector sobre el partido,
locución ni gesto de apertura a partir de estos intents. El recorrido cálido
desde Mis partidos sigue pendiente de esta nueva sesión.


TalkBack desde Mis partidos: segunda tarjeta activada mediante `input tap`
con servicio activo, **PASS visual J22**, seguida de **PASS retorno con Atrás**
(marco verde sobre la segunda tarjeta).
[J22 cálido](evidence/tournament-browsing-2026-10-11/android-ios-only-talkback-warm-j22.png),
[retorno](evidence/tournament-browsing-2026-10-11/android-ios-only-talkback-return-j22.png).
La inyección de un toque activa el handler y no se equipara al doble toque
del lector.

Primera tarjeta J23: **PASS visual desde Mis partidos**, con el encuentro y
sus dos controles completos; **PASS retorno con Atrás**, con el marco verde
sobre la primera tarjeta de origen.
[J23 cálido](evidence/tournament-browsing-2026-10-11/android-ios-only-talkback-warm-j23.png),
[retorno](evidence/tournament-browsing-2026-10-11/android-ios-only-talkback-return-j23.png).
Estos dos recorridos cálidos complementan los dos COLD con TalkBack; siguen
sin certificar doble toque, locución ni foco del lector sobre el destino.

Cierre de sesión: TalkBack restaurado apagado (`null`, `0`), emulador Android
cerrado y servidor ADB detenido. VoiceOver y simulador iOS ya estaban apagados.
`make dev-down` termina correctamente y no quedan listeners en 8080, 8082 ni
8098. Se conservan volúmenes y evidencias. No se activa observabilidad.

Retrospectiva técnica: el ajuste mínimo cambia únicamente el índice de iOS.
Las variantes de foco y temporizadores se descartan porque sus PASS parciales
no resolvían los dos destinos. Probar el índice en Android detecta una regresión
y evita ampliar el alcance del cambio. Las verificaciones estáticas, exportación
web y compilaciones Release pasan; los cinco enlaces Android automatizados
pasan con el código final. La evidencia visual con lectores complementa esa
automatización, pero no demuestra audio ni gestos reales. Esas comprobaciones,
el orden completo de lectura y los casos pendientes de la matriz permanecen
abiertos; no se declara cerrado todo el QA de accesibilidad ni de las features.


### Candidato de publicación v1.13.0 — 2026-10-12

El usuario autoriza publicar las features y fixes pendientes con el proceso
completo. El alcance incluye ADR-0161/0162, API, renderer y web; las builds
nativas de QA no equivalen a distribución en tiendas. Los pendientes de gesto
de refresco iOS, locución y orden completo de lectores se conservan explícitos.
No hay migración nueva: las consultas usan el esquema 00022 ya productivo.
Antes de activar se exige make verify y CI, promoción a main con tag anotado,
validación dev, backup/restauración y artefactos del mismo SHA con rollback.
