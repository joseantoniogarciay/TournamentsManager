# ADR-0184: Copa que se forma y desvanece en las cargas de pantalla completa

- **Estado:** Aceptado; implementado; publicación y QA visual pendientes
- **Fecha:** 2026-11-02
- **Decisor:** Usuario: prefiere una copa que se forme, se desvanezca con alpha y vuelva a formarse en los loaders grandes.
- **Supera a:** Apariencia del indicador de LoadingTransition; conserva su contrato y accesibilidad.

## Problema y alternativas

El usuario acepta las ilustraciones de ADR-0183 y solicita ese lenguaje también
en las cargas que ocupan gran parte de la pantalla. Mantener el spinner cuesta
menos, pero no cumple la preferencia. Crear loaders por feature duplicaría dibujo,
accesibilidad y mantenimiento. Se recomienda y acepta cambiar LoadingTransition
compartido reutilizando la geometría de copa Lottie existente.

## Decisión

La copa se dibuja, queda formada brevemente y se desvanece hasta alpha cero;
repite el ciclo solo mientras la carga está activa y la app visible. Los botones
y cargas parciales conservan sus indicadores contextuales. Los cinco momentos
de ADR-0183 siguen siendo de reproducción única; el bucle es exclusivo del loader.

Movimiento reducido muestra la copa completamente formada, sin el ciclo ni fade
de transición. El mensaje localizado y el rol progressbar de LoadingTransition
se conservan. La copa es decorativa y no representa porcentaje, éxito ni tiempo
restante. No se añade una duración mínima artificial a la operación.

El loader debe funcionar también en SessionProvider, fuera de una pantalla:
su reproducción no depende del foco de navegación. Se comparte la lectura de
preferencia/app visibility mediante un hook independiente; los momentos de
pantalla conservan además su foco. Al terminar, ocultar o desmontar se detiene
el reproductor, sin dejar un bucle en segundo plano.

## Validación y cierre

Regresiones del primer DOMLoaded, bucle activo y parada/fotograma formado;
render claro/oscuro, transparencia al final y ausencia de animación con
movimiento reducido. Typecheck, exportaciones y make verify. No se añade módulo
nativo, negocio, contrato HTTP ni infraestructura. Se incluye en la nueva subida
web solicitada antes de activar v1.18.0, con evidencia/retrospectiva y límites de
QA visual real. Los servicios local/dev y observabilidad permanecen apagados.

## Evidencia de implementación

Ocho regresiones Lottie/invitación, typecheck, lint y make verify local correctos;
integración PostgreSQL genérica omitida sin DB configurada. Exportaciones
web/Hermes iOS/Android correctas. Render de los 60 frames en ambos temas,
incluidos el frame formado de fallback y el final transparente. Véase
[evidencia y retrospectiva](../engineering/CUP_LOADING_2026-11-02.md).
