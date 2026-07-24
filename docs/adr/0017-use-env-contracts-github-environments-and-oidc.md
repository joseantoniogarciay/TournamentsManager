# ADR-0017: Usar contratos de entorno, GitHub Environments y OIDC

- **Estado:** Aceptado
- **Fecha:** 2026-07-24
- **Decisor:** Usuario, mediante elección explícita de la alternativa A
- **Propietario del análisis:** Asistente como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

El sistema necesitará configuración para backend, cliente universal, entorno
local, CI y despliegues. El repositorio es público, por lo que hay que decidir
cómo separar configuración no sensible, secretos y datos operativos sensibles
sin perder reproducibilidad.

## Contexto y restricciones

- ADR-0006 acepta GitHub público con secretos fuera de Git.
- Backend Go, cliente Expo, Docker Compose, GitHub Actions y AWS futura tendrán
  necesidades de configuración distintas.
- El cliente web/mobile no puede recibir secretos: todo valor embebido en un
  bundle debe tratarse como público.
- El entorno local debe parecerse a producción en contratos importantes, no en
  topología exacta.
- Todavía no existen `apps/client`, `compose.yaml`, workflows de CI ni despliegue.

## Criterios de decisión

1. evitar secretos en Git, logs, bundles e imágenes;
2. mantener un contrato claro de variables por aplicación;
3. permitir desarrollo local sencillo;
4. encajar con GitHub Actions y entornos protegidos;
5. evitar credenciales cloud de larga vida cuando sea posible;
6. no introducir un gestor de secretos prematuro.

## Alternativas

### Alternativa A — `.env` local, ejemplos, GitHub Environments y OIDC

Usar archivos `.env` reales ignorados por Git, archivos `.env.example`
versionados con nombres y valores ficticios, GitHub variables/secrets por
entorno y OIDC para cloud cuando el proveedor lo permita.

- **Ventajas:** estándar, barato, fácil de entender, compatible con Docker
  Compose, GitHub Actions, VPS y AWS futura.
- **Inconvenientes:** exige disciplina; los `.env` locales no se sincronizan; la
  validación de configuración debe implementarse explícitamente.
- **Coste de adopción:** bajo.
- **Coste de mantenimiento:** bajo o medio.

### Alternativa B — Gestor de secretos desde el inicio

Centralizar secretos en AWS Secrets Manager, HashiCorp Vault, Doppler, 1Password
u otro servicio.

- **Ventajas:** auditoría, control centralizado, rotación y menor dependencia de
  archivos locales.
- **Inconvenientes:** añade proveedor, permisos, CLI, onboarding y coste antes de
  tener despliegue real.
- **Coste de adopción:** medio o alto.
- **Coste de mantenimiento:** medio o alto, especialmente si se opera Vault.

### Alternativa C — Configuración cifrada en Git con SOPS

Versionar archivos cifrados y descifrarlos localmente o en CI mediante claves
controladas.

- **Ventajas:** trazabilidad en Git, buena compatibilidad futura con GitOps y
  Kubernetes.
- **Inconvenientes:** gestión de claves y rotación antes de necesitar GitOps.
- **Coste de adopción:** medio.
- **Coste de mantenimiento:** medio.

### Alternativa D — Configuración manual sin contrato

Configurar variables a mano en shells, servidores y CI, documentando solo de
forma informal.

- **Ventajas:** arranque rápido.
- **Inconvenientes:** deriva entre entornos, errores difíciles de diagnosticar y
  auditoría pobre.
- **Coste de adopción:** bajo.
- **Coste de mantenimiento:** alto.

## Comparación

La alternativa A cubre las necesidades inmediatas con el menor coste y sin
cerrar la evolución a gestores de secretos o SOPS. B y C son opciones válidas
cuando haya varios entornos reales, rotación frecuente o GitOps. D ahorra trabajo
al principio, pero contradice el objetivo de reproducibilidad y seguridad.

## Recomendación

**Opinión/recomendación:** alternativa A. Es la base mínima profesional para un
repositorio público y un equipo pequeño: contratos explícitos, secretos fuera de
Git, entornos protegidos en GitHub y credenciales temporales para cloud futura.

## Decisión del usuario

**Aceptada:** usar `.env` locales ignorados, `.env.example` versionados, GitHub
Environments para entornos de despliegue y OIDC para cloud cuando esté
disponible.

## Reglas de implementación

- `.env`, `.env.*`, claves privadas, tokens, estado Terraform e inventarios
  sensibles no se versionan.
- `.env.example` contiene nombres, comentarios y valores ficticios; nunca
  secretos reales.
- Cada aplicación o servicio tendrá su propio contrato de configuración cuando
  exista: backend, cliente, Compose, CI e infraestructura no comparten un `.env`
  global por comodidad.
- El backend Go leerá configuración desde variables de entorno y fallará al
  arrancar si falta una variable obligatoria o tiene formato inválido.
- Los secretos no se imprimen en logs, errores, métricas ni trazas.
- Las variables del cliente Expo que lleguen a JavaScript deben usar
  `EXPO_PUBLIC_` y se consideran públicas.
- Ningún secreto se expone al bundle web/mobile aunque el nombre de variable no
  use `EXPO_PUBLIC_`.
- Docker Compose podrá usar `env_file` para local; Compose secrets se usarán
  solo cuando aporten aislamiento real frente a una variable normal.
- GitHub Actions separará valores no sensibles en variables y secretos en
  `secrets`.
- Producción usará GitHub Environment protegido con permisos mínimos y revisión
  antes de acceder a secretos.
- AWS futura usará OIDC y credenciales temporales siempre que el servicio lo
  permita; no se almacenarán access keys de larga vida salvo excepción
  documentada.
- Un VPS, si se usa, tendrá identidad de despliegue dedicada; no se reutilizan
  claves personales.

## Consecuencias

### Positivas

- El contrato de configuración será visible sin publicar secretos.
- Local, CI y producción podrán evolucionar de forma trazable.
- La futura integración cloud puede evitar credenciales persistentes en GitHub.
- Expo queda protegido frente al error común de tratar variables públicas como
  secretas.

### Negativas y deuda aceptada

- Los `.env` locales requieren creación manual o tooling posterior.
- La rotación y auditoría centralizada quedan aplazadas.
- Habrá que implementar validación explícita de configuración en cada runtime.

## Validación

- `.gitignore` bloquea `.env`, claves, Terraform state y planes.
- El primer backend tendrá validación de configuración al arrancar.
- Los ejemplos de entorno contendrán solo valores ficticios.
- Los workflows no accederán a secretos salvo en jobs y environments que los
  necesiten.
- Cualquier variable `EXPO_PUBLIC_*` se revisará como dato público.

## Disparadores de revisión

- Aparecen varios entornos reales con rotación frecuente de secretos.
- Se incorpora un equipo que necesita onboarding y auditoría de acceso a
  secretos.
- Kubernetes/GitOps hace valioso versionar secretos cifrados.
- Un proveedor exige credenciales persistentes que no pueden evitarse con OIDC.
- Un incidente o casi incidente demuestra que `.env` y GitHub secrets no bastan.

## Documentación afectada

- [TECHNICAL_BASELINE.md](../governance/TECHNICAL_BASELINE.md)
- [SECURITY.md](../engineering/SECURITY.md)
- [DEVELOPMENT.md](../engineering/DEVELOPMENT.md)
- [DEPLOYMENT.md](../operations/DEPLOYMENT.md)
- [DECISIONS.md](../governance/DECISIONS.md)
- [DECISIONS_TO_REVISIT.md](../governance/DECISIONS_TO_REVISIT.md)
- [LEARNING.md](../project/LEARNING.md)

## Fuentes técnicas

- [GitHub Actions secrets](https://docs.github.com/en/actions/concepts/security/secrets)
- [GitHub deployment environments](https://docs.github.com/en/actions/concepts/workflows-and-actions/deployment-environments)
- [GitHub OIDC reference](https://docs.github.com/en/actions/reference/security/oidc)
- [Docker Compose environment variables](https://docs.docker.com/compose/how-tos/environment-variables/set-environment-variables/)
- [Docker Compose secrets](https://docs.docker.com/reference/compose-file/secrets/)
- [Expo environment variables](https://docs.expo.dev/guides/environment-variables/)
