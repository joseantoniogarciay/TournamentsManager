# ADR-0181: Mi perfil privado con resultados y estadísticas

- **Estado:** Aceptado
- **Fecha:** 2026-11-02
- **Decisor:** Usuario
- **Supera parcialmente a:** ADR-0161, solo en las nuevas proyecciones privadas.

## Requisitos aceptados

Cuenta incorpora una fila Mi perfil. Esa ruta reúne analítica privada a la cuenta
actual, historial de partidos y resultados de sus participaciones, separados por
deporte. El usuario elige historial deportivo, no auditoría de las modificaciones
realizadas como administrador. No hay perfil público ni clasificación global.

## Evidencia y restricciones

Mis partidos ya relaciona cuenta y equipo; el backend conserva cambios internos
de resultados. Ninguna de estas capacidades equivale a estadísticas personales.
La app no tiene jugadores ni plantillas: participación significa equipo o
participante vinculado, no rendimiento individual de una persona en una plantilla.
La privacidad se aplica en backend, no solo escondiendo una ruta o enlazándola
desde Cuenta. Los resultados públicos de torneos conservan su visibilidad previa.

## Alternativas y mantenimiento

Calcular en cliente descargando páginas: coste aparente bajo, totales incompletos
y acoplamiento al volumen. Proyección agregada backend: coste moderado, totales
consistentes y control de privacidad. Almacén analítico separado: mantenimiento
alto y complejidad sin evidencia de necesidad. Se recomienda proyección backend,
sin nuevo servicio ni dashboard externo.

## Alcance de la primera implementación

Ruta autenticada propia, sin parámetro de otra cuenta; historial paginado con
filtros por deporte/torneo y agregados calculados sobre el conjunto completo.
Métricas aceptadas: disputados, victorias, empates cuando el deporte los
admita, derrotas y porcentaje de victorias por deporte; porcentaje sin partidos
sin división por cero. No sumar goles, puntos y sets en una misma métrica.
Correcciones de resultados deben reflejarse; sin inferir participación por ser
organizador, delegado o seguidor. Sin comparar contra otras cuentas.

El usuario acepta explícitamente «Ese criterio deportivo (recomendado)»:
los agregados incluyen únicamente partidos realmente jugados. Byes y resultados
administrativos quedan fuera tanto del numerador como del denominador de las
métricas; los registros pertinentes permanecen identificados en el historial.
El porcentaje de victorias usa victorias / disputados; sin partidos jugados se
presenta sin datos, no como una derrota ni como una tasa demostrada del 0 %.
Los empates solo se muestran cuando el deporte/formato los permita.

El contrato y las consultas deben verificar la evidencia disponible ante
correcciones, retiradas e incidencias. No deducir un resultado jugado a partir de
un marcador administrativo ni convertir byes en partidos disputados. Las reglas
se prueban en backend antes de publicar la proyección; los totales no se calculan
sumando únicamente las páginas descargadas.

## Validación

Aislamiento entre cuentas y sesiones, ausencia de participación, múltiples
nombres de equipo, varios deportes/ediciones, byes e incidencias, correcciones y
retiradas; precisión de denominadores; paginación sin totales parciales, UTC y
orden estables, sin PII en observabilidad, estados de carga/vacío/error y accesibilidad.

## Decisión y retrospectiva

Ubicación, privacidad, historial deportivo, separación por deporte y criterio de
cómputo aceptados por el usuario. Implementación, contrato y pruebas pendientes. Analítica privada no convierte
los partidos públicos en datos privados ni atribuye estadísticas individuales
que el modelo de equipos no puede demostrar.

## Documentación afectada

PRODUCT, API, DATABASE, SECURITY, OBSERVABILITY, DESIGN_SYSTEM, DECISIONS y LEARNING.


## Implementación — 2026-11-02

Implementado el contrato GET /me/profile, consulta sqlc sin migración, agregado
completo, historial paginado con filtros de deporte y nombre del torneo, y ruta
Cuenta → Mi perfil con catálogos es/en/it/fr. Se aplica el resultado oficial
vigente: una retirada que lo convierte en administrativo lo excluye del agregado;
no se reconstruyen marcadores antiguos. Pruebas de aislamiento, correcciones,
byes, incidencias, penaltis y ocho deportes; QA visual y publicación pendientes.
[Alcance, checklist y evidencia](../engineering/PRIVATE_PROFILE_2026-11-02.md).


Publicado en web/API/renderer como v1.16.0; QA visual y distribución móvil
pendientes. Véase la sección de publicación del informe de implementación.
