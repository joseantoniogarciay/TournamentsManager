# ADR-0052: Hacer opcional el borrador en el alta

- **Estado:** Aceptado
- **Fecha:** 2026-07-27
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** ADR-0048, exclusivamente en la obligatoriedad del borrador al alta
- **Superado por:** Ninguno

## Problema

`RegisterRequest` exige actualmente `draft`, aunque una persona debe poder crear
una cuenta sin haber empezado una liga. Esta discrepancia bloquea un alta válida.

## Contexto y restricciones

- ADR-0048 sigue exigiendo `email`, `password` y `username` en el alta local.
- ADR-0031 conserva el borrador previo al acceso cuando la persona decide
  transferirlo al servidor.
- `DraftInput` ya exige `name` y entre dos y 64 equipos; cada `TeamInput` exige
  un nombre no vacío.
- OpenAPI es la fuente de verdad del contrato y el cliente TypeScript se genera
  desde ella.

## Criterios de decisión

1. permitir crear una cuenta sin liga;
2. no aceptar borradores incompletos;
3. conservar un contrato y cliente generados coherentes;
4. no añadir estados ni endpoints innecesarios.

## Alternativas

### Alternativa A — Mantener `draft` obligatorio

- **Ventajas:** un único recorrido de alta y borrador.
- **Inconvenientes:** impide el alta sin una liga preparada.
- **Coste de adopción:** nulo.
- **Coste de mantenimiento:** bajo, a costa de una restricción funcional incorrecta.
- **Riesgos:** clientes obligados a enviar datos ficticios o inválidos.

### Alternativa B — `draft` opcional y validación completa cuando exista

- **Ventajas:** permite ambas entradas reales; reutiliza las restricciones ya
  expresadas por `DraftInput`; no añade estado persistente.
- **Inconvenientes:** el backend debe distinguir ausencia de objeto de objeto
  presente e inválido.
- **Coste de adopción:** bajo: cambiar el `required` de `RegisterRequest` y
  regenerar el cliente.
- **Coste de mantenimiento:** bajo: las reglas del borrador siguen centralizadas
  en un único esquema.
- **Riesgos:** tratar erróneamente un objeto parcial como ausencia.

### Alternativa C — Aceptar borradores parciales

- **Ventajas:** menos fricción aparente durante el alta.
- **Inconvenientes:** introduce un segundo modelo de borrador y desplaza
  invariantes hasta la publicación.
- **Coste de adopción:** medio.
- **Coste de mantenimiento:** medio o alto por validación y estados adicionales.
- **Riesgos:** datos incompletos persistidos y reglas ambiguas.

## Comparación

**Hecho:** OpenAPI aplica `required` del objeto padre de forma independiente de
las restricciones del esquema referenciado. Por ello, retirar `draft` de
`RegisterRequest.required` no relaja `DraftInput.required`, `minItems` ni los
límites de sus nombres.

La alternativa B satisface los cuatro criterios sin el estado adicional de C;
A mantiene una limitación que no representa el flujo deseado.

## Recomendación

**Opinión/recomendación:** alternativa B. Es la solución mínima suficiente:
ausencia de borrador significa “no transferir ninguno”; presencia significa
validación completa del borrador existente.

## Decisión del usuario

**Aceptada el 2026-07-27:** `draft` deja de ser obligatorio en
`RegisterRequest`. Si se incluye, debe contener `name` y al menos dos equipos,
con un `name` válido para cada equipo, conforme a `DraftInput` y `TeamInput`.

## Consecuencias

### Positivas

- Se puede crear una cuenta sin preparar una liga.
- El contrato sigue rechazando borradores parciales o con menos de dos equipos.
- El tipo generado expresa que `draft` puede omitirse.

### Negativas y deuda aceptada

- La futura implementación HTTP debe validar la entrada en runtime; los tipos
  TypeScript no sustituyen la validación del servidor.

## Validación

- OpenAPI pasa el lint.
- El cliente generado declara `draft?: DraftInput`.
- El contrato acepta un alta sin `draft` y rechaza un borrador enviado sin
  `name`, con menos de dos equipos o con un equipo sin nombre.

## Disparadores de revisión

- Se necesita guardar borradores incompletos de forma deliberada.
- El alta separa la creación de cuenta de la transferencia de borrador mediante
  un endpoint posterior.

## Documentación afectada

- [OpenAPI v1](../../contracts/openapi/v1/openapi.yaml)
- [API](../engineering/API.md)
- [Producto](../project/PRODUCT.md)
- [Decisiones](../governance/DECISIONS.md)
