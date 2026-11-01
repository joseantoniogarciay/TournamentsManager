# ADR-0174: Corregir glob y minimatch con compatibilidad acotada

- **Estado:** Aceptado
- **Fecha:** 2026-10-31
- **Decisor:** Usuario: «Menos la de hono el resto vamos de una y con la de hono planteame opciones».
- **Complementa:** ADR-0172 y ADR-0138 mediante excepciones de edad exactas.

## Problema y evidencia

Quedan tres registros moderados. El usuario autoriza corregir ahora glob y
minimatch, tras explicar su edad y cambios de major; Hono se mantiene para
comparar opciones. Glob 13.0.6 no interrumpe todo el recorrido al cancelar;
minimatch antiguo permite amplificación de patrones. El grafo incluye
minimatch 10.2.6 y 3.1.5. Sus consumidores son Expo, TypeDoc, Spectral y glob.

## Alternativas y coste

- Esperar a los consumidores: coste bajo, conserva los dos avisos.
- Actualizar todas las herramientas: coste alto y cambios ajenos a la corrección.
- Selectores exactos y adaptación mínima de consumidores: coste medio de pruebas
  y adaptaciones de consumidores; conserva herramientas y matriz Expo.

Se recomienda y se aplica la tercera opción dentro de la autorización.
Glob 7.2.3 no está en el rango vulnerable, pero sus consumidores internos no
son compatibles con minimatch 11. Su único consumidor, plugin-commonjs 22.0.2,
usa solo glob.sync. Se resuelve ese enlace a glob 14.0.3, que conserva sync,
y se prueba una generación real de Rollup. No se añade un wrapper global ni
se cambia la API publicada de minimatch.

## Decisión y alcance exacto

Resolver glob@13.0.6 → 14.0.3, plugin-commonjs@22.0.2>glob → 14.0.3
y minimatch@3.1.5/10.2.6 → 11.0.3.
Las revisiones finales de parche incorporan las correcciones iniciales y sus
ajustes posteriores de matching/rendimiento. Añadir excepciones exactas para
esas dos versiones y las transitivas necesarias path-scurry@3.0.0 y
@isaacs/nonrecursive-glob-match@1.0.3, sin ampliar el control global.

Spectral 1.22.0 y 1.24.0 importan la función por default desde CommonJS.
Dos parches de importación, vinculados a sus versiones exactas, adaptan la
importación a minimatch.minimatch. No cambian las opciones ni las reglas de
filtrado. Retirar el parche cuando Spectral soporte la API nueva.
TypeDoc 0.28.20 usa matchOne, retirada en minimatch 11. Un tercer parche
sustituye ese recorrido por match(path, true), conservando la negación para
la decisión de descenso de directorios. Su prueba expande entradas glob y
excluye archivos. No se cambia el compilador privado de TypeDoc.
No se fuerza ninguna resolución Hono ni se acepta una de sus alternativas.

## Verificación

Candidato aislado; revisar todas las versiones nuevas y fechas; instalación
congelada y audit autorizado. Probar filtros Spectral, la generación Rollup del consumidor de glob.sync,
glob moderno y cancelación, consumidor Expo/Fingerprint y generación
OpenAPI/TypeDoc. Regresiones, exportación web y make verify antes de subir a
develop. No se autoriza main, despliegue u observabilidad.

## Resultado y retrospectiva

Implementado y verificado: audit 3 → 1 moderado, únicamente Hono, sin IDs
nuevos. Pasa make verify y 130 regresiones Node; la matriz Expo no cambia.
Las pruebas de consumidores detectaron fallos que el conteo no muestra:
exportación default de Spectral, estructura interna esperada por glob 7 y
matchOne eliminada en TypeDoc. Se adaptan las importaciones y la API pública
de TypeDoc, y se actualiza el único consumidor de glob 7 a glob 14.

El [informe](../engineering/GLOB_MINIMATCH_SECURITY_FIX_2026-11-01.md) conserva
las cuatro fechas de retirada de excepciones y las opciones Hono sin aplicar.
La recomendación para Hono es actualizar únicamente void-server; la prueba
de resolución añade tres versiones maduras y elimina la rama 1, pero aún
requiere decisión del usuario y pruebas de ejecución.
