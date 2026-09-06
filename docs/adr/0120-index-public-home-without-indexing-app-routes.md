# ADR-0120: Indexar la home pública sin indexar las rutas de aplicación

- **Estado:** Aceptado
- **Fecha:** 2026-09-06
- **Decisor:** Usuario, mediante decisión explícita
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** ADR-0016, en la estrategia de salida web inicial
- **Superado por:** Ninguno

## Problema

FastTourney necesita que las personas puedan descubrir el producto desde su
propia home en Google y en experiencias de búsqueda generativa, sin convertir
las rutas autenticadas ni las ligas leíbles por ID en resultados de búsqueda.

## Contexto y restricciones

- La ruta `/` ya explica el producto a una persona sin sesión y permite crear
  un borrador local; una cuenta solo es necesaria para persistir o publicar.
- Las ligas visibles se pueden leer sin sesión mediante su ID público, pero no
  se ha decidido su descubrimiento mediante búsqueda (ADR-0049).
- `www.fasttourney.com` redirige permanentemente al host canónico
  `https://fasttourney.com` (ADR-0089).
- La exportación web actual es SPA (`single`) y no incluye HTML indexable de la
  home. La salida estática de Expo genera HTML en build sin requerir un servidor.
- Este cambio no introduce SSR, un segundo sitio, un CMS ni una dependencia
  nueva. Las rutas dinámicas de liga conservan el fallback de la aplicación.
- El entorno `dev` conserva `X-Robots-Tag: noindex, nofollow, noarchive`.

## Criterios de decisión

1. hacer descubrible el producto con el menor cambio posible;
2. no exponer ligas, cuentas, recuperación ni flujos de creación en buscadores;
3. conservar una única home y la experiencia de borrador local existente;
4. evitar infraestructura o contenido duplicado antes de que exista evidencia;
5. poder verificar el resultado desde el artefacto exportado y Search Console.

## Alternativas

### Alternativa A — Mantener toda la SPA fuera de la estrategia de búsqueda

- **Ventajas:** no añade metadatos ni controles de rastreo.
- **Inconvenientes:** la marca y su propuesta de valor no tienen una superficie
  pública declarada para indexación.
- **Coste de adopción y mantenimiento:** nulo.
- **Riesgos:** confundir una app accesible por URL con un producto descubrible.

### Alternativa B — Indexar la home existente y excluir las rutas de aplicación

- **Ventajas:** reutiliza `/`, su explicación para invitado y su borrador local;
  genera HTML estático de `/` y añade URL canónica, `robots.txt`, sitemap y
  metadatos sin otro runtime.
- **Inconvenientes:** la exportación también genera HTML de rutas internas, que
  deben permanecer excluidas del índice; la cobertura pública queda
  deliberadamente limitada a una sola URL.
- **Coste de adopción y mantenimiento:** bajo; al añadir otra página pública se
  decide explícitamente si entra en el sitemap.
- **Riesgos:** confiar solo en `robots.txt`; se mitiga con `X-Robots-Tag` para
  toda ruta que no sea la home. La salida estática es necesaria: Expo no trata
  el HTML de su salida SPA como indexable.

### Alternativa C — Crear una landing independiente o migrar a SSR

- **Ventajas:** control máximo de HTML inicial, campañas y contenido público.
- **Inconvenientes:** duplica navegación, copy, pruebas y despliegue, o añade
  un runtime servidor que el producto aún no necesita.
- **Coste de adopción y mantenimiento:** medio o alto.
- **Riesgos:** sobreingeniería y divergencia respecto a la home real.

### No cambiar

Conservar el rendering client-side sin una home indexable mantiene el alcance
privado original, pero no satisface el objetivo de descubrimiento confirmado.

## Comparación

La A conserva el menor coste pero no aporta descubrimiento. La C resuelve
necesidades futuras que todavía no existen. La B satisface el objetivo mediante
la home real, preserva los límites de privacidad de las rutas de aplicación y
mantiene la evolución a static rendering o SSR como una decisión posterior.

## Recomendación

**Opinión/recomendación:** adoptar B. `robots.txt` comunica preferencias de
rastreo, pero el borde debe emitir además `X-Robots-Tag` para impedir indexar
las rutas no públicas aunque otro sitio las enlace.

## Decisión del usuario

**Aceptada el 2026-09-06:** `https://fasttourney.com/` es la única superficie
web indexable inicial de FastTourney. La propia home explica el producto y ofrece
crear un torneo mediante el borrador local ya aceptado. Por ajuste explícito del
usuario el mismo día, la caja principal recupera «Crear torneo» y retira las
acciones añadidas de explorar, registro y acceso; estos últimos siguen en Cuenta.

Las rutas de cuenta, creación, recuperación, biblioteca, enlaces y ligas no se
indexan. El release web publica `robots.txt`, un sitemap con solo la home y sus
metadatos canónicos y sociales mediante exportación estática. No se añade
`llms.txt`, marcado especial para IA, SSR ni una landing independiente.

## Consecuencias

### Positivas

- La propuesta de valor se descubre desde la misma ruta que usa el producto.
- Google y otros rastreadores reciben una frontera explícita entre la home y la
  aplicación.
- El sitemap no revela IDs de ligas ni convierte su lectura pública en catálogo.

### Negativas y deuda aceptada

- La salida estática genera también HTML para rutas internas conocidas. El borde
  debe conservar su directiva `noindex` y el sitemap no debe listarlas.
- El HTML y los metadatos iniciales se publican en inglés bajo la URL canónica
  única; el cliente adapta su idioma al navegador. No se declaran `hreflang` sin
  URLs localizadas equivalentes.

## Validación

1. La exportación web incluye `/robots.txt` y `/sitemap.xml` con la home como
   única URL.
2. La home exportada contiene título, descripción, canonical y Open Graph.
3. En producción `/` no devuelve `X-Robots-Tag: noindex`; una ruta de aplicación
   sí lo devuelve.
4. `www` redirige a la URL canónica y Search Console verifica el dominio,
   inspecciona `/` y recibe el sitemap.
5. Una persona sin sesión puede empezar un torneo desde la caja principal y
   acceder al registro o inicio de sesión desde Cuenta; las apps nativas
   mantienen la misma intención funcional.

## Disparadores de revisión

- Search Console no procesa correctamente el HTML o metadatos de `/`.
- Se necesita otra página pública, previews específicos o contenido frecuente.
- Se decide hacer una liga, ranking, club o perfil descubrible.
- Se requieren URLs localizadas independientes o campañas con contenido propio.
- Una liga compartida necesita metadatos sociales dinámicos (resuelto para el
  documento raíz por ADR-0121, sin ampliar la indexación).

## Documentación afectada

- [ADR-0016](0016-use-client-side-web-rendering-initially.md)
- [Producto](../project/PRODUCT.md)
- [Arquitectura](../engineering/ARCHITECTURE.md)
- [Desarrollo](../engineering/DEVELOPMENT.md)
- [Despliegue](../operations/DEPLOYMENT.md)
- [Decisiones](../governance/DECISIONS.md)
- [Decisiones a revisar](../governance/DECISIONS_TO_REVISIT.md)
- [Aprendizaje](../project/LEARNING.md)
