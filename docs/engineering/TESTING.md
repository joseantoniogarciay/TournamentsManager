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
| Deriva SQL/esquema/Go | Generación `sqlc` y compilación |
| Evolución de esquema | Migración desde vacío y desde versión anterior |
| Contrato externo | Prueba de contrato |
| Deriva OpenAPI/Go | Validación del backend contra la descripción |
| Deriva OpenAPI/TypeScript | Regeneración determinista y diff limpio |
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
- Las consultas, restricciones y transacciones se validan contra PostgreSQL real;
  un mock de `pgx` no sustituye esa evidencia.
- La generación `sqlc` debe ser determinista y fallar si una consulta deja de ser
  compatible con el esquema.
- Las migraciones `goose` se prueban sin depender del arranque de la API.
- Toda corrección de bug incluye una reproducción automatizada cuando sea viable.
- Las pruebas deben ser deterministas, aisladas y diagnosticables.
- Los datos de prueba no contienen información real sensible.
- El tipado TypeScript mejora el feedback de desarrollo, pero no reemplaza la
  validación en runtime ni las pruebas del backend.
- La paridad funcional no exige snapshots idénticos: cada plataforma puede
  adaptar composición y navegación conservando el mismo resultado observable.
- La matriz de navegadores, sistemas operativos, dispositivos y anchos se
  definirá por riesgo, no intentando cubrir todas las combinaciones.

## Baseline ejecutable de Go

ADR-0012 establece los comandos generales, sin decidir todavía librerías,
fixtures, integración con PostgreSQL ni cobertura:

- `make test`: todos los paquetes del módulo;
- `make test-race`: detector de carreras para una comprobación más lenta;
- `make check`: formato, lint y tests como feedback local;
- `make verify`: añade módulos limpios, build y vulnerabilidades.

Durante el desarrollo se prefiere el test más cercano al cambio:

```bash
go -C apps/backend test ./ruta/del/paquete/...
```

No se crean tests vacíos para demostrar que el runner funciona. El primer test se
añadirá con el primer comportamiento observable o riesgo real.

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
