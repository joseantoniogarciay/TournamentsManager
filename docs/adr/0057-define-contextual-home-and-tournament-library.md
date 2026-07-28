# ADR-0057: Definir una home contextual y una biblioteca de torneos por relación

- **Estado:** Aceptado
- **Fecha:** 2026-07-28
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

La primera pantalla debe orientar tanto a quien aún no ha accedido como a quien
ya tiene una cuenta, sin mezclar permisos de administración con el simple hecho
de seguir una liga ni imponer al navegador una navegación propia de una app.

## Contexto y restricciones

- Una persona sin sesión puede preparar un borrador local; persistir o publicar
  exige una cuenta verificada (ADR-0031).
- Creador y administrador delegado son relaciones de administración; seguidor
  es una relación independiente y sin permisos (ADR-0034).
- El cliente usa Expo Router y adapta la presentación por plataforma (ADR-0055).
- El contrato vigente permite leer una liga por ID público, pero no lista las
  ligas relacionadas con la sesión. Esta decisión no autoriza inventar datos en
  la interfaz: el endpoint autenticado y su paginación se decidirán
  contract-first antes de implementarlos.

## Criterios de decisión

1. Que la acción principal de crear una liga sea inmediata y válida con o sin
   sesión.
2. Separar con claridad administración y seguimiento, sin convertir la home en
   una copia completa del catálogo.
3. Mantener URLs web directas y el historial habitual del navegador.
4. Conservar contexto al cambiar de sección en una app sin persistir navegación
   obsoleta indefinidamente.
5. Añadir la menor estructura posible al cliente.

## Alternativas

### A — Home estática y un único listado de torneos

- **Ventajas:** menos rutas y consultas iniciales.
- **Inconvenientes:** no distingue qué puede gestionar una persona de lo que
  solo sigue; obliga a inspeccionar cada liga para orientarse.
- **Coste de adopción:** bajo.
- **Coste de mantenimiento:** bajo al inicio, creciente cuando se añadan roles.
- **Riesgos:** confundir visibilidad, seguimiento y autorización.

### B — Home contextual, accesos rápidos y biblioteca separada por relación

- **Ventajas:** prioriza las ligas de uso frecuente y representa relaciones ya
  existentes sin crear permisos nuevos.
- **Inconvenientes:** exige una consulta autenticada adicional y estados de
  carga, vacío y error.
- **Coste de adopción:** medio; requiere ampliar el contrato de API después.
- **Coste de mantenimiento:** bajo; las categorías son proyecciones de
  relaciones de dominio existentes.
- **Riesgos:** duplicar una liga en dos grupos si no se fija una regla de
  precedencia.

### C — Navegación persistida igual para web y apps

- **Ventajas:** comportamiento aparentemente uniforme.
- **Inconvenientes:** contradice las expectativas del navegador, complica URLs
  y puede restaurar estados desactualizados.
- **Coste de adopción:** medio o alto.
- **Coste de mantenimiento:** alto por sincronización y casos de restauración.
- **Riesgos:** navegación sorprendente y enlaces poco fiables.

## Comparación

La alternativa B satisface la orientación solicitada con las relaciones que ya
existen en producto. No necesita un nuevo rol ni una librería de estado. La C
añade persistencia sin valor proporcional: la paridad exigida es funcional, no
una copia del historial entre plataformas.

## Recomendación

**Opinión/recomendación:** adoptar B y conservar el historial según la
convención de cada plataforma. Es la solución mínima que permite encontrar y
distinguir torneos sin anticipar búsqueda global, notificaciones ni caché.

## Decisión del usuario

**Aceptada el 2026-07-28:**

- La ruta `/` es la home. Siempre muestra una acción principal «Crear torneo» y
  un acceso a cuenta en la botonera.
- Sin sesión, «Crear torneo» inicia o retoma el borrador local. Con sesión
  verificada, la home muestra accesos rápidos a ligas administradas y guardadas.
- «Administro» incluye las ligas creadas por la cuenta y aquellas donde es
  administrador delegado. «Guardados» muestra las ligas seguidas. Si una liga
  pertenece a ambos conjuntos, solo aparece en «Administro».
- La sección «Torneos» separa las mismas colecciones en «Administro» y «Sigo»;
  esta clasificación mejora la navegación, pero no concede ni sustituye
  autorización del backend.
- En iOS y Android cada sección de la botonera conserva su pila de navegación
  mientras la app está activa. No se restaura tras reiniciar la app en esta
  versión. En web, la botonera actúa como acceso directo: cada URL carga como
  página y el historial del navegador conserva el flujo habitual.

## Consecuencias

### Positivas

- La home presenta una siguiente acción útil antes y después de iniciar sesión.
- Administración y seguimiento quedan nombrados de forma consistente con el
  dominio.
- La navegación nativa retiene contexto sin introducir almacenamiento, una store
  global ni comportamiento web artificial.

### Negativas y deuda aceptada

- La API debe añadir una proyección autenticada de ligas relacionadas y definir
  filtro, orden, paginación y representación de la relación antes de que la UI
  consuma datos reales.
- Búsqueda, exploración global, filtros adicionales, notificaciones y una
  persistencia de navegación entre reinicios quedan fuera de este corte.

## Validación

- Sin sesión se ven «Crear torneo» y el acceso a cuenta; no se presentan listas
  personalizadas como si existieran.
- Con sesión, una liga administrada aparece en «Administro» y no se duplica en
  «Guardados».
- La sección «Torneos» conserva la separación sin permitir una mutación no
  autorizada.
- Una URL de web se puede abrir directamente y usa atrás/adelante del navegador;
  al alternar secciones en móvil se conserva la pantalla previa durante la
  sesión actual.

## Disparadores de revisión

- La búsqueda o una biblioteca grande exige paginación, filtros o una estrategia
  de caché medida.
- La restauración entre reinicios aporta valor demostrado y justifica persistir
  estado de navegación.
- Nuevas relaciones de usuario con una liga dejan insuficientes «Administro» y
  «Sigo».

## Documentación afectada

- [Producto](../project/PRODUCT.md)
- [Arquitectura](../engineering/ARCHITECTURE.md)
- [Decisiones](../governance/DECISIONS.md)
- [Aprendizaje](../project/LEARNING.md)
