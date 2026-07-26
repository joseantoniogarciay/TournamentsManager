# ADR-0034: Usar equipos como participantes y administración delegada directa

- **Estado:** Aceptado
- **Fecha:** 2026-07-26
- **Decisor:** Usuario
- **Propietario del análisis:** Codex
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

La liga inicial necesita diferenciar a quienes compiten de las personas que la
consultan o gestionan. Además, el creador debe poder delegar administración sin
introducir aún solicitudes de incorporación, jugadores, email ni notificaciones
push.

## Contexto y restricciones

- La liga de fútbol inicial tiene equipos creados por su organizador y una
  composición que se fija al publicarla, según ADR-0032.
- Las ligas publicadas son no listadas y consultables mediante enlace en modo de
  solo lectura, según ADR-0033.
- Las cuentas verificadas tienen una identidad interna independiente de email y
  proveedores de login, según ADR-0010.
- No se diseñan aún jugadores, pertenencia de usuarios a equipos, invitaciones
  con aceptación, email ni push.

## Criterios de decisión

1. modelar la competición con la menor ambigüedad;
2. separar participación deportiva, lectura y autorización administrativa;
3. delegar administración sin cadenas de permisos ni flujos de aceptación;
4. preservar la propiedad y una revocación simple;
5. evitar exponer el email como identificador de búsqueda.

## Alternativas

### Alternativa A — Usuarios como participantes directos

- **Ventajas:** relación inmediata entre cuenta y liga.
- **Inconvenientes:** no representa que compiten equipos y obliga a definir
  jugadores, afiliación y múltiples papeles antes de necesitarlos.
- **Coste de adopción:** moderado.
- **Coste de mantenimiento:** alto por la matriz persona-equipo-liga.
- **Riesgos:** mezclar administración con participación deportiva.

### Alternativa B — Equipos como únicos participantes y administración directa

Los equipos creados por el organizador compiten. Una cuenta verificada puede
seguir una liga y el creador puede asignar directamente administradores mediante
un `username` público y único.

- **Ventajas:** expresa el formato de liga; permisos claros y revocables; no
  expone correos ni requiere invitaciones aceptadas.
- **Inconvenientes:** el usuario asignado puede recibir un rol no solicitado y
  debe abandonarlo si no le interesa.
- **Coste de adopción:** moderado.
- **Coste de mantenimiento:** bajo a moderado por la gestión de roles y nombres
  de usuario.
- **Riesgos:** abuso de asignaciones si no se añaden límites y bloqueo en el
  futuro.

### Alternativa C — Equipos con invitaciones aceptadas y notificaciones desde el inicio

- **Ventajas:** consentimiento explícito y aviso inmediato.
- **Inconvenientes:** añade estados de invitación, entrega de email/push, tokens,
  preferencias y recuperación de fallos antes de validar la liga.
- **Coste de adopción:** alto.
- **Coste de mantenimiento:** alto.
- **Riesgos:** ampliar el primer corte con infraestructura de notificaciones.

### No cambiar

Mantener este modelo abierto impide definir la autorización, la lista de ligas
del usuario y la futura asignación de resultados.

## Comparación

La alternativa A adelanta un modelo de jugadores que no es necesario. La C
resuelve más casos, pero supera el alcance inicial. La B distingue tres relaciones
simples: equipo que compite, usuario que sigue y usuario que administra; añade
una delegación controlada por el creador sin notificaciones operativas.

## Recomendación

**Recomendación:** alternativa B. Mantiene una representación directa del
deporte y permite delegar operación sin convertir el primer corte en un sistema
de membresías e invitaciones.

## Decisión del usuario

**Aceptada el 2026-07-26:** adoptar la alternativa B.

- Los equipos creados por el organizador son los únicos participantes
  competitivos de una liga. Las personas no se incorporan a equipos ni a la
  competición en este corte.
- Solo una cuenta verificada puede seguir una liga; esta relación de lectura se
  muestra en «ligas seguidas». Abrir un enlace sin sesión no crea seguimiento.
- Toda cuenta debe completar un `username` público y único al activar el perfil
  verificado; en un primer acceso federado ocurre después de acreditar la
  identidad del proveedor. No se expone ni se usa el email para buscar usuarios.
- El `username` no se puede cambiar inicialmente. Un cambio futuro no alterará
  la identidad interna ni las relaciones existentes.
- El creador mantiene la propiedad y es el único que puede asignar o retirar
  administradores mediante `username`; la asignación es inmediata y no exige
  aceptación.
- Un administrador delegado puede abandonar por sí mismo una liga gestionada.
  El creador puede retirarlo en cualquier momento y el efecto es inmediato.
- Las capacidades operativas exactas de un administrador, incluidos resultados,
  se decidirán antes de implementarlas.
- Email, push con deep links, invitaciones con aceptación, límites antiabuso y
  bloqueo de organizadores quedan para decisiones posteriores.

## Consecuencias

### Positivas

- La estructura de competición no se acopla a cuentas de usuario.
- «Ligas seguidas» es recuperable entre dispositivos para cuentas autenticadas.
- La propiedad de la administración permanece clara y reversible.
- El `username` permite seleccionar usuarios sin divulgar su correo.

### Negativas y deuda aceptada

- Una persona puede ser nombrada administradora sin consentimiento previo.
- No existe aún bloqueo, límite de asignaciones ni avisos de administración.
- Bloquear cambios de `username` reduce flexibilidad inicial.
- Falta decidir la matriz concreta de permisos de los administradores.

## Aclaración de permisos — 2026-07-26

El usuario precisa que los administradores delegados se limitan a gestionar
resultados. No pueden modificar equipos, publicar o cancelar la liga, ni asignar
o retirar administradores. La mecánica de registro, modificación y confirmación
de un marcador continúa pendiente de decisión.

## Validación

- Crear o publicar una liga no crea participantes humanos; solo sus equipos
  forman parte de la competición.
- Un visitante sin sesión puede leer mediante enlace pero no seguir la liga.
- Un usuario autenticado puede seguir una liga y recuperarla en «ligas seguidas».
- Solo el creador puede asignar o retirar a otro administrador; la retirada y la
  salida voluntaria eliminan de inmediato el acceso delegado.
- Dos cuentas no pueden compartir `username`, ni una cuenta cambiarlo en este
  corte.
- Ninguna asignación de administrador necesita email, push o una aceptación.

## Disparadores de revisión

- Abuso de asignaciones no solicitadas o necesidad de consentimiento.
- Necesidad de jugadores, membresías de equipo o participación individual.
- Necesidad de cambiar `username`, nombres reservados o políticas de
  normalización más complejas.
- La gestión de resultados exige varios niveles administrativos.
- Se requiere avisar o bloquear a organizadores.

## Documentación afectada

- [PRODUCT.md](../project/PRODUCT.md)
- [IDENTITY.md](../engineering/IDENTITY.md)
- [ROADMAP.md](../project/ROADMAP.md)
- [LEARNING.md](../project/LEARNING.md)
- [DECISIONS.md](../governance/DECISIONS.md)
