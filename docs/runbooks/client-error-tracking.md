# Crashes de cliente en producción

Decisión: [ADR-0149](../adr/0149-reserve-the-single-posthog-project-for-production.md).
Estado: configuración preparada; clave CLI creada en PostHog y guardada en el
archivo privado con permisos 0600; CLI instalada en el Mac de preparación;
subida real de símbolos y prueba de crashes pendientes.

## Destino y alcance

El proyecto PostHog **255144** es el existente, «Default project», en EU Cloud;
su token público se ha comparado con el cliente. El usuario lo reserva para
producción por su límite actual de un proyecto. No se crea otro ni se borra el
histórico beta. Revisar en la cuenta retención, filtrado y gasto máximo 0 €.

| Entorno | Variable | Captura nueva |
| --- | --- | --- |
| local | ninguna | SDK apagado |
| development | ninguna | SDK apagado |
| production | `EXPO_PUBLIC_POSTHOG_PRODUCTION_API_KEY` | fallos mínimos |

La clave pública existente se migra en `apps/client/.env` y
`infra/home/secrets/production-web.env`; se retira de la configuración beta.
Los archivos privados no se imprimen ni versionan. Un placeholder no activa SDK.
El entorno runtime se lee de Expo Constants, coherente con `app.config.ts`.

No hay cliente activo en la fachada de analítica de producto, vistas, resultados
ni correlación API. Se retiran sus controles de Inicio/Ajustes. Replay, flags,
GeoIP, surveys, ciclo de vida, logs y push siguen apagados. Producción permite solo
`$exception` en el filtro JS y descarta excepciones con secretos/emails reconocidos.
Comprobar aparte filtrado y contenido del canal nativo: el filtro JS no acredita
por sí solo todos los eventos nativos.

## Source maps web

El archivo público de producción declara
`FASTTOURNEY_PROD_POSTHOG_SYMBOLS_CONFIG=infra/home/secrets/posthog-production.env`.
El archivo privado de símbolos, preparado con modo 0600, fija:

```dotenv
POSTHOG_CLI_HOST=https://eu.posthog.com
POSTHOG_CLI_PROJECT_ID=255144
POSTHOG_CLI_API_KEY=
```

La clave personal «FastTourney production symbols» se ha creado con el preset
«Source map upload» y acceso solo al proyecto 255144. La interfaz confirma
«Personal API key ready» tras reautenticación. Su valor está guardado en el archivo
privado, fuera del chat/Git; se han comprobado formato, proyecto, host y permisos
sin imprimir el secreto. Nunca debe llamarse `EXPO_PUBLIC_…`.

CLI oficial instalada globalmente el 2026-10-04 en este Mac:

```sh
npm install --global @posthog/cli@0.18.7
posthog-cli --version
```

El binario disponible en PATH devuelve `posthog-cli 0.18.7`. Se ha comprobado
soporte de `--dotenv-file`, `sourcemap process`, flags de release y borrado de
maps, además de una ejecución `--dry-run` sin contactar PostHog. La CLI 0.18.7
se publicó el 2026-09-24 y su única dependencia instalada, `detect-libc` 2.1.2,
el 2025-10-05: ambas cumplen ADR-0138. No se instala `latest` ni se modifica
el lockfile de la app. Repetir esta instalación en cualquier otro equipo de
build; registrar la versión al validar la primera release.

`stage-prod-web.sh` exporta con maps y llama a `prepare-web-telemetry.mjs` antes
de completar el staging. Usa el archivo explícito, descarta credenciales PostHog
heredadas y fuerza EU. Vincula `fasttourney-web-production` al SHA. La falta de
credenciales o una subida fallida bloquea esa release; no cambia `current`.
Después elimina los mapas y `sourceMappingURL` públicos. Beta también los retira,
pero no sube símbolos ni conecta SDK. Con placeholders no hay subida.

## iOS y Android

Metro incorpora debug IDs solo en producción. El plugin
`posthog-react-native/expo` prepara maps Hermes, dSYM y R8. No añade fuentes
nativas al upload ni desactiva sandboxing Xcode desde el plugin.

En el proceso de build fijar `APP_ENV=production`, la clave pública correcta y
`POSTHOG_CLI_DOTENV_FILE` con la ruta absoluta al archivo de símbolos. Evitar
`POSTHOG_CLI_API_KEY`, `POSTHOG_CLI_PROJECT_ID`, aliases o host heredados de otra
configuración: la CLI prioriza env sobre dotenv. Nunca colocar esas claves
personales en Expo `extra` ni en el código.

Regenerar por CNG y reconstruir/instalar antes de probar. Las carpetas nativas
existentes no reciben fases nuevas por cambiar JavaScript. Si Xcode requiere
inputs de scripts, declararlos; no desactivar protecciones para ocultar el fallo.
El usuario difiere reconstrucción y prueba real. Expo Go no valida crash nativo.

## Validación posterior

1. Revisar proyecto 255144, región, claves/ID CLI, retención, privacidad y gasto.
2. Build release web/iOS/Android: provocar excepción JS y crash nativo móvil;
   relanzar, comprobar recepción y archivos/funciones/líneas legibles del release.
3. Verificar beta/local sin SDK aun con claves y preferencias previas. Producción
   sin vistas, resultados, replay, push, flags ni correlación API.
4. Revisar payloads y frames sin tokens, emails, formularios, cuerpos HTTP o IDs,
   incluido el canal nativo; comprobar que maps no se publican por HTTP.
5. Registrar evidencia y cerrar el gate ADR-0109 antes de distribuir.

## Retrospectiva

Capturar, entregar y simbolizar son tres resultados distintos. Exportar bundles
solo acredita compilación. Reservar el proyecto para prod evita mezclar nuevas
señales con beta sin añadir coste. No se han iniciado local/dev, observabilidad
ni despliegues en esta preparación.

Fuentes: [PostHog error tracking](https://posthog.com/docs/error-tracking/start-here)
y [CLI oficial](https://github.com/PostHog/posthog/tree/master/cli).
