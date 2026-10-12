# Cierre de checks — API security 2026-10-12

Implementación autorizada mediante ADR-0163 sobre develop e724c0e más cambios
del checkout. No se realiza commit, integración de rama ni despliegue en esta
sesión. Los logs iniciales de la auditoría se conservan por separado.

| Comprobación | Resultado | Evidencia |
| --- | --- | --- |
| PostgreSQL 18.6 efímero, esquema y 22 migraciones, -race -run Integration -count=1 | 55 tests y subcasos pasan | fixed-postgres-race.log |
| Suite Go completa -race ./... | Pasa; PostgreSQL omitido aquí y acreditado arriba | fixed-go-race.log |
| Lint Go / build | 0 issues / pasa | fixed-go-lint.log, fixed-go-build.log |
| Lint OpenAPI | Pasa; una regla ya ignorada por configuración | fixed-openapi-lint.log |
| Generación OpenAPI dos veces | Hashes SHA-256 de todos los archivos idénticos | fixed-generation-repeat.log; comparación ejecutada por script de cierre |
| sqlc-generate-check | Sin deriva | fixed-sqlc.log |
| Typecheck / lint cliente | Pasan | fixed-typecheck.log, fixed-client-lint.log |
| Features y sesión cliente | 70 tests pasan | fixed-client-tests.log |
| Exportación web | Pasa, artefacto temporal /tmp/tm-api-security-web-export | fixed-web-export.log |
| Govulncheck | 0 alcanzables; aviso de módulo openpgp no importado | fixed-vuln.log |
| Formato Go y Prettier del repositorio | Pasan | Checks ejecutados al cierre; sin cambio funcional tras formato |
| Seguridad operacional | 11 tests pasan | tests/operational-safety.test.py; salida observada en sesión |
| Cierre operacional | Contenedor tm-api-security-review-20261012 eliminado | Script con finally/docker stop y verificación posterior de ausencia |

Volumen local medido una vez bajo detector de carreras: familia de 5001 ediciones,
51 filas consultadas, 11,690 ms; liga de 64 equipos con 2016 encuentros,
51 filas personales, 11,554 ms. Las 51 incluyen la fila adicional de paginación.
No se extrapola throughput, percentiles o latencia de producción.

Las pruebas temporales no usan bases de despliegue, observabilidad, local/dev,
datos reales o Apple real. El rol efímero posee el esquema y no acredita los
privilegios del rol PostgreSQL de producción. No se hace QA visual nativa.

Retrospectiva: la revocación necesita un snapshot posterior al bloqueo. El hash
necesita cuota y admisión antes de trabajo, y los generadores que limpian carpetas
se ejecutan secuencialmente antes de las pruebas de tipos. El detector de carreras
se ejecuta con PostgreSQL real para no ocultar el recorrido detrás de un skip.
