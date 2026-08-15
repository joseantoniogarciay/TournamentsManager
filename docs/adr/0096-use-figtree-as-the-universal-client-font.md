# ADR-0096: Usar Figtree como tipografía universal del cliente

- **Estado:** Aceptado
- **Fecha:** 2026-08-14
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** ADR-0054, solo en la elección de familia tipográfica.
- **Superado por:** Ninguno

## Problema

La tipografía de sistema hace que la web, iOS y Android usen familias distintas
y da a la interfaz Pulse una apariencia correcta pero genérica. El producto
necesita una voz tipográfica reconocible y consistente sin adoptar la fuente
propietaria de otra marca.

## Contexto y restricciones

- ADR-0054 estableció inicialmente fuente de sistema y previó que una necesidad
  de marca activaría su revisión.
- El cliente soporta español, inglés, italiano y francés, y debe cargar la
  familia localmente en web, iOS y Android; no se consulta un CDN al usar la app.
- `Text`, `TextField`, cabeceras y etiquetas de tabs consumen tokens compartidos.
- Airbnb Cereal es una fuente creada por Airbnb con Dalton Maag, no un asset que
  el producto pueda reutilizar.

## Criterios de decisión

1. Misma familia y pesos reales en las tres plataformas.
2. Licencia apta para distribuir el producto.
3. Legibilidad en formularios y tablas compactas.
4. Integración mínima con Expo y los tokens existentes.
5. Carga inicial sin mostrar brevemente una fuente distinta.

## Alternativas

### A — Mantener fuentes de sistema

- Ventajas: no añade assets ni tiempo de carga.
- Inconvenientes: cada plataforma conserva una personalidad distinta y no crea
  una voz de marca.
- Coste de adopción y mantenimiento: nulo.
- Riesgos: la interfaz continúa percibiéndose genérica.

### B — Figtree local con cuatro pesos

- Ventajas: familia abierta, cálida y legible; conserva regular, medium,
  semibold y bold como pesos reales y se aplica desde un único token.
- Inconvenientes: añade cuatro assets tipográficos al bundle y retrasa el primer
  render hasta que se cargan.
- Coste de adopción: bajo; una carga compartida y los tokens existentes.
- Coste de mantenimiento: bajo; actualizar un único paquete y mantener los
  pesos efectivamente utilizados.
- Riesgos: usar `fontWeight` sobre la fuente regular produciría una síntesis de
  peso; cada peso debe usar su familia Figtree correspondiente.

### C — Fuente comercial o personalizada

- Ventajas: diferenciación más exclusiva.
- Inconvenientes: licencia, presupuesto, negociación y gestión de archivos.
- Coste de adopción y mantenimiento: medio o alto.
- Riesgos: dependencia de proveedor y mayor complejidad antes de validar marca.

## Comparación

La alternativa A minimiza bytes, pero no satisface la paridad visual solicitada.
La C puede aportar una identidad más exclusiva, pero introduce coste sin una
necesidad demostrada. B satisface la paridad, licencia y legibilidad con el
menor cambio transversal: la fuente viaja empaquetada con Expo y los tokens
evitan editar cada pantalla.

## Recomendación

**Opinión/recomendación:** Figtree local con los pesos 400, 500, 600 y 700 es la
solución mínima suficiente. No se imitan familias ni elementos de Airbnb.

## Decisión del usuario

**Aceptada el 2026-08-14:** sustituir la familia de sistema por Figtree en todos
los textos controlables del cliente web, iOS y Android. Se cargan los pesos 400,
500, 600 y 700 antes de montar la interfaz.

## Consecuencias

### Positivas

- Web, iOS y Android comparten la misma familia, pesos y jerarquía visual.
- Las pantallas existentes cambian a través de `typography.family`, sin copiar
  reglas por feature.
- Los assets se empaquetan con la aplicación; la experiencia no depende de una
  solicitud a Google Fonts ni de la fuente instalada en el dispositivo.

### Negativas y deuda aceptada

- El bundle incorpora cuatro archivos TTF y el arranque espera su carga.
- Las etiquetas de navegación nativa se configuran con la API de `NativeTabs`;
  futuras capacidades nativas que no expongan familia tipográfica deberán
  documentar su excepción antes de adoptar un fallback de sistema.

## Validación

- La exportación web contiene Figtree y no deja referencias a `system-ui` en los
  textos del producto.
- `pnpm run typecheck` y la exportación Expo web finalizan correctamente.
- En iOS y Android, una development build muestra Figtree en contenido,
  cabeceras y tabs antes de validar una distribución.

## Disparadores de revisión

- La legibilidad o accesibilidad exige otra escala, peso o una familia distinta.
- La descarga del bundle o el tiempo de arranque medidos justifican reducir los
  pesos incluidos.
- Una fuente corporativa licenciada sustituye a la familia inicial.

## Documentación afectada

- `docs/adr/0054-use-pulse-design-tokens-and-shared-form-feedback.md`
- `docs/governance/DECISIONS.md`
- `docs/engineering/DESIGN_SYSTEM.md`
- `docs/engineering/DEVELOPMENT.md`
- `docs/project/LEARNING.md`
