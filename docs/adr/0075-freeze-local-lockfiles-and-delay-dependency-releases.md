# ADR-0075: Congelar el lockfile local y retrasar versiones nuevas

- **Estado:** Aceptado
- **Fecha:** 2026-08-04
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** ADR-0077, solo para actualizaciones de compatibilidad de Expo solicitadas por Expo CLI

## Problema

El repositorio ya instala con `--frozen-lockfile` en CI, pero una instalación
local ordinaria podía actualizar la resolución. Además, aceptar de inmediato
versiones recién publicadas expone tanto dependencias directas como transitivas
a incidentes de cadena de suministro antes de que haya tiempo razonable para su
detección pública.

## Contexto y restricciones

- ADR-0014 fija Node, pnpm, versiones exactas y un lockfile único; esta decisión
  lo complementa, sin cambiar de gestor de paquetes.
- La alerta de Shai-Hulud demostró que un `preinstall` malicioso puede ejecutarse
  antes de las comprobaciones habituales y afectar credenciales del entorno.
- El proyecto usa pnpm 11.17.0, cuya configuración versionable permite exigir
  lockfile congelado y una edad mínima de publicación también para el árbol
  transitivo.
- La espera no prueba que una versión sea segura ni sustituye revisar advisories
  o rotar secretos si existe exposición.

## Criterios de decisión

1. impedir cambios accidentales de resolución durante el trabajo local;
2. aplicar la misma espera a dependencias directas y transitivas;
3. fallar de forma visible, sin instalar una versión joven por degradación;
4. evitar un scanner, bot o servicio adicional mientras pnpm cubra el caso;
5. conservar una vía explícita y auditable para cambios intencionados.

## Alternativas

### A — Configuración nativa de pnpm: lockfile congelado y siete días

Configurar `frozenLockfile: true`, `minimumReleaseAge: 10080`, modo estricto y
rechazo de metadatos sin fecha en `pnpm-workspace.yaml`.

- **Ventajas:** se versiona junto al workspace; cubre dependencias transitivas;
  no añade credenciales, servicio ni automatización propia.
- **Inconvenientes:** una actualización urgente puede quedar bloqueada hasta que
  exista una versión madura compatible; requiere una acción explícita para
  actualizar el lockfile.
- **Coste de adopción:** mínimo.
- **Coste de mantenimiento:** bajo; revisar las excepciones exactas ya fijadas
  y los diffs de dependencia.
- **Riesgos:** una versión con siete días puede seguir ser maliciosa o vulnerable.

### B — Lockfile congelado sin periodo de maduración

- **Ventajas:** máxima simplicidad y sin retrasar actualizaciones.
- **Inconvenientes:** al cambiar el lockfile puede resolverse una publicación de
  minutos de antigüedad, incluida una dependencia transitiva.
- **Coste de adopción y mantenimiento:** mínimo.
- **Riesgos:** no reduce la exposición a compromisos recién publicados.

### C — Scanner o bot externo de dependencias

- **Ventajas:** podría añadir fuentes de inteligencia o políticas más complejas.
- **Inconvenientes:** añade integración, permisos, coste y mantenimiento antes
  de demostrar que la capacidad nativa es insuficiente.
- **Coste de adopción y mantenimiento:** medio.
- **Riesgos:** falsa sensación de cobertura y ampliación de la cadena de suministro.

### No cambiar

CI seguiría congelado, pero las instalaciones locales y actualizaciones
ordinarias conservarían una política menos restrictiva.

## Comparación

La alternativa A satisface los cinco criterios con la capacidad ya incluida en
pnpm. B solo protege la resolución existente. C puede reevaluarse si hay una
necesidad demostrada de inteligencia adicional, pero ahora sería sobreingeniería.

## Recomendación

**Opinión/recomendación:** alternativa A, como solución mínima suficiente.

## Decisión del usuario

**Aceptada el 2026-08-04:** las instalaciones locales usarán lockfile congelado
por defecto y pnpm no instalará una versión publicada hace menos de siete días.
La política aplica también a dependencias transitivas, es estricta y rechaza
metadatos de registro sin fecha. No se toman exclusiones generales para obtener
la última versión.

## Consecuencias

- `pnpm install` falla localmente si manifiestos y lockfile divergen.
- Los cambios intencionados se realizan mediante `pnpm add`, `pnpm update` o
  `pnpm install --no-frozen-lockfile`, con revisión del diff antes de integrarlo.
- Si no existe una versión compatible con al menos siete días, la resolución
  falla; no se rebaja silenciosamente a una versión reciente.
- Las exclusiones existentes continúan siendo versiones exactas ya fijadas, no
  permisos por paquete para aceptar futuras publicaciones.

## Validación

- `pnpm config list --location project` muestra `frozenLockfile: true`,
  `minimumReleaseAge: 10080`, `minimumReleaseAgeStrict: true` y
  `minimumReleaseAgeIgnoreMissingTime: false`.
- `pnpm install` funciona con el lockfile vigente y no lo modifica.
- Cambiar un manifiesto sin actualizar el lockfile hace fallar `pnpm install`.
- Una resolución de paquete nuevo con menos de siete días falla sin escribir el
  lockfile.

## Disparadores de revisión

- Una corrección de seguridad no puede aplicarse porque no existe versión madura
  compatible.
- El registro público deja de servir metadatos de tiempo de forma fiable.
- Aparece un incidente que demuestre insuficiente la edad mínima o justifique
  una fuente de inteligencia adicional.
- El proyecto incorpora un registro privado con requisitos distintos.

## Documentación afectada

- [Toolchain TypeScript](../engineering/DEVELOPMENT.md)
- [Contribución](../../CONTRIBUTING.md)
- [Decisiones](../governance/DECISIONS.md)
- [Aprendizaje](../project/LEARNING.md)
- [Changelog](../../CHANGELOG.md)

## Fuentes técnicas

- [pnpm: `minimumReleaseAge`](https://pnpm.io/settings/dependency-resolution#minimumreleaseage)
- [pnpm: instalación y lockfile congelado](https://pnpm.io/cli/install#--frozen-lockfile)
