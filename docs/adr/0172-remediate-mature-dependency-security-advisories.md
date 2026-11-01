# ADR-0172: Corregir por lotes los avisos de dependencias maduras

- **Estado:** Aceptado
- **Fecha:** 2026-10-31
- **Superado parcialmente por:** ADR-0173 en la espera de node-forge, mediante excepción exacta autorizada para 1.4.2; ADR-0174 en glob/minimatch y sus transitivas exactas.
- **Decisor:** Usuario: «Ok vamos», tras la revisión de los 43 registros y la recomendación de lotes.

## Problema, contexto y evidencia

La migración de ADR-0171 deja 43 registros anteriores: 15 altos, 18 moderados
y 10 bajos, ninguno crítico. La [revisión](../engineering/DEPENDENCY_SECURITY_REVIEW_2026-10-31.md)
contrasta 39 GHSA únicos y 13 versiones candidatas. Veintiocho registros tienen
candidatos maduros en las ramas actuales; otro requiere revisar Hono 1 → 2.
Catorce necesitan espera y revisión adicional. El gate técnico está cerrado.

## Alternativas y coste

- Esperar todos los paquetes: coste inmediato bajo, mantiene correcciones maduras pendientes.
- Lotes acotados: coste medio de comprobar consumidores y resolución; limita regresiones.
- Actualización global o audit --fix: mayor coste de revisar majors y transitivas ajenas.

## Recomendación y decisión explícita

El usuario acepta los lotes acotados. Aplicar fast-uri 3.1.9/4.2.3,
fastify 5.12.6, hono 4.13.11, markdown-it 14.3.3, sharp 0.35.5,
undici 7.30.1, vite 8.1.6 y zod 4.6.6. Evaluar @hono/node-server 2.1.3
sin ampliar artificialmente los peers ni forzar consumidores incompatibles.
Si un consumidor no admite la corrección, conservar su recorrido pendiente.

Se prefieren consumidores maduros compatibles. Cuando actualizar el consumidor
amplíe innecesariamente el grafo, usar selectores de versiones afectadas para
la corrección mínima y probar su integración real. No añadir overrides globales.
Conservar ADR-0138, las excepciones exactas previas de Expo y su matriz nativa.
No instalar node-forge/glob/minimatch jóvenes ni reinterpretar altos como críticos.

## Verificación y límites

Resolver primero un candidato aislado; revisar diff y edad de toda versión nueva,
incluidos binarios opcionales. Instalar congelado, repetir auditoría autorizada,
comparar avisos por ID, ejecutar check/regresiones, generar OpenAPI dos veces,
exportar web y comprobar Scalar y Wrangler/Miniflare en loopback.
La autorización incluye estas pruebas temporales; terminarlas sin servicios
persistentes ni observabilidad. No se autoriza publicación o despliegue.

El informe de revisión registrará resultados y pendientes. No se presenta
ningún lote parcial como cierre de todos los avisos ni como seguridad de producción.

## Retrospectiva

Implementado el primer lote: auditoría 43 → 15, sin IDs nuevos; quedan
6 altos, 6 moderados y 3 bajos. Hono 2 pasa a 2.1.3, pero la rama 1 requerida
por void-server/node-ws sigue pendiente. Se verifican 56 versiones maduras,
check, 125 pruebas Node, dos generaciones idénticas y exportación web.
Las fixtures de Sharp/librsvg, Vite y Hono pasan y se cierran al terminar.
Un selector previo no alcanzaba Undici fijado a 7.29.1: revisar todo el grafo
evita dar por corregida una vulnerabilidad que conserva otra copia instalada.
