# Runbooks

Los runbooks son procedimientos verificables para operar o recuperar el sistema.
El primer procedimiento ejecutable se incorpora al decidir e implementar el
entorno local; su estado de prueba se declara dentro del propio runbook.

## Runbooks disponibles

- [PostgreSQL local con Docker Compose](local-postgresql.md)
- [Backup y restauración PostgreSQL de dev](postgresql-backup-dev.md)
- [Diagnóstico del refresh de sesión](session-refresh-observability.md)

## Requisitos

- síntoma y alcance;
- prerequisitos y permisos;
- diagnóstico seguro antes de mutar;
- mitigación y recuperación diferenciadas;
- verificación desde el punto de vista del usuario;
- rollback o escalado;
- riesgos de cada acción;
- fecha de última prueba.

Usa [template.md](template.md). Un runbook no se considera válido hasta haber sido
ejecutado en un entorno apropiado.
