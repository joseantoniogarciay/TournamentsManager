# Adaptación de dependencias web — 2026-10-03

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

## Retrospectiva

Separar corrección y adaptación permitió retirar tres avisos sin migrar el SDK.
Las pruebas deben ejercitar consumidores reales y archivos, no solo comprobar
versiones del lockfile. Una alerta pendiente sin release no se arregla haciendo
que el auditor la ignore.

## Fuentes

- [Registro image-size](https://registry.npmjs.org/image-size).
- [Registro decode-uri-component](https://registry.npmjs.org/decode-uri-component).
- [Cambio de API de image-size 2](https://github.com/image-size/image-size/releases/tag/v2.0.0).
- [Decodificador de una pasada](https://github.com/SamVerschueren/decode-uri-component/releases/tag/v0.5.0).
