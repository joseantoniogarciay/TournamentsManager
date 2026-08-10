# Estrategia de pruebas

> Estado: estrategia por riesgo y capas aceptada. CI ejecuta `make verify` como
> señal informativa; las herramientas específicas, presupuestos y gates por
> riesgo continúan pendientes.

## Objetivo

Las pruebas reducen riesgo y aportan feedback. [ADR-0019](../adr/0019-use-risk-based-layered-testing.md)
acepta seleccionar la evidencia por riesgo, no perseguir una cifra de cobertura
aislada ni una pirámide rígida.

## Selección por riesgo

| Riesgo | Evidencia preferida |
|---|---|
| Invariante de dominio | Prueba rápida del dominio/caso de uso |
| Consulta o transacción | Integración con PostgreSQL real |
| Deriva SQL/esquema/Go | Generación `sqlc` y compilación |
| Esquema inicial | Aplicación desde vacío y generación `sqlc` |
| Contrato externo | Prueba de contrato |
| Deriva OpenAPI/Go | Validación del backend contra la descripción |
| Deriva OpenAPI/TypeScript | Regeneración determinista y diff limpio |
| Cableado de componentes | Integración o smoke test |
| Flujo crítico | Prueba end-to-end mínima |
| Paridad del cliente | Mismo escenario funcional en web, iOS y Android |
| Layout adaptativo | Pruebas visuales y de interacción en móvil, tablet y escritorio |
| Calidad web privada inicial | Accesibilidad, navegación por teclado, URLs internas y rendimiento percibido |
| Calidad web pública futura | Metadatos, previews sociales, SEO y estrategia de rendering si se acepta visibilidad pública |
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
- El esquema inicial se aplica explícitamente, sin depender del arranque de la API.
- Toda corrección de bug incluye una reproducción automatizada cuando sea viable.
- Las pruebas deben ser deterministas, aisladas y diagnosticables.
- Los datos de prueba no contienen información real sensible.
- El tipado TypeScript mejora el feedback de desarrollo, pero no reemplaza la
  validación en runtime ni las pruebas del backend.
- La paridad funcional no exige snapshots idénticos: cada plataforma puede
  adaptar composición y navegación conservando el mismo resultado observable.
- La matriz de navegadores, sistemas operativos, dispositivos y anchos se
  definirá por riesgo, no intentando cubrir todas las combinaciones.

## Evidencia de integración actual

La suite PostgreSQL ejecutada en CI cubre las transacciones de mayor riesgo del
vertical slice actual: creación, inicio y cancelación de una liga; registro y
corrección de resultados con historial; clasificación calculada desde los
marcadores persistidos; autorización de las mutaciones; y cierre atómico con
rechazo de partidos pendientes, co-campeones conservados y solicitudes de
cierre concurrentes. También cubre
restablecimiento de contraseña (consumo único, revocación de sesiones y nueva
sesión) y cambio de contraseña o vinculación de Google desde opciones de cuenta
con ticket de reautenticación de un solo uso.
Las validaciones, formatos y respuestas HTTP que no dependen de semántica real
de PostgreSQL permanecen en pruebas unitarias o de handler.

El adaptador Google valida su firma y audiencia con JWT y JWKS generados en la
propia prueba, sin pedir tokens, credenciales ni certificados a Google. Las
invariantes de identidad que afectan al dominio se prueban en el servicio
federado antes de acceder a persistencia.

## Capas aceptadas

- **Dominio y casos de uso:** pruebas unitarias rápidas con la biblioteca estándar
  `testing`. Los dobles se usan solo en puertos externos reales, no para imitar
  `pgx` o tablas.
- **Persistencia:** integración con PostgreSQL real para SQL, restricciones,
  transacciones, esquema inicial y generación `sqlc`. Cada ejecución usa una base
  de pruebas efímera creada desde vacío, nunca la base local de
  desarrollo.
- **HTTP y contratos:** handlers con `httptest`; validación y generación
  determinista para impedir deriva de OpenAPI respecto al backend y cliente.
- **Flujos completos:** end-to-end mínimos, reservados a recorridos críticos una
  vez existan API y cliente funcionales.

Cobertura es una señal secundaria. Toda corrección de bug incluirá una prueba de
regresión cuando sea viable.

## Baseline ejecutable de Go

ADR-0012 establece los comandos generales. ADR-0019 adopta `testing` y
`httptest` como base inicial, sin añadir librerías de assertions, mocks o
contenedores hasta que una necesidad repetida lo justifique:

- `make test`: todos los paquetes del módulo;
- `make test-integration`: aplica el esquema inicial y prueba PostgreSQL real cuando recibe una URL
  de base aislada; en CI la proporciona el servicio efímero del workflow;
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

- librerías de assertions y mocks, si se demuestran necesarias;
- limpieza adicional de la base efímera de pruebas si el aislamiento por runner deja de ser suficiente;
- organización, tags y presupuesto temporal;
- ejecución local y CI: PostgreSQL 18.4 efímero en GitHub Actions para las
  transacciones que declaran `TM_INTEGRATION_DATABASE_URL` (ADR-0071);
- pruebas de carga, resiliencia y seguridad.
- herramientas para pruebas del cliente universal y dispositivos objetivo;
- matriz mínima de web, iOS, Android, móvil, tablet y escritorio;
- reparto de pruebas compartidas y específicas de plataforma.
