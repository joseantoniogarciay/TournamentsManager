# Decisiones y supuestos a revisar

Este registro evita convertir elecciones provisionales en permanentes. No duplica
ADR: enlaza la decisión y define qué evidencia obliga a reabrirla.

| Tema | Estado actual | Disparador de revisión | Documento |
|---|---|---|---|
| Redis o Valkey | Candidato: Redis; Valkey por evaluar | Antes de introducir cache | [DATABASE.md](DATABASE.md) |
| Stack de observabilidad | Lista objetivo por evaluar | Inicio de Fase 3 | [OBSERVABILITY.md](OBSERVABILITY.md) |
| Kubernetes/k3d | Dirección de aprendizaje | Antes de Fase 4 y tras medir Compose | [DEPLOYMENT.md](DEPLOYMENT.md) |
| AWS | Cloud inicial objetivo | Inicio de Fase 5 o restricción de coste | [DEPLOYMENT.md](DEPLOYMENT.md) |
| MinIO/S3 | Dirección objetivo | Antes del primer caso de uso con objetos | [ARCHITECTURE.md](ARCHITECTURE.md) |
| Frontend | Sin decidir | Cuando exista un flujo que necesite interfaz | [ROADMAP.md](ROADMAP.md) |
| Cloud agnostic | Principio | Cuando una abstracción añada más coste que portabilidad | [ARCHITECTURE.md](ARCHITECTURE.md) |

## Cómo revisar

1. Cambiar el ADR afectado a **En revisión** o crear una propuesta sucesora.
2. Recoger evidencia del disparador.
3. Repetir el proceso de comparación.
4. Pedir decisión explícita al usuario.
5. Marcar el ADR anterior como **Superado** si cambia la decisión.
6. Actualizar este registro y los documentos dependientes.
