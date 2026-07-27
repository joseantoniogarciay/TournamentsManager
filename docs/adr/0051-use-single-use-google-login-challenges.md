# ADR-0051: Usar challenges de un solo uso para el login con Google

- **Estado:** Aceptado
- **Fecha:** 2026-07-27
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

El primer incremento incorpora Google. Un ID token no debe poder reutilizarse para
crear sesiones durante su vigencia ni crear cuentas o vínculos duplicados.

## Alternativas

### A — Aceptar directamente el ID token

- **Ventajas:** un endpoint y menos estado.
- **Inconvenientes:** una repetición válida puede abrir más de una sesión.
- **Coste de mantenimiento:** bajo.

### B — Challenge persistido de nonce, corto y de un solo uso

- **Ventajas:** enlaza el inicio y cierre del login; impide la repetición tras el
  primer consumo y funciona igual en web, iOS y Android.
- **Inconvenientes:** requiere una tabla temporal, purga y dos llamadas.
- **Coste de mantenimiento:** bajo.

### C — Authorization Code Flow con PKCE completo

- **Ventajas:** el backend intercambia directamente el código.
- **Inconvenientes:** añade redirects, estado y configuración por plataforma antes
  de demostrar el límite común de identidad.
- **Coste de mantenimiento:** medio o alto.

## Decisión del usuario

**Aceptada el 2026-07-27:** adoptar B.

- El cliente solicita un challenge Google, recibe `id` y `nonce`, y entrega el
  nonce a Google.
- Entrega después `idToken` y `challengeId` por `POST` al backend.
- El backend valida firma, `iss`, `aud`, `exp` y `nonce`; bloquea y consume el
  challenge al crear una cuenta, vínculo o sesión.
- Los challenges duran cinco minutos, se purgan y nunca contienen una sesión.
- `external_identities` guarda el vínculo estable `(issuer, subject)`; su
  unicidad evita cuentas duplicadas bajo concurrencia.
- Una alta Google nueva crea una cuenta pendiente y usa la verificación de email
  existente antes de emitir sesión. Una coincidencia con cuenta local inicia el
  desafío de vinculación explícita de ADR-0010, nunca una unión automática.

## Consecuencias

- El frontend deshabilita el botón mientras el intento está en curso. Si un
  challenge ya fue consumido, consulta la sesión actual en web o inicia un nuevo
  intento en móvil si se perdió la respuesta Bearer.
- Un segundo uso no crea otra cuenta ni vínculo; devuelve un estado recuperable.
- No se implementa todavía Authorization Code Flow ni Apple.

## Validación

- Un challenge consumido, vencido o cuyo nonce no coincide no crea sesión.
- `(issuer, subject)` solo puede pertenecer a una cuenta.
- El contrato y la migración separan challenges, identidades externas,
  credenciales locales y sesiones.

## Fuentes técnicas

- [Google: verificar el ID token en el servidor](https://developers.google.com/identity/gsi/web/guides/verify-google-id-token)
- [Google OpenID Connect API Reference](https://developers.google.com/identity/openid-connect/reference)
