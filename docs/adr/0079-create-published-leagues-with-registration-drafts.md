# ADR-0079: Crear una liga publicada al transferir el borrador en el alta

- **Estado:** Aceptado
- **Fecha:** 2026-08-09
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** ADR-0078, exclusivamente en el estado persistido tras transferir el borrador
- **Superado por:** Ninguno

## Problema

El borrador local debe enviarse junto al alta, pero crear un segundo estado
persistido `draft` para una liga válida duplica una frontera que ya impone la
verificación de la cuenta.

## Alternativas

### A — Persistir la liga como `draft`

- **Ventajas:** diferencia explícita entre preparación y publicación.
- **Inconvenientes:** introduce una transición y consultas adicionales para un
  recurso que ya tiene todos los datos de una liga normal.
- **Coste de mantenimiento:** medio.

### B — Crear la liga como `published` junto al alta

- **Ventajas:** reutiliza el ciclo actual de liga y no añade estado ni transición.
- **Inconvenientes:** la liga existe antes de que su organizadora pueda iniciar
  sesión; su ID no se comunica antes de verificar.
- **Coste de mantenimiento:** bajo.

## Decisión del usuario

**Aceptada el 2026-08-09:** un borrador válido permanece local hasta enviar el
alta. El servidor crea una liga normal `published` y sus equipos junto a la
cuenta pendiente. Antes de verificar no hay sesión válida para administrarla ni
consultar los torneos de la cuenta; al verificar o iniciar sesión aparece en
«Administro».

## Validación

- El alta válida crea cuenta pendiente, liga `published` y equipos atómicamente.
- Antes de verificar, la cuenta no puede obtener una sesión ni su colección.
- Después de verificar, la liga aparece en la colección administrada sin crear
  otro recurso.

## Documentación afectada

- [Producto](../project/PRODUCT.md)
- [Modelo inicial de datos](../engineering/INITIAL_DATA_MODEL.md)
- [Decisiones](../governance/DECISIONS.md)
- [OpenAPI v1](../../contracts/openapi/v1/openapi.yaml)
