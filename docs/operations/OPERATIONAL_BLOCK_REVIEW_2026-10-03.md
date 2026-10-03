# Revisión del bloque operativo — 2026-10-03

## Problema y decisión vigente

Tras conciliar el workspace quedaban sin integrar retención, observabilidad,
herramientas locales, restauración de backups y su evidencia documental. El gate
técnico está cerrado; ADR-0139 y ADR-0140 están aceptados y ADR-0108 y ADR-0112
amparan backup y empaquetado de observabilidad. No se abre una decisión funcional.

**Recomendación aplicada:** revisar el bloque existente y validarlo con imágenes
fijadas y destinos temporales. Añadir otro servicio o framework de pruebas
incrementaría mantenimiento sin resolver una necesidad. La integración local en
Git y la promoción al runtime siguen siendo operaciones distintas.

## Evidencia de validación

- `make verify` aprobado, incluyendo las seis regresiones nuevas de
  `make test-operational-safety`. Formato, lint, tipos, contrato, pruebas,
  exportación web, generación, módulos, build y análisis de vulnerabilidades
  terminan correctamente. La integración PostgreSQL de ese target se omite al
  no definir `TM_INTEGRATION_DATABASE_URL`; no se presenta como ejecutada.
- Los tres Compose pasan `config --quiet`, sin mostrar valores resueltos.
  Loki 3.5.1 y ambos Promtail 3.5.0 validan configuración; Tempo 2.8.2 pasa
  `-config.verify=true`.
- En un Loki temporal, una entrada reciente recibe `204` y una de 25 horas
  recibe `400`. Promtail descarta el probe correcto y conserva un probe fallido,
  un evento de aplicación y JSON inválido. Esto comprueba admisión y filtrado,
  no una purga física instantánea a las 24 horas.
- Helm 3.18.6 temporal, con SHA-256 oficial comprobado, renderiza los charts
  fijados de Alloy 1.12.1, Loki 18.11.7 y Prometheus 29.27.0. Las configuraciones
  renderizadas validan con Alloy 1.19.2 y Loki 3.7.7. Las siete fixtures atraviesan
  el pipeline Alloy renderizado y un Loki temporal: categorías y descarte
  coinciden con lo esperado. No se usa Kubernetes real para estas fixtures.
- Promtool de Prometheus 3.14.0, versión renderizada del chart, valida la
  configuración y sus once reglas. Las nueve suites de almacenamiento cubren
  normalidad, aviso, crítico, medición ausente o vencida, bloques vencidos y
  compactor ausente o detenido. No se fuerza falta de espacio en producción.
- Las seis regresiones de seguridad comprueban conservación ante datos recientes,
  contenedor existente y fallo de inspección; alcance exacto local/dev,
  inspección sin red y en lectura, rechazo de `prod` y ausencia de borrado
  forzado ante una carrera de conexión de un contenedor. La medición del host
  falla ante volúmenes ausentes/ambiguos o timeout; no publica un falso cero.
- En PostgreSQL 18.4 temporal se aplica el esquema inicial y el DDL de las
  migraciones mediante `local-migrate.sh`, adaptando únicamente `/app` a la
  ruta host. Goose llega a versión 19, la segunda ejecución no aplica cambios
  y conserva una fila de prueba previa. No se ejecuta `make dev-migrate`
  sobre una base activa ni se validan de nuevo todos los adaptadores PostgreSQL.
- `verify-backup-restore.py` vuelve a restaurar la incremental
  `20261003-154848F_20261003-155132I`: recuperación terminada y agregados iguales
  a dev, esquema 19. El repositorio se monta en lectura, sin red ni datos
  activos; contenedor y volumen temporal se retiran. La comparación de conteos
  no acredita igualdad de cada fila ni PITR a cualquier instante.

## Comprobación de producción en solo lectura

Aproximadamente a las 18:39 CEST se consultaron Prometheus y el host por SSH:

- Las siete alertas `production-storage` tienen salud `ok` y estado `inactive`.
- La antigüedad mínima de TSDB es 25 722 segundos (unas 7 h 9 min), frente a los
  bloques del 24–25 de septiembre observados por la mañana. La condición de
  bloques vencidos ya no está presente; no se borraron manualmente.
- Loki informa de un ciclo de retención correcto hace 64 segundos.
- El timer está activo y `systemd-analyze verify` acepta ambas unidades.
- Espacio disponible: 15 504 764 928 bytes; tamaños agregados: Loki
  117 932 032, Tempo 8 155 136 y Prometheus 2 449 408 bytes.

No se han aplicado cambios, reiniciado servicios ni enviado notificaciones en
esta comprobación. Retención, rotación y alertas no convierten local-path en
una cuota física. La eliminación de Loki sigue siendo asíncrona.

## Integración y límites de cierre

Este cierre integra en Git el bloque conservado por la conciliación y sus
regresiones. La configuración local/dev del workspace no se ha promovido al
checkout estable de despliegue: conserva la configuración del release v1.9.0.
No se publica una versión ni se actualiza la API de producción. Los temporales
Docker de validación y las imágenes descargadas exclusivamente para esta
revisión se retiran; se conservan imágenes y volúmenes preexistentes.

## Retrospectiva técnica

Una validación sintáctica no demuestra filtrado, recuperación ni purga efectiva.
Las pruebas de comportamiento y la comprobación posterior de producción cierran
preguntas distintas. Ante incertidumbre, las herramientas de limpieza y medición
conservan datos o fallan de forma visible. Estas fronteras merecen regresiones
pequeñas sin añadir dependencias; el aprendizaje se incorpora al gate existente.
