# Revisión del inventario de dependencias — 2026-10-31

## Evaluación inicial y alcance

Se revisan los 43 registros del inventario de la migración TS 7: 15 altos,
18 moderados y 10 bajos; cero críticos. Son 12 paquetes y 39 GHSA únicos:
los cuatro avisos de fast-uri aparecen separados para las ramas 3 y 4.
Se contrastan los 39 avisos públicos de GitHub y las fechas oficiales de npm
para 13 versiones candidatas. No se vuelve a enviar el grafo al auditor.

**Estado antes de implementar:** los 43 registros estaban abiertos.
Hay candidatos maduros para 29 registros (9 altos, 13 moderados y 7 bajos).
Los otros 14 (6 altos, 5 moderados y 3 bajos) necesitan esperar y revisar
compatibilidad o confirmar la corrección. Tener candidato maduro no acredita
compatibilidad ni una auditoría futura limpia.

La revisión inicial conserva el trabajo previo y prepara la decisión del usuario.
La ejecución posterior autorizada se registra a continuación.

## Ejecución autorizada y resultado

El usuario acepta «Ok vamos». [ADR-0172](../adr/0172-remediate-mature-dependency-security-advisories.md)
registra la decisión antes de implementar. Se resuelve primero un candidato
con manifiestos, parches y lockfile separados; se incorporan después los
selectores exactos y el lockfile verificado al checkout.

**Auditoría final: 15 avisos, 6 altos, 6 moderados y 3 bajos; cero críticos.**
Desaparecen 28 IDs del inventario original y no aparecen IDs nuevos. La
[evidencia final](evidence/dependency-security-2026-10-31/result.json) conserva
comparación, hashes, pendientes y las 56 fechas de publicación verificadas.
Todas las versiones nuevas son maduras, incluidos binarios Sharp/Rolldown.

Se aplican los nueve paquetes del primer lote y @hono/node-server 2.1.3 para
los consumidores que antes fijaban 2.0.1. La revisión del grafo descubre que
el selector previo undici@7.28.0 no alcanza a quienes fijan 7.29.1: se añade
un segundo selector exacto. El grafo final no contiene Undici 7.29.1.
No cambia ninguna dependencia directa ni la matriz Expo nativa.

Actualizar Scalar completo a la versión madura 2.13.1 ampliaría el grafo,
introduciría Undici 8 y mantendría dependencias que necesitan corrección.
Miniflare fija Sharp exactamente. Se eligen correcciones acotadas en lugar de
actualizaciones amplias de consumidores. Los overrides previos se actualizan
y se añaden solo sharp@0.35.4, vite@8.0.16, zod@4.6.5 y undici@7.29.1.
Retirar estos selectores cuando sus consumidores incorporen correcciones.

### Pendientes efectivos

| Paquete | Registros | Motivo y siguiente paso |
| --- | --- | --- |
| node-forge 1.4.0 | 12: 6 altos, 3 moderados, 3 bajos | Candidato 1.4.1 joven hasta el 3 de noviembre; confirmar además GHSA-86w9 antes de cerrar. |
| glob 13.0.6 | 1 moderado | Espera hasta el 3 de noviembre y compatibilidad de major 14. |
| minimatch 3.1.5 / 10.2.6 | 1 moderado | Espera hasta el 3 de noviembre y compatibilidad de major 11. |
| @hono/node-server 1.19.17 | 1 moderado | @scalar/void-server 2.5.2 requiere ^1.19.10 y @hono/node-ws 1.3.1 exige peer ^1.19.11. La última versión madura consultada de node-ws conserva ese peer; no se fuerza 2.x ni se falsea su rango. |

El peer previo de node-ws sigue avisando de la raíz Hono 2; su recorrido 1.x
permanece en el grafo. El peer previo neverpanic/TypeScript 5 tampoco se oculta.
El segundo lote queda parcialmente implementado, con ese aviso abierto;
los catorce registros sujetos a espera no se instalan anticipadamente.

### Verificaciones

- Instalaciones congeladas correctas; 56/56 versiones nuevas con siete días.
- Check completo: Prettier, Oxlint, tipos y contrato OpenAPI correctos.
- 125/125 regresiones Node correctas, incluida la integración Wrangler/Miniflare.
- Dos generaciones idénticas de los 145 archivos OpenAPI de partida.
- Exportación web correcta en `/tmp/tm-security-web-20261031`.
- Sharp 0.35.5 con librsvg 2.63.2 transforma SVG en PNG y conserva dimensiones;
  Vite de Scalar transforma HTML y Hono 2.1.3 responde por HTTP loopback.
  Los servidores de fixtures se cierran en finally y Miniflare en sus hooks.
- Gate completo `make verify` correcto sobre los tres commits integrados:
  incluye generación comprobada de OpenAPI/sqlc y seguridad operativa.
- Govulncheck: cero vulnerabilidades alcanzables y cero en paquetes importados;
  conserva un aviso a nivel de módulo requerido sin llamadas desde el código.
- Backend acumulado: formato, golangci-lint, pruebas Go, build y ambos tidy -diff correctos.
  No se repite integración PostgreSQL real en esta sesión; la evidencia anterior
  permanece en DEPENDENCY_UPGRADE_2026-10-31.md.

La revisión automática rechazó inicialmente repetir audit por enviar el grafo
al servicio público npm. El usuario autoriza explícitamente «Sí, autorizo enviar
el grafo a npm»; la ejecución autorizada termina con exit 1 por los 15 avisos,
no por error de conectividad. Logs y respuestas completas quedan en
`.tmp/dependency-security-fix-20261031/`. No se publica, despliega ni activa
observabilidad o servicios local/dev persistentes.

La checklist del cliente se aplica al alcance: resolución transitiva de tooling,
sin cambios nuevos de UI, idiomas, endpoints, adaptadores o módulos nativos.
Tipos y exportación web pasan; no se atribuyen builds nativas nuevas a este lote.
El trabajo acumulado conserva sus verificaciones y límites documentados por
ADR-0166 a ADR-0171. El usuario pide subirlo a develop; main y producción
conservan ADR-0119.

## Contexto y evidencia

El manifiesto DOCX y su transcripción exigen alternativas y decisión explícita
antes de un cambio importante. El gate técnico está cerrado. ADR-0138 conserva
siete días para cada versión nueva, también transitiva; una severidad alta no
activa la excepción crítica. ADR-0171 cierra el tooling y deja este inventario
separado. Las excepciones exactas de Expo de ADR-0167 no autorizan otras.

La [evidencia versionada](evidence/dependency-security-2026-10-31/review.json)
conserva hash del audit original, los 43 IDs, versiones instaladas, rangos npm,
rangos GitHub y metadatos de candidatos. Las respuestas completas consultadas
quedan en `.tmp/dependency-security-review-20261031/`; el audit original en
`.tmp/typescript7-migration-20261031/audit.json`. Las rutas proceden de ese audit;
los requisitos de consumidores se contrastan con sus package.json instalados.
El inventario es del checkout, no del artefacto actualmente publicado.

## Candidatos y coste de compatibilidad

Las fechas se expresan en UTC. El enlace de cada paquete apunta a sus metadatos
oficiales npm. Se elige la versión mínima indicada para cubrir el conjunto de
avisos de su rama, sin equipararla automáticamente a una versión aprobada.

| Paquete | Candidato | Publicación UTC | Maduración / condición |
| --- | --- | --- | --- |
| [@hono/node-server](https://registry.npmjs.org/@hono%2fnode-server) | 2.1.3 | 2026-09-29 06:10:34 | Madura; validar consumidor y transitivas; 1.x → 2.x requiere decisión |
| [fast-uri](https://registry.npmjs.org/fast-uri) | 3.1.9 | 2026-10-19 23:05:26 | Madura; validar consumidor y transitivas |
| [fast-uri](https://registry.npmjs.org/fast-uri) | 4.2.3 | 2026-10-19 22:51:36 | Madura; validar consumidor y transitivas |
| [fastify](https://registry.npmjs.org/fastify) | 5.12.6 | 2026-10-13 06:50:58 | Madura; validar consumidor y transitivas |
| [glob](https://registry.npmjs.org/glob) | 14.0.2 | 2026-10-27 22:07:00 | Esperar hasta 2026-11-03 22:07:00 UTC; cambio de major |
| [hono](https://registry.npmjs.org/hono) | 4.13.11 | 2026-09-29 06:15:49 | Madura; validar consumidor y transitivas |
| [markdown-it](https://registry.npmjs.org/markdown-it) | 14.3.3 | 2026-10-16 01:47:54 | Madura; validar consumidor y transitivas |
| [minimatch](https://registry.npmjs.org/minimatch) | 11.0.1 | 2026-10-27 22:26:03 | Esperar hasta 2026-11-03 22:26:03 UTC; cambio de major |
| [node-forge](https://registry.npmjs.org/node-forge) | 1.4.1 | 2026-10-27 04:44:07 | Esperar hasta 2026-11-03 04:44:07 UTC; GHSA-86w9 no confirma parche |
| [sharp](https://registry.npmjs.org/sharp) | 0.35.5 | 2026-09-27 13:46:24 | Madura; validar consumidor y transitivas |
| [undici](https://registry.npmjs.org/undici) | 7.30.1 | 2026-10-15 14:18:31 | Madura; validar consumidor y transitivas |
| [vite](https://registry.npmjs.org/vite) | 8.1.6 | 2026-10-06 04:51:51 | Madura; validar consumidor y transitivas |
| [zod](https://registry.npmjs.org/zod) | 4.6.6 | 2026-10-11 16:18:23 | Madura; validar consumidor y transitivas |

No se han resuelto ni comprobado todas las transitivas de estos candidatos.
Sharp distribuye binarios: revisar también opcionales por plataforma y librsvg.
Node 24.18.0 satisface los engines declarados por los candidatos consultados;
eso no prueba compatibilidad de sus APIs.

## Superficie y condiciones de explotación

**Hecho local:** las rutas del audit llevan a Scalar CLI, Orval/TypeDoc,
Wrangler/Miniflare y herramientas Expo. La imagen runtime Go se construye en
scratch sin node_modules (`apps/backend/Dockerfile`); el Worker de la web usa
APIs de plataforma sin importar esos paquetes (`infra/cloudflare/web-worker.mjs`).
La búsqueda en `apps/client/src` no encuentra importaciones de los 12 paquetes.
**Inferencia limitada:** la superficie identificada está en herramientas;
no se demuestra ausencia de código vulnerable en todos los bundles ni se audita
el CDN de Scalar usado por la documentación pública. El flag dev=false del
auditor para Expo no demuestra ejecución de node-forge en la app distribuida.

| Paquete / entrada local | Condición relevante y evaluación |
| --- | --- |
| node-forge: Expo CLI y code-signing-certificates | PEM, firmas RSA, certificados, ASN.1, PKCS#12, TLS y opciones criptográficas según cada GHSA. Los consumidores importan Forge para firma/certificados; no cabe descartar todo el grupo como código muerto. Los avisos TLS requieren ese cliente TLS; no se ha demostrado su uso por Expo. Revisar material de firma no confiable y rutas concretas antes de una sesión de build. |
| sharp: Wrangler → Miniflare | SVG procesado por librsvg vulnerable; el aviso alto incluye posible RCE bajo condiciones de Linux/glibc. No se rebaja por usar macOS ni se extrapola al Worker publicado. Miniflare fija sharp 0.35.4 exactamente: cambiarlo requiere actualizar consumidor o justificar override y probar binarios. |
| undici: Scalar, Orval/json-magic, Miniflare | Respuestas HTTP/TLS hostiles; algunos avisos requieren EventSource, WebSocket, SRI o interceptores específicos. El bypass de TLS exige callback en tls; la confusión de caché exige Agent sin origen, caché compartida y orden redirect/cache. No equivale a todo fetch del producto: el cliente usa transporte del navegador/nativo y el backend es Go. Revisar entradas remotas del tooling. |
| markdown-it: Orval → TypeDoc opcional | Markdown hostil: smartquotes requiere typographer; HTML inline requiere html=true; el otro DoS afecta referencias multilínea con opciones por defecto. No se considera seguro solo por ser opcional: aparece instalado. Revisar generación de documentación con contenido externo. |
| vite: Scalar CLI | Tres avisos afectan apertura en editor y sus condiciones de editor/plataforma; otro necesita servidor expuesto y una ruta importada que coincida con un archivo del host; el último requiere HTML transformado con URL no saneada en middleware personalizado. No se arranca Scalar para probar estos supuestos. Scalar fija Vite 8.0.16 exactamente; probar su integración con 8.1.6. |
| hono y @hono/node-server: Scalar | Doble decodificación de serveStatic: bypass de middleware en prefijos estáticos protegidos. Servir solo archivos públicos no cumple esa condición. No equivale a traversal fuera del root. Verificar configuración del servidor Scalar si se utiliza. |
| fastify: Scalar CLI | Restricciones de nombres de cabecera mediante propertyNames enum/const con mayúsculas. El servidor de API Go no usa este validador. Validar comportamiento de Scalar tras actualizar. |
| fast-uri: Scalar/Fastify/Ajv y otros consumidores del grafo | Normalización y resolución pueden discrepar sobre esquema, host y rutas relativas. El aviso de sufijos exige reference=suffix. La explotación depende de usar esa representación para permisos, enlaces o destinos; no se ha acreditado ese flujo en el producto. Mantener las ramas 3 y 4 evita un salto innecesario de API. |
| zod: Scalar/types | Colecciones hostiles sobredimensionadas con restricciones: consumo de memoria al validarlas. Revisar documentos y datos que consume el tooling; no se encuentra importación directa del producto. |
| glob: Expo y Scalar/static | Una cancelación puede terminar la promesa mientras sigue el recorrido de disco; relevante si se confía en AbortSignal para acotar árboles no confiables. Los consumidores declaran ^13; 14.0.2 queda fuera de rango. |
| minimatch: Expo, TypeDoc, Scalar/Spectral | Amplificación de patrones no confiables; el ejemplo del aviso requiere optimizationLevel >= 2 en 10.2.6. El rango oficial también marca 3.1.5: se conserva el registro, sin asumir idéntico recorrido de código en esa rama. TypeDoc y fingerprint requieren 10.x; no forzar 11 sobre las ramas 3/10 sin migración. |

## Discrepancias que deben conservarse

- npm indica node-forge >=1.4.1 para GHSA-86w9-cpqp-85rv; la respuesta de GitHub
  mantiene <=1.4.0 afectado, pero first_patched_version es null y no confirma
  una corrección. No declarar ese aviso cerrado hasta contrastar el parche
  upstream y repetir la auditoría del candidato.
- @hono/node-server tiene dos versiones afectadas, 1.19.17 y 2.0.10, dentro de
  un solo registro. 2.1.3 es madura, pero resolver solo la rama 2 no elimina
  el aviso de la rama 1. La corrección cruza major para esa rama.
- Los 43 registros son 39 GHSA: deduplicar para explicar causas no autoriza
  eliminar las dos ramas afectadas de fast-uri del seguimiento.

## Alternativas comparadas y plan aceptado

1. Mantener todo y esperar: coste inmediato bajo, deja también 29 registros
   con candidatos maduros. No se recomienda como cierre del bloque.
2. Corregir por lotes acotados: coste medio de pruebas y trazabilidad; mantiene
   la matriz Expo y concentra los cambios en consumidores y ramas afectadas.
   **Recomendación:** esta opción.
3. audit --fix, latest u overrides globales: reduce el trabajo inicial de
   edición, pero aumenta el coste de compatibilidad y puede saltar majors,
   políticas de edad o dependencias fijadas. No se recomienda.

Plan aceptado para la opción 2 mediante ADR-0172; resultados y límites arriba:

- Primer lote: fast-uri 3.1.9/4.2.3, fastify 5.12.6, hono 4.13.11,
  markdown-it 14.3.3, sharp 0.35.5, undici 7.30.1, vite 8.1.6 y
  zod 4.6.6. Cubren 28 registros con candidatos maduros en las ramas actuales;
  priorizar sharp, undici y herramientas que procesen entradas externas.
  Comprobar si un consumidor maduro actualiza su dependencia; si no,
  justificar selectores exactos, sin añadir overrides globales por defecto.
- Segundo lote: resolver ambas ramas de @hono/node-server hacia una integración
  compatible con 2.1.3 o posterior madura. Revisar @hono/node-ws y sus peers;
  no ampliar sus rangos para ocultar incompatibilidad. Es el registro maduro
  adicional, hasta 29, condicionado a compatibilidad.
- Tercer lote: node-forge después del 3 de noviembre a las 04:44:08 UTC,
  confirmando primero el aviso sin parche acreditado; glob después de las
  22:07:01 UTC y minimatch después de las 22:26:04 UTC del mismo día.
  Esas fechas corresponden solo a las versiones candidatas de la tabla;
  no autorizan instalar versiones posteriores más jóvenes. Preferir soporte
  de consumidores para majors 14/11 frente a forzarlos en Expo/TypeDoc.

## Criterios de verificación de los lotes

- Resolver en un candidato aislado, conservar siete días y verificar todas las
  versiones nuevas, incluidos binarios y opcionales. Revisar diff y peers,
  instalar con lockfile congelado y repetir audit completo, comparando IDs y
  versiones con esta evidencia. Un descenso de conteo no demuestra por sí solo
  corrección de todos los recorridos.
- Ejecutar check, regresiones de dependencias y tooling; generar OpenAPI dos
  veces sin cambios inesperados y exportar web. Probar Scalar y la integración
  Wrangler/Miniflare afectada bajo una sesión de pruebas autorizada.
- Si cambia matriz o módulo nativo, aplicar la checklist de cliente y build
  nativa correspondiente. La mera presencia de Expo CLI no implica cambio
  nativo. No activar observabilidad; cerrar servicios de prueba al terminar.
- Registrar resultado por aviso y riesgo residual. No publicar ni desplegar
  como consecuencia implícita de corregir el lockfile.

## Retrospectiva y límites de cierre

Se completa el primer lote, eliminando 28 registros; quedan 15 abiertos.
No se afirma haber probado la explotabilidad de cada herramienta. La revisión de rangos,
fechas, consumidores y condiciones permite separar espera de maduración,
cambio de major y pruebas pendientes. El estándar práctico es corregir con
trazabilidad y validar el consumidor; la elección de lotes anterior es una
decisión aceptada en ADR-0172. Comprobar todas las copias instaladas fue
necesario para detectar Undici fijado a dos versiones distintas.

## Inventario trazable

Cada fila corresponde a un ID npm del audit original; el enlace lleva al aviso
público. A/M/B significan alto/moderado/bajo. Se conserva el rango de corrección
que informó npm, sujeto a los matices anteriores.

| ID npm | Paquete / versiones locales | GHSA | Nivel | Corrección según npm |
| --- | --- | --- | --- | --- |
| 1243661 | @hono/node-server 1.19.17, 2.0.10 | [GHSA-rmxm-3fg6-px4f](https://github.com/advisories/GHSA-rmxm-3fg6-px4f) | M | >=2.1.3 |
| 1243097 | fast-uri 4.1.5 | [GHSA-w74x-v6f9-j3xg](https://github.com/advisories/GHSA-w74x-v6f9-j3xg) | M | >=4.2.3 |
| 1243098 | fast-uri 3.1.8 | [GHSA-w74x-v6f9-j3xg](https://github.com/advisories/GHSA-w74x-v6f9-j3xg) | M | >=3.1.9 |
| 1243100 | fast-uri 4.1.5 | [GHSA-xhq7-xmqj-769h](https://github.com/advisories/GHSA-xhq7-xmqj-769h) | B | >=4.2.3 |
| 1243101 | fast-uri 3.1.8 | [GHSA-xhq7-xmqj-769h](https://github.com/advisories/GHSA-xhq7-xmqj-769h) | B | >=3.1.9 |
| 1243103 | fast-uri 4.1.5 | [GHSA-wjr9-j795-jxcr](https://github.com/advisories/GHSA-wjr9-j795-jxcr) | M | >=4.2.3 |
| 1243104 | fast-uri 3.1.8 | [GHSA-wjr9-j795-jxcr](https://github.com/advisories/GHSA-wjr9-j795-jxcr) | M | >=3.1.9 |
| 1243106 | fast-uri 4.1.5 | [GHSA-333x-w4r6-6crp](https://github.com/advisories/GHSA-333x-w4r6-6crp) | B | >=4.2.3 |
| 1243107 | fast-uri 3.1.8 | [GHSA-333x-w4r6-6crp](https://github.com/advisories/GHSA-333x-w4r6-6crp) | B | >=3.1.9 |
| 1242860 | fastify 5.12.5 | [GHSA-389g-gv7q-r4x8](https://github.com/advisories/GHSA-389g-gv7q-r4x8) | B | >=5.12.6 |
| 1244013 | glob 13.0.6 | [GHSA-3g3m-2fjr-6pw4](https://github.com/advisories/GHSA-3g3m-2fjr-6pw4) | M | >=14.0.2 |
| 1242595 | hono 4.13.7 | [GHSA-5r4p-p66f-jhc7](https://github.com/advisories/GHSA-5r4p-p66f-jhc7) | M | >=4.13.11 |
| 1242600 | markdown-it 14.3.1 | [GHSA-r7fv-28h4-cvq7](https://github.com/advisories/GHSA-r7fv-28h4-cvq7) | A | >=14.3.2 |
| 1242646 | markdown-it 14.3.1 | [GHSA-vfgh-jcmw-g37q](https://github.com/advisories/GHSA-vfgh-jcmw-g37q) | B | >=14.3.3 |
| 1242672 | markdown-it 14.3.1 | [GHSA-gfr2-rv44-r23c](https://github.com/advisories/GHSA-gfr2-rv44-r23c) | A | >=14.3.3 |
| 1244012 | minimatch 10.2.6, 3.1.5 | [GHSA-g3hv-75wg-m249](https://github.com/advisories/GHSA-g3hv-75wg-m249) | M | >=11.0.1 |
| 1240912 | node-forge 1.4.0 | [GHSA-86w9-cpqp-85rv](https://github.com/advisories/GHSA-86w9-cpqp-85rv) | A | >=1.4.1 |
| 1244014 | node-forge 1.4.0 | [GHSA-4642-5mcj-2rqw](https://github.com/advisories/GHSA-4642-5mcj-2rqw) | B | >=1.4.1 |
| 1244015 | node-forge 1.4.0 | [GHSA-6wvx-3r32-xmcg](https://github.com/advisories/GHSA-6wvx-3r32-xmcg) | M | >=1.4.1 |
| 1244016 | node-forge 1.4.0 | [GHSA-525v-9gp9-w22h](https://github.com/advisories/GHSA-525v-9gp9-w22h) | M | >=1.4.1 |
| 1244017 | node-forge 1.4.0 | [GHSA-7wch-p9px-956x](https://github.com/advisories/GHSA-7wch-p9px-956x) | A | >=1.4.1 |
| 1244018 | node-forge 1.4.0 | [GHSA-h6cr-wjv9-3cpp](https://github.com/advisories/GHSA-h6cr-wjv9-3cpp) | B | >=1.4.1 |
| 1244019 | node-forge 1.4.0 | [GHSA-4487-j8r4-4g2f](https://github.com/advisories/GHSA-4487-j8r4-4g2f) | B | >=1.4.1 |
| 1244020 | node-forge 1.4.0 | [GHSA-5cwg-6vw7-rj6p](https://github.com/advisories/GHSA-5cwg-6vw7-rj6p) | M | >=1.4.1 |
| 1244021 | node-forge 1.4.0 | [GHSA-jc66-7q78-782g](https://github.com/advisories/GHSA-jc66-7q78-782g) | A | >=1.4.1 |
| 1244022 | node-forge 1.4.0 | [GHSA-8hrf-8g6f-7xwp](https://github.com/advisories/GHSA-8hrf-8g6f-7xwp) | A | >=1.4.1 |
| 1244023 | node-forge 1.4.0 | [GHSA-hqq3-7jhf-c5jx](https://github.com/advisories/GHSA-hqq3-7jhf-c5jx) | A | >=1.4.1 |
| 1244024 | node-forge 1.4.0 | [GHSA-5x38-2vff-x55p](https://github.com/advisories/GHSA-5x38-2vff-x55p) | A | >=1.4.1 |
| 1241331 | sharp 0.35.4 | [GHSA-wq5f-xc86-pv6w](https://github.com/advisories/GHSA-wq5f-xc86-pv6w) | A | >=0.35.5 |
| 1243078 | undici 7.29.1 | [GHSA-4qv8-2p9j-2685](https://github.com/advisories/GHSA-4qv8-2p9j-2685) | M | >=7.30.1 |
| 1243081 | undici 7.29.1 | [GHSA-f72v-6w45-q472](https://github.com/advisories/GHSA-f72v-6w45-q472) | M | >=7.30.1 |
| 1243113 | undici 7.29.1 | [GHSA-xjj6-f33q-929c](https://github.com/advisories/GHSA-xjj6-f33q-929c) | A | >=7.30.1 |
| 1243473 | undici 7.29.1 | [GHSA-4w7w-5xrc-9h26](https://github.com/advisories/GHSA-4w7w-5xrc-9h26) | A | >=7.30.1 |
| 1243475 | undici 7.29.1 | [GHSA-2m5v-pch6-74hx](https://github.com/advisories/GHSA-2m5v-pch6-74hx) | A | >=7.30.1 |
| 1243478 | undici 7.29.1 | [GHSA-qrxm-fqhp-jwvm](https://github.com/advisories/GHSA-qrxm-fqhp-jwvm) | A | >=7.30.1 |
| 1243481 | undici 7.29.1 | [GHSA-9fm3-97px-rjf6](https://github.com/advisories/GHSA-9fm3-97px-rjf6) | M | >=7.30.1 |
| 1243484 | undici 7.29.1 | [GHSA-55wc-3pg8-jfw6](https://github.com/advisories/GHSA-55wc-3pg8-jfw6) | M | >=7.30.1 |
| 1242624 | vite 8.0.16 | [GHSA-698v-4r3x-3vcv](https://github.com/advisories/GHSA-698v-4r3x-3vcv) | M | >=8.1.6 |
| 1242653 | vite 8.0.16 | [GHSA-r825-w8mm-5pxw](https://github.com/advisories/GHSA-r825-w8mm-5pxw) | M | >=8.1.6 |
| 1242660 | vite 8.0.16 | [GHSA-587h-7wmj-h5pm](https://github.com/advisories/GHSA-587h-7wmj-h5pm) | A | >=8.1.6 |
| 1242726 | vite 8.0.16 | [GHSA-9jrq-w75r-8gcw](https://github.com/advisories/GHSA-9jrq-w75r-8gcw) | B | >=8.1.6 |
| 1242732 | vite 8.0.16 | [GHSA-rq7h-c2jc-7f22](https://github.com/advisories/GHSA-rq7h-c2jc-7f22) | M | >=8.1.6 |
| 1243066 | zod 4.6.5 | [GHSA-mrhx-c7xc-m3w5](https://github.com/advisories/GHSA-mrhx-c7xc-m3w5) | A | >=4.6.6 |
