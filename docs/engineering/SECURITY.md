# Seguridad

> Estado: baseline de proceso. Los controles concretos dependen del producto.

## Principio

La seguridad se diseña con el flujo, no se añade al final. Toda decisión debe
considerar confidencialidad, integridad, disponibilidad, abuso y recuperación.

## Proceso mínimo

Antes del primer vertical slice:

1. identificar activos, actores y límites de confianza;
2. clasificar datos;
3. modelar amenazas del flujo;
4. decidir identidad, autenticación y autorización;
5. definir gestión de secretos;
6. seleccionar controles y pruebas;
7. documentar riesgos aceptados por el usuario.

## Límites de confianza iniciales

- navegador web;
- aplicación mobile y almacenamiento del dispositivo;
- API pública;
- proveedor o módulo de identidad;
- canal de email para verificación y recuperación;
- datos públicos del torneo frente a datos de cuenta y gestión.

La autenticación demuestra identidad. La autorización para crear, ver datos no
públicos, unirse o administrar se evalúa dentro del contexto del torneo.

La recuperación de contraseña no revelará si existe una cuenta y nunca devolverá
la contraseña anterior.

## Reglas iniciales

- mínimo privilegio para personas y workloads;
- denegar por defecto;
- secretos fuera del repositorio, logs e imágenes;
- dependencias fijadas y revisables;
- entradas no confiables validadas en el límite;
- errores externos sin detalles sensibles;
- acciones sensibles auditables;
- cifrado y retención definidos según el tipo de dato;
- recuperación y respuesta a incidentes ensayables.

## Repositorio público y CI/CD

- Código, handbook y definiciones declarativas pueden ser públicos.
- `.env`, claves, tokens, estados Terraform e inventarios sensibles no se
  versionan.
- Variables incorporadas a bundles web/mobile se tratan como públicas.
- Producción usa secrets por environment y aprobación antes del despliegue.
- Los workflows reciben permisos mínimos y no ejecutan contribuciones no
  confiables con secretos.
- No se ejecutan runners self-hosted del repositorio público en el VPS ni en una
  red con acceso privilegiado.
- Cloud usa OIDC y credenciales temporales cuando esté disponible.
- El VPS usa una identidad de despliegue dedicada y limitada.

La decisión completa está en
[ADR-0006](../adr/0006-public-github-repository-security-boundary.md).

## Gates futuros

| Gate | Evidencia |
|---|---|
| Diseño | Modelo de amenazas y clasificación de datos |
| Cambio | Pruebas de autorización, validación y abuso relevante |
| Despliegue | Secretos, permisos, superficie expuesta y rollback revisados |
| Operación | Alertas, runbook e historial auditable |

No se declarará “seguro” un componente; se documentarán amenazas consideradas,
controles, evidencia y riesgo residual.
