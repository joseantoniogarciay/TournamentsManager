# Decisiones y supuestos a revisar

Este registro evita convertir elecciones provisionales en permanentes. No duplica
ADR: enlaza la decisión y define qué evidencia obliga a reabrirla.

| Tema | Estado actual | Disparador de revisión | Documento |
|---|---|---|---|
| Redis o Valkey | Candidato: Redis; Valkey por evaluar | Antes de introducir cache | [DATABASE.md](../engineering/DATABASE.md) |
| Stack de observabilidad | Lista objetivo por evaluar | Inicio de Fase 3 | [OBSERVABILITY.md](../operations/OBSERVABILITY.md) |
| Kubernetes/k3d | Dirección de aprendizaje | Antes de Fase 4 y tras medir Compose | [DEPLOYMENT.md](../operations/DEPLOYMENT.md) |
| AWS | Cloud inicial objetivo | Inicio de Fase 5 o restricción de coste | [DEPLOYMENT.md](../operations/DEPLOYMENT.md) |
| MinIO/S3 | Dirección objetivo | Antes del primer caso de uso con objetos | [ARCHITECTURE.md](../engineering/ARCHITECTURE.md) |
| Cliente universal React Native | Aceptado en ADR-0008; framework pendiente | Divergencia sustancial, o incumplimiento de SEO, accesibilidad o rendimiento web | [ADR-0008](../adr/0008-use-a-universal-react-native-client.md) |
| Identidad propia federada | Aceptada en ADR-0010 | Incidente, coste operativo excesivo o requisitos de assurance superiores | [ADR-0010](../adr/0010-own-identity-with-federated-login.md) |
| PostgreSQL + pgx + sqlc + goose | Aceptada en ADR-0011 | Consultas dinámicas inmantenibles, requisito de otro motor o migraciones que necesiten tooling superior | [ADR-0011](../adr/0011-use-postgresql-pgx-sqlc-and-goose.md) |
| Cloud agnostic | Principio | Cuando una abstracción añada más coste que portabilidad | [ARCHITECTURE.md](../engineering/ARCHITECTURE.md) |

## Cómo revisar

1. Cambiar el ADR afectado a **En revisión** o crear una propuesta sucesora.
2. Recoger evidencia del disparador.
3. Repetir el proceso de comparación.
4. Pedir decisión explícita al usuario.
5. Marcar el ADR anterior como **Superado** si cambia la decisión.
6. Actualizar este registro y los documentos dependientes.
