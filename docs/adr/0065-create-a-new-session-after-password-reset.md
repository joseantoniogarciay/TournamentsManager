# ADR-0065: Crear una sesión nueva tras restablecer la contraseña

- **Estado:** Aceptado
- **Fecha:** 2026-07-31
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** ADR-0064, exclusivamente sobre el tratamiento de sesiones al completar un restablecimiento
- **Superado por:** Ninguno

## Problema

El restablecimiento prueba control actual del email y exige una contraseña nueva.
Pedir login adicional o conservar una sesión previa crea un recorrido distinto de
la verificación de registro.

## Alternativas

### A — Crear una sesión nueva y revocar todas las anteriores

- **Ventajas:** mismo modelo que la verificación; una sola credencial nueva y
  una navegación coherente.
- **Inconvenientes:** el dispositivo que ya tenía sesión también la sustituye.
- **Coste de mantenimiento:** bajo; reutiliza el reemplazo de sesión existente.

### B — Rotar y conservar la sesión presentada

- **Ventajas:** no emite otra sesión en el dispositivo actual.
- **Inconvenientes:** añade ramas y experiencia distinta de registro.
- **Coste de mantenimiento:** medio.

## Recomendación

**Recomendación:** alternativa A por consistencia y menor superficie de estado.

## Decisión del usuario

**Aceptada el 2026-07-31:** alternativa A. Tras un restablecimiento válido, se
revocan todas las sesiones de la cuenta, se crea una sesión nueva para el
dispositivo que completó el flujo y el cliente restablece su navegación igual
que durante la verificación de registro.

## Validación

- Ninguna sesión previa autentica después del cambio, incluida la presentada.
- El resultado entrega una única sesión nueva y el cliente la persiste según su
  transporte.
- Web reemplaza la URL por `/`; móvil reconstruye Inicio, Torneos y Cuenta.

## Documentación afectada

- `docs/engineering/IDENTITY.md`
- `contracts/openapi/v1/openapi.yaml`
- `docs/project/LEARNING.md`
