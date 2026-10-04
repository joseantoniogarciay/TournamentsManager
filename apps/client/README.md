# Cliente universal

Cliente de TournamentsManager para web, iOS y Android, construido con Expo SDK
57, React Native y Expo Router. Es una aplicación del monorepo: instala las
dependencias y ejecuta las comprobaciones desde la raíz del repositorio.

## Arranque

```bash
pnpm install
pnpm --filter @tournaments-manager/client start
```

Con Metro activo, Expo Go permite abrir el proyecto en un dispositivo o
simulador. Para elegir una plataforma directamente:

```bash
pnpm --filter @tournaments-manager/client ios
pnpm --filter @tournaments-manager/client android
pnpm --filter @tournaments-manager/client web
```

La web local escucha siempre en `http://localhost:8082`. Si el puerto está
ocupado, libéralo antes de arrancar Expo: no se acepta desplazar el cliente a
otro puerto, porque los enlaces de correo y el origen CORS local usan `8082`.

## Entornos de aplicación

La configuración de Expo se resuelve en `app.config.ts`, con dos variantes:
desarrollo (`Fast Tourney Dev`) y producción (`Fast Tourney`). Ambas usan el
mismo código y se distinguen mediante `APP_ENV`; el icono compartido es
`public/apple-touch-icon.png` (1024 × 1024). La configuración web reutiliza
esa fuente mediante `web.favicon`; cada exportación genera `/favicon.ico` y su
declaración en el HTML para navegador y buscadores. El documento web enlaza
además el mismo fichero público mediante `apple-touch-icon` para que
Safari pueda representar la web guardada o minimizada sin recurrir al globo.

```bash
pnpm --filter @tournaments-manager/client start:dev
pnpm --filter @tournaments-manager/client ios:dev
pnpm --filter @tournaments-manager/client ios:prod
pnpm --filter @tournaments-manager/client ios:public-dev
```

Mientras CNG esté vigente no se crean ni editan targets Xcode versionados. Si
una dependencia nativa necesita existir únicamente en desarrollo o producción,
se declarará de forma condicional en `app.config.ts` o en un config plugin; solo
se reabre la alternativa de targets manuales si esa configuración no basta. Los
Las variantes iOS usan `com.fasttourney.app.dev` y `com.fasttourney.app`, por
lo que pueden instalarse a la vez. Antes de distribuirlas se verificará que el
dominio y ambos identificadores estén registrados en la cuenta de Apple.

El enlace de verificación usa la ruta HTTPS `/link/confirm` y el de recuperación
`/link/password-reset`. Producción usa `https://fasttourney.com` y desarrollo
`https://dev.fasttourney.com`; cada build recibe ese mismo origen mediante
`PUBLIC_BASE_URL` del backend y `EXPO_PUBLIC_APP_LINK_URL`. Las plantillas de los ficheros que se publicarán
en `/.well-known/` están en [infra/app-links](../../infra/app-links/README.md):
requieren el Team ID de Apple y las huellas de firma Android reales, que no se
deben adivinar ni sustituir por valores de ejemplo en producción. La plantilla
iOS incluye también `webcredentials`, necesario para asociar las credenciales
guardadas del dominio con la aplicación.

Cuando un cambio requiera código nativo, crea una development build bajo demanda
con `pnpm --filter @tournaments-manager/client exec expo run:ios` o
`pnpm --filter @tournaments-manager/client exec expo run:android`. Los
directorios `ios/` y `android/` son generados por Expo y no se editan ni se
versionan manualmente. La guía operativa completa, incluidos los requisitos locales, está en
[Desarrollo](../../docs/engineering/DEVELOPMENT.md#cliente-expo).

## Estado actual

La navegación principal tiene tres secciones, en este orden: Inicio, Torneos y
Cuenta. Cuenta conserva su propio flujo; en web, cerrar una ruta de ese flujo
vuelve explícitamente a su raíz para no depender del historial de otra tab tras
una recarga. Pulsar en web una tab que ya está activa también reemplaza su URL
por la raíz de esa tab: así una recarga en una ruta profunda no conserva una pila
inexistente. En iOS y Android, la botonera usa `NativeTabs` de Expo Router para
delegar este reinicio al sistema; esa API sigue siendo experimental y se
reevaluará al actualizar Expo.

Ajustes y Notificaciones conservan sus URLs bajo `/account` y se presentan desde
el stack raíz como modales nativos en iOS y Android. En web usan la página
estable del stack para que un cambio de tamaño no sustituya la ruta activa.

Inicio muestra la orientación y las acciones disponibles sin inventar sesión ni
colecciones. Torneos y Cuenta expresan su estado actual mientras llegan los
flujos autenticados y los datos reales. El alcance funcional aceptado se
documenta en [Producto](../../docs/project/PRODUCT.md).

Cuenta ofrece el recorrido local de acceso y una ruta de registro. Google pide
primero un challenge propio y pasa su nonce al proveedor; el ID token resultante
se entrega al backend, que establece la sesión o pide un `username` para una
cuenta nueva. No hay vínculo automático por email. Desde la rueda de ajustes se
puede elegir tema claro, oscuro o sistema; la preferencia se guarda localmente y
no requiere sesión. Las notificaciones no se solicitan todavía: siguen fuera del
alcance aceptado y el control lo comunica sin simular un permiso del sistema.

El login Google se deshabilita deliberadamente en local: no se declaran clientes
ni audiencias allí. La primera publicación web de `prod` recibe solo el ID OAuth
web, que coincide con la audiencia `GOOGLE_CLIENT_IDS` de su API. Los IDs iOS y
Android se añaden al publicar sus builds firmadas; hasta entonces el botón se
deshabilita de forma natural en esas plataformas. El export de desarrollo los inyecta desde
`infra/home/deploy-dev-web.sh`; para una prueba nativa contra `dev` usa
`pnpm --filter @tournaments-manager/client ios:public-dev` o
`pnpm --filter @tournaments-manager/client android:public-dev`. Antes de
producción se sustituye el contacto por un correo operativo del dominio propio
(por ejemplo, `support@fasttourney.com`) y se verifican dominio, web y URLs
públicas de producto y privacidad.

## Observabilidad de producto

El proyecto UE **255144** se reserva para fallos mínimos de producción, conforme
al ADR-0149. `local` y beta `development` no inicializan PostHog, aunque conserven
una clave o preferencia de uso antigua. La variable pública de producción es
`EXPO_PUBLIC_POSTHOG_PRODUCTION_API_KEY`; se ha migrado la clave existente a ella.
La analítica de uso queda apagada en todos los entornos y sus controles se retiran
de Inicio/Ajustes. No hay vistas, resultados ni correlación API; la fachada de
producto permanece sin cliente activo. No se identifican cuentas ni se habilitan
replay, flags, GeoIP, interacción automática, logs o push.

Metro añade debug IDs únicamente en producción y el plugin Expo prepara source
maps Hermes, dSYM y R8. Las claves personales CLI son solo de build, fuera de Git
y del bundle. La reconstrucción nativa y la prueba de entrega/simbolización se
harán después; esta configuración no acredita que un crash ya llegue legible.
Véase el [runbook de crashes](../../docs/runbooks/client-error-tracking.md).

## Estructura

```text
src/
  app/             Rutas de Expo Router y composición de pantallas
  api/generated/   Cliente OpenAPI generado; no se edita a mano
  shared/
    feedback/      Feedback global
    i18n/          Catálogos y resolución de idioma
    ui/             Primitivas de interfaz reutilizables
```

Las rutas no llaman directamente al cliente OpenAPI generado: cada capacidad
incorpora su adaptación y estado fuera de la pantalla. Los tokens viven en
`packages/design-tokens` y las reglas de arquitectura están en
[ARCHITECTURE.md](../../docs/engineering/ARCHITECTURE.md).

## Localización y diseño

Los textos visibles, etiquetas de accesibilidad y mensajes de feedback se
guardan en catálogos JSON planos de `src/shared/i18n/locales/`. Las claves son
semánticas y estables, por ejemplo `common_cancel` o
`home_create_tournament`. Los idiomas iniciales son español, inglés, italiano y
francés; cuando el dispositivo no está soportado se usa inglés.

Las pantallas usan tokens semánticos y primitivas de `shared/ui`. Una `Card`
mantiene su padding interno y tiene siempre 20 px de margen horizontal exterior;
los layouts reservan también 20 px entre cards hermanas. El detalle de estas
reglas y la checklist obligatoria están en [AGENTS.md](AGENTS.md) y
[DESIGN_SYSTEM.md](../../docs/engineering/DESIGN_SYSTEM.md).

## Verificación

Antes de cerrar un cambio de cliente:

```bash
pnpm run typecheck
make client-web-export
```

`make client-web-export` exporta la web temporalmente en `/tmp` y comprueba el
router y el bundling de Expo. La puerta completa del repositorio es:

```bash
make verify
```

No sustituye una revisión visual manual en las plataformas afectadas. Consulta
[AGENTS.md](AGENTS.md) antes de modificar pantallas, rutas, copy o estilos.

## Acceso Apple y Google

ADR-0147 añade Apple en web, iOS y Android. Ambos proveedores se presentan con
ASWebAuthenticationSession en iOS y Custom Tabs en Android; no se abre un navegador
externo ni se usa una WebView. Expo Router deja el retorno caliente a AuthSession
para conservar el modal y su borrador. Un inicio en frío reinicia el acceso.

El Services ID Apple público se declara en `EXPO_PUBLIC_APPLE_SERVICE_ID`; el
placeholder no activa el botón. Team ID, Key ID y clave .p8 pertenecen únicamente
al backend. La feature invoca operaciones generadas mediante `apiFetch`; el
callback Apple es una navegación del proveedor, no una petición de la feature.
Los conflictos permiten usar el acceso ya existente, la tasa pide esperar, el
transporte usa feedback de conexión y los demás fallos conservan el mensaje
común seguro. Cancelar no muestra un banner.

Preparación y gates pendientes: [runbook de acceso social](../../docs/runbooks/social-login.md).
