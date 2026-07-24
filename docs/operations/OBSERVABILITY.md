# Observabilidad

> Dirección objetivo: OpenTelemetry, Prometheus, Grafana, Loki y Tempo; composición
> final pendiente de evaluación en Fase 3.

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

No se desplegará toda la lista objetivo solo por completitud. La unidad mínima es
una pregunta operativa respondida de extremo a extremo.
