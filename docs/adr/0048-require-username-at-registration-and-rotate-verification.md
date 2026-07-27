# ADR-0048: Requerir username en el alta y rotar la verificación pendiente

- **Estado:** Aceptado
- **Fecha:** 2026-07-27
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** ADR-0034 y ADR-0045, parcialmente, en el momento de elegir el username
- **Superado por:** ADR-0052, exclusivamente en la obligatoriedad del borrador al alta

## Problema

El registro debe recoger todos los datos de identidad desde el principio. Elegir
el `username` solo al verificar añade un paso y deja una cuenta pendiente con un
perfil incompleto. Un login correcto de una cuenta pendiente debe ayudar a
completar la verificación sin reutilizar enlaces antiguos.

## Alternativas

### A — Username obligatorio en el alta y nuevo token al login pendiente

- **Ventajas:** un único formulario de identidad; el email es la única prueba
  pendiente; los enlaces antiguos se invalidan explícitamente.
- **Inconvenientes:** una cuenta sin verificar reserva temporalmente un username.
- **Coste de mantenimiento:** bajo; requiere límite de tasa y purga ya acordados.

### B — Elegir username al verificar y conservar el token inicial

- **Ventajas:** no reserva usernames antes de verificar.
- **Inconvenientes:** un paso adicional y enlaces antiguos activos tras nuevos
  intentos de login.
- **Coste de mantenimiento:** medio.

## Decisión del usuario

**Aceptada el 2026-07-27:** adoptar alternativa A.

- Todo registro local exige `email`, contraseña, `username` y borrador.
- El username es minúsculo, único e inmutable en este corte; no necesita versión
  de presentación ni normalizada.
- El email conserva su forma original para envío; una clave de comparación
  insensible a mayúsculas impide duplicados para el producto.
- Una cuenta pendiente contiene todos los datos y solo espera acreditar el email.
- Un login con contraseña correcta de una cuenta pendiente invalida el token de
  verificación activo, crea otro de 24 horas y solicita su envío; no crea sesión.
- Un token tiene estados mutuamente excluyentes: activo, consumido o invalidado.
- El alta social crea directamente la cuenta verificada y sesión una vez que el
  backend valide la credencial del proveedor y reciba el username. Sigue fuera
  del primer incremento implementado.

## Consecuencias

- La verificación ya no recibe username: consume un token activo y activa una
  cuenta completa de forma atómica.
- La respuesta de login pendiente es `202 Accepted`; credenciales incorrectas
  siguen recibiendo una respuesta de autenticación no reveladora.
- Mientras la migración inicial siga siendo exclusivamente local y no compartida,
  se consolida en `00001_initial_schema.sql`. Una vez compartida, cada cambio se
  hará mediante una migración nueva e inmutable.

## Validación

- No se insertan cuentas pendientes sin username o con username en mayúsculas.
- El segundo token invalida el primero y solo el activo puede verificar.
- Un login pendiente no crea sesión; una verificación válida sí.
- El contrato OpenAPI y cliente generado ya no envían username al verificar.

## Disparadores de revisión

- Abuso medido de reserva temporal de usernames.
- Necesidad de cambiar username, admitir mayúsculas de presentación o nombres
  internacionalizados.
- Un proveedor social no acredita un email suficiente para activar la cuenta.

## Documentación afectada

- [Modelo inicial de datos](../engineering/INITIAL_DATA_MODEL.md)
- [Identidad](../engineering/IDENTITY.md)
- [Producto](../project/PRODUCT.md)
- [OpenAPI](../../contracts/openapi/v1/openapi.yaml)
