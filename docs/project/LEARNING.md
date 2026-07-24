# Registro de aprendizaje

## Método

Para cada capacidad se sigue el ciclo:

1. explicar el fundamento con palabras propias;
2. comparar alternativas;
3. construir la versión mínima;
4. observar su comportamiento;
5. provocar y diagnosticar un fallo;
6. documentar lo aprendido y los límites;
7. revisarlo en la retrospectiva de fase.

## Mapa de competencias

| Área | Resultado demostrable | Estado |
|---|---|---|
| Arquitectura | Explicar límites, dependencias y trade-offs | En curso |
| Go | Construir y mantener un servicio idiomático | No iniciado |
| PostgreSQL | Diseñar, migrar y operar datos con criterio | No iniciado |
| API | Diseñar contratos evolutivos y observables | No iniciado |
| Testing | Elegir pruebas por riesgo y velocidad de feedback | No iniciado |
| Seguridad | Modelar amenazas y aplicar controles verificables | No iniciado |
| Contenedores | Crear un entorno local reproducible | No iniciado |
| Observabilidad | Diagnosticar con logs, métricas y trazas | No iniciado |
| Kubernetes | Desplegar, escalar y recuperar cargas | No iniciado |
| Terraform/AWS | Aprovisionar y operar infraestructura | No iniciado |

## Diario

### 2026-07-24 — Navegación del handbook

- **Aprendido:** mantener todos los documentos en la raíz aumenta visibilidad al
  principio, pero deja de escalar cuando oculta las unidades del monorepo.
- **Evidencia:** handbook agrupado por proyecto, gobierno, ingeniería y
  operaciones, con índices y enlaces validados.
- **Coste aceptado:** mantener rutas e índices como parte de cualquier movimiento
  documental.
- **Siguiente decisión:** arquitectura de identidad.

### 2026-07-24 — Contrato API y cliente generado

- **Aprendido:** la API HTTP es un adaptador del backend, no todo el backend;
  OpenAPI coordina el servidor Go y el consumidor TypeScript.
- **Evidencia:** ADR-0009 con REST contract-first, generación del cliente y
  límites respecto al dominio.
- **Coste aceptado:** mantener lint, generación y compatibilidad, especialmente
  con aplicaciones instaladas que se actualizan más tarde.
- **Siguiente decisión:** arquitectura de identidad.

### 2026-07-24 — Estrategia universal de cliente

- **Aprendido:** compartir producto y comportamiento no obliga a usar una
  presentación idéntica; responsive y adaptativo resuelven tamaños y capacidades
  de entrada diferentes.
- **Evidencia:** ADR-0008 con paridad funcional, límites de plataforma y
  disparadores de revisión.
- **Coste aceptado:** aislar excepciones web/native y validar por separado SEO,
  accesibilidad, rendimiento y releases.
- **Siguiente decisión:** estilo y contrato de API.

### 2026-07-24 — Topología técnica inicial

- **Aprendido:** monorepo no significa un único pipeline, y monolito no significa
  ausencia de límites.
- **Evidencia:** ADR de monorepo, publicación segura y monolito modular.
- **Coste aceptado:** disciplina para mantener módulos y pipelines independientes.
- **Siguiente decisión:** estrategia web/mobile y frontera de reutilización.

### 2026-07-23 — Fundación documental

- **Aprendido:** una arquitectura profesional comienza por explicitar autoridad,
  proceso, estados y criterios de salida.
- **Evidencia:** manifiesto transcrito, ADR iniciales, mapa del handbook y
  plantillas operativas.
- **Incertidumbre:** requisitos del producto y decisiones de implementación.
- **Siguiente experimento:** definir el alcance y primer caso de uso antes del
  entorno o del backend.

## Regla de evidencia

“Entendido” exige una explicación propia y una demostración. Un comando que
funciona o una respuesta del asistente no son evidencia suficiente por sí solos.
