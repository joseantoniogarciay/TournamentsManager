# Gobierno de decisiones

## Autoridad

El usuario toma la decisión final. El asistente y cualquier colaborador preparan
el análisis, diferencian hechos de opiniones y formulan una recomendación. Una
recomendación no se convierte en decisión por aparecer en un documento.

## Cuándo hace falta un ADR

Se crea un ADR si la elección:

- cambia límites arquitectónicos o la dirección de dependencias;
- introduce una tecnología, servicio o dependencia operativa relevante;
- afecta datos, seguridad, disponibilidad, coste o portabilidad;
- es difícil o costosa de revertir;
- establece una convención transversal;
- reemplaza una decisión aceptada.

Las decisiones locales y reversibles pueden documentarse junto al cambio. Si se
repiten o empiezan a condicionar otras áreas, se elevan a ADR.

## Estados

- **Propuesto:** análisis abierto; no autoriza implementación.
- **Aceptado:** decisión final del usuario.
- **Rechazado:** analizado y no elegido.
- **Superado:** reemplazado por otro ADR.
- **En revisión:** un disparador exige reconsideración.

Solo un ADR aceptado puede presentarse como “decisión tomada”.

## Proceso obligatorio

1. Definir el problema y el resultado que se necesita.
2. Aclarar restricciones, supuestos y criterios.
3. Comparar al menos dos alternativas razonables, incluido “no hacer nada” cuando
   aplique.
4. Explicar ventajas, inconvenientes, riesgos y coste de mantenimiento.
5. Separar estándar de industria, evidencia y opinión.
6. Presentar una recomendación.
7. Obtener la decisión explícita del usuario.
8. Registrar el ADR y actualizar los documentos afectados.
9. Definir cómo se validará y cuándo debe revisarse.

Playbook completo: [decision-process.md](docs/playbooks/decision-process.md).

## Índice de ADR

| ADR | Título | Estado | Fecha |
|---|---|---|---|
| [0000](docs/adr/0000-record-architecture-decisions.md) | Registrar decisiones arquitectónicas | Aceptado | 2026-07-23 |
| [0001](docs/adr/0001-pragmatic-clean-architecture.md) | Clean Architecture pragmática con principios hexagonales | Aceptado | 2026-07-23 |
| [0002](docs/adr/0002-handbook-before-code.md) | Construir el handbook antes que el código | Aceptado | 2026-07-23 |
| [0003](docs/adr/0003-use-git-for-version-control.md) | Usar Git para control de versiones | Aceptado | 2026-07-23 |
| [0004](docs/adr/0004-technical-baseline-before-product-design.md) | Confirmar la base técnica antes del diseño de producto | Aceptado | 2026-07-23 |
| [0005](docs/adr/0005-use-a-product-monorepo.md) | Usar un monorepo de producto | Aceptado | 2026-07-23 |
| [0006](docs/adr/0006-public-github-repository-security-boundary.md) | Publicar el monorepo en GitHub sin publicar secretos | Aceptado | 2026-07-23 |
| [0007](docs/adr/0007-use-a-modular-monolith-backend.md) | Usar un monolito modular para el backend | Aceptado | 2026-07-24 |
| [0008](docs/adr/0008-use-a-universal-react-native-client.md) | Usar un cliente universal con React Native | Aceptado | 2026-07-24 |

## Trazabilidad de un cambio

Toda propuesta importante debe enlazar:

`problema → análisis → decisión → cambio → prueba → documentación → aprendizaje`

Si falta un eslabón, el cambio no está terminado.
