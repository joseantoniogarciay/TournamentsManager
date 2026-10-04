# Conciliación del workspace — 2026-10-03

## Problema y evidencia

El checkout `develop` seguía en `17eab6f` (v1.8.0), mientras las publicaciones
v1.8.1–v1.9.0 se habían integrado desde checkouts aislados. Tras actualizar las
referencias remotas, `origin/develop` está en `6967457` y `origin/main`, igual
que el tag anotado v1.9.0, en `69533ac`. La base local estaba 16 commits atrás.

De los 148 archivos modificados o nuevos del workspace original:

- 97 coincidían byte a byte con `origin/develop`: no eran trabajo sin entregar.
- 23 conservaban cambios locales sobre archivos no cambiados en remoto.
- 13 eran archivos locales que aún no existían en remoto.
- 15 habían cambiado en ambos lados y requerían comparar su contenido.

Esta clasificación describe archivos y contenido, no demuestra por sí sola
que una configuración esté aplicada a un runtime. La evidencia de publicación
está en [DEPLOYMENT.md](DEPLOYMENT.md), y la de cambios operativos en
[LOG_RETENTION.md](LOG_RETENTION.md), [OBSERVABILITY.md](OBSERVABILITY.md) y los
runbooks correspondientes.

## Alternativas y actuación

Reaplicar sin revisar todo el diff antiguo podía retirar pruebas ya publicadas o
restaurar la imagen anterior de API. Mantener la base v1.8.0 perpetuaba la mezcla
de entregado y pendiente. Se optó por avanzar la base y conservar únicamente el
diff pendiente comparado con la revisión remota, sin reescribir historia.

Antes del avance se guardaron una copia privada de los 148 archivos y los diffs
en `.config/reconciliation/`, ignorado por Git. Se conserva también el stash
`pre-reconciliation 2026-10-03: preserved local work`. Su contenido parte de la
base antigua: no debe aplicarse entero sobre el checkout actual sin revisión.

`develop` avanzó por fast-forward a `6967457`; la referencia local `main` avanzó
a `69533ac`, tras comprobar que era descendiente de su referencia anterior.
El checkout permanece en `develop`. No se crearon commits en estas ramas ni se
hicieron pushes o despliegues durante la conciliación.

Se recuperaron las pruebas `test:match-incidents`, su inclusión en `make verify`
y la referencia declarativa de API v1.9.0. Los cambios funcionales, las cinco
migraciones 00015–00019, el contrato y el cliente coinciden con la base integrada.
El ajuste generado sin salto final de `expo-env.d.ts` quedó normalizado al archivo
versionado; no hay cambios de interfaz nuevos.

## Entregado y pendiente

| Bloque | Estado verificable | Archivos y evidencia |
| --- | --- | --- |
| Deportes, sets e incidencias | Integrados en v1.9.0; publicación documentada | Dominio, HTTP, PostgreSQL, migraciones 00015–00019, OpenAPI y cliente; [publicación](DEPLOYMENT.md) |
| Dependencias y compatibilidad web | Integradas en v1.8.1–v1.8.3 y conservadas en v1.9.0 | Módulos Go, lockfile, overrides, patches y tests; [informe histórico](WEB_DEPENDENCY_COMPATIBILITY_2026-10-03.md) |
| Retención y espacio de producción | Aplicación al runtime registrada; archivos aún pendientes de integrar | ADR-0140, valores Loki/Alloy/Prometheus, fixtures, reglas de alerta, timer, métricas del host y journal; [runbook K3s](../runbooks/k3s-observability.md) |
| Retención local/dev y limpieza | Configuración del workspace pendiente de integrar; no promovida al checkout estable de dev | ADR-0139, Compose, Loki, Promtail, Tempo y `clean-development-data.sh`; [retención](LOG_RETENTION.md) |
| Mitigación de módulos del host | Aplicada con autorización según evidencia; archivo pendiente de integrar | `fasttourney-unused-kernel-modules.conf`; [mitigación](SECURITY_PATCHES_2026-10-03.md) |
| Migración y herramientas locales | Código y documentación pendientes de integrar | `local-migrate.sh`, `mk/postgres.mk`, guías de desarrollo y Compose; [runbook local](../runbooks/local-postgresql.md) |
| Backup cifrado dev | WAL, incremental programada y restauración verificados; conservación local activada; corrección del comando pendiente de integrar | `verify-backup-restore.py`, Make y [runbook de backup](../runbooks/postgresql-backup-dev.md) |
| Evidencia y gobierno | Aclaraciones e informes pendientes de integrar | Auditoría de dependencias, diario, índices, plazos de retención y cronología de ADR-0075/0077 |

La entrada v1.9.0 del changelog vuelve a contener exclusivamente lo publicado en
ese release. `Unreleased` conserva el trabajo operativo pendiente de integrar,
indicando cuándo ya se aplicó al runtime. Los informes de seguridad mantienen sus
versiones históricas y enlazan el estado actual para no presentar v1.8.1–v1.8.3
como versiones activas del producto después de v1.9.0.

## Verificación y límites

- `develop == origin/develop == 6967457`; `main == origin/main == v1.9.0 == 69533ac`.
- Sin diff de cliente, dominio, adaptadores, contrato, dependencias o gate de
  pruebas respecto de la revisión integrada.
- `make test-dependencies test-match-incidents`: nueve pruebas aprobadas.
- `git diff --check` y enlaces locales de la documentación modificada comprobados.
- No se ejecutó `make verify` completo ni se revalidó o publicó todo el bloque
  operativo conservado: esta revisión concilia su procedencia y estado.
- El usuario confirmó después: «Comprobé un torneo de padel entero en movil y
  0 problema». Queda registrado el recorrido completo de pádel en móvil sin
  incidencias reportadas. No se especificaron navegador/app, sistema operativo,
  entorno ni los casos particulares recorridos; no acredita por sí solo el
  teclado nativo o todos los deportes.

## Retrospectiva y siguiente incremento

Un archivo que parece modificado respecto de una base antigua puede estar ya
publicado. Comparar los blobs con la revisión integrada evita repetir trabajo y
retirar pruebas de un release. Integración Git, publicación y aplicación de
configuración operativa son hechos distintos.

La base ya permite continuar. El bloque operativo pendiente debe integrarse tras
su revisión y validación proporcional, conservando el checkout estable de dev.
El usuario ya completó un torneo de pádel en móvil sin problemas reportados.
El siguiente bloque recomendado es revisar y validar los cambios operativos
pendientes para integrarlos; cualquier ampliación funcional requiere un problema
concreto y su decisión, sin repetir este recorrido de pádel por defecto.

## Cierre posterior del bloque operativo

La [revisión posterior](OPERATIONAL_BLOCK_REVIEW_2026-10-03.md) valida e integra
el bloque operativo conservado aquí. Añade pruebas de seguridad al gate,
revalida la restauración aislada y comprueba en producción que la condición de
bloques Prometheus vencidos ya no está presente. La clasificación anterior es
la fotografía de la conciliación; la configuración local/dev sigue sin promover
al checkout estable de despliegue.
