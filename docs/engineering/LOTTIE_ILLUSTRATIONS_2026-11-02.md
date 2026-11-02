# Ilustraciones de torneos — 2026-11-02

Decisión: [ADR-0182](../adr/0182-add-lottie-tournament-illustrations.md).

## Comportamiento y mantenimiento

`TournamentIllustration` comparte el dibujo entre web, iOS y Android. El módulo
`tournament-illustration-data.ts` es el recurso vectorial original: curvas,
rellenos y fotogramas Lottie, sin imágenes, fuentes, expresiones ni URLs externas.
Recibe los tokens semánticos del tema. Los nombres, explicaciones y acciones
siguen localizados por la feature; no se añaden textos permanentes.

- Trofeo: 48 fotogramas a 30 fps (1,6 s), dibujo progresivo y cuatro destellos
  pequeños. Sustituye al símbolo en el diálogo existente después de completar
  correctamente. Conserva campeón/cocampeones y acciones originales.
- Biblioteca: 60 fotogramas a 30 fps (2 s). Aparecen cuatro equipos, conexiones y
  copa central. Acompaña a Administro y Sigo vacíos; cada relación conserva su
  copy. La ampliación a Sigo es posterior a v1.17.0 y está pendiente de publicar.
  Cambiar entre segmentos vacíos remonta solo la ilustración para reproducir
  una vez la secuencia, conservando la biblioteca y sus datos.
- Una reproducción por montaje, sin bucle. El último fotograma queda visible.
  Perder foco o pasar a segundo plano termina la secuencia en ese dibujo y no
  la repite al volver. El desmontaje destruye el reproductor.
- Movimiento reducido muestra el dibujo final. Mientras se obtiene la
  preferencia tampoco se inicia movimiento; si su lectura falla, se conserva
  el comportamiento estático. Una notificación más reciente del sistema gana
  a la lectura inicial asíncrona.
- Dibujos decorativos sin eventos táctiles ni anuncios al lector. Cajas con
  proporción fija que se reducen al ancho disponible dentro de los contenedores
  desplazables/modal existentes.

Expo resuelve `lottie-react-native` 7.5.0. Web usa `lottie-web` 5.13.0, build SVG
light, para conservar reproducción local sin WASM/CDN ni expresiones. El
adaptador web clona los datos que Lottie muta internamente. Cambiar el tema
reconstruye el dibujo sin repetir la secuencia.

## Evidencia y límites

### Corrección posterior — inicio de reproducción web

El usuario observa el trofeo estático en Safari iOS. La foto coincide con el
fotograma final; por sí sola no distingue una secuencia ya terminada, movimiento
reducido o un fallo. Se reproduce en prueba un defecto concreto: Lottie 5.13.0
marca `isLoaded=true` antes de emitir `DOMLoaded` mediante un timeout. Nuestro
efecto podía iniciar y consumir la reproducción antes de ese evento; el callback
posterior detenía el dibujo al final sin volver a iniciarlo.

El adaptador web espera explícitamente `DOMLoaded` por instancia, reinicia esa
disponibilidad al cambiar de tema/desmontar y descarta callbacks de otra instancia.
Se mantienen la duración, la reproducción única y el movimiento reducido. Tres
pruebas de regresión cubren el orden de eventos, la preferencia tardía, salida,
desmontaje y cambio de tema. El caso de carrera falla antes del cambio.
Es una reproducción determinista del defecto, no una ejecución real en Safari;
el origen exacto del caso del usuario sigue sin acreditarse. Corrección pendiente
de publicación y de QA visual en Safari iOS.
Typecheck, lint, formato y exportación web correctos. Las tres regresiones pasan
y se incorporan al gate `make verify` mediante `make test-lottie-playback`.

La [lámina claro/oscuro](evidence/lottie-2026-11-02/artwork.png) muestra los
fotogramas finales generados por el reproductor SVG real sobre un DOM XML en
memoria y rasterizados con Sharp. Se renderizaron también los fotogramas 0, 12 y
24 de ambas ilustraciones en ambos temas. Comprueba geometría y color del
recurso; **no es una captura de la app ni prueba interacción o accesibilidad**.
La [secuencia GIF](evidence/lottie-2026-11-02/motion.gif) usa los mismos
fotogramas SVG para revisar dibujo progresivo y tiempos; tampoco sustituye QA.

- Typecheck, lint y Prettier de los archivos afectados: correctos.
- 22 pruebas de tokens, paginación y browsing: correctas.
- 15 pruebas de compatibilidad de dependencias y generación Android: correctas.
  La primera ejecución no pudo abrir el loopback por el sandbox; la repetición
  con ese permiso pasó y cerró el servidor efímero.
- `expo install --check`: compatible con el SDK fijado.
- Exportación web y bundles Hermes iOS/Android: correctos.
- `pnpm peers check`: dos avisos previos (@hono/node-server y
  TypeScript/neverpanic); no aparecen conflictos de Lottie.
- CNG regenerado y Pods instalados, incluyendo lottie-ios 4.6.0 y
  lottie-react-native 7.5.0.
- Android: `assembleDebug` y `lintDebug` correctos (1031 tareas, 13m07s),
  APK generado; lint de app con 0 errores y 46 avisos. Daemon Gradle detenido
  al terminar.
- iOS: build Debug arm64 de simulador genérico correcta (599,823 s), con
  Lottie incorporado y app firmada para simulador. Logs de compilación en
  `/tmp/tm-lottie-ios-generic-build.log` y `/tmp/tm-lottie-android-build.log`.
  La incompatibilidad del servicio de simulador impide instalar/ejecutar y
  acreditar QA en ese entorno; no invalida la compilación genérica.
- Navegador: el control está desactivado en Work; no se acredita QA de rutas,
  resize ni preferencia de movimiento en navegador.
- Simulador iOS deshabilitado por incompatibilidad CoreSimulator 1174.9.2 frente
  a Xcode 1174.21.0. La compilación genérica se comprueba por separado; no se
  cambia el sistema ni se acredita ejecución nativa. Device Hub agota su tiempo.

## Checklist del cliente

Preflight: ADR-0054/55/56/57, identidad Grada ADR-0176, secciones ADR-0178,
arquitectura, desarrollo y sistema de diseño. Solo cambian dos puntos de
composición y las primitivas/recursos de ilustración. No se añaden controles,
rutas permanentes, márgenes globales, navegación ni permisos. Colores y tamaños
consumen tokens; no hay nuevas cadenas que localizar. La ruta efímera de prueba
se retira antes de la exportación final.

No se modifica ninguna operación OpenAPI, adaptador, DTO, contrato o salida del
backend. El trofeo sigue en la rama de respuesta validada de
`completeTournamentRequest`, que usa operación generada y `apiFetch`. Rechazos,
errores y cancelaciones conservan el feedback existente. No se prueban mutaciones
sobre datos reales ni se declara revisado un endpoint sin cambios.

Pendiente de QA visual en las tres plataformas: reproducción única, salida durante
secuencia, cambio de tema/movimiento, diálogo con nombres largos y cocampeones,
cuatro idiomas, texto ampliado, ventanas bajas, rotación e insets. Los checks
estáticos y el dibujo no sustituyen esos recorridos.

## Retrospectiva técnica

Dos ilustraciones comparten un recurso sin construir un sistema de animación
general. La parte delicada es su ciclo de vida: una preferencia asíncrona o una
vista que sigue montada bajo otra ruta puede iniciar movimiento inesperado. El
estado inicial estático y la reproducción consumida limitan esos casos. El coste
añadido queda en dos reproductores y builds nativas; no se extiende Lottie a
controles o transiciones que ya tienen solución.

La publicación posterior autorizada activa solo la web como v1.17.0, con símbolos
PostHog UE 255144 autorizados expresamente y retirados del artefacto público.
[Evidencia del despliegue y rollback](../operations/DEPLOYMENT.md#publicación-v1170--2026-11-02).
No se ha arrancado local/dev, observabilidad ni tareas programadas. La QA visual
indicada arriba y la distribución móvil continúan pendientes.
