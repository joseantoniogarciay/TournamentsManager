# ADR-0175: Actualizar void-server para retirar Hono 1

- **Estado:** Aceptado
- **Fecha:** 2026-11-01
- **Decisor:** Usuario: «Vamos con la 1 pero explicame exactamente para que se utiliza actualmente».
- **Complementa:** ADR-0172 y ADR-0174; no modifica la política de edad.

## Problema y contexto

El último aviso npm pertenece a @hono/node-server 1.19.17, declarado por
@scalar/void-server 2.5.2. La comparación previa encuentra una actualización
acotada de este consumidor que permite retirar la rama vulnerable sin
actualizar toda la CLI ni mantener un parche propio de seguridad.

Void-server es el servidor de eco de `scalar document void`. El proyecto
instala Scalar para `scalar document serve` mediante openapi:ui; no contiene
scripts que llamen a document void. El módulo se importa al cargar la CLI,
pero su servidor se crea solo al ejecutar el comando void. No se usa en la
API Go, las apps o el visor público que carga Scalar API Reference desde CDN.

## Alternativas, coste y decisión

- Actualizar solo void-server: coste bajo/medio de comprobar integración; tres
  versiones maduras nuevas y ningún parche de seguridad propio.
- Backport a Hono 1: mantenimiento medio/alto y aviso por versión posiblemente
  persistente, aunque se demuestre el parche.
- Esperar: conserva el aviso y exige seguimiento, sin fecha garantizada.

El usuario acepta la primera opción. Aplicar únicamente el selector
@scalar/void-server@2.5.2 → 2.5.19. La resolución prevista añade void-server
2.5.19, @hono/node-server 2.1.5 y hono 4.13.13, todos con siete días.
Se actualiza el destino del override de seguridad preexistente de hono a
4.13.13: el valor 4.13.11 anterior reescribía peers y causaba una discrepancia
nueva con el Hono 4.13.13 requerido por void-server. Se consolida esa revisión
madura sin falsear los rangos oficiales. No se añaden excepciones de edad ni
se amplía el peer antiguo de node-ws.
Retirar el selector cuando Scalar incorpore una versión corregida por su rango.

## Verificación y límites

Candidato aislado, diff exacto de paquetes y maduración; instalación congelada,
audit con la autorización vigente de enviar el grafo a npm. Probar espejo HTTP,
estados, WebSocket y comandos Scalar, además de make verify y exportación web.
Las pruebas usan datos sintéticos y servidores temporales que se cierran.
No enviar credenciales reales al servidor de eco ni dejar servicios persistentes.
Se sube a develop conforme a la instrucción vigente; no se publica ni despliega,
no se activa observabilidad y no se altera producto o contrato HTTP.

El peer previo de node-ws (^1.19.11, enlazado con 2.1.3) se conserva visible:
no es el recorrido de dependencia Hono 1 que corrige este cambio.

## Resultado y retrospectiva

Implementado: solo se añaden las tres versiones maduras previstas. Se retiran
void-server 2.5.2, node-server 1.19.17 y la revisión hono 4.13.11 ya innecesaria.
Audit termina con exit 0 y cero avisos. Pasan make verify, 131 regresiones y
ambos comandos Scalar con --once; la prueba permanente cubre HTTP y WebSocket.
Permanecen los dos peers anteriores, sin ninguno nuevo.

La retrospectiva es distinguir la utilidad opcional de eco del visor OpenAPI
realmente configurado, y el framework hono 4 de su adaptador node-server 2.
El [informe](../engineering/HONO_SECURITY_FIX_2026-11-01.md) conserva la
explicación de uso, las versiones y los límites del cierre.
