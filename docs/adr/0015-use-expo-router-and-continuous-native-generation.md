# ADR-0015: Usar Expo, Expo Router y Continuous Native Generation

- **Estado:** Aceptado
- **Fecha:** 2026-07-24
- **Decisor:** Usuario, mediante confirmación explícita
- **Propietario del análisis:** Asistente como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

El cliente universal necesita un framework que haga viable la entrega web, iOS y
Android para un equipo pequeño. También necesita navegación basada en URLs y una
política explícita para los proyectos nativos generados de iOS y Android.

## Contexto y restricciones

- ADR-0008 acepta un cliente universal basado en React Native, no un framework ni
  un router concreto.
- El producto requiere web responsive, aplicaciones instalables y paridad
  funcional entre plataformas.
- La identidad aceptada usa enlaces HTTPS y deep linking para confirmaciones.
- El toolchain TypeScript, pnpm y workspaces está aceptado en ADR-0014.
- Todavía no se ha creado `apps/client`, ni se ha decidido rendering web,
  adaptación visual, router alternativo, design system o testing del cliente.
- Compartir texto, URLs o archivos no exige código nativo propio: React Native y
  Expo proporcionan APIs y módulos para ello.

## Criterios de decisión

1. navegación coherente en web, iOS y Android;
2. URLs y deep links trazables hasta una pantalla;
3. configuración nativa declarativa y reproducible;
4. posibilidad de integrar capacidades nativas sin mantener Swift, Kotlin,
   Xcode y Gradle antes de necesitarlos;
5. upgrades con coste proporcional;
6. salida posible a código nativo cuando haya evidencia.

## Alternativas

### Alternativa A — Expo, Expo Router y CNG

Usar Expo como framework, Expo Router para rutas por archivos y Continuous Native
Generation (CNG) para generar `ios/` y `android/` desde configuración y plugins.

- **Ventajas:** tooling integrado; rutas universales; deep linking natural;
  proyectos nativos reproducibles; acceso a módulos Expo sin código nativo propio.
- **Inconvenientes:** seguir convenciones Expo; cualquier personalización nativa
  excepcional debe expresarse como config plugin o justificar abandonar CNG.
- **Coste de adopción:** bajo o medio.
- **Coste de mantenimiento:** bajo mientras las necesidades nativas sean
  convencionales.
- **Riesgos:** confundir generación reproducible con ausencia de runtime nativo;
  configurar directamente un directorio generado y perder el cambio al regenerar.

### Alternativa B — Expo con React Navigation directo y CNG

Conservar Expo y generación nativa, pero definir manualmente la navegación y los
links mediante React Navigation.

- **Ventajas:** control explícito de la jerarquía de navegación.
- **Inconvenientes:** más configuración para URLs web, deep links y convenciones
  de rutas que Expo Router ya integra.
- **Coste de adopción:** medio.
- **Coste de mantenimiento:** medio.
- **Riesgos:** divergencia entre rutas web y enlaces nativos.

### Alternativa C — Proyectos nativos versionados y modificados directamente

Generar y conservar `ios/` y `android/` en Git, editándolos desde Xcode y Android
Studio.

- **Ventajas:** control inmediato de cada detalle nativo y experimentación rápida
  con código de plataforma.
- **Inconvenientes:** upgrades, conflictos y configuración nativa pasan a ser
  mantenimiento permanente del equipo.
- **Coste de adopción:** medio o alto.
- **Coste de mantenimiento:** alto antes de que exista una necesidad concreta.
- **Riesgos:** acoplar el producto a cambios manuales difíciles de reproducir.

### Alternativa D — React Native bare sin Expo

Mantener directamente los proyectos nativos y ensamblar navegación, módulos y
tooling manualmente.

- **Ventajas:** control total y ausencia de convenciones Expo.
- **Inconvenientes:** coste inicial de integración, upgrades y operación nativa
  superior; no resuelve por sí mismo URLs ni deep links.
- **Coste de adopción y mantenimiento:** alto.
- **Riesgos:** adelantar complejidad nativa sin necesidad de producto.

## Comparación

La alternativa A satisface rutas universales y configuración reproducible con el
menor coste. B conserva el framework, pero reintroduce trabajo manual sin una
ventaja identificada. C y D son válidas cuando se requieren cambios nativos
propios, pero anticipan una responsabilidad operativa que el producto todavía no
necesita.

## Recomendación

**Opinión/recomendación:** alternativa A. Empezar con CNG no cierra la puerta a
código nativo: Expo puede generar los proyectos en local, los config plugins
pueden personalizarlos y una necesidad demostrada puede reabrir este ADR.

## Decisión del usuario

**Aceptada:** usar Expo como framework del cliente universal, Expo Router para
navegación universal y CNG para los proyectos nativos.

`apps/client/ios` y `apps/client/android` se generarán bajo demanda y no se
versionarán. El código fuente de producto, `app.config`, dependencias y config
plugins serán la fuente de verdad de la configuración nativa.

## Reglas de implementación

- Las rutas viven en `apps/client/src/app`; componentes reutilizables viven fuera
  de ese directorio.
- Una pantalla tiene una URL; los grupos de rutas solo organizan y no añaden un
  segmento a la URL.
- Los layouts de Router contienen navegación y providers de interfaz, no reglas
  de negocio ni llamadas de infraestructura.
- El deep linking HTTPS y la asociación de dominio se configurarán de forma
  declarativa al resolver configuración y secretos; Router no sustituye esa
  asociación de plataforma.
- `apps/client/ios` y `apps/client/android` quedan ignorados por Git y no se
  editan manualmente mientras CNG esté vigente.
- `expo prebuild --clean` puede regenerar los directorios nativos; se ejecutará
  solo con un árbol Git limpio y como operación explícita.
- Los permisos, iconos, bundle identifiers, enlaces y módulos nativos se declaran
  en configuración o config plugins versionados.
- Para compartir texto o URLs se evaluará primero `Share` de React Native; para
  archivos se evaluará `expo-sharing`. No se añadirá una extensión de compartir
  entrante sin un caso de uso aceptado.
- La versión exacta de Expo SDK se fijará al crear `apps/client`, mediante el
  template y la instalación reproducible de pnpm; no se selecciona por este ADR.

## Consecuencias

### Positivas

- Web y aplicaciones nativas comparten una estructura de rutas y deep links.
- Los cambios TypeScript habituales usan Fast Refresh sin recompilar código
  nativo.
- Las integraciones nativas comunes no fuerzan mantenimiento directo de Xcode o
  Gradle.
- El proyecto puede abrirse en simuladores y compilarse localmente cuando haga
  falta.

### Negativas y deuda aceptada

- Algunas bibliotecas o requisitos nativos exigirán config plugins o una
  development build.
- El runtime web conserva límites distintos: el compartir web depende de Web
  Share API y HTTPS.
- CNG exige disciplina para no editar directorios generados.
- Las páginas públicas con requisitos de SEO o rendering avanzado pueden exigir
  configuración o adaptación adicional.

## Validación

- `apps/client` arrancará en web, iOS y Android desde el mismo proyecto Expo.
- Una ruta de prueba se abrirá como URL web y deep link nativo.
- Los directorios nativos generados no aparecerán como cambios Git.
- Una development build iOS se podrá generar en local sin modificar Swift ni el
  proyecto Xcode manualmente.
- El share sheet nativo se probará con un texto/URL o archivo cuando exista el
  primer caso de uso.

## Disparadores de revisión

- Se necesita modificar Swift, Kotlin, Gradle o Xcode de forma permanente.
- Una biblioteca esencial no funciona mediante CNG ni config plugin mantenido.
- Expo Router no cubre una necesidad de navegación, URL o accesibilidad.
- Requisitos web de SEO o rendering obligan a una estrategia distinta.
- La configuración nativa generada deja de ser reproducible o fiable.

## Documentación afectada

- [TECHNICAL_BASELINE.md](../governance/TECHNICAL_BASELINE.md)
- [ARCHITECTURE.md](../engineering/ARCHITECTURE.md)
- [DEVELOPMENT.md](../engineering/DEVELOPMENT.md)
- [DECISIONS.md](../governance/DECISIONS.md)
- [DECISIONS_TO_REVISIT.md](../governance/DECISIONS_TO_REVISIT.md)
- [LEARNING.md](../project/LEARNING.md)

## Fuentes técnicas

- [Expo Router](https://docs.expo.dev/router/basics/core-concepts/)
- [Continuous Native Generation](https://docs.expo.dev/workflow/overview/)
- [Expo Sharing](https://docs.expo.dev/versions/latest/sdk/sharing/)
