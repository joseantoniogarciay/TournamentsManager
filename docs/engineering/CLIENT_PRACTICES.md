# Buenas prácticas del cliente universal

> Guía propia para React Native, Expo Router y web. Resume principios de React,
> React Native y Expo; no copia un estilo personal ni código de otro autor.

## Objetivo

Entregar una interfaz rápida, accesible y coherente en web, iOS y Android sin
convertir optimizaciones hipotéticas en arquitectura permanente. Se mide antes de
optimizar y se comparte solo lo que ya se ha repetido.

## Reglas de arquitectura

- Organizar por feature: ruta, pantalla, componentes exclusivos, hook de
  coordinación, validación y adaptador del cliente OpenAPI viven juntos.
- La renderización es pura: no inicia peticiones, no muta estado externo ni
  genera valores inestables durante el render. Los efectos viven en eventos o
  efectos explícitos.
- Las reglas de negocio y autorización son del backend. El cliente valida formato,
  experiencia y estado transitorio; nunca presupone que su validación es control
  de seguridad.
- `shared` contiene solo tokens, UI, feedback y utilidades reutilizadas por más
  de una feature. No se crea un repositorio, caso de uso o store global vacío.
- Estado local por defecto. Contexto solo para tema, locale, sesión y feedback.
  Una nueva librería de estado o caché exige evidencia de repetición o perfilado.

## Rendimiento

- Medir en build de producción o profiling antes de añadir `memo`, `useMemo` o
  `useCallback`; no son una convención automática.
- Mantener props simples y estables para componentes repetidos. Evitar pasar
  objetos, arrays o callbacks recién creados cuando produzcan renders costosos
  medidos.
- No bloquear el hilo JavaScript con transformaciones grandes, serialización o
  cálculos dentro de una pulsación; paginar, diferir o mover el cálculo si la
  medición lo justifica.
- Optimizar imágenes por tamaño de uso, cache y formato antes de mostrar muchas
  a la vez. No cargar una imagen original grande para un avatar o miniatura.
- Comprobar memoria, FPS, red y renders con React Native DevTools/Profiler y el
  monitor de rendimiento de Expo; las métricas de desarrollo no sustituyen una
  comprobación de producción.

## Listas y colecciones

- Para listas cortas y estáticas, una composición normal es suficiente.
- Para colecciones potencialmente largas, desplazables o paginadas en móvil, usar
  `FlatList`/`SectionList`, no `ScrollView` con todos los elementos montados.
- Cada fila recibe una clave estable del dato (`league.id`, `team.id`), nunca el
  índice ni un valor aleatorio.
- Extraer `renderItem`; mantener las filas pequeñas y aisladas. Aplicar memoización
  solo si un perfil demuestra renders innecesarios.
- Definir `keyExtractor`; si todas las filas tienen altura conocida y el perfil lo
  necesita, añadir `getItemLayout`. Ajustar `windowSize`, `initialNumToRender` o
  `removeClippedSubviews` solo tras observar huecos, memoria o falta de respuesta.
- La web mantiene la semántica que corresponda: una colección de datos tabulares
  se presenta como tabla accesible cuando aporta comprensión, no como tarjetas
  por imitar el móvil.

## Componentes, formularios y feedback

- Consumir tokens semánticos de `@tournaments-manager/design-tokens`; no repetir
  hexadecimales, tamaños o textos de estado dentro de una pantalla.
- Reutilizar `Button`, `TextField`, `Picker`, `Card`, `InlineMessage` y `Banner`
  cuando estén estabilizados. No crear variantes “casi iguales” por feature.
- Validar formato al abandonar el campo y siempre al enviar. El mensaje se muestra
  bajo su control, conserva el valor y explica cómo corregirlo.
- Mientras una acción está en vuelo, el botón muestra progreso y queda desactivado
  para impedir envíos duplicados. La UI se vuelve a habilitar al recibir respuesta.
- Error de red: mensaje recuperable; error no tipado: mensaje genérico seguro.
  Los errores globales van al banner superior, no exponen trazas ni cuerpos HTTP.

## Accesibilidad, tema y localización

- Toda acción tiene rol, etiqueta y estado accesible (`disabled`, `busy`,
  `selected`, etc.). Iconos sin texto requieren etiqueta; el color nunca es el
  único portador de significado.
- Mantener objetivos táctiles de al menos 44 px y orden de foco lógico. Un modal
  debe poder cerrarse por teclado, gesto y lector de pantalla.
- Probar contraste y ambos temas. Los componentes usan tokens semánticos, por lo
  que claro, oscuro y sistema no requieren ramas de color en cada pantalla.
- Todo texto visible procede de catálogos `es`, `en`, `it` o `fr`. La app usa el
  idioma del sistema; web detecta navegador y puede guardar la selección. Fallback
  siempre en inglés.

## Web y nativo

- Compartir intención y feature; aislar diferencias completas en archivos
  `.web.tsx` o `.native.tsx`. `Platform.select` se reserva para una diferencia
  pequeña y local.
- Una URL profunda es canónica. Web la presenta como ruta; móvil puede presentarla
  modalmente y debe definir el cierre en inicio en frío.
- No usar HTML ni APIs de navegador directamente en código compartido sin un
  adaptador web. No usar APIs nativas en web sin comprobar la plataforma.

## Definition of Done para una pantalla

- usa componentes/tokens y textos localizables;
- trata carga, vacío, éxito y error;
- no repite solicitud durante envío;
- tiene navegación y cierre correctos en web y móvil;
- pasa typecheck, lint y prueba proporcional al riesgo;
- si contiene lista larga, incluye estrategia de virtualización y clave estable;
- se ha probado con tamaño de fuente, tema y lector de pantalla cuando cambian
  controles o navegación.

## Fuentes y aprendizaje

- [Pureza de componentes — React](https://react.dev/learn/keeping-components-pure)
- [Claves en colecciones — React](https://react.dev/learn/rendering-lists)
- [Optimización de FlatList — React Native](https://reactnative.dev/docs/0.74/optimizing-flatlist-configuration)
- [Accesibilidad — React Native](https://reactnative.dev/docs/accessibility)
- [Herramientas de perfilado — Expo](https://docs.expo.dev/debugging/tools/)
