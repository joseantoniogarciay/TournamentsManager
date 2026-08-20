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

ADR-0044 fija Argon2id para contraseñas locales y sesiones opacas revocables;
ADR-0062 añade access y refresh opacos rotatorios. En web ambos secretos viven
solo en cookies persistentes `HttpOnly`, y en móvil se entregan como Bearer al
almacenamiento seguro. Mailpit se usa solo para capturar email en local; el
desarrollo público usa Resend SMTP con STARTTLS y una clave de solo envío
(ADR-0093). JWT no se introduce sin un disparador de arquitectura distribuida.

ADR-0059 centraliza la validación de cookie o Bearer en middleware para rutas
protegidas. La autenticación no sustituye autorización por recurso. Antes de la
primera mutación autenticada desde web se añadirá una defensa CSRF separada; una
lectura `GET` y el transporte Bearer móvil no la requieren.

## Reglas iniciales

- mínimo privilegio para personas y workloads;
- una API pública usa una identidad PostgreSQL de ejecución sin permisos de
  esquema; migraciones y propiedad de objetos usan identidades separadas según
  ADR-0097;
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
- para el registro local, el cliente inicia automáticamente el `POST` de un solo
  uso después de retirar el token de la URL; véase ADR-0061;
- la URL base procede de configuración confiable y el token no se propaga por
  historial, referencias, analytics o recursos de terceros;
- al completar un registro desde un cliente con sesión, la credencial presentada
  se revoca antes de entregar la nueva, conforme a ADR-0061;
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
- CI usa `pull_request`, no `pull_request_target`, y solo permisos de lectura
  hasta que un job de despliegue aprobado necesite otra capacidad.
- No se ejecutan runners self-hosted del repositorio público en el VPS ni en una
  red con acceso privilegiado.
- Cloud usa OIDC y credenciales temporales cuando esté disponible.
- El VPS usa una identidad de despliegue dedicada y limitada.

Para AWS, ADR-0026 separa `management`, `nonprod` y `prod` mediante AWS
Organizations. El acceso humano usa IAM Identity Center con MFA y roles
temporales; los usuarios IAM y access keys persistentes no son el mecanismo
ordinario de acceso. El usuario root se reserva para acciones excepcionales.

ADR-0027 mantiene estado Terraform local únicamente sin infraestructura AWS
real. ADR-0028 fija HCP Terraform Free como backend remoto inicial, con
ejecución local y sin auto-apply. Antes de un `apply` cloud se verificará su
locking, historial recuperable y el secreto de acceso; GitHub no almacenará
estados, porque pueden contener valores sensibles y no aporta locking de
Terraform.

Para la publicación futura, ADR-0029 fija un ALB como único punto de entrada
público. El security group de la API solo aceptará el puerto de aplicación desde
el security group del ALB; PostgreSQL no tendrá IP pública y solo aceptará la
conexión necesaria desde la API. No se usará NAT inicialmente, decisión que se
revisará si aparece una necesidad de egress privado o cumplimiento.

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
- CORS se configura mediante una allowlist exacta de orígenes públicos; no se
  usa `*` porque la web podrá enviar cookies de sesión. La allowlist no sustituye
  la protección CSRF de mutaciones autenticadas por cookie. Para esas
  mutaciones, los mismos orígenes validados se registran explícitamente como
  confiables en la protección CSRF; un origen ajeno sigue rechazado.
- Los límites de abuso se evalúan en la API. Cuando una API se publica a través
  de Caddy y Docker, Caddy copia `CF-Connecting-IP` en `X-Client-IP` y la API
  acepta esa cabecera solo si la conexión inmediata procede de
  `TRUSTED_PROXY_CIDRS`. Esa lista contiene únicamente la red interna del proxy,
  nunca rangos públicos ni direcciones de visitantes; una cabecera enviada por
  un peer no confiable se ignora.
- ADR-0089 fija esos orígenes públicos en `https://dev.fasttourney.com` para
  desarrollo y `https://fasttourney.com` para producción. Cada host publica
  exclusivamente la asociación de su propia aplicación nativa.

## Gates futuros

| Gate       | Evidencia                                                    |
| ---------- | ------------------------------------------------------------ |
| Diseño     | Modelo de amenazas y clasificación de datos                  |
| Cambio     | Pruebas de autorización, validación y abuso relevante        |
| Despliegue | Secretos, permisos, superficie expuesta y rollback revisados |
| Operación  | Alertas, runbook e historial auditable                       |

No se declarará “seguro” un componente; se documentarán amenazas consideradas,
controles, evidencia y riesgo residual.
# Reautenticación para cambios de credenciales

Los tickets de reautenticación son secretos de 256 bits, se persisten como
SHA-256 con prefijo de propósito y se consumen transaccionalmente junto con la
mutación. El ticket exige la misma sesión opaca válida y no se registra. La
verificación de contraseña decodifica el formato Argon2id y compara la salida
en tiempo constante; un formato inválido se trata como credencial no válida.
