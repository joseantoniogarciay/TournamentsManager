# ADR-0004: Confirmar la base técnica antes del diseño de producto

- **Estado:** Aceptado
- **Fecha:** 2026-07-23
- **Decisor:** Usuario, mediante instrucción explícita
- **Supera a:** Orden de decisión inicial de `SYSTEM_OPTIONS.md`
- **Superado por:** Ninguno

## Problema

El mapa inicial mezclaba decisiones de producto —formato, participantes,
visibilidad e incorporación— con decisiones de plataforma. El usuario quiere
confirmar primero la base técnica sobre la que se desarrollará y aprenderá.

## Contexto y restricciones

Ya existen requisitos de contexto aceptados: producto de torneos, web y mobile,
acceso invitado, cuentas y fútbol como primer deporte. No se quiere diseñar aún
el comportamiento detallado del torneo.

## Alternativas

### Producto primero

Permite que cada tecnología responda a casos de uso concretos y reduce diseño
técnico especulativo. Retrasa la confirmación del entorno y del stack de
aprendizaje.

### Técnica primero

Ordena repositorio, aplicaciones, contratos, identidad, datos, tooling y operación
antes del dominio. Aporta una base pedagógica clara, pero exige evitar decisiones
que dependan de requisitos todavía desconocidos.

### Evolución intercalada

Alterna un requisito de producto con la mínima decisión técnica necesaria. Reduce
supuestos, aunque puede dificultar una visión completa del itinerario técnico.

## Decisión del usuario

Confirmar la base técnica antes de tomar nuevas decisiones de negocio o producto.

## Consecuencias

- Las preguntas funcionales de `PRODUCT.md` quedan pausadas.
- Las recomendaciones técnicas se decidirán una a una y terminarán en ADR.
- No se creará scaffolding ni código para una alternativa no aceptada.
- Las decisiones técnicas declararán qué supuestos de producto necesitan.
- Si una decisión no puede tomarse responsablemente sin un requisito funcional,
  se aplaza en vez de inventar el requisito.

## Validación

El gate termina cuando `TECHNICAL_BASELINE.md` no contiene decisiones bloqueantes
en estado pendiente y cada elección importante tiene ADR aceptado.

## Disparadores de revisión

- Una elección técnica depende de una regla de producto todavía desconocida.
- La base se convierte en arquitectura especulativa.
- Un experimento técnico invalida una decisión aceptada.

## Documentación afectada

- [TECHNICAL_BASELINE.md](../../TECHNICAL_BASELINE.md)
- [ROADMAP.md](../../ROADMAP.md)
- [SYSTEM_OPTIONS.md](../../SYSTEM_OPTIONS.md)
- [PRODUCT.md](../../PRODUCT.md)
