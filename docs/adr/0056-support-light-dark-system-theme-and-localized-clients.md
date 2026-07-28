# ADR-0056: Soportar tema claro, oscuro o sistema y clientes localizados

- **Estado:** Aceptado
- **Fecha:** 2026-07-28
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico

## Problema

El cliente universal debe respetar preferencias visuales y de idioma sin que web,
iOS y Android diverjan ni que cada pantalla implemente su propia detección.

## Alternativas

### A — Solo claro y español

- **Ventajas:** implementación mínima.
- **Inconvenientes:** no respeta preferencias ni el público definido.
- **Coste de mantenimiento:** bajo, pero producto insuficiente.

### B — Tema e idioma fijados por cada pantalla

- **Ventajas:** control local inmediato.
- **Inconvenientes:** comportamiento inconsistente y duplicación.
- **Coste de mantenimiento:** alto.

### C — Preferencias centralizadas y tokens semánticos

- **Ventajas:** una fuente de verdad para todas las rutas; las componentes
  consumen semántica y no colores ni textos literales.
- **Inconvenientes:** requiere providers pequeños desde el inicio.
- **Coste de mantenimiento:** bajo.

## Decisión del usuario

**Aceptada el 2026-07-28:** adoptar la alternativa C.

- La versión inicial del producto y del cliente es `1.0.0`.
- Tema: `system`, `light` o `dark`; la elección explícita persiste localmente y
  prevalece sobre el sistema.
- Idiomas soportados: español (`es`), inglés (`en`), italiano (`it`) y francés
  (`fr`). Si el idioma detectado no coincide, el fallback es inglés.
- En iOS y Android el idioma depende exclusivamente del sistema y no se muestra
  selector. En web se detecta el navegador y se ofrece selector persistente.

## Consecuencias

- Todo texto visible se incorporará a catálogos por locale; no se escribe en una
  pantalla nueva sin una clave de traducción.
- Las componentes consumen tokens de tema semánticos, no valores de color fijos.
- El selector web no controla el idioma de la app instalada.

## Validación

- Un sistema o navegador no soportado muestra inglés.
- Al forzar tema, reiniciar conserva la elección.
- Web conserva el locale elegido; apps no ofrecen ese ajuste.
- Ningún control pierde contraste al cambiar de tema.
