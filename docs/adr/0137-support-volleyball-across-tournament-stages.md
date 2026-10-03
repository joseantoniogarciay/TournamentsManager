# ADR-0137: Soportar voleibol en liga, eliminatoria y formato mixto

- **Estado:** Aceptado
- **Fecha:** 2026-10-02
- **Decisor:** Usuario, aceptación explícita del alcance y ambas políticas
- **Supera a:** ADR-0135, exclusivamente para ampliar los perfiles por sets

## Problema y contexto

El usuario ha elegido voleibol después de tenis de mesa y ha confirmado liga,
eliminatoria y formato mixto. Reutilizar solo el marcador agregado oculta los
parciales que determinan puntos y clasificación. La fase de liga exige además
adaptar retirada, desempates y siembra del cuadro a este deporte.

## Alternativas, coste y recomendación

- Agregado 3–0/3–1/3–2 sin parciales: coste bajo, pierde información necesaria
  para comparar cocientes de tantos y no valida la secuencia del encuentro.
- Perfil explícito con sets y clasificación propios: adopción media/alta,
  mantenimiento medio; conserva fases, transacciones e historial existentes.
- Motor configurable de reglamentos: coste alto, combinaciones y UI sin demanda.

La recomendación es el perfil explícito. La eliminatoria sola tiene coste medio,
pero el usuario ha elegido también ligas y mixto; no se reduce ese alcance.

## Alcance aceptado

- Perfil `volleyball`, con equipos, liga a una o dos vueltas, eliminatoria directa
  y liga más eliminatoria (tabla única o grupos) usando las fases existentes.
- Resultados por sets ordenados y agregado derivado; historial atómico.

## Decisión del usuario — aceptada el 2026-10-02

1. Partidos siempre al mejor de cinco: primeros cuatro sets a 25; quinto a 15.
   Cada set exige diferencia de dos y no se juega después de decidir el partido.
2. Clasificación de liga: victorias, puntos, cociente de sets y cociente de tantos.
   Ganar 3–0/3–1 concede 3 puntos, ganar 3–2 concede 2, perder 2–3 concede 1,
   otras derrotas 0. Cocientes se comparan exactamente, sin redondeo; con cero
   derrotas y valor positivo se consideran superiores a cocientes finitos; 0/0
   se trata como cero. Una igualdad exacta comparte posición, sin enfrentamiento
   directo añadido. En un corte de clasificación se usa la fase de desempate
   existente; sus partidos también son de voleibol y usan los mismos criterios.
3. Retirada durante liga: todos los partidos del equipo pasan a derrota 0–3 con
   tres sets 0–25; se conservan las instantáneas previas y se excluye el equipo de
   plazas de la siguiente fase como en los perfiles actuales. No se incorpora
   retirada en eliminatoria ni abandono parcial del partido.
4. Siembra entre grupos: se ordena por posición y los criterios de voleibol;
   la semilla de alta solo aporta orden estable después de resolver elegibilidad.

El usuario aceptó explícitamente las políticas de clasificación y retirada.
Este perfil es una regla de producto, no una declaración de soporte completo de
cualquier reglamento FIVB. La «i» de clasificación explica por deporte el tanteo,
puntos, columnas, orden y desempates, incluidos cocientes y retirada en voleibol.

## Validación realizada

- Sets a 25 y quinto a 15, deuce, victoria anticipada y rechazo de agregados solos.
- Clasificación 3–0 frente a 3–2, prioridades y cocientes con denominador cero.
- Retirada con parciales administrativos e historial original conservado.
- Liga, grupos, corte empatado, desempate y propagación a eliminatoria.
- Contrato generado, borradores de registro/login y UI localizada en cuatro idiomas.
- Respuestas HTTP seguras, diagnóstico cerrado y cancelación sin feedback.

Pruebas de dominio y PostgreSQL aprobadas, incluida finalización por cociente
de tantos, corrección y retirada con historial, grupos a dos vueltas y ciclo
de desempate repetido. Cliente comprobado con typecheck, contrato generado y
exportación; revisión visual web en escritorio y 390 px. Binarios nativos y
despliegue no ejecutados. La auditoría de dependencias detecta un aviso previo
de OpenTelemetry (GO-2026-6505), registrado en aprendizaje; no está resuelto
por este cambio funcional.

## Fuentes

- [FIVB: reglas 2025–2028, resultado y default](https://www.fivb.com/wp-content/uploads/2025/01/FIVB-Volleyball_Rules2025_2028-EN.pdf)

## Documentación afectada

Producto, API, modelo, diseño, observabilidad, decisiones y aprendizaje.
