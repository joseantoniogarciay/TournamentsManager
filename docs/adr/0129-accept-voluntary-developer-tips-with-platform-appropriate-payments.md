# ADR-0129: Aceptar propinas voluntarias al desarrollo con pagos adecuados por plataforma

- **Estado:** Aceptado
- **Fecha:** 2026-09-19
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

FastTourney puede recibir apoyo económico voluntario sin convertir sus funciones
en un producto de pago. Web, App Store y Google Play tienen mecanismos y reglas
de pago diferentes; elegir un único enlace externo para todas las plataformas
arriesgaría incumplir la política de las tiendas y añadir una infraestructura de
pagos innecesaria.

## Contexto y restricciones

- La aportación es una propina al desarrollo, no una donación a una entidad
  benéfica ni una compra deducible fiscalmente.
- No desbloquea funciones, contenido, eliminación de publicidad, monedas,
  insignias, prioridad, premios ni ninguna otra contraprestación presente o
  futura.
- Web ya está publicada; las apps nativas todavía no están distribuidas.
- Stripe Payment Links ofrece checkout alojado y se incluye en sus tarifas
  estándar sin cuota fija; las tarifas vigentes y obligaciones fiscales se
  revisarán al activarlo, no se fijan en este ADR.
- Apple permite propinas al desarrollador mediante compras in-app. Google Play
  exige normalmente su facturación para pagos digitales dentro de la app; sus
  excepciones y programas regionales cambian con frecuencia.
- Ninguna clave de Stripe, recibo de tienda, dato de tarjeta ni estado de pago
  pertenece al cliente ni se guarda en Git.

Fuentes de política consultadas el 2026-09-19: [Apple App Review
Guidelines](https://developer.apple.com/app-store/review/guidelines/uk/),
[Google Play Payments](https://support.google.com/googleplay/android-developer/answer/9858738?hl=en)
y [Stripe Payment Links](https://stripe.com/es/payments/payment-links).

## Criterios de decisión

1. mantener la propina opcional y claramente separada del producto;
2. respetar las reglas de distribución de cada plataforma;
3. evitar tratar datos de pago o crear un ledger propio sin necesidad;
4. limitar coste fijo y mantenimiento;
5. conservar una evolución posible si en el futuro se venden capacidades.

## Alternativas

### A — Stripe Payment Links en web y enlaces externos también en las apps

- **Ventajas:** un único proveedor y una implementación inicialmente pequeña.
- **Inconvenientes:** los enlaces de pago desde apps distribuidas pueden
  infringir las reglas de Apple o Google Play.
- **Coste de adopción y mantenimiento:** bajo técnicamente; alto riesgo de
  revisión de tienda.
- **Riesgos:** rechazo de la app, retirada de distribución y una experiencia
  desigual por región.

### B — Stripe en web; compra consumible nativa en cada tienda

- **Ventajas:** checkout web alojado sin datos sensibles en FastTourney y flujo
  compatible con cada tienda en móvil.
- **Inconvenientes:** tres catálogos/configuraciones de pago y comisiones de
  plataforma distintas.
- **Coste de adopción y mantenimiento:** bajo en web; medio al publicar iOS y
  Android por la validación de recibos y requisitos de las tiendas.
- **Riesgos:** cambios de política, fiscalidad y cuentas de comerciante.

### C — No aceptar propinas

- **Ventajas:** no añade obligaciones operativas, fiscales ni de revisión.
- **Inconvenientes:** no ofrece a quienes valoran el producto una forma directa
  de apoyarlo.
- **Coste de adopción y mantenimiento:** nulo.

## Comparación

A reduce código pero desplaza el riesgo a la aprobación de tiendas. C es la
opción más simple, pero no cumple el deseo de permitir apoyo voluntario. B
mantiene cada pago dentro del carril exigido por su plataforma y no modela el
dinero como parte del dominio del producto.

## Recomendación

**Opinión/recomendación:** B, activada gradualmente. Empezar por web con un
Payment Link alojado de Stripe; no introducir SDK de Stripe, endpoint, webhook
ni persistencia de pagos mientras el único resultado sea agradecimiento. Añadir
StoreKit y Google Play Billing solo al preparar la distribución real de cada
app.

## Decisión del usuario

**Aceptada el 2026-09-19:** FastTourney podrá ofrecer propinas voluntarias al
desarrollador. En web usará Stripe Payment Links de pago único; en iOS, una
compra in-app consumible mediante StoreKit; en Android, una compra única
consumible mediante Google Play Billing. La interfaz deberá declarar que no hay
contraprestación ni deducción fiscal y las apps no enlazarán a Stripe ni a otro
método externo.

La primera implementación se limita a web y a importes visibles de 2 €, 5 € y
10 €. El usuario decidirá expresamente cuándo abrir la cuenta de Stripe,
publicar el enlace y realizar la verificación legal/fiscal correspondiente.
El catálogo de Stripe usa un único nombre y descripción en inglés; FastTourney
mantiene localizada su propia interfaz. Esta decisión evita multiplicar
productos, precios y Payment Links por cada idioma mientras no exista evidencia
de que esa traducción adicional justifique su mantenimiento.

## Consecuencias

### Positivas

- la web puede empezar con una página de checkout alojada y una superficie de
  ataque reducida;
- el producto no crea privilegios, deuda de conciliación ni una economía
  interna;
- las apps tienen una ruta compatible con sus distribuidores cuando existan.

### Negativas y deuda aceptada

- las propinas tienen comisiones por transacción y pueden crear obligaciones
  contables o fiscales para el receptor;
- los catálogos de App Store y Play requieren cuentas, revisión y pruebas
  separadas;
- los precios, políticas regionales y requisitos de transparencia deben volver
  a verificarse antes de activar cada canal.

## Validación

Antes de activar web:

1. revisar tarifa, identidad del receptor, condiciones y obligaciones fiscales
   aplicables;
2. comprobar en modo de prueba que el enlace va al checkout de Stripe y que no
   expone secretos;
3. confirmar que el texto visible indica aportación voluntaria, sin recompensa
   ni afirmación fiscal;
4. comprobar que fallo, cancelación y éxito no cambian permisos ni datos de
   producto.

Antes de distribuir móvil:

1. revisar las políticas vigentes de Apple y Google Play para los territorios
   objetivo;
2. configurar cada producto consumible y probar compra, cancelación y fallo en
   sandbox;
3. confirmar que no existe enlace ni llamada a Stripe dentro de las apps;
4. actualizar las declaraciones de privacidad, fiscales y de tienda que
   correspondan.

## Disparadores de revisión

- venta de una función, suscripción, premio o beneficio asociado al pago;
- primera distribución en App Store o Google Play;
- cambio de política, comisión o país de operación relevante;
- necesidad de conciliación, reembolsos, recibos, varios destinatarios o una
  entidad jurídica distinta;
- fraude, disputa o incidente de privacidad.

## Documentación afectada

- [Producto](../project/PRODUCT.md)
- [Decisiones](../governance/DECISIONS.md)
- [Decisiones a revisar](../governance/DECISIONS_TO_REVISIT.md)
- [Aprendizaje](../project/LEARNING.md)
