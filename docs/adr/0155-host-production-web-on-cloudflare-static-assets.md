# ADR-0155: Alojar la web de producción en Cloudflare Static Assets

- **Estado:** Aceptado
- **Fecha:** 2026-10-10
- **Decisor:** Usuario, mediante «Vamos con ello» tras recomendar la migración
- **Supera parcialmente:** ADR-0090 y ADR-0121 para el origen web de producción

## Problema y contexto

Cloudflare Tunnel transporta actualmente la web al Caddy del Mac, que sirve el
export estático de Expo. Una interrupción doméstica impide incluso cargar la
interfaz. Cloudflare permite alojar esos archivos sin coste por almacenamiento
ni peticiones estáticas. API, PostgreSQL y los datos permanecen en K3s doméstico.

## Alternativas y mantenimiento

- Mantener Caddy: sin migración, pero conserva la dependencia doméstica de todos
  los archivos y la operación de releases locales.
- Pages con función para previews: alojamiento estático gratuito; requiere
  también adaptar el documento dinámico y operar su configuración.
- Workers Static Assets con función mínima: archivos gratuitos y un único
  despliegue de archivos y renderer. Añade Wrangler y cuota de ejecución para
  las peticiones dinámicas, sin migrar Expo a SSR ni la API a otro runtime.

## Recomendación y decisión

Se adopta Workers Static Assets. El usuario autoriza migrar la web conservando
la API doméstica y el paso explícito de publicación. No se contrata Workers Paid
ni servicios adicionales. Los estáticos se atienden antes de la función; solo
las rutas que necesitan fallback o metadatos la invocan. Workers Free tiene
cuota diaria de ejecución compartida por cuenta; agotarla puede impedir esas
rutas dinámicas, aunque los archivos estáticos continúen disponibles.

Se conserva `fasttourney.com`, `www` con redirección canónica, cabeceras de
seguridad, indexación exclusiva de la home, asociaciones móviles cuando existan
y `/tournament/{uuid}` como documento con metadatos y 404 real. El renderer de
borde consulta únicamente `GET /v1/tournaments/{uuid}` en la API pública,
sin cookies, sesión, token privado de Caddy ni acceso a PostgreSQL. Esta es una
operación de infraestructura fuera del cliente OpenAPI de la aplicación.
El HTML proviene de los mismos assets publicados, evitando desajustes de bundle.

Desarrollo público conserva Caddy y el renderer Go, apagados fuera de pruebas
según ADR-0146. El origen doméstico productivo se conserva para rollback durante
la migración; no se cambia el túnel ni DNS de la API. Servir la interfaz desde
Cloudflare no hace funcionar operaciones de API con el Mac apagado.

## Publicación y validación

Wrangler se fija a una versión madura en el lockfile. Preparar un artefacto no
publica producción. Se prueban routing de Expo, fallback de rutas profundas,
cabeceras, metadatos escapados, 404, fallos de API seguros, métodos, `.well-known`
y ausencia de source maps públicos. La conmutación usa el release productivo
identificado por SHA; no publica inadvertidamente cambios de `develop`.
Las siguientes promociones mantienen ADR-0119 y una publicación explícita.

No afirmar que la migración está activa hasta verificar las respuestas HTTPS
del dominio y un despliegue Cloudflare trazable. Documentar estado efectivo,
recuperación y evidencia en el runbook.

## Fuentes

- [Static Assets y facturación](https://developers.cloudflare.com/workers/static-assets/billing-and-limitations/)
- [Routing de la función](https://developers.cloudflare.com/workers/static-assets/routing/worker-script/)
- [Cabeceras de assets](https://developers.cloudflare.com/workers/static-assets/headers/)
- [Wrangler](https://developers.cloudflare.com/workers/wrangler/commands/workers/)
