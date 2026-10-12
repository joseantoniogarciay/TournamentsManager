# ADR-0162: Consultar ediciones relacionadas desde el torneo

- **Estado:** Aceptado
- **Fecha:** 2026-10-11
- **Decisor:** Usuario
- **Propietario del análisis:** Codex
- **Supera parcialmente:** ADR-0159, ampliando lectura de ediciones

## Problema

Repetir conserva torneos independientes, pero la ficha no permite recorrer
ediciones relacionadas. El usuario propone Ediciones cuando haya al menos dos.

## Contexto y evidencia

Gate técnico y v1 cerrados. ADR-0159 conserva identidades, resultados, reglas y
permisos independientes. La migración 00022 ya guarda repetition_source_id:
existe procedencia explícita, aunque todavía no una colección de ediciones.
No deducir relaciones por nombre ni reconstruir asociaciones inventadas.

## Criterios

Histórico intacto, relación explícita, mismo recorrido desde cualquier edición,
lectura y autorización actuales, consistencia concurrente y mantenimiento mínimo.

## Alternativas y comparación

- Mantener torneos separados: coste nulo y ninguna navegación entre ediciones.
- Recorrer procedencia existente y sus descendientes: coste moderado de consulta
  y contrato; evita otro agregado. Mantenimiento bajo–medio, sujeto a medir la
  consulta. Recomendación inicial.
- Introducir entidad Serie y membresía: coste mayor en datos, migración y
  mutaciones; útil con gestión de temporadas, pero excesivo para esta lectura.

## Recomendación de comportamiento

- Una familia incluye el torneo raíz y todos sus descendientes de repetición.
  Repetir una edición antigua crea una rama de la misma familia; no otra serie.
- El botón Ediciones aparece en todas sus fichas cuando hay al menos dos.
  Para una única edición no se muestra bloque vacío ni control deshabilitado.
- Lista de creación más reciente a más antigua, con desempate estable. No
  ordenar por nombre ni atribuir fechas deportivas inexistentes.
- Mostrar nombre, estado y campeones confirmados cuando esté finalizada;
  identificar la edición consultada. Pulsar abre su ficha canónica.
- La lista incluye publicadas, en curso, finalizadas y canceladas que permitan
  lectura conforme a la política actual. El historial no concede permisos ni
  transmite propiedad, participantes, seguidores o delegados entre ediciones.
- No habilita unir/separar familias manualmente ni estadísticas acumuladas.
  Los nombres iguales de equipos no acreditan identidad entre ediciones.
- Las repeticiones previas con procedencia persistida se pueden incluir; torneos
  sin relación registrada siguen independientes. No cambia el flujo de repetición.

## Estados y recuperación

En apps se actualiza con pull to refresh, sin botón de refresh en cabecera.
El error inicial usa Reintentar: carga únicamente Ediciones, mantiene el torneo
consultado y recupera su lista; nunca muestra Mis partidos como resultado.

Carga inicial centrada y error seguro con Reintentar; fallo de actualización
conserva lista y muestra banner. Reintento de página adicional conserva filas
si se requiere paginación. La edición consultada queda visible/identificada sin
duplicados. Un enlace directo a una familia con una sola edición muestra esa
edición; no inventa una segunda. Si la ficha base no se puede leer, usa su estado
seguro vigente. Cancelaciones intencionadas no muestran error.
El botón necesita un dato confirmado en la proyección de lectura; un fallo de
consulta no equivale a una familia de una edición. Transporte y fallos técnicos
siguen common_network_error/common_request_error, sin errores brutos del backend.

## Decisión del usuario

**Aceptado el 2026-10-11**, tras revisar y corregir la maqueta: «Parece que todo
esta bien a futuro revisaré pero vayamos con la implementación, si tienes dudas
o hay cosas que se salgan del estilo que ya tenemos me preguntas antes».
Se implementa el alcance descrito y el pull to refresh acordado. Cualquier
ampliación material o cambio de estilo requiere consulta previa.

## Consecuencias

Continuidad sin alterar resultados ni crear identidad global de equipos.
La lectura debe resolver ancestros y descendientes de forma segura, sin N+1
desde el cliente. Medir índices/consulta antes de adoptar otra entidad; no
introducir materialización ni caché sin evidencia. La regla de familia pertenece
al negocio, la consulta SQL a su adaptador. Visibilidad por enlace actual implica
que conocer una edición permite descubrir las otras vinculadas: revisar este
efecto explícitamente al aceptar el diseño.

## Validación prevista

Una edición, cadena, ramas desde una antigua, cambios de nombre, transferencia
de propiedad, diferentes estados, campeones compartidos y torneos independientes
con el mismo nombre. Reintentos/concurrencia de repetición no duplican ediciones.
Colección idéntica desde cualquier miembro y límite de consulta con datos reales.
Mostrar/ocultar botón, estado consultado, carga/error/reintento, navegación en
frío y regreso, accesibilidad, temas, idiomas y QA web/nativa por separado.
Para endpoints nuevos/modificados recorrer éxito, validación, rate limit,
rechazos de negocio/autorización, cada fallo técnico y cancelación; documentar
contrato, inventario de OBSERVABILITY.md y causas cerradas seguras en span raíz.

## Disparadores de revisión

Series privadas, desvinculación, temporadas editables, estadísticas entre
ediciones o coste de consulta medido que justifique identidad de familia.

## Documentación afectada

PRODUCT.md, API.md, DATABASE.md, DESIGN_SYSTEM.md, OBSERVABILITY.md,
DECISIONS.md y LEARNING.md tras aceptación e implementación.
Diseño para revisión: [propuesta](../engineering/PRODUCT_PROPOSAL_MATCHES_EDITIONS_2026-10-11.md).

## Estado de implementación

Implementado. [Evidencia y límites de QA](../engineering/PRODUCT_QA_MATCHES_EDITIONS_2026-10-11.md);
QA de dispositivos nativos y matriz visual ampliada pendientes, declaradas por separado.
