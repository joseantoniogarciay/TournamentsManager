# ADR-0006: Publicar el monorepo en GitHub sin publicar secretos

- **Estado:** Aceptado
- **Fecha:** 2026-07-23
- **Decisor:** Usuario, mediante confirmación explícita
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

El proyecto puede desarrollarse de forma abierta y aprovechar CI/CD público, pero
los workflows no deben exponer credenciales ni convertirse en una vía de acceso al
VPS o al cloud.

## Alternativas

### Repositorio privado

Reduce exposición del código y de los workflows. No elimina la necesidad de
gestionar secretos y runners de forma segura.

### Repositorio público con límites de seguridad

Hace público código, documentación y definiciones declarativas. Mantiene
credenciales, estados y acceso operativo fuera de Git, con despliegues protegidos.

### Código público y repositorio privado de despliegue

Aísla pipelines y datos operativos, pero divide la trazabilidad y añade
coordinación. Puede adoptarse más adelante si existe una necesidad real.

## Decisión del usuario

Usar GitHub como remoto público del monorepo, aplicando límites explícitos para
secretos y despliegues. La creación efectiva del remoto queda pendiente de indicar
cuenta, organización y nombre.

## Consecuencias

- Nunca se versionan `.env`, claves privadas, credenciales, tokens, estado de
  Terraform ni inventarios sensibles.
- `.env.example` solo contiene nombres y valores ficticios.
- Todo valor incorporado al bundle web o mobile se considera público.
- Los secretos de CI se almacenan como GitHub Actions/Environment Secrets.
- Producción usa un environment protegido, aprobación y ramas o tags autorizados.
- Cloud usará OIDC y credenciales temporales cuando el proveedor lo permita.
- Un VPS tendrá una identidad de despliegue dedicada, sin root y con permisos
  mínimos; nunca se reutiliza la clave SSH personal.
- No se instalará un runner self-hosted del repositorio público dentro del VPS o
  una red sensible.
- Código de forks o pull requests no confiables no se ejecutará con secretos.
- Los workflows y cambios de permisos recibirán revisión específica.

## Validación

Antes del primer remoto y del primer despliegue se comprobarán:

- `.gitignore` y escaneo de secretos;
- permisos mínimos de `GITHUB_TOKEN`;
- aislamiento por environments;
- ausencia de credenciales persistentes de cloud;
- ruta de despliegue sin runner público dentro del VPS.

## Disparadores de revisión

- Necesidad de ocultar inventario o arquitectura.
- Incorporación de colaboradores con permisos de escritura.
- Aparición de datos regulados.
- Uso de infraestructura que no pueda protegerse con identidades de corta vida.

## Referencias

- [GitHub: secretos](https://docs.github.com/en/actions/concepts/security/secrets)
- [GitHub: uso seguro de Actions](https://docs.github.com/en/actions/reference/security/secure-use)
- [GitHub: OIDC con cloud](https://docs.github.com/en/actions/how-tos/secure-your-work/security-harden-deployments/oidc-in-cloud-providers)
- [Expo: variables de entorno](https://docs.expo.dev/guides/environment-variables/)
