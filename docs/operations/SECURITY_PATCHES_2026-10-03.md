# Parches de seguridad de producción — 2026-10-03

El usuario autorizó completar la auditoría y aplicar las correcciones necesarias.
El cambio de aplicación parte de v1.8.0 y contiene únicamente dependencias Go;
no incorpora funcionalidad, migraciones ni cambios del cliente pendientes.

## Correcciones verificadas

- Backend: OpenTelemetry 1.45.0, x/crypto 0.56.0, x/net 0.58.0 y gRPC 1.83.2,
  junto con su selección transitiva. Todas las versiones cambiadas se publicaron
  antes del 26 de septiembre, según el proxy oficial de Go.
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
  AMD SEV no corresponde a la arquitectura ARM64 de esta VM. Un bloqueo
  persistente de módulos requiere autorización específica antes de aplicarse.
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
