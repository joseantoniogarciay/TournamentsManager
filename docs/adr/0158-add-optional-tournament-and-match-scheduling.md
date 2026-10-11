# ADR-0158: Fecha, hora y lugar opcionales de torneos y partidos

- **Estado:** Aceptado
- **Fecha:** 2026-10-11
- **Decisor:** Usuario, mediante «perfecto, adelante» tras revisar los diseños
- **Supera parcialmente:** ADR-0034, en el permiso delegado para estos datos informativos

## Problema, contexto y restricciones

Una persona organizadora puede necesitar coordinar encuentros reales, mientras
que un torneo rápido entre amigos no necesita planificación. El gate técnico y
la v1 están cerrados; esta capacidad es un incremento independiente.

## Alternativas y coste

- No añadir información: coste nulo, coordinación fuera del producto.
- Modos rápido/planificado o calendario automático: más controles, reglas y
  mantenimiento antes de demostrar necesidad.
- Campos opcionales dentro del detalle existente: coste moderado de contrato,
  persistencia, permisos y formularios; un único recorrido de producto.

## Recomendación y decisión del usuario

Se acepta la tercera alternativa y los diseños revisados. Creador y
administradores designados pueden editar el lugar común del torneo y la fecha,
hora y lugar propios de cada partido. Las demás personas solo leen los valores
existentes; no ven controles deshabilitados ni bloques vacíos.

El lugar común es opcional y se hereda al leer un partido sin lugar propio.
Vaciar el lugar propio restablece esa herencia. No se copia el lugar común a
cada partido. Los campos se editan después de crear el torneo, sin añadir
requisitos al alta, inicio o resultados. No hay mapas, reservas, recordatorios
ni planificación automática.

## Concreción y consecuencias

- Lugar: texto plano recortado, hasta 160 caracteres Unicode, sin controles.
- Fecha: día civil `YYYY-MM-DD`; hora: `HH:mm` de 24 horas. Son opcionales
  e independientes. Una hora conserva una zona IANA explícita; se muestra
  la hora de esa zona, sin convertirla a la del lector. No se inventa un instante
  UTC cuando falta fecha. Sin hora no se almacena zona.
- PUT reemplaza los campos del recurso; las cadenas vacías eliminan el dato.
  La lectura omite campos vacíos. Un formulario conserva su entrada al fallar.
- Se permite editar en torneos publicados o en curso. Los finalizados y
  cancelados conservan la información en lectura. Un bye no tiene planificación.
  Los datos no alteran resultados, clasificación ni la congelación deportiva.
- El backend verifica el permiso en una transacción serializada por torneo,
  también ante retirada de administradores o transferencia concurrente.
- Las dos operaciones devuelven la proyección completa y actualizan la actividad.
  Validación, permisos, ausencia y cierre tienen respuestas seguras; errores
  técnicos nunca incluyen fecha, lugar ni otros inputs en telemetría.

## Validación

Validar opcionalidad, fechas reales, hora y zona, límites de texto, eliminación
herencia y persistencia; propietario/delegado frente a lector, recurso cruzado,
retirada de permiso, torneo cerrado y bye. Recorrer respuestas HTTP, fallos de
persistencia y cancelación, parser de cliente, localización y exportación web.
La comprobación nativa se declara por separado si no se ha realizado.

## Disparadores de revisión

Avisos, calendario automático, cambios concurrentes de una misma planificación,
zonas múltiples por partido o necesidad de auditoría de esta información.

## Documentación afectada

PRODUCT.md, API.md, DATABASE.md, DESIGN_SYSTEM.md, OBSERVABILITY.md,
DECISIONS.md, LEARNING.md y CHANGELOG.md.

## Evidencia de cierre — 2026-10-11

- `make verify` completo en verde: Go/TypeScript, suites de cliente,
  OpenAPI/sqlc sin deriva, exportación web, build y vulnerabilidades aplicables.
- Suite PostgreSQL real `-run Integration -count=1` en base Docker efímera,
  esquema inicial y 21 migraciones; permiso, transferencia, cierre y bye.
- Down/Up de 00021 comprobados únicamente en esa base aislada.
- UI web real con datos ficticios: escritorio y 390×844, formularios, validación
  por blur, guardado/herencia y lectura sin permisos ni datos.
- Contenedor efímero y servidor de preview detenidos al cerrar. No se publican
  producción ni migraciones en bases de despliegue. QA visual iOS/Android
  pendiente; la exportación web no acredita esa revisión.
