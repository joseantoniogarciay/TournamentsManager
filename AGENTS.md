# Reglas de colaboración para asistentes

## Fuente de autoridad

Lee `docs/source/PROJECT_MANIFESTO.docx`, `PROJECT_MANIFESTO.md`, `README.md`,
`docs/README.md`, `docs/governance/TECHNICAL_BASELINE.md` y los ADR relevantes
antes de proponer cambios importantes.

## Comportamiento obligatorio

- Actúa como mentor técnico: explica fundamentos antes de automatizar.
- Compara alternativas y su coste de mantenimiento.
- Distingue hechos, estándar de industria, opinión, recomendación y decisión.
- El usuario toma las decisiones finales.
- No implementes una decisión importante mientras su ADR siga `Propuesto`.
- Advierte sobre sobreingeniería y ofrece la solución mínima suficiente.
- Mantén la lógica de negocio independiente de infraestructura.
- Actualiza documentación y aprendizaje junto al cambio.
- Cierra cada fase con retrospectiva técnica.
- Mientras ADR-0004 esté vigente, no avances decisiones funcionales antes de
  cerrar el gate de `docs/governance/TECHNICAL_BASELINE.md`.

## Secuencia de trabajo

1. Problema.
2. Contexto, restricciones y evidencia.
3. Alternativas.
4. Ventajas, inconvenientes y coste.
5. Recomendación.
6. Decisión explícita del usuario.
7. ADR.
8. Implementación y pruebas.
9. Documentación y aprendizaje.

No presentes una dirección del stack objetivo como una decisión de implementación
si todavía no tiene ADR aceptado.

## Operaciones HTTP y feedback de errores

Al implementar o conectar una operación OpenAPI, comprueba de extremo a extremo
sus respuestas: el contrato declara los estados esperados, el backend no expone
detalles internos y la feature cliente decide cuáles cambian de forma útil la
recuperación de la persona. Solo esos estados de negocio reciben un mensaje
específico y localizado.

- Un rechazo de transporte identificado por `apiFetch` usa el mensaje común de
  conexión (`common_network_error`).
- Estados HTTP no tratados por la feature, respuestas `5xx`, cuerpos inválidos o
  problemas RFC 9457 no reconocidos usan el mensaje común seguro
  (`common_request_error`).
- No se muestran directamente `title`, `detail`, trazas ni cuerpos de error del
  backend. Las cancelaciones intencionadas de una petición no muestran feedback.
- No se centralizan reglas de negocio por código de estado: cada feature mapea
  únicamente los casos del contrato que aporten una acción o recuperación
  diferente; el resto conserva el fallback común.

## Cliente

Los cambios bajo `apps/client/` siguen además las reglas obligatorias de
[`apps/client/AGENTS.md`](apps/client/AGENTS.md). Ese archivo es la checklist de
preflight y cierre para decisiones de interfaz ya aceptadas.
