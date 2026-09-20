# ADR-0131: Sincronizar el último nombre de equipo como preferencia de cuenta

- **Estado:** Aceptado
- **Fecha:** 2026-09-20
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** ADR-0130, exclusivamente en que el último nombre de equipo se
  guardaba solo en el dispositivo
- **Superado por:** Ninguno

## Problema

El cliente recordaba el último nombre de equipo en una clave local única. Al
cerrar sesión y entrar con otra cuenta, ambas compartían la misma sugerencia; al
usar otro navegador o la app móvil, en cambio, la sugerencia no viajaba con la
cuenta. El comportamiento no representa correctamente una preferencia personal.

## Contexto y restricciones

- El nombre es una sugerencia editable, no una identidad permanente del equipo.
- Un borrador local de creación conserva precedencia sobre cualquier sugerencia.
- Solo se recuerda un nombre después de crear el equipo inicial de un torneo o
  de incorporarse correctamente mediante una invitación.
- Los equipos sin cuenta que añade una organizadora no cambian su preferencia.
- Web, iOS y Android comparten sesión y API; no se introduce sincronización
  dispositivo a dispositivo fuera del backend.
- El nombre no se exporta a logs, spans, métricas ni motivos de fallo.

## Alternativas

### A — Separar la clave local por cuenta

- **Ventajas:** cambio pequeño y sin persistencia de servidor.
- **Inconvenientes:** corrige la contaminación entre cuentas solo dentro del
  mismo dispositivo; web y móvil siguen divergiendo.
- **Coste de mantenimiento:** bajo, con semántica incompleta.

### B — Persistir la preferencia explícita en la cuenta

- **Ventajas:** la misma cuenta recibe la misma sugerencia en web y móvil; el
  dato se actualiza dentro de la transacción que confirma el equipo.
- **Inconvenientes:** añade una columna, amplía la proyección de sesión y obliga
  a mantener actualizada la caché local de la sesión móvil.
- **Coste de mantenimiento:** bajo; es un único valor opcional sin CRUD propio.

### C — Derivar el nombre del último equipo vinculado

- **Ventajas:** no añade una preferencia separada.
- **Inconvenientes:** el orden de actividad, borrado o historial deportivo pasa
  a decidir una intención de interfaz; además una consulta histórica no expresa
  qué nombre confirmó la persona más recientemente.
- **Coste de mantenimiento:** medio y acoplado al modelo de torneos.

## Recomendación

**Opinión/recomendación:** alternativa B. Guardar un único valor opcional en la
cuenta y devolverlo dentro de la identidad de sesión ya usada para arrancar el
cliente. No crear un endpoint ni un servicio genérico de preferencias mientras
solo exista este dato.

## Decisión del usuario

**Aceptada el 2026-09-20:** sincronizar el último nombre de equipo por cuenta.
Se actualiza atómicamente al crear el equipo inicial propio o al incorporarse
mediante invitación, y se devuelve en toda sesión autenticada. El cliente relee
esa proyección al abrir cualquiera de los dos formularios, la usa como
sugerencia editable cuando no hay borrador y actualiza la copia de sesión tras
el éxito para no esperar al siguiente acceso o refresh.

## Consecuencias

### Positivas

- cambiar de cuenta deja de reutilizar la sugerencia anterior;
- la cuenta comparte el valor entre navegador, iOS y Android;
- una respuesta perdida no puede guardar la preferencia sin haber creado el
  equipo, porque ambas escrituras comparten transacción;
- no se añade una API de preferencias prematura.

### Negativas y deuda aceptada

- la sesión contiene un dato de presentación adicional y el almacenamiento
  seguro móvil conserva una copia que debe reemplazarse tras cada éxito;
- no existe edición explícita de la preferencia fuera de crear o inscribir un
  equipo;
- no se conserva historial de nombres ni se intenta inferir uno para cuentas
  existentes.

## Validación

- crear un torneo e incorporarse actualizan la preferencia solo si la operación
  completa confirma;
- alta con borrador y acceso con borrador devuelven el nombre recién confirmado
  en la propia sesión;
- restaurar o refrescar sesión devuelve la preferencia actual de esa cuenta;
- abrir creación o incorporación relee el valor para recoger cambios realizados
  desde otro dispositivo ya autenticado;
- cerrar sesión y acceder con otra cuenta reemplaza la sugerencia;
- la app móvil persiste la identidad de sesión actualizada sin una clave global
  de nombre de equipo;
- un borrador local sigue prevaleciendo y el campo continúa siendo editable.

## Disparadores de revisión

- necesidad de varias identidades deportivas o plantillas por cuenta;
- edición explícita desde un perfil;
- aparición de suficientes preferencias sincronizadas para justificar un recurso
  o caso de uso propio.

## Documentación afectada

- [Producto](../project/PRODUCT.md)
- [Modelo inicial de datos](../engineering/INITIAL_DATA_MODEL.md)
- [API](../engineering/API.md)
- [Decisiones](../governance/DECISIONS.md)
- [Aprendizaje](../project/LEARNING.md)

## Retrospectiva técnica de implementación

- **Resultado:** una columna opcional y la proyección de usuario de sesión
  sustituyen la clave local compartida; no fue necesario crear un recurso de
  preferencias ni otro servicio de dominio.
- **Qué funcionó:** escribir el nombre dentro de las transacciones existentes
  mantuvo juntos equipo, vínculo y preferencia. Releer `GET /sessions` al entrar
  en el formulario cubre también dispositivos que ya estaban autenticados.
- **Ajuste importante:** la app móvil conserva la proyección dentro del mismo
  registro seguro que sus tokens. Así el cambio de cuenta reemplaza toda la
  identidad local y no deja una sugerencia global huérfana.
- **Coste conservado:** no se infiere historial ni se añade edición de perfil;
  el valor solo cambia cuando una operación confirma realmente un equipo propio.
