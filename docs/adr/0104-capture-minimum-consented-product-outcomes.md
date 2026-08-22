# ADR-0104: Capturar resultados mínimos de producto con consentimiento

- **Estado:** Aceptado
- **Fecha:** 2026-08-22
- **Decisor:** Usuario, mediante decisión explícita
- **Propietario del análisis:** Codex como mentor técnico
- **Complementa a:** ADR-0102 y ADR-0103
- **Superado por:** Ninguno

## Problema

Los eventos técnicos de red permiten seguir una petición, pero no responden por
sí solos qué resultado de producto consiguió una persona. Inferirlo de una ruta
o de un endpoint convertiría la instrumentación en una copia frágil del contrato
HTTP y confundiría reintentos o cargas automáticas con acciones deliberadas.

## Decisión del usuario

**Aceptada:** después del opt-in y solo en `development`, las features emiten
un catálogo pequeño de eventos de resultado confirmado. No importan el SDK de
PostHog: utilizan una fachada de analítica compartida. Los eventos que suceden
después de una respuesta API correcta reciben el `interaction_id` de esa
respuesta.

| Evento | Momento | Propiedades adicionales permitidas |
| --- | --- | --- |
| `registration_submitted` | Se envía un registro ya válido | Ninguna |
| `account_registered` | Se confirma el registro | Ninguna |
| `account_signed_in` | Se obtiene una sesión | `method`: `password` o `google` |
| `password_recovery_requested` | Se acepta la solicitud | Ninguna |
| `password_recovery_completed` | Se cambia la contraseña | Ninguna |
| `league_creation_submitted` | Se envía una creación ya válida | Ninguna |
| `league_created` | Se crea la liga | Ninguna |
| `league_started` | Se inicia la liga | Ninguna |
| `match_result_recorded` | Se guarda un resultado | Ninguna |
| `league_completed` | Se completa la liga | Ninguna |
| `league_administrator_assigned` | Se asigna una persona administradora | Ninguna |
| `league_administrator_removed` | Se retira una persona administradora | Ninguna |

No se registran pulsaciones genéricas, URL, IDs de ligas, equipos, partidos o
cuentas, nombres de usuario, direcciones de correo, marcadores, cuerpos,
credenciales, tokens ni errores brutos. Los estados rechazados conservan el
evento técnico de la petición; solo obtienen un evento de producto si una
recuperación futura justifica uno explícito.

## Alternativas descartadas

- **Derivar el evento desde cada endpoint:** exige mantener una taxonomía de
  red, etiqueta reintentos como acciones y no expresa el resultado aplicado.
- **Registrar cada pulsación:** eleva volumen y ruido, y mide intención aunque
  la validación local o la API no completen la operación.
- **Instrumentar todas las mutaciones desde el inicio:** dificulta revisar la
  privacidad y responder qué pregunta de producto motivó cada evento.

## Validación

1. Una operación exitosa emite su evento semántico y comparte el
   `interaction_id` de su respuesta API.
2. Un rechazo, cuerpo inválido o cancelación no emite un falso resultado de
   producto.
3. Una feature no importa `posthog-react-native`.
4. Sin consentimiento o fuera de `development`, no se emite evento semántico.

## Revisión

Antes de añadir otro evento se documentará qué pregunta operativa o de producto
responde. Se reconsidera el catálogo al medir cuota, revisar privacidad o abrir
producción con un proyecto PostHog separado.
