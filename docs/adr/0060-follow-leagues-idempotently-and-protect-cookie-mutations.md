# ADR-0060: Seguir ligas de forma idempotente y proteger mutaciones web con cookie

- **Estado:** Aceptado
- **Fecha:** 2026-07-28
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

La colección «Guardados» necesita una primera mutación autenticada para crear y
eliminar la relación de seguimiento. Al admitir sesión web por cookie, cualquier
mutación debe incorporar una defensa CSRF antes de exponerse.

## Contexto y restricciones

- Una persona verificada puede guardar una liga consultada mediante enlace; no
  recibe permisos ni participación deportiva (ADR-0034).
- Las ligas visibles se leen por ID público; borradores no son visibles
  (ADR-0049).
- La sesión se valida en middleware y Bearer móvil no es vulnerable al envío
  automático de cookies (ADR-0059).
- PostgreSQL mantiene sesiones opacas: una defensa debe conservar la separación
  entre autenticación, autorización y transporte.

## Criterios de decisión

1. Expresar seguir y dejar de seguir sin duplicados ni estados transitorios.
2. Repetir peticiones por reintentos de red sin producir error ni cambiar otro
   recurso.
3. Proteger las mutaciones web con cookie sin imponer un token adicional a apps
   Bearer.
4. Evitar tablas o dependencias que no aporten valor a este primer flujo.

## Alternativas de relación HTTP

### A — `PUT` y `DELETE` idempotentes sobre el seguimiento

`PUT /me/leagues/{leagueId}/follow` crea la relación si no existe y devuelve
`204`; `DELETE` la elimina si existe y también devuelve `204` si ya no existe.

- **Ventajas:** semántica REST directa; los reintentos son seguros; no exige un
  identificador de seguimiento ni estado en el cliente.
- **Inconvenientes:** el nombre de la subruta expresa una relación, no un recurso
  independiente.
- **Coste de mantenimiento:** bajo.

### B — `POST` para seguir y `DELETE` para dejar de seguir

- **Ventajas:** `POST` resulta familiar para una acción.
- **Inconvenientes:** reintentar un `POST` obliga a definir conflicto o clave de
  idempotencia sin aportar información nueva.
- **Coste de mantenimiento:** medio.

### C — Recurso de seguimiento con ID propio

- **Ventajas:** deja espacio para preferencias o alertas futuras.
- **Inconvenientes:** introduce una entidad y rutas que el producto todavía no
  necesita.
- **Coste de mantenimiento:** medio o alto.

## Alternativas de defensa CSRF para cookie

### A — Token sincronizador almacenado por sesión

El servidor emite un secreto CSRF y exige que la web lo reenvíe en un header de
cualquier mutación autenticada por cookie.

- **Ventajas:** patrón conocido y compatible con navegadores antiguos.
- **Inconvenientes:** añade emisión, almacenamiento, renovación y distribución
  de un nuevo secreto; la app web debe cargarlo antes de mutar.
- **Coste de mantenimiento:** medio.

### B — Protección estándar de Go basada en Fetch Metadata y origen

Aplicar `net/http.CrossOriginProtection` a mutaciones autenticadas por cookie;
el navegador moderno informa el contexto cross-site y el servidor bloquea la
petición antes del handler. Bearer móvil queda fuera de esta comprobación.

- **Ventajas:** no añade token, tabla ni endpoint; encaja con Go 1.26.5 y el
  cliente web moderno definido.
- **Inconvenientes:** requiere revisar proxy/orígenes autorizados y una política
  de fallback antes de admitir clientes web no modernos.
- **Coste de mantenimiento:** bajo.

### C — Confiar solo en `SameSite=Lax`

- **Ventajas:** no requiere código adicional.
- **Inconvenientes:** es una defensa complementaria, no una prueba explícita de
  origen; no satisface la regla de defensa en profundidad acordada.
- **Coste de mantenimiento:** bajo, con riesgo alto.

## Comparación

La alternativa A de relación representa exactamente una asociación única
cuenta-liga y hace seguros los reintentos. Para CSRF, B cubre el navegador
objetivo sin añadir estado duplicado; A se reserva si la compatibilidad o los
requisitos de seguridad futuros exigen un desafío explícito por sesión.

## Recomendación

**Opinión/recomendación:** aceptar A para seguir/dejar de seguir y B para CSRF.

## Decisión del usuario

**Aceptada el 2026-07-28:** alternativa A para la relación HTTP y alternativa B
para CSRF. Se usan `PUT` y `DELETE` idempotentes sobre la relación de
seguimiento; Go `CrossOriginProtection` se aplica a mutaciones autenticadas con
cookie, no a Bearer móvil.

## Consecuencias previstas

- Solo una cuenta autenticada y verificada podrá crear o borrar su propia fila
  en `league_followers`.
- Solo se podrá seguir una liga visible; una relación existente sobrevivirá si
  después la liga se cancela, para que aparezca su estado actualizado.
- Las mutaciones por Bearer no necesitarán token CSRF; las realizadas con cookie
  se bloquearán si no satisfacen la protección de origen.

## Validación prevista

- Dos `PUT` consecutivos dejan una sola relación y dos `DELETE` consecutivos
  terminan correctamente.
- Un invitado, una cuenta pendiente o una liga inexistente/no visible no crean
  una relación.
- Una mutación cross-site con cookie se rechaza antes del caso de uso; la misma
  operación mediante Bearer válido puede ejecutarse desde la app.

## Disparadores de revisión

- Compatibilidad con navegadores que no envíen Fetch Metadata.
- Necesidad de preferencias, alertas, notas o auditoría propias del seguimiento.
- Operación sensible que exija una confirmación explícita adicional.

## Documentación afectada al aceptar

- `contracts/openapi/v1/openapi.yaml`
- `docs/project/PRODUCT.md`
- `docs/engineering/API.md`
- `docs/engineering/SECURITY.md`
- `docs/governance/DECISIONS.md`
- `docs/project/LEARNING.md`
