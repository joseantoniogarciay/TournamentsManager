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

## Decisiones pendientes

- librerías de assertions y mocks, si hacen falta;
- estrategia de base de datos por test;
- organización y tags;
- ejecución local y CI;
- presupuestos de tiempo;
- cobertura como señal secundaria;
- pruebas de carga, resiliencia y seguridad.
