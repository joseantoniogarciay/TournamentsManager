# Retención de registros

> Estado: decisión aceptada en ADR-0106. Última revisión: 2026-08-22.

| Categoría | Contenido permitido | Plazo | Destino |
| --- | --- | ---: | --- |
| Diagnóstico | nivel, ruta, estado, latencia, IDs técnicos de correlación y causa cerrada | 30 días | Loki |
| Seguridad | evento cerrado de autenticación o límite, sin email, IP, token ni cuerpo | 12 meses | registro de seguridad pendiente de backend de producción |
| Trazas | atributos técnicos sin PII ni secretos | 7 días | Tempo |
| Métricas `dev` | agregados sin identificadores | 24 horas | Prometheus |
| Evidencia de aceptación | versión, huella de términos, momento, canal y huella de email | cuenta activa + 5 años bloqueada tras baja | PostgreSQL y backup legal independiente |

Los plazos son máximos; un incidente o requerimiento legal puede imponer un
bloqueo documentado. Las copias de la evidencia legal se cifran, se limitan a
personal autorizado y se purgan con una retención que nunca sea menor a la del
registro original. El backup no sustituye el futuro plan de recuperación de
PostgreSQL.

## Copia legal incremental

La purga de cuentas y evidencia vencida se ejecuta diariamente a las 03:15. El
backup se ejecuta a las 03:30, después de esa purga. Cada archivo
está cifrado antes de llegar a `iCloud Drive/FastTourney/legal-audit-backups` y
solo incluye filas nuevas o modificadas desde el último punto de control. Si no
hay cambios, no se sube un archivo vacío. La clave local no se sincroniza con
iCloud ni se incorpora a Git.

El estado local del backup relaciona únicamente identificadores técnicos con el
archivo y su fecha de retención. Cuando todos los registros de un archivo han
vencido, el mismo job diario elimina ese archivo cifrado. No guarda emails,
contraseñas, tokens ni contenido legal en claro.

La restauración se ejecuta de forma explícita sobre un destino aislado mediante
`legal-audit-restore`; no se descifra en la carpeta sincronizada ni se mezcla
con datos de la aplicación.
