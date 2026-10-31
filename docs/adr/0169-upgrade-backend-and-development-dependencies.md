# ADR-0169: Actualizar dependencias backend y herramientas de desarrollo

- **Estado:** Aceptado
- **Fecha:** 2026-10-31
- **Decisor:** Usuario mediante autorización explícita en la conversación
- **Complementa:** ADR-0012, ADR-0014, ADR-0046 y ADR-0138.

## Problema, alternativas y coste

La revisión identifica parches de seguridad del compilador Go y mejoras de
robustez de persistencia, instrumentación y herramientas. Mantener versiones
reduce cambios inmediatos, pero conserva defectos corregidos. Actualizar toda
dependencia al último release amplía riesgo y puede incumplir maduración.
Se recomienda actualizar los candidatos revisados por bloques, conservando
TypeScript y PostHog y la matriz de Expo ya validada. El coste es bajo para
formato, medio para generadores, persistencia, instrumentación y gestor pnpm.

## Decisión explícita

El usuario decide: «Vamos con todo el listado y dejamos para otro momento
typescript y posthog». Se aceptan Go 1.26.10, pgx 5.11.0, OpenTelemetry Go
1.47.0 con otelhttp 0.72.0, Prettier 3.9.10, ESLint 10.13.0,
typescript-eslint 8.72.0, Orval 8.44.0 y Redocly CLI 2.71.0. Se incluye pnpm
12.15.0 del listado, manteniendo Node 24.18.0 y TypeScript 6.0.3.

Cada versión nueva, también transitiva, debe cumplir siete días salvo la
excepción exacta de x/net documentada a continuación. Se conservan las excepciones exactas
de la migración Expo autorizada en ADR-0167. Una incompatibilidad requiere
adaptación mínima o registrar el bloqueo; no rebajar controles silenciosamente.

## Implementación y verificación

Conservar los cambios previos del workspace y revisar el diff por bloque.
Alinear compilador en ambos módulos y Dockerfile, resolver grafos Go por
versiones exactas y verificar fechas oficiales de los módulos cambiados.
Mantener lockfile congelado y controles de maduración al cambiar pnpm.
Regenerar el cliente OpenAPI, revisar cambios reales y comprobar determinismo.
Ejecutar formato, lint, tipos, regresiones, build, vulnerabilidades e integración
PostgreSQL aislada cuando sea viable. La actualización no autoriza despliegue
de producción ni activación de observabilidad. Cerrar cualquier servicio de
pruebas temporal y documentar resultados y retrospectiva.

## Fuentes

- [Go: releases](https://go.dev/doc/devel/release#go1.26.10).
- [pgx 5.11](https://github.com/jackc/pgx/releases/tag/v5.11.0).
- [OpenTelemetry Go 1.47](https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.47.0).
- [pnpm 12.15](https://github.com/pnpm/pnpm/releases/tag/v12.15.0).
- [Orval 8.44](https://github.com/orval-labs/orval/releases/tag/v8.44.0).
- [Registro npm](https://registry.npmjs.org/).
- [Registro de módulos Go](https://proxy.golang.org/).

## Excepción adicional autorizada: x/net 0.62.0

El gate `govulncheck` identifica GO-2026-6860 y GO-2026-6866 en las llamadas
`html.Parse` y `html.Render` de la preview social. La corrección x/net 0.62.0
se publicó el 27 de octubre a las 19:00:16 UTC; madura el 3 de noviembre a las
20:00:16 de Madrid. Los avisos no se presentan como vulnerabilidades críticas.
Conservar 0.60.0 mantendría el gate de seguridad fallido. Adelantar únicamente
la corrección tiene coste acotado de pruebas del parser y de preview.

El usuario decide explícitamente «Autorizar excepción puntual para x/net 0.62.0».
Se supera ADR-0138 solo para esa versión exacta y sus requisitos indispensables;
no se cambia la política general. Se actualizan ambos módulos, se revisan sus
transitivas y se repiten pruebas y análisis. No se publica ni despliega.

- [GO-2026-6860](https://pkg.go.dev/vuln/GO-2026-6860).
- [GO-2026-6866](https://pkg.go.dev/vuln/GO-2026-6866).
- [Fecha exacta de x/net](https://proxy.golang.org/golang.org/x/net/@v/v0.62.0.info).

## Estado de ejecución y retrospectiva

Implementación terminada. Instalación congelada, políticas de suministro,
regeneración OpenAPI determinista, tipos, formato, lint, exportación web,
regresiones Node/Python, pruebas Go con carreras, build y 62 integraciones
PostgreSQL correctas. Govulncheck final sin vulnerabilidades alcanzables;
el aviso OpenPGP del módulo x/crypto no corresponde a un paquete importado.
No se ejecuta la comparación de rendimiento opt-in ni se publican artefactos.
Contenedor efímero eliminado y servicios local/dev/observabilidad apagados.

Retrospectiva: validar la salida del generador y el parser de parches es parte
del cambio de herramientas. La corrección de seguridad surgida durante el gate
necesitó una decisión explícita adicional; esa excepción no modifica la política
general. Véase el [informe](../engineering/DEPENDENCY_UPGRADE_2026-10-31.md).
