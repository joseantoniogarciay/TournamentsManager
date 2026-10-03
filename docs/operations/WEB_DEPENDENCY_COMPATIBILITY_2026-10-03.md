# Adaptación de dependencias web — 2026-10-03

> Registro cronológico de los parches de seguridad de v1.8.1–v1.8.3.
> La publicación posterior v1.9.0 y las revisiones activas están en
> [DEPLOYMENT.md](DEPLOYMENT.md); las referencias a versiones anteriores
> describen el momento de cada comprobación.

## Problema, evidencia y decisión

La publicación v1.8.2 conserva cuatro avisos altos y uno moderado en el grafo
JavaScript. Dos altos pertenecen a image-size 1.2.1 en Metro y el moderado a
decode-uri-component 0.2.2 en query-string 7.1.3, consumido por Expo Router.
Sus alternativas corregidas están publicadas y cumplen ADR-0138: image-size
2.0.4 desde el 14 de septiembre y decode-uri-component 0.5.0 desde el 29 de junio,
según las fechas del registro npm. El cambio no necesita una excepción de edad.

El usuario autorizó adaptar los consumidores, revisar las pruebas, publicar
primero en dev y promover a producción solo tras comprobar dev.

Se comparó actualizar toda la matriz Expo/React Native con adaptar los dos
consumidores. La matriz nueva incorpora módulos nativos y revisiones todavía
jóvenes, con pruebas y mantenimiento mayores. La adaptación mantiene las
versiones directas y requiere conservar dos parches pequeños hasta que los
consumidores admitan las nuevas APIs. Es la solución mínima para estas tres
alertas; no cambia arquitectura, contratos HTTP ni comportamiento de negocio.

## Implementación reproducible

- Override por consumidor: `metro@0.84.4>image-size` en 2.0.4. El parche de Metro
  lee los bytes con fs.promises.readFile antes de medir la imagen. Conserva el
  caso no gráfico, las escalas y la firma de getAssetData. Se actualizan juntos
  JavaScript distribuido y la fuente Flow. getAssetSize ya consume bytes.
- Override por consumidor: `query-string@7.1.3>decode-uri-component` en 0.5.0.
  El parche ajusta la importación CommonJS al export default ESM. Node 24 y la
  transformación de Metro conservan los consumidores existentes de query-string.
- pnpm patchedDependencies versiona los diffs y sus hashes en el lockfile. La
  instalación normal sigue congelada, con siete días y sin exclusiones.

Los parches no son revisiones de seguridad propias del parser: seleccionan las
correcciones publicadas y adaptan únicamente su interfaz. Al actualizar Metro o
query-string hay que comprobar si incorporan esa compatibilidad y retirar el
override y el parche conjuntamente. No se regeneran parches sobre una revisión
distinta sin revisar el diff y las regresiones.

## Verificación y promoción

`tests/dependency-compatibility.test.mjs` cubre cinco regresiones con los
consumidores que resuelve el workspace: PNG por archivo y por bytes, escalas,
assets no gráficos, rechazo de imágenes vacías/truncadas, consultas con Unicode,
reservados y valores repetidos, y entradas UTF-8/percent malformadas largas.
La suite queda en make verify mediante test-dependencies, junto con las
verificaciones previas. También se ejecutan lint, typecheck y exportación web.

La auditoría prod corregida indica dos altos, cero moderados y cero críticos.
Persisten braces 3.0.3 y node-forge 1.4.0: sus correcciones 3.0.4 y 1.4.1 no
están publicadas. No se declaran corregidos ni se inventa una versión.

Se integra primero en develop tras CI. Dev valida arranque y disponibilidad de
API, exportación con su configuración, inicio/cuenta y navegación con queries,
imágenes y ausencia de errores del bundle. Solo tras ese gate se integra en
main y prepara/activa la exportación productiva. Los releases web anteriores
se conservan para rollback; no se ejecutan migraciones por este cambio web.

## Resultado operativo

- PR #3 integró el cambio en develop tras CI `37122925939` aprobado para
  ec731c2. Dev publica `1350514f39ef5750cadba0e02fb75cc4df16ff74` desde
  2026-10-03 12:35:43 UTC. Su API estaba parada (502); se recuperaron únicamente
  PostgreSQL, API y Tempo del entorno público, conservando el volumen y la
  migración 14. Se promovió también el renderer de ese commit. El entorno local
  retirado no se arrancó y no se ejecutaron migraciones.
- Dev comprobó web/API 200, CORS 204 y sesión anónima 401 conforme al contrato
  GET /v1/sessions. En navegador cargaron inicio, cuenta con consultas Unicode,
  repetidas y malformadas y creación sin guardar. La imagen Google cargó 200×204
  y no aparecieron errores de consola. El bundle usa exclusivamente la API dev.
- PR #4 promovió develop a main después de ese gate y del CI `37123472114`
  aprobado. Su código y lockfile coinciden con los comprobados en dev.
- Producción publica v1.8.3, `91bcbcd24e6c7bb066f831045456f5a31d1bb740`,
  construido el 2026-10-03 12:43:54 UTC. La API y su renderer conservan v1.8.1.
  La web preparada conservó las entradas públicas y solo apunta a API prod.
  Se verificaron identidad de deployment.json, navegación con consultas,
  inicio/cuenta, imagen 200×204 y consola sin errores; API e imagen respondieron
  200. La entrada del bundle publicada es entry-a7cc0a16c6740c0695603651178b40cc.js.
- El rollback web productivo permanece en v1.8.2, SHA e86afeb, y el de dev en
  ea837f8. La auditoría prod conserva dos altos, cero moderados y cero críticos.
  No se declara que toda la infraestructura tenga cero avisos.

## Retrospectiva de cierre

Separar corrección y adaptación permitió retirar tres avisos sin migrar el SDK.
Las pruebas deben ejercitar consumidores reales y archivos, no solo comprobar
versiones del lockfile. Una alerta pendiente sin release no se arregla haciendo
que el auditor la ignore.

## Fuentes

- [Registro image-size](https://registry.npmjs.org/image-size).
- [Registro decode-uri-component](https://registry.npmjs.org/decode-uri-component).
- [Cambio de API de image-size 2](https://github.com/image-size/image-size/releases/tag/v2.0.0).
- [Decodificador de una pasada](https://github.com/SamVerschueren/decode-uri-component/releases/tag/v0.5.0).
