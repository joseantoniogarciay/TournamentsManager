# ADR-0042: Cancelar ligas sin motivo obligatorio ni avisos automáticos

- **Estado:** Aceptado
- **Fecha:** 2026-07-26
- **Decisor:** Usuario
- **Propietario del análisis:** Codex
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

El creador puede cancelar una liga publicada o en curso. Hay que definir si esa
acción requiere explicar un motivo y cómo conocen el cambio quienes la seguían o
la consultaban mediante enlace.

## Contexto y restricciones

- Solo el creador cancela desde `publicado` o `en_curso` y la cancelación conserva
  datos (ADR-0040).
- Las ligas son no listadas y se consultan por enlace (ADR-0033).
- Los seguidores autenticados recuperan sus ligas guardadas (ADR-0034).
- Email y notificaciones push están fuera del primer corte.

## Criterios de decisión

1. mantener la cancelación rápida y reversible solo mediante decisiones futuras;
2. preservar la trazabilidad del estado sin obligar a introducir texto;
3. no añadir infraestructura de avisos al primer vertical slice;
4. informar de forma honesta a quien vuelva a consultar la liga.

## Alternativas

### Alternativa A — Sin motivo obligatorio ni notificaciones automáticas

El creador cancela sin texto requerido. El enlace sigue mostrando la liga como
cancelada y los seguidores ven ese estado al entrar en sus ligas guardadas.

- **Ventajas:** flujo mínimo; no exige moderación, plantillas ni infraestructura
  de entrega.
- **Inconvenientes:** nadie recibe un aviso activo y no se conoce la causa.
- **Coste de adopción:** bajo.
- **Coste de mantenimiento:** bajo.
- **Riesgos:** una persona puede descubrir tarde la cancelación.

### Alternativa B — Motivo obligatorio visible

- **Ventajas:** aporta contexto a seguidores y visitantes.
- **Inconvenientes:** obliga a validar, moderar y retener texto que puede ser
  sensible o inapropiado.
- **Coste de adopción:** moderado.
- **Coste de mantenimiento:** moderado.
- **Riesgos:** explicaciones vacías, datos personales o conflictos públicos.

### Alternativa C — Avisos por email y push

- **Ventajas:** comunicación inmediata.
- **Inconvenientes:** requiere preferencias, tokens de dispositivo, entrega,
  reintentos, deep links y observabilidad de notificaciones.
- **Coste de adopción:** alto.
- **Coste de mantenimiento:** alto.
- **Riesgos:** fallos de entrega, spam y expansión prematura de infraestructura.

### No cambiar

La cancelación mantendría un efecto ambiguo para los enlaces y las ligas seguidas.

## Comparación

La B y la C añaden comunicación útil, pero no necesaria para validar la gestión
de una liga cerrada. La A conserva el estado visible y el historial sin ampliar
el producto con texto libre o notificaciones.

## Recomendación

**Recomendación:** alternativa A, sin motivo obligatorio ni avisos automáticos.

## Decisión del usuario

**Aceptada el 2026-07-26:** cancelar una liga no exige motivo. La liga conserva
sus datos, sigue consultable mediante enlace mostrando estado `cancelado` y los
seguidores verán ese estado al volver a su lista. No se envían email ni push en
el primer corte.

**Aclaración registrada el 2026-08-10:** antes de ejecutar la cancelación, la
pantalla de detalle pide una confirmación explícita mediante el diálogo
compartido definido por el sistema de diseño. Al confirmarla, la liga pasa a
`cancelado`, que es terminal en este corte: ya no admite registrar ni corregir
resultados, ni finalizarse. La restauración por el creador se pospone para una
decisión e incremento futuros; no forma parte del comportamiento actual.

Las vistas que ya proyectan el estado de una liga —su detalle y las cajas desde
las que se accede a ella— actualizan esa misma proyección a «Liga cancelada».
No se añade una etiqueta ni una superficie de estado adicional.

## Consecuencias

### Positivas

- Cancelar no queda bloqueado por un formulario ni una entrega externa.
- El estado final permanece visible y trazable.
- El primer slice no necesita una plataforma de notificaciones.
- La acción destructiva requiere una confirmación antes de cambiar el estado.

### Negativas y deuda aceptada

- No hay explicación de la cancelación.
- Los seguidores no reciben comunicación proactiva.
- Una cancelación no se puede restaurar todavía.

## Validación

- El creador cancela sin aportar texto.
- La pantalla de detalle no ejecuta la cancelación hasta la confirmación
  explícita de la persona creadora.
- El enlace de una liga cancelada muestra el estado, sin habilitar cambios.
- Las ligas seguidas muestran la cancelación al consultarse.
- No se emite email ni push por la acción.

## Disparadores de revisión

- Necesidad de explicar cancelaciones a grupos.
- Necesidad de avisos urgentes o notificaciones configurables.
- Requisitos de moderación o auditoría adicionales.
- Restauración de una liga cancelada por su creador.

## Documentación afectada

- [PRODUCT.md](../project/PRODUCT.md)
- [DESIGN_SYSTEM.md](../engineering/DESIGN_SYSTEM.md)
- [ROADMAP.md](../project/ROADMAP.md)
- [LEARNING.md](../project/LEARNING.md)
- [DECISIONS.md](../governance/DECISIONS.md)
