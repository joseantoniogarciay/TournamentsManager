# ADR-0020: Usar observabilidad mínima correlacionada

- **Estado:** Aceptado
- **Fecha:** 2026-07-25
- **Decisor:** Usuario, mediante aceptación explícita de la alternativa B
- **Propietario del análisis:** Asistente como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

El futuro backend necesita permitir diagnosticar un fallo o una degradación sin
convertir la observabilidad en una plataforma independiente. Debe ser posible
relacionar una petición concreta con sus logs, sus medidas agregadas y el tiempo
empleado en sus límites de infraestructura.

## Contexto y restricciones

- El manifiesto propone OpenTelemetry, Prometheus, Grafana, Loki y Tempo como
  dirección que aún requería evaluación.
- ADR-0007 fija un monolito modular: inicialmente existe una sola unidad de
  ejecución, aunque HTTP y PostgreSQL ya son límites que importa diagnosticar.
- ADR-0011 fija PostgreSQL y ADR-0018 mantiene la API en el host y las
  dependencias locales en Compose.
- ADR-0017 prohíbe secretos en logs, errores, métricas y trazas.
- No hay todavía un vertical slice ni una API implementada. Esta decisión no
  inventa eventos de negocio, dashboards, alertas, SLO ni retenciones: se
  concretarán contra un flujo y un fallo reales.

## Criterios de decisión

1. diagnosticar latencia, errores y dependencia PostgreSQL de una petición;
2. correlacionar logs, métricas y trazas con convenciones abiertas;
3. ofrecer una experiencia local visual y reproducible;
4. evitar servicios y procesamiento que una sola aplicación no necesita aún;
5. no acoplar el dominio a SDK, backend ni proveedor de observabilidad;
6. conservar una ruta clara hacia múltiples procesos y cloud.

## Alternativas

### Alternativa A — Stack completo desde el inicio

Operar OpenTelemetry Collector, Prometheus, Grafana, Loki y Tempo antes de que
exista un flujo funcional.

- **Ventajas:** topología muy completa y Collector disponible para procesar o
  enrutar todas las señales.
- **Inconvenientes:** configura cinco servicios y políticas de retención,
  muestreo y alertas sin evidencia de uso.
- **Coste de adopción:** alto.
- **Coste de mantenimiento:** alto desde el inicio.
- **Riesgos:** aprender operaciones de la plataforma antes de aprender a
  formular una pregunta operativa útil.

### Alternativa B — Señales correlacionadas con backends locales mínimos

Emitir logs JSON a salida estándar, métricas Prometheus y trazas OpenTelemetry;
usar Prometheus, Grafana, Loki y Tempo como backends locales. La aplicación
exporta directamente a los backends y no se incorpora OpenTelemetry Collector
por ahora.

- **Ventajas:** Grafana permite consultar las tres señales y pasar de una traza
  a sus logs correlacionados; OpenTelemetry conserva un contrato portable; cada
  componente responde a una función concreta.
- **Inconvenientes:** ya se operan cuatro servicios de observabilidad y la
  aplicación tendrá configuración temporal de exportación directa.
- **Coste de adopción:** medio y diferido hasta que exista un flujo real.
- **Coste de mantenimiento:** medio; hay que versionar configuración, controlar
  cardinalidad y mantener almacenamiento local desechable.
- **Riesgos:** usar etiquetas con identificadores de alta cardinalidad o
  registrar datos sensibles. Se mitiga con revisión de atributos y prohibición
  explícita de secretos y PII.

### Alternativa C — Servicio de observabilidad gestionado

Instrumentar con OpenTelemetry y enviar las señales a un proveedor SaaS.

- **Ventajas:** menos operación de backends, acceso remoto y retención
  gestionada.
- **Inconvenientes:** cuenta, costes, credenciales, dependencia de proveedor y
  menor aprendizaje de la operación local.
- **Coste de adopción:** bajo o medio.
- **Coste de mantenimiento:** bajo para el software propio, con coste recurrente
  externo.
- **Riesgos:** adoptar límites y precios de un proveedor antes de conocer el
  volumen de señales.

### No cambiar

Mantener logs ad hoc en consola sin métricas, trazas ni visualización común.

- **Consecuencias:** no permite medir una degradación ni seguir una petición a
  través de HTTP y PostgreSQL de forma consistente.

## Comparación

La alternativa A adelanta la complejidad propia de varias aplicaciones y de un
entorno cloud. La C reduce operación, pero introduce coste y dependencia antes
de conocer las necesidades. La B conserva el estándar y ofrece una interfaz
única para las tres señales; su principal deuda es operar cuatro piezas locales,
acotadas a desarrollo y aprendizaje.

## Recomendación

**Opinión/recomendación:** alternativa B. Es la mínima suficiente para practicar
diagnóstico correlacionado y no fuerza Collector ni SaaS sin una necesidad
medida.

## Decisión del usuario

**Aceptada:** alternativa B, con estas reglas:

- la aplicación emitirá logs estructurados JSON a salida estándar mediante la
  biblioteca estándar de Go; `fmt.Println` no es la interfaz de logging de
  producción;
- logs, métricas y trazas nunca incluirán secretos, tokens, credenciales ni PII;
- las métricas usarán Prometheus; las trazas usarán OpenTelemetry y contexto de
  propagación estándar; Loki almacenará logs y Tempo trazas; Grafana será la
  interfaz de consulta;
- la instrumentación automática cubrirá límites técnicos establecidos —entrada
  HTTP, salida HTTP cuando exista y PostgreSQL— usando convenciones semánticas
  de OpenTelemetry;
- los autores del código solo añadirán spans manuales para operaciones con
  significado operativo que la instrumentación automática no vea. No se crea un
  span por función ni se introducen nombres de negocio sin un caso de uso
  aceptado;
- los logs incluirán contexto de traza cuando esté disponible para permitir la
  navegación cruzada en Grafana;
- OpenTelemetry Collector, perfiles, alertas, SLO, backend gestionado, retención
  de producción y muestreo avanzado quedan aplazados hasta que una pregunta
  operativa o varios procesos los justifiquen.

## Consecuencias

### Positivas

- Un diagnóstico puede comenzar en una métrica, traza o log y mantener el
  contexto de la petición.
- OpenTelemetry evita que la lógica de negocio dependa de un backend concreto.
- Loki aporta búsqueda visual de logs, no solo salida de terminal.
- Prometheus, Grafana, Loki y Tempo pueden ejecutarse localmente sin licencia de
  servicio; su operación y almacenamiento siguen teniendo coste.

### Negativas y deuda aceptada

- Se mantienen cuatro servicios y sus datos locales.
- La exportación directa desde la aplicación se sustituirá por Collector si la
  topología, filtrado, muestreo o destinos múltiples lo requieren.
- Hasta disponer de un flujo crítico no hay paneles, alertas ni objetivos de
  servicio significativos.

## Validación

Con el primer flujo técnico o vertical slice se demostrará que:

- una petición HTTP visible en Tempo contiene sus spans HTTP y PostgreSQL;
- Grafana permite llegar desde esa traza a los logs JSON con el mismo contexto;
- Prometheus muestra volumen, errores y latencia HTTP, además de una señal del
  pool PostgreSQL cuando esté disponible;
- una indisponibilidad provocada de PostgreSQL puede diagnosticarse desde los
  tres tipos de señales y queda documentada en un runbook;
- las pruebas verifican que los atributos registrados no contengan secretos ni
  PII, y que la instrumentación no invada el dominio;
- sin backend configurado, el servicio conserva un modo local seguro y no falla
  por no poder exportar telemetría.

## Disparadores de revisión

- hay más de una aplicación o destino de telemetría y la exportación directa se
  duplica;
- se necesita filtrado, redacción, tail sampling o control de costes antes de
  almacenar señales;
- la retención, volumen o disponibilidad hacen insuficiente el stack local;
- un incidente muestra que las señales o atributos no permiten diagnosticarlo;
- se adopta Kubernetes o un backend cloud que requiera Collector u operación
  gestionada.

## Documentación afectada

- [OBSERVABILITY.md](../operations/OBSERVABILITY.md)
- [TECHNICAL_BASELINE.md](../governance/TECHNICAL_BASELINE.md)
- [DECISIONS.md](../governance/DECISIONS.md)
- [DECISIONS_TO_REVISIT.md](../governance/DECISIONS_TO_REVISIT.md)
- [LEARNING.md](../project/LEARNING.md)
- [CHANGELOG.md](../../CHANGELOG.md)

## Fuentes técnicas

- [OpenTelemetry: conceptos](https://opentelemetry.io/docs/what-is-opentelemetry/)
- [OpenTelemetry: Collector](https://opentelemetry.io/docs/collector/)
- [OpenTelemetry: Go](https://opentelemetry.io/docs/languages/go/)
- [Prometheus: introducción](https://prometheus.io/docs/introduction/overview/)
- [Grafana Loki: introducción](https://grafana.com/docs/loki/latest/get-started/overview/)
- [Grafana Tempo: trazas](https://grafana.com/docs/tempo/latest/set-up-for-tracing/)
