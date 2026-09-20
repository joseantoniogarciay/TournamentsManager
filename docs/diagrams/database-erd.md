# Mapa entidad-relación de PostgreSQL

> Propósito: mostrar las tablas y relaciones del esquema efectivo.
>
> Alcance: `initial_schema.sql` y migraciones hasta `00011`.
>
> Última revisión: 2026-09-19.

Este mapa es una vista explicativa. La fuente de verdad ejecutable continúa
siendo el [esquema y sus migraciones](../../apps/backend/db/); ante cualquier
divergencia prevalece el SQL aplicado en orden. Se divide en dos vistas porque
un único diagrama con las 21 tablas ocultaría las cardinalidades.

## Identidad y acceso

```mermaid
%%{init: {"theme":"base","themeVariables":{"darkMode":true,"background":"#0d1117","primaryColor":"#1f2937","primaryTextColor":"#f8fafc","primaryBorderColor":"#94a3b8","secondaryColor":"#111827","secondaryTextColor":"#f8fafc","secondaryBorderColor":"#94a3b8","tertiaryColor":"#334155","tertiaryTextColor":"#f8fafc","tertiaryBorderColor":"#94a3b8","lineColor":"#94a3b8","textColor":"#f8fafc","edgeLabelBackground":"#111827"}}}%%
erDiagram
    direction LR

    ACCOUNTS ||--o| LOCAL_CREDENTIALS : "posee CASCADE"
    ACCOUNTS ||--o{ EMAIL_VERIFICATION_TOKENS : "verifica CASCADE"
    ACCOUNTS ||--o{ PASSWORD_RESET_TOKENS : "recupera CASCADE"
    ACCOUNTS ||--o{ SESSIONS : "abre CASCADE"
    SESSIONS ||--o{ SESSION_REFRESH_TOKENS : "rota CASCADE"
    ACCOUNTS ||--o{ REAUTHENTICATION_TICKETS : "solicita CASCADE"
    SESSIONS ||--o{ REAUTHENTICATION_TICKETS : "autoriza CASCADE"
    ACCOUNTS ||--o{ EXTERNAL_IDENTITIES : "vincula CASCADE"
    ACCOUNTS o|..o{ LEGAL_ACCOUNT_ACCEPTANCES : "acepta SET NULL"
    ACCOUNTS ||--o{ PRODUCT_SUGGESTIONS : "envia CASCADE"

    ACCOUNTS {
        uuid id PK
        text email UK
        text username UK
        text state
        text locale
    }
    LOCAL_CREDENTIALS {
        uuid account_id PK, FK
        text password_hash
    }
    EMAIL_VERIFICATION_TOKENS {
        uuid id PK
        uuid account_id FK
        bytea token_hash UK
        timestamptz expires_at
    }
    PASSWORD_RESET_TOKENS {
        uuid id PK
        uuid account_id FK
        bytea token_hash UK
        timestamptz expires_at
    }
    SESSIONS {
        uuid id PK
        uuid account_id FK
        bytea token_hash UK
        timestamptz idle_expires_at
        timestamptz absolute_expires_at
    }
    SESSION_REFRESH_TOKENS {
        uuid id PK
        uuid session_id FK
        bytea token_hash UK
        timestamptz expires_at
    }
    REAUTHENTICATION_TICKETS {
        uuid id PK
        uuid account_id FK
        uuid session_id FK
        bytea token_hash UK
        timestamptz expires_at
    }
    EXTERNAL_IDENTITIES {
        uuid id PK
        uuid account_id FK
        text provider
        text issuer
        text subject
    }
    LEGAL_ACCOUNT_ACCEPTANCES {
        uuid id PK
        uuid account_id FK "nullable"
        bytea email_hash
        text terms_version
        timestamptz retention_until "nullable"
    }
    FEDERATED_LOGIN_CHALLENGES {
        uuid id PK
        text provider
        bytea nonce_hash UK
        timestamptz expires_at
    }
    GOOGLE_RISC_EVENTS {
        text id PK
        timestamptz received_at
        timestamptz expires_at
    }
    PRODUCT_SUGGESTIONS {
        uuid id PK
        uuid account_id FK
        text body
        timestamptz created_at
    }
```

`federated_login_challenges` y `google_risc_events` no tienen clave foránea: el
primero protege un intento federado previo a conocer la cuenta y el segundo
deduplica eventos RISC por su identificador externo.

## Torneos y competición

```mermaid
%%{init: {"theme":"base","themeVariables":{"darkMode":true,"background":"#0d1117","primaryColor":"#1f2937","primaryTextColor":"#f8fafc","primaryBorderColor":"#94a3b8","secondaryColor":"#111827","secondaryTextColor":"#f8fafc","secondaryBorderColor":"#94a3b8","tertiaryColor":"#334155","tertiaryTextColor":"#f8fafc","tertiaryBorderColor":"#94a3b8","lineColor":"#94a3b8","textColor":"#f8fafc","edgeLabelBackground":"#111827"}}}%%
erDiagram
    direction LR

    ACCOUNTS ||--o{ TOURNAMENTS : "organiza RESTRICT"
    ACCOUNTS ||--o{ TOURNAMENT_ADMINISTRATORS : "administra CASCADE"
    TOURNAMENTS ||--o{ TOURNAMENT_ADMINISTRATORS : "delega CASCADE"
    ACCOUNTS ||--o{ TOURNAMENT_FOLLOWERS : "sigue CASCADE"
    TOURNAMENTS ||--o{ TOURNAMENT_FOLLOWERS : "es seguido CASCADE"
    ACCOUNTS ||--o{ ACCOUNT_NOTIFICATIONS : "recibe CASCADE"
    TOURNAMENTS ||--o{ ACCOUNT_NOTIFICATIONS : "origina CASCADE"
    TOURNAMENTS ||--o{ TOURNAMENT_TEAMS : "inscribe CASCADE"
    TOURNAMENTS ||--o{ TOURNAMENT_STAGES : "ordena CASCADE"
    TOURNAMENTS ||--o{ MATCHES : "contiene CASCADE"
    TOURNAMENT_STAGES ||--o{ TOURNAMENT_STAGE_TEAMS : "compone CASCADE"
    TOURNAMENT_TEAMS ||--o{ TOURNAMENT_STAGE_TEAMS : "participa RESTRICT"
    TOURNAMENT_STAGES ||--o{ MATCHES : "programa CASCADE"
    TOURNAMENT_TEAMS o|..o{ MATCHES : "juega local RESTRICT"
    TOURNAMENT_TEAMS o|..o{ MATCHES : "juega visitante RESTRICT"
    TOURNAMENT_TEAMS o|..o{ MATCHES : "gana RESTRICT"
    MATCHES o|..o{ MATCHES : "alimenta plaza local"
    MATCHES o|..o{ MATCHES : "alimenta plaza visitante"
    MATCHES ||--o{ MATCH_RESULT_CHANGES : "audita CASCADE"
    ACCOUNTS o|..o{ MATCH_RESULT_CHANGES : "corrige SET NULL"
    TOURNAMENTS ||--o{ TOURNAMENT_CHAMPIONS : "proclama CASCADE"
    TOURNAMENT_TEAMS ||--o| TOURNAMENT_CHAMPIONS : "obtiene RESTRICT"

    ACCOUNTS {
        uuid id PK
        text username UK
    }
    TOURNAMENTS {
        uuid id PK
        uuid organizer_account_id FK
        uuid source_draft_id "nullable"
        text sport
        text format
        text state
    }
    TOURNAMENT_ADMINISTRATORS {
        uuid tournament_id PK, FK
        uuid account_id PK, FK
        timestamptz assigned_at
    }
    TOURNAMENT_FOLLOWERS {
        uuid tournament_id PK, FK
        uuid account_id PK, FK
        timestamptz followed_at
    }
    ACCOUNT_NOTIFICATIONS {
        uuid id PK
        uuid account_id FK
        uuid tournament_id FK
        text kind
        timestamptz read_at "nullable"
    }
    TOURNAMENT_TEAMS {
        uuid id PK
        uuid tournament_id FK
        text name
        int position
        timestamptz withdrawn_at "nullable"
    }
    TOURNAMENT_STAGES {
        uuid id PK
        uuid tournament_id FK
        int position
        text type
        text state
        text league_structure "nullable"
        int qualifier_count "nullable"
        int group_count "nullable"
        int qualifiers_per_group "nullable"
    }
    TOURNAMENT_STAGE_TEAMS {
        uuid stage_id PK, FK
        uuid team_id PK, FK
        int seed_position
        int group_number "nullable"
    }
    MATCHES {
        uuid id PK
        uuid tournament_id FK
        uuid stage_id FK
        int group_number "nullable"
        uuid home_team_id FK "nullable"
        uuid away_team_id FK "nullable"
        uuid winner_team_id FK "nullable"
        uuid home_source_match_id FK "nullable"
        uuid away_source_match_id FK "nullable"
        text state
    }
    MATCH_RESULT_CHANGES {
        uuid id PK
        uuid match_id FK
        uuid changed_by_account_id FK "nullable"
        text result_type
        timestamptz changed_at
    }
    TOURNAMENT_CHAMPIONS {
        uuid tournament_id PK, FK
        uuid team_id PK, FK
    }
```

Las tres relaciones entre `tournament_teams` y `matches` representan los roles
local, visitante y ganador. Las dos autorrelaciones de `matches` permiten que la
ganadora de un partido alimente la plaza local o visitante de uno posterior.

## Cómo leer el mapa

- `PK`, `FK` y `UK` identifican clave primaria, foránea y única.
- `||` significa exactamente uno; `o|`, cero o uno; `o{`, cero o varios.
- `CASCADE`, `RESTRICT` y `SET NULL` resumen el efecto de borrar la fila padre.
- Las claves únicas compuestas, como `(issuer, subject)` o
  `(organizer_account_id, source_draft_id)`, se explican en el modelo de datos
  pero no se marcan como unicidad individual. Sus restricciones exactas
  permanecen en el SQL.
