# Roadmap de aprendizaje e implementación

El roadmap es secuencial por aprendizaje, no necesariamente por entregas. No se
avanza de fase porque exista código: se avanza cuando se cumplen sus criterios de
salida y se completa la retrospectiva.

## Fase 0 — Documentación

**Objetivo:** establecer propósito, proceso y baseline técnica antes del código.

**Entregables:**

- manifiesto preservado y transcrito;
- handbook navegable;
- proceso ADR y plantillas;
- arquitectura acordada registrada;
- mapa de decisiones pendientes;
- plantillas de aprendizaje, playbook, runbook y retrospectiva.

**Salida:** todos los enlaces internos son válidos; las decisiones aceptadas se
distinguen de propuestas y el siguiente decision gate está identificado.

### Gate 0A — Confirmación de la base técnica

Antes de Fase 1 se decide:

- topología del repositorio;
- topología del backend;
- estrategia web/mobile;
- contrato API e identidad;
- persistencia y toolchains;
- configuración, entorno local y pruebas;
- observabilidad, CI y despliegue.

El estado y orden están en
[TECHNICAL_BASELINE.md](../governance/TECHNICAL_BASELINE.md).

### Gate 0B — Definición del primer vertical slice

Después de confirmar la base técnica se decide:

- formato inicial o inscripción/publicación sin competición;
- usuario, equipo o ambos como participante;
- visibilidad del torneo;
- mecanismo de incorporación.

Las preguntas quedan pausadas en [PRODUCT.md](PRODUCT.md).

## Fase 1 — Entorno local

**Objetivo:** crear un entorno reproducible que se parezca a producción sin
introducir Kubernetes antes de necesitarlo.

**Decisiones previas:** versiones soportadas, Docker Compose, gestión de
configuración y secretos, estrategia de migraciones, datos de desarrollo y
comandos de trabajo.

**Salida:** una persona puede preparar, arrancar, comprobar y limpiar el entorno
siguiendo documentación; los fallos comunes tienen guía de diagnóstico.

## Fase 2 — Backend

**Objetivo:** entregar un primer vertical slice de negocio en Go.

**Decisiones previas:** alcance, API, modelo de dominio, persistencia, errores,
validación y estrategia de pruebas.

**Salida:** caso de uso funcional de extremo a extremo, reglas de dependencia
verificadas, pruebas proporcionales al riesgo y documentación actualizada.

## Fase 3 — Observabilidad

**Objetivo:** poder explicar el estado del sistema y diagnosticar fallos.

**Decisiones previas:** señales y objetivos de servicio, instrumentación,
retención, stack mínimo y coste.

**Salida:** logs, métricas y trazas correlacionables para el vertical slice;
dashboard y runbook de una ruta crítica.

## Fase 4 — Kubernetes

**Objetivo:** aprender orquestación cuando el servicio ya sea operable.

**Decisiones previas:** necesidad, k3d, empaquetado, health checks, recursos,
configuración, secretos y estrategia de despliegue.

**Salida:** despliegue local reproducible, recuperación demostrada y comparación
documentada frente a Docker Compose.

## Fase 5 — Cloud

**Objetivo:** desplegar y operar en AWS mediante Terraform sin acoplar el dominio
al proveedor.

**Decisiones previas:** cuenta y seguridad, red, cómputo, datos, almacenamiento,
coste, backup, CI/CD y estrategia de rollback.

**Salida:** infraestructura reproducible, despliegue verificable, observabilidad,
presupuesto y procedimiento de recuperación documentados.

## Retrospectiva obligatoria

Cada fase termina usando
[phase-retrospective.md](../playbooks/phase-retrospective.md). Sus conclusiones
actualizan `LEARNING.md`, el handbook y, si procede, los ADR.
