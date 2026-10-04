# ADR-0148: Preparar fiabilidad de producción con proyectos PostHog separados

- **Estado:** Superado por ADR-0149
- **Superado por:** [ADR-0149](0149-reserve-the-single-posthog-project-for-production.md), tras confirmar el límite de un proyecto y la decisión de reservarlo para producción
- **Fecha:** 2026-10-04
- **Decisor:** Usuario: configurar producción ahora, diferir la prueba y separar los entornos también en PostHog
- **Supera a:** ADR-0105 únicamente en el bloqueo de fiabilidad de producción; conserva ADR-0109 como gate de prueba distribuible

## Problema, evidencia y alternativas

El SDK está instalado y captura fallos en beta, pero producción está bloqueada.
Los entornos web ya tienen hosts y releases separados; una variable pública única
no acredita proyectos PostHog separados ni la simbolización de crashes.

Compartir proyecto y filtrar por entorno tiene menor configuración, pero mezcla
datos y cuotas. Proyectos separados requieren dos claves y configuración de
subida por proyecto, a cambio de aislar la información. Añadir otro proveedor
duplicaría SDK y mantenimiento sin resolver una necesidad adicional.

## Recomendación y decisión explícita

El usuario acepta preparar producción y exige separación de entornos. Se usan
claves públicas distintas para beta y producción, sin fallback entre ellas.
La clave beta existente conserva su nombre; producción dispone de una variable
propia. Local permanece apagado. No se crean proyectos ni credenciales ficticias:
el responsable conectará los proyectos UE reales cuando estén disponibles.

Producción habilita únicamente fiabilidad mínima, independiente del switch de
analítica. La analítica de uso permanece limitada a beta y consentimiento.
Replay, flags, GeoIP, logs, push y captura de interacción permanecen apagados.

## Implementación y coste

Se conserva el SDK fijado. Metro incorpora IDs de depuración y el plugin Expo
prepara source maps Hermes, dSYM y R8 para releases públicos. Las credenciales
CLI son solo de build, por proyecto y fuera de Git/bundle. No se desactiva el
sandbox de scripts Xcode como efecto implícito de este cambio. La configuración
no despliega ni inicia servicios ni transmite símbolos durante la preparación.

## Validación pendiente y retrospectiva

Tests locales verifican selección de entorno, placeholders, separación de claves
y límites de eventos. Exportaciones y configuración no acreditan entrega real.
Antes de distribuir se reconstruyen las apps y se prueba JS y crash nativo en
iOS/Android, recepción, simbolización, filtrado, retención y gasto máximo 0 € en
ambos proyectos. Una producción conectada sin esa evidencia sigue sin validar.

La separación de runtime no sustituye la separación del destino de telemetría;
una opción `nativeCrashes: true` no sustituye símbolos y pruebas de recepción.

Fuentes: [PostHog error tracking](https://posthog.com/docs/error-tracking/start-here)
y configuración del SDK `posthog-react-native` 4.63.2 instalado en el proyecto.
