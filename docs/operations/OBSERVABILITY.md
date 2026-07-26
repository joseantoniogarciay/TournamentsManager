# Observabilidad

> Base aceptada: OpenTelemetry, Prometheus, Grafana, Loki y Tempo. El
> OpenTelemetry Collector queda aplazado hasta que una necesidad medida lo
> justifique. Véase [ADR-0020](../adr/0020-use-minimal-correlated-observability.md).

## Resultado buscado

La observabilidad debe permitir responder:

- ¿está funcionando el servicio para el usuario?
- ¿qué cambió?
- ¿dónde está el cuello de botella o fallo?
- ¿qué usuarios, operaciones o dependencias están afectados?
- ¿qué acción reduce el impacto?

## Señales

- **Logs:** eventos estructurados y accionables, sin secretos.
- **Métricas:** comportamiento agregado, capacidad y objetivos de servicio.
- **Trazas:** recorrido y latencia entre límites.
- **Perfiles/eventos:** solo cuando respondan una pregunta concreta.

Las señales deben compartir contexto de correlación y convenciones de nombres.

## Base mínima aceptada

- **Logs:** JSON a salida estándar mediante `log/slog`; Loki los almacena y
  Grafana permite buscarlos. Un log es un evento discreto, no una traza ni una
  sustitución de `fmt.Println` en producción.
- **Métricas:** Prometheus recopila medidas agregadas; Grafana las visualiza.
- **Trazas:** OpenTelemetry instrumenta límites técnicos y Tempo conserva el
  recorrido de una operación. Una traza se compone de *spans* —por ejemplo,
  HTTP entrante y consulta PostgreSQL—, no solo de llamadas de red.
- **Correlación:** cada log incluirá el identificador de traza y span cuando el
  contexto exista. No se registran secretos, tokens, credenciales ni PII.

La instrumentación automática cubre HTTP y PostgreSQL. Quien implementa el
código decide los spans manuales solo cuando representen una operación
operativamente significativa que no esté cubierta; no se añade un span por
función. Los nombres y atributos técnicos siguen las convenciones semánticas de
OpenTelemetry. Los eventos de negocio se decidirán junto al caso de uso que los
necesite.

El servicio debe degradarse de forma segura si un backend de telemetría no está
configurado o no está disponible. El dominio no importa SDKs ni tipos de los
backends.

## Orden de diseño

1. definir el flujo crítico y su resultado correcto;
2. definir indicadores y objetivos;
3. enumerar modos de fallo;
4. elegir señales y atributos necesarios;
5. decidir instrumentación y backend;
6. crear visualización, alerta y runbook;
7. provocar un fallo y verificar el diagnóstico.

## Criterios para evaluar el stack

- estándares abiertos y portabilidad;
- integración con Go y Kubernetes;
- correlación entre señales;
- coste de operación y almacenamiento;
- retención y cardinalidad;
- experiencia local;
- seguridad de datos;
- facilidad de backup, upgrade y diagnóstico.

No se crearán paneles, alertas, SLO, retenciones de producción ni perfiles por
completitud. La unidad mínima es una pregunta operativa respondida de extremo a
extremo y validada provocando un fallo.
