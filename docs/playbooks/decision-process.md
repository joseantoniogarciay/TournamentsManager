# Playbook: tomar una decisión técnica

## Cuándo usarlo

Cuando una elección pueda afectar arquitectura, datos, seguridad, operación,
coste, portabilidad o convenciones transversales.

## Paso 1 — Formular el problema

Describe el resultado, la evidencia y qué ocurre si no se decide ahora. Evita
preguntas del tipo “¿qué herramienta es mejor?” sin contexto.

## Paso 2 — Establecer criterios

Ordena criterios y define cómo se observarán. Incluye restricciones y supuestos.

## Paso 3 — Explorar alternativas

Compara al menos dos opciones viables. Incluye mantener el estado actual cuando
sea realista. Descarta pronto solo por una restricción explícita.

## Paso 4 — Analizar coste completo

Para cada opción cubre aprendizaje, implementación, mantenimiento, actualización,
seguridad, diagnóstico, recuperación, salida y dependencia de proveedor.

## Paso 5 — Separar etiquetas

- **Hecho/evidencia:** verificable y con fuente.
- **Estándar:** práctica ampliamente aceptada, con su contexto.
- **Opinión:** juicio razonado.
- **Recomendación:** opción preferida por el analista.
- **Decisión:** elección explícita del usuario.

## Paso 6 — Recomendar y pedir decisión

Expón la recomendación, la principal desventaja aceptada y qué evidencia podría
cambiarla. No implementes mientras el estado sea `Propuesto`.

## Paso 7 — Registrar y propagar

Completa el ADR, el índice, los documentos especializados y el registro de
revisión. Define prueba o experimento de validación.

## Checklist

- [ ] El problema está respaldado por una necesidad actual.
- [ ] Las alternativas son realmente viables.
- [ ] El coste de mantenimiento es explícito.
- [ ] La recomendación no se presenta como decisión.
- [ ] El usuario decidió.
- [ ] El ADR y documentos están enlazados.
- [ ] Existen validación y disparadores de revisión.
