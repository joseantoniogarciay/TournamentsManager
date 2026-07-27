# ADR-0049: Usar IDs públicos de liga para la lectura de solo lectura

- **Estado:** Aceptado
- **Fecha:** 2026-07-27
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** ADR-0033 y la parte de enlaces compartibles de ADR-0045
- **Superado por:** Ninguno

## Problema

Una liga publicada debe poder consultarse sin sesión y compartirse con una URL
estable. El enlace no listado introducía un secreto, una tabla y un endpoint
adicionales aunque el usuario acepta que conocer o adivinar el identificador de
una liga permita únicamente su lectura pública.

## Contexto y restricciones

- Las mutaciones de una liga requieren sesión y autorización de negocio; conocer
  un identificador nunca concede permisos de edición, resultados o administración.
- Los UUIDv7 identifican recursos, pero no son secretos ni deben ser la única
  barrera si la lectura necesitara privacidad.
- La primera migración ya está aplicada en la base local; se conserva inmutable y
  el cambio se expresa mediante una migración nueva y reversible.
- Invitaciones, audiencias restringidas y búsqueda pública permanecen fuera del
  primer incremento.

## Criterios de decisión

1. minimizar modelo, contrato y mantenimiento;
2. permitir URLs estáticas construidas a partir del ID de la liga;
3. conservar una frontera estricta entre lectura pública y autorización de
   escritura;
4. explicitar la consecuencia de privacidad en lugar de confiar en que un ID sea
   difícil de enumerar.

## Alternativas

### Alternativa A — Enlace no listado con token independiente

- **Ventajas:** la lectura depende de poseer un secreto aleatorio independiente
  del ID de la liga.
- **Inconvenientes:** necesita generar, almacenar y buscar una huella de token,
  además de una ruta de API y una URL específicas.
- **Coste de adopción:** bajo o moderado.
- **Coste de mantenimiento:** bajo inicialmente; mayor si se añade rotación,
  revocación o caducidad.
- **Riesgos:** interpretar el enlace como un permiso de edición o añadir una
  abstracción que el alcance no necesita.

### Alternativa B — Lectura pública por ID de liga

- **Ventajas:** elimina el token y `league_share_links`; el cliente construye una
  URL estable con el ID recibido y el contrato usa una única ruta de recurso.
- **Inconvenientes:** quien conozca un ID de una liga visible puede leer su
  proyección pública.
- **Coste de adopción:** bajo.
- **Coste de mantenimiento:** bajo.
- **Riesgos:** tratar UUIDv7 como secreto u olvidar proteger las mutaciones con
  autenticación y autorización.

### No cambiar

- **Consecuencias:** se conserva una capacidad secreta y su tabla sin que la
  privacidad adicional sea un requisito del usuario.

## Comparación

La alternativa A es apropiada cuando la URL debe funcionar como un secreto de
lectura. El requisito aceptado es distinto: la liga publicada es legible por
cualquiera que conozca su ID. Por tanto, B satisface el uso con menos estado y
sin degradar la protección de operaciones mutables, que no depende del enlace.

## Recomendación

**Recomendación:** adoptar B. Es la solución mínima suficiente mientras la
lectura sea pública por diseño y no se necesite revocar el acceso de lectura.

## Decisión del usuario

**Aceptada el 2026-07-27:** las ligas visibles se consultan sin sesión mediante
`GET /leagues/{leagueId}`. El ID de la liga es público y permite construir la URL
en el cliente. Se eliminan el token de compartición, la tabla
`league_share_links`, el endpoint de enlaces compartidos y `shareUrl`.

## Consecuencias

### Positivas

- Publicar una liga ya no crea un secreto ni una fila adicional.
- El cliente puede enlazar estáticamente una liga desde su `id`.
- La lectura pública y la autorización de escritura quedan expresadas como reglas
  independientes.

### Negativas y deuda aceptada

- Cualquier persona que conozca o adivine el ID puede consultar la proyección
  pública de una liga visible.
- Si en el futuro se necesita restringir o revocar la lectura, habrá que decidir
  enlaces con capacidad, invitaciones o una política de audiencia explícita.

## Validación

- La publicación no crea un token ni una fila de enlace.
- `GET /leagues/{leagueId}` devuelve solo la proyección pública de una liga
  visible sin sesión; un borrador no se expone.
- Conocer el ID no permite ninguna mutación sin la sesión y autorización
  correspondientes.
- El esquema final y el cliente OpenAPI no conservan referencias operativas a
  `league_share_links`, `shareToken` ni `shareUrl`.

## Disparadores de revisión

- Se introducen datos que no deban ser legibles por quien conozca el ID.
- Se necesita revocar, caducar o auditar accesos de lectura.
- El producto requiere invitaciones, una audiencia restringida o descubrimiento
  público con búsqueda.

## Documentación afectada

- [Producto](../project/PRODUCT.md)
- [Roadmap](../project/ROADMAP.md)
- [Modelo inicial de datos](../engineering/INITIAL_DATA_MODEL.md)
- [API](../engineering/API.md)
- [OpenAPI v1](../../contracts/openapi/v1/openapi.yaml)
- [Migraciones PostgreSQL](../../apps/backend/db/migrations/)
- [Decisiones](../governance/DECISIONS.md)
- [Aprendizaje](../project/LEARNING.md)
