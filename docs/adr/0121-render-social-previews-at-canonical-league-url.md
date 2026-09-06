# ADR-0121: Renderizar previews sociales en la URL canónica de liga

- **Estado:** Aceptado
- **Fecha:** 2026-09-06
- **Decisor:** Usuario, mediante decisión explícita
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

Una liga visible por su identificador se puede compartir directamente, pero una
SPA estática no ofrece su nombre ni contexto a WhatsApp, iMessage, Slack u otros
robots que no ejecutan JavaScript. La tarjeta debe corresponder a la misma URL
canónica que abre la aplicación, sin una ruta `/share` ni un redireccionamiento.

## Contexto y restricciones

- ADR-0120 mantiene Expo en salida estática y solo indexa la home. Las ligas
  continúan siendo públicas por enlace, pero `noindex`.
- `GET /v1/leagues/{leagueId}` ya devuelve la proyección pública y responde 404
  para una liga inexistente, borrador o no visible.
- Caddy ya es el borde local de `fasttourney.com` y el único componente que
  posee el token privado para alcanzar la API de producción.
- El renderer no debe leer PostgreSQL, conocer sesiones ni interpretar agentes
  de usuario. Ha de devolver el mismo documento a navegador y crawler.
- El release estático activo se mantiene bajo el enlace atómico `current`; el
  HTML que se hidrate debe proceder de ahí para no duplicar la app.

## Criterios de decisión

1. conservar exactamente `/league/{uuid}` como URL compartida y canónica;
2. no convertir toda la aplicación Expo en SSR;
3. reutilizar solo el contrato público existente y no ampliar privilegios;
4. entregar 404 real cuando la liga no sea pública;
5. mantener `noindex` aun cuando exista una tarjeta social.

## Alternativas

### A — No renderizar una preview

- **Ventajas:** coste operativo nulo.
- **Inconvenientes:** las plataformas sociales solo ven el título genérico de
  la home o no construyen una tarjeta útil.
- **Coste y riesgo:** bajo coste, experiencia de compartir deficiente.

### B — Ruta de compartir que redirige a la liga

- **Ventajas:** aislada de la app.
- **Inconvenientes:** duplica URL, parte del historial muestra `/share`, y los
  previews pueden cachear el salto en vez de la identidad canónica.
- **Coste y riesgo:** medio; introduce enlaces y soporte duplicados.

### C — Migrar Expo completo a SSR

- **Ventajas:** metadatos dinámicos para cualquier ruta futura.
- **Inconvenientes:** Expo no mezcla esta exportación estática con SSR en un
  mismo artefacto; añade runtime, despliegue y superficie de fallo para toda la
  app antes de que exista esa necesidad.
- **Coste y riesgo:** alto y desproporcionado para un único documento dinámico.

### D — Renderer externo mínimo para la ruta canónica

- **Ventajas:** Caddy deriva solo el patrón UUID exacto a un proceso Go local;
  este obtiene la proyección pública a través del Caddy loopback ya autenticado,
  modifica el `<head>` del `index.html` activo y sirve la misma URL. El resto
  continúa siendo salida estática de Expo.
- **Inconvenientes:** añade un LaunchAgent y una comprobación operativa al
  publicar web.
- **Coste y riesgo:** bajo-medio y acotado. Un fallo del renderer solo afecta
  la preview de una liga; se devuelve 503 seguro, sin detalles internos.

## Recomendación

**Opinión/recomendación:** adoptar D. Es la mínima pieza dinámica que resuelve
el preview real sin forzar SSR general ni crear una segunda URL.

## Decisión del usuario

**Aceptada el 2026-09-06:** FastTourney usará un renderer Go externo para el
documento raíz exacto `/league/{uuid}` en producción. Caddy conserva esa URL,
la envía al renderer local y mantiene la salida estática de Expo para todo lo
demás; no hay `/share`, 303 ni detección por User-Agent.

Antes de habilitarlo en producción, desarrollo público reproduce el mismo
recorrido mediante un proceso, puerto, release y API separados. La validación
usa una liga de prueba real en `dev.fasttourney.com`; no convierte dev en una
fuente indexable.

El renderer consulta únicamente `GET /v1/leagues/{uuid}` por el loopback de
Caddy con su host API. Inyecta título, descripción, canonical, Open Graph,
Twitter Card, imagen 1200×630 y `noindex`; la liga no disponible devuelve 404.
El navegador recibe ese mismo shell y la aplicación normal se hidrata después.

## Consecuencias

### Positivas

- Una URL copiada en WhatsApp conserva el nombre de la liga y una imagen de
  marca sin revelar datos privados ni crear una URL alternativa.
- La fuente de datos sigue siendo la proyección pública de la API; no hay
  acceso a base de datos ni secretos nuevos en el renderer.
- `X-Robots-Tag` se emite desde Caddy y desde el renderer como defensa en
  profundidad. Una tarjeta social no implica indexación en Google.

### Negativas y deuda aceptada

- La disponibilidad de `/league/{uuid}` depende ahora también del LaunchAgent
  local y de la API pública. Caddy devuelve 502 si el proceso no está disponible.
- Las plataformas sociales controlan su propia caché: un cambio de nombre o
  estado puede tardar en actualizarse aunque la respuesta lleve `no-store`.
- El copy de los metadatos se publica inicialmente en inglés porque la URL no
  lleva un locale; no se inventan URLs `hreflang` inexistentes.

## Validación

1. Las pruebas del renderer verifican nombre escapado, URL canónica, metadatos,
   404 y que no captura subrutas de liga.
2. Caddy valida el patrón UUID antes del fallback estático.
3. Con una liga pública real, `curl -I` y el HTML de `/league/{uuid}` devuelven
   200, canonical idéntico, `og:title`, `og:image` y `X-Robots-Tag`.
4. Una liga inexistente devuelve 404; `/league/{uuid}/standings` sigue
   resolviendo en la aplicación estática.
5. Tras publicar, se prueba una URL real en el inspector de previews de una
   plataforma social y se registra la evidencia saneada.

## Disparadores de revisión

- Más rutas públicas requieren documentos dinámicos.
- Se requieren previews localizados, imágenes específicas por liga o volumen
  que haga necesario caché explícita.
- El Mac deja de ser el borde de producción o se adopta un runtime web SSR.

## Documentación afectada

- [ADR-0120](0120-index-public-home-without-indexing-app-routes.md)
- [Despliegue](../operations/DEPLOYMENT.md)
- [Publicación de producción](../runbooks/production-web-publication.md)
- [Decisiones](../governance/DECISIONS.md)
- [Aprendizaje](../project/LEARNING.md)
