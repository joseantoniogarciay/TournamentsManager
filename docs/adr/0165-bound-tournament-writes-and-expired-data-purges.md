# ADR-0165: Agrupar calendarios y acotar escrituras y purgas

- **Estado:** Aceptado
- **Fecha:** 2026-10-13
- **Decisor:** Usuario: «Pues hagamos los ajustes que dices», después de revisar
  inserciones de calendario, escrituras de partidos sin cambios y purgas.

## Problema y contexto

Iniciar una liga de 64 equipos a dos vueltas ejecuta 4032 inserciones secuenciales
bajo el bloqueo del torneo. Registrar un resultado eliminatorio actualiza todo
el cuadro aunque muchas filas no cambien. PurgeExpired limita las cuentas, pero
no todos los registros auxiliares caducados. Son riesgos observados en código;
no hay una medición de saturación productiva. El gate técnico está cerrado.

## Alternativas y coste

- Inserciones individuales: simples, pero acumulan viajes a PostgreSQL.
- INSERT con arrays y UNNEST en lotes: SQL fijo, parámetros tipados y una única
  transacción; mantenimiento bajo con pgx existente.
- COPY: eficiente para grandes cargas, con un protocolo y mapeo adicional que
  no se necesitan para este máximo de 4032 partidos.
- Colas o commits parciales: mayor operación o pérdida de atomicidad; no cumplen
  la garantía actual de iniciar el torneo completo.

## Recomendación y decisión aceptada

- Insertar calendarios de liga y ciclos de desempate en lotes de hasta 256
  partidos mediante SQL parametrizado. Una liga máxima necesita 16 inserciones.
  Todos los lotes pertenecen a la transacción existente: no hay commits parciales.
- Evitar consultas UPDATE de partidos eliminatorios cuya proyección persistida
  no cambia. Mantener historial y sets del resultado solicitado.
- Aplicar el límite de purga también a eventos RISC, desafíos Apple y evidencia
  legal caducada. Cada ejecución procesa como máximo ese límite por categoría;
  futuras ejecuciones continúan el trabajo, sin bucles que vacíen toda la cola.
- Mantener autorización, bloqueo por torneo, contrato HTTP, deadlines, retención
  legal y durabilidad. No añadir índices, dependencias ni infraestructura.

256 es un punto de partida de ingeniería, no un SLO ni un óptimo medido. La
duración de cada lote depende de la conexión, índices, restricciones y carga.
Los límites de ADR-0163 continúan vigentes.

## Validación y revisión

Comprobar calendario máximo y grupos impares, rollback tras fallar un lote
posterior, propagación y corrección de eliminatorias con historial, y purgas
acotadas que conservan registros vigentes. Revisar el tamaño ante evidencia de
latencia o contención, antes de cambiar el bloqueo de autorización.

Documentar evidencia ejecutada y retrospectiva en DATABASE.md y LEARNING.md.
No se autoriza publicación ni despliegue por este ADR.
