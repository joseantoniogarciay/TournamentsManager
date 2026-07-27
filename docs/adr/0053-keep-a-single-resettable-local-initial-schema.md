# ADR-0053: Mantener un único esquema inicial local reseteable

- **Estado:** Aceptado
- **Fecha:** 2026-07-27
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** ADR-0047, únicamente en la política temporal de evolución local
- **Superado por:** Ninguno

## Problema

El proyecto está construyendo y validando el primer modelo de datos solo en el
entorno PostgreSQL local. Acumular migraciones incrementales antes de necesitar
conservar datos hace más difícil leer y corregir el modelo inicial.

## Contexto y restricciones

- No hay entorno compartido, despliegue ni datos que deban conservarse.
- El usuario acepta borrar la base local al cambiar el modelo.
- PostgreSQL, Goose y sqlc continúan siendo las herramientas aceptadas.
- El esquema debe seguir siendo una fuente de verdad ejecutable y reproducible.

## Alternativas

### A — Migraciones incrementales desde ahora

- **Ventajas:** replica desde el principio la disciplina necesaria con datos compartidos.
- **Inconvenientes:** conserva pasos intermedios que no aportan valor mientras el modelo cambia libremente.
- **Coste de mantenimiento:** medio.

### B — Un único esquema inicial reescribible con reset local

- **Ventajas:** el estado inicial es legible; cada cambio deja una sola representación actual del modelo.
- **Inconvenientes:** el reset destruye todos los datos locales.
- **Coste de mantenimiento:** bajo mientras no haya datos que conservar.

### C — No versionar ningún esquema

- **Ventajas:** menos archivos inicialmente.
- **Inconvenientes:** impide reproducir y revisar la base; contradice la fuente de verdad ejecutable.
- **Coste de mantenimiento:** alto por deriva manual.

## Decisión del usuario

**Aceptada el 2026-07-27:** elegir la alternativa B hasta que exista un entorno
compartido o datos que deban conservarse. `00001_initial_schema.sql` es el único
esquema inicial: se edita y se resetea PostgreSQL local cuando cambie el modelo.
No se añaden migraciones incrementales durante esta etapa.

## Consecuencias

- `make db-reset`, `make db-up` y `make db-migrate` reconstruyen el estado local
  desde el único esquema inicial.
- La API y cualquier dato local deben considerarse desechables durante el reset.
- Antes del primer entorno compartido, backup o dato que se deba preservar, se
  abre una decisión sucesora para volver a migraciones inmutables e incrementales.

## Validación

- Desde un volumen local eliminado, Goose aplica solamente `00001`.
- El esquema resultante contiene las tablas de identidad local, Google, borrador
  y liga vigentes, sin `league_share_links`.
- Una segunda ejecución de `make db-migrate` no introduce cambios.

## Disparadores de revisión

- Se crea un entorno compartido o se comparte la base con otra persona.
- Aparecen datos locales que sea necesario conservar.
- Se prepara el primer despliegue o una copia de seguridad.

## Documentación afectada

- [Datos y persistencia](../engineering/DATABASE.md)
- [Desarrollo](../engineering/DEVELOPMENT.md)
- [Decisiones](../governance/DECISIONS.md)
- [Runbook PostgreSQL local](../runbooks/local-postgresql.md)
