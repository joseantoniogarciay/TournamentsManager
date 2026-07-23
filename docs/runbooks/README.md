# Runbooks

Los runbooks son procedimientos verificables para operar o recuperar el sistema.
En Fase 0 solo existe la plantilla; crear comandos sin un sistema real sería
documentación especulativa.

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
