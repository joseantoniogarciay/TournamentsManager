# Estrategia de pruebas

> Estado: principios iniciales; herramientas y umbrales pendientes.

## Objetivo

Las pruebas reducen riesgo y aportan feedback. No se persigue una cifra de
cobertura aislada ni una pirámide rígida.

## Selección por riesgo

| Riesgo | Evidencia preferida |
|---|---|
| Invariante de dominio | Prueba rápida del dominio/caso de uso |
| Consulta o transacción | Integración con PostgreSQL real |
| Contrato externo | Prueba de contrato |
| Cableado de componentes | Integración o smoke test |
| Flujo crítico | Prueba end-to-end mínima |
| Paridad del cliente | Mismo escenario funcional en web, iOS y Android |
| Layout adaptativo | Pruebas visuales y de interacción en móvil, tablet y escritorio |
| Calidad web pública | Accesibilidad, navegación por teclado, URLs, metadatos y rendimiento |
| Seguridad/abuso | Pruebas negativas y de autorización |
| Operabilidad | Inyección de fallo y verificación de señales/runbook |

## Reglas

- Probar comportamiento observable, no detalles accidentales.
- Usar dobles solo en límites que aportan valor.
- Preferir dependencias reales en integración cuando el comportamiento específico
  importa.
- Toda corrección de bug incluye una reproducción automatizada cuando sea viable.
- Las pruebas deben ser deterministas, aisladas y diagnosticables.
- Los datos de prueba no contienen información real sensible.
- La paridad funcional no exige snapshots idénticos: cada plataforma puede
  adaptar composición y navegación conservando el mismo resultado observable.
- La matriz de navegadores, sistemas operativos, dispositivos y anchos se
  definirá por riesgo, no intentando cubrir todas las combinaciones.

## Decisiones pendientes

- librerías de assertions y mocks, si hacen falta;
- estrategia de base de datos por test;
- organización y tags;
- ejecución local y CI;
- presupuestos de tiempo;
- cobertura como señal secundaria;
- pruebas de carga, resiliencia y seguridad.
- herramientas para pruebas del cliente universal y dispositivos objetivo;
- matriz mínima de web, iOS, Android, móvil, tablet y escritorio;
- reparto de pruebas compartidas y específicas de plataforma.
