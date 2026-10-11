# ADR-0159: Repetir torneos finalizados como nuevas ediciones

- **Estado:** Aceptado
- **Fecha:** 2026-10-11
- **Decisor:** Usuario, mediante «Si, perfecto palante» tras revisar la propuesta
- **Propietario del análisis:** Codex
- **Supera a:** Ninguno
- **Superado parcialmente por:** ADR-0160, en el alcance general de elegir equipos mediante invitación

## Problema, contexto y criterios

Repetir un torneo acabado exige volver a introducir su composición y reglas.
El gate técnico está cerrado. Se debe conservar el histórico, verificar la
propiedad actual y mantener independientes las identidades y permisos de cada
edición, con el mínimo mantenimiento sobre el ciclo existente.

## Alternativas y comparación

- Crear desde cero: coste de desarrollo nulo, pero repite el trabajo manual.
- Nueva edición con revisión: coste moderado en contrato, persistencia y cliente;
  conserva el histórico y reutiliza creación, inscripciones e inicio.
- Reiniciar el original: menos pasos, pero altera el histórico y mezcla permisos,
  resultados y enlaces. Su mantenimiento y riesgo son mayores.

## Recomendación y decisión del usuario

Se acepta la nueva edición con revisión, mismos equipos y configuración, e
inscripciones renovadas. Solo la persona propietaria actual puede repetir un
torneo `completed`. Revisa nombre y equipos y elige cuál será su equipo; se
copian deporte y reglas deportivas. El formato y la configuración de liga o
fase clasificatoria se conservan como propuesta editable antes del inicio.

El nuevo torneo nace `published`, con ID y equipos nuevos, sin partidos,
resultados, retiradas, campeones, programación, administradores, seguidores ni
invitaciones anteriores. No se copia el lugar común. El original no cambia.
El primer equipo enviado se vincula a la organizadora como en el alta actual;
las demás cuentas no se arrastran. En una nueva edición, su nueva invitación
permite reclamar un equipo existente sin cuenta por coincidencia de nombre
normalizado, o crear uno nuevo si hay capacidad. No verifica identidad deportiva:
como la inscripción actual, depende del acceso a la invitación compartida.
El usuario concreta que quienes entran no deben ir a ciegas: la invitación
muestra los nombres de equipos libres como opciones y conserva el campo de
nombre libre para crear otro si hay capacidad. Un equipo ocupado se retira de
las opciones al recargar; el servidor resuelve carreras de inscripción.

## Concreción y consecuencias

POST de repetición valida propiedad y cierre en la misma transacción que crea
la edición. Una identidad UUID de intención evita duplicados por reintentos,
incluso concurrentes; cada nueva revisión recibe una intención diferente.
Las reglas heredadas se leen en servidor. La configuración propuesta se guarda
sin materializar fases ni partidos y se descarta al iniciar. Su compatibilidad
con la composición se comprueba al iniciar con las reglas ya aceptadas.

La nueva pantalla reutiliza primitivas y patrones de cierre del cliente; no
sobrescribe el borrador local de creación. Errores de permiso, estado y ausencia
permiten feedback localizado; transporte y errores técnicos mantienen el fallback
seguro. Ni observabilidad ni fallos HTTP exportan entradas o errores internos.

## Validación

Propietaria actual frente a delegado, lector, transferencia y estados abiertos;
reintentos concurrentes; independencia de IDs, histórico y relaciones; todos los
formatos y reglas de deportes; reclamación única, equipo ocupado y capacidad 64.
Recorrer éxito, validación, autenticación/CSRF, rechazo de negocio, fallos de
base y cancelación. Verificar parser, navegación directa, formulario, feedback,
locales, typecheck y exportación web. Declarar por separado la QA nativa.

## Disparadores de revisión

Ediciones enlazadas para estadísticas, reclamaciones con verificación de identidad,
plantillas reutilizables, repetición por delegados o planificación recurrente.

## Documentación afectada

PRODUCT.md, API.md, DATABASE.md, DESIGN_SYSTEM.md, OBSERVABILITY.md,
DECISIONS.md, LEARNING.md y CHANGELOG.md.
