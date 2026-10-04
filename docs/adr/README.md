# Architecture Decision Records

Los ADR conservan el contexto y las consecuencias de decisiones importantes. No
son actas extensas ni documentación de una tecnología.

## Convención

- Nombre: `NNNN-titulo-en-kebab-case.md`.
- Numeración: secuencial, cuatro dígitos; no reutilizar números.
- Idioma: español.
- Un ADR aceptado es inmutable en su decisión; las aclaraciones menores se añaden
  con fecha. Un cambio de decisión crea un ADR sucesor.
- El índice canónico está en [DECISIONS.md](../governance/DECISIONS.md).

## Ciclo de vida

`Propuesto → Aceptado | Rechazado → En revisión → Superado`

Solo el usuario cambia una propuesta a **Aceptado**. El autor registra la evidencia
de esa decisión en el campo correspondiente.

## Contenido obligatorio

- problema;
- contexto y restricciones;
- criterios;
- alternativas;
- comparación;
- recomendación;
- decisión del usuario;
- consecuencias;
- validación;
- disparadores de revisión;
- documentos afectados.

Usa [template.md](template.md) y el
[playbook de decisiones](../playbooks/decision-process.md).

- [ADR-0138: Esperar siete días salvo vulnerabilidades críticas](0138-wait-seven-days-except-critical-vulnerabilities.md) — Aceptado.

- [ADR-0145: Drenaje de la API antes de terminar el contenedor](0145-drain-api-before-container-termination.md) — Aceptado.
