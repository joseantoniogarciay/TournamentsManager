# ADR-0170: Actualizar PostHog maduro y revisar TypeScript 7

- **Estado:** Aceptado para mantenimiento de PostHog; TypeScript conserva ADR-0014
- **Fecha:** 2026-10-31
- **Decisor:** Usuario: «Ya podemos ir a por typescript y posthog»
- **Complementa:** ADR-0014, ADR-0138, ADR-0149 y ADR-0169.
- **Superado por:** ADR-0171 en el aplazamiento de TypeScript; PostHog conserva esta decisión.

## Problema, evidencia y alternativas

ADR-0169 aplazó estas dependencias. El registro oficial publica TypeScript
7.0.2 y PostHog React Native 4.85.1, pero latest no acredita compatibilidad ni
maduración. TypeScript 7 sustituye el compilador JavaScript por uno nativo y
no proporciona su API estable; typescript-eslint 8.72.0 y 8.72.1 requieren
`>=4.8.4 <6.1.0`. Orval 8.44.0 declara TypeScript 6.0.3 como dependencia.

Mantener TypeScript 6.0.3 conserva un compilador compatible con coste bajo.
Ejecutar 7 junto a 6 mediante alias es una alternativa documentada por Microsoft,
pero obliga a mantener dos compiladores, comandos y configuración del editor.
Reemplazar directamente 6 por 7 incumpliría la compatibilidad aceptada.
Se recomienda conservar 6 hasta compatibilidad completa; la adopción dual
requiere decisión explícita adicional y no se implementa en este cambio.

El usuario pregunta después por alternativas a ESLint. Se recomienda evaluar
Oxlint manteniendo Prettier: su lint con tipos estable se basa en TS 7.0.2,
aunque el proyecto actual no activa reglas con tipos. Biome integra también
formato, con diferencias de reglas que requieren revisión. Doce archivos de
pruebas usan `transpileModule`; el editor apunta a `typescript/lib`. El cambio
de linter por sí solo no resuelve estos consumidores. La sustitución del tooling
sigue pendiente de decisión explícita y de un ADR que supere esa parte de ADR-0014.

PostHog 4.83.0 y su plugin 2.12.6 son las versiones estables más recientes con
siete días a la fecha de revisión. Mantener las anteriores evita reconstrucción,
pero pierde correcciones. Instalar latest requiere una excepción de maduración
que esta petición no concede. Actualizar las maduras tiene coste acotado de
regresión y reconstrucción nativa.

## Decisión y límites

La autorización de mantenimiento permite actualizar PostHog mediante Expo,
con versiones exactas 4.83.0 y 2.12.6, revisando también sus transitivas.
No se altera la política de suministro ni la matriz Expo. TypeScript continúa
en 6.0.3 con las opciones estrictas aceptadas, pendiente de resolver la
alternativa dual con el usuario.

Se conserva ADR-0149: fiabilidad mínima solo en producción, beta/local apagados,
sin analítica de producto, replay, flags ni identificación. La actualización
no autoriza distribución, despliegue, carga de símbolos ni envío de telemetría.
La prueba distribuible y simbolización reales de ADR-0109 siguen pendientes.

## Validación y retrospectiva

El [informe de revisión](../engineering/TYPESCRIPT_POSTHOG_REVIEW_2026-10-31.md)
registra fechas, grafo, pruebas y límites de verificación. Una exportación web
no acredita compilación nativa, y una build debug no acredita un crash recibido
y simbolizado en producción. La maduración también incluye CocoaPods y Maven.

## Fuentes

- [Microsoft: TypeScript 7 y convivencia con 6](https://devblogs.microsoft.com/typescript/announcing-typescript-7-0/).
- [Compatibilidad de typescript-eslint](https://typescript-eslint.io/users/dependency-versions/).
- [Registro TypeScript](https://registry.npmjs.org/typescript).
- [Registro PostHog React Native](https://registry.npmjs.org/posthog-react-native).
- [Registro del plugin](https://registry.npmjs.org/@posthog/react-native-plugin).
