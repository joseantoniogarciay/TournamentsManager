# Auditoría de dependencias de la aplicación productiva — 2026-10-03

Este documento conserva la primera fase de auditoría. La revisión ampliada y
las correcciones posteriores están en [Parches de seguridad](SECURITY_PATCHES_2026-10-03.md).

## Resultado y alcance

No se confirmó una vulnerabilidad crítica en las dependencias de aplicación
revisadas. Este resultado no certifica toda la infraestructura ni descarta
vulnerabilidades desconocidas.

La web activa y ambas réplicas de API ejecutan v1.8.0, revisión
`17eab6fd1bd71dc90afaba13e7340de12a8bf772`. La identidad se comprobó en el
manifiesto del release web y en los Pods reales, sin usar como prueba el tag
desactualizado del manifiesto YAML del checkout.

Se exportó en modo lectura la imagen activa de containerd y se extrajeron sus
tres ejecutables. Su configuración tiene el digest
`sha256:1bee6078a34a6410cd5f9cdb75bd68f6fc34119b77c08a04df9dad30ecf8b88f`,
coincidente con el `imageID` de ambas réplicas. Los binarios productivos usan
Go 1.26.6; el renderer del Mac también incorpora esa versión y la revisión
activa. No se cambiaron dependencias, servicios ni despliegues.

## Método y evidencia

- `govulncheck` v1.6.0, ya aceptado por ADR-0012, contra `vuln.go.dev`; la base
  consultada indicaba actualización el 2026-10-01 a las 20:24:15 UTC.
- Análisis binario de `/api`, `/legal-audit-backup`, `/legal-audit-restore`
  extraídos de la imagen y del renderer productivo.
- Análisis de fuentes de los tres comandos de la imagen, desde una copia
  temporal obtenida con `git archive` del SHA activo, con `GOOS=linux` y
  `GOARCH=arm64`. Se confirmó Go 1.26.6 en el resultado. Lanzar el auditor
  desde el módulo evita atribuir al artefacto el Go 1.26.5 global del Mac.
- `pnpm audit --prod --json` sobre el lockfile del mismo SHA, sin instalar
  paquetes ni construir el cliente. El resultado no equivale a un análisis
  de alcance de funciones del bundle estático.
- `make vuln` sobre el checkout actual, conservando sus cambios previos.

## Go: hallazgos y condiciones

| Área | Evidencia | Interpretación |
| --- | --- | --- |
| Biblioteca estándar Go 1.26.6 | Sin avisos en los binarios revisados | No hay corrección del compilador identificada por este análisis |
| OpenTelemetry, GO-2026-6505 / CVE-2026-81870 | Alcanzable en fuentes; exporters 1.43.0 y SDK 1.44.0; corregido en 1.45.0 | Gravedad baja, 2,0/10 según el mantenedor; requiere logs internos Info explícitos y acceso a ellos |
| gRPC, GO-2026-6443 | 1.83.1, corregido en 1.83.2; paquete importado, sin llamadas vulnerables identificadas en fuentes | Aviso alto para servidores xDS; no se encontró creación de ese servidor en la aplicación |
| SSH, GO-2026-6354, GO-2026-6355 y GO-2026-6303 | x/crypto 0.54.0; correcciones en 0.56.0 y 0.55.0; solo nivel módulo en fuentes | No se encontró uso de SSH en el código de aplicación |
| OpenPGP, GO-2026-5932 | Aviso de paquete sin mantenimiento dentro de x/crypto, sin versión corregida | No se encontró uso; actualizar x/crypto no elimina por sí solo este aviso a nivel módulo |
| Backup, restore y renderer | Análisis binario sin vulnerabilidades | Resultado limitado a estos ejecutables y a la base consultada |

La aplicación configura `slog`, pero no se encontró `otel.SetLogger` ni un
logger de diagnóstico interno detallado. Por ello, el aviso de OpenTelemetry
es alcanzable según el scanner, pero no se confirmó su condición de exposición.
No se inspeccionaron logs de usuarios ni se buscaron secretos en ellos.

El scanner de la API binaria informó cinco avisos a nivel símbolo y otro a
nivel módulo. El análisis de fuentes dejó un aviso alcanzable, otro a nivel
paquete y cuatro a nivel módulo. Los modos no son equivalentes: la presencia
conservadora en el binario no demuestra una ruta de ejecución explotable.

El checkout local ya contiene OpenTelemetry 1.45.0, gRPC 1.83.2 y x/crypto
0.56.0. `make vuln` terminó correctamente: cero avisos alcanzables, cero en
paquetes importados y uno en módulos requeridos. Esos cambios previos aún
no son la versión desplegada y esta auditoría no los promociona.

## Cliente

El resumen de npm indica cero críticos, 29 altos y siete moderados. Hay 27
identificadores de aviso distintos; los recuentos del resumen no son 36
vulnerabilidades explotables confirmadas en la web.

Los paquetes señalados son `uuid`, `brace-expansion`, `braces`, `js-yaml`,
`nanoid`, `decode-uri-component`, `@xmldom/xmldom`, `image-size` y `node-forge`.
Varias rutas pasan por Expo CLI, config plugins o Metro: un paquete situado
en dependencias de producción del workspace puede pertenecer al proceso de
build y no al JavaScript servido. Queda por determinar su presencia y uso en
el bundle antes de afirmar exposición de la web.

## Recomendación y límites

Recomendación: incorporar las correcciones ya preparadas en una promoción
revisada, verificando fechas de publicación y la espera de siete días de
ADR-0138. El aviso bajo de OpenTelemetry no justifica una excepción crítica.
No actualizar todo el grafo ni desplegar el árbol de trabajo con cambios ajenos
como consecuencia automática de una auditoría.

Se obtuvo el inventario real del clúster, pero no se auditaron las dependencias
internas de sus otras imágenes, los paquetes Ubuntu, K3s, PostgreSQL/pgBackRest,
Traefik, Caddy, cloudflared ni la pila de observabilidad. Tampoco se examinó
el migrador Goose. Estos límites impiden afirmar que toda producción esté
libre de vulnerabilidades críticas.

## Retrospectiva y fuentes

El aprendizaje principal es auditar primero el artefacto desplegado y después
su código exacto: ni un checkout más reciente ni un scanner binario aislado
resuelven por sí solos versión, alcance y condiciones de explotación.

- [Gestión de vulnerabilidades de Go](https://go.dev/doc/security/vuln/):
  la base Go no asigna categorías de gravedad; ausencia de etiqueta no significa
  ausencia de criticidad.
- [Aviso OpenTelemetry y condiciones de exposición](https://github.com/open-telemetry/opentelemetry-go/security/advisories/GHSA-8wmf-6v46-5gfg).
- [Aviso gRPC, gravedad alta y servidor xDS](https://github.com/grpc/grpc-go/security/advisories/GHSA-2v4p-qf9q-27wj).
- [Aviso OpenPGP](https://pkg.go.dev/vuln/GO-2026-5932).
- [Política vigente de maduración](../adr/0138-wait-seven-days-except-critical-vulnerabilities.md).
