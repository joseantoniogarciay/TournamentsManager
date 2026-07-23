# ADR-0005: Usar un monorepo de producto

- **Estado:** Aceptado
- **Fecha:** 2026-07-23
- **Decisor:** Usuario, mediante confirmación explícita
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

Backend, web, mobile, infraestructura, contratos y handbook evolucionarán a ritmos
distintos, pero pertenecen inicialmente al mismo producto y equipo. Se necesita
una topología de repositorios que preserve trazabilidad sin añadir coordinación
prematura.

## Criterios

1. trazabilidad entre contrato, implementación y documentación;
2. simplicidad para aprender y operar;
3. reutilización controlada entre web y mobile;
4. independencia de builds y despliegues;
5. mantenimiento del tooling;
6. posibilidad de separar componentes en el futuro.

## Alternativas

### Monorepo de producto

Un único Git contiene handbook, backend, web, mobile e infraestructura. Permite
cambios atómicos y evita publicar contratos internos desde el primer día. Mezcla
varios toolchains y exige CI selectivo.

### Un repositorio por aplicación

Aísla permisos, pipelines y releases. Obliga a versionar contratos, coordinar
cambios y mantener varios repositorios desde el inicio.

### Topología híbrida

Separa backend/infra de web/mobile. Reduce parte de la mezcla de toolchains, pero
conserva la coordinación multirepo y deja menos clara la fuente documental.

## Decisión del usuario

Usar un monorepo de producto.

## Consecuencias

- Handbook, API, web, mobile e infraestructura compartirán repositorio.
- Cada aplicación conservará build, artefacto, versión, secretos y despliegue
  independientes.
- Los cambios en paquetes compartidos activarán la validación de sus consumidores.
- Monorepo no autoriza todavía Nx, Turborepo, pnpm ni una estructura concreta.
- La separación a otros repositorios seguirá siendo posible si aparecen equipos,
  permisos o ciclos de vida realmente independientes.

## Validación

La futura CI deberá demostrar que un cambio exclusivo de una aplicación no
construye ni despliega innecesariamente las demás.

## Disparadores de revisión

- Equipos con propiedad y permisos independientes.
- Historial o checkout inmanejable.
- Tooling de cambios afectados desproporcionadamente complejo.
- Contratos públicos con ciclos de versión independientes.

## Documentación afectada

- [TECHNICAL_BASELINE.md](../../TECHNICAL_BASELINE.md)
- [CONTRIBUTING.md](../../CONTRIBUTING.md)
- [SYSTEM_OPTIONS.md](../../SYSTEM_OPTIONS.md)
