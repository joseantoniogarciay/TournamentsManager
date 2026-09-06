# ADR-0119: Versionar los hitos de producción con tags SemVer y GitHub Releases

- **Estado:** Aceptado
- **Fecha:** 2026-09-06
- **Decisor:** Usuario, mediante decisión explícita
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

La promoción de `develop` a `main` identifica un bloque estable, pero un SHA no
comunica qué hito de producto representa ni ofrece unas notas curadas sobre su
alcance, validación y límites. La web ya está publicada en producción y necesita
una referencia inmutable y recuperable sin crear releases por cada integración.

## Contexto y restricciones

- ADR-0013 mantiene `develop` como integración y `main` como estado estable.
- ADR-0092 reserva tags inmutables y GitHub Releases para producción o hitos
  distribuidos, no para cada push de `develop`.
- ADR-0056 fija `1.0.0` como versión inicial del producto y cliente.
- El primer hito publicado corresponde a la web; las apps nativas no están aún
  distribuidas. Su ausencia no altera la versión del código fuente compartido.
- Un release debe poder relacionar tag, SHA y artefacto activo, sin afirmar que
  componentes no incluidos hayan sido distribuidos o validados.

## Criterios de decisión

1. conservar una referencia inmutable, legible y recuperable;
2. distinguir hitos productivos de integraciones diarias;
3. no añadir automatización, ramas o artefactos innecesarios;
4. explicar con precisión alcance y limitaciones;
5. mantener compatible la distribución futura de web, iOS y Android.

## Alternativas

### Alternativa A — Mantener solo SHAs y el changelog

- **Ventajas:** no añade comandos ni metadatos de release.
- **Inconvenientes:** un SHA no expresa intención de producto; recuperar un
  hito exige reconstruir contexto desde commits y documentos.
- **Coste de adopción y mantenimiento:** bajo, con trazabilidad manual creciente.
- **Riesgos:** confundir una integración diaria con una referencia estable.

### Alternativa B — Crear un tag móvil `v1` y una Release informal

- **Ventajas:** nombre breve y visible en GitHub.
- **Inconvenientes:** `v1` no identifica una versión SemVer completa y podría
  moverse, anulando el valor de la referencia.
- **Coste de adopción y mantenimiento:** bajo, con ambigüedad permanente.
- **Riesgos:** no poder asociar con precisión un artefacto a su versión.

### Alternativa C — Tags SemVer anotados y GitHub Releases selectivas

- **Ventajas:** cada hito tiene versión exacta, tag inmutable, notas curadas y
  vínculo directo con el SHA; es compatible con el flujo de ramas aceptado.
- **Inconvenientes:** requiere preparar notas y comprobar que el artefacto
  activo corresponde al SHA etiquetado.
- **Coste de adopción y mantenimiento:** bajo mientras las releases sean selectivas.
- **Riesgos:** unas notas imprecisas pueden sobredimensionar el alcance; se evita
  declarando componentes incluidos, no incluidos y validación.

### No cambiar

Mantener solo promociones a `main` conserva estabilidad técnica, pero no crea
una referencia de producto ni una explicación pública de cada hito.

## Comparación

La alternativa A cubre el trabajo diario, que ya conserva Git. La B aporta
visibilidad pero no precisión. La C satisface la trazabilidad de producción sin
introducir CI/CD, una rama `release/*` permanente ni empaquetado adicional; se
limita a los hitos que ADR-0092 ya autoriza.

## Recomendación

**Opinión/recomendación:** adoptar la alternativa C. El primer tag será
`v1.0.0`, no `v1`, pues comunica una referencia SemVer exacta y coincide con la
versión inicial aceptada. La llegada de una app nativa no exige por sí sola un
cambio de major: el próximo número depende de la compatibilidad y alcance del
producto; los números de build de cada tienda siguen su ciclo independiente.

## Decisión del usuario

**Aceptada el 2026-09-06:** los hitos de producción o distribución se versionan
con un tag Git anotado e inmutable conforme a SemVer y una GitHub Release. El
primer hito es `v1.0.0`, definido como baseline productivo web.

El orden de promoción es:

1. ejecutar y revisar `make verify` en `develop`;
2. promocionar mediante `git merge --no-ff develop` en `main`;
3. crear el tag anotado sobre ese commit de merge;
4. publicar una GitHub Release con SHA, componentes incluidos, validaciones,
   limitaciones y ruta de rollback;
5. preparar y activar el artefacto de producción del SHA etiquetado;
6. avanzar `develop` hasta `main` antes de continuar el trabajo ordinario.

## Consecuencias

### Positivas

- `v1.0.0` permite recuperar y comparar un hito sin interpretar una secuencia
  de commits.
- La Release comunica qué se entregó y qué queda fuera, incluidas las apps
  nativas pendientes.
- El SHA del tag y el manifiesto de despliegue forman una cadena revisable.

### Negativas y deuda aceptada

- La persona operadora prepara notas y verifica tag y artefacto; no se
  automatiza todavía porque la frecuencia no justifica otro pipeline.
- La release web no reemplaza las versiones, firmas, revisiones ni publicación
  de iOS y Android.

## Validación

1. El tag es anotado, apunta al merge de `main` y no se mueve.
2. La GitHub Release enlaza tag e identifica SHA, alcance y límites.
3. El manifiesto del artefacto web activo contiene el SHA del tag.
4. `develop` vuelve a contener el merge de `main`.
5. Changelog, `make verify` y CI registran el hito y quedan en verde.

## Disparadores de revisión

- distribución de la primera app nativa;
- más de una persona promoviendo releases;
- necesidad de staging, selección de funcionalidades o CI/CD;
- un hotfix desde `main`;
- incompatibilidad de API o datos que requiera revisar SemVer.

## Documentación afectada

- [Decisiones](../governance/DECISIONS.md)
- [Contribución](../../CONTRIBUTING.md)
- [Despliegue](../operations/DEPLOYMENT.md)
- [Changelog](../../CHANGELOG.md)
- [Aprendizaje](../project/LEARNING.md)
