# ADR-0077: Permitir actualizaciones inmediatas del conjunto compatible de Expo

- **Estado:** Aceptado
- **Fecha:** 2026-08-09
- **Decisor:** Usuario, mediante decisión explícita
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** ADR-0075, solo en la excepción de edad para una actualización de compatibilidad de Expo
- **Superado por:** Ninguno

## Problema

La política de siete días de ADR-0075 evitó resolver parches jóvenes de Expo,
pero dejó el cliente con una combinación parcial de paquetes Expo y React Native
que falla al montar vistas nativas. Expo pide actualizar el conjunto compatible
en bloque; esperar siete días puede prolongar una incompatibilidad que impide
probar y distribuir una build nativa válida.

## Contexto y restricciones

- ADR-0015 adopta Expo y CNG: el código JavaScript y los módulos nativos deben
  ser compatibles y se reconstruyen juntos cuando cambian dependencias nativas.
- ADR-0075 sigue protegiendo toda actualización ordinaria, directa y transitiva,
  con lockfile congelado, edad mínima y revisión explícita.
- pnpm admite exclusiones de edad por paquete y versión. No conoce el concepto
  de «actualización requerida por Expo», por lo que la condición se controla en
  el procedimiento y en el diff revisado, no con un comodín de paquete.
- Expo recomienda actualizar las dependencias que correspondan al SDK instalado
  mediante `expo install --fix` y verificar después el resultado.

## Criterios de decisión

1. recuperar de inmediato un conjunto nativo que Expo declara compatible;
2. no desactivar la maduración para paquetes o futuras versiones no solicitadas;
3. conservar una excepción visible, exacta y eliminable;
4. fijar una resolución reproducible para desarrollo, CI y build nativa;
5. exigir validación proporcional al riesgo nativo.

## Alternativas

### A — Excepciones exactas, solo para el conjunto que Expo solicita

Ejecutar `expo install --fix` cuando `expo install --check` detecte una matriz
incompatible. Añadir a `minimumReleaseAgeExclude` únicamente las versiones
exactas, directas o transitivas, que el resolvedor necesita para esa operación;
fijar los paquetes directos actualizados a la versión resultante y revisar el
lockfile.

- **Ventajas:** mantiene la regla de siete días fuera de este caso; la excepción
  es visible y acotada; admite el conjunto completo que Expo valida.
- **Inconvenientes:** requiere revisar y registrar una lista de versiones; la
  primera actualización puede incluir varias dependencias transitivas.
- **Coste de mantenimiento:** bajo, limitado a upgrades reales de Expo.

### B — Excluir permanentemente `expo`, `@expo/*`, React Native y Reanimated

- **Ventajas:** no requiere editar la lista por actualización.
- **Inconvenientes:** autoriza versiones jóvenes futuras aunque Expo no las haya
  solicitado y puede no cubrir transitivas ajenas a esos prefijos.
- **Coste de mantenimiento:** bajo, con una superficie de suministro demasiado
  amplia.

### C — Mantener siete días sin excepción

- **Ventajas:** máxima uniformidad de la política de edad.
- **Inconvenientes:** conserva una matriz que Expo marca incompatible y puede
  bloquear correcciones nativas urgentes.
- **Coste de mantenimiento:** bajo, a costa de disponibilidad y diagnóstico.

## Recomendación

**Recomendación:** alternativa A. Es la excepción mínima suficiente: concreta,
revisable y reversible, sin convertir a Expo en una fuente permanentemente no
sujeta a maduración.

## Decisión del usuario

**Aceptada:** cuando Expo solicite actualizar en bloque el conjunto compatible
del SDK, el proyecto actualizará inmediatamente. La excepción a siete días se
limita a las versiones exactas requeridas por esa ejecución; no se admiten
comodines ni exclusiones permanentes por familia de paquetes. Los paquetes
directos afectados se fijan a las versiones resultantes y el lockfile conserva
el árbol transitivo exacto.

## Reglas de implementación

1. Ejecutar `pnpm --filter @tournaments-manager/client exec expo install --check`.
2. Si Expo pide una corrección del conjunto, registrar las versiones jóvenes
   necesarias en `minimumReleaseAgeExclude` como `nombre@versión` y ejecutar
   `pnpm --filter @tournaments-manager/client exec expo install --fix`.
3. Convertir a versiones exactas los paquetes directos modificados por Expo;
   no actualizar paquetes no solicitados para aprovechar la excepción.
4. Revisar manifiesto, exclusiones y `pnpm-lock.yaml`. Las exclusiones dejan de
   ser necesarias al superar siete días y se eliminan en una actualización
   posterior deliberada.
5. Ejecutar `expo install --check`, typecheck, exportación web y una build
   limpia en iOS o Android afectado; comprobar el arranque y el flujo tocado.

## Consecuencias

- La regla ordinaria de siete días y el lockfile congelado continúan vigentes.
- Una excepción equivocada puede introducir código joven en el árbol nativo;
  se mitiga mediante versiones exactas, revisión del diff y pruebas nativas.
- Las dependencias directas dejan de usar rangos semiautomáticos tras una
  corrección de compatibilidad de Expo. Las actualizaciones futuras serán
  deliberadas y trazables.

## Validación

- `pnpm install` permanece congelado y no modifica la resolución aprobada.
- Sin una excepción exacta, pnpm rechaza una versión de menos de siete días.
- Con la lista exacta de esta actualización, `expo install --fix` resuelve el
  conjunto solicitado y `expo install --check` termina correctamente.
- La build nativa limpia no muestra una incompatibilidad de `ComponentView` y
  el flujo de liga se puede recorrer en el simulador.

## Disparadores de revisión

- Expo deja de recomendar actualizaciones de conjunto mediante su CLI.
- La lista de excepciones crece de forma sostenida o requiere comodines.
- Una actualización inmediata introduce una regresión de seguridad o runtime.
- El proyecto adopta una fuente de inteligencia de seguridad que permita una
  política más precisa que la edad de publicación.

## Documentación afectada

- [ADR-0075](0075-freeze-local-lockfiles-and-delay-dependency-releases.md)
- [Desarrollo](../engineering/DEVELOPMENT.md)
- [Contribución](../../CONTRIBUTING.md)
- [Decisiones](../governance/DECISIONS.md)
- [Aprendizaje](../project/LEARNING.md)
- [Changelog](../../CHANGELOG.md)

## Fuentes técnicas

- [pnpm: configuración de resolución](https://pnpm.io/settings#minimumreleaseage)
- [Expo: actualizar el SDK](https://docs.expo.dev/workflow/upgrading-expo-sdk-walkthrough/)
