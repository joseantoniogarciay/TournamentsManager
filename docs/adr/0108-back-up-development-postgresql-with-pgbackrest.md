# ADR-0108: Respaldar PostgreSQL de desarrollo con pgBackRest

- **Estado:** Aceptado
- **Fecha:** 2026-08-23
- **Decisor:** Usuario

## Problema

El entorno público `dev` conserva datos y evidencia legal. Perder el volumen de
PostgreSQL o una mutación accidental no tiene hoy una restauración integral ni
un punto de recuperación verificable.

## Contexto y restricciones

- ADR-0091 separa el volumen de `dev` de `local` y de la futura producción.
- ADR-0106 conserva una exportación incremental, cifrada y limitada de evidencia
  legal; no es un backup de la base completa.
- PostgreSQL requiere una copia base y una secuencia continua de WAL para PITR.
- El repositorio inicial será una carpeta privada de iCloud Drive distinta del
  backup legal. Está fuera del volumen PostgreSQL, pero comparte Mac y cuenta
  iCloud; no constituye por sí solo una estrategia de producción.

## Criterios

1. restaurar todo el clúster y llegar a un instante posterior a una copia base;
2. cifrar antes de que los datos alcancen la carpeta sincronizada;
3. mantener una operación pequeña, reproducible y separada de la API;
4. conservar una ventana razonable sin añadir un servicio gestionado.

## Alternativas

### A — No hacer backup integral todavía

- **Ventaja:** ningún coste ni operación adicional.
- **Inconveniente:** una pérdida del volumen o un borrado no es recuperable.
- **Mantenimiento:** nulo, con riesgo inaceptable mientras `dev` conserva datos.

### B — `pg_dump` periódico

- **Ventaja:** sencillo y útil como exportación lógica.
- **Inconveniente:** no proporciona PITR ni reproduce WAL; restaura más lento y
  no cubre el clúster físico completo.
- **Mantenimiento:** bajo, pero insuficiente para el objetivo.

### C — pgBackRest con copia base, incrementales y WAL archivado

- **Ventaja:** automatiza backup, retención, cifrado, WAL y restauración PITR.
- **Inconveniente:** añade una imagen PostgreSQL propia, jobs y una prueba de
  restauración periódica.
- **Mantenimiento:** moderado y explícito.

## Recomendación

**Recomendación:** C. Es la mínima solución que cumple recuperación integral y
PITR sin acoplar el dominio a la infraestructura de backup.

## Decisión del usuario

**Aceptada el 2026-08-23:** `tournaments-manager-dev` usa pgBackRest 2.59.1,
compatible con PostgreSQL 18, y guarda un repositorio cifrado separado en iCloud
Drive. Archiva WAL con un máximo de una hora sin actividad, crea una copia
completa semanal y un incremental diario de lunes a sábado. Conserva dos copias
completas y sus dependencias, una ventana aproximada de 7 a 14 días.

La configuración usa la ruta de datos real de la imagen PostgreSQL 18
(`/var/lib/postgresql/18/docker`), comprobada con `SHOW data_directory`, no la
ruta histórica de imágenes anteriores.

## Consecuencias

- La clave `PGBACKREST_REPO1_CIPHER_PASS` es exclusiva, aleatoria y queda en el
  `.env` no versionado; no se reutiliza la clave de evidencia legal.
- `pgBackRest` no se ejecuta al arrancar la API ni durante una migración.
- El primer `make dev-public-backup-init` crea la stanza, valida WAL y toma la
  primera copia completa. `launchd` ejecuta después los calendarios versionados.
- Antes de producción se decide una segunda ubicación independiente, custodia y
  rotación de claves, RPO/RTO y pruebas periódicas; iCloud no aporta inmutabilidad.

## Validación

1. `pgbackrest check` confirma que el WAL llega al repositorio cifrado.
2. La primera copia completa y un incremental aparecen en `pgbackrest info`.
3. Se restaura una copia elegida en un destino aislado y PostgreSQL arranca sin
   tocar el volumen de `dev`.
4. La expiración conserva dos conjuntos completos recuperables.

**Evidencia inicial (2026-08-23):** stanza y `check` correctos; copia completa
`20260823-094744F`, incremental `20260823-094744F_20260823-094800I` y
restauración aislada verificada con `pg_is_in_recovery() = false`.

## Disparadores de revisión

- La ventana de 7–14 días, el RPO de una hora o el espacio de iCloud son
  insuficientes.
- Se exige producción, varios operadores, inmutabilidad o recuperación ante
  pérdida de la cuenta iCloud.
- El tamaño o latencia de WAL afecta al servidor.

## Documentación afectada

- [Datos y persistencia](../engineering/DATABASE.md)
- [Despliegue](../operations/DEPLOYMENT.md)
- [Runbook PostgreSQL](../runbooks/local-postgresql.md)
- [Decisiones](../governance/DECISIONS.md)
