# ADR-0183: Completar las animaciones contextuales de torneos

- **Estado:** Aceptado; implementado; publicación y QA visual pendientes
- **Fecha:** 2026-11-02
- **Decisor:** Usuario: «mete todas las otras animaciones de las otras propuestas y hacemos nueva subida».
- **Supera a:** ADR-0182 en la limitación inicial a dos ilustraciones.
- **Complementado por:** ADR-0184, loader de carga en bucle añadido a la misma subida.

## Problema y evidencia

El usuario acepta las otras propuestas de la conversación: silbato al iniciar,
camiseta al aceptar una invitación e identidad animada en la bienvenida. Incluye
el cuadro de Sigo y la corrección de inicio web pendientes de publicar. El gate
técnico está cerrado y se mantienen Grada, primitivas y navegación vigentes.

## Alternativas, coste y recomendación

Nuevos diálogos o retardos para cada animación aumentarían interrupciones y
mantenimiento. Otra biblioteca duplicaría el reproductor y su ciclo de vida.
Se recomienda y acepta reutilizar los recursos vectoriales Lottie y mostrar los
nuevos momentos dentro del contenido existente, sin bloquear acciones.

## Decisión y alcance

- Silbato corto en el resumen de la ficha solo tras respuesta confirmada de
  iniciar el torneo. No al visitar un torneo que ya está en curso.
- Camiseta que se integra en un grupo en la ficha de destino tras aceptar una
  invitación. Se conserva la navegación inmediata. Una marca efímera en memoria,
  acotada a torneo y cuenta, transmite el éxito confirmado y se consume una vez;
  no se usa una bandera de URL que pueda inventar un resultado. No persiste ni
  concede permisos. Se conserva el texto/dato real de la ficha, sin banners nuevos.
- Bienvenida sin sesión: balón que recorre un cuadro y da paso a la marca
  vectorial (copa, conexiones y avance). Una reproducción por sesión JS, sin
  repetirla en cada navegación. No sustituye splash ni introduce espera inicial.
- Las cinco ilustraciones comparten colores semánticos, geometría vectorial
  local, dibujo final estático, reducción de movimiento y parada al perder foco.
  No hay sonido, flashes, bucles, fuentes, imágenes o solicitudes externas nuevas.

## Validación y publicación

Pruebas del orden DOMLoaded/preferencia/foco, consumo de invitación por cuenta y
torneo y ausencia de repetición. Render de fotogramas claros/oscuros de los nuevos
recursos. Checklist cliente, typecheck, lint, exportaciones y gate `make verify`.
Se revisan las respuestas de los adaptadores ya existentes; no se modifican
contrato, reglas HTTP, backend ni observabilidad de endpoints. Se conserva el
feedback seguro y las cancelaciones existentes.

Publicación exclusivamente web mediante ADR-0119; sin distribución móvil ni
activación local/dev u observabilidad. La evidencia distinguirá pruebas y QA
visual real, que está pendiente por los límites de las superficies disponibles.
Se documentarán alcance, resultado, rollback y retrospectiva junto al cambio.

## Evidencia de implementación

`make verify` local completo correcto, incluidas seis regresiones del ciclo de
vida y de invitación; integración PostgreSQL genérica omitida sin DB configurada.
Exportaciones web/Hermes iOS/Android correctas. Lámina y secuencia de los tres
recursos nuevos renderizadas y revisadas en claro/oscuro. Véase
[evidencia y límites](../engineering/LOTTIE_MOMENTS_2026-11-02.md).
