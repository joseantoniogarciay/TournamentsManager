# Corrección de glob/minimatch y opciones Hono — 2026-11-01

> Decisión posterior: el usuario acepta la opción 1 mediante ADR-0175.
> [Void-server actualizado](HONO_SECURITY_FIX_2026-11-01.md) retira Hono 1 y
> la auditoría actual del checkout queda en cero avisos. La comparación
> siguiente conserva las alternativas previas a esa decisión.

## Resultado del bloque autorizado

El usuario autoriza corregir glob y minimatch y pide opciones para Hono.
[ADR-0174](../adr/0174-correct-glob-and-minimatch-with-scoped-compatibility.md)
registra la decisión antes de implementar. La resolución real mantiene Hono.

**Auditoría npm: 3 → 1 registro moderado**, GHSA-rmxm-3fg6-px4f en
@hono/node-server 1.19.17. Cero altos, bajos o críticos y ningún ID nuevo.
Desaparecen los registros de glob y minimatch, incluidas ambas ramas de este.
El resultado corresponde al checkout, no a una publicación de producción.

Se instalan glob 14.0.3 y minimatch 11.0.3, con path-scurry 3.0.0 y
@isaacs/nonrecursive-glob-match 1.0.3. Son las únicas cuatro versiones nuevas.
Se conservan Expo, Orval, TypeDoc, Scalar CLI y los módulos nativos fijados.
La revisión incluye los hotfixes posteriores a las correcciones iniciales,
sin actualizar el conjunto de herramientas a latest.

## Compatibilidad: qué se tuvo que comprobar

Spectral 1.22.0 y 1.24.0 importan minimatch 3 como función default. Dos parches
exactos cambian esa importación a minimatch.minimatch, sin modificar opciones
ni reglas. Se prueban ambas versiones con matchBase, globstar, llaves y archivos
ocultos. Retirar los parches cuando sus consumidores admitan la API nueva.

La primera prueba también detectó que glob 7.2.3 usa propiedades internas del
matcher antiguo: actualizar solo minimatch rompe su recorrido. El único
consumidor de glob 7, plugin-commonjs 22.0.2, utiliza únicamente glob.sync.
Se actualiza ese enlace concreto a glob 14.0.3, que conserva sync. No se crea
un wrapper de callbacks ni se modifica la API de minimatch o del plugin.
La regresión ejecuta Rollup con dynamicRequireTargets y comprueba el valor
exportado al ejecutar el CommonJS generado.

TypeDoc 0.28.20 usa matchOne, retirada en minimatch 11. Un tercer parche
usa match(path, true) para la coincidencia parcial de directorios, conservando
la negación. Una regresión convierte entradas TypeScript descubiertas mediante
glob anidado y comprueba que un archivo excluido no aparece en las reflexiones.
Se mantiene el compilador TS 6 privado de TypeDoc.

Las otras pruebas comparan la búsqueda síncrona y por promesa, el rechazo de
una búsqueda cancelada y filtros positivos/negados de Expo Fingerprint.
No se afirma haber demostrado todas las condiciones posibles de explotación.

## Excepciones de publicación autorizadas

La decisión de instalar ahora estos dos paquetes incluye sus cuatro versiones
necesarias. Se mantiene la espera global de siete días y se añaden únicamente
estas exclusiones; retirarlas en las fechas UTC indicadas, sin revertir fixes.

| Versión exacta | Publicación UTC | Retirar exclusión UTC |
| --- | --- | --- |
| glob 14.0.3 | 2026-10-28 04:08:39.624 | 2026-11-04 04:08:39.624 |
| minimatch 11.0.3 | 2026-10-28 20:00:34.713 | 2026-11-04 20:00:34.713 |
| path-scurry 3.0.0 | 2026-10-27 21:14:48.842 | 2026-11-03 21:14:48.842 |
| @isaacs/nonrecursive-glob-match 1.0.3 | 2026-10-28 19:50:48.349 | 2026-11-04 19:50:48.349 |

## Hono: alternativas concretas, todavía sin decisión

**Hechos del grafo actual:** la única rama 1.x pertenece a
@scalar/void-server 2.5.2. node-ws 1.3.1 mantiene un peer ^1.19.11,
pero actualmente se enlaza a la rama 2 y muestra un aviso de peer previo.
Son dos cuestiones distintas: resolver el paquete vulnerable de void-server
no corrige automáticamente la compatibilidad declarada de node-ws.

### 1. Actualizar únicamente void-server — recomendación

Usar el selector exacto @scalar/void-server@2.5.2 → 2.5.19. Su manifiesto
oficial declara @hono/node-server ^2.1.5. Se ha resuelto el candidato en un
directorio temporal, sin instalarlo ni copiarlo al proyecto: desaparecen
void-server 2.5.2 y node-server 1.19.17; se añaden únicamente void-server
2.5.19, node-server 2.1.5 y hono 4.13.13. Las tres versiones cumplen siete días
(publicadas respectivamente el 23, 13 y 4 de octubre). No se necesita ampliar
la excepción de edad ni actualizar todo Scalar.

**Coste:** bajo/medio: revisar API de integración y probar HTTP, WebSocket y
comandos Scalar antes de aplicar. La resolución prueba viabilidad del grafo;
no prueba todavía el comportamiento en ejecución ni una auditoría nueva de
ese candidato. Mantiene el peer previo de node-ws como asunto independiente.
Retirar el selector cuando Scalar adopte el consumidor corregido.

### 2. Backport del parche a Hono 1.19.17

Portar la corrección serveStatic a la versión actual y añadir una regresión
con rutas codificadas y middleware protegido. Evita actualizar el consumidor,
pero mantiene un parche de seguridad propio. El auditor por versión puede
seguir señalando 1.19.17: se necesitaría demostrar el backport y documentar
la discrepancia, sin silenciar el aviso globalmente.

**Coste:** medio/alto de mantenimiento; no se recomienda existiendo una versión
madura del consumidor que admite Hono 2.

### 3. Esperar a una actualización del consumidor sin override

Conservar el aviso y revisar periódicamente las dependencias de Scalar. Trabajo
inmediato bajo, pero duración incierta. El fallo exige servir archivos bajo
prefijos protegidos con serveStatic; la mera instalación no acredita ese uso.
No se ha probado exhaustivamente que ningún modo de Scalar pueda activarlo.
No iniciar servidores con contenido privado y rutas estáticas protegidas
mientras se decide esta alternativa.

**Coste:** bajo de código, con seguimiento y riesgo pendiente. El usuario no ha
aceptado ese riesgo ni elegido esta alternativa; Hono permanece sin cambios
por su instrucción de comparar opciones primero.

Actualizar Scalar CLI completo también es posible, pero amplía el grafo y sus
herramientas; aporta más coste que la primera alternativa para este problema.
No se propone como solución mínima.

## Verificación y límites

- Candidato aislado y diff revisado: cuatro versiones nuevas, Hono sin cambios.
- Instalación congelada correcta; todos los selectores y parches son exactos.
- Cuatro regresiones nuevas cubren Spectral, Rollup/glob, Expo Fingerprint y TypeDoc.
- `make verify` completo correcto: formato, lint, tipos, regresiones,
  exportación web, generación comprobada OpenAPI/sqlc, build y govulncheck.
- 130/130 regresiones Node correctas, sin omisiones.
- Integración PostgreSQL omitida localmente por ausencia de base aislada;
  no se presenta como ejecutada en esta sesión.
- La nueva auditoría reutiliza la autorización explícita de enviar el grafo a npm.
  El primer intento se detuvo por cuota de revisión automática; el reintento
  autorizado sí se ejecutó y devolvió el único moderado de Hono.

La evidencia compacta está en
[result.json](evidence/dependency-security-2026-11-01/result.json). Fuentes,
logs y el candidato de Hono quedan en `.tmp/glob-minimatch-security-20261031/`.
No se inicia local/dev persistente ni observabilidad, no se publica ni despliega.
Las fixtures cierran sus recursos temporales. La checklist de cliente se revisa
como mantenimiento transitivo: no cambian UI, endpoints, locales ni nativos;
se exige tipos y exportación web.

## Retrospectiva

Un conteo reducido de auditoría no descubre incompatibilidades de ejecución.
Las pruebas de consumidores detectaron el cambio de exportación de minimatch,
la dependencia de glob 7 de sus detalles internos y el uso de matchOne en TypeDoc. Migrar el
único consumidor de esa rama resulta más simple que mantener un adaptador.
Para Hono, mirar el consumidor exacto ofrece una alternativa acotada frente a
forzar una major o actualizar todas las herramientas.

## Fuentes

- [Aviso glob](https://github.com/advisories/GHSA-3g3m-2fjr-6pw4).
- [Aviso minimatch](https://github.com/advisories/GHSA-g3hv-75wg-m249).
- [Versiones y fechas glob](https://registry.npmjs.org/glob).
- [Versiones y fechas minimatch](https://registry.npmjs.org/minimatch).
- [path-scurry](https://registry.npmjs.org/path-scurry).
- [Matcher no recursivo](https://registry.npmjs.org/@isaacs%2fnonrecursive-glob-match).
- [Manifiesto void-server 2.5.19](https://registry.npmjs.org/@scalar%2fvoid-server/2.5.19).
- [Manifiesto node-ws 1.3.1](https://registry.npmjs.org/@hono%2fnode-ws/1.3.1).
- [Aviso Hono](https://github.com/advisories/GHSA-rmxm-3fg6-px4f).
