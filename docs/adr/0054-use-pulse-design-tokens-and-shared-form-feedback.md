# ADR-0054: Usar la dirección Pulse, tokens semánticos y feedback compartido de formularios

- **Estado:** Aceptado
- **Fecha:** 2026-07-28
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** [ADR-0096](0096-use-figtree-as-the-universal-client-font.md), solo en la
  elección de familia tipográfica.

## Problema

Web, iOS y Android deben conservar una experiencia reconocible sin duplicar
colores, medidas, textos de interfaz y reglas de formulario por pantalla.

## Alternativas

### A — Campo

- **Ventajas:** personalidad editorial y deportiva marcada.
- **Inconvenientes:** la tipografía expresiva y las ilustraciones condicionan más
  las pantallas densas de gestión.
- **Coste de mantenimiento:** medio.

### B — Club

- **Ventajas:** eficaz para administración de muchos datos.
- **Inconvenientes:** resulta menos amable como primera experiencia móvil.
- **Coste de mantenimiento:** medio.

### C — Pulse

- **Ventajas:** interfaz luminosa, clara y mobile-first; tarjetas, acciones y
  jerarquía funcionan igualmente en web y aplicaciones nativas.
- **Inconvenientes:** requiere moderar ilustración y color para no ocultar datos.
- **Coste de mantenimiento:** bajo con tokens semánticos y componentes básicos.

## Decisión del usuario

**Aceptada el 2026-07-28:** adoptar la dirección C — Pulse.

- Los colores, espaciados, tipografía, radios, elevaciones y duraciones se
  consumen mediante tokens semánticos compartidos.
- Se empieza con fuentes de sistema por plataforma; no se descarga una fuente
  externa hasta que una necesidad de marca lo justifique.
- Los campos muestran su error de validación inmediatamente debajo del control
  cuando el validador lo determine.
- Las acciones de envío bloquean pulsaciones duplicadas y muestran progreso hasta
  recibir respuesta.
- Los errores globales se muestran en un banner superior, se cierran solos y se
  pueden descartar mediante gesto o botón. Red inaccesible usa un mensaje de
  recuperación; cualquier error no tipado usa un mensaje genérico seguro.

## Consecuencias

- La primera home se construirá después de definir su contenido, sobre los
  mismos tokens y primitivas que registro y futuros formularios.
- Los textos no son tokens visuales: viven en un catálogo semántico para poder
  revisar el copy y localizarlo sin alterar componentes.
- No se introduce todavía una librería de UI, una fuente remota, animaciones
  complejas ni una abstracción de navegación adicional.

## Validación

- Web, iOS y Android usan el mismo nombre de token para la misma intención.
- Un campo inválido conserva foco y muestra un mensaje accesible bajo él.
- Un segundo toque durante un envío no emite otra solicitud.
- El banner global no expone errores internos del backend.

## Disparadores de revisión

- La accesibilidad, la marca o una fuente corporativa exigen cambiar tipografía.
- Una divergencia nativa justifica un token o componente específico de plataforma.
- Tres o más pantallas repiten un patrón que no cubren las primitivas iniciales.

## Documentación afectada

- [Sistema de diseño](../engineering/DESIGN_SYSTEM.md)
- [Desarrollo](../engineering/DEVELOPMENT.md)
- [Decisiones](../governance/DECISIONS.md)
- [Aprendizaje](../project/LEARNING.md)
