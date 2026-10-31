# ADR-0173: Aplicar la corrección de Forge con una excepción exacta de edad

- **Estado:** Aceptado
- **Fecha:** 2026-10-31
- **Decisor:** Usuario: «Si soluciona lo metemos», tras explicar los seis altos, el candidato y la espera pendiente.
- **Complementa:** ADR-0172; excepción puntual a ADR-0138 para node-forge@1.4.2.

## Problema y evidencia

Los seis avisos altos restantes y otros seis registros afectan a Forge 1.4.0,
resuelto por Expo CLI y code-signing-certificates. La ficha antigua
GHSA-86w9-cpqp-85rv carece de versión corregida; el mantenedor confirma el
mismo CVE-2026-85393 en GHSA-2g46-22mc-6gwp y la corrección en 1.4.1.

El changelog de 1.4.1 incluye los seis altos. La versión 1.4.2 restaura el uso
de Buffer en pem.decode(), roto por la corrección de 1.4.1. Publicada el
2026-10-27T14:09:15.471Z, aún no cumple siete días y no declara dependencias.
No se presenta ninguno de estos avisos como crítico.

## Alternativas, coste y recomendación

- Esperar hasta el 3 de noviembre: coste bajo, conserva los avisos altos.
- Instalar 1.4.1: corrige seguridad, pero deja una regresión PEM conocida.
- Instalar 1.4.2 exacta: conserva las correcciones y repara esa compatibilidad;
  requiere excepción puntual y pruebas de seguridad y del consumidor Expo.

Se recomienda la tercera opción como corrección mínima suficiente. No se
actualiza Expo ni se añaden librerías o mecanismos de auditoría.

## Decisión explícita y alcance

El usuario autoriza instalar la corrección si resuelve los avisos. Se concreta
la autorización en node-forge@1.4.2, tras comprobar el parche del mantenedor y
la regresión de 1.4.1. Añadir únicamente el selector node-forge@1.4.0 → 1.4.2
y la excepción minimumReleaseAgeExclude exacta de 1.4.2. Mantener el control
global y todas las demás esperas. Retirar la excepción al cumplir siete días,
el 2026-11-03T14:09:15.471Z, conservando la resolución corregida.

## Verificación y límites

Resolver candidato aislado, revisar que cambia solo Forge, instalar congelado,
comprobar el código npm frente al tag, ejecutar regresiones oficiales relevantes
y probar la generación, lectura y validación de certificados de Expo.
Repetir audit con la autorización de envío del grafo a npm vigente en la sesión;
contrastar los doce IDs previos y cualquier nuevo aviso, sin equiparar un conteo
menor a ausencia absoluta de vulnerabilidades. Ejecutar check, regresiones y
exportación web. Subir el bloque a develop según la instrucción vigente; no
promocionar main, desplegar, arrancar local/dev persistente ni observabilidad.

## Resultado y retrospectiva

Implementado: solo cambia Forge 1.4.0 → 1.4.2. Audit elimina sus doce
registros y conserva tres moderados, sin altos, bajos, críticos ni IDs nuevos.
Pasan make verify, 126 regresiones Node y 254 pruebas oficiales seleccionadas.
Cuatro comparaciones deterministas de keygen se omiten por la generación
nativa de Node. El vector RSA antiguo acepta la firma malformada en 1.4.0
y se rechaza en 1.4.2; Expo conserva PEM Buffer y firma/verificación.

La retrospectiva es contrastar CVE y parche del mantenedor cuando el agregador
no está actualizado, y probar también el consumidor: 1.4.2 evita la regresión
de compatibilidad introducida por 1.4.1. El [informe](../engineering/FORGE_SECURITY_FIX_2026-10-31.md)
conserva evidencia, límites y fecha de retirada de la excepción.

## Fuentes

- [Aviso del mantenedor para CVE-2026-85393](https://github.com/digitalbazaar/forge/security/advisories/GHSA-2g46-22mc-6gwp).
- [Changelog de 1.4.2](https://github.com/digitalbazaar/forge/blob/v1.4.2/CHANGELOG.md).
- [Registro npm](https://registry.npmjs.org/node-forge).
