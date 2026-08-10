# ADR-0078: Transferir el borrador local junto al alta

- **Estado:** Aceptado
- **Fecha:** 2026-08-09
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** ADR-0069, exclusivamente en la persistencia y transferencia del borrador
- **Superado por:** Ninguno

## Problema

Una persona que prepara una liga antes de tener cuenta debe conservar ese
borrador al crear y verificar su cuenta, y recuperarlo en futuros inicios de
sesión.

## Contexto y restricciones

- Antes del alta el borrador sigue siendo local y no requiere identidad ni
  acceso al servidor.
- El registro ya admite la ausencia de borrador.
- La cuenta pendiente no recibe sesión ni puede consultar recursos protegidos.
- OpenAPI contract-first y PostgreSQL como fuente de verdad continúan vigentes.

## Criterios de decisión

1. Conservar la preparación al cruzar la frontera de crear una cuenta.
2. Mantener un único modelo de liga, equipos y autorización.
3. No persistir borradores incompletos ni exponerlos públicamente.
4. Recuperar el borrador tras verificar o iniciar sesión desde cualquier instalación.

## Alternativas

### A — Transferir el borrador válido como liga privada en estado `draft`

- **Ventajas:** reutiliza liga y equipos; evita sincronización posterior.
- **Inconvenientes:** modelo y consultas deben distinguir `draft` de una liga publicada.
- **Coste de adopción:** medio: contrato, transacción, esquema, autorización, cliente y pruebas.
- **Coste de mantenimiento:** bajo: una entidad y ciclo explícito.
- **Riesgos:** listar o publicar un borrador antes de verificar.

### B — Tabla y endpoints independientes de borradores

- **Ventajas:** separación física de preparación y ligas.
- **Inconvenientes:** duplica validación, consultas, autorización y transición.
- **Coste de adopción y mantenimiento:** medio o alto.
- **Riesgos:** deriva entre modelos de equipos.

### No cambiar

- **Consecuencias:** el borrador solo sobrevive en la instalación original.

## Comparación

La alternativa A conserva el borrador entre dispositivos con un único modelo de
dominio. B no aporta una capacidad adicional proporcional a su coste.

## Recomendación

**Opinión/recomendación:** alternativa A, la solución mínima suficiente.

## Decisión del usuario

**Aceptada el 2026-08-09:** el borrador es local hasta que la persona envía el
alta. Si es válido, el cliente lo incluye en `RegisterRequest` y el servidor lo
persiste atómicamente junto a la cuenta pendiente. Tras verificar o iniciar
sesión aparece en los torneos administrados. Un alta sin borrador sigue siendo
válida.

## Consecuencias

### Positivas

- El borrador sobrevive a verificación, cambio de instalación e inicio de sesión.
- Publicarlo transforma el mismo recurso `draft` en `published`.

### Negativas y deuda aceptada

- El backend incorpora datos temporales de cuentas pendientes y los purgará con
  la cuenta expirada.

## Validación

- Alta con borrador válido: cuenta pendiente, liga `draft` y equipos en una transacción.
- Un borrador no se lee por enlace ni aparece antes de verificar.
- Tras verificar o iniciar sesión aparece solo en «Administro».
- Publicar conserva ID y equipos y cambia el estado a `published`.

## Disparadores de revisión

- Edición concurrente desde varios dispositivos.
- Necesidad de varios borradores con una organización distinta.

## Documentación afectada

- [Producto](../project/PRODUCT.md)
- [API](../engineering/API.md)
- [Modelo inicial de datos](../engineering/INITIAL_DATA_MODEL.md)
- [Decisiones](../governance/DECISIONS.md)
- [Aprendizaje](../project/LEARNING.md)
- [OpenAPI v1](../../contracts/openapi/v1/openapi.yaml)
