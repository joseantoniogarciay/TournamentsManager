# ADR-0176: Adoptar la identidad Grada conservando componentes y navegación nativa

- **Estado:** Aceptado; implementado; builds verificadas, QA visual parcial
- **Fecha:** 2026-11-01
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** ADR-0054 en dirección cromática y ADR-0096 en dirección tipográfica,
  en la implementación del cliente. La validación visual de la transición sigue
  pendiente; no se cambia la anatomía de las primitivas ni la navegación.
- **Superado por:** Ninguno

## Problema

El usuario percibe la interfaz como demasiado sobria y genérica. Solicita estudiar
tres alternativas con datos ficticios y elige Grada por sus colores y tipografía,
conservando los componentes actuales y las tab bars nativas.

## Contexto y evidencia

El gate técnico está cerrado. Se compararon tres exploraciones de Inicio, ficha
de torneo y clasificación: Grada, Pista y Club. Son referencias visuales, no
implementaciones del cliente universal ni evidencia de ejecución nativa.

El cliente actual usa tokens semánticos, primitivas compartidas y `NativeTabs`
de Expo Router en móvil. La maqueta Grada usa navegación HTML de demostración;
esa navegación no forma parte de lo elegido. La estructura responsive vigente,
las rutas, los insets, la accesibilidad y la localización conservan sus acuerdos.

## Alternativas y coste

- **Mantener Pulse:** coste inmediato nulo y mantenimiento bajo; conserva la
  percepción que motiva este estudio.
- **Grada, identidad sobre componentes existentes:** color cálido y tipografía
  deportiva mediante tokens y primitivas compartidos. Adopción moderada por
  fuentes locales, contraste y verificación transversal; mantenimiento bajo al
  conservar componentes y navegación.
- **Grada, trasladando también la composición de la maqueta:** aumenta el trabajo
  de layout, navegación y QA, sin responder a la preferencia expresada de
  conservar los componentes actuales. Adopción y mantenimiento mayores.
- **Pista o Club:** alternativas visuales comparadas y no elegidas. Pista exige
  especial cuidado con contraste y densidad; Club da más protagonismo a la serif
  y a la composición editorial.

## Recomendación

**Opinión/recomendación:** aplicar la identidad Grada sobre el sistema existente.
La separación entre identidad y comportamiento permite ganar personalidad sin
crear otro sistema de componentes. No hace falta una nueva biblioteca de UI.

## Decisión explícita del usuario

**Aceptada el 2026-11-01:** «El mejor diseño y colores es el de grada utilizariamos
tipografia y colores pero se mantiene lo que son nuestros componentes, quiero
utilizar tab bara nativas y tal no como esta en este diseño».

El alcance aceptado es:

- Adoptar como referencia cromática el naranja quemado, los fondos cálidos y la
  tinta oscura de Grada. La muestra clara usa `#B93B19`, `#F5F1E8` y `#24241F`;
  son referencias para construir tokens semánticos accesibles, no literales que
  copiar por pantalla. El modo oscuro requiere su adaptación y validación.
- Adoptar la dirección tipográfica de Grada: Barlow Condensed para titulares
  expresivos, DM Sans para lectura y controles; IBM Plex Mono como referencia
  auxiliar para rótulos y cifras cuando aporte legibilidad. Los pesos y usos
  concretos deben verificarse con los componentes reales y texto ampliado.
- Conservar las primitivas actuales de tarjetas, botones, campos, diálogos,
  feedback y layout, aplicando color y tipografía mediante el sistema compartido.
- Conservar `NativeTabs`, cabeceras y controles nativos, incluidos los materiales
  y comportamientos que corresponden a cada plataforma. Se adapta su tintado y
  tipografía solo donde la API nativa lo permita.
- Conservar las secciones Inicio, Torneos, Mis partidos y Cuenta y sus rutas.
  La elección no adopta la cabecera, navegación HTML, tamaños de titulares,
  márgenes, bordes, radios ni disposición de tarjetas de la maqueta.
- Mantener carga local de fuentes en el cliente universal, temas claro/oscuro,
  idiomas y escalado accesible. El CDN usado en la exploración no es el mecanismo
  de distribución del producto.

## Consecuencias y estado

La elección define la identidad, no una reescritura funcional. No modifica API,
datos ni permisos. La implementación deberá revisar contrastes y los consumidores
de los tokens para evitar que queden colores de marca anteriores mezclados con
Grada. No se presupone que una sola familia de fuente sirva para todos los roles.

**Implementación del 2026-11-01:** tokens cálidos claro/oscuro, familias locales
por rol, primitivas existentes, tintado de NativeTabs y tema de React Navigation.
Las cifras de clasificación usan IBM Plex Mono; navegación y controles usan
DM Sans, y los titulares de contenido, Barlow Condensed. Se conserva la escala.
El fondo de splash consume el canvas compartido. Tras la petición «Haz lo que
falte», se adaptan también el icono, PNG transparente de splash, preview social
y email de verificación, conservando composiciones y fallback tipográfico de
sistema en correo. No se publica ni despliega este cambio.

La [evidencia de implementación](../engineering/GRADA_IDENTITY_2026-11-01.md)
registra contraste, fuentes, exportaciones y límites. Las builds iOS/Android están verificadas y se acredita QA parcial de iOS en Duo
e iPad, incluidos los cuatro idiomas en Inicio. La revisión web, la inspección
visual Android y la matriz completa de gestos/ventanas siguen pendientes por
los límites documentados; no se presenta ese cierre como completo.

## Validación requerida para cerrar la transición

- Comparar Inicio, ficha y clasificación usando las primitivas reales.
- Verificar contraste, feedback semántico y controles en claro y oscuro.
- Comprobar los cuatro locales, nombres largos, ampliación de texto y reflow.
- Ejecutar typecheck y exportación web según la checklist del cliente.
- Validar por separado iOS y Android en una build nativa: tabs, cabeceras,
  materiales, insets, rotación y estado al navegar. Una maqueta HTML no demuestra
  esos recorridos.

## Aprendizaje y retrospectiva de la exploración

La identidad de marca y la anatomía de los componentes son decisiones separables.
Una exploración puede validar color y tipografía sin aprobar sus layouts o su
navegación. El feedback del usuario acota aquí esa diferencia y evita convertir
una muestra visual en una sustitución innecesaria del sistema nativo.

## Disparadores de revisión

- La condensación tipográfica impide leer títulos largos o texto ampliado.
- Un color no alcanza contraste en un rol o tema concreto.
- Una API nativa limita la personalización; se documenta la adaptación antes de
  reemplazar un control nativo por otro personalizado.

## Documentación afectada

- [Registro de decisiones](../governance/DECISIONS.md).
- ADR-0054, ADR-0096, DESIGN_SYSTEM.md y aprendizaje reflejan la
  implementación y los límites de validación.
