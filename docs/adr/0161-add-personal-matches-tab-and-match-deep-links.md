# ADR-0161: Mis partidos y acceso al encuentro dentro del torneo

- **Estado:** Aceptado
- **Fecha:** 2026-10-11
- **Decisor:** Usuario
- **Propietario del análisis:** Codex
- **Supera parcialmente:** ADR-0057, en la composición de la botonera

## Problema

Una participante debe abrir cada torneo y localizar su encuentro. Se propone
reunir los partidos personales y abrir el detalle en el encuentro elegido.

## Contexto y evidencia

El gate técnico y v1 están cerrados. Una cuenta representa un equipo por torneo;
administración y seguimiento no equivalen a participación. Fecha, hora y lugar
son opcionales (ADR-0158). La ficha actual usa listas y cuadros, con varias fases.
El usuario propone una sección de la tab bar y exige estados vacíos y de reintento.

## Criterios

Participación real, navegación contextual, ausencia de datos inventados,
recuperación segura, accesibilidad y paridad web/iOS/Android.

## Alternativas y comparación

- Mantener la ficha actual: coste nulo; sigue exigiendo buscar entre torneos.
- Filtro personal dentro de cada torneo: coste bajo–medio; no reúne encuentros.
- Tab personal y destino por partido: coste moderado en consulta, contrato,
  navegación y QA; mantenimiento moderado por listas y cuadros. Recomendación,
  porque resuelve ambos pasos sin otra pantalla de detalle de partido.

## Recomendación de comportamiento

- Botonera: Inicio, Torneos, Mis partidos, Cuenta. Sin sesión conserva la sección
  con explicación y acceso al login; nunca consulta una colección personal.
- Pendientes y Resultados reúnen encuentros del equipo vinculado a la sesión.
  Administrar o seguir sin ese vínculo no incluye encuentros.
- Pendientes contiene encuentros reales pendientes de torneos en curso con la
  participación activa. Resultados contiene encuentros resueltos, incluidas
  incidencias y resultados administrativos, también en torneos finalizados.
  Un bye no es un encuentro y no aparece. Torneos aún sin iniciar no inventan
  partidos; cancelados no ofrecen pendientes. No se inventan rivales de rondas
  futuras: mostrar solo partidos con ambos participantes determinados.
- Orden propuesto: pendientes con fecha ascendente, luego sin fecha; hora
  ausente se identifica, sin inventar un instante. Fechas/horas conservan la
  zona del partido según ADR-0158. Resultados por resolución más reciente.
  Desempates de orden estables y paginación se concretarán contract-first.
- Tarjeta: torneo, participantes, fase/ronda cuando ayude, programación existente
  o Sin fecha, y resultado/incidencia cuando corresponda.
- Abrir usa la ficha canónica con identificador de partido: selecciona fase,
  grupo/ronda y posición vertical/horizontal necesarias, revela el encuentro y
  lo identifica brevemente. Respeta movimiento reducido y anuncia el destino
  de forma accesible. No abre el editor ni cambia permisos.
- Volver desde el recorrido iniciado en la lista restaura pestaña y posición;
  enlaces en frío y atrás web respetan ADR-0055. Un partido que ya no exista
  produce feedback seguro y permite consultar el torneo si sigue disponible.

## Estados y recuperación

En iOS y Android la actualización se hace con pull to refresh, también sobre
vacíos autenticados, sin botón de refresh en cabecera. Conserva la vista y el
contenido durante la consulta. El error inicial conserva el botón Reintentar.
Web sigue su convención de actualización. Corrección explícita del usuario.

| Situación | Presentación y acción |
| --- | --- |
| Sin sesión | Explicación y acceso al login |
| Carga inicial | LoadingTransition centrado; no flash de vacío |
| Sin partidos en ninguna vista | Explicación: participar y generar encuentros; acceso a Torneos |
| Solo Pendientes vacío | No tienes partidos pendientes; Resultados sigue disponible |
| Solo Resultados vacío | Todavía no tienes resultados; Pendientes sigue disponible |
| Fallo inicial | RequestErrorCard con Reintentar en la vista seleccionada |
| Fallo de actualización | Conservar contenido confirmado, banner seguro y nueva actualización |
| Fallo de página adicional, si existe | Conservar filas, reintento en el pie sin duplicados |
| Torneo inaccesible al abrir | Estado seguro según lectura vigente, sin revelar datos |
| Partido ausente en torneo disponible | Aviso específico y acceso al torneo |
| Cancelación intencionada | Sin feedback de error |

Transporte identificado por apiFetch usa common_network_error; errores técnicos,
5xx, cuerpos inválidos y problemas no reconocidos usan common_request_error.
Nunca mostrar title/detail del servidor. No presentar vacío si la consulta falló.
Las respuestas vacías de una página no prueban que ambas colecciones estén vacías:
el contrato debe aportar evidencia suficiente para distinguir esos estados.
Reintentos bloquean solicitudes duplicadas; cambiar sesión descarta datos previos.

## Decisión del usuario

**Aceptado el 2026-10-11**, tras revisar y corregir la maqueta: «Parece que todo
esta bien a futuro revisaré pero vayamos con la implementación, si tienes dudas
o hay cosas que se salgan del estilo que ya tenemos me preguntas antes».
Se implementa el alcance descrito y el pull to refresh acordado. Cualquier
ampliación material o cambio de estilo requiere consulta previa.

## Consecuencias

Mayor utilidad para participantes; nueva colección autenticada y cuarta tab.
No incorpora recordatorios, calendario, caché persistente ni otra librería.
La consulta se proyecta en backend sin descargar cada torneo desde el cliente.
La lógica de participación permanece independiente de transporte y persistencia.

## Validación prevista

Equipo vinculado frente a seguimiento/delegación; sesión cambiada/cerrada;
torneos sin empezar, activos, terminados/cancelados; retiradas, byes, incidentes,
rivales aún indeterminados, fechas incompletas y zonas distintas. Todos los
estados de la tabla, reintento repetido, carrera de actualización y paginación.
Enlaces a liga, grupos, desempate y cuadro, incluidos cambios de fase; posición
restaurada, inicio en frío, teclado/lector, escalado, movimiento reducido,
temas e idiomas. QA web y nativa se acredita por separado.

Recorrer explícitamente éxito, validación, rate limit, autorización, negocio,
fallos de cada límite técnico y cancelación del nuevo endpoint. El contrato
declara respuestas; spans raíz conservan plantilla y causas cerradas seguras
según OBSERVABILITY.md. No exportar equipos, cuentas, IDs ni entradas.

## Disparadores de revisión

Agenda unificada en zonas distintas, arbitraje, plantillas individuales,
pendientes de torneos cancelados, necesidad de datos offline o volumen medido.

## Documentación afectada

PRODUCT.md, API.md, DATABASE.md, DESIGN_SYSTEM.md, OBSERVABILITY.md,
DECISIONS.md y LEARNING.md tras aceptación e implementación.
Diseño para revisión: [propuesta](../engineering/PRODUCT_PROPOSAL_MATCHES_EDITIONS_2026-10-11.md).

## Estado de implementación

Implementado. [Evidencia y límites de QA](../engineering/PRODUCT_QA_MATCHES_EDITIONS_2026-10-11.md);
QA de dispositivos nativos y matriz visual ampliada pendientes, declaradas por separado.
