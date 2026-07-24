# ADR-0013: Usar `develop` como rama de integración

- **Estado:** Aceptado
- **Fecha:** 2026-07-24
- **Decisor:** Usuario, mediante confirmación explícita
- **Propietario del análisis:** Asistente como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

El repositorio necesita un flujo de ramas que mantenga `main` estable sin imponer
una rama y una pull request para cada cambio durante una etapa de trabajo
principalmente individual.

## Contexto y restricciones

- El monorepo es público y `main` es su rama por defecto.
- El trabajo avanza por bloques técnicos antes de empezar el detalle funcional.
- Inicialmente hay un único desarrollador y se busca evitar proceso innecesario.
- Los cambios importantes conservan su trazabilidad mediante ADR y commits
  coherentes, aunque no tengan una rama por feature.
- CI, protecciones de rama y mecanismo exacto de promoción siguen pendientes.

## Criterios

1. mantener `main` como estado estable y comprensible;
2. reducir fricción en el trabajo diario individual;
3. permitir agrupar y verificar bloques antes de publicarlos en `main`;
4. conservar una vía para aislar cambios de riesgo o trabajo paralelo;
5. poder añadir revisión y protecciones sin rediseñar todo el flujo.

## Alternativas

### Alternativa A — Trunk-based sobre `main`

Todo cambio se integra directamente o mediante ramas muy cortas en `main`.

- **Ventajas:** historial y flujo simples; evita divergencia entre ramas largas.
- **Inconvenientes:** exige que cada commit de `main` sea integrable y que CI dé
  feedback rápido; ofrece menos separación entre trabajo en curso y bloques
  terminados.
- **Coste de mantenimiento:** bajo.

### Alternativa B — `develop` sin ramas por feature como norma

El trabajo diario se realiza en `develop`. Al terminar y verificar un bloque,
`develop` se promociona a `main`.

- **Ventajas:** `main` representa hitos estables; poca ceremonia durante el
  desarrollo; permite verificar un bloque completo antes de promocionarlo.
- **Inconvenientes:** dos ramas de larga vida pueden divergir; un cambio
  incompleto en `develop` puede retrasar toda la promoción; hay menos aislamiento
  y revisión previa entre cambios simultáneos.
- **Coste de mantenimiento:** bajo mientras trabaje una sola persona; aumenta con
  colaboradores o varios cambios paralelos.

### Alternativa C — Rama por feature y pull request

Cada cambio nace en una rama temporal y se integra tras revisión.

- **Ventajas:** aislamiento, revisión granular y CI previo al merge.
- **Inconvenientes:** más ramas y ceremonia para una sola persona; puede
  fragmentar artificialmente bloques de aprendizaje pequeños.
- **Coste de mantenimiento:** medio en la etapa actual.

## Comparación

La alternativa A es la más simple y suele encajar con entrega continua. La C
aporta el mejor aislamiento cuando existe colaboración o trabajo paralelo. La B
acepta el coste de una rama larga adicional a cambio de conservar `main` como
escaparate estable y mantener fluido el trabajo individual.

## Recomendación

**Opinión/recomendación:** alternativa B en la etapa actual, con `develop`
actualizada respecto de `main` y excepciones explícitas para no forzar cambios
arriesgados o paralelos dentro de una única rama.

## Decisión del usuario

**Aceptada:** trabajar habitualmente sobre `develop`, sin crear ramas por feature.
Al terminar bloques coherentes y verificados, se integrarán en `main` y el trabajo
continuará de nuevo en `develop`.

Podrán usarse ramas temporales cuando exista un hotfix de producción, un
experimento arriesgado, colaboración paralela o una necesidad real de aislamiento.

## Consecuencias

### Positivas

- `main` muestra bloques terminados.
- El trabajo cotidiano requiere poca gestión de ramas.
- Un bloque puede verificarse completo antes de promocionarlo.

### Negativas y deuda aceptada

- `develop` puede acumular trabajo no publicable.
- Los cambios simultáneos dentro de `develop` no quedan aislados entre sí.
- Hay que sincronizar `develop` después de cualquier hotfix en `main`.
- El modelo debe revisarse si crece el equipo o se necesita entrega continua.

## Reglas operativas

- El trabajo ordinario y sus commits se hacen sobre `develop`.
- No se hacen commits cotidianos directamente en `main`.
- Antes de promocionar un bloque se ejecutan las verificaciones aplicables y se
  revisa el diff respecto de `main`.
- Tras integrar un bloque, `develop` debe contener el estado de `main` antes de
  continuar.
- No se reescribe el historial compartido de `main` ni `develop`.
- La estrategia exacta de merge, CI y protección de ramas se decidirá en el gate
  de CI y política de calidad.

## Validación

- Existen `main` y `develop` en local y en GitHub.
- El workspace queda situado en `develop`.
- `develop` parte del último commit publicado en `main`.
- La guía de contribución describe el mismo flujo.

## Disparadores de revisión

- Incorporación de colaboradores habituales.
- Necesidad de desarrollar o publicar varios bloques en paralelo.
- `develop` permanece sin poder promocionarse durante periodos prolongados.
- Se adopta entrega continua desde `main`.
- Un incidente exige una política más estricta de revisión o protección.

## Documentación afectada

- [CONTRIBUTING.md](../../CONTRIBUTING.md)
- [TECHNICAL_BASELINE.md](../governance/TECHNICAL_BASELINE.md)
- [DECISIONS.md](../governance/DECISIONS.md)
- [LEARNING.md](../project/LEARNING.md)

