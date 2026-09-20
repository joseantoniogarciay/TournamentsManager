# ADR-0130: Permitir inscribir equipos mediante enlaces de invitación al torneo

- **Estado:** Aceptado
- **Fecha:** 2026-09-19
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** ADR-0034, exclusivamente en que las cuentas no se vinculaban a
  equipos y las invitaciones de participación quedaban fuera de alcance
- **Superado por:** Ninguno

## Problema

Crear un torneo exige actualmente que la organizadora introduzca todos los
equipos. Esto concentra un trabajo que cada participante puede resolver mejor:
elegir el nombre de su propio equipo. El producto necesita permitir esa
incorporación sin convertir a las personas participantes en administradoras ni
confundir el enlace público de lectura con permiso para modificar el torneo.

## Contexto y restricciones

- El equipo sigue siendo el participante deportivo; una cuenta vinculada no
  representa una plantilla de jugadores ni recibe permisos sobre resultados.
- La organizadora debe crear el torneo con su propio equipo. Puede añadir además
  equipos anónimos, sin cuenta asociada, mientras el torneo no haya empezado.
- El torneo publicado continúa siendo legible mediante su identificador
  público, pero conocerlo solo concede lectura.
- Solo una cuenta verificada puede conservar una incorporación entre
  dispositivos y aparecer en su biblioteca personal.
- La composición se congela al iniciar el torneo y mantiene el límite vigente de
  64 equipos. El inicio sigue exigiendo al menos dos.
- No se añaden email, push, invitaciones individuales, jugadores, plantillas ni
  aceptación administrativa.
- El nombre usado más recientemente es una comodidad del cliente, no identidad
  ni dato necesario del dominio.

## Criterios de decisión

1. reducir el trabajo de la organizadora sin ceder administración;
2. separar lectura pública, inscripción deportiva, seguimiento y permisos;
3. impedir incorporaciones después de congelar la composición;
4. conservar equipos anónimos para participantes sin cuenta;
5. hacer recuperable el torneo para quien se incorpora;
6. evitar infraestructura de notificaciones y perfiles deportivos prematuros.

## Alternativas

### A — Permitir incorporarse desde el enlace público de lectura

- **Ventajas:** no añade otro enlace y reduce la interfaz inicial.
- **Inconvenientes:** conocer la URL pública permitiría modificar la composición;
  lectura y mutación compartirían una capacidad que hoy es de solo lectura.
- **Coste de adopción:** bajo.
- **Coste de mantenimiento:** medio por los controles compensatorios y la
  dificultad de explicar qué permite cada enlace.
- **Riesgos:** altas no deseadas, distribución accidental del permiso y abuso.

### B — Usar un enlace específico de inscripción y conservar equipos anónimos

- **Ventajas:** la organizadora decide cuándo compartir o revocar la capacidad;
  cada cuenta crea su propio equipo sin recibir administración; los equipos sin
  cuenta siguen siendo posibles.
- **Inconvenientes:** añade un secreto revocable, una relación cuenta-equipo y
  estados de interfaz para login, conflicto y expiración funcional.
- **Coste de adopción:** medio.
- **Coste de mantenimiento:** bajo a medio; existe una única invitación activa
  por torneo y no requiere entrega propia.
- **Riesgos:** el enlace puede reenviarse mientras esté activo y debe tratarse
  como secreto sin aparecer en logs o métricas.

### C — Crear plazas o equipos previamente y permitir reclamarlos

- **Ventajas:** la organizadora controla de antemano el número y la identidad de
  las plazas.
- **Inconvenientes:** obliga a preparar la composición y traslada a la
  organizadora el trabajo que se quiere eliminar.
- **Coste de adopción y mantenimiento:** medio.
- **Riesgos:** reclamaciones equivocadas, plazas abandonadas y un flujo de
  reasignación adicional.

### No cambiar

La organizadora seguiría introduciendo todos los equipos y las cuentas
participantes no podrían recuperar automáticamente el torneo como seguido.

## Comparación

La alternativa A es pequeña, pero convierte un enlace diseñado para lectura en
permiso de escritura. La C conserva el control a costa de no resolver el
problema principal. La B añade únicamente la capacidad nueva necesaria: un
enlace de inscripción distinto y revocable, sin entrega de mensajes ni roles
administrativos nuevos.

## Recomendación

**Opinión/recomendación:** alternativa B. Mantener equipos como participantes y
añadir una relación mínima entre una cuenta y el equipo que inscribió. Crear el
equipo, esa relación y el seguimiento dentro de una sola transacción.

## Decisión del usuario

**Aceptada el 2026-09-19:** la organizadora crea el torneo con un equipo propio
obligatorio. Hasta iniciar el torneo puede añadir cualquier número válido de
equipos anónimos y puede compartir un enlace específico para que otras cuentas
verificadas creen su equipo al vuelo.

La invitación es reutilizable mientras esté activa, distinta del enlace público
de lectura y revocable o regenerable por la organizadora. Cada cuenta puede
mantener como máximo un equipo vinculado por torneo; si la organizadora elimina
ese equipo antes del inicio, la cuenta puede volver a inscribirse. Incorporarse crea el equipo, lo
vincula a la cuenta y guarda el torneo en «Sigo», pero no concede administración
ni permiso para registrar resultados. La precedencia visual de «Administro» se
mantiene cuando una cuenta reúne ambas relaciones.

El equipo inicial queda vinculado a la cuenta organizadora. Los equipos añadidos
directamente por esta después son anónimos. Todos compiten con las mismas reglas;
la vinculación solo conserva quién incorporó el equipo y la relación personal
con el torneo.

El cliente recuerda localmente el último nombre de equipo confirmado y lo
propone, siempre editable, tanto al crear un torneo como en la siguiente
inscripción. Un borrador de creación conservado tiene precedencia sobre esa
sugerencia. No se sincroniza como preferencia de cuenta en este incremento.

La interfaz explica la composición como dos caminos disponibles hasta el inicio:
«Añadir equipo» crea un equipo sin cuenta asociada y «Compartir invitación» deja
que cada persona cree el suyo. Evita usar «anónimo» como única etiqueta visible,
porque puede sugerir que se oculta una identidad; «sin cuenta» describe la
diferencia de producto. Antes de confirmar una invitación se indica que el
torneo aparecerá en «Sigo» y que la incorporación no concede administración.
Mientras el torneo no haya empezado, el detalle de la organizadora despliega esa
gestión. La acción de inicio permanece deshabilitada con un solo equipo y se
habilita desde dos, pero alcanzar el mínimo no compacta la composición ni la
presenta como terminada.

## Consecuencias

### Positivas

- la organizadora puede publicar con su equipo y delegar la introducción de
  nombres sin delegar gestión;
- una persona incorporada encuentra el torneo en «Sigo» sin una segunda acción;
- el enlace público conserva su semántica de solo lectura;
- los grupos sin cuenta siguen representados mediante equipos anónimos;
- no se introduce un modelo de jugadores ni plantillas.

### Negativas y deuda aceptada

- debe persistirse y protegerse una invitación reutilizable y una relación
  cuenta-equipo;
- quien reciba un enlace reenviado puede incorporarse mientras siga activo;
- no se decide todavía si una cuenta puede renombrar o abandonar su equipo;
- borrar antes del inicio un equipo vinculado no elimina automáticamente el
  seguimiento independiente de la cuenta y permite reemplazarlo mediante una
  nueva inscripción;
- el último nombre no se comparte entre dispositivos.

## Validación

- crear un torneo exige exactamente un equipo inicial vinculado a la
  organizadora, despliega la gestión hasta que empiece y permite iniciarlo solo
  cuando contiene al menos dos equipos;
- la organizadora puede añadir y eliminar equipos anónimos únicamente antes del
  inicio, respetando nombres únicos y el máximo de 64;
- el enlace público nunca permite inscribir y la invitación revocada o
  regenerada deja de aceptar el secreto anterior;
- una cuenta verificada puede inscribir un solo equipo por invitación válida;
  equipo, vínculo y seguimiento se confirman o revierten juntos;
- dos incorporaciones concurrentes no superan el límite ni duplican cuenta o
  nombre;
- una incorporación válida aparece en «Sigo» y no autoriza ninguna operación de
  administración o resultados;
- iniciar el torneo invalida funcionalmente cualquier incorporación posterior;
- tokens, nombres e identificadores de cuenta no se exportan en motivos de fallo,
  logs ni métricas;
- el cliente solo guarda el último nombre después de una creación o inscripción
  correcta, lo propone en ambos formularios cuando no existe un borrador y
  permite reemplazarlo antes de enviar;
- creación, gestión de equipos e incorporación explican los dos caminos, el
  cierre al iniciar y la ausencia de permisos administrativos sin depender de
  que la persona deduzca esas reglas del estado de los botones.

## Disparadores de revisión

- necesidad de renombrar, abandonar o transferir el equipo vinculado;
- necesidad de varias cuentas por equipo, jugadores o plantillas;
- abuso por reenvío del enlace que requiera aprobación, cupos o invitaciones
  individuales;
- necesidad de sincronizar el último nombre entre dispositivos;
- incorporación de participantes individuales en otro deporte.

## Documentación afectada

- [Producto](../project/PRODUCT.md)
- [Modelo inicial de datos](../engineering/INITIAL_DATA_MODEL.md)
- [API](../engineering/API.md)
- [Seguridad](../engineering/SECURITY.md)
- [Observabilidad](../operations/OBSERVABILITY.md)
- [Decisiones](../governance/DECISIONS.md)
- [Aprendizaje](../project/LEARNING.md)

## Retrospectiva técnica de implementación

- **Resultado:** el incremento quedó resuelto con dos relaciones pequeñas, tres
  operaciones de capacidad y una pantalla de incorporación; no hizo falta
  introducir perfiles deportivos, mensajería ni un rol nuevo.
- **Qué funcionó:** diseñar contrato, transacción y copy como una sola unidad
  evitó que «participar», «seguir» y «administrar» volvieran a mezclarse. La
  prueba PostgreSQL confirma que equipo, vínculo y seguimiento se confirman
  juntos y que iniciar invalida la invitación.
- **Ajuste importante:** transportar el secreto en el fragmento del enlace y
  retirarlo antes de inspeccionar reduce su exposición a servidor web, historial
  de rutas y telemetría. La API recibe el token solo en cuerpos `POST` y guarda
  únicamente su hash contextual.
- **Coste conservado:** se mantiene una única invitación activa y un único
  vínculo cuenta-equipo. Renombrar, abandonar, aprobar incorporaciones o asociar
  varias cuentas sigue fuera hasta que exista evidencia para asumir ese modelo.
