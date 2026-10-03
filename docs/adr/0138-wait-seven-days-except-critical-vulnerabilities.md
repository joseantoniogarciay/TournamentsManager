# ADR-0138: Esperar siete días salvo vulnerabilidades críticas

- **Estado:** Aceptado
- **Fecha:** 2026-10-03
- **Decisor:** Usuario, mediante decisión explícita en la conversación
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** ADR-0077, en su excepción por compatibilidad de Expo; complementa ADR-0075
- **Superado por:** Ninguno

## Problema y contexto

La excepción de Expo permitía instalar inmediatamente versiones jóvenes por
compatibilidad del SDK. El usuario prefiere mantener siete días de maduración
salvo que exista una vulnerabilidad crítica. El gate técnico está cerrado;
esta decisión modifica mantenimiento y suministro, no funcionalidad del producto.

## Alternativas y coste

- **Esperar siete días con excepción crítica:** política uniforme y coste bajo;
  puede retrasar una corrección de compatibilidad o de seguridad no crítica.
- **Conservar la excepción de Expo:** recupera antes su matriz compatible, pero
  amplía los motivos para aceptar publicaciones jóvenes y exige mantener listas.
- **Esperar siempre:** coste mínimo, pero retrasa incluso una corrección crítica.

## Recomendación y decisión explícita

**Recomendación:** primera alternativa, usando el control existente de pnpm y
revisión manual para gestores sin ese control; no añadir bots ni otro scanner.

**Decisión del usuario:** «Salvo que sea vulnerabilidad critica prefiero esperar
7 dias». Se acepta esperar al menos siete días desde la publicación de toda
versión nueva, directa o transitiva, incluidas dependencias de Expo, Go y
herramientas. Compatibilidad, novedades y avisos no críticos no permiten saltar
esa espera. Las versiones ya resueltas se mantienen; no se reinstala ni arranca
el entorno local que el usuario ha pedido retirar.

## Implementación y excepción

- Conservar lockfile congelado, `minimumReleaseAge: 10080`, modo estricto y
  rechazo de fechas ausentes. Retirar las anteriores exclusiones de Expo.
- En gestores sin control nativo de edad, verificar la fecha de publicación en
  la fuente oficial antes de actualizar; no afirmar que pnpm controla Go.
- Una excepción necesita un aviso de seguridad que califique la vulnerabilidad
  como crítica, evidencia de que afecta al proyecto y una versión corregida
  publicada. Registrar fuente, motivo y versiones exactas necesarias.
- Limitar la excepción a la corrección y sus requisitos transitivos; no usar
  comodines, desactivar el control global ni actualizar paquetes ajenos.
- Revisar el diff y ejecutar las verificaciones correspondientes. Un cambio
  nativo mantiene la validación de conjunto y build limpia prevista para Expo.
- Retirar cada excepción cuando la versión cumpla siete días. La autorización
  de esta categoría no convierte avisos altos o una incompatibilidad en críticos.

## Validación y consecuencias

La configuración mantiene siete días y no contiene exclusiones. La matriz de
Expo detectada el 2026-10-03 queda aplazada; no se aplica una excepción de edad
por ese motivo. El entorno local sigue desmontado. La comprobación de este
cambio es documental y estática; no se reinstalan dependencias para validar
una política que no modifica código ni resolución.

## Documentación afectada

- ADR-0075, ADR-0077 y registro de decisiones.
- Contribución, Desarrollo, Aprendizaje y Changelog.
