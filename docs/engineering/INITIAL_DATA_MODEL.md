# Modelo inicial de datos

> Estado: diseño aceptado y ajustado por ADR-0122 y ADR-0123. No es una
> migración ni un modelo Go.

## Alcance

Este modelo cubre alta local y con Google, verificación, sesión,
publicación/lectura de torneos, sus relaciones de seguimiento o administración
delegada, fases ordenadas y los orígenes de plaza de un bracket. El primer corte
admite una única fase de liga o eliminatoria directa; liga por grupos y la
composición de varias fases quedan preparadas, pero no implementadas. No
incorpora Apple ni resultados oficiales por
cuenta: estos últimos siguen esperando la decisión de vinculación de cuentas a
equipos.

## Entidades y relaciones

```text
accounts 1 ── 0..1 local_credentials
    │ 1
    ├──── * external_identities
    ├──── * email_verification_tokens
    ├──── * sessions
    └──── * tournaments (organizer)

tournaments 1 ── * tournament_administrators ── 1 accounts
tournaments 1 ── * tournament_followers ── 1 accounts

tournaments 1 ── * tournament_teams
tournaments 1 ── * tournament_stages 1 ── * matches
```

Todos los IDs son UUIDv7. Los secretos son aleatorios opacos de al menos 128
bits; solo se almacena `token_hash = SHA-256(contexto || secreto)`. Nunca se
incluyen secretos ni hashes en DTOs, logs o métricas.

| Tabla                        | Campos esenciales                                                                                                      | Restricciones de dominio                                                                                                                                                                                                            |
| ---------------------------- | ---------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `accounts`                   | `id`, `email`, `locale`, `state`, `username`, `created_at`, `verified_at`, `expires_at`, `deletion_requested_at`        | `locale` es uno de `es`, `en`, `it` o `fr`; índice único sobre `lower(email)` para acceso; username único y en minúsculas; estado `pending_verification`, `verified` o `deletion_pending`; pendiente expira a los 7 días y una baja conserva `deletion_requested_at` durante 30 días antes de la purga. |
| `local_credentials`          | `account_id`, `password_hash`, `created_at`, `updated_at`                                                              | PK/FK uno a uno; hash Argon2id; ninguna contraseña recuperable.                                                                                                                                                                     |
| `external_identities`        | `id`, `account_id`, `provider`, `issuer`, `subject`, `created_at`                                                      | `(issuer, subject)` único; una identidad Google pertenece a una cuenta y una cuenta tiene como máximo una identidad Google. Se añade solo desde Seguridad a la cuenta autenticada y nunca se mueve entre cuentas.                   |
| `federated_login_challenges` | `id`, `provider`, `nonce_hash`, `expires_at`, `consumed_at`, `created_at`                                              | Google únicamente; nonce de 5 min, de un solo uso y sin sesión asociada.                                                                                                                                                            |
| `email_verification_tokens`  | `id`, `account_id`, `token_hash`, `expires_at`, `consumed_at`, `invalidated_at`, `created_at`                          | hash único por contexto; expira a 24 h; activo, consumido e invalidado son excluyentes; solo hay un token activo por cuenta.                                                                                                        |
| `sessions`                   | `id`, `account_id`, `token_hash`, `created_at`, `last_seen_at`, `idle_expires_at`, `absolute_expires_at`, `revoked_at` | hash único; válida solo si la cuenta está verificada, no revocada y ambos vencimientos son futuros.                                                                                                                                 |
| `tournaments`                | `id`, `organizer_account_id`, `name`, `sport`, `state`, `created_at`, `published_at`, `last_activity_at`                | raíz de identidad, equipos, permisos y visibilidad; no decide por sí misma reglas de liga o eliminatoria. |
| `tournament_stages`          | `id`, `tournament_id`, `position`, `type`, `state`                                                                     | orden único dentro del torneo; `league` y `single_elimination` son tipos distintos. |
| `tournament_administrators`  | `tournament_id`, `account_id`, `assigned_at`                                                                           | PK compuesta; el creador se conserva en `tournaments.organizer_account_id`, no se duplica. |
| `tournament_followers`       | `tournament_id`, `account_id`, `followed_at`                                                                           | guardar un torneo no concede permisos. |
| `tournament_teams`           | `id`, `tournament_id`, `name`, `position`, `withdrawn_at`                                                              | nombre y posición únicos por torneo; el orden es la siembra congelada si la fase lo requiere. |
| `matches`                    | `id`, `tournament_id`, `stage_id`, `round_number`, `sequence`, fuentes de local y visitante, `winner_team_id`, `state` | cada partido pertenece a una fase. Una plaza procede de equipo sembrado, ganadora, *bye* o, en el futuro, clasificación de grupo; un bracket puede tener equipos aún desconocidos. |
| `match_result_changes`        | `id`, `match_id`, `changed_by_account_id` opcional, marcador anterior y nuevo, `changed_at`                            | cada registro o corrección conserva la administradora y el marcador previo mientras exista su cuenta; al purgarla, la autora pasa a `NULL` y el marcador se conserva.                                                                  |
| `tournament_champions`        | `tournament_id`, `team_id`                                                                                              | conserva una única campeona de eliminatoria o todas las co-campeonas de una liga al cerrar el torneo. |

El email conserva el valor aportado para entrega; `lower(email)` es solo la clave
de comparación del producto. El locale de cuenta es una preferencia validada
para localizar emails, no un atributo de identidad o autorización. Equipos
mantienen una columna normalizada para unicidad. El username se valida como
minúsculo antes de guardar.

## Transacciones e invariantes

1. **Alta:** crea `accounts(pending_verification, locale)`, credencial, borrador
   y token de verificación en una transacción. Si ya existe el email, responde
   igual sin revelar ni modificar la cuenta existente.
2. **Verificación:** bloquea el token y cuenta, comprueba estado y vencimiento,
   consume token, fija `verified`, crea sesión y conserva el borrador; todo o
   nada.
3. **Login:** compara Argon2id; crea una sesión si la cuenta está verificada o,
   si está pendiente, invalida el token activo y crea uno nuevo sin sesión.
   El login Google valida y consume un challenge, resuelve `(issuer, subject)` y
   crea la misma clase de sesión. Una identidad nueva crea una cuenta Google
   solo si su email no pertenece ya a otra cuenta; no hay vinculación ni fusión
   basada en coincidencia de email.
4. **Publicación desde el alta:** el borrador válido crea un torneo `published`
   junto a la cuenta pendiente; los partidos se generan después, al iniciar la
   torneo tras verificar.
5. **Lectura pública:** busca por ID de torneo, exige que sea visible y devuelve solo
   proyección pública. No crea relaciones ni actualiza permisos.
6. **Baja programada:** una cuenta verificada sin torneos propios pasa a
   `deletion_pending` y conserva la fecha de solicitud. En una transacción se
   invalidan sesiones y tokens, y se eliminan seguimientos y administraciones
   delegadas. Un comando interno diario purga en lotes de hasta 100 las cuentas
   con 30 días vencidos; los `match_result_changes` conservan el resultado y
   quedan sin autora. La recuperación se define después.
7. **Inicio:** bloquea el torneo, congela una fase `league` a una o dos vueltas
   o `single_elimination` a partido único y genera sus partidos. Cada partido
   referencia esa fase; cada plaza del cuadro conserva equipo, ganadora previa
   o *bye* como fuente.
8. **Resultado:** bloquea torneo y partido, exige `in_progress` y administración;
   en liga admite el marcador, y en eliminatoria exige una ganadora, propaga su
   identidad solo a las plazas dependientes y rechaza corregirla si ya existe un
   resultado posterior. Conserva marcador anterior y nuevo, incluidos penaltis.
9. **Finalización:** bloquea el torneo, exige organizadora, estado `in_progress`
   y ningún partido pendiente. En liga persiste las posiciones 1; en
   eliminatoria, la ganadora final. Torneo y fase pasan a `completed` en la misma
   transacción.
10. **Baja de equipo:** bloquea el torneo y el equipo, exige una fase de liga y estado
   `in_progress`; marca la baja y completa todos los partidos de ese equipo con
   `3-0` para el rival. No se aplica a eliminatorias. Cada cambio conserva el marcador anterior y la autora
   en `match_result_changes`, todo en la misma transacción.

La purga es un proceso operativo explícito, idempotente y auditable por conteos,
sin registrar emails ni tokens. El `ON DELETE` y las FKs se concretarán en la
migración para que la purga de una cuenta pendiente elimine exclusivamente sus
datos temporales.

## Límites intencionales

No se persiste el borrador anónimo. Tampoco se añaden tablas genéricas de roles,
proveedores o eventos: no son necesarias para esta entrega y adelantarían
decisiones ya aplazadas.
