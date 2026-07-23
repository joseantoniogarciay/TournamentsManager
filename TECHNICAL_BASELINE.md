# Technical Baseline

> Estado: decision gate activo.
>
> Objetivo: confirmar la base técnica antes de diseñar el comportamiento detallado
> del producto. No autoriza scaffolding ni código por sí mismo.

## Reglas del gate

1. Se toma una decisión por vez, siguiendo sus dependencias.
2. Cada análisis incluye alternativas, ventajas, inconvenientes, mantenimiento y
   recomendación.
3. El usuario acepta la decisión.
4. Se registra el ADR.
5. Solo entonces se prepara el siguiente análisis.
6. Las versiones exactas se fijan al crear el entorno reproducible, con política
   de actualización, no como números aislados.
7. Una elección que necesite comportamiento de negocio se marca como bloqueada.

## Estado

| Orden | Decisión | Estado | Registro |
|---|---|---|---|
| 0 | Control de versiones: Git | Aceptada | [ADR-0003](docs/adr/0003-use-git-for-version-control.md) |
| 0 | Arquitectura clean/hexagonal pragmática | Aceptada | [ADR-0001](docs/adr/0001-pragmatic-clean-architecture.md) |
| 1 | Topología de repositorios | En decisión | Este documento |
| 2 | Topología del backend | Pendiente | Monolito modular / servicios / funciones |
| 3 | Estrategia web y mobile | Pendiente | Separadas / universal / web-first |
| 4 | Estilo y contrato de API | Pendiente | REST/OpenAPI / GraphQL / RPC |
| 5 | Arquitectura de identidad | Pendiente | Propia / gestionada / híbrida |
| 6 | Persistencia y acceso a datos | Pendiente | PostgreSQL, SQL y migraciones |
| 7 | Toolchain Go | Pendiente | Versión, módulo, formato, lint y análisis |
| 8 | Toolchain TypeScript | Pendiente | Runtime, package manager y workspaces |
| 9 | Framework web | Pendiente | Comparación tras decisión 3 |
| 10 | Framework mobile | Pendiente | Comparación tras decisión 3 |
| 11 | Configuración y secretos | Pendiente | Contrato local, CI y cloud |
| 12 | Entorno local | Pendiente | Procesos / Docker Compose |
| 13 | Estrategia de pruebas | Pendiente | Capas, herramientas y gates |
| 14 | Observabilidad mínima | Pendiente | Señales, OpenTelemetry y backends |
| 15 | CI y política de calidad | Pendiente | Checks, artefactos y seguridad |
| 16 | Contenedores y despliegue | Pendiente | Imagen, runtime y promoción |
| 17 | IaC y AWS | Pendiente | Terraform, cuentas, red y estado |
| 18 | Kubernetes local/cloud | Aplazada | Fase 4 del manifiesto |

“Pendiente” significa que no existe decisión. Las direcciones del manifiesto son
candidatos preferentes, no autorización para seleccionar librerías o crear
configuración.

## Decisión 1 — Topología de repositorios

### Problema

Backend, web, mobile, infraestructura, contratos y handbook evolucionarán a ritmos
distintos, pero pertenecen al mismo producto y inicialmente al mismo equipo. Hay
que decidir cuántos repositorios Git serán fuente de verdad.

### Criterios

En orden:

1. trazabilidad entre contrato, implementación y documentación;
2. simplicidad para aprender y operar;
3. reutilización controlada entre web y mobile;
4. independencia de builds y despliegues;
5. rendimiento y mantenibilidad del tooling;
6. posibilidad de separar componentes en el futuro.

### Alternativa A — Monorepo de producto

Un repositorio contiene handbook, backend, web, mobile e infraestructura.

**Ventajas**

- cambios atómicos de API, clientes, pruebas y documentación;
- un único historial, onboarding y política de calidad;
- compartir contratos no exige publicar paquetes desde el primer día;
- encaja con un equipo pequeño y una arquitectura inicialmente simple.

**Inconvenientes**

- varios ecosistemas conviven en el mismo árbol;
- CI debe detectar qué ha cambiado;
- puede tentar a compartir código o acoplar despliegues sin necesidad;
- si crece mucho, puede requerir tooling adicional.

**Mantenimiento**

Inicialmente bajo con Git, Go y workspaces del package manager. Nx, Turborepo u
otra capa se añadiría solo ante tiempos de build o coordinación medidos.

### Alternativa B — Un repositorio por aplicación

Repositorios separados para API, web, mobile e infraestructura.

**Ventajas**

- límites y pipelines independientes;
- permisos y releases aislados;
- menor mezcla de toolchains por repositorio.

**Inconvenientes**

- cambios de contrato coordinados entre varios historiales;
- paquetes compartidos necesitan versionado y distribución;
- más configuración, bots, permisos y mantenimiento;
- documentación transversal puede quedar desalineada.

**Mantenimiento**

Mayor desde el primer día: varios pipelines, dependencias entre repositorios y
estrategia de compatibilidad.

### Alternativa C — Híbrida

Un repositorio para backend/infra y otro para web/mobile, con handbook en uno de
ellos.

**Ventajas**

- separa Go/cloud del ecosistema TypeScript;
- web y mobile comparten packages con facilidad.

**Inconvenientes**

- el contrato cruza repositorios;
- la fuente de verdad documental resulta menos natural;
- mantiene buena parte de la coordinación de multirepo sin aislamiento total.

**Mantenimiento**

Intermedio; exige versionar el contrato o automatizar compatibilidad desde el
inicio.

### Recomendación

**Opinión:** alternativa A, monorepo de producto.

Es la solución con menor coste para un equipo pequeño, permite commits atómicos y
mantiene el handbook junto al sistema. Monorepo no significa un único build ni un
único despliegue: cada aplicación conservará dependencias, pruebas y artefactos
independientes.

La estructura concreta se decidirá después; aceptar monorepo no implica aceptar
Nx, Turborepo, pnpm ni una convención `apps/packages`.

### Decisión del usuario

**Pendiente.**

## Resultado del gate

Cuando todas las decisiones bloqueantes estén aceptadas tendremos:

- un mapa de aplicaciones y repositorios;
- contratos de integración;
- límites de identidad y datos;
- toolchains reproducibles;
- entorno local;
- estrategia de pruebas, seguridad y observabilidad;
- camino de CI, despliegue e infraestructura.

Después se reanudará el diseño de producto y se creará el scaffolding mínimo
compatible con ambos.
