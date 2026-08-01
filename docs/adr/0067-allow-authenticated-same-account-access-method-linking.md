# ADR-0067: Permitir añadir métodos de acceso solo a la cuenta autenticada

- **Estado:** Aceptado
- **Fecha:** 2026-08-01
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** ADR-0066, exclusivamente en la prohibición de vincular una identidad externa que todavía no pertenezca a ninguna cuenta
- **Superado por:** Ninguno

## Problema

Una persona debe poder usar contraseña y un proveedor social para una única
cuenta interna, sin convertir el registro o el login inicial en una fusión de
cuentas.

## Contexto y restricciones

- La propiedad de ligas y demás datos siempre pertenece a una única `account_id`.
- El email no es una clave de identidad ni autoriza una vinculación.
- El registro no debe enumerar emails ni métodos de acceso existentes.
- `(issuer, subject)` identifica de forma única una identidad de proveedor.

## Criterios de decisión

1. permitir varios métodos de acceso para una misma cuenta;
2. impedir mover una identidad o datos entre cuentas;
3. no revelar proveedores durante el registro;
4. mantener el flujo pequeño y verificable.

## Alternativas

### Alternativa A — No permitir añadir métodos

- **Ventajas:** menor superficie de identidad.
- **Inconvenientes:** obliga a usar siempre el método inicial.
- **Coste de mantenimiento:** bajo.

### Alternativa B — Añadir métodos desde una cuenta autenticada

- **Ventajas:** permite contraseña y proveedores sociales sin fusionar cuentas;
  el límite de propiedad es claro.
- **Inconvenientes:** requiere pantalla de Seguridad y reautenticación reciente.
- **Coste de mantenimiento:** bajo o medio.

### Alternativa C — Vincular durante registro o login por coincidencia de email

- **Ventajas:** menos pasos aparentes.
- **Inconvenientes:** enumera o confunde cuentas y abre el límite hacia fusiones.
- **Coste de mantenimiento:** medio o alto.

## Comparación

B satisface el objetivo de varios accesos sin usar el email para unir cuentas.
A impide una necesidad legítima y C reduce seguridad y aumenta estados. El coste
adicional de Seguridad es menor que operar vínculos entre cuentas.

## Recomendación

**Recomendación:** alternativa B.

## Decisión del usuario

**Aceptada el 2026-08-01:** alternativa B.

- Desde `Cuenta > Seguridad`, una persona con sesión debe reautenticarse con un
  método actual antes de añadir una contraseña o identidad social a su propia
  cuenta.
- Al añadir una identidad social, si `(issuer, subject)` ya pertenece a otra
  cuenta, la operación falla sin mover ni duplicar la identidad.
- Si ya pertenece a la misma cuenta, la operación es idempotente.
- Desde registro o login no se vincula por email. Ante conflicto, el cliente
  muestra un aviso genérico para iniciar sesión con el método habitual y añadir
  otro acceso desde Seguridad; no nombra proveedor alguno.
- No hay fusión de cuentas ni traslado de datos de torneos.

## Consecuencias

### Positivas

- Una sola cuenta puede combinar contraseña y Google, y después otros proveedores
  aceptados.
- La unicidad de `(issuer, subject)` evita que una identidad social habilite dos
  cuentas.
- El formulario público no revela qué cuenta o proveedor existe.

### Negativas y deuda aceptada

- Para añadir un método hay que acceder primero con uno existente.
- Cuentas distintas creadas por error permanecen separadas.

## Validación

- Una cuenta autenticada puede añadir una identidad social no vinculada.
- Un `sub` de otra cuenta devuelve conflicto y conserva ambas cuentas intactas.
- Registro y login con email en conflicto no revelan el proveedor existente.
- Ninguna vinculación cambia `account_id` en datos de torneos.

## Disparadores de revisión

- Evidencia de que el flujo de Seguridad impide recurrentemente recuperar acceso.
- Requisito de un proveedor o assurance que exija otro mecanismo de vinculación.

## Documentación afectada

- `docs/engineering/IDENTITY.md`
- `docs/engineering/INITIAL_DATA_MODEL.md`
- `docs/governance/DECISIONS.md`
- `contracts/openapi/v1/openapi.yaml`
