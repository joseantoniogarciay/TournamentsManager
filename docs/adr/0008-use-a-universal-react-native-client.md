# ADR-0008: Usar un cliente universal con React Native

- **Estado:** Aceptado
- **Fecha:** 2026-07-24
- **Decisor:** Usuario, mediante elección explícita de la alternativa A
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

El producto debe estar disponible en navegadores de escritorio, móvil y tablet,
además de como aplicación instalable en teléfonos y tablets. Hay que decidir si
esas superficies evolucionan desde un único cliente universal o desde
aplicaciones web y mobile especializadas.

## Contexto y restricciones

- Web, iOS y Android representan el mismo producto y deben conservar paridad
  funcional.
- Una misma capacidad debe mantener reglas, estados y significado entre
  plataformas.
- El diseño debe ser responsive y adaptarse a móvil, tablet y escritorio.
- Web, iOS y Android conservan artefactos, canales de distribución y despliegues
  diferentes.
- El repositorio es un monorepo de producto conforme a
  [ADR-0005](0005-use-a-product-monorepo.md).
- Esta decisión no selecciona todavía Expo, router, estrategia de renderizado,
  gestión de estado ni design system.

Paridad funcional no significa identidad píxel a píxel. Navegación, densidad,
entrada por teclado o ratón, gestos, semántica y APIs de plataforma pueden exigir
presentaciones adaptadas.

## Criterios de decisión

1. consistencia funcional entre web, iOS y Android;
2. calidad responsive en móvil, tablet y escritorio;
3. reutilización de pantallas, navegación y comportamiento;
4. mantenibilidad para un equipo pequeño;
5. accesibilidad, rendimiento y descubrimiento de la web pública;
6. independencia de build, despliegue y publicación.

## Alternativas

### Alternativa A — Cliente universal React Native

Un mismo árbol de cliente sirve web, iOS y Android mediante React Native para
web. Puede contener implementaciones específicas por plataforma cuando una
diferencia real lo justifique.

- **Ventajas:** máxima coherencia funcional; una implementación principal de las
  pantallas; menor duplicación; evolución coordinada entre superficies.
- **Inconvenientes:** la web pública, las pantallas densas y la accesibilidad
  avanzada pueden necesitar adaptaciones; las diferencias de plataforma pueden
  acumularse en el mismo árbol.
- **Coste de adopción:** bajo o moderado; exige diseñar desde el inicio layouts
  adaptativos y límites de plataforma.
- **Coste de mantenimiento:** bajo mientras las experiencias permanezcan
  alineadas; aumenta con cada excepción específica.
- **Riesgos:** sacrificar calidad web para maximizar reutilización o convertir el
  código universal en una red de condiciones de plataforma.

### Alternativa B — Clientes web y mobile especializados

Una aplicación React web y otra React Native mobile comparten contratos, cliente
API, validación y lógica portable, pero mantienen UI y navegación propias.

- **Ventajas:** adaptación natural a cada plataforma; control directo de SEO,
  semántica web y experiencias nativas.
- **Inconvenientes:** más duplicación y riesgo de deriva funcional.
- **Coste de adopción:** moderado; dos aplicaciones y dos superficies de pruebas.
- **Coste de mantenimiento:** moderado y explícito.
- **Riesgos:** resolver dos veces una misma pantalla o entregar capacidades en
  momentos distintos.

### Alternativa C — Web/PWA primero

Se entrega primero una web responsive y se aplazan las aplicaciones nativas.

- **Ventajas:** menor alcance inicial y validación rápida en navegador.
- **Inconvenientes:** rompe temporalmente la paridad y difiere requisitos nativos.
- **Coste de adopción:** bajo inicialmente.
- **Coste de mantenimiento:** diferido hasta incorporar mobile.
- **Riesgos:** que la arquitectura y la experiencia queden sesgadas hacia web.

## Comparación

La alternativa A satisface mejor la paridad y la reutilización que el usuario
considera prioritarias. La alternativa B ofrece más libertad por plataforma, pero
introduce dos implementaciones visuales aunque el producto deba permanecer
alineado. La alternativa C reduce trabajo inicial a costa de posponer una parte
del producto aceptado.

La calidad web no queda resuelta por escoger A: deberá validarse de forma
independiente en responsive, accesibilidad, rendimiento, URLs y descubrimiento.

## Recomendación

**Opinión/recomendación previa del analista:** alternativa B, porque la web
pública y las herramientas de administración podrían beneficiarse de una
implementación especializada.

**Recomendación ajustada al requisito aclarado:** la alternativa A es coherente
si la paridad funcional 1:1 es una restricción del producto y se acepta que la
reutilización no prevalece sobre la calidad de cada plataforma.

## Decisión del usuario

Adoptar la alternativa A: un cliente universal con React Native para web, iOS y
Android.

El usuario fundamenta la decisión en que la web debe funcionar de forma
responsive en navegadores de móvil y tablet, las aplicaciones también deben
instalarse en tablets y todas las superficies representan el mismo producto con
paridad funcional.

## Consecuencias

### Positivas

- Una capacidad se implementará de forma universal por defecto.
- Web, iOS y Android compartirán comportamiento, rutas y UI cuando la experiencia
  resultante sea adecuada.
- Los layouts se diseñarán desde el inicio para móvil, tablet y escritorio.
- La paridad será más fácil de revisar y probar.
- Se reduce el riesgo de deriva entre dos clientes independientes.

### Negativas y deuda aceptada

- El cliente deberá manejar diferencias reales de entrada, navegación, densidad,
  semántica y APIs de plataforma.
- Las páginas públicas podrán exigir rendering y metadatos específicos para web.
- Tablas, cuadros de torneo o herramientas de administración podrán necesitar
  componentes específicos.
- Los tres targets comparten una base de código, pero no un único pipeline ni un
  único release.

### Reglas de implementación

- “Universal por defecto” no prohíbe archivos o componentes específicos de web o
  native.
- No se reducirá accesibilidad, rendimiento ni usabilidad para aumentar un
  porcentaje de código compartido.
- No se introducirán condiciones de plataforma dentro de la lógica de negocio.
- Las diferencias se aislarán en componentes, adaptadores o entrypoints de
  plataforma.
- La paridad exigida es funcional y semántica; la composición visual y la
  navegación pueden adaptarse al dispositivo.

## Validación

Antes de considerar listo un flujo de cliente deberá comprobarse:

- el mismo comportamiento funcional en web, iOS y Android;
- layouts utilizables en anchos representativos de móvil, tablet y escritorio;
- navegación mediante tacto y, donde corresponda, teclado y ratón;
- accesibilidad y semántica apropiadas en web;
- URLs, metadatos y estrategia de rendering adecuadas para contenido público;
- builds y entregas independientes para web, iOS y Android.

La matriz concreta de dispositivos, navegadores y pruebas se decidirá en la
estrategia de testing.

## Disparadores de revisión

- La mayoría de una capacidad necesita implementaciones distintas para web y
  native.
- No se pueden cumplir objetivos acordados de SEO, accesibilidad o rendimiento
  web sin eludir el runtime universal.
- Las herramientas de administración requieren una superficie web
  sustancialmente distinta.
- El acoplamiento de dependencias o releases ralentiza de forma medible una
  plataforma.
- Los equipos o ciclos de producto pasan a ser independientes por plataforma.

## Documentación afectada

- [TECHNICAL_BASELINE.md](../../TECHNICAL_BASELINE.md)
- [ARCHITECTURE.md](../../ARCHITECTURE.md)
- [PRODUCT.md](../../PRODUCT.md)
- [SYSTEM_OPTIONS.md](../../SYSTEM_OPTIONS.md)
- [TESTING.md](../../TESTING.md)
- [DEPLOYMENT.md](../../DEPLOYMENT.md)
