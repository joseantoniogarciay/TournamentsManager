# ADR-0060: Usar PostHog de forma diferida para observabilidad de producto del cliente

- **Estado:** Aceptado
- **Fecha:** 2026-07-29
- **Decisor:** Usuario, mediante confirmación explícita
- **Propietario del análisis:** Asistente como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

Cuando el cliente llegue a usuarios reales, el equipo necesita diagnosticar
errores y crashes desde el comportamiento que los precede, medir la adopción de
flujos y poder desactivar cambios de interfaz con seguridad. La observabilidad
del backend ya usa una dirección abierta basada en OpenTelemetry, pero no cubre
por sí sola la experiencia en web, iOS y Android ni ofrece session replay.

No existe todavía una beta distribuida ni evidencia de volumen, crashes o una
pregunta de producto que justifique integrar un SDK ahora.

## Contexto y restricciones

- ADR-0020 acepta OpenTelemetry, Prometheus, Grafana, Loki y Tempo para el
  diagnóstico técnico del backend; no adopta un SaaS ni instrumentación móvil.
- ADR-0015 acepta Expo, Expo Router y CNG. Los módulos nativos requieren una
  development build o una build distribuida; no se configura código nativo en
  Expo Go.
- ADR-0017 exige que valores expuestos al cliente se traten como públicos y
  prohíbe secretos, tokens y PII en logs, errores, métricas y trazas.
- El backend sigue siendo la autoridad de identidad, autorización y reglas de
  negocio. Un flag remoto no puede conceder permisos ni modificar reglas,
  precios, resultados o datos protegidos.
- Session replay, analytics y crash reporting pueden tratar datos personales o
  sensibles. Antes de activarlos se revisarán base jurídica, información al
  usuario, consentimiento cuando aplique, DPA, región, retención, enmascarado y
  exclusiones de pantalla.
- Esta decisión no autoriza crear una cuenta, instalar dependencias, enviar
  datos, activar replay ni cambiar el cliente antes del disparador definido.

## Criterios de decisión

1. correlacionar errores del cliente con la secuencia de interacción previa y
   con una versión o actualización concreta;
2. cubrir web, iOS y Android sin romper CNG;
3. conservar OpenTelemetry como fuente de diagnóstico técnico del backend;
4. evitar datos sensibles, costes no acotados y acoplar lógica de negocio a un
   proveedor;
5. mantener coste inicial cero y una salida razonable;
6. introducir la menor cantidad de servicios para analytics, replay, errores y
   flags visuales cuando exista una necesidad real.

## Alternativas

### Alternativa A — PostHog Cloud para el cliente, con OpenTelemetry en backend

Adoptar PostHog Cloud en región UE cuando se distribuya la primera beta. Usarlo
para eventos de producto, session replay restringido, error tracking y flags de
interfaz. Mantener las señales técnicas del backend en OpenTelemetry. La
correlación entre ambas capas se realiza con un `request_id` opaco, no con
payloads ni trazas completas enviados a PostHog.

- **Ventajas:** una sola herramienta para experiencia, eventos, replay, errores
  y flags; guía e integración de Expo/EAS; soporte para web y apps; cuotas
  gratuitas separadas y límite de gasto configurable.
- **Inconvenientes:** dependencia de SaaS y tratamiento externo de telemetría;
  el replay y crashes nativos requieren builds no Expo Go; la correlación con
  Tempo/Grafana no es una navegación automática.
- **Coste de adopción:** bajo o medio, diferido hasta una beta distribuida.
- **Coste de mantenimiento:** bajo mientras el esquema de eventos sea pequeño;
  requiere revisar privacidad, subir source maps/símbolos y retirar flags
  temporales.
- **Riesgos:** sobrecaptura, replay de datos sensibles o crecimiento de coste.
  Se mitigan con enmascarado por defecto, exclusiones, muestreo, catálogo de
  eventos, región UE y límite de gasto cero.

### Alternativa B — Firebase: Analytics, Crashlytics y Remote Config

Usar React Native Firebase para analítica, crash reporting, rollouts y
configuración remota.

- **Ventajas:** productos integrados y maduros para crashes, analítica y
  rollouts; Crashlytics aporta breadcrumbs y métricas de estabilidad.
- **Inconvenientes:** no aporta session replay visual; integra varios SDKs
  nativos y productos Google; Remote Config no es seguro para secretos ni
  decisiones de autorización; añade una dependencia que no cubre el diagnóstico
  técnico OpenTelemetry del backend.
- **Coste de adopción:** medio; requiere config plugins y development builds.
- **Coste de mantenimiento:** medio; hay que gobernar varios productos, eventos
  y su privacidad.
- **Riesgos:** ampliar Firebase más allá del alcance por comodidad y duplicar la
  autoridad del backend.

### Alternativa C — Sentry para errores y replay; otro producto para analytics y flags

Usar Sentry para crashes, errores y session replay, y elegir posteriormente una
herramienta distinta para analítica y feature flags.

- **Ventajas:** excelente foco en diagnóstico de errores y buena integración con
  Expo/EAS.
- **Inconvenientes:** dos o más proveedores para cubrir todos los casos;
  correlación y gobernanza de datos divididas.
- **Coste de adopción:** medio.
- **Coste de mantenimiento:** medio por los contratos y SDKs adicionales.
- **Riesgos:** complejidad prematura antes de saber si se necesitan analytics o
  flags.

### Alternativa D — No incorporar observabilidad de producto del cliente todavía

Mantener únicamente logs y observabilidad futura del backend hasta contar con
un incidente o una beta distribuida.

- **Ventajas:** coste y tratamiento de datos cero; no añade dependencias.
- **Inconvenientes:** un crash de cliente no tendrá stack trace simbolizado ni
  una secuencia de interacción que permita reproducirlo.
- **Coste de adopción y mantenimiento:** nulo.
- **Riesgos:** mayor tiempo de diagnóstico y decisiones de producto sin
  evidencia al comenzar la beta.

## Comparación

PostHog es la única alternativa evaluada que reúne replay, analytics, errores y
flags de interfaz con una integración específica de Expo. Firebase cubre bien
crashes y rollouts, pero no ofrece replay visual y separaría la experiencia del
diagnóstico OTel. Sentry es preferible si el único objetivo fuese crash
reporting, pero exige otro servicio para analytics y flags. No cambiar conserva
la simplicidad mientras no se distribuya el cliente, pero deja descubierto el
primer incidente de usuario real.

## Recomendación

**Opinión/recomendación:** alternativa A, pero activada solo al abrir una beta
distribuida. Es la solución mínima suficiente para conectar comportamiento y
errores del cliente sin sustituir OpenTelemetry ni delegar decisiones de negocio
en un proveedor externo.

## Decisión del usuario

**Aceptada:** adoptar de forma diferida PostHog Cloud en región UE para la
observabilidad de producto del cliente. Hasta el disparador no se crea cuenta ni
se modifica el código. Al implementarlo, el gasto mensual máximo se configura
en 0 €; al agotarse una cuota gratuita se acepta perder señal adicional antes
que incurrir en coste.

PostHog se limita a eventos de producto, session replay restringido, error
tracking y flags de interfaz. El backend conserva OpenTelemetry como fuente de
verdad para logs, métricas y trazas. La correlación se hará mediante un
`request_id` opaco y de alta cardinalidad solo en logs y trazas, nunca como
etiqueta de métrica ni como sustituto de identidad.

## Consecuencias

### Positivas

- Los errores del cliente se podrán relacionar con una interacción previa,
  plataforma, release y actualización de Expo.
- Analytics, replay y flags visuales comparten un proveedor y un catálogo de
  eventos pequeño.
- El backend no queda acoplado al SDK ni al modelo de datos de PostHog.
- El límite de gasto elimina el riesgo de coste antes de validar el valor.

### Negativas y deuda aceptada

- PostHog Cloud es una dependencia externa para telemetría de cliente.
- Las señales de usuario y las trazas técnicas estarán en sistemas distintos;
  el `request_id` permite cruzarlas de forma deliberada, no automática.
- La integración nativa se reevaluará junto con la transición futura de
  CocoaPods a Swift Package Manager.
- Si el volumen supera las cuotas gratuitas, se perderán replays, eventos o
  evaluaciones adicionales hasta el siguiente ciclo de facturación.

## Validación

Al abrir la primera beta distribuida se demostrará que:

1. PostHog Cloud está configurado en región UE, sin gasto permitido y con alertas
   de uso antes de cada cuota.
2. El catálogo inicial contiene únicamente eventos necesarios, nombres
   semánticos y propiedades revisadas; no hay PII, credenciales, tokens, cuerpos
   HTTP ni datos de formularios.
3. Session replay está desactivado en desarrollo, usa muestreo bajo y enmascara
   texto; login, registro, recuperación y pantallas sensibles están excluidos.
4. Los errores JavaScript y los crashes nativos de prueba se simbolizan contra
   la release o EAS Update correcta.
5. Un fallo controlado puede seguirse desde el replay/error hasta el
   `request_id` y la traza/log del backend en Grafana, sin propagar PII.
6. Un flag de interfaz puede desactivar una vista experimental, pero una llamada
   de backend sigue aplicando autorización y reglas aunque el cliente manipule
   el flag.

## Disparadores de revisión

- La primera beta distribuida o el primer incidente de cliente abre la
  implementación.
- El coste, las cuotas o la privacidad impiden conservar el límite de gasto
  cero.
- Una revisión legal o de privacidad exige otra región, consentimiento,
  autoalojamiento o no usar replay.
- Expo, React Native o Swift Package Manager vuelven incompatible el SDK o su
  config plugin.
- Se necesita correlación de trazas extremo a extremo madura entre cliente y
  backend; se reevalúa OpenTelemetry en React Native.
- Sentry, Firebase u otra alternativa ofrece mejor cobertura con menor coste o
  riesgo demostrado.

## Documentación afectada

- [README.md](../../README.md)
- [DECISIONS.md](../governance/DECISIONS.md)
- [DECISIONS_TO_REVISIT.md](../governance/DECISIONS_TO_REVISIT.md)
- [LEARNING.md](../project/LEARNING.md)
- [CHANGELOG.md](../../CHANGELOG.md)

## Fuentes técnicas

- [PostHog: precios](https://posthog.com/pricing)
- [Expo: usar PostHog](https://docs.expo.dev/guides/using-posthog/)
- [Expo: usar Firebase](https://docs.expo.dev/guides/using-firebase/)
- [Expo: usar Sentry](https://docs.expo.dev/guides/using-sentry/)
- [Firebase Remote Config: seguridad](https://firebase.google.com/docs/remote-config)
- [OpenTelemetry: propagación de contexto](https://opentelemetry.io/docs/concepts/context-propagation/)
