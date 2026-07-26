# Retrospectiva técnica — Fase 0

- **Fecha:** 2026-07-26
- **Objetivo:** establecer documentación, proceso de decisión y baseline antes de
  código de producto.
- **Participantes:** Usuario y Codex.

## Resultado frente al objetivo

El manifiesto y el handbook son la fuente de trabajo; la base técnica está
aceptada y el Gate 0B define el primer vertical slice de una liga de fútbol.
Las decisiones de producto relevantes están registradas en ADR-0031 a ADR-0042.

Queda fuera de esta fase la implementación: entorno local, esquema, OpenAPI,
cliente y pruebas de producto pertenecen a las fases siguientes.

## Decisiones

- **Funcionaron:** decisiones pequeñas y secuenciales permitieron distinguir
  identidad, visibilidad, participación, administración, resultados y ciclo de
  vida sin acoplarlas a PostgreSQL o HTTP.
- **Coste inesperado:** la decisión inicial de congelar al publicar no reflejaba
  el flujo deseado; ADR-0040 la sucedió parcialmente y desplazó esa frontera al
  inicio de la liga.
- **Revisar:** reglas configurables de liga, incidencias tipadas, cambio de
  `username`, invitaciones con aceptación, notificaciones y prevención de abuso.
- **ADR ausentes:** ninguno para las decisiones que bloquean el primer vertical
  slice.

## Aprendizaje

- Una liga no se define solo por equipos: necesita estados, una frontera para
  generar partidos y reglas explícitas para resultados, bajas y cierre.
- Visibilidad, seguimiento, participación y autorización son relaciones
  diferentes; compartir un enlace no concede permisos.
- Un modelo pequeño puede preservar trazabilidad: los cambios de resultado se
  guardan aunque el estado actual se aplique de inmediato.

## Calidad profesional

- **Seguridad:** email no se usa para buscar usuarios; los enlaces no listados
  requieren identificadores aleatorios; no se han diseñado aún avisos ni datos
  sensibles adicionales.
- **Pruebas:** las validaciones de los ADR-0031 a ADR-0042 definen la evidencia
  mínima para las pruebas futuras.
- **Observabilidad y operación:** no hay producto desplegable todavía; las
  decisiones de plataforma aceptadas guían la fase de entorno local.
- **Coste:** no se han creado recursos cloud ni dependencias de notificaciones.
- **Documentación:** se actualizaron producto, identidad, roadmap, decisiones y
  aprendizaje junto a cada decisión.

## Complejidad

La solución se mantiene como liga de una vuelta, marcador local-visitante y
administración limitada a resultados. Se aplazan motor multideporte, calendario,
incidencias deportivas, jugadores, notificaciones y permisos avanzados.

## Acciones

| Acción | Propietario | Disparador | Destino |
| --- | --- | --- | --- |
| Preparar entorno local reproducible | Usuario/Codex | Inicio de Fase 1 | [ROADMAP.md](ROADMAP.md) |
| Diseñar esquema y OpenAPI del slice | Usuario/Codex | Tras salida de Fase 1 | ADR y documentos de ingeniería |
| Revisar decisiones aplazadas | Usuario/Codex | Disparadores registrados | [DECISIONS_TO_REVISIT.md](../governance/DECISIONS_TO_REVISIT.md) |

## Cierre

La Fase 0 queda cerrada en documentación y decisiones. El siguiente paso es la
Fase 1: entorno local reproducible; no se implementará negocio antes de preparar
esa base.
