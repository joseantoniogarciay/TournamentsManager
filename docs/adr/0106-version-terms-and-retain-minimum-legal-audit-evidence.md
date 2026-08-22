# ADR-0106: Versionar los términos y conservar evidencia legal mínima

- **Estado:** Aceptado
- **Fecha:** 2026-08-22
- **Decisor:** Usuario

## Problema

FastTourney crea cuentas mediante contraseña y Google, pero no demuestra la
aceptación de condiciones contractuales. Además, los logs técnicos no tenían un
plazo de conservación explícito ni una copia separada de la evidencia legal.

## Decisión

- El alta local y la primera alta Google exigen una casilla no premarcada de
  aceptación de los Términos de uso vigentes. La política de privacidad se
  enlaza e informa; no se presenta como consentimiento necesario para ejecutar
  el contrato.
- Los Términos son un documento versionado e inmutable. El backend acepta solo
  la versión vigente y persiste, en la misma transacción que crea la cuenta, la
  versión, huella de contenido, fecha del servidor, canal de alta y una huella
  de comparación del email. No registra contraseñas, tokens, cuerpos HTTP ni IP
  en ese registro.
- Al purgar una cuenta, el registro legal se desacopla de ella y queda bloqueado
  durante cinco años para la formulación, ejercicio o defensa de reclamaciones.
  Después se elimina. La copia de seguridad conserva exactamente ese mismo
  plazo y no es una copia de la base de datos completa.
- Retención operativa inicial: logs diagnósticos JSON, 30 días; eventos de
  seguridad sin PII, 12 meses; trazas, 7 días; métricas, 24 horas en `dev`.
  Se revisará con volumen, incidentes y requisitos de producción.
- Antes de producción se configurará una exportación diaria cifrada de la
  evidencia legal hacia almacenamiento independiente, con acceso restringido y
  prueba periódica de restauración. Una copia en el mismo volumen no cuenta como
  backup.

La exportación es incremental: contiene únicamente altas o cambios desde el
último punto de control; si no hay cambios, no se crea un archivo. Se conserva
un máximo aproximado de 1.826 archivos diarios durante cinco años.

## Consecuencias

La aceptación añade un paso pequeño y homogéneo en todos los canales de alta,
pero mejora la prueba de la relación contractual. La evidencia conservada sigue
siendo dato personal seudonimizado: se bloquea, minimiza y no entra en Loki,
trazas, métricas ni analítica. El diseño de backups completos de PostgreSQL queda
fuera de esta decisión.

## Validación

- No se crea una cuenta local o Google sin aceptación de la versión vigente.
- La aceptación y la cuenta se confirman o revierten juntas.
- La purga de cuenta no elimina antes de plazo la evidencia legal y la purga de
  evidencia vencida no conserva datos identificables.
- La política de privacidad declara las categorías y plazos de conservación.
