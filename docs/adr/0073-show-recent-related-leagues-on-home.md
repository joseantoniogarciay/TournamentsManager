# ADR-0073: Mostrar en Inicio las ligas relacionadas con actividad reciente

- **Estado:** Aceptado
- **Fecha:** 2026-08-02
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

Una persona autenticada necesita ver desde Inicio qué ligas relacionadas han
tenido cambios recientemente. La colección vigente separa relaciones y ordena
por creación, por lo que no puede responder con fidelidad a «las últimas ligas
con actividad».

## Contexto y restricciones

- «Administro» y «Sigo» siguen siendo colecciones separadas y paginadas
  (ADR-0057 y ADR-0058); la clasificación no concede permisos.
- La home no inventa actividad ni fusiona colecciones en el cliente.
- El cliente debe poder recargar la proyección mediante pull-to-refresh sin
  ocultar el contenido que ya se ha cargado.
- Por ahora existen creación/publicación e inicio de liga; resultados, edición
  y otros cambios posteriores aún no están implementados.

## Alternativas

### A — Marca `last_activity_at` por liga y una proyección reciente autenticada

- **Ventajas:** orden verificable en el servidor, una liga no se duplica aunque
  la cuenta la administre y siga, y Inicio solicita solo cinco elementos.
- **Inconvenientes:** cada futura mutación de liga debe decidir explícitamente
  si representa actividad.
- **Coste de adopción:** medio: columna, índice, contrato y consulta dedicada.
- **Coste de mantenimiento:** bajo: una marca temporal y una proyección acotada.

### B — Feed de eventos de actividad

- **Ventajas:** puede mostrar qué cambió, cuándo y quién lo hizo.
- **Inconvenientes:** introduce una entidad, retención, autorización y copy de
  eventos antes de que Inicio los necesite.
- **Coste de adopción y mantenimiento:** alto.

### C — Ordenar por creación la colección existente

- **Ventajas:** no requiere cambios de backend.
- **Inconvenientes:** una liga antigua con resultados o inicio reciente queda
  fuera; no cumple el significado de actividad.
- **Coste de mantenimiento:** bajo, con semántica incorrecta.

## Recomendación

**Opinión/recomendación:** alternativa A. Es la mínima proyección que expresa
la necesidad sin adelantar un sistema de feed.

## Decisión del usuario

**Aceptada el 2026-08-02:**

- `leagues.last_activity_at` se inicializa al crear/publicar una liga y se
  actualiza al iniciarla. Las futuras mutaciones de contenido o estado de liga
  deberán actualizarla si cambian algo relevante para participantes o seguidores.
  Seguir o dejar de seguir no modifica la actividad de la liga.
- `GET /v1/me/recent-leagues` exige sesión y devuelve, como máximo, cinco ligas
  relacionadas, sin duplicados, ordenadas por `lastActivityAt` descendente y por
  ID descendente como desempate. Una relación administrada prevalece sobre
  seguida.
- Inicio muestra esa proyección solo con sesión. Un resultado vacío explica que
  ahí aparecerán las últimas ligas con actividad; sin sesión conserva su
  contenido público actual.
- Inicio y Torneos ofrecen pull-to-refresh solo con sesión. Cada pantalla
  refresca exclusivamente sus datos; los fallos siguen el feedback común seguro.

## Consecuencias

- La home no necesita cargar ni fusionar las páginas de biblioteca.
- La lista reciente no se pagina: su límite fijo de cinco responde a un resumen,
  no a una biblioteca.
- El primer modelo de actividad no explica el tipo de cambio; eso se evaluará
  solo si el producto necesita un historial visible.

## Validación

- Una liga iniciada aparece antes que otra creada anteriormente.
- Una liga administrada y seguida aparece una vez y con relación administrada.
- Una cuenta sin relaciones recibe una lista vacía y ve el estado vacío.
- Pull-to-refresh vuelve a consultar la proyección sin sustituir contenido por
  una pantalla de carga.

## Disparadores de revisión

- Necesidad de explicar qué cambio produjo la actividad o quién lo realizó.
- Más de cinco elementos o filtros temporales aportan valor demostrado.
- Nuevas mutaciones no pueden actualizar `last_activity_at` de forma atómica.

## Documentación afectada

- [Producto](../project/PRODUCT.md)
- [API](../engineering/API.md)
- [Modelo inicial de datos](../engineering/INITIAL_DATA_MODEL.md)
- [Decisiones](../governance/DECISIONS.md)
- [Aprendizaje](../project/LEARNING.md)
