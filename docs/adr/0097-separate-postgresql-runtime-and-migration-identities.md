# ADR-0097: Separar las identidades PostgreSQL de ejecución y migración

- **Estado:** Aceptado
- **Fecha:** 2026-08-16
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

La API se conecta hoy con la identidad que inicializa PostgreSQL. En los entornos
locales y de desarrollo eso simplifica el arranque, pero esa identidad puede ser
propietaria del esquema. Si una vulnerabilidad de inyección SQL o una credencial
de la API se comprometiera en un entorno compartido, el impacto incluiría cambios
de esquema y permisos que la API no necesita para atender peticiones.

Se necesita limitar la identidad de ejecución a las operaciones de datos de la
API y reservar la evolución del esquema a una identidad diferente, sin añadir un
gestor de secretos, una abstracción de persistencia ni un runner de migraciones
nuevo antes de necesitarlos.

## Contexto y restricciones

- ADR-0011 fija PostgreSQL con `pgx`, `sqlc` y migraciones SQL; el SQL y los
  adaptadores permanecen fuera del dominio.
- El adaptador actual usa parámetros `$n`; una revisión de las consultas de
  producción no encontró SQL formado mediante concatenación o interpolación.
  La separación de roles es defensa en profundidad, no sustituto de consultas
  parametrizadas ni de validación de entradas.
- ADR-0017 exige secretos fuera de Git y credenciales mínimas por entorno.
- `docs/engineering/DATABASE.md` mantiene el esquema inicial reseteable y no
  activa todavía Goose; los datos locales son descartables.
- El runtime doméstico público de `dev` usa PostgreSQL aislado; producción y su
  proceso de migración siguen pendientes. `dev` se reinicia desde una base vacía
  para crear los roles antes de su esquema inicial.
- Los roles PostgreSQL son un control de la infraestructura: no se exponen al
  dominio ni se duplican como autorizaciones de producto.

## Criterios de decisión

1. Reducir el alcance de una credencial comprometida de la API.
2. Mantener las consultas parametrizadas como barrera primaria y hacer visible
   cualquier excepción de SQL dinámico.
3. Separar el despliegue de cambios de esquema del arranque ordinario de la API.
4. Conservar un flujo local sencillo mientras la base siga siendo descartable.
5. Evitar un gestor de secretos o una herramienta de análisis adicional sin una
   necesidad medida.

## Alternativas

### Alternativa A — Roles separados de ejecución, migración y propiedad sin login

Crear por base y entorno compartido:

- `tm_schema_owner` sin `LOGIN`, propietario de la base, esquema y objetos;
- `tm_db_migrator` con `LOGIN`, usado solo por el paso explícito de migración y
  autorizado para actuar como el propietario;
- `tm_app_runtime` con `LOGIN`, usado solo por la API y con `CONNECT`, `USAGE`
  en el esquema, permisos DML explícitos en las tablas necesarias y `USAGE`/
  `SELECT` en las secuencias necesarias.

Se revocan los privilegios por defecto de `PUBLIC` que no sean necesarios. La
API no recibe `CREATE`, `ALTER`, `DROP`, `TRUNCATE`, gestión de roles ni la
credencial de migración. No se conceden permisos por defecto a objetos futuros:
cada nueva tabla o secuencia exige un `GRANT` de runtime dentro de la misma
migración.

- **Ventajas:** mínimo privilegio real; el ownership no depende de una credencial
  de uso ordinario; credenciales, auditoría y rotación quedan separadas; encaja
  con migraciones SQL explícitas de ADR-0011.
- **Inconvenientes:** hay tres identidades y los `GRANT` deben evolucionar con
  el esquema; un permiso omitido falla en despliegue hasta corregir la migración.
- **Coste de adopción:** medio, al introducir el primer entorno con datos no
  descartables o al formalizar las migraciones.
- **Coste de mantenimiento:** bajo o medio; revisar permisos en cada migración
  y rotar dos secretos operativos en lugar de uno.
- **Riesgos:** privilegios demasiado amplios o insuficientes. Se mitigan con SQL
  de roles versionado, revisión y una prueba de integración con `tm_app_runtime`.

### Alternativa B — Un único rol con acceso total para API y migraciones

Mantener una única `DATABASE_URL` que inicializa el esquema y ejecuta la API.

- **Ventajas:** configuración y desarrollo inicial mínimos.
- **Inconvenientes:** una inyección SQL o credencial comprometida conserva el
  poder de alterar objetos y permisos; no permite auditar ni rotar por uso.
- **Coste de adopción y mantenimiento:** bajo.
- **Riesgos:** radio de impacto innecesariamente alto en un entorno público o
  con datos que se deban conservar.

### Alternativa C — Roles por tabla o por módulo desde ahora

Crear identidades independientes para cada módulo del monolito y concederles
acceso solo a sus tablas.

- **Ventajas:** aislamiento máximo dentro de la base.
- **Inconvenientes:** la API es un único proceso y sus casos de uso ya cruzan
  tablas y transacciones; complica permisos, depuración y cambios de producto
  sin un límite de despliegue real.
- **Coste de adopción y mantenimiento:** alto.
- **Riesgos:** falsa modularidad y errores operativos por sobreingeniería.

### No cambiar

La API mantiene una identidad propietaria. Es razonable solo mientras la base
sea estrictamente local, aislada y descartable; no satisface el criterio de
mínimo privilegio para un entorno público o compartido.

## Comparación

La B reduce trabajo inmediato, pero contradice el principio de mínimo privilegio
ya documentado en `SECURITY.md`. La C aporta aislamiento que un monolito con una
sola unidad desplegable todavía no puede aprovechar y eleva el coste continuo.
La A limita el daño ante compromiso, separa claramente responsabilidades y no
cambia el dominio ni las consultas. Su coste aparece donde ya hay un cambio de
esquema y un despliegue explícito, por lo que es verificable y recuperable.

## Recomendación

**Opinión/recomendación:** aceptar la alternativa A para cada entorno público o
compartido antes de conservar datos en él. Mantener local con su bootstrap actual
hasta que se active Goose o exista una necesidad operativa concreta. Es la
solución mínima suficiente: tres roles por base, no una identidad por módulo ni
un gestor de secretos nuevo.

Como convención complementaria, toda consulta seguirá usando parámetros. Si se
necesitan identificadores SQL dinámicos, se elegirán exclusivamente desde una
allowlist cerrada en código. En la próxima decisión de política CI se evaluará
si una regla Semgrep aporta evidencia proporcional; no se añade todavía una
dependencia solo para detectar patrones que la revisión de SQL ya hace visibles.

## Decisión del usuario

**Aceptada el 2026-08-16:** alternativa A. `tournaments-manager-dev` adopta los
roles separados ahora, antes del siguiente despliegue de la API. El bootstrap es
una operación explícita sobre una base vacía: crea roles, aplica el esquema como
migrador y concede después el runtime. Los despliegues ordinarios no reciben la
credencial de migración ni cambian privilegios de base de datos.

## Consecuencias

### Positivas

- La credencial de la API no puede alterar el esquema ni administrar roles.
- Un cambio de base exige declarar y revisar los permisos de runtime necesarios.
- La migración pasa a ser una operación explícita y auditable, separada de la
  disponibilidad ordinaria de la API.

### Negativas y deuda aceptada

- Dos secretos operativos deben gestionarse y rotarse por entorno.
- Los despliegues fallarán de forma segura si un `GRANT` no acompaña a un objeto
  nuevo; el procedimiento debe contemplar ese diagnóstico y forward-fix.
- La revisión automatizada de construcción dinámica de SQL queda aplazada hasta
  medir que la revisión manual no detecta regresiones suficientes.

## Validación

1. Un script SQL reproducible crea los tres roles sin imprimir contraseñas.
2. Una conexión con `tm_app_runtime` completa las pruebas de integración de la
   API y puede operar solo las tablas y secuencias previstas.
3. Esa misma conexión falla al crear, alterar, eliminar o truncar una tabla y al
   otorgar privilegios.
4. La migración se ejecuta con `tm_db_migrator`; la API no recibe esa URL ni su
   secreto en imagen, logs, trazas o variables públicas.
5. Una migración que añada una tabla necesaria para la API incluye y comprueba
   sus `GRANT` correspondientes.
6. Las pruebas de consultas siguen demostrando que entradas como comillas y
   comentarios SQL se tratan como datos, nunca como parte de una sentencia.

## Disparadores de revisión

- La API se divide en varios procesos con límites de datos reales.
- Se introduce RLS, multitenencia, una réplica de lectura o una cuenta de
  reporting que requiera una política distinta.
- Un incidente, auditoría o requisito de cumplimiento exige rotación central,
  credenciales efímeras o auditoría de base más detallada.
- La gestión manual de secretos o permisos se vuelve frecuente y justifica un
  gestor de secretos o automatización adicional.

## Documentación afectada

- [DECISIONS.md](../governance/DECISIONS.md)
- [DECISIONS_TO_REVISIT.md](../governance/DECISIONS_TO_REVISIT.md)
- [DATABASE.md](../engineering/DATABASE.md)
- [SECURITY.md](../engineering/SECURITY.md)
- [DEPLOYMENT.md](../operations/DEPLOYMENT.md)
- [ADR-0011](0011-use-postgresql-pgx-sqlc-and-goose.md)
- [ADR-0017](0017-use-env-contracts-github-environments-and-oidc.md)
