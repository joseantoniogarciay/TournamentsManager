# ADR-0050: Incluir login federado con Google en el primer incremento

- **Estado:** Aceptado
- **Fecha:** 2026-07-27
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** ADR-0043 y ADR-0044, exclusivamente en la exclusión de Google
- **Superado por:** Ninguno

## Problema

El primer incremento demuestra identidad local, pero no el modelo aceptado de una
cuenta interna con varios métodos de acceso. Se necesita incluir un proveedor
federado real para hacer visible la relación entre credencial local, identidad
externa y sesión propia sin añadir simultáneamente la complejidad específica de
Apple.

## Contexto y restricciones

- ADR-0010 acepta identidad propia federada y prohíbe vincular cuentas
  automáticamente por coincidencia de email.
- Una sesión siempre pertenece a una cuenta interna y se mantiene en
  `sessions`; una credencial local es opcional.
- Google establece que el backend debe validar los tokens recibidos y usar `sub`,
  no el email, como identificador estable de la cuenta Google.
- Apple, recuperación de contraseña, MFA y gestión multidispositivo siguen fuera
  del incremento.

## Criterios de decisión

1. demostrar un login federado completo con una sesión propia revocable;
2. conservar el modelo extensible a Apple sin implementar Apple todavía;
3. no usar email como identificador ni como autorización de vinculación;
4. limitar configuración externa, superficies de error y pruebas iniciales.

## Alternativas

### Alternativa A — Mantener solo identidad local

- **Ventajas:** menor alcance inmediato.
- **Inconvenientes:** no valida el vínculo entre proveedor, cuenta y sesión.
- **Coste de adopción:** bajo.
- **Coste de mantenimiento:** bajo.
- **Riesgos:** aplazar el principal límite de la identidad federada aceptada.

### Alternativa B — Google como único proveedor federado inicial

- **Ventajas:** demuestra el flujo real con una sola configuración de proveedor;
  el modelo queda preparado para Apple.
- **Inconvenientes:** requiere validar credenciales de Google, configurar
  audiencias por cliente y resolver vinculación explícita con cuentas locales.
- **Coste de adopción:** medio.
- **Coste de mantenimiento:** medio.
- **Riesgos:** aceptar claims sin verificación completa o usar email como clave.

### Alternativa C — Google y Apple desde el primer incremento

- **Ventajas:** paridad inmediata de proveedores.
- **Inconvenientes:** duplica configuración, validación por plataforma y casos
  de prueba antes de demostrar el flujo común.
- **Coste de adopción:** alto.
- **Coste de mantenimiento:** alto inicialmente.
- **Riesgos:** retrasar el vertical slice por particularidades de Apple.

### No cambiar

- **Consecuencias:** el modelo federado permanece documental y no se valida en el
  primer recorrido ejecutable.

## Comparación

La alternativa A conserva el corte más pequeño, pero no muestra el mapa de
identidad buscado. C entrega más paridad de la necesaria y adelanta riesgos de
plataforma. B incorpora el límite común —identidad externa resuelta a cuenta y
sesión propia— con el menor número de proveedores.

## Recomendación

**Recomendación:** alternativa B. Google permite aprender y probar el flujo
federado completo; Apple puede añadirse después sobre las mismas fronteras.

## Decisión del usuario

**Aceptada el 2026-07-27:** el primer incremento incluye login federado con
Google. Apple queda fuera del incremento.

- Una identidad externa de Google se identifica por el par normalizado
  `(issuer, subject)` y se vincula a una sola cuenta interna.
- Tras validar una credencial de Google, el backend resuelve esa identidad y
  emite la sesión opaca propia en `sessions`; no crea `local_credentials`.
- Una primera entrada Google sin cuenta candidata crea una cuenta interna tras
  recoger el `username` requerido.
- Si el email recibido coincide con una cuenta local, no se vincula de forma
  automática: se mantiene el desafío de confirmación fresca de ADR-0010.
- Antes de implementar el adaptador, OpenAPI y el threat model concretarán el
  artefacto OIDC, nonce, CSRF web, audiencias por plataforma y librería de
  verificación.

## Consecuencias

### Positivas

- El primer incremento muestra claramente que autenticación federada y sesión
  propia son responsabilidades distintas.
- `local_credentials` permanece opcional y el modelo admite Apple después.
- La autorización de ligas continúa dependiendo solo de la cuenta y la sesión.

### Negativas y deuda aceptada

- El incremento requiere configuración de OAuth/OIDC y credenciales de Google
  por entorno y plataforma antes de ejecutar el flujo real.
- El caso de vinculación con una cuenta local exige persistencia y entrega de un
  desafío de un solo uso.
- Apple conserva su trabajo específico para un incremento posterior.

## Validación

- Una identidad Google ya vinculada inicia una sesión propia sin contraseña.
- Una cuenta local sin identidad Google no se vincula solo por coincidir email.
- Una credencial con firma, emisor, audiencia, vencimiento o nonce inválidos no
  crea cuenta, vínculo ni sesión.
- El modelo y contrato que se aprueben para implementar Google diferencian
  `external_identities` de `sessions` y de `local_credentials`.

## Disparadores de revisión

- Se necesita Apple u otro proveedor.
- Google cambia materialmente su mecanismo de credenciales o validación.
- Un incidente evidencia que el flujo de vinculación, nonce o CSRF es
  insuficiente.

## Documentación afectada

- [Identidad](../engineering/IDENTITY.md)
- [Producto](../project/PRODUCT.md)
- [Roadmap](../project/ROADMAP.md)
- [Primer incremento](0043-deliver-publish-and-read-league-first-backend-increment.md)
- [Sesiones y autenticación local](0044-use-opaque-sessions-local-smtp-and-argon2id.md)
- [Decisiones](../governance/DECISIONS.md)
- [Aprendizaje](../project/LEARNING.md)

## Fuentes técnicas

- [Google OpenID Connect API Reference](https://developers.google.com/identity/openid-connect/reference)
- [Verificar el ID token de Google en el servidor](https://developers.google.com/identity/gsi/web/guides/verify-google-id-token)
