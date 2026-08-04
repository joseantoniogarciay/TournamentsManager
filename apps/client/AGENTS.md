# Reglas de cliente universal

Estas reglas complementan el `AGENTS.md` de la raíz y son obligatorias para
cualquier cambio en `apps/client/`, incluidos cambios aparentemente pequeños de
copy, estilo o rutas.

## Preflight obligatorio

Antes de editar, revisa los ADR y documentos aplicables. Para una pantalla o
componente de interfaz, como mínimo:

- `docs/adr/0054-use-pulse-design-tokens-and-shared-form-feedback.md`
- `docs/adr/0055-use-feature-first-client-architecture-and-platform-adaptive-navigation.md`
- `docs/adr/0056-support-light-dark-system-theme-and-localized-clients.md`
- `docs/adr/0057-define-contextual-home-and-tournament-library.md` cuando se
  modifique home o navegación de torneos.
- `docs/engineering/DESIGN_SYSTEM.md`, `docs/engineering/ARCHITECTURE.md` y
  `docs/engineering/DEVELOPMENT.md`.

Si una regla aceptada no tiene aún infraestructura suficiente para cumplirla,
no la eludas en una pantalla. Construye el mínimo compartido necesario o detén
el cambio y pide dirección si eso amplía materialmente el alcance.

## Reglas no negociables de UI

- Ningún texto visible ni `accessibilityLabel`, `accessibilityHint` o mensaje de
  feedback se escribe como literal en una ruta o componente. Debe vivir en los
  catálogos planos por locale de `shared/i18n/locales/`, con claves semánticas
  estables como `common_cancel` u `home_create_tournament`. Los idiomas
  iniciales son `es`, `en`, `it` y `fr`; cualquier locale no soportado usa
  inglés.
- El selector web persistente y las preferencias de tema e idioma pertenecen a
  un provider compartido, nunca a una pantalla. La raíz propaga el tema resuelto
  también a React Navigation mediante el `ThemeProvider` de Expo Router; es
  obligatorio para que las transiciones nativas no muestren un destello del tema
  contrario. `NativeTabs` recibe además el tema resuelto en su host mediante
  `unstable_nativeProps={{ colorScheme: resolvedTheme }}`: de otro modo iOS
  hereda su propio esquema al crear o reutilizar una tab y muestra la apariencia
  equivocada. No se fuerza `Appearance.setColorScheme` para corregir este efecto.
  iOS y Android no muestran selector de idioma.
- Las pantallas usan tokens semánticos y primitivas de `shared/ui`; no introducen
  hexadecimales, medidas repetidas, fuentes remotas ni una librería de UI sin ADR
  aceptado.
- Una card mantiene su padding interno definido por la primitiva y añade siempre
  20 px de margen exterior horizontal. El layout reserva además 20 px entre
  cards hermanas; no se corrige esa separación alterando el padding de la card.
- `Screen` no añade padding horizontal: ese margen exterior pertenece a `Card`.
  Añadirlo en ambos sitios duplica la separación lateral.
- Toda pantalla deja al menos 10 px entre el área segura o el borde inferior de
  una navigation bar y su primer contenido; la implementación vigente usa 12 px
  mediante `Screen`.
- Las rutas bajo tabs reservan el espacio de la botonera en el
  `contentContainerStyle` de su contenido desplazable mediante
  `useTabContentBottomPadding`, nunca en `Screen`. El cálculo compartido suma
  `space[12]` al inset seguro en iOS y Android; en web, donde ese inset es cero
  pero la botonera permanece superpuesta, suma además `space[10]` (40 px).
- No se usa `SafeAreaView` de React Native cuando haya que sumar padding propio:
  en iOS puede ignorarlo. `Screen` combina `useSafeAreaInsets` con una `View`.
  Las rutas bajo una cabecera nativa declaran `topInset="navigation-bar"` para
  no sumar por segunda vez el inset superior que ya aporta la navegación.
- Toda ruta respeta la URL canónica y la presentación acordada: página directa e
  historial normal en web; modal y cierre seguro en móvil cuando corresponda.
- Una pantalla no llama directamente al cliente OpenAPI generado. La feature
  contiene su adaptación, hook y estado local; las reglas de negocio y
  autorización permanecen en el backend.
- Una feature no hace `fetch` directo para una operación definida en OpenAPI:
  su adaptador invoca la operación generada y le entrega `apiFetch`, el
  transporte común que resuelve URL base y futuras credenciales. Una excepción
  exige justificar por escrito que el destino no pertenece al contrato.
- La interfaz no inventa sesión, permisos, colecciones ni resultados. Debe
  expresar el estado real disponible y sus estados de carga, vacío y error.
- Botones y controles mantienen semántica accesible, un objetivo táctil mínimo
  de 44 px y no permiten envíos duplicados.
- La validación de formato de un formulario se muestra al abandonar cada campo
  y al intentar enviarlo. Cada campo conserva su propia marca de interacción;
  el intento de envío puede mostrar los errores de todos los campos, pero el
  `blur` de uno nunca marca ni muestra errores en los demás.
- `TextField` puede recibir `validationTrigger="change"` solo cuando el
  feedback inmediato ayuda materialmente a completar un requisito, como la
  longitud mínima de una contraseña. El comportamiento normal sigue siendo
  `"blur"`; el indicador complementario no aparece hasta cumplir ese requisito.
- Los formularios no muestran una acción de cancelar. La salida de la ruta se
  hace mediante el botón de atrás de la barra de navegación, que no muestra
  texto.

## Desarrollo en simulador con Expo Go

- Si se solicita Expo Go en el simulador iOS, inicia Metro con
  `pnpm --filter @tournaments-manager/client exec expo start --lan`, mantenlo
  activo y abre en Expo Go la URL `exp://` LAN que muestra Metro. No uses
  `--localhost` ni construyas manualmente una URL `exp://127.0.0.1:8081`:
  Metro puede quedar escuchando solo en IPv6 (`::1`) y Expo Go no alcanzará esa
  dirección IPv4.
- En este entorno, Metro debe ejecutarse en una terminal interactiva que se
  mantenga abierta; un `nohup ... &` lanzado desde una ejecución efímera no
  persiste de forma fiable. Como `expo-dev-client` hace que Expo seleccione una
  development build inicialmente, pulsa `s` para cambiar a **Expo Go**, espera
  a que Metro imprima `› Metro: exp://<IP-LAN>:8081` y abre exactamente esa URL
  en el simulador. No abras la URL `com.fasttourney...://expo-development-client`.
- Antes de abrir el proyecto, ejecuta
  `pnpm --filter @tournaments-manager/client exec expo install --check`. Expo Go
  solo carga módulos nativos incluidos en su SDK; usa las versiones compatibles
  que indique Expo en lugar de versiones más recientes no incluidas.
- Al añadir o actualizar una dependencia con módulo nativo, no fijes la última
  versión publicada por npm ni ejecutes `pnpm add` directamente. Ejecuta
  `pnpm --filter @tournaments-manager/client exec expo install <paquete>` y
  conserva la versión que Expo resuelva para el SDK fijado. Después, vuelve a
  ejecutar `expo install --check`. Solo se aparta de esa versión con una decisión
  explícita del usuario y un development build que compile el módulo nativo.
- Tras añadir, actualizar o configurar una dependencia con módulo nativo,
  recompila e instala una development build antes de validar o diagnosticar su
  comportamiento: los cambios de JavaScript no incorporan un pod, framework o
  config plugin a una app ya instalada. Si la consola muestra un aviso como
  `Unable to get the view config for <Módulo>`, la build activa no contiene el
  módulo; detén el ajuste de estilos o lógica, ejecuta `expo run:ios` o
  `expo run:android`, instala el binario resultante y verifica que el aviso haya
  desaparecido antes de continuar.

## Cierre obligatorio

Antes de entregar un cambio de cliente:

1. Recorre esta checklist de forma explícita contra los archivos modificados.
2. Ejecuta `pnpm run typecheck` y la exportación web:

   ```bash
   pnpm --filter @tournaments-manager/client exec expo export --platform web
   ```

3. Actualiza la documentación y `docs/project/LEARNING.md` cuando el cambio
   concrete una regla o aprendizaje reutilizable.
4. Declara cualquier parte de un ADR aceptado que siga pendiente; nunca la
   presentes como terminada por una implementación parcial.
5. Para cada operación OpenAPI tocada, confirma que la feature usa la función
   generada mediante su adaptador y `apiFetch`, sin reconstruir URL, `fetch` ni
   DTOs a mano.
