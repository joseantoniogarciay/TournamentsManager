# ADR-0152: Corregir dos errores en la generación de recursos Android de Expo

- **Estado:** Propuesto
- **Fecha:** 2026-10-09
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** Ninguno; concreta el mantenimiento CNG de ADR-0015.
- **Superado por:** Ninguno

## Problema

El contraste completo de Android Lint con K1 alcanza el informe de la app y
encuentra dos errores: MissingPrefix por data-generated en intent-filter y
NewApi por android:windowSplashScreenBehavior en values/styles.xml aunque
minSdk es 24 y el atributo requiere API 33. La fuente está en
@expo/config-plugins 57.0.8 y expo-splash-screen 57.0.7, respectivamente.
El lint del módulo local global-feedback termina con cero errores y seis avisos.

## Contexto y restricciones

CNG genera android/; no se corrigen sus archivos a mano. Mantener los enlaces
aceptados, sus dominios, autoVerify, identidad y categorías; mantener splash,
logo, temas y soporte Android desde API 24. No subir minSdk para ocultar el
problema ni introducir permisos, baseline o supresiones. Conservar versiones
fijadas y siete días de maduración de ADR-0138. El diagnóstico K1 no autoriza
sustituir el gate K2, que sigue bloqueado por una excepción upstream.

El proyecto ya usa patchedDependencies de pnpm. Añadir dos parches de generación
amplía su mantenimiento: cada actualización de Expo debe revisarlos y retirarlos
cuando la fuente corregida produzca el mismo resultado sin ellos.

## Criterios de decisión

Recursos válidos en todas las API soportadas, generación idempotente, contrato
funcional intacto, ausencia de supresiones y coste de mantenimiento acotado.

## Alternativas

### A — Dos parches exactos de los generadores fijados

- Corregir la marca de generación del intent-filter con metadata en el namespace
  tools, migrando la marca antigua y conservando la eliminación de filtros
  generados previos al regenerar. El manifest declara ese namespace y el APK
  no debe conservar la metadata de herramientas.
- Mantener los atributos de splash compatibles en el estilo base. Generar el
  estilo correspondiente a values-v33 con icon_preferred; conservar la herencia,
  logo y referencias de color que resuelven el tema claro u oscuro.
- Versionar los parches para @expo/config-plugins 57.0.8 y expo-splash-screen
  57.0.7 y registrar su hash en el lockfile; no cambiar las versiones.
- Ventajas: corrige el origen, permanece reproducible con CNG y no debilita lint.
- Inconvenientes/coste: dos parches que revisar en upgrades; validación de
  generación repetida y builds nativas. Coste medio inicial, localizado después.
- Riesgos: duplicar filtros o estilos, dejar metadata antigua, alterar splash
  en API antiguas. Deben comprobarse antes de adoptar el cambio.

### B — Esperar versiones corregidas de Expo

- Ventajas: evita ampliar los parches locales; coste de mantenimiento bajo.
- Inconvenientes: conserva ambos errores y exige identificar publicación,
  maduración y compatibilidad antes de actualizar. Sin fecha garantizada.
- No adoptar una versión solo por anunciar una corrección: reproducir el lint.

### No cambiar ni registrar el resultado

Los errores quedan ocultos detrás del crash de K2. No es un cierre válido del QA.

## Comparación y recomendación

**Opinión/recomendación:** A, dentro del mecanismo de parches ya existente,
con pruebas de idempotencia y matriz API 24/34. Corrige dos salidas concretas
sin migrar AGP ni cambiar el diseño del producto. B reduce mantenimiento si
el usuario prefiere conservar esas incidencias hasta un upgrade compatible.

## Decisión del usuario

**Pendiente.** No se aplican parches ni cambios de configuración mientras este
ADR siga Propuesto. Esta decisión no incluye activar producción, instalar una
versión joven ni seleccionar K1 como motor permanente.

## Consecuencias

La alternativa A añade mantenimiento temporal y evidencia obligatoria en cada
upgrade. La B mantiene abiertos los dos errores además del fallo K2. Ninguna
alternativa certifica automáticamente accesibilidad ni todo el QA del cliente.

## Validación

- Generar dos veces con la misma configuración: mismos filtros, ningún duplicado;
  migrar un manifest con data-generated y conservar filtros no generados.
- Cambiar el dominio/retirar filtros en configuración: retirar solo los generados.
- Manifest XML válido, namespace declarado y metadata ausente en manifest de APK.
- Estilo base sin atributo de API 33 y estilo v33 completo; no sobrescribir otros
  estilos del recurso, conservar referencias de colores claro/oscuro.
- Lint complementario de app y módulo local sin los dos errores, sin supresiones.
  Registrar avisos y conservar el límite del gate K2 explícito.
- Build Android instalada; arranque y enlace en API 34 y arranque en API 24,
  respetando fixtures y sin envíos de formularios.
- Typecheck, exportación web, gate común e iOS ante cualquier efecto compartido.

## Disparadores de revisión

Publicación madura que corrige los generadores, cambio de SDK, pérdida de
idempotencia, error de recursos o cambio visual de splash en una API soportada.

## Documentación y fuentes

- [Diagnóstico de herramientas](../engineering/ANDROID_LINT_ENGINE_ANALYSIS_2026-10-09.md).
- PRODUCT_QA_REVIEW_2026-10-04.md, DEVELOPMENT.md y LEARNING.md.
- [Android: recursos alternativos y calificadores de versión](https://developer.android.com/guide/topics/resources/providing-resources).
- [Android: atributos del namespace tools](https://developer.android.com/studio/write/tool-attributes).
- Generadores fijados de Expo y los informes XML privados de lint de esta fase.
