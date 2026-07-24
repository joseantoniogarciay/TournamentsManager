# ADR-0016: Usar rendering web client-side inicialmente

- **Estado:** Aceptado
- **Fecha:** 2026-07-24
- **Decisor:** Usuario, mediante confirmación explícita
- **Propietario del análisis:** Asistente como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

El cliente universal aceptado en ADR-0008 y ADR-0015 debe ejecutarse en web,
iOS y Android. Falta decidir cómo renderizar la web y cómo aislar diferencias de
plataforma sin romper la paridad funcional del producto.

## Contexto y restricciones

- El cliente será universal con React Native, Expo y Expo Router.
- El producto inicial está pensado para torneos privados entre amistades, clubes
  o grupos cerrados.
- No existe requisito inicial de descubrimiento público, SEO, páginas
  indexables ni landing pública rica.
- La visibilidad exacta de torneos queda pendiente de decisión funcional.
- Web, iOS y Android deben ofrecer el mismo producto, pero la presentación puede
  adaptarse a dispositivo, entrada, accesibilidad y APIs disponibles.
- No se ha creado todavía `apps/client`.

## Criterios de decisión

1. simplicidad operativa inicial;
2. paridad funcional entre web, iOS y Android;
3. bajo coste para un equipo pequeño;
4. capacidad de evolucionar si aparecen torneos o páginas públicas;
5. aislamiento limpio de diferencias web/native;
6. ausencia de infraestructura web innecesaria.

## Alternativas

### Alternativa A — Universal adaptativo con salida web estática

Usar el cliente universal y generar HTML estático para rutas públicas cuando
aporte SEO, previews sociales o carga inicial.

- **Ventajas:** mantiene Expo universal y mejora páginas públicas; despliegue en
  hosting estático; buena base para contenido indexable conocido en build.
- **Inconvenientes:** exige modelar qué rutas se generan; las rutas dinámicas
  públicas pueden necesitar `generateStaticParams` o estrategia adicional.
- **Coste de adopción:** medio si todavía no hay contenido público real.
- **Coste de mantenimiento:** bajo o medio según crezca el contenido público.

### Alternativa B — Universal con rendering web client-side

La web se comporta inicialmente como una aplicación renderizada en el navegador.
Expo Router conserva rutas universales y navegación web, pero no se optimiza el
HTML inicial para SEO público.

- **Ventajas:** menor complejidad; no requiere servidor web ni generación
  estática; encaja con un producto privado tras login; conserva paridad y rutas
  universales.
- **Inconvenientes:** peor para indexación, previews sociales y contenido público
  rico; una futura superficie pública puede requerir revisar la decisión.
- **Coste de adopción:** bajo.
- **Coste de mantenimiento:** bajo mientras el producto siga siendo privado o de
  acceso limitado.

### Alternativa C — Universal con server-side rendering

Renderizar la web en servidor en cada request.

- **Ventajas:** resuelve mejor rutas dinámicas públicas, metadatos por recurso y
  contenido fresco.
- **Inconvenientes:** requiere runtime servidor, caching y operación adicional;
  en Expo Router el SSR actual es una capacidad reciente/alpha según la
  documentación oficial.
- **Coste de adopción:** medio o alto.
- **Coste de mantenimiento:** medio o alto.

### Alternativa D — Cliente universal y web pública especializada

Mantener Expo para la aplicación principal y crear una superficie web específica
para páginas públicas si la calidad web lo exige.

- **Ventajas:** máximo control de SEO, semántica HTML y páginas públicas.
- **Inconvenientes:** duplica UI, rutas y pruebas; rompe parcialmente la
  estrategia universal.
- **Coste de adopción:** alto para el momento actual.
- **Coste de mantenimiento:** alto salvo que la web pública tenga valor claro.

## Comparación

La alternativa B es la solución mínima que satisface el producto inicial privado.
Evita introducir generación estática, SSR o una web especializada antes de tener
una necesidad pública real. A es una evolución natural si aparecen torneos o
páginas públicas conocidas en build. C y D quedan reservadas para requisitos web
más fuertes.

## Recomendación

**Opinión/recomendación:** alternativa B. Dado que los torneos serán privados al
inicio y no hay requisito de SEO, el coste de A, C o D sería prematuro.

La regla de diseño será: universal por defecto, adaptación explícita donde la
plataforma lo justifique, y revisión de rendering si el producto incorpora
superficies públicas.

## Decisión del usuario

**Aceptada:** usar rendering web client-side inicialmente para el cliente Expo
universal.

No se optimizará la web inicial para SEO público, generación estática ni SSR. Las
diferencias por plataforma se aislarán mediante componentes, adaptadores o
archivos específicos (`.web.tsx`, `.native.tsx`, `.ios.tsx`, `.android.tsx`)
cuando una diferencia concreta lo justifique.

## Reglas de implementación

- El cliente comparte pantallas y comportamiento por defecto.
- Las diferencias de plataforma no se introducen dentro de lógica de negocio.
- `Platform` o `Platform.select` se usarán solo para diferencias pequeñas y
  locales.
- Los archivos específicos de plataforma se usarán cuando una implementación
  completa sea distinta.
- Web debe seguir siendo responsive en móvil, tablet y escritorio.
- iOS y Android deben probar la misma intención funcional que web, aunque la
  navegación o interacción visual se adapten.
- Las páginas públicas, si aparecen, exigirán revisar metadata, URLs, previews,
  accesibilidad, rendimiento y estrategia de rendering.

## Consecuencias

### Positivas

- Menos infraestructura inicial para web.
- Menos decisiones prematuras antes del primer producto privado.
- El cliente Expo puede crearse sin servidor web propio.
- La paridad funcional sigue siendo la regla principal.

### Negativas y deuda aceptada

- La web inicial no estará optimizada para SEO ni previews sociales ricos.
- Un futuro catálogo público, landing o torneo indexable puede requerir static
  rendering, SSR o una superficie web especializada.
- Habrá que vigilar que las condiciones de plataforma no se dispersen por el
  código.

## Validación

- Una ruta de prueba debe ejecutarse en web, iOS y Android desde el mismo
  proyecto Expo.
- La web debe funcionar como aplicación responsive client-side.
- Una diferencia de plataforma demostrable debe quedar aislada en componente o
  adaptador.
- Los directorios nativos generados seguirán fuera de Git.

## Disparadores de revisión

- Torneos, clubes, rankings o invitaciones pasan a necesitar indexación pública.
- Se requieren metadatos sociales específicos por torneo o página.
- La carga inicial web o accesibilidad no alcanza el estándar acordado.
- La mayoría de una pantalla necesita implementación distinta en web y native.
- Expo Router incorpora o estabiliza capacidades de rendering que reduzcan el
  coste de adoptar static/SSR para este proyecto.

## Documentación afectada

- [TECHNICAL_BASELINE.md](../governance/TECHNICAL_BASELINE.md)
- [ARCHITECTURE.md](../engineering/ARCHITECTURE.md)
- [DEVELOPMENT.md](../engineering/DEVELOPMENT.md)
- [TESTING.md](../engineering/TESTING.md)
- [DEPLOYMENT.md](../operations/DEPLOYMENT.md)
- [PRODUCT.md](../project/PRODUCT.md)
- [DECISIONS.md](../governance/DECISIONS.md)
- [DECISIONS_TO_REVISIT.md](../governance/DECISIONS_TO_REVISIT.md)
- [LEARNING.md](../project/LEARNING.md)

## Fuentes técnicas

- [Expo static rendering](https://docs.expo.dev/router/web/static-rendering/)
- [Expo server rendering](https://docs.expo.dev/router/web/server-rendering/)
- [React Native platform-specific code](https://reactnative.dev/docs/platform-specific-code.html)
