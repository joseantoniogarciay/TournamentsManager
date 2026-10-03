# ADR-0136: Soportar tenis de mesa con juegos por puntos

- **Estado:** Aceptado
- **Fecha:** 2026-10-02
- **Decisor:** Usuario
- **Supera a:** ADR-0135, exclusivamente para ampliar deportes y formatos de sets

## Problema y evidencia

Tenis de mesa necesita conservar juegos a 11 puntos con diferencia de dos. La
estructura de sets existente sirve, pero su validación de juegos de tenis no.
El usuario ha elegido este deporte después de corregir la validación de v1.8.0.

## Alternativas y coste

- Guardar solo juegos ganados: coste bajo, pierde el tanteo y su validación.
- Perfil explícito sobre sets ordenados: coste medio de adopción y bajo de
  mantenimiento; reutiliza contrato e historial con una regla deportiva propia.
- Motor configurable: coste alto y combinaciones sin demanda demostrada.

## Recomendación y decisión del usuario

La recomendación es el perfil explícito, mínimo suficiente. El usuario aceptó
el deporte y confirmó 3, 5 y 7 juegos en eliminatoria directa.

- Perfil `table_tennis`, con `bestOfSets: 3 | 5 | 7`, por defecto 3 en cliente.
- Cada juego termina a 11 con perdedor hasta 9, o por diferencia exactamente
  de dos a partir del 10–10. No se aceptan juegos después de alcanzar la victoria.
- Nombres de personas o parejas bajo el vocabulario «participantes»; sin plantillas.
- Solo eliminatoria directa; ligas y clasificación requieren otra decisión.
- El agregado se deriva, las correcciones son atómicas y mantienen historial.
- El límite técnico de cada tanteo sigue siendo 32767 (`smallint`); no es un
  límite deportivo de puntos. El cliente y contrato lo declaran explícitamente.

## Consecuencias y validación

Se amplía a siete el número de sets almacenables y se generaliza el validador
por perfiles cerrados. Las reglas permanecen en el dominio. Se prueban 11–9,
12–10, 13–11, juegos parciales, empates, victoria anticipada, formato inválido,
persistencia, propagación de ganadora y corrección con historial.

## Fuentes

- [ITTF: reglas y formatos de juegos](https://www.ittf.com/2021/07/22/table-tennis-101/?from=6)

## Documentación afectada

Producto, API, modelo, sistema de diseño, observabilidad, decisiones y aprendizaje.
