# ADR-0014: Usar Node LTS, pnpm y TypeScript estricto

- **Estado:** Aceptado
- **Fecha:** 2026-07-24
- **Decisor:** Usuario, mediante confirmación explícita
- **Propietario del análisis:** Asistente como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

El cliente universal y el cliente generado desde OpenAPI necesitan un entorno
TypeScript reproducible. Hay que fijar el runtime de herramientas, el gestor de
paquetes, los workspaces, la política de tipos, el formato, el lint y los comandos
comunes sin adelantar la elección del framework.

## Contexto y restricciones

- TypeScript y el cliente API generado están aceptados en ADR-0009.
- El monorepo de producto está aceptado en ADR-0005.
- El framework universal y su versión todavía están pendientes.
- Node ejecutará tooling y bundling; no será necesariamente el runtime de la
  aplicación en navegador o dispositivo.
- El manifiesto exige reproducibilidad y evitar tooling adicional sin evidencia.
- Expo admite workspaces de npm, pnpm, Yarn y Bun, pero su elección sigue
  pendiente del siguiente gate.

## Criterios

1. compatibilidad con React Native, web y generación OpenAPI;
2. versiones reproducibles en desarrollo y CI;
3. dependencias explícitas y aisladas en el monorepo;
4. feedback temprano de tipos, lint y formato;
5. integración sencilla con VS Code;
6. coste de mantenimiento proporcional a un equipo pequeño.

## Alternativas

### Alternativa A — Node LTS, pnpm, TypeScript estricto y ESLint/Prettier

- **Ventajas:** runtime estable; workspaces y lockfile únicos; dependencias
  aisladas; herramientas maduras y compatibles con el ecosistema objetivo.
- **Inconvenientes:** pnpm añade una herramienta frente a npm; algunas
  dependencias nativas antiguas pueden asumir un árbol hoisted.
- **Coste de mantenimiento:** bajo o medio.

### Alternativa B — Node LTS, npm workspaces y ESLint/Prettier

- **Ventajas:** npm acompaña a Node y reduce requisitos de bootstrap.
- **Inconvenientes:** instalación menos estricta por defecto y ergonomía de
  workspaces menos orientada a monorepos que pnpm.
- **Coste de mantenimiento:** bajo.

### Alternativa C — Bun, workspaces y Biome

- **Ventajas:** instalación y herramientas rápidas; menor número de procesos.
- **Inconvenientes:** algunos flujos Expo todavía requieren Node LTS; se
  mantendrían dos runtimes y la integración específica de React Native es más
  reciente.
- **Coste de mantenimiento:** medio por compatibilidad y diagnóstico.

## Comparación

La alternativa B es la mínima en número de herramientas. La C optimiza velocidad
antes de que exista un problema medido y no elimina Node en el ecosistema
objetivo. La A añade pnpm, pero ofrece aislamiento de dependencias, workspaces
explícitos y un camino oficialmente soportado por Expo.

## Recomendación

**Opinión/recomendación:** alternativa A. No añadir Nx ni Turborepo mientras los
scripts recursivos de pnpm y Make sean suficientes.

## Decisión del usuario

**Aceptada:** usar:

- Node 24 LTS, inicialmente 24.18.0;
- pnpm 11.17.0 con workspaces y lockfile único;
- instalación aislada por defecto;
- TypeScript 6.0.3;
- `strict`, `noUncheckedIndexedAccess`, `exactOptionalPropertyTypes`,
  `noImplicitOverride` y `noEmit`;
- ESLint 10 con flat config y reglas recomendadas/estrictas de
  `typescript-eslint`;
- Prettier como formateador;
- formato al guardar en VS Code;
- comandos `format`, `format-check`, `lint`, `typecheck`, `check` y `verify`.

La versión 6.0.3 de TypeScript se fija porque es la última compatible con el rango
oficial actual de `typescript-eslint` (`>=4.8.4 <6.1.0`). Una etiqueta `latest`
incompatible no prevalece sobre el conjunto validado.

## Reglas de implementación

- `package.json` raíz es privado y concentra el tooling compartido.
- `pnpm-workspace.yaml` declara `apps/*` y `packages/*`.
- Las dependencias usan versiones exactas y `pnpm-lock.yaml` se versiona.
- `devEngines.runtime` selecciona Node 24.18.0 para los scripts del proyecto sin
  modificar el Node global de la máquina.
- `devEngines.packageManager` y `packageManager` fijan pnpm 11.17.0.
- CI instalará mediante lockfile congelado cuando se decida su workflow.
- Cada workspace declara sus dependencias directas.
- Se usa el linker aislado; cambiar a `hoisted` exige un problema de
  compatibilidad reproducido y documentado.
- `tsconfig.base.json` contiene las reglas transversales; cada aplicación o
  paquete tendrá su propio `tsconfig`.
- `tsconfig.json` raíz comprueba la configuración JavaScript del tooling y
  añadirá referencias al crear workspaces TypeScript.
- El código generado se formatea y comprueba, pero no se edita manualmente.
- La elección no autoriza Expo, routing, testing, Nx, Turborepo ni un generador
  OpenAPI concreto.

## Consecuencias

### Positivas

- Desarrollo y CI compartirán runtime, package manager y lockfile.
- El tipado estricto detectará estados ausentes y contratos ambiguos antes de
  ejecutar la aplicación.
- pnpm impedirá depender accidentalmente de paquetes no declarados.
- Editor, terminal y futuros checks de CI usarán la misma configuración.

### Negativas y deuda aceptada

- El bootstrap necesita una instalación compatible de pnpm o Corepack.
- El modo aislado puede descubrir librerías nativas con declaraciones incorrectas.
- `exactOptionalPropertyTypes` exige distinguir ausencia de propiedad y valor
  `undefined`.
- `noUncheckedIndexedAccess` obliga a tratar como posiblemente ausentes los
  accesos por índice.
- ESLint y Prettier son dos herramientas con responsabilidades separadas.

## Validación

- pnpm selecciona Node 24.18.0 y pnpm 11.17.0 para el proyecto.
- `pnpm install` genera un lockfile reproducible.
- `pnpm run check` verifica formato, lint y tipos.
- `make check` y `make verify` incluyen Go y TypeScript.
- VS Code usa la versión de TypeScript del workspace y Prettier al guardar.
- Una alteración de formato o un error TypeScript hace fallar el check
  correspondiente cuando existan fuentes.

## Disparadores de revisión

- Node 24 entra en mantenimiento o fin de vida.
- El framework elegido no soporta alguna versión fijada.
- Una dependencia necesaria no funciona con instalación aislada.
- ESLint/Prettier causan coste sostenido frente a una alternativa madura.
- El monorepo necesita cache o planificación de tareas por tiempos medidos.
- `typescript-eslint` soporta una nueva major de TypeScript y aporta una mejora
  relevante.

## Documentación afectada

- [TECHNICAL_BASELINE.md](../governance/TECHNICAL_BASELINE.md)
- [DEVELOPMENT.md](../engineering/DEVELOPMENT.md)
- [STYLEGUIDE.md](../engineering/STYLEGUIDE.md)
- [CONTRIBUTING.md](../../CONTRIBUTING.md)
- [DECISIONS.md](../governance/DECISIONS.md)
- [LEARNING.md](../project/LEARNING.md)

## Fuentes técnicas

- [Versiones de Node.js](https://nodejs.org/en/about/previous-releases)
- [Workspaces con Expo](https://docs.expo.dev/guides/monorepos/)
- [Configuración de pnpm](https://pnpm.io/package_json)
- [Compatibilidad de typescript-eslint](https://typescript-eslint.io/users/dependency-versions/)
- [Opciones de TSConfig](https://www.typescriptlang.org/tsconfig/)
