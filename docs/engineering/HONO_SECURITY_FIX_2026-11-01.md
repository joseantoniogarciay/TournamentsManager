# Corrección Hono y uso actual de Scalar — 2026-11-01

## Decisión y resultado

El usuario acepta actualizar únicamente void-server y pide explicar su uso.
[ADR-0175](../adr/0175-update-scalar-void-server-to-remove-hono-one.md) registra
la decisión antes de implementar. Se añade el selector exacto
@scalar/void-server@2.5.2 → 2.5.19 y se actualiza el destino del override
preexistente de hono a 4.13.13.

**Auditoría npm: 1 → 0 avisos**, en todas las severidades. Desaparece
GHSA-rmxm-3fg6-px4f y no aparecen IDs nuevos. Se retira la rama
@hono/node-server 1.19.17. Este resultado pertenece al checkout; no se ha
publicado ni desplegado un artefacto ni se afirma cubrir otros scanners.

## Para qué se utiliza actualmente

| Pieza | Uso comprobado en el proyecto |
| --- | --- |
| @scalar/cli 2.0.1 | DevDependency directa. package.json define openapi:ui como scalar document serve contracts/openapi/v1/openapi.yaml --watch --port 8083; make openapi-ui ejecuta ese script. |
| document serve | Carga el contrato local y sirve una página Scalar API Reference. Con watch, informa al navegador de cambios del archivo mediante un canal de eventos. Es la utilidad que el repositorio sí configura. |
| @scalar/void-server | Dependencia de la CLI. La CLI la importa al cargar y llama createVoidServer solo en la acción document void. No hay scripts del repositorio que invoquen esa acción. |
| document void | Servidor de eco para depurar clientes HTTP: devuelve método, ruta, cabeceras, query, cookies y cuerpo. También permite simular códigos de estado y formatos de respuesta. Se inicia bajo petición; no arranca por instalar Scalar o ejecutar document serve. |
| hono | Framework de enrutado HTTP de esas herramientas. La revisión final es 4.13.13. No corresponde a las versiones 1.x/2.x del adaptador. |
| @hono/node-server | Adaptador que permite atender en Node una aplicación Hono basada en fetch. La CLI conserva 2.1.3 y void-server declara 2.1.5; ambas ramas son 2.x. |

El paquete void-server también exporta attachVoidWebSocket para eco de frames.
La CLI 2.0.1 solo llama createVoidServer en document void; no conecta allí ese
helper WebSocket. La prueba del helper comprueba su compatibilidad como API
del paquete, sin atribuir su uso al proyecto o al comando actual.

No hay importación de void-server/Hono en código de negocio, app o Worker.
La API propia es Go. El visor público de desarrollo de infra/home/api-docs
carga @scalar/api-reference 1.63.0 desde CDN; es una integración de navegador
diferente de esta CLI y de su servidor de eco.

La inspección corresponde al código y scripts actuales; no demuestra si alguien
ha ejecutado manualmente document void en otra ocasión. Por eso se distingue
presencia instalada de un servicio usado o arrancado por el proyecto.

## Cambio mínimo y maduración

La resolución inicial aceptada añade void-server 2.5.19, node-server 2.1.5 y
hono 4.13.13. Las tres versiones nuevas son maduras, publicadas el 23, 13 y
4 de octubre respectivamente. No se añaden excepciones de edad.

El override previo hono@4.12.32 → 4.13.11 reescribía también peers y causaba
una discrepancia nueva con el Hono 4.13.13 de void-server. Se actualiza su
destino a 4.13.13, una versión ya requerida por el cambio aceptado. Así se
retira también Hono 4.13.11 y se conserva una sola revisión del framework,
sin ampliar artificialmente los rangos oficiales.

El diff final añade tres versiones y retira tres: void-server 2.5.2,
node-server 1.19.17 y hono 4.13.11. No cambia la CLI, Expo, módulos nativos,
contrato OpenAPI, negocio ni configuración de observabilidad. Retirar el
selector de void-server cuando Scalar incorpore un consumidor corregido.

## Verificaciones y límites

- Candidato aislado, comparación del grafo y tres fechas oficiales verificadas.
- Instalación congelada correcta. El primer intento offline carecía de metadatos
  de una dependencia ya fijada; se repite con acceso al registro sin cambiar
  el lockfile para completar la política de suministro.
- Regresión permanente HTTP sobre loopback: espejo POST/JSON, cabecera sintética,
  query repetida y simulación del estado 418.
- Regresión WebSocket: texto, binario y cierre 1000 por timeout. Se cierran cliente,
  WebSocketServer y servidor HTTP en el hook de limpieza.
- CLI real: document void y document serve responden correctamente en puertos
  efímeros y terminan con --once. Se usa únicamente un payload sintético y el
  contrato público local; no se inicia el watch persistente.
- make verify completo correcto y 131/131 regresiones Node, sin omisiones.
  Incluye tipos, exportación web, generación comprobada OpenAPI/sqlc, build y
  govulncheck. PostgreSQL real se omite localmente por ausencia de base aislada.
- Audit autorizado con envío del grafo a npm: exit 0 y cero avisos.
- Peers: permanecen solo los dos anteriores de node-ws/node-server y
  neverpanic/TypeScript. No hay peer nuevo ni se falsea ninguno para ocultarlo.

Los logs y fuentes de ejecución quedan en `.tmp/hono-security-fix-20261101/`.
La [evidencia versionada](evidence/dependency-security-2026-11-01/hono-result.json)
registra hashes, versiones, fechas y comparación de avisos. La checklist de
cliente se aplica como mantenimiento de tooling: no hay cambios de pantallas,
idiomas, navegación, endpoints o módulos nativos; tipos y exportación web pasan.
No se arrancan local/dev persistentes ni observabilidad. Subir a develop no
promociona main ni publica o despliega producción.

## Retrospectiva

Leer el comando concreto evita atribuir al producto un servidor de eco que solo
viene incluido en la CLI. También evita confundir hono 4 con su adaptador Node 2.
Actualizar el consumidor retira la rama vulnerable sin un backport propio;
revisar los overrides evita introducir un peer artificialmente desalineado.
La auditoría a cero se conserva como evidencia de este grafo y fecha, junto a
las pruebas del comportamiento real y los peers que siguen pendientes.

## Evidencia de uso y fuentes

- package.json: openapi:ui y @scalar/cli en devDependencies.
- mk/typescript.mk: make openapi-ui.
- Scalar CLI 2.0.1 instalado: ServeCommand, VoidCommand y bootServer2 en index.js.
- Void-server 2.5.19 instalado: create-void-server.js, get-request-data.js,
  get-body.js y void-websocket.js.
- infra/home/api-docs/index.html: Scalar API Reference de navegador.
- [Manifiesto void-server 2.5.19](https://registry.npmjs.org/@scalar%2fvoid-server/2.5.19).
- [Manifiesto node-server 2.1.5](https://registry.npmjs.org/@hono%2fnode-server/2.1.5).
- [Aviso corregido](https://github.com/advisories/GHSA-rmxm-3fg6-px4f).
