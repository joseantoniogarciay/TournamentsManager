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
  un provider compartido, nunca a una pantalla. iOS y Android no muestran
  selector de idioma.
- Las pantallas usan tokens semánticos y primitivas de `shared/ui`; no introducen
  hexadecimales, medidas repetidas, fuentes remotas ni una librería de UI sin ADR
  aceptado.
- Una card mantiene su padding interno definido por la primitiva y añade siempre
  20 px de margen exterior horizontal. El layout reserva además 20 px entre
  cards hermanas; no se corrige esa separación alterando el padding de la card.
- Toda ruta respeta la URL canónica y la presentación acordada: página directa e
  historial normal en web; modal y cierre seguro en móvil cuando corresponda.
- Una pantalla no llama directamente al cliente OpenAPI generado. La feature
  contiene su adaptación, hook y estado local; las reglas de negocio y
  autorización permanecen en el backend.
- La interfaz no inventa sesión, permisos, colecciones ni resultados. Debe
  expresar el estado real disponible y sus estados de carga, vacío y error.
- Botones y controles mantienen semántica accesible, un objetivo táctil mínimo
  de 44 px y no permiten envíos duplicados.

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
