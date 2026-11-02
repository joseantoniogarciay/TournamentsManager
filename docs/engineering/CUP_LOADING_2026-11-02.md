# Copa de carga — 2026-11-02

Decisión: [ADR-0184](../adr/0184-use-a-forming-cup-for-full-screen-loading.md).
Incluido en la publicación web v1.18.0.
[Evidencia y rollback](../operations/DEPLOYMENT.md#publicación-v1180--2026-11-02).

## Comportamiento

`LoadingTransition` conserva active/message, overlay, bloqueo de interacción y
rol progressbar con mensaje localizado. Cambia el spinner por `CupLoader`:

- Ciclo de 60 fotogramas a 30 fps (2 s). Copa trazada y formada hasta el frame
  30, pausa breve y fade alpha 100→0 entre 36 y 56; reinicio invisible al final.
- Bucle exclusivamente mientras active, app visible, preferencia resuelta y
  movimiento permitido. Ocultar, finalizar o desmontar detiene el reproductor.
- Movimiento reducido, preferencia aún desconocida y app en segundo plano
  muestran el frame 32 completamente formado. Nunca se usa el frame final
  transparente como fallback accesible.
- El texto permanece visible durante los ciclos. No es un porcentaje ni un
  éxito del torneo; la carga no espera a que termine un ciclo ni recibe un mínimo
  artificial. Las cinco ilustraciones contextuales siguen reproduciéndose una vez.
- Cargas pequeñas de botón, búsqueda y paginación conservan sus indicadores.

La copa reutiliza las dos capas vectoriales del trofeo, sin halo ni destellos,
coloreadas con tokens Grada. `CupLoader` tiene adaptadores web y nativo; no se
añade dependencia. `useIllustrationEnvironment` comparte preferencia/app state
sin navegación. Esto permite usarlo en SessionProvider, fuera de una pantalla;
`useIllustrationPlayback` añade el foco solo a los momentos de pantalla.

El overlay anima desde su alpha actual, sin volver a cero al resolver la
preferencia. El contenido reserva el margen horizontal compartido y centra el
mensaje, manteniendo su escalado de texto y semántica accesible.

## Evidencia y límites

[Fotogramas 0/16/32/44/59](evidence/cup-loader-2026-11-02/frames.png) y
[bucle claro/oscuro](evidence/cup-loader-2026-11-02/motion.gif), renderizados por
Lottie SVG sobre DOM XML y Sharp. Se comprueba que el frame formado está visible
y el final es transparente. Es evidencia del recurso, no una captura de la app.

Ocho regresiones del conjunto Lottie/invitación correctas, incluidas espera de
DOMLoaded, loop, parada formada, cambio de permiso antes de cargar el SVG y
desmontaje. Typecheck, lint, exportaciones web/Hermes iOS/Android y make verify
local completos correctos; integración PostgreSQL genérica omitida sin DB
configurada. No hay cambios de HTTP, datos, backend ni contratos.

Se revisa el preflight de apps/client/AGENTS y las mismas decisiones de interfaz
de ADR-0183. Primitivas, mensajes existentes es/en/it/fr, márgenes, tema y rol
se conservan; no se añaden controles ni dependencias nativas. QA real en Safari,
rotación, texto ampliado y dispositivos sigue pendiente por los límites conocidos.
No se distribuyen binarios ni se arrancan local/dev u observabilidad.

## Retrospectiva

Un loader global no debe depender de un hook de foco de pantalla. Separar ese
foco de la preferencia/app state permite reutilizar las reglas accesibles sin
acoplar la transición de sesión al navegador. Un bucle con fade requiere además
un frame estático elegido antes de la transparencia final: detenerlo al final
ocultaría precisamente el indicador que debe seguir comunicando espera.
