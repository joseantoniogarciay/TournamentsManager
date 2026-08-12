# ADR-0092: Conservar dos despliegues recuperables de desarrollo

- **Estado:** Aceptado
- **Fecha:** 2026-08-12
- **Decisor:** Usuario, mediante decisión explícita
- **Propietario del análisis:** Asistente como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

El desarrollo público doméstico ejecuta una API runtime y una exportación web
estática, pero hasta ahora ambos artefactos se reemplazaban sin conservar el
SHA desplegado ni una versión local anterior. Git preserva el código, pero no
permite recuperar el servicio anterior sin reconstruirlo. Tampoco corresponde
crear una GitHub Release por cada integración diaria de `develop`.

## Alternativas

### Alternativa A — Reemplazar siempre el artefacto actual

- **Ventajas:** no requiere directorios, manifiestos ni limpieza.
- **Inconvenientes:** rollback manual y lento; no queda evidencia local de la
  combinación API/web ejecutada.
- **Coste de mantenimiento:** mínimo, con recuperación débil.

### Alternativa B — Conservar dos artefactos locales por SHA

- **Ventajas:** rollback inmediato del par API/web; evidencia legible sin
  secretos; coste y automatización mínimos.
- **Inconvenientes:** solo cubre el último despliegue anterior y depende del
  disco del Mac; no protege PostgreSQL.
- **Coste de mantenimiento:** bajo: dos directorios web, dos imágenes Docker y
  un manifiesto por despliegue.

### Alternativa C — Registry, CI/CD y releases por cada push de dev

- **Ventajas:** artefactos remotos e historial centralizado.
- **Inconvenientes:** credenciales, coste, retención, automatización y una
  GitHub Release ruidosa para cada integración diaria.
- **Coste de mantenimiento:** medio o alto, desproporcionado para el runtime
  doméstico actual.

## Decisión del usuario

**Aceptada:** alternativa B. El despliegue manual de dev parte exclusivamente
de un árbol limpio cuyo `HEAD` coincide con `origin/develop`. Construye la API
runtime con tag `git-<SHA-completo>`, exporta la web bajo un directorio por SHA y
guarda un `deployment.json` sin secretos con commit, imagen y fecha.

Caddy sirve un enlace simbólico `current` hacia la versión web activa. Se
conservan el despliegue actual y el anterior; la limpieza elimina tanto el
directorio web como la imagen asociada más antiguos. Un rollback explícito vuelve
a seleccionar uno de los SHA conservados para API y web, sin tocar PostgreSQL.

Git y GitHub siguen siendo la fuente de historial de código, configuración y
decisiones; los artefactos, imágenes y manifiestos de dev viven fuera de Git en
el Mac. No se crean GitHub Releases por los pushes ordinarios de `develop`.
Tags inmutables y GitHub Releases se reservan para producción o hitos que se
distribuyan fuera del equipo.

## Consecuencias

- Un rollback de código no es una restauración de datos ni revierte el esquema.
- El entorno no tiene aún backups de PostgreSQL: sus datos siguen siendo
  descartables. Cuando sean valiosos se decidirán dump, retención y restauración
  probada antes de afirmar que existe protección.
- La entrega permanece manual; CI y `make verify` se completan antes del
  despliegue, sin introducir credenciales de despliegue en GitHub.

## Validación

1. El comando de despliegue rechaza ramas, árboles o SHAs no publicados en
   `origin/develop`.
2. Tras desplegar, API y web señalan al mismo SHA y superan sus health checks.
3. Solo permanecen dos directorios de release y sus dos imágenes asociadas.
4. El rollback devuelve API y web a uno de esos SHA sin modificar PostgreSQL.

## Disparadores de revisión

- datos de dev que haya que conservar;
- más de una persona desplegando;
- necesidad de retener más de una versión anterior;
- despliegue automático, disponibilidad mayor o traslado estable fuera del Mac.
