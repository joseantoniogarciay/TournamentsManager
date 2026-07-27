# Modelo inicial de datos

> Estado: diseño aceptado en ADR-0045 y ajustado por ADR-0048. No es una
> migración ni un modelo Go.

## Alcance

Este modelo cubre alta local y con Google, verificación, sesión, borrador asociado
y publicación/lectura de una liga. No incorpora Apple, recuperación de
contraseña, seguimiento, administradores delegados, resultados ni cambios de
estado posteriores a `published`.

## Entidades y relaciones

```text
accounts 1 ── 0..1 local_credentials
    │ 1
    ├──── * external_identities
    ├──── * identity_link_attempts
    ├──── 0..1 league_drafts ── * draft_teams
    ├──── * email_verification_tokens
    ├──── * sessions
    └──── * leagues (organizer)

leagues 1 ── * league_teams
leagues 1 ── * matches
```

Todos los IDs son UUIDv7. Los secretos son aleatorios opacos de al menos 128
bits; solo se almacena `token_hash = SHA-256(contexto || secreto)`. Nunca se
incluyen secretos ni hashes en DTOs, logs o métricas.

| Tabla | Campos esenciales | Restricciones de dominio |
| --- | --- | --- |
| `accounts` | `id`, `email`, `state`, `username`, `created_at`, `verified_at`, `expires_at` | índice único sobre `lower(email)` para acceso; username único y en minúsculas; estado `pending_verification` o `verified`; pendiente expira a los 7 días. |
| `local_credentials` | `account_id`, `password_hash`, `created_at`, `updated_at` | PK/FK uno a uno; hash Argon2id; ninguna contraseña recuperable. |
| `external_identities` | `id`, `account_id`, `provider`, `issuer`, `subject`, `created_at` | `(issuer, subject)` único; una identidad Google pertenece a una cuenta y una cuenta tiene como máximo una identidad Google. |
| `federated_login_challenges` | `id`, `provider`, `nonce_hash`, `expires_at`, `consumed_at`, `created_at` | Google únicamente; nonce de 5 min, de un solo uso y sin sesión asociada. |
| `identity_link_attempts` | `id`, `candidate_account_id`, `provider`, `issuer`, `subject`, `token_hash`, `expires_at`, `consumed_at`, `created_at` | confirma un vínculo Google con una cuenta local mediante desafío fresco; no crea sesión hasta consumirlo. |
| `email_verification_tokens` | `id`, `account_id`, `token_hash`, `expires_at`, `consumed_at`, `invalidated_at`, `created_at` | hash único por contexto; expira a 24 h; activo, consumido e invalidado son excluyentes; solo hay un token activo por cuenta. |
| `sessions` | `id`, `account_id`, `token_hash`, `created_at`, `last_seen_at`, `idle_expires_at`, `absolute_expires_at`, `revoked_at` | hash único; válida solo si la cuenta está verificada, no revocada y ambos vencimientos son futuros. |
| `league_drafts` | `id`, `account_id`, `name`, `created_at`, `updated_at`, `expires_at` | FK a cuenta pendiente, una fila por cuenta en este incremento; se borra con la purga de la cuenta. |
| `draft_teams` | `id`, `draft_id`, `name`, `position` | nombre normalizado único por borrador; `position` conserva el orden del cliente. |
| `leagues` | `id`, `organizer_account_id`, `name`, `sport`, `format`, `state`, `created_at`, `published_at` | `sport=football`, `format=league`, reglas 1 vuelta y 3-1-0; este incremento crea `draft` y publica a `published`. |
| `league_teams` | `id`, `league_id`, `name`, `position` | nombre normalizado único por liga; se crean desde el borrador en la publicación. |
| `matches` | `id`, `league_id`, `round_number`, `sequence`, `home_team_id`, `away_team_id`, `state` | un partido por pareja no ordenada de equipos; sin marcador o fecha; estado inicial `pending`. |

El email conserva el valor aportado para entrega; `lower(email)` es solo la clave
de comparación del producto. Equipos mantienen una columna normalizada para
unicidad. El username se valida como minúsculo antes de guardar.

## Transacciones e invariantes

1. **Alta:** crea `accounts(pending_verification)`, credencial, borrador y token
   de verificación en una transacción. Si ya existe el email, responde igual sin
   revelar ni modificar la cuenta existente.
2. **Verificación:** bloquea el token y cuenta, comprueba estado y vencimiento,
   consume token, fija `verified`, crea sesión y conserva el borrador; todo o
   nada.
3. **Login:** compara Argon2id; crea una sesión si la cuenta está verificada o,
   si está pendiente, invalida el token activo y crea uno nuevo sin sesión.
   El login Google valida y consume un challenge, resuelve `(issuer, subject)` y
   crea la misma clase de sesión; una identidad nueva sigue la verificación de
   correo y una coincidencia local inicia vinculación explícita.
4. **Publicación:** exige sesión válida y propiedad. Valida nombre y al menos dos
   equipos; crea liga, equipos y todos los partidos, y elimina el borrador dentro
   de una transacción.
5. **Lectura pública:** busca por ID de liga, exige liga visible y devuelve solo
   proyección pública. No crea relaciones ni actualiza permisos.

La purga es un proceso operativo explícito, idempotente y auditable por conteos,
sin registrar emails ni tokens. El `ON DELETE` y las FKs se concretarán en la
migración para que la purga de una cuenta pendiente elimine exclusivamente sus
datos temporales.

## Límites intencionales

No se persiste el borrador anónimo. Tampoco se añaden tablas genéricas de roles,
proveedores, auditoría de resultados o eventos: no son necesarias para esta
entrega y adelantarían decisiones ya aplazadas.
