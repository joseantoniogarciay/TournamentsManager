# ADR-0007: Usar un monolito modular para el backend

- **Estado:** Aceptado
- **Fecha:** 2026-07-24
- **Decisor:** Usuario, mediante elección explícita de la alternativa A
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

El backend necesita una topología inicial que permita aprender dominio,
persistencia, seguridad, observabilidad y despliegue sin introducir complejidad
distribuida antes de conocer los límites reales del producto.

## Criterios

1. simplicidad y comprensión del sistema completo;
2. límites internos claros y verificables;
3. facilidad para probar transacciones y reglas de negocio;
4. operabilidad local y en producción;
5. coste de infraestructura y mantenimiento;
6. capacidad de evolucionar con evidencia.

## Alternativas

### Monolito modular

Un servicio Go desplegable contiene módulos internos por capacidad. Comparte
proceso y puede compartir PostgreSQL, manteniendo dependencias y propiedad de
datos explícitas.

### Microservicios

Cada capacidad se despliega de forma independiente. Añade red, consistencia
distribuida, compatibilidad de contratos, observabilidad y operación desde el
primer caso de uso.

### Funciones serverless

Casos de uso o eventos se despliegan como funciones gestionadas. Reduce la gestión
del runtime, pero fragmenta el sistema y aumenta el acoplamiento al proveedor.

## Decisión del usuario

Adoptar la alternativa A: monolito modular en Go.

## Consecuencias

- Inicialmente habrá un artefacto y una unidad de despliegue backend.
- “Monolito” no autoriza acoplamiento arbitrario entre capacidades.
- Los módulos se definirán desde el dominio cuando se reanude el diseño funcional;
  no se inventan ahora.
- Las dependencias entre módulos serán explícitas y verificables.
- Compartir PostgreSQL no implica que cualquier módulo pueda acceder a cualquier
  tabla.
- No se añade un bus interno, repositorio genérico o capa común por simetría.
- Un worker o servicio solo se extraerá con evidencia de carga, aislamiento,
  seguridad, autonomía o ciclo de despliegue.
- Framework HTTP, acceso a datos, paquetes y estructura de carpetas siguen
  pendientes.

## Validación

La implementación futura deberá producir un único backend desplegable y demostrar
mediante pruebas o análisis que las dependencias respetan los límites aceptados.

## Disparadores de revisión

- Escalado claramente diferente de una capacidad.
- Requisito de aislamiento de seguridad o disponibilidad.
- Equipos con propiedad y despliegues independientes.
- Un proceso en background necesita ciclo de vida propio.
- Los límites internos no pueden conservarse de forma razonable.

## Documentación afectada

- [ARCHITECTURE.md](../engineering/ARCHITECTURE.md)
- [TECHNICAL_BASELINE.md](../governance/TECHNICAL_BASELINE.md)
- [DEPLOYMENT.md](../operations/DEPLOYMENT.md)
