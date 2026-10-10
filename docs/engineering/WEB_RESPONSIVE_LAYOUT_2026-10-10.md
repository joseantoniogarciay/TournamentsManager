# Ancho y columnas web — 2026-10-10

## Alcance y decisión

El usuario aplaza la distribución móvil y pide limitar el ancho desktop,
adaptarlo al viewport y evaluar layouts. Acepta después columnas en Inicio y
biblioteca. [ADR-0154](../adr/0154-constrain-responsive-web-content-width.md)
registra alternativas y decisión. Marco 1200 px, lectura 720 px y columnas a
960 px CSS mediante tokens compartidos. Ficha y clasificación conservan su
organización; las celdas mantienen ADR-0153.

## Evidencia acotada

| Muestra | Resultado observado |
| --- | --- |
| Inicio invitado y autenticado a 1600 px | Marco centrado, creación a la izquierda y onboarding/recientes a la derecha; sugerencias en fila completa |
| Biblioteca autenticada, Administro 45 y Sigo 5 | Dos columnas, filtros funcionan y última tarjeta impar conserva su ancho; cards 570 px a x=220 y x=810, separación 20 px |
| Inicio a 959 y 960 px | Una columna a 959; dos a 960; scrollWidth coincide con viewport |
| Inicio autenticado a 320 px | Una columna, texto refluye, sin scroll horizontal de página; botonera visible |
| Biblioteca a 390 px | Una columna, filtro y acción flotante visibles; sin scroll horizontal de página |
| Crear torneo a 1600 y 390 px | Card centrada en lectura; deportes refluyen y campos/cierre permanecen disponibles; no se envía |
| Ajustes claro/oscuro, 1600 y 320 px | Fondo completo, contenido de lectura y cierre funcional; preferencia Sistema restaurada |
| Marco/botonera a 1600 px | Botonera 1200 px a x=200; sin overflow horizontal de página |

La revisión de código confirma que nativo no recibe el marco web y conserva
una columna. Los márgenes de Card no cambian. La biblioteca conserva filtros,
paginación y orden de datos; CSS cambia únicamente la distribución. No se
repite la matriz de lectores ni las mutaciones de negocio. No se certifica
zoom real del navegador, locución ni todas las rutas: la media query responde
al ancho CSS y la muestra de reflow verifica sus límites.

Capturas de sesión: `/tmp/tm-web-responsive-home.jpg`,
`/tmp/tm-web-responsive-library.jpg` y `/tmp/tm-web-responsive-desktop.jpg`.
Son evidencia temporal, no assets del producto.

## Validación técnica y cierre

Prettier, ESLint de archivos afectados, `pnpm run typecheck` y exportación web
(36 rutas estáticas). No cambian contratos, infraestructura ni dependencias.
Metro y API/PostgreSQL/Mailpit se usan solo durante esta muestra y se apagan
al cierre; no se activa observabilidad y se conservan volúmenes.

## Retrospectiva técnica

El marco común resuelve también cabeceras y tabs, que limitar solo contenido
habría dejado dispersas. Limitar el hijo de Screen conserva canvas completo:
la primera muestra detectó y corrigió bandas negras del navegador en oscuro.
La muestra con sesión recuperada detectó que TipCard devuelve null; evitar
reservarle una celda mantiene útil la segunda fila. CSS mantiene el mismo DOM
al cambiar ancho y evita trasladar breakpoints a cada feature. La comprobación
por patrones y unos tamaños frontera basta para este cambio reversible.

## Ampliación aceptada: equipos y ficha de liga

El usuario acepta las dos oportunidades tras comparar su utilidad y coste.
Gestión e invitación comparten lateral con un listado de equipos; sin alta o
invitación no se reserva una columna vacía. La variante sidebar reutiliza la
primitiva de columnas y mantiene los márgenes; CSS fija el panel al desplazar
contenido. Nativo mantiene una columna y los diálogos originales.

La ficha de liga iniciada sitúa resumen/accesos en un lateral de 360 px;
partidos conservan su SectionList y su referencia de viewport, ahora medida
solo en la columna principal. El lateral tiene desplazamiento propio para
que sus controles sigan alcanzables. Al estrechar la ventana, el resumen
regresa al encabezado sin trasladar ni remontar la SectionList. El selector de
fase permanece junto a los partidos. Preparación, desempates y cuadros no
reciben este lateral. La gestión de equipos compartida también se adapta
cuando aparece dentro de la ficha en preparación.

Muestra ejecutada con fixtures existentes, sin escrituras de negocio:

- QA bajas equipos web iOS: ficha en preparación y ruta Equipos a 1600 px,
  controles a la izquierda, lista a la derecha; a 320 px se apilan, sin overflow.
  Apertura/cierre del diálogo Añadir equipo, sin escribir ni guardar.
- QA web football league: ficha finalizada a 1600 px, lateral compacto y
  jornadas en la columna principal; al cambiar a 959 px el resumen regresa
  al encabezado; a 320 px todo refluye y scrollWidth coincide con viewport.
- Lista de equipos de consulta: mantiene contenido centrado a 720 px.
- QA formulario football: cuadro final sigue usando el marco amplio.

Capturas: `/tmp/tm-web-team-sidebar.jpg` y `/tmp/tm-web-league-sidebar.jpg`.
Typecheck, ESLint y exportación web completan la muestra. Revisión del diff
confirma que solicitudes, permisos, validación y textos localizados conservan
su implementación. No se certifican scroll sticky con listas largas ni lector
por esta muestra; no se repite la matriz anterior ni se crean fixtures.

Retrospectiva: ScrollView añade flexGrow por defecto; establecer solo width
no reserva un lateral estrecho. La muestra detectó reparto de ancho excesivo
y se corrigió flexGrow a cero antes del cierre. Mantener la lista de partidos
en el mismo padre evita reiniciar su instancia al alternar el lateral.
