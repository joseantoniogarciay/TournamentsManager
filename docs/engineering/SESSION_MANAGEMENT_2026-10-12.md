# Cierre de gestión de sesiones — 2026-10-12

[ADR-0164](../adr/0164-manage-active-opaque-sessions.md) registra la aceptación del
usuario. Implementación en Cuenta > Sesiones activas y contrato OpenAPI.

## Resultado

Se consultan familias propias activas, con paginación, marca actual y categorías
orientativas. Cierre individual o de todas las demás con confirmación, espera de
respuesta y revocación atómica de access/refresh. La sesión actual conserva su
logout. Refresh y revocación bloquean primero la familia. No se usan JWT ni se
almacenan IP, nombres de dispositivo o User-Agent bruto.

## Validación ejecutada

| Comprobación                                                                                                  | Resultado                                                   |
| ------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------- |
| Suite Go completa (`go -C apps/backend test ./...`)                                                           | Pasa; integración separada requiere su activación explícita |
| Suite PostgreSQL completa (`go test -race ./internal/adapters/postgres -run Integration -count=1`)            | Pasa, base exclusiva desechable PostgreSQL 18.6             |
| Sesiones: aislamiento, ID tras refresh, cierre selectivo, todas las demás y token renovado inutilizable       | Pasa                                                        |
| Paginación de 56 familias y exclusión de familia sin acceso ni refresh                                        | Pasa                                                        |
| 20 carreras de refresh contra revocación                                                                      | Pasa, sin resurrección de acceso o renovación               |
| Recorrido HTTP con repositorio real: lista, revocación, access/refresh rechazados y sesión actual preservada  | Pasa                                                        |
| HTTP: validación, autenticación, CSRF, tasa y causas técnicas seguras                                         | Pasa; sin PII ni spans por CTE/hashes                       |
| Cliente: parsing completo, transporte generado, cuerpos inválidos, estados no tratados, errores y cancelación | 4 pruebas pasan                                             |
| Regresión de refresh y navegación, junto a las nuevas pruebas                                                 | 23 pruebas pasan                                            |
| `pnpm run check`                                                                                              | Formato, ESLint, typecheck y OpenAPI pasan                  |
| `make lint-go`, `make format-check-go`, `git diff --check`                                                    | Pasan                                                       |
| Regeneración OpenAPI y sqlc                                                                                   | Misma huella de todos los archivos antes y después          |
| Exportación web final                                                                                         | Pasa; incorpora /account/sessions                           |

El sandbox impedía algunos sockets y el uso de la caché habitual; se ejecutaron
las pruebas necesarias con acceso revisado. La caché aislada creada para esta
tarea agotó el espacio al compilar herramientas y race: se retiró únicamente esa
caché y se reutilizó la habitual. Se recreó la base desechable y la suite completa
posterior terminó correctamente. El contenedor de pruebas se apagó y eliminó.
Local/dev permanecen apagados y no se inició observabilidad.

## Checklist de cliente

- Preflight de ADR-0054/0055/0056 y handbook de diseño/arquitectura/desarrollo.
- Ruta normal del stack de Cuenta; reutiliza navegación y salida existentes.
- Textos y categorías en es/en/fr/it, sin mensajes internos ni literales de copy
  en componentes. Fechas desde datos validados, formateadas por idioma del sistema.
- Tokens, Card, Screen, Button, LoadingTransition, RequestErrorCard y confirmación
  compartidos; márgenes exteriores, separación y reserva de tab bar existentes.
- Feature con adaptador y hook; funciones generadas con authenticatedApiFetch,
  que delega en apiFetch. No reconstruye URLs ni usa fetch directo.
- Carga inicial, vacío y error con reintento; fallos parciales conservan contenido
  y usan banner seguro. Sin duplicación de feedback inicial ni banner de éxito.
- Mutaciones serializadas; controles deshabilitados durante peticiones, datos
  eliminados tras 204 y cancelación al abandonar la ruta sin aviso.
- Autorización y pertenencia revalidadas en backend, nunca deducidas por hints.

## Límites y siguiente operación

No se ha publicado ni aplicado la migración a bases de local/dev/producción.
Aplicar **00023_add_session_client_categories.sql antes de arrancar la API nueva**.
No hay migración de secretos ni de identidad. Las sesiones previas conservan
unknown hasta volver a usarse. La comprobación visual en navegador y en binarios
nativos queda fuera de la validación ejecutada; no se declara como realizada.

La retrospectiva está en [LEARNING.md](../project/LEARNING.md): familia estable,
refresh activo con access vencido, pistas de cliente falsificables y orden de locks.
