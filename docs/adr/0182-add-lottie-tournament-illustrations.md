# ADR-0182: Ilustraciones Lottie para el cierre y la biblioteca vacía

- **Estado:** Aceptado; implementado; QA visual de la app pendiente
- **Fecha:** 2026-11-02
- **Decisor:** Usuario: «Vamos con tu recomendación».
- **Supera a:** ADR-0054 solo en la exclusión inicial de animaciones complejas.

## Problema y contexto

El usuario solicita animaciones Lottie para dar personalidad a FastTourney y
acepta comenzar con un trofeo al finalizar y una ilustración de biblioteca vacía.
El gate técnico está cerrado. Grada (ADR-0176), las primitivas existentes,
la navegación nativa y los estados reales del producto siguen vigentes.

## Alternativas y coste

- Mantener los símbolos estáticos: coste nulo, sin el movimiento solicitado.
- Dibujar y animar cada elemento con primitivas React Native: evita un reproductor,
  pero exige mantener por separado el dibujo y su secuencia entre plataformas.
- Lottie para las dos ilustraciones: añade dependencia nativa y QA de tres
  plataformas; concentra el dibujo y los fotogramas en un recurso vectorial común.
  Las transiciones de controles siguen utilizando las herramientas existentes.

## Recomendación y decisión del usuario

La recomendación aceptada es la tercera alternativa, limitada a dos ilustraciones.
El trofeo sustituye el símbolo del diálogo existente tras una finalización
confirmada. No se reproduce por abrir un torneo ya terminado ni tras un fallo.
El cuadro acompaña al vacío real de Administro; Sigo conserva su explicación.
No se introducen pantallas, banners de éxito ni bloqueos de navegación.

La implementación usará la versión de `lottie-react-native` que resuelva
`expo install`, conforme al SDK fijado. En web se prioriza un reproductor SVG
local sin recursos CDN, WASM remoto ni expresiones de After Effects; el
adaptador de plataforma permite usar `lottie-web` light sin afectar al nativo.
Los recursos serán originales, vectoriales y sin texto, imágenes ni fuentes
externas; sus colores proceden de los tokens del tema.

## Restricciones y consecuencias

Una reproducción breve, sin bucle, sonido, flashes ni confeti de pantalla
completa. Se conserva el dibujo final. Movimiento reducido muestra ese mismo
estado estático, también mientras se resuelve la preferencia. Al perder foco,
cerrar o pasar la app a segundo plano se detiene la reproducción. Los nombres
de campeones y las acciones permanecen como texto y controles accesibles reales;
la ilustración decorativa no duplica anuncios al lector.

La dependencia nativa exige recompilar antes de validar en iOS/Android. No se
publica ni despliega y no se inicia infraestructura ni observabilidad.

## Validación y cierre

Typecheck, lint y exportación web; revisar ambas ilustraciones en claro/oscuro,
movimiento reducido y ventanas estrechas. Verificar reproducción única,
cancelación al desmontar y conservación de las acciones del diálogo. Las builds
nativas y su QA se registran por separado: una exportación web no las acredita.
Documentar resultados y pendientes en DESIGN_SYSTEM y aprendizaje, con
retrospectiva. Revisar esta decisión si el coste de las builds o del reproductor
resulta desproporcionado respecto a dos ilustraciones.

## Referencias técnicas

- [Lottie React Native](https://github.com/lottie-react-native/lottie-react-native).
- [Lottie web y builds light](https://github.com/airbnb/lottie-web).

## Resultado de implementación

Las dos ilustraciones están conectadas. Typecheck, lint, exportaciones web y
Hermes, 37 pruebas de regresión/compatibilidad, compilación Android con lint y
compilación iOS arm64 genérica correctos. Se revisan fotogramas del reproductor
SVG en claro/oscuro. La QA visual de la app permanece pendiente por los límites
del navegador y CoreSimulator; véase [evidencia y retrospectiva](../engineering/LOTTIE_ILLUSTRATIONS_2026-11-02.md).
No se publica ni despliega.

## Autorización posterior de publicación

El usuario pide «Despliega». Se autoriza preparar y publicar el hito web v1.17.0
siguiendo ADR-0119. No incluye distribución en tiendas ni cambia API, renderer,
esquema o los límites de QA visual documentados.
