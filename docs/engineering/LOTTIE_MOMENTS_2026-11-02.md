# Momentos Lottie y nueva publicación — 2026-11-02

Decisión: [ADR-0183](../adr/0183-complete-contextual-lottie-moments.md), aceptada
por «mete todas las otras animaciones de las otras propuestas y hacemos nueva subida».
Amplía las dos ilustraciones de ADR-0182, conservando sus reproductores y tokens.

## Resultado de implementación

| Recurso | Disparador real | Duración | Presentación |
| --- | --- | --- | --- |
| Trofeo | Finalización confirmada | 1,6 s | Diálogo existente |
| Cuadro | Administro o Sigo vacío | 2 s | Estado vacío; una vez por relación montada |
| Silbato | Inicio confirmado con estado in_progress | 1,6 s | Resumen de la ficha, sin diálogo nuevo |
| Camisetas | Invitación aceptada y navegación al torneo | 1,6 s | Resumen de la ficha de destino |
| Bienvenida | Inicio sin sesión | 2,6 s | Hero existente; una vez por sesión JS |

La bienvenida conserva el dibujo final cuando se vuelve a Inicio; `GuestWelcome`
marca la presentación en un efecto para que el export estático no consuma una
visita. El balón recorre el cuadro y da paso al avance de la marca deportiva.
No sustituye al splash, no reproduce sonido y no retiene a la persona.

La invitación conserva `router.replace` inmediato. `invitation-moment.ts` guarda
un único resultado confirmado en memoria, acotado por torneo/cuenta y consumido
una vez al resolver la ficha de destino. Otra cuenta descarta la señal anterior;
no hay almacenamiento, URLs de éxito, roles, permisos ni estado de negocio nuevo.
El silbato solo procede de iniciar con éxito en esa pantalla y se oculta si el
torneo deja de estar en curso. Las ilustraciones solo decoran datos existentes.

Todas usan colores semánticos Grada, proporción adaptable, SVG/formas nativas y
fallback final con movimiento reducido. `animate=false` permite conservar la
bienvenida estática en visitas posteriores sin eludir la preferencia del SO.
Foco, segundo plano y desmontaje conservan las reglas de ADR-0182.

Se incluye el arreglo web comunicado tras la foto de Safari iOS: `DOMLoaded`
debe llegar antes de consumir la reproducción. La propiedad `isLoaded` puede
ser verdadera antes del evento; iniciar entonces dejaba que el callback inicial
detuviera la animación en el fotograma final. La regresión reproduce el fallo
antes del arreglo. No demuestra que fuera la causa exacta del móvil del usuario.

## Operaciones y feedback

No se cambia OpenAPI, backend ni su inventario de observabilidad. La revisión
de los dos adaptadores y sus respuestas conserva:

- Inicio: función generada + `authenticatedApiFetch`; 200 parseado antes del
  silbato, 422 reconocido con mensaje localizado de configuración y resto con
  fallback seguro existente.
- Invitación: operación generada + `authenticatedApiFetch`; 201 parseado antes
  de recordar el momento. 404 y 409 conocidos conservan recuperación/localización
  de la feature; problemas no reconocidos y otros estados conservan el fallback.
- Transporte identificado: `common_network_error`; 5xx, cuerpos inválidos y
  problemas inesperados: `common_request_error`. No se muestran title/detail
  ni cuerpos internos. AbortError intencional no produce feedback ni momentos.
- No se generan celebraciones en catch ni al visitar una URL con parámetros
  inventados. No se añaden logs ni spans con identificadores de cuenta/torneo.

## Validación y límites

- Seis regresiones en `make test-lottie-playback`: carrera DOMLoaded, preferencia
  tardía, foco/desmontaje/tema, bienvenida estática y consumo de invitación por
  cuenta/torneo. Se incluyen en `make verify`.
- Typecheck, lint dirigido y exportaciones web/Hermes iOS/Android correctos.
- [Lámina claro/oscuro](evidence/lottie-moments-2026-11-02/artwork.png) y
  [secuencia animada](evidence/lottie-moments-2026-11-02/motion.gif) renderizadas
  con Lottie SVG sobre DOM XML y Sharp. Se revisan todos los fotogramas, sin
  backend ni datos reales. No son capturas de la aplicación.
- `make verify` local completo correcto; integración PostgreSQL genérica omitida
  sin DB configurada. CI y publicación se registran al cerrar el despliegue.
- Preflight de cliente: ADR-0054/55/56/57/0176/0178, arquitectura, desarrollo,
  diseño y AGENTS. Primitivas/márgenes existentes, arte decorativo, sin cadenas
  nuevas que localizar, controles o dependencias nativas nuevas. No se recompila
  un módulo que no cambia; los bundles nativos no equivalen a QA en dispositivo.
- QA visual real de Safari iOS, demás navegadores, iOS/Android, rotación,
  texto ampliado y recorridos autenticados pendiente por los límites conocidos.
  No se distribuyen binarios móviles. No se arrancan local/dev ni observabilidad.

## Retrospectiva

El coste de ampliar una ilustración es pequeño si conserva el reproductor y sus
reglas. La integración entre rutas requiere más cuidado que el dibujo: un
resultado real se transmite mediante una señal efímera y acotada, nunca mediante
una bandera de URL. Una marca por sesión evita convertir la bienvenida en una
animación repetida de navegación. Los fotogramas, pruebas del ciclo de vida y
checks de hosting aportan evidencias distintas de la QA visual real pendiente.
