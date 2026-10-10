# ADR-0154: Limitar el ancho del contenido web de forma responsive

- **Estado:** Aceptado
- **Fecha:** 2026-10-10
- **Decisor:** Usuario
- **Supera a:** Ninguno; concreta la presentación web de ADR-0054 y ADR-0055.

## Problema y evidencia

El usuario solicita que desktop no expanda todo de extremo a extremo, limitar
el ancho máximo, adaptar el contenido y evaluar oportunidades de layout.
Screen usa flex:1 sin límite y Card solo aporta 20 px laterales. Las cabeceras
y tabs pertenecen al navegador, fuera de Screen; limitar solo las cards dejaría
controles de navegación separados del contenido. La botonera web tiene además
una regla fixed respecto al viewport.

## Alternativas y coste

- Un único ancho estrecho para todo: bajo mantenimiento, pero desaprovecha
  clasificación, listas y cuadros.
- Marco web común y dos anchos de contenido: bajo mantenimiento; separa
  lectura/formularios de datos extensos y conserva navegación alineada.
- Rediseño completo con sidebar y rejillas por ruta: coste alto y más cambios
  de navegación. No es necesario para resolver el ancho solicitado.

## Decisión del usuario

**Aceptada mediante la petición del 2026-10-10:** limitar el ancho máximo web
con adaptación responsive y evaluar posibles mejoras al disponer de espacio.
Se implementa el marco común dentro del stack existente, sin dependencia nueva.
Como valores iniciales de implementación, tokens web de 1200 px para el marco
y 720 px para lectura/formularios. Inicio, biblioteca, ficha y clasificación
usan el ancho amplio; las otras rutas conservan una columna de lectura.
Ambos anchos son máximos, nunca mínimos: por debajo se usa el ancho disponible.
El fondo ocupa el viewport y el marco se centra; conserva los 20 px existentes
sin sumar padding horizontal a Screen ni Card. Cabeceras y tabs comparten marco.

El usuario acepta después explícitamente «Sí, añadir columnas en Inicio y
biblioteca». Inicio coloca creación y torneos recientes (o explicación de
invitado) en dos columnas; sugerencias conserva una segunda fila completa. TipCard no tiene contenido
actual y no reserva una columna vacía.
La biblioteca distribuye sus tarjetas por filas de dos; filtros y paginación
conservan su posición. Una media query a 960 px vuelve a una columna al reducir
el viewport CSS, también con zoom. El DOM conserva el mismo orden y estructura
en ambos tamaños. La primitiva compensa medio margen de Card en cada celda
para mantener 20 px exteriores y entre columnas, sin cambiar Card. La ampliación aceptada después mediante «Vamos a por ellas» coloca gestión e
invitación junto al listado de equipos cuando hay acciones disponibles; sin
ellas mantiene una lista de lectura. En la ficha con partidos de liga, resumen
y accesos pasan a una columna lateral desplazable y los partidos conservan su
lista. Preparación, desempates y cuadro eliminatorio conservan su organización.
Una ventana estrecha devuelve el resumen al encabezado sin remontar la lista.
No cambia la excepción de texto de clasificación de ADR-0153. Móvil nativo conserva su layout.

## Validación

Typecheck y exportación web, más muestra visual acotada de Inicio, formulario,
biblioteca y datos amplios en desktop y viewport estrecho. Comprobar zoom/reflow,
controles de cierre, acceso a botonera, fondo claro/oscuro y ausencia de overflow
horizontal de página. Los scrolls internos de tabla conservan su función.
No se envían formularios ni se cambian fixtures para validar layout.

## Consecuencias

Tokens compartidos, una propiedad de Screen y una primitiva de columnas evitan medidas por ruta. Una pantalla
nueva elige lectura o datos amplios; no introduce anchos arbitrarios. El límite
es ajustable si la evidencia visual muestra falta de espacio. La distribución
en columnas no se deduce de la anchura: responde al contenido y mantiene orden
de lectura. No se inicia distribución móvil ni se contratan cuentas.

## Documentos afectados y revisión

Design System, checklist del cliente, registro de decisiones y aprendizaje.
La [muestra de cierre](../engineering/WEB_RESPONSIVE_LAYOUT_2026-10-10.md)
distingue ejecución y revisión de código. Revisar estos valores ante un
desbordamiento reproducible, un nuevo layout amplio o un cambio de navegación.
