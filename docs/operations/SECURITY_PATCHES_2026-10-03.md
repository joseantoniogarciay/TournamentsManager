# Parches de seguridad de producción — 2026-10-03

> Registro cronológico de los parches de seguridad de v1.8.1–v1.8.3.
> La publicación posterior v1.9.0 y las revisiones activas están en
> [DEPLOYMENT.md](DEPLOYMENT.md); las referencias a versiones anteriores
> describen el momento de cada comprobación.

El usuario autorizó completar la auditoría y aplicar las correcciones necesarias.

La adaptación posterior de consumidores se publicó como v1.8.3 tras comprobar
dev y prod. El grafo JavaScript queda en dos altos y cero moderados/críticos;
su detalle y evidencia están en [Adaptación de dependencias web](WEB_DEPENDENCY_COMPATIBILITY_2026-10-03.md).
El cambio de aplicación parte de v1.8.0 y contiene únicamente dependencias Go;
no incorpora funcionalidad, migraciones ni cambios del cliente pendientes.

## Correcciones verificadas

- Backend: OpenTelemetry 1.45.0, x/crypto 0.56.0, x/net 0.58.0 y gRPC 1.83.2,
  junto con su selección transitiva. Todas las versiones cambiadas se publicaron
  antes del 26 de septiembre, según el proxy oficial de Go.
- Migrador Goose: se conserva 3.27.1 y se parchea su grafo separado de
  herramientas (crypto, net, gRPC, cel-go, compress y tooling Go). El análisis
  de fuentes del grafo anterior encontró GO-2026-6348 alcanzable en gRPC
  1.83.0, corregido desde 1.83.1. Todas las versiones cambiadas cumplen siete
  días; no se ejecutan migraciones como parte de este parche de dependencias.
- Traefik: 3.7.8 → 3.7.13, publicado el 4 de septiembre; imagen fijada por
  digest en `infra/k3s/core/traefik-config.yaml`. El HelmChartConfig conserva el
  parche frente a reconciliaciones de K3s. El nuevo análisis no encuentra el
  aviso crítico CVE-2026-88007. Su explotación requería HTTP/3 y un backend
  NTLM/Negotiate; esas condiciones no se encontraron en la configuración.
- K3s: 1.36.3+k3s1 → 1.36.4+k3s1, publicado el 27 de agosto. El binario oficial
  se verificó por SHA-256 y utiliza Go 1.26.7. Antes de cambiarlo se conservaron
  copias protegidas del binario y del datastore detenido. El nodo volvió a Ready
  y los Pods productivos siguieron disponibles.
- Ubuntu: curl y sus dos bibliotecas en 8.5.0-2ubuntu10.15, libexpat1 en
  2.6.1-2ubuntu0.6 y libpcap0.8t64 en 1.10.4-4.1ubuntu3.1. La simulación no
  añadió ni retiró paquetes. Se comprobaron fechas reales de publicación en
  Launchpad (24 y 25 de septiembre), no solo fechas de changelog.

## Auditoría y límites que siguen abiertos

Se analizaron las 17 imágenes distintas del inventario inicial del clúster,
la imagen del migrador v1.8.0, los binarios Caddy/cloudflared y los 751 paquetes
del inventario Ubuntu, además de la aplicación. Trivy 0.74.0 se utilizó como
auditor puntual desde un directorio temporal, tras verificar su checksum y
publicación del 14 de agosto; no se adoptó como servicio o gate permanente.

Las bases y los modos de análisis tienen límites: los avisos de versión no
demuestran ejecución de la función ni exposición de la condición vulnerable.
No se afirma que todos los contadores estén a cero.

- Kernel Noble 6.8: cinco CVE críticos distintos en NFSD, NVMe/TCP, SCTP,
  SRP target y AMD SEV. Ubuntu aún no publica corrección para varios en esta
  línea. Los cuatro módulos de red/almacenamiento se encontraron descargados;
  AMD SEV no corresponde a la arquitectura ARM64 de esta VM. Con autorización
  específica posterior del usuario se bloqueó la carga normal de los cuatro
  módulos; es una mitigación, no una corrección del paquete del kernel.
- PostgreSQL Bookworm: quedan avisos de Perl, SQLite y libxml2 sin versión
  corregida en esa distribución. Debian considera algunos menores o aplazados.
  El aviso Perl CVE-2026-8376 requiere una build de 32 bits, distinta de esta
  imagen ARM64. La API usa PostgreSQL y no ofrece ejecución de SQL SQLite ni
  extracción de archivos Perl; no se confirmó explotación de estos avisos.
- zlib CVE-2023-45853 corresponde a MiniZip, no a las funciones normales de
  zlib. El paquete zlib no debe declararse arreglado ni vulnerable en ejecución
  solo por compartir la versión del source package.
- gosu 1.19 contiene Go 1.24.6: Trivy señala CVE-2025-68121 de TLS a nivel
  biblioteca, pero `govulncheck` del ejecutable encuentra cero avisos a nivel
  símbolo. No se encontró la función TLS vulnerable en este helper.
- Caddy 2.11.4 tiene avisos en sus dependencias. Las nuevas versiones 2.11.6
  y 2.11.7 se publicaron el 1 y 3 de octubre; no se saltó la espera de siete
  días por avisos no críticos ni se fabricó una build propia de Caddy.
- cloudflared 2026.9.3 es la última versión oficial comprobada. Todavía contiene
  dependencias señaladas por el análisis binario; estar actualizado no prueba
  que todo el grafo esté corregido.
- Kernel 6.8.0-146.146, OpenSSL y otros parches Ubuntu tienen publicaciones
  recientes. Se conservaron los parches maduros y no se eludió la espera.
- El lockfile del cliente v1.8.0 indica cero críticos, con avisos altos y
  moderados, varios en Expo CLI/Metro. No se forzaron actualizaciones mayores
  transitivas de herramientas como si fueran parches del bundle web.

## Validación y retrospectiva

El backend parcheado pasa `go test ./...`, `go test -race ./...`, build,
`go mod tidy -diff` y `govulncheck`: cero avisos alcanzables, con el aviso
OpenPGP del módulo x/crypto sin uso por la aplicación. Tras los cambios de
infraestructura, web y `/healthz` de API devolvieron HTTP 200 y PostgreSQL
conservó una réplica lista.

La lección es fechar el paquete realmente publicado, auditar la imagen activa
y distinguir un parche disponible de una mitigación o un aviso sin corrección.
Cambiar de distribución o línea de kernel para vaciar un scanner sería una
decisión adicional, con pruebas y mantenimiento propios.

## Fuentes

- [Aviso crítico Traefik](https://github.com/traefik/traefik/security/advisories/GHSA-qqjf-53cj-pwvv).
- [K3s 1.36.4](https://github.com/k3s-io/k3s/releases/tag/v1.36.4%2Bk3s1).
- [OpenTelemetry: gravedad baja y corrección](https://github.com/open-telemetry/opentelemetry-go/security/advisories/GHSA-8wmf-6v46-5gfg).
- [Ubuntu: NFSD y estado de corrección](https://ubuntu.com/security/CVE-2026-53398).
- [Debian: Perl en 32 bits](https://security-tracker.debian.org/tracker/CVE-2026-8376).
- [Debian: SQLite](https://security-tracker.debian.org/tracker/CVE-2025-7458).
- [Debian: zlib y MiniZip](https://security-tracker.debian.org/tracker/CVE-2023-45853).

## Promoción de aplicación y correcciones JavaScript

La API y el renderer productivos ejecutan v1.8.1, commit
`f6a199ad42bd7e535db9905b132cd75d69486396`. Las dos réplicas están listas,
con imageID `sha256:341a9cca2cb09b72992c2a7cfe7f6d53f4c976b226eadbbb9f1b209a57e7d12d`;
web y API devuelven 200 y PostgreSQL conserva 1/1 listo. Goose queda importado
para el próximo uso; no se ejecutaron migraciones. CI `make verify` pasó para
`eaa9705`. Las imágenes finales de API y migrador presentan solo el aviso
OpenPGP sin gravedad, sin uso identificado por el análisis de fuentes.

El lockfile JavaScript se corrige conservando las dependencias directas de Expo
y React Native. Los overrides seleccionan ramas compatibles de xmldom,
brace-expansion, js-yaml y nanoid, y consumidores concretos como uuid de xcode
(que usa v4) y Fastify de Scalar (línea 5). También corrigen avisos de herramientas
del workspace. El diff se limita a configuración de suministro y resoluciones;
no cambia código de interfaz ni contratos. Las nuevas resoluciones se verifican
contra las fechas del registro npm; no hay exclusiones por edad. ADR-0138
registra la decisión de maduración ya aceptada por el usuario.

`pnpm audit --prod` del lockfile corregido pasa de 29 altos y 7 moderados a
4 altos y 1 moderado, con cero críticos. Es un recuento de avisos, no de
funciones explotables. La exportación web se debe reconstruir con este lockfile
para llevar sus correcciones a producción.

Persisten node-forge 1.4.0 y braces 3.0.3, cuyas correcciones indicadas por el
auditor no están publicadas; decode-uri-component 0.2.2, cuya alternativa
publicada cambia CommonJS por ESM y requiere actualizar query-string; e
image-size 1.2.1, cuya versión 2 cambia la lectura de rutas que usa Metro.
Forzar esos overrides rompería consumidores. Se mantienen pendientes hasta una
corrección publicada y madura o una actualización compatible de la matriz Expo.
La nueva matriz Expo sugerida todavía no cumple siete días; la web conserva
el SDK y las dependencias nativas existentes. No se reinstala ni arranca el
entorno local retirado: CI y el checkout aislado validan el artefacto de producción.

El manifiesto declarativo de API conserva ahora la imagen de v1.8.1 ya activa;
se corrige su referencia anterior desactualizada para que una aplicación futura
del YAML no revierta los parches. La publicación web no cambia esa imagen.

## Cierre operativo de v1.8.2

La web está publicada como v1.8.2, commit
`e86afeb18cdadf45b41e827a4b4bb30385d40ec3`, construida el 3 de octubre
a las 11:50:16 UTC. CI `make verify` aprobó el último commit de PR #2
`ffdc99f`; typecheck y exportación web también pasaron en el checkout aislado.
Se comprobaron las fechas de las 19 resoluciones JavaScript nuevas en npm.
La web preparada conservó las entradas públicas previas, canonical y sitemap,
y el bundle apunta a la API de producción. La activación mantuvo el release
v1.8.0 como rollback web. Inicio y cuenta cargaron en navegador sin errores de
consola; web, API y el renderer del torneo controlado respondieron HTTP 200.

La API y su renderer permanecen en v1.8.1 (SHA f6a199a); no se ejecutaron
migraciones. K3s está Ready, los Pods en ejecución están listos y PostgreSQL
conserva 1/1. El manifiesto API del checkout conserva la imagen parcheada.

La revisión automática rechazó inicialmente instalar el bloqueo persistente de
nfsd, nvmet_tcp, sctp e ib_srpt por falta de autorización específica y posible
impacto funcional. Después de explicar sus funciones y comprobar de nuevo su
ausencia de uso, el usuario autorizó: «Pues bloqueemos sin problema». Con esa
autorización se instaló el bloqueo y se verificó, como se detalla a continuación.

Los cambios se integraron mediante PR #1 y #2 desde checkouts aislados. En aquel momento, el
checkout develop conservaba cambios funcionales y debía incorporar el historial
de los hotfixes. La conciliación del 2026-10-03 actualiza ahora su base a
`6967457`, conserva el trabajo operativo y mantiene las pruebas ya integradas;
no quedan cambios funcionales locales respecto de esa revisión.

## Mitigación de módulos del kernel aplicada

El 3 de octubre, con autorización explícita, se instaló
`infra/k3s/host/fasttourney-unused-kernel-modules.conf` en
`/etc/modprobe.d/fasttourney-unused-kernel-modules.conf`, propiedad root y modo
0644. Cada módulo tiene `blacklist` y una regla `install … /bin/false`: la primera
evita la selección automática por alias y la segunda rechaza su carga normal
mediante modprobe. No se descargaron módulos ni se reinició el host.

Antes de instalar se confirmó que ninguno de los cuatro módulos estaba cargado.
La revisión adicional no encontró servicios asociados, montajes de red ni sockets
SCTP; los volúmenes Kubernetes usan local-path y sus servicios usan TCP/UDP.
Después de instalar, `modprobe --dry-run --verbose` mostró `install /bin/false`
para los cuatro módulos. K3s siguió Ready, API 2/2, PostgreSQL 1/1 y web/API en
HTTP 200.

La mitigación es reversible retirando únicamente este archivo de modprobe.d.
Debe revisarse antes de introducir un servidor NFS, un target NVMe/TCP,
SCTP o un target SRP, y al disponer de un kernel corregido. No impide que un
administrador privilegiado omita modprobe ni elimina los avisos de versión del
scanner. La retrospectiva es reducir superficie no utilizada con evidencia y
autorización, conservando pendiente la actualización que corrige el kernel.
