# ADR-0094: Gestionar métodos de acceso añadibles y eliminables

- **Estado:** Aceptado
- **Fecha:** 2026-08-13
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** ADR-0068, exclusivamente en la gestión de métodos de acceso
- **Superado por:** Ninguno

## Problema

Una cuenta creada con Google no tiene contraseña, pero el flujo anterior exigía
una contraseña actual para crearla. Además, un método ya vinculado seguía
pareciendo accionable y no había una eliminación segura.

## Decisión del usuario

**Aceptada el 2026-08-13:** una cuenta mantiene siempre al menos un método de
acceso. Se permiten contraseña y Google como métodos independientes:

- una cuenta solo Google añade contraseña tras acreditar Google;
- una cuenta con contraseña puede cambiarla tras acreditar una credencial
  actual y vincular Google tras acreditar la contraseña;
- con ambos métodos se permite eliminar Google solo tras acreditar la
  contraseña, y eliminar la contraseña solo tras acreditar Google;
- con un único método no se ofrece ni permite eliminarlo.

Toda mutación exige un ticket opaco, de un solo uso, cinco minutos, ligado a la
sesión, cuenta y finalidad (`set-local-password`, `link-google`,
`unlink-google` o `remove-local-password`). La contraseña se recupera por email
solo si ya existe una credencial local; crear la primera contraseña es una
acción autenticada y no una recuperación.

## Consecuencias

- Datos de acceso muestra estados explícitos y acciones independientes; una
  fila vinculada no inicia otra vinculación.
- La vinculación y las eliminaciones usan `ModalDialog`, no rutas modales ni
  cards incrustadas. El challenge de Google se prepara al abrir el diálogo y se
  renueva antes de vencer.
- OpenAPI, persistencia y pruebas deben hacer cumplir el último método y la
  finalidad del ticket en el backend.

## Validación

- Una cuenta solo Google puede crear contraseña tras una prueba Google válida.
- Ninguna operación puede borrar el último método.
- Un ticket de una finalidad no autoriza otra mutación.
- Tras cada mutación, `GET /me/access-methods` refleja el método real.

## Disparadores de revisión

- Incorporación de Apple, passkeys o MFA.
- Requisito de revocar todas las sesiones tras cambios de credenciales.
