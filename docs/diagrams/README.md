# Diagramas

Los diagramas explican relaciones que el texto no muestra con claridad. Deben ser
versionables, sencillos y enlazar la decisión o documento que representan.

## Reglas

- Mermaid como formato inicial cuando sea suficiente.
- Título, propósito, alcance y fecha de revisión.
- Etiquetas de tecnología solo si están decididas.
- No inventar componentes para completar visualmente el dibujo.
- Actualizar el diagrama junto al cambio.
- El texto normativo y los ADR prevalecen.

## Catálogo

| Diagrama                                               | Estado                                                | Fuente                                                                  |
| ------------------------------------------------------ | ----------------------------------------------------- | ----------------------------------------------------------------------- |
| Dirección conceptual de dependencias                   | Vigente                                               | [ARCHITECTURE.md](../engineering/ARCHITECTURE.md)                       |
| [Contexto de producto](system-context.md)              | Conceptual                                            | [PRODUCT.md](../project/PRODUCT.md)                                     |
| [Mapa entidad-relación de PostgreSQL](database-erd.md) | Vigente hasta migración `00008`                       | [`apps/backend/db`](../../apps/backend/db/)                             |
| Contenedores/componentes                               | No requiere diagrama separado en v1                   | [ARCHITECTURE.md](../engineering/ARCHITECTURE.md)                       |
| Despliegue local                                       | Cerrado y documentado textualmente                    | [DEPLOYMENT.md](../operations/DEPLOYMENT.md)                            |
| Despliegue Kubernetes                                  | Cerrado y documentado mediante manifiestos y runbooks | [Retrospectiva de Fase 4](../project/PHASE_4_RETROSPECTIVE.md)          |
| Despliegue AWS                                         | Cancelado                                             | [ADR-0128](../adr/0128-close-roadmap-after-k3s-and-cancel-aws-phase.md) |
