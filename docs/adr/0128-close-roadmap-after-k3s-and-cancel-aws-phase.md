# ADR-0128: Cerrar el roadmap tras K3s y cancelar la Fase AWS

- **Estado:** Aceptado
- **Fecha:** 2026-09-18
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera parcialmente a:** ADR-0088 y ADR-0101, solo en su previsión de
  laboratorios AWS/EKS posteriores
- **Superado por:** Ninguno

## Problema

El roadmap reservaba una Fase 5 para laboratorios AWS después de completar el
runtime doméstico K3s. Con la Fase 4 cerrada, el objetivo de aprender los
conceptos de Kubernetes ya se cumplió y AWS añadiría coste, configuración y
mantenimiento sin una necesidad actual de producto u operación.

## Contexto y restricciones

- `prod` opera en K3s doméstico con publicación, observabilidad, backups y
  recuperación demostrada.
- No se creó ninguna cuenta, recurso, estado remoto ni gasto AWS.
- Las decisiones AWS ya aceptadas conservan valor documental como alternativas
  evaluadas, pero no autorizan ni exigen su implementación.
- La lógica de negocio continúa independiente de proveedor e infraestructura.

## Criterios de decisión

1. preservar el aprendizaje útil ya conseguido;
2. evitar coste y mantenimiento sin demanda demostrada;
3. no convertir una dirección futura en trabajo obligatorio;
4. conservar una vía explícita y reversible si cambian las necesidades.

## Alternativas

### A — Continuar con la Fase 5 AWS

- **Ventajas:** práctica directa de IAM, Terraform, red y servicios AWS.
- **Inconvenientes:** abre cuentas, permisos, presupuesto y operación ajenos a
  la necesidad actual.
- **Coste de adopción y mantenimiento:** medio o alto.

### B — Cancelar la Fase 5 y cerrar el roadmap tras K3s

- **Ventajas:** concentra el tiempo en profundizar capacidades ya útiles y evita
  gasto cloud sin propósito inmediato.
- **Inconvenientes:** no se adquiere experiencia práctica de AWS en este
  proyecto.
- **Coste de adopción y mantenimiento:** bajo.

### No cambiar

- **Consecuencias:** AWS seguiría apareciendo como siguiente fase, creando una
  expectativa de trabajo y coste que ya no responde al objetivo del usuario.

## Comparación

La alternativa A aporta aprendizaje adicional, pero no satisface una necesidad
operativa actual. B mantiene el runtime funcional y elimina el coste de una
plataforma que no se va a usar. No cambiar deja el roadmap desalineado.

## Recomendación

**Recomendación:** B. La solución mínima suficiente es cerrar el itinerario al
cumplir el objetivo de K3s y reabrir cloud solo ante evidencia nueva.

## Decisión del usuario

**Aceptada el 2026-09-18:** se cancela la Fase 5 de AWS. El roadmap activo
termina al cerrar la Fase 4. No se crearán ni configurarán recursos AWS,
Terraform ni HCP Terraform para este proyecto. El usuario profundizará en otros
temas sobre la base de Kubernetes ya aprendida.

## Consecuencias

### Positivas

- no se incurre en coste cloud ni mantenimiento de cuentas, red o estado remoto;
- el roadmap refleja el objetivo real de aprendizaje;
- K3s doméstico permanece como único runtime de `prod`.

### Negativas y deuda aceptada

- AWS, EKS, IAM y Terraform no cuentan con práctica ejecutada en este proyecto;
- una necesidad futura de cloud requerirá retomar el análisis y la autorización
  de coste desde cero.

## Validación

- roadmap, handbook y despliegue no presentan AWS como trabajo pendiente;
- no se crean recursos, cuentas ni estado cloud;
- las decisiones previas quedan identificadas como históricas, no activas.

## Disparadores de revisión

- requisito de disponibilidad, carga, cumplimiento o residencia que el runtime
  doméstico no pueda cubrir;
- pérdida o coste excesivo del hosting doméstico;
- decisión explícita del usuario de volver a aprender o adoptar cloud.

## Documentación afectada

- [Roadmap](../project/ROADMAP.md)
- [Despliegue](../operations/DEPLOYMENT.md)
- [Decisiones](../governance/DECISIONS.md)
- [Decisiones a revisar](../governance/DECISIONS_TO_REVISIT.md)
- [Aprendizaje](../project/LEARNING.md)
