# ADR-0142: Registrar incidencias por partido en todos los deportes

- **Estado:** Aceptado
- **Fecha:** 2026-10-03
- **Decisor:** Usuario
- **Supera a:** ADR-0039 y los perfiles deportivos, solo respecto a incidencias por partido

## Problema, contexto y evidencia

Una incomparecencia o abandono no puede registrarse como un marcador realmente
jugado. La retirada de un equipo en liga modifica todos sus partidos; no sirve
para una incidencia individual. El gate técnico está cerrado. La app dispone de
resultados, historial atómico, clasificación y dependencias entre fases/cuadro.

## Alternativas y coste

1. Solo raqueta/eliminatoria: coste moderado, deja ligas y otros deportes fuera.
2. Incidencias cerradas para todos los deportes y fases admitidas: coste medio,
   exige separar tanteo real parcial y resultado administrativo por deporte.
3. Motor de sanciones/configuración por torneo: coste alto y reglas arbitrarias;
   sobreingeniería para dos casos. No se adopta.

## Decisión explícita

El usuario solicitó «Habria que hacerlo para todo», aprobó las maquetas del
selector Marcador/Incidencia dentro del diálogo y autorizó «Perfecto vamos con
ello». A la pregunta específica sobre resultado fijo para ambas incidencias y
tanteo parcial separado respondió «Vamos con eso si».

Se adopta la alternativa 2, sin cambiar los formatos admitidos por cada deporte:

- `no_show`: un lado no se presenta; no admite tanteo parcial.
- `retirement`: un lado abandona; admite tanteo parcial opcional. La app elige el
  lado afectado; el rival recibe la victoria incluso si iba perdiendo.
- Liga y desempate de clasificación: fútbol 3–0, baloncesto 20–0, balonmano 10–0,
  voleibol 3–0 con tres sets 25–0. La clasificación usa exclusivamente el
  resultado administrativo, con puntos/cocientes y política de empate vigentes.
- Eliminatoria: los mismos perfiles para deportes de equipo; raqueta adjudica
  los sets/juegos necesarios al rival, sin inventar parciales. La UI identifica
  la incidencia y la ganadora, sin presentar el agregado como tanteo jugado.
- Tanteo parcial separado: goles/puntos para deportes sin sets; para deportes
  por sets se admiten sets completos y un último parcial, incluidos empates.
  No se admiten sets después de ganar el partido, múltiples sets incompletos,
  tanteos por encima del límite ni sets que ya deberían haber terminado.
  Tenis/pádel no incluyen tanteo interno de puntos ni del tie-break.
- Las incidencias no retiran al participante ni modifican otros partidos. Solo
  organizadora/administradores durante la fase activa, con historial y cierre
  final explícito. No hay cambios al roster ni permisos nuevos.
- Correcciones de incidente a jugado o a otro incidente guardan historial. Se
  conservan los bloqueos por descendientes, fase congelada y ciclos resueltos.
  La retirada completa existente conserva su política y bloqueo de voleibol.
- Doble ausencia, suspensión, expulsión y reglas configurables quedan fuera:
  no se elige arbitrariamente una ganadora.

## Presentación y contrato

Un único acceso Añadir/Editar resultado; selector Marcador/Incidencia debajo de
participantes. Incidencia muestra motivo, lado afectado, parcial opcional si hay
abandono y victoria antes de Guardar dentro del scroll. La lectura conserva el
motivo, ganadora, tanteo parcial y resultado administrativo de liga etiquetado.
Cuatro idiomas, primitivas/tokens y feedback seguro existentes.

La operación de resultado existente acepta una tercera variante exclusiva
`incident`. La proyección pública incluye su metadata tipada. Se conserva el
historial de instantáneas, con metadata actual y previa. PostgreSQL usa JSONB
para el objeto pequeño y cerrado, validado por el dominio, sin otro motor ni
servicio. Migración incremental; no se reescribe historia ni despliega aquí.

## Validación y cierre

Pruebas de dominio por deporte y límite parcial; HTTP para formas exclusivas,
respuestas y trazas seguras; PostgreSQL desechable para lectura, clasificación,
cuadro, correcciones, retirada e historial. Revisión visual móvil, cuatro
catálogos, typecheck, lint y exportación web. Recorrer éxito, validación, tasa,
rechazo de negocio, cada fallo técnico y cancelación. Actualizar API, producto,
modelo, observabilidad y aprendizaje con retrospectiva al cerrar.

### Evidencia y retrospectiva

Implementación completa en dominio, contrato, PostgreSQL y cliente. Suite Go y
PostgreSQL desechable aprobadas, con matriz de ocho deportes en cuadro, cuatro
en liga y desempate mixto; correcciones, dependencias e historial verificados.
Pruebas del módulo cliente real, lint, typecheck, OpenAPI y exportación web
aprobados. Revisión visual del JSX real a 320×480 y 390×844; ruta de prueba
eliminada, sin enviar resultados. Nombres largos ajustados y Guardar accesible
por scroll. No hay despliegue ni migración en entornos persistentes. El teclado
nativo sigue pendiente de comprobación en dispositivo; se mantiene el aviso
Expo de compatibilidad aplazado por ADR-0138.

Retrospectiva: un normalizador puro y metadata cerrada cubren la necesidad sin
motor de sanciones. Separar tanteo parcial evita contaminar clasificación;
preservar instantáneas y metadata al resolver otros partidos del cuadro exige
pruebas de corrección y persistencia, además del primer guardado.
