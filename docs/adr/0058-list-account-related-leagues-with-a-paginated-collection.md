# ADR-0058: Listar las ligas relacionadas con la cuenta mediante una colección paginada

- **Estado:** Aceptado
- **Fecha:** 2026-07-28
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

La home y «Torneos» necesitan recuperar de forma autenticada las ligas que una
cuenta administra o sigue. El contrato actual solo permite consultar una liga
visible por ID y el esquema aún no persiste las relaciones de seguimiento ni de
administración delegada.

## Contexto y restricciones

- Creador, administrador delegado y seguidor son relaciones distintas
  (ADR-0034); la clasificación de UI no sustituye autorización.
- La home definida en ADR-0057 no debe inventar datos ni tener un contrato propio
  que duplique el de la biblioteca.
- REST/OpenAPI contract-first, PostgreSQL, sqlc y el esquema inicial reseteable
  están aceptados (ADR-0009, ADR-0011, ADR-0046 y ADR-0053).

## Criterios de decisión

1. Una colección reutilizable por home y biblioteca.
2. Separación clara y estable entre administración y seguimiento.
3. Paginación que no cargue una colección completa por anticipado.
4. Sin tablas genéricas de roles ni endpoint de agregación específico de home.

## Alternativas

### A — Endpoint exclusivo de home

- **Ventajas:** una sola petición inicial.
- **Inconvenientes:** duplica el contrato de la biblioteca y fija tamaños de
  accesos rápidos en el backend.
- **Coste de mantenimiento:** medio.

### B — Una colección autenticada filtrada por relación

- **Ventajas:** home y biblioteca reutilizan el mismo recurso; cada colección
  pagina de forma independiente.
- **Inconvenientes:** la home hace hasta dos peticiones pequeñas.
- **Coste de mantenimiento:** bajo.

### C — Listar todas las ligas y filtrar en cliente

- **Ventajas:** ruta aparentemente simple.
- **Inconvenientes:** expone datos no necesarios, impide paginar correctamente y
  traslada autorización al cliente.
- **Coste de mantenimiento:** alto.

## Recomendación

**Opinión/recomendación:** alternativa B. Mantiene una sola proyección paginada
y permite que el servidor aplique la relación autorizada.

## Decisión del usuario

**Aceptada el 2026-07-28:**

- Se persisten `league_administrators` y `league_followers`. El creador se
  reconoce desde `leagues.organizer_account_id`; no se duplica en la tabla de
  administradores.
- `GET /v1/me/leagues` exige sesión válida y acepta
  `relationship=administered|followed`, `limit` de 1 a 50 y un `cursor` UUIDv7
  opcional. Ordena por ID UUIDv7 descendente; el último ID devuelto es el cursor
  de la página siguiente.
- `administered` incluye propiedad y administración delegada. `followed` excluye
  las ligas que la cuenta administra para mantener colecciones disjuntas.
- Cada elemento expone `id`, `name`, `state`, `createdAt` y una relación
  específica (`organizer`, `delegated` o `follower`). No expone emails, sesiones
  ni permisos de mutación.

## Consecuencias

- La home pide un límite pequeño de cada relación; «Torneos» reutiliza el mismo
  endpoint y solicita páginas posteriores.
- La consulta valida la sesión opaca contra PostgreSQL antes de acceder a las
  relaciones. El cliente no filtra permisos.
- Búsqueda global, filtros por estado, caché compartida, contadores y un resumen
  agregado de home quedan fuera de este corte.

## Validación

- Una sesión ausente, vencida o revocada recibe `401`.
- Un organizador y un administrador delegado aparecen en `administered`.
- Una liga seguida y además administrada no aparece en `followed`.
- Dos páginas consecutivas no repiten elementos y el cursor no expone secreto.

## Disparadores de revisión

- La colección supera el límite o la UX requiere búsqueda y filtros compuestos.
- El orden UUIDv7 no resulta suficiente para el orden de producto deseado.
- Aparecen nuevas relaciones que requieran otra perspectiva de biblioteca.

## Documentación afectada

- [Producto](../project/PRODUCT.md)
- [Modelo inicial de datos](../engineering/INITIAL_DATA_MODEL.md)
- [API](../engineering/API.md)
- [Decisiones](../governance/DECISIONS.md)
- [Aprendizaje](../project/LEARNING.md)
