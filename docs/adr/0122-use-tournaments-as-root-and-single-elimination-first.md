# ADR-0122: Usar torneos como recurso raíz e incorporar la eliminatoria directa

- **Estado:** Aceptado
- **Fecha:** 2026-09-06
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** ADR-0032, ADR-0037, ADR-0038, ADR-0039, ADR-0040, ADR-0041,
  ADR-0043, ADR-0049, ADR-0058, ADR-0060, ADR-0070, ADR-0073, ADR-0079,
  ADR-0080, ADR-0081, ADR-0082, ADR-0083, ADR-0085 y ADR-0121, exclusivamente
  en la nomenclatura de recurso y en el alcance de liga como único formato
- **Superado por:** Ninguno

## Problema

El producto ya gestiona una liga de fútbol, pero ahora debe permitir una
eliminatoria. Un recurso llamado `league` no puede representar de forma honesta
un cuadro cuyos participantes futuros proceden de ganadores anteriores ni una
competición sin clasificación. Además, el formato mixto liga más eliminatoria es
una evolución plausible, sin que hoy exista información suficiente para diseñar
sus reglas de clasificación.

## Contexto y restricciones

- El cliente, contrato, dominio, persistencia, enlaces públicos y documentación
  actuales usan `league` como nombre del recurso.
- El producto inicial mantiene fútbol y equipos como participantes.
- La persona usuaria acepta una migración incompatible: no se conservarán rutas,
  contratos ni enlaces canónicos de `leagues`.
- La liga existente sigue siendo un formato de torneo. El primer formato nuevo
  es `single_elimination`: partido único, de 2 a 64 equipos, cuadro congelado
  al iniciar y *byes* deterministas si el total no es potencia de dos.
- Cada partido resuelto debe identificar una única ganadora. Un marcador igual
  exige un desempate explícito; no puede dejar una plaza posterior ambigua.
- Organizador y administradores delegados mantienen la gestión inmediata y
  trazable de resultados. Una corrección que afectaría a un partido descendiente
  ya resuelto se rechaza.
- No se decide todavía una entidad o motor genérico de fases, clasificación
  entre fases, doble eliminación, tercer puesto, ida/vuelta ni reordenación del
  cuadro.

## Criterios de decisión

1. que API y lenguaje de dominio describan correctamente ambos formatos;
2. preservar invariantes de avance y una única campeona en el cuadro;
3. evitar duplicar permisos, equipos, enlaces y ciclo de vida;
4. no anticipar un sistema de fases sin reglas de producto aceptadas;
5. asumir explícitamente el coste de una migración incompatible ahora, antes de
   consolidar más consumidoras.

## Alternativas

### Alternativa A — Extender `leagues` con `format: bracket`

- **Ventajas:** menor cambio inmediato de rutas y tablas.
- **Inconvenientes:** una eliminatoria seguiría expuesta como liga; respuestas
  con clasificación y co-campeones perderían semántica.
- **Coste de adopción:** bajo inicialmente.
- **Coste de mantenimiento:** creciente por condiciones de formato en todo el
  contrato y cliente.
- **Riesgos:** deuda de vocabulario y evolución más difícil hacia formatos
  mixtos.

### Alternativa B — Recurso `tournaments` y primer formato de eliminatoria

- **Ventajas:** el recurso raíz representa la competición, permite que formato
  determine su proyección y conserva una única base para equipos, permisos y
  enlaces.
- **Inconvenientes:** exige migrar dominio, esquema, OpenAPI, rutas, cliente,
  previews y documentación.
- **Coste de adopción:** alto, aceptado por la persona usuaria.
- **Coste de mantenimiento:** bajo a medio; los formatos contienen sus reglas
  sin fingir que son una liga.
- **Riesgos:** diseñar accidentalmente una abstracción de fases antes de conocer
  sus reglas.

### Alternativa C — Recurso `brackets` paralelo a `leagues`

- **Ventajas:** aísla el cambio y conserva las rutas existentes.
- **Inconvenientes:** duplica el ciclo de vida, permisos, equipos, lectura,
  seguimiento y enlaces; no crea una base limpia para una competición mixta.
- **Coste de adopción:** medio.
- **Coste de mantenimiento:** alto por dos recursos equivalentes.
- **Riesgos:** divergencia de comportamientos y migración posterior inevitable.

### No cambiar

El producto seguiría limitado a ligas y un bracket se modelaría de forma
incoherente o no se podría implementar.

## Comparación

La alternativa A ahorra una migración pero convierte el nombre y varias
proyecciones del contrato en falsedades. La C aplaza el problema a costa de
duplicar capacidades transversales. La B paga una única migración incompatible,
autorizada expresamente, y mantiene el mínimo modelo extensible: torneo raíz y
un formato concreto, sin fases genéricas.

## Recomendación

**Opinión/recomendación:** alternativa B. Es la mínima estructura que hace
correcta la terminología sin adelantar el futuro formato mixto.

## Decisión del usuario

**Aceptada el 2026-09-06:** adoptar la alternativa B y ejecutar una migración
incompatible completa. El recurso canónico, sus rutas y enlaces pasan a llamarse
`tournaments`; no se mantendrá compatibilidad con `leagues`.

La liga existente se migra y se conserva como formato `league`. El primer
formato **nuevo** implementado será `single_elimination`, a partido único. Al
iniciar, el servidor fija la siembra, materializa el árbol completo y crea
huecos derivados de ganadores o *byes*. La interfaz permitirá recorrer rondas y
ver el origen de cada hueco. Toda eliminatoria termina cada partido con una
única ganadora; los empates de marcador se resuelven con desempate explícito.

## Consecuencias

### Positivas

- El API, dominio y experiencia llaman torneo a una competición y bracket a su
  formato, sin ambigüedad.
- El servidor puede validar el avance del cuadro y devolver una proyección que
  explica sus plazas futuras.
- Liga y eliminatoria comparten identidad, equipos, permisos y visibilidad sin
  fingir que tienen la misma estructura deportiva.

### Negativas y deuda aceptada

- Se rompen rutas `/v1/leagues`, clientes generados, enlaces públicos y nombres
  de tablas existentes; la migración debe ser atómica y comprobada.
- Las reglas aceptadas específicas de liga se conservan al migrarla y deben
  diferenciarse de las reglas nuevas de eliminatoria en el dominio y contrato.
- La combinación liga-eliminatoria sigue fuera de alcance hasta decidir fuente
  de clasificación, número de plazas y reglas de empate.

## Validación

- No permanece ninguna ruta, contrato, enlace canónico ni proyección pública
  que use `leagues` como recurso actual.
- Un torneo de eliminatoria válido genera un árbol de tamaño potencia de dos,
  con *byes* reproducibles y una única final.
- Un hueco posterior solo se completa con su origen declarado; no se puede
  resolver un partido sin vencedora ni modificar una ascendencia ya disputada.
- OpenAPI, generación de cliente, migraciones, pruebas de dominio e interfaces
  verifican el contrato nuevo y sus respuestas de error seguras.

## Disparadores de revisión

- Se acepta una competición con varias fases y se conocen sus reglas de
  clasificación.
- Se necesitan doble eliminación, tercer puesto, calendarios o desempates con
  semántica deportiva más rica.

## Documentación afectada

- [Producto](../project/PRODUCT.md)
- [API](../engineering/API.md)
- [Modelo inicial de datos](../engineering/INITIAL_DATA_MODEL.md)
- [OpenAPI v1](../../contracts/openapi/v1/openapi.yaml)
- [Observabilidad](../operations/OBSERVABILITY.md)
- [Decisiones](../governance/DECISIONS.md)
- [Aprendizaje](../project/LEARNING.md)
