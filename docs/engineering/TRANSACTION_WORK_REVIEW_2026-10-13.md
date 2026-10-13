# Revisión de trabajo transaccional — 2026-10-13

## Problema, autorización y solución

El usuario autoriza los ajustes recomendados y solicita medir las versiones
anterior y nueva. [ADR-0165](../adr/0165-bound-tournament-writes-and-expired-data-purges.md)
registra alternativas, coste y decisión. Se conservan las garantías de autorización,
atomicidad y bloqueo del torneo; no se despliega ni cambia el esquema.

Calendarios de liga y ciclos de desempate insertan hasta 256 partidos por
sentencia. Los resultados eliminatorios omiten UPDATE cuando sus campos
persistidos no cambian. Las purgas auxiliares respetan el límite recibido.

## Comparación ejecutada

- Anterior: commit `b64840c77b7916a68bd622e0ab557b0b60489553`, exportado mediante
  git archive en una copia temporal, sin modificar la rama ni el checkout actual.
- Nueva: código de ADR-0165, incluida la conversión de los arrays mediante int64
  que evita un estrechamiento inseguro antes del codificador PostgreSQL.
- PostgreSQL 18.6 Homebrew nativo arm64, base sintética aislada en localhost,
  shared_buffers=128MB, max_connections=100, fsync/synchronous_commit/full_page_writes
  activados. [Configuración registrada](evidence/transaction-work-2026-10-13/postgres-settings.log).
  Binarios y cluster temporales; no se instala un servicio.
  La relocalización solo adapta rutas de bibliotecas y recursos del paquete.
- Go 1.26.9; mismo pool runtime y sus límites de ADR-0163 en ambas versiones.
- Mismo probe sobre APIs públicas del repositorio, copiado al snapshot anterior.
  Primero se mide la anterior y después la nueva. Cada escenario tiene una
  muestra de calentamiento descartada y cinco muestras; se informa la mediana.
- Liga y eliminatoria nuevas para cada muestra, con 64 equipos; purga con 1000
  registros caducados por categoría auxiliar y límite 100. La preparación de
  datos se excluye del tiempo. Inicio y resultado incluyen commit y lectura de
  la respuesta completa. El tracer de prueba mide consultas sin exportar SQL,
  parámetros ni datos personales. No se activa infraestructura de observabilidad.

| Escenario | Anterior | Nueva | Trabajo SQL |
| --- | ---: | ---: | --- |
| Iniciar liga de 4032 partidos | 689,978 ms | 76,376 ms | 4032 INSERT → 16 INSERT |
| Registrar un resultado en cuadro de 64 equipos | 13,114 ms | 4,893 ms | 63 UPDATE de partidos → 2 UPDATE |
| Primera ejecución de purga | 1,114 ms | 1,370 ms | 3000 registros auxiliares → 300 |
| Purgar toda la misma acumulación | 1,114 ms / 1 ejecución | 10,000 ms / 10 ejecuciones | 3000 registros en ambas |

En la versión nueva, el tiempo mediano acumulado de las 16 inserciones es
61,396 ms: unos **3,84 ms por lote** de media, incluyendo el último lote de
192 partidos. La mediana del máximo de cada muestra es 4,192 ms. No son
percentiles de producción ni un límite prometido por lote. Iniciar la liga
resulta aproximadamente nueve veces más rápido en esta prueba.

La purga pequeña no acelera: seleccionar y ordenar candidatos, y repetir
sentencias para vaciar todo, añade coste. Su beneficio es acotar trabajo por
ejecución. El bucle de vaciado solo existe en el probe comparativo; producción
continúa la limpieza en futuras ejecuciones programadas.

Muestras completas: [anterior](evidence/transaction-work-2026-10-13/baseline.log),
[nueva](evidence/transaction-work-2026-10-13/updated.log) y
[medianas](evidence/transaction-work-2026-10-13/medians.json).

## Validación y reproducción

El probe `TestIntegrationTransactionPerformance` requiere
TM_MEASURE_TRANSACTIONS=1, TM_RUN_INTEGRATION=1 y una
TM_INTEGRATION_DATABASE_URL explícita de una base desechable. integrationPool
trunca cuentas y cascadas; el probe de purga trunca las tres tablas auxiliares.
Nunca se ejecuta contra local, dev ni producción con datos que se deban conservar.

Para comparar, exportar el backend del commit indicado en un directorio temporal,
copiar el mismo transaction_performance_integration_test.go a esa exportación
y ejecutar en ambas raíces:

```sh
TM_RUN_INTEGRATION=1 TM_MEASURE_TRANSACTIONS=1 \
  go -C apps/backend test ./internal/adapters/postgres \
  -run '^TestIntegrationTransactionPerformance$' -count=1 -v
```

La URL aislada se proporciona mediante el entorno, con el esquema inicial y
migraciones aplicados como en make test-integration. No ejecutar las dos versiones
en paralelo sobre la misma base.

Regresiones reales: fallo de FK en un segundo lote revierte también el primero
y la etapa; resultado nuevo escribe únicamente dos partidos y uno idéntico
ninguno, conservando dos entradas de historial; purgas repetidas acotadas mantienen
los registros vigentes. La suite existente cubre grupos impares a dos vueltas,
propagación y corrección del cuadro, sets, permisos, cancelaciones, conflictos y
concurrencia. Se conserva evidencia de [pruebas](evidence/transaction-work-2026-10-13/tests.log)
y [análisis estático](evidence/transaction-work-2026-10-13/lint.log).
Las [regresiones nuevas](evidence/transaction-work-2026-10-13/regressions.log) y
la [matriz de errores HTTP](evidence/transaction-work-2026-10-13/http-write-failures.log)
conservan evidencia independiente. Pasan la suite Go completa con integración
activada y detector de carreras, format-check-go y lint-go (cero incidencias).
PostgreSQL temporal se apaga tras las comprobaciones, conservando cluster y
evidencia; local/dev y observabilidad permanecen apagados.

## Límites y retrospectiva

Los lotes no liberan el bloqueo entre sentencias. Retirada, cierre, programación
y repetición todavía tienen trabajo bajo el bloqueo; este cambio no acredita
capacidad global bajo carga. Una base local caliente y cinco muestras secuenciales
no reproducen la VM, red, volumen histórico ni concurrencia de producción.
No se deducen p95, capacidad máxima ni SLO de estos resultados.

Se mantiene el tamaño 256 como punto de partida de bajo mantenimiento. Antes
de introducir COPY, colas o bloqueos más finos, medir carga concurrente y latencia
operativa. La purga acota entidades seleccionadas por categoría, no todas las
filas dependientes eliminadas por cascadas de cuentas. La retrospectiva distingue
aceleración del calendario y reducción de escrituras del objetivo de contención
de las purgas; no todas las optimizaciones mejoran el mismo indicador.
