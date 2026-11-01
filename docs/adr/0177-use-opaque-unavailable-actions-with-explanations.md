# ADR-0177: Acciones no disponibles opacas y con explicación al pulsar

- **Estado:** Aceptado
- **Fecha:** 2026-11-01
- **Decisor:** Usuario
- **Supera a:** ADR-0054/0176 solo en apariencia e interacción de acciones desactivadas.

## Problema y evidencia

La captura del usuario muestra contenido que se transparenta bajo una acción
flotante desactivada. Button, ConfigurationOption y botones sociales aplicaban
opacity 0.55 y los controles físicamente disabled no explicaban su bloqueo.

## Alternativas y coste

- Cambiar solo opacity: coste mínimo, pero no resuelve la explicación solicitada.
- Guardar la acción y explicar el motivo con el feedback compartido: coste bajo,
  conserva componentes y evita peticiones cuando la acción no está disponible.
- Crear diálogos o tooltips locales por control: coste mayor y experiencia
  divergente entre web, iOS y Android, sin necesidad funcional.

## Recomendación y decisión explícita

Se recomienda el guard compartido y el banner existente, con motivos decididos
por cada consumidor. El usuario lo acepta mediante «Los botones desactivados no
pueden ir con alpha distinto de 1 o colores transparentes, siempre opaco y si lo
pulsas que diga por qué está desactivado».

Los controles bloqueados usan superficie, texto y borde sólidos, opacity 1,
sin degradados transparentes. Conservan objetivo táctil, foco de teclado y
semántica de no disponibilidad. Pulsar explica el motivo localizado; nunca
invoca la acción original. Las operaciones en curso explican la espera y siguen
bloqueando envíos duplicados. El motivo de negocio pertenece a la feature,
no se deduce centralmente de estados HTTP ni se obtiene de cuerpos del backend.

## Validación y retrospectiva

Comprobar variantes y temas, pulsación/teclado, motivos específicos, carga y
transición a disponible. Ejecutar typecheck, regresiones y exportación web.
Los colores disabled deben contrastar para leer la explicación y el nombre de
la acción. La opacidad del control no se usa como sustituto de un estado visual.

Validado: 63 regresiones cliente (incluidas 17 nuevas), TypeScript, lint y
exportación web estática correctos. Se comprueba HTML real de RN Web con foco
conservado, semántica heredada y ausencia de `disabled` físico. Las pruebas de
callbacks por plataforma no sustituyen QA visual ni un recorrido real con
lector de pantalla. No se han arrancado servicios local/dev ni observabilidad.
