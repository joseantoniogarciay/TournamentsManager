# ADR-0021: Usar CI informativa con puerta de calidad local

- **Estado:** Aceptado
- **Fecha:** 2026-07-25
- **Decisor:** Usuario, mediante aceptación explícita de la alternativa A ajustada
- **Propietario del análisis:** Asistente como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

El repositorio necesita repetir en un entorno limpio las verificaciones de
calidad y detectar diferencias respecto del equipo local. Al haber un solo
desarrollador, CI no debe exigir pull requests ni auto-revisiones que no añaden
independencia, ni convertirse en una barrera para publicar trabajo remoto.

## Contexto y restricciones

- ADR-0013 establece `develop` como rama de integración diaria y `main` como
  hito estable, sin ramas por feature como norma mientras trabaja una persona.
- `make verify` ya reúne formato, lint, tipos, pruebas, consistencia de módulos,
  build y vulnerabilidades; no modifica archivos.
- ADR-0019 exige evidencia proporcional al riesgo, no cobertura ni suites
  vacías.
- El repositorio es público. Los runners estándar alojados por GitHub son de uso
  gratuito para repositorios públicos; los runners mayores, artefactos y el
  posible paso a repositorio privado requieren vigilancia de coste.
- ADR-0006 y ADR-0017 prohíben secretos en contribuciones no confiables y
  reservan los secretos y OIDC para despliegues futuros.
- No hay todavía imágenes, artefactos de entrega, integración PostgreSQL en CI,
  tests E2E ni consumidores de reportes. Esta decisión no los inventa.

## Criterios de decisión

1. ejecutar la misma comprobación completa local y remota;
2. mantener el trabajo individual y los pushes remotos disponibles;
3. no fingir revisión independiente mediante una PR autoaprobada;
4. minimizar permisos, secretos, dependencias y coste de runners;
5. detectar regresiones de formato, build, dependencias y vulnerabilidades;
6. permitir endurecer la política al incorporar colaboración o despliegues.

## Alternativas

### Alternativa A — CI informativa con puerta local

Ejecutar `make verify` en GitHub Actions tras cada push a `develop` o `main` y
en pull requests cuando existan. El desarrollador ejecuta el mismo comando antes
de promover `develop` a `main`; los resultados de CI informan y no bloquean.

- **Ventajas:** misma fuente de verdad local y remota; mantiene el flujo de una
  persona; detecta diferencias de un runner Linux limpio; no depende de PRs ni
  de cuota para publicar.
- **Inconvenientes:** la promoción no queda técnicamente bloqueada si CI falla o
  está indisponible; la disciplina local sigue siendo necesaria.
- **Coste de adopción:** bajo.
- **Coste de mantenimiento:** bajo; un workflow y un comando versionado.
- **Riesgos:** ignorar un resultado rojo. Se mitiga con promoción por bloques,
  revisión del resultado y el disparador de revisión al incorporar colaboradores.

### Alternativa B — Checks obligatorios y pull requests protegidas

Requerir `verify` verde y una pull request para promover a `main` o integrar en
`develop`.

- **Ventajas:** GitHub impide fusionar un commit sin el check requerido y deja un
  historial de revisión.
- **Inconvenientes:** una auto-revisión no aporta independencia; una cuota o
  indisponibilidad de Actions puede bloquear la promoción aunque el entorno
  local esté verificado.
- **Coste de adopción:** medio.
- **Coste de mantenimiento:** medio; reglas, excepciones y resolución de checks
  atascados.
- **Riesgos:** burocracia para un único desarrollador y falsa sensación de
  revisión.

### Alternativa C — Calidad solo local

Documentar `make verify` y no ejecutar CI.

- **Ventajas:** coste y configuración mínimos.
- **Inconvenientes:** no hay entorno limpio independiente ni resultado visible
  en remoto; las divergencias de sistema operativo y dependencias se detectan
  tarde.
- **Coste de adopción:** mínimo.
- **Coste de mantenimiento:** bajo.
- **Riesgos:** confiar solo en el estado de una máquina.

## Comparación

La B es adecuada cuando una revisión de otra persona o un control de promoción
aporta evidencia adicional. Con una sola persona, esa evidencia no existe y el
bloqueo añade fricción. La C elimina una comprobación útil y barata para un
repositorio público. La A conserva la reproducción en Linux y la visibilidad de
los resultados sin sustituir el juicio del desarrollador ni impedir el trabajo.

## Recomendación

**Opinión/recomendación:** alternativa A. Es la solución mínima suficiente para
el flujo actual y mantiene una ruta explícita de endurecimiento basada en
evidencia.

## Decisión del usuario

**Aceptada:** alternativa A, con estas reglas:

- GitHub Actions ejecutará `make verify` en un runner estándar Ubuntu alojado
  por GitHub en cada push a `develop` y `main`, y en cada pull request que exista;
- el resultado de CI es informativo, no un check obligatorio de protección;
- antes de promover un bloque de `develop` a `main`, el desarrollador ejecuta
  `make verify` localmente y revisa el resultado remoto cuando esté disponible;
- no se exigen pull requests ni aprobaciones mientras haya un único
  desarrollador; la promoción puede ser directa desde `develop` a `main`;
- `main` y `develop` no permiten force-push ni borrado;
- los workflows usan permisos mínimos (`contents: read`), no reciben secretos ni
  acceso de despliegue, y usan `pull_request`, nunca `pull_request_target`;
- las acciones de terceros se fijarán a SHA completa y se revisarán como
  dependencias de la cadena de suministro;
- no se generan artefactos, caches explícitas, matrices, CodeQL, cobertura,
  escaneo de imágenes, integración PostgreSQL en CI ni tests de dispositivos
  hasta que un componente y un riesgo concreto los justifiquen;
- se cancelará una ejecución anterior en curso para la misma rama o pull request
  cuando un nuevo commit la sustituya.

## Consecuencias

### Positivas

- `make verify` es la fuente de verdad común entre estación local y CI.
- Se detectan diferencias de entorno sin depender de una auto-revisión.
- El uso de runners estándar de GitHub es gratuito mientras el repositorio sea
  público y no se empleen runners mayores.
- Los workflows no amplían la superficie de secretos o despliegue.

### Negativas y deuda aceptada

- Un resultado rojo no bloquea técnicamente la promoción; la responsabilidad de
  detenerla sigue en el desarrollador.
- No existe revisión independiente ni obligación de checks hasta que se forme un
  equipo o aparezca un riesgo que lo requiera.
- No se conserva aún evidencia descargable de tests o builds porque no hay un
  consumidor de esos artefactos.

## Validación

La implementación deberá demostrar que:

- un push a `develop` y uno a `main` ejecutan `make verify` en Linux limpio;
- un fallo intencionado de formato o de tipos falla en local y en CI;
- el workflow no usa secretos, permisos de escritura, environments ni triggers
  privilegiados;
- un cambio nuevo cancela la ejecución anterior de la misma rama;
- no se publican artefactos ni se crean costes de almacenamiento innecesarios;
- la documentación de contribución refleja promoción directa y calidad local.

## Disparadores de revisión

- se incorpora un colaborador que pueda hacer revisión independiente;
- un resultado rojo se ignora o llega una regresión a `main`;
- el repositorio pasa a privado, se usan runners mayores o aparece coste medido;
- el tiempo de `make verify` supera un presupuesto de feedback acordado;
- integración PostgreSQL, E2E, imágenes, releases o despliegues requieren jobs,
  artefactos, secretos o gates específicos;
- un incidente de seguridad exige checks obligatorios o protección reforzada.

## Documentación afectada

- [CONTRIBUTING.md](../../CONTRIBUTING.md)
- [DEVELOPMENT.md](../engineering/DEVELOPMENT.md)
- [TESTING.md](../engineering/TESTING.md)
- [SECURITY.md](../engineering/SECURITY.md)
- [TECHNICAL_BASELINE.md](../governance/TECHNICAL_BASELINE.md)
- [DECISIONS.md](../governance/DECISIONS.md)
- [DECISIONS_TO_REVISIT.md](../governance/DECISIONS_TO_REVISIT.md)
- [LEARNING.md](../project/LEARNING.md)
- [CHANGELOG.md](../../CHANGELOG.md)

## Fuentes técnicas

- [GitHub Actions: facturación](https://docs.github.com/en/billing/concepts/product-billing/github-actions)
- [GitHub Actions: uso seguro](https://docs.github.com/en/actions/reference/security/secure-use)
- [GitHub Actions: sintaxis de workflows](https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax)
- [GitHub: ramas protegidas](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches/about-protected-branches)
