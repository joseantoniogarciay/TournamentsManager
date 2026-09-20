# ADR-0133: Componer liga y eliminatoria en un mismo torneo

- **Estado:** Aceptado
- **Fecha:** 2026-09-20
- **Decisor:** Usuario, mediante aceptación explícita de la propuesta completa
- **Propietario del análisis:** Asistente como mentor técnico
- **Supera a:** ADR-0122 y ADR-0123, exclusivamente en el aplazamiento del
  formato mixto y de sus reglas de clasificación
- **Superado por:** Ninguno

## Problema

Una liga recompensa la regularidad y produce una clasificación útil, pero por
sí sola no ofrece el cierre de una eliminatoria. El producto necesita permitir
que la clasificación de una liga general o de varios grupos alimente un cuadro
posterior sin que la persona organizadora tenga que reconstruir otro torneo ni
elegir manualmente sus participantes.

## Contexto y restricciones

- ADR-0122 adopta `tournament` como recurso raíz y ADR-0123 ya modela sus
  partidos dentro de fases ordenadas.
- Liga y eliminatoria directa existen como formatos independientes para fútbol
  y baloncesto.
- La configuración se congela al iniciar; los equipos no cambian después.
- El backend es autoridad de clasificación, avance y emparejamientos.
- El formato mixto debe explicar antes de empezar cuántos equipos necesita y
  rechazar una composición incompleta.
- No se incorporan bombos, sorteos configurables, mejores terceros, repesca,
  doble eliminación, ida y vuelta en el cuadro ni un lenguaje genérico de
  reglas.

## Criterios de decisión

1. que la configuración sea comprensible y validable antes de empezar;
2. que cada plaza de la eliminatoria proceda de una clasificación reproducible;
3. que los mejores clasificados queden separados de forma deportiva en el
   cuadro;
4. que una corrección no pueda cambiar participantes de una fase ya iniciada;
5. reutilizar las fases aceptadas sin construir un motor genérico.

## Alternativas

### Alternativa A — Un formato monolítico específico

- **Ventajas:** una única operación podría generar toda la competición.
- **Inconvenientes:** duplica reglas ya presentes en liga y eliminatoria y
  mezcla sus ciclos de vida.
- **Coste de adopción:** medio.
- **Coste de mantenimiento:** alto por ramas específicas en resultados,
  clasificación y cierre.
- **Riesgos:** que el formato mixto diverja de los formatos independientes.

### Alternativa B — Componer dos fases existentes

- **Ventajas:** la primera fase conserva sus reglas de liga y la segunda las de
  eliminación directa; la transición tiene una frontera explícita.
- **Inconvenientes:** exige configuración de clasificación, grupos y estado por
  fase, además de una acción para confirmar la transición.
- **Coste de adopción:** medio o alto por afectar dominio, datos, contrato y UI.
- **Coste de mantenimiento:** bajo o medio; cada fase mantiene su política.
- **Riesgos:** comparar grupos de forma injusta si no tienen el mismo tamaño o
  distinto número de partidos.

### Alternativa C — Motor configurable de competiciones

- **Ventajas:** permitiría combinaciones arbitrarias de fases y reglas.
- **Inconvenientes:** requiere un lenguaje de configuración, validación de
  grafos y una interfaz difícil de explicar sin casos reales.
- **Coste de adopción y mantenimiento:** alto.
- **Riesgos:** sobreingeniería y combinaciones deportivas incoherentes.

### No cambiar

Liga y eliminatoria seguirían siendo torneos separados y la organizadora
tendría que trasladar manualmente equipos y siembra.

## Comparación

La alternativa A entrega el caso concreto a costa de duplicar reglas. La C
optimiza una variabilidad todavía desconocida. La B aprovecha la frontera
`Tournament -> Stage -> Match` ya aceptada y añade solo la configuración y la
transición necesarias para este caso.

## Recomendación

**Opinión/recomendación:** alternativa B, con una primera fase de liga general o
por grupos y una segunda fase de eliminatoria directa sin pases automáticos.

## Decisión del usuario

**Aceptada el 2026-09-20:** adoptar la alternativa B con estas reglas:

- la opción visible es «Liga + eliminatoria» y compone una fase `league` seguida
  de una fase `single_elimination`;
- la liga general permite una o dos vueltas y un número total de clasificados;
- la liga por grupos permite elegir número de grupos, una o dos vueltas dentro
  de cada grupo y número de clasificados por grupo;
- los grupos tienen el mismo número de equipos y al menos un equipo de cada
  grupo queda sin clasificar;
- la composición inicial de grupos usa distribución serpentina conforme al
  orden de siembra persistido;
- el total de clasificados es una potencia de dos entre 2 y 64; el formato mixto
  no genera _byes_;
- si faltan equipos, los grupos no quedan equilibrados o la clasificación no
  forma un cuadro completo, el torneo no empieza y la interfaz explica la
  cantidad necesaria; el backend repite autoritativamente la validación;
- la liga general siembra por su clasificación; en grupos se ordenan primero
  las posiciones obtenidas dentro de cada grupo y después rendimiento,
  criterios deportivos vigentes y orden de siembra persistido como último
  desempate;
- la primera ronda enfrenta mejor contra peor y el árbol separa los cabezas de
  serie para que 1 y 2 solo puedan coincidir en la final;
- al terminar todos los partidos de liga, la organizadora confirma «Iniciar
  eliminatoria»; esa acción congela la primera fase, fija clasificados y
  emparejamientos e inicia la segunda;
- los resultados de la primera fase se pueden corregir antes de esa
  confirmación, pero quedan congelados después.

**Ampliación aceptada el 2026-09-20 tras revisar casos extremos:**

- un equipo retirado conserva su fila histórica, pero queda fuera de la
  clasificación elegible; su plaza pasa al siguiente equipo activo de la tabla
  general o del mismo grupo;
- si no quedan suficientes equipos activos para llenar todas las plazas, la
  transición se rechaza sin reconstruir grupos ni crear un cuadro parcial;
- una vez iniciado el cuadro, el cliente permite alternar entre los partidos de
  liga y eliminatorias;
- quien no sea propietario ve la transición preparada como una acción
  deshabilitada y sabe que debe confirmarla la persona propietaria vigente.

## Consecuencias

### Positivas

- La regularidad de la liga determina objetivamente acceso y siembra del cuadro.
- Liga, grupos y eliminatoria conservan límites y estados visibles.
- Una configuración imposible se detecta antes de crear partidos.
- No se necesitan equipos temporales ni selección manual de clasificados.

### Negativas y deuda aceptada

- La proyección pública deberá distinguir clasificaciones por grupo y mostrar
  las dos fases sin confundir sus partidos.
- El orden de siembra es un último desempate determinista; no se añaden todavía
  partidos de desempate ni sorteo posterior.
- No se evita expresamente una revancha del mismo grupo en la primera ronda si
  resulta del orden global aceptado.
- La implementación requiere una operación contractual nueva para iniciar la
  segunda fase y una migración de datos, incorporadas en el incremento que
  ejecuta esta decisión.

## Evidencia de implementación

- El contrato discrimina las dos configuraciones mixtas y expone la transición
  explícita de la segunda fase.
- La migración `00011` persiste configuración, pertenencia, grupo y siembra por
  fase sin convertirlos en conceptos genéricos fuera del caso aceptado.
- Las pruebas de dominio y PostgreSQL recorren configuración inválida, grupos
  equilibrados e impares, una y dos vueltas, clasificación, retiradas, cuadro
  completo, concurrencia y congelación de la liga en fútbol y baloncesto.
- El cliente valida y explica la composición necesaria, separa grupos, permite
  alternar fases, informa de la espera de la propietaria y solicita confirmación
  antes de iniciar la eliminatoria.

## Validación

- Las pruebas de dominio cubren configuraciones válidas e inválidas, grupos
  equilibrados, clasificación exacta y siembra del cuadro.
- Persistencia y contrato conservan dos fases ordenadas y no crean el cuadro
  antes de la confirmación.
- Una corrección anterior a la transición puede cambiar la clasificación; tras
  iniciarse la eliminatoria se rechaza sin modificar datos.
- Cliente y backend coinciden en las precondiciones, pero una petición directa
  inválida sigue siendo rechazada por el servidor.
- Las salidas de inicio y transición recorren éxito, validación, permiso,
  conflicto, límites técnicos y cancelación en el inventario de observabilidad.

## Disparadores de revisión

- Se solicitan mejores terceros, grupos de distinto tamaño o cruces prefijados
  entre grupos.
- Se necesitan bombos, sorteos manuales o restricciones para evitar revancha de
  grupo.
- Un deporte requiere series, ida y vuelta, repesca o desempate mediante otro
  partido.
- La composición deja de poder expresarse honestamente como dos fases
  consecutivas.

## Documentación afectada

- [Producto](../project/PRODUCT.md)
- [API](../engineering/API.md)
- [Modelo inicial de datos](../engineering/INITIAL_DATA_MODEL.md)
- [Observabilidad](../operations/OBSERVABILITY.md)
- [Decisiones](../governance/DECISIONS.md)
- [Aprendizaje](../project/LEARNING.md)
