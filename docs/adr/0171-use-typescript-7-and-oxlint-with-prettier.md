# ADR-0171: Adoptar TypeScript 7 y Oxlint conservando Prettier

- **Estado:** Aceptado
- **Fecha:** 2026-10-31
- **Decisor:** Usuario: «Vamos con la migración sí», tras recomendar Oxlint + Prettier + TypeScript 7.0.2
- **Supera a:** ADR-0014 en compilador, linter y configuración del editor; ADR-0170 en el aplazamiento de TypeScript.

## Problema y evidencia

TypeScript 7.0.2 es estable, pero su API no reemplaza la de TypeScript 6 usada
por typescript-eslint y doce archivos de pruebas. La configuración existente
activa 76 reglas recommended/strict, sin reglas dependientes de información de
tipos. El editor apunta al antiguo tsserver mediante typescript.tsdk.

## Alternativas y coste

- Conservar ESLint y TS 6: coste bajo, aplaza la adopción solicitada.
- TS 7 para tipos y TS 6 para herramientas: viable mediante alias, pero mantiene
  dos compiladores y su configuración en el proyecto.
- Oxlint con Prettier: coste medio de migrar reglas, pruebas y editor; preserva
  el formato y desacopla el lint del API JavaScript anterior.
- Biome: integra formato y lint; requiere revisar también diferencias de
  formato y equivalencia de reglas, ampliando el cambio.

## Recomendación y decisión explícita

El usuario acepta Oxlint + Prettier y TypeScript 7.0.2. Se fija Oxlint 1.89.0,
la publicación más reciente con siete días de maduración. Se migran las reglas
efectivas y se documenta cualquier diferencia sin ocultarla con supresiones.
No se añaden reglas con tipos ni tsgolint, porque el baseline no los utiliza.
Prettier y las opciones estrictas de tipos se conservan.

Las pruebas usan esbuild 0.28.1 para transformar TS/TSX a CommonJS y Babel Parser
7.29.9 para extraer la declaración de navegación desde su AST. Ambas versiones
ya están presentes transitivamente y se declaran como dependencias directas de
pruebas. No se introduce otro runner ni se reescriben los casos funcionales.
Orval conserva su compilador 6 interno si lo necesita: retirar la dependencia
raíz de TS 6 no implica eliminarlo del grafo de herramientas externas.

Se adapta VS Code a la extensión TypeScript 7 y Oxc. Se mantiene tsc como
comprobación de tipos separada del lint. La revisión del root tsconfig debe
conservar una comprobación real tras retirar eslint.config.mjs.

## Verificación y límites

Instalación congelada, maduración también de binarios por plataforma, mapa de
reglas, diagnóstico de código inválido, tipos estrictos, regresiones, generación
OpenAPI determinista y exportación web. No se inicia API, Metro ni observabilidad,
no se despliega ni se modifica producto, contrato o política de privacidad.

## Resultado y retrospectiva

Migración implementada: formato, lint, tipos, 125 regresiones Node, exportación
web, instalación congelada y dos generaciones idénticas de 145 archivos.
Las 43 versiones revisadas cumplen maduración. TypeDoc resuelve realmente TS
6.0.3 mediante hook limitado a 0.28.20; el compilador raíz sigue en 7.0.2.
La migración automática no conserva todos los scopes: se reconstruyen reglas
efectivas y se prueban rechazos deliberados, incluido modo estricto de CJS.

La auditoría npm autorizada devuelve 43 avisos en paquetes/versiones presentes
antes de esta migración (15 altos, 18 moderados, 10 bajos; cero críticos).
Queda un inventario de seguridad separado, sin aplicar correcciones automáticas
ni presentar el cierre del tooling como ausencia de vulnerabilidades.

El [informe de migración](../engineering/TYPESCRIPT_7_OXLINT_MIGRATION_2026-10-31.md)
registrará resultados y retrospectiva. La equivalencia de nombres no garantiza
idéntico comportamiento: se revisan opciones y diferencias de implementación.

## Fuentes

- [Microsoft: TypeScript 7](https://devblogs.microsoft.com/typescript/announcing-typescript-7-0/).
- [Oxlint: migración de ESLint](https://oxc.rs/docs/guide/usage/linter/migrate-from-eslint.html).
- [Oxlint: editor](https://oxc.rs/docs/guide/usage/linter/editors).
- [esbuild: transform](https://esbuild.github.io/api/#transform).
- [Babel Parser](https://babeljs.io/docs/babel-parser).
