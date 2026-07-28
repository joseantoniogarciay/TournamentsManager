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

Cuando un cambio requiera código nativo, crea una development build bajo demanda
con `pnpm --filter @tournaments-manager/client exec expo run:ios` o
`pnpm --filter @tournaments-manager/client exec expo run:android`. Los
directorios `ios/` y `android/` son generados por Expo y no se editan ni se
versionan manualmente. La guía operativa completa, incluidos los requisitos locales, está en
[Desarrollo](../../docs/engineering/DEVELOPMENT.md#cliente-expo).

## Estado actual

La navegación principal tiene tres secciones, en este orden: Inicio, Torneos y
Cuenta. Cuenta conserva su propio flujo. En iOS, la botonera usa `NativeTabs`
de Expo Router para delegar el acabado de la barra al sistema; esa API sigue
siendo experimental y se reevaluará al actualizar Expo.

Inicio muestra la orientación y las acciones disponibles sin inventar sesión ni
colecciones. Torneos y Cuenta expresan su estado actual mientras llegan los
flujos autenticados y los datos reales. El alcance funcional aceptado se
documenta en [Producto](../../docs/project/PRODUCT.md).

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
