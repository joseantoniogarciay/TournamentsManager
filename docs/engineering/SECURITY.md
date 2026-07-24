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
- datos visibles sin cuenta frente a datos de cuenta y gestión.

La autenticación demuestra identidad. La autorización para crear, ver datos no
visibles, unirse o administrar se evalúa dentro del contexto del torneo.

La identidad será propia y federada: el backend Go gestionará credenciales
locales y sesiones, y verificará identidades Apple/Google antes de resolver un
usuario interno. Email no será la clave de una identidad externa y una
coincidencia no autoriza vinculación automática. Véanse
[ADR-0010](../adr/0010-own-identity-with-federated-login.md) e
[IDENTITY.md](IDENTITY.md).

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
- prueba fresca antes de vincular identidades o cambiar canales;
- ningún intento de vinculación pendiente concede sesión o permisos;
- deep links de identidad mediante HTTPS asociado, sin tokens de sesión en URL;
- abrir un deep link mediante `GET` no consume el intento ni cambia la cuenta;
- la confirmación explícita mediante `POST` es de un solo uso y resistente a
  repetición y concurrencia;
- la URL base procede de configuración confiable y el token no se propaga por
  historial, referencias, analytics o recursos de terceros;
- una sesión no puede cambiar de propietario durante un switch de cuenta;
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

## Configuración y secretos

La gestión de configuración sigue
[ADR-0017](../adr/0017-use-env-contracts-github-environments-and-oidc.md).

- `.env`, `.env.*`, claves privadas, tokens, inventarios sensibles y estado
  Terraform no se versionan.
- `.env.example` documenta nombres y valores ficticios, no credenciales reales.
- Cada variable incorporada al cliente Expo con `EXPO_PUBLIC_` se considera
  pública.
- Los secretos de CI viven en GitHub Secrets o Environment Secrets y solo se
  exponen a jobs que los necesitan.
- Producción requiere GitHub Environment protegido antes de acceder a secretos.
- Cloud usará OIDC y credenciales temporales siempre que esté disponible.
- Los secretos no se imprimen en logs, métricas, trazas, errores ni resultados
  de comandos.

## Gates futuros

| Gate | Evidencia |
|---|---|
| Diseño | Modelo de amenazas y clasificación de datos |
| Cambio | Pruebas de autorización, validación y abuso relevante |
| Despliegue | Secretos, permisos, superficie expuesta y rollback revisados |
| Operación | Alertas, runbook e historial auditable |

No se declarará “seguro” un componente; se documentarán amenazas consideradas,
controles, evidencia y riesgo residual.
