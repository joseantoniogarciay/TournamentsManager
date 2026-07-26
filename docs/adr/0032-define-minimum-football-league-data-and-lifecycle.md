# ADR-0032: Definir los datos mínimos y ciclo de vida de una liga de fútbol

- **Estado:** Superado parcialmente
- **Fecha:** 2026-07-26
- **Decisor:** Usuario
- **Propietario del análisis:** Codex
- **Supera a:** Ninguno
- **Superado por:** ADR-0040, en la generación de partidos y la edición posterior
  a publicar

## Problema

El primer vertical slice debe permitir preparar, persistir y publicar una liga
de fútbol. Sin definir qué constituye una liga y cuáles son sus transiciones, no
se puede validar el borrador, proteger sus cambios ni diseñar después el esquema
de datos y los contratos API.

## Contexto y restricciones

- El fútbol y el formato de liga son el primer alcance aceptado.
- El organizador autenticado y verificado es inicialmente el único administrador
  y crea los equipos.
- Un invitado puede preparar el borrador localmente; solo una cuenta verificada
  puede persistirlo y publicarlo, según ADR-0031.
- Visibilidad, participantes, invitaciones, fechas, resultados, clasificación,
  bajas y administración delegada siguen fuera de esta decisión.
- La lógica de ciclo de vida pertenece al dominio y no debe depender de
  PostgreSQL, HTTP ni de la interfaz de cliente.

## Criterios de decisión

1. representar una liga real sin adelantar un motor deportivo genérico;
2. mantener pequeño el primer vertical slice;
3. fijar reglas que permitan validación y trazabilidad;
4. impedir cambios que invaliden unos emparejamientos ya publicados;
5. dejar separadas y revisables las decisiones de calendario, resultados y
   visibilidad.

## Alternativas

### Alternativa A — Registro mínimo sin partidos

Una liga contiene solo nombre y equipos y transita de `borrador` a `publicado`.

- **Ventajas:** menor modelo y menor trabajo inicial.
- **Inconvenientes:** el formato liga es solo una etiqueta; no existe la
  estructura competitiva que deberá gestionar el producto.
- **Coste de adopción:** bajo.
- **Coste de mantenimiento:** bajo al inicio, pero obliga a rediseñar la
  publicación al introducir partidos.
- **Riesgos:** confundir un directorio de equipos con una competición.

### Alternativa B — Liga estructural mínima

La liga contiene sus equipos, reglas iniciales y emparejamientos. Al publicarla
se valida y congela la composición, se generan los partidos sin fecha ni hora y
su ciclo es `borrador → publicado → en_curso → finalizado`, con `cancelado` como
estado terminal alternativo.

- **Ventajas:** expresa el formato de negocio desde el primer corte; protege la
  consistencia de los emparejamientos; aplaza decisiones operativas no necesarias.
- **Inconvenientes:** requiere modelar partidos antes de poder registrar sus
  resultados.
- **Coste de adopción:** moderado.
- **Coste de mantenimiento:** bajo a moderado; las reglas de generación y las
  transiciones deben probarse.
- **Riesgos:** asumir que una vuelta y la puntuación inicial sirven para siempre.

### Alternativa C — Liga operativa completa

Además de la estructura, incorpora fechas, resultados, clasificación, bajas y
replanificación desde el primer vertical slice.

- **Ventajas:** ofrece una competición operable de extremo a extremo.
- **Inconvenientes:** introduce decisiones aún abiertas sobre calendario,
  autoridad de resultados y excepciones deportivas.
- **Coste de adopción:** alto.
- **Coste de mantenimiento:** alto por reglas, casos límite y soporte.
- **Riesgos:** sobreingeniería y diseño especulativo.

### No cambiar

Mantener la pregunta abierta bloquea el modelo de dominio, las validaciones y el
contrato de publicación del primer vertical slice.

## Comparación

La alternativa A es más pequeña, pero no satisface el requisito de representar
una liga. La C satisface más necesidades, a costa de adelantar decisiones de
producto no aceptadas. La B fija únicamente la estructura y las invariantes que
la publicación necesita; deja los resultados y el calendario detallado para
decisiones posteriores.

## Recomendación

**Recomendación:** alternativa B, liga estructural mínima. Es la mínima solución
suficiente para que publicar una liga tenga significado de negocio, sin convertir
el primer corte en un sistema completo de competición.

## Decisión del usuario

**Aceptada el 2026-07-26:** adoptar la alternativa B.

Una liga de fútbol contiene como mínimo:

- identificador, nombre, deporte `fútbol`, formato `liga`, organizador, estado y
  marcas temporales;
- equipos, cada uno con identificador, nombre único dentro de la liga y vínculo
  con ella;
- reglas iniciales inmutables: una vuelta, tres puntos por victoria, uno por
  empate y cero por derrota;
- partidos generados, con jornada u orden, equipo local, visitante y estado
  pendiente, sin fecha, hora, marcador ni clasificación en este corte.

El ciclo de vida es:

```text
borrador ──publicar──> publicado ──iniciar──> en_curso ──cerrar──> finalizado
    │                     │                    │
    └─────────────────────┴────cancelar────────┴───────────────> cancelado
```

- Solo en `borrador` se pueden añadir, renombrar o eliminar equipos.
- `publicar` requiere una cuenta verificada, nombre válido y al menos dos
  equipos; genera los partidos una sola vez y fija la composición.
- `en_curso` y `finalizado` reservan el lugar para la futura gestión de
  resultados, sin autorizarla ni definirla todavía.
- `cancelado` conserva trazabilidad y no borra la liga ni sus equipos o
  partidos.

La visibilidad es una propiedad distinta del ciclo de vida y se decidirá aparte.

## Consecuencias

### Positivas

- El primer vertical slice puede validar qué se publica y mostrar una estructura
  de competición estable.
- Los emparejamientos no cambian silenciosamente después de publicar.
- El modelo conserva espacio explícito para resultados y clasificación futuros.

### Negativas y deuda aceptada

- La composición no podrá alterarse tras publicar; resolver bajas o cambios de
  equipo requiere una decisión posterior.
- Las reglas de una vuelta y puntuación 3-1-0 serán inmutables en el primer
  corte, aunque otros formatos puedan requerir configuración futura.
- Falta decidir quién y cómo registra o confirma resultados antes de implementar
  la transición a `en_curso` con actividad deportiva real.

## Validación

- Un borrador admite cambios de equipos y no tiene partidos persistidos.
- La publicación rechaza menos de dos equipos, nombres inválidos o una cuenta no
  verificada.
- Publicar una liga válida crea exactamente un emparejamiento por cada pareja de
  equipos y bloquea cambios de composición.
- Solo las transiciones declaradas alcanzan `en_curso`, `finalizado` o
  `cancelado`; cancelar conserva los datos.
- Ninguna transición ni dato de dominio requiere conocer HTTP o PostgreSQL.

## Disparadores de revisión

- Se requiere ida y vuelta, distinta puntuación o desempates configurables.
- Se necesita modificar equipos después de publicar.
- La visibilidad, las invitaciones o la incorporación exigen separar publicación
  de disponibilidad para jugar.
- Resultados, calendario o clasificación revelan que los estados elegidos no
  representan el flujo real.

## Documentación afectada

- [PRODUCT.md](../project/PRODUCT.md)
- [ROADMAP.md](../project/ROADMAP.md)
- [LEARNING.md](../project/LEARNING.md)
- [DECISIONS.md](../governance/DECISIONS.md)
