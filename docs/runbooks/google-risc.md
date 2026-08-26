# Activación y comprobación de Google RISC

> Estado: implementado en la API; activación externa pendiente en cada proyecto
> OAuth público. Última prueba: pendiente.

## Síntoma y alcance

Sin un stream RISC registrado, Google no puede avisar a FastTourney cuando una
revocación o compromiso exige cerrar sus sesiones propias. Este runbook activa
Google Cross-Account Protection y verifica `POST /v1/risc/events`.

## Prerrequisitos y permisos

- La API pública responde HTTPS en `https://dev-api.fasttourney.com/v1/risc/events`
  para `dev`, o su URL equivalente de `prod`, con `GOOGLE_CLIENT_IDS` configurado.
- La persona operadora administra el proyecto Google Cloud correcto y puede
  aceptar las condiciones RISC en nombre de su organización.
- Una cuenta de servicio con solo **RISC Configuration Admin**
  (`roles/riscconfigs.admin`) guarda su JSON exclusivamente en el gestor de
  secretos operativo; nunca en Git, una imagen ni salida de terminal.

## Diagnóstico seguro

1. En Google Auth Platform, confirmar el proyecto y los clientes OAuth realmente
   desplegados, sin copiar secretos.
2. Desde fuera de la red local, una petición sin SET a la ruta debe devolver
   `400`, sin crear cuentas ni sesiones.
3. Consultar el stream actual con la cuenta de servicio. Si apunta a otro
   entorno, detenerse y corregirlo antes de actualizarlo.

## Activación

1. Habilitar RISC API en el mismo proyecto OAuth y aceptar sus condiciones.
2. Crear y guardar la cuenta de servicio mínima anterior fuera del repositorio.
3. Registrar mediante `stream:update` el receptor HTTPS público y solicitar
   `sessions-revoked`, `tokens-revoked` y `account-disabled`.
4. Ejecutar `stream:verify` con un estado aleatorio no sensible. Google debe
   recibir `202`; comprobar solo código y traza correlacionada, nunca el SET.

## Verificación desde la persona usuaria

1. Iniciar sesión en el artefacto público con Google y confirmar una sesión.
2. Revocar el acceso de FastTourney desde la cuenta Google de prueba.
3. Comprobar que la siguiente solicitud autenticada exige iniciar sesión otra vez.
   Si existe contraseña local, se puede usar ese método.

## Recuperación, rollback y escalado

- Si el receptor devuelve `5xx`, restaurar la API y repetir `stream:verify`;
  Google solo reintenta durante un tiempo limitado.
- Para detener entregas, actualizar el stream a `disabled`. No borrar clientes
  OAuth ni identidades externas para resolver un incidente de webhook.
- Escalar si un SET válido no revoca sesiones, la firma es inesperada o se
  expone la cuenta de servicio. Rotar esa clave y revisar logs sin publicar datos.

## Riesgos

- RISC no cubre actualmente cuentas Google Workspace.
- No sustituye reautenticación ni controles propios de fraude.
- Un SET repetido es seguro: el `jti` procesado no revoca un login posterior.

## Fuentes

- [Google Cross-Account Protection](https://developers.google.com/identity/protocols/risc?hl=es-419)
