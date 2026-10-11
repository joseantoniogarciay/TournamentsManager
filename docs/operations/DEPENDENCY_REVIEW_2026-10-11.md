# Revisión de mantenimiento fuera del cliente — 2026-10-11

## Resultado y alcance

Consulta de manifiestos y releases oficiales, sin instalar infraestructura,
conectar al host productivo ni arrancar servicios. Las versiones de partida
son las del repositorio o la auditoría del 3 de octubre; no certifican el runtime
actual. ADR-0138 exige siete días para cada versión y su grafo.

| Área             | Base declarada        | Candidato       | Evidencia y prioridad                                                                                                                           |
| ---------------- | --------------------- | --------------- | ----------------------------------------------------------------------------------------------------------------------------------------------- |
| pgBackRest       | 2.59.1                | 2.59.3          | Publicado 4 octubre 09:02:09 UTC; madura 11 octubre 11:02:09 Madrid. Prioridad alta de revisión por cifrado y checksums.                        |
| PostgreSQL       | 18.4                  | 18.6            | Release oficial con correcciones de seguridad; permanece en major 18. Revisar runtime, imagen y restauración.                                   |
| Caddy            | 2.11.4 en auditoría   | 2.11.7          | Publicado 3 octubre 05:58:33 UTC; ya maduro. Corrige regresiones HTTP/2 y streams de 2.11.6; evaluar configuración y dependencias del binario.  |
| K3s              | 1.36.4 en auditoría   | 1.36.5+k3s1     | Publicado 30 septiembre 16:10:03 UTC. Preferir parche de la línea existente frente al salto a 1.37.                                             |
| OpenTelemetry Go | 1.45.0                | 1.47.0          | Publicado 2 octubre; ya maduro. Límites de profundidad de atributos y tamaño de respuesta OTLP. No adoptar Logs SDK solo por su estabilización. |
| pgx              | 5.10.0                | 5.11.0          | Publicado 7 septiembre; candidato maduro, beneficio pendiente de revisión de cambios.                                                           |
| gRPC             | 1.83.2 transitivo     | 1.84.0          | Publicado 17 septiembre; candidato maduro, sin justificación para forzarlo aislado del consumidor.                                              |
| cloudflared      | 2026.9.3 en auditoría | 2026.10.0       | Publicado 5 octubre 17:43:45 UTC; esperar al 12 octubre 19:43:45 Madrid.                                                                        |
| Traefik          | 3.7.13                | 3.7.14          | Publicado 6 octubre 07:49:56 UTC; esperar al 13 octubre 09:49:56 Madrid.                                                                        |
| x/crypto y x/net | 0.57.0 / 0.60.0       | 0.58.0 / 0.61.0 | Publicados 9 octubre; todavía jóvenes. Mantener política vigente.                                                                               |

## Recomendación y coste

Primero revisar pgBackRest y PostgreSQL, y después Caddy. Tienen razones de
mantenimiento concretas; no incorporar otra herramienta ni actualizar todo
el grafo. Comparar el binario desplegado con estos candidatos antes de cambiar
versiones. La base de datos y backups tienen coste de prueba superior:
restauración aislada, continuidad de WAL, lectura de backups anteriores y rollback
documentado. K3s y backend pueden abordarse en bloques separados.

Promtail sigue declarado en Compose dev, pero alcanzó EOL el 2 marzo 2026.
Producción ya tiene configuración Alloy. Revisar migración de dev al próximo
trabajo explícito de observabilidad, conservándolo apagado mientras tanto.

## pgBackRest: condición y límites

Versiones anteriores a 2.59.3 no comprobaban fallos al generar aleatoriedad para
cifrado. El mantenedor cita errores raros de OpenSSL o políticas que bloquean
getrandom. Solo afecta a repositorios cifrados. El proyecto tiene configuración
de cifrado, pero esta revisión no demuestra que se generase una clave débil.
Actualizar evita futuras generaciones silenciosas; no reemplaza las subclaves
persistidas de archive.info y backup.info. Si una comprobación segura confirma
debilidad, la remediación oficial requiere un repositorio nuevo, conservando el
anterior mientras sea necesario. No leer ni imprimir claves en logs del asistente,
no borrar backups ni alterar retención durante la revisión.

## Retrospectiva

Separar publicación, maduración, exposición y versión desplegada evita confundir
una novedad con una corrección efectiva. Un parche de cifrado requiere revisar
estado persistido, no solo el ejecutable. No se aplica infraestructura por esta
consulta ni se activa observabilidad.

## Fuentes primarias

- [pgBackRest: aviso y remediación](https://pgbackrest.org/news.html#weak-encryption).
- [pgBackRest 2.59.3](https://github.com/pgbackrest/pgbackrest/releases/tag/release/2.59.3).
- [PostgreSQL 18.6](https://www.postgresql.org/docs/release/18.6/).
- [Caddy 2.11.7](https://github.com/caddyserver/caddy/releases/tag/v2.11.7).
- [K3s 1.36.5](https://github.com/k3s-io/k3s/releases/tag/v1.36.5%2Bk3s1).
- [OpenTelemetry 1.47.0](https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.47.0).
- [Go proxy: fechas por módulo](https://proxy.golang.org/).
- [cloudflared 2026.10.0](https://github.com/cloudflare/cloudflared/releases/tag/2026.10.0).
- [Traefik 3.7.14](https://github.com/traefik/traefik/releases/tag/v3.7.14).
- [Grafana: fin de soporte de Promtail](https://grafana.com/docs/alloy/latest/set-up/migrate/from-promtail/).

## Aplicación autorizada

ADR-0157 acepta los tres parches y la excepción puntual de pgBackRest.
PostgreSQL 18.6 y pgBackRest 2.59.3 aplicados con restauración previa/posterior
y copia completa sin expiración. Caddy 2.11.7 activo y API pública 200;
ver Deployment para evidencia y diagnóstico de la transición.
