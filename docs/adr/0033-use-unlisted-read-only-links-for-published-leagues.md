# ADR-0033: Usar enlaces no listados de solo lectura para ligas publicadas

- **Estado:** Aceptado
- **Fecha:** 2026-07-26
- **Decisor:** Usuario
- **Propietario del análisis:** Codex
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

Una liga publicada debe poder compartirse con personas ajenas al organizador
antes de decidir invitaciones, participantes o administradores delegados. Hay
que definir si se descubre públicamente, se comparte mediante enlace o queda
exclusivamente privada.

## Contexto y restricciones

- El producto inicial se dirige a grupos cerrados y el formato inicial es una
  liga de fútbol.
- Publicar valida la liga, genera sus emparejamientos y fija su composición,
  conforme a ADR-0032.
- Una cuenta verificada es necesaria para persistir y publicar; un invitado puede
  preparar un borrador local, conforme a ADR-0031.
- Participantes, invitaciones, varios administradores y gestión de resultados
  están pendientes y no se infieren de la visibilidad.
- La visibilidad debe ser una propiedad distinta del ciclo de vida de la liga.

## Criterios de decisión

1. permitir compartir una liga publicada con la mínima fricción;
2. evitar descubrimiento público y sus necesidades de búsqueda y moderación;
3. no conceder permisos por consultar una liga;
4. conservar una base que admita invitaciones y administración delegada después;
5. no confundir una comodidad de interfaz con una transición de dominio distinta.

## Alternativas

### Alternativa A — Privada solo para el organizador

- **Ventajas:** autorización y superficie expuesta mínimas.
- **Inconvenientes:** impide compartir la liga antes de crear invitaciones.
- **Coste de adopción:** bajo.
- **Coste de mantenimiento:** bajo.
- **Riesgos:** retrasa la utilidad básica de informar al grupo.

### Alternativa B — No listada mediante enlace de solo lectura

La liga publicada se consulta mediante un enlace con identificador aleatorio no
predecible. Quien lo abre puede verla sin sesión, pero no recibe permisos ni una
relación de participación.

- **Ventajas:** permite compartir de inmediato sin registro ni sistema de
  invitaciones; evita una superficie de descubrimiento público.
- **Inconvenientes:** quien reciba el enlace puede reenviarlo; no identifica a
  quien consulta la liga.
- **Coste de adopción:** bajo a moderado, por la generación y protección del
  identificador de compartición.
- **Coste de mantenimiento:** bajo; la rotación o revocación del enlace se
  decidirá si el producto la necesita.
- **Riesgos:** tratar erróneamente el secreto del enlace como una autorización
  de edición o usar identificadores secuenciales fáciles de enumerar.

### Alternativa C — Pública y descubrible

- **Ventajas:** máxima visibilidad y base para explorar torneos.
- **Inconvenientes:** requiere búsqueda, indexación, moderación, controles de
  privacidad y una decisión de producto que el alcance inicial no necesita.
- **Coste de adopción:** alto.
- **Coste de mantenimiento:** alto.
- **Riesgos:** exponer torneos de grupos cerrados y adelantar producto público.

### No cambiar

No decidir la visibilidad mantiene ambigua la semántica de publicar y bloquea la
superficie de lectura del primer vertical slice.

## Comparación

La alternativa A conserva la máxima privacidad, pero no permite el uso de
compartición que se busca. La C excede el alcance de grupos cerrados. La B ofrece
consulta directa y acotada: el enlace sirve para localizar el recurso, mientras
la autorización sigue reservada al organizador y a futuras invitaciones.

## Recomendación

**Recomendación:** alternativa B, enlaces no listados de solo lectura. Es la
solución mínima suficiente para compartir ligas sin diseñar todavía un sistema de
incorporación ni una plataforma pública.

## Decisión del usuario

**Aceptada el 2026-07-26:** adoptar la alternativa B.

- Una liga en `publicado`, `en_curso` o `finalizado` tiene visibilidad no
  listada y se puede consultar mediante un enlace compartible.
- El enlace contiene un identificador aleatorio no predecible; no expone ni
  deriva de un identificador secuencial interno.
- La consulta mediante enlace es de solo lectura y no exige cuenta.
- Consultar mediante enlace no crea un participante, invitado ni administrador,
  ni habilita edición o registro de resultados.
- Una liga en `borrador` no es accesible por enlace. La interfaz puede ofrecer
  “crear y publicar” cuando los datos son válidos; es la misma transición de
  dominio que publicar un borrador, no un estado adicional.
- Invitaciones, varios administradores y permisos para resultados se decidirán
  en un ADR posterior.

## Consecuencias

### Positivas

- El organizador puede informar al grupo desde el primer vertical slice.
- La lectura compartida no requiere crear cuentas ni administrar invitaciones.
- Los permisos futuros pueden añadirse sin reinterpretar una visita como una
  relación con la liga.

### Negativas y deuda aceptada

- El enlace puede reenviarse; no es una garantía de confidencialidad frente a
  quienes lo reciban.
- No se decide aún la revocación, rotación, caducidad ni analítica de enlaces.
- La implementación debe proteger el identificador frente a enumeración y evitar
  incluir datos que no correspondan a una lectura pública no listada.

## Validación

- Un borrador no se consulta con un enlace compartible.
- Una liga publicada se muestra en modo lectura sin sesión mediante su enlace.
- Un enlace válido no autoriza cambios, resultados ni administración.
- Los identificadores internos consecutivos no permiten deducir enlaces válidos.
- Una acción de interfaz de crear y publicar cumple las mismas validaciones que
  publicar un borrador.

## Disparadores de revisión

- Se necesita revocar o rotar enlaces compartidos.
- Se introducen datos sensibles que no deban estar disponibles para quien tenga
  el enlace.
- El producto requiere búsqueda pública o una audiencia abierta.
- Invitaciones o administradores delegados necesitan distinguir más niveles de
  acceso.

## Documentación afectada

- [PRODUCT.md](../project/PRODUCT.md)
- [ROADMAP.md](../project/ROADMAP.md)
- [LEARNING.md](../project/LEARNING.md)
- [DECISIONS.md](../governance/DECISIONS.md)
