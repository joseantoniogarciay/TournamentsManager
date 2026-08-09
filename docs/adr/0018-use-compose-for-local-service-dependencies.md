# ADR-0018: Usar Docker Compose para dependencias locales de servicio

- **Estado:** Superado por ADR-0076
- **Fecha:** 2026-07-25
- **Decisor:** Usuario, mediante aceptación explícita de la alternativa B
- **Propietario del análisis:** Asistente como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** ADR-0076

## Problema

El equipo necesita un entorno local reproducible para desarrollar y validar la
persistencia PostgreSQL real. Debe reducir la diferencia operativa relevante con
producción sin convertir las herramientas de desarrollo ni los clientes en
contenedores prematuros.

## Contexto y restricciones

- El manifiesto fija Docker y Docker Compose como dirección y exige que local se
  parezca a producción en sus contratos importantes, no en toda su topología.
- ADR-0011 acepta PostgreSQL, `pgx`, `sqlc` y `goose`; las migraciones SQL se
  ejecutan separadas del arranque de la API.
- ADR-0012 y ADR-0014 fijan toolchains reproducibles de Go, Node y pnpm que se
  ejecutan desde el host durante el desarrollo.
- ADR-0015 acepta Expo para web, iOS y Android; sus herramientas y simuladores
  son nativos del host.
- ADR-0017 establece contratos `.env` locales ignorados y ejemplos versionados.
- Redis/Valkey, MinIO/S3, observabilidad y Kubernetes siguen pendientes; no se
  incorporan en este entorno por anticipación.
- Todavía no existe API, cliente ni imagen de aplicación. Esta decisión no fija
  la versión de PostgreSQL, la estructura de archivos ni los comandos concretos;
  esos detalles se fijarán al implementar un entorno reproducible.

## Criterios de decisión

1. reproducir PostgreSQL y su ciclo de vida sin instalación por máquina;
2. conservar un bucle rápido de edición, depuración y pruebas para Go y Expo;
3. mantener visibles configuración, salud, persistencia y destrucción de datos;
4. evitar servicios sin necesidad actual;
5. facilitar la futura construcción de imágenes y despliegue sin hacerlos
   requisito para desarrollar;
6. mantener una ruta de diagnóstico y recuperación sencilla.

## Alternativas

### Alternativa A — Contenerizar aplicaciones y dependencias en Compose

Ejecutar PostgreSQL, API Go y el cliente web mediante Docker Compose; tratar los
targets móviles como una excepción en host.

- **Ventajas:** la API comparte desde el inicio un modelo de imagen y red con un
  futuro despliegue; `depends_on` puede coordinar servicios contenidos.
- **Inconvenientes:** bind mounts, recarga, permisos, red y builds añaden fricción
  al bucle de desarrollo antes de tener una imagen de backend o un requisito de
  despliegue; no resuelve los simuladores iOS/Android.
- **Coste de adopción:** medio.
- **Coste de mantenimiento:** medio; exige mantener Dockerfiles y configuración
  de desarrollo además de los toolchains locales.
- **Riesgos:** confundir un contenedor de desarrollo con el diseño de despliegue
  y retrasar el feedback cotidiano.

### Alternativa B — Compose para dependencias; aplicaciones en el host

Ejecutar PostgreSQL como servicio de Docker Compose y ejecutar API Go y cliente
Expo desde el host con sus comandos versionados.

- **Ventajas:** PostgreSQL tiene imagen, volumen, salud y ciclo de vida
  reproducibles; Go y Expo conservan recarga, depuración y simuladores nativos;
  cada pieza usa la herramienta que mejor corresponde a su función.
- **Inconvenientes:** la API futura no forma parte de `depends_on` y debe
  comunicar con claridad que PostgreSQL no está disponible.
- **Coste de adopción:** bajo.
- **Coste de mantenimiento:** bajo; el Compose inicial contiene una sola
  dependencia y no duplica configuraciones de desarrollo de las aplicaciones.
- **Riesgos:** tratar `localhost` como sustituto de la configuración externa de
  producción. Se mitiga usando variables de entorno y validación explícita.

### Alternativa C — Instalar PostgreSQL en cada host

Usar el gestor de paquetes de cada sistema para ejecutar PostgreSQL directamente.

- **Ventajas:** elimina Docker del ciclo diario.
- **Inconvenientes:** versión, extensiones, puertos, directorios de datos y
  operaciones varían por máquina; el onboarding y la recuperación dejan de ser
  declarativos.
- **Coste de adopción:** bajo inicialmente.
- **Coste de mantenimiento:** medio o alto por deriva entre equipos.
- **Riesgos:** errores que no se reproducen fuera de una máquina concreta.

### No cambiar

No existe aún un procedimiento reproducible para arrancar PostgreSQL real. Esto
bloquea migraciones, integración de consultas y el primer vertical slice.

## Comparación

Compose declara dependencias versionables y permite levantar un entorno local con
un ciclo de vida común. Un contenedor iniciado no equivale necesariamente a un
servicio listo; un `healthcheck` permite expresar esa diferencia. La alternativa
A adelanta una imagen y una topología que corresponden a la decisión de
contenedores y despliegue. La C reduce pasos aparentes, pero desplaza la
complejidad al estado no versionado de cada host.

La alternativa B satisface la paridad que importa ahora —PostgreSQL real,
configuración externa, persistencia, salud y migraciones explícitas— y preserva
el flujo nativo requerido por Go y Expo.

## Recomendación

**Opinión/recomendación:** alternativa B. Usar Docker Compose solo para las
dependencias de infraestructura que el desarrollo necesita hoy. Mantener API Go,
web Expo e iOS/Android fuera de Docker en desarrollo. No se crea un contenedor
del frontend: el cliente web se servirá con Expo en el host y los targets móviles
necesitan herramientas y simuladores nativos.

## Decisión del usuario

**Aceptada:** alternativa B, con las siguientes reglas:

- Docker Compose operará inicialmente únicamente PostgreSQL;
- API Go y cliente Expo (web, iOS y Android) se ejecutarán en el host durante el
  desarrollo;
- no se contenerizará ningún frontend localmente;
- una imagen de la API y, si aporta valor, un artefacto o contenedor web serán
  decisiones posteriores de empaquetado y despliegue;
- el servicio PostgreSQL usará una imagen oficial fijada a una versión exacta al
  implementar el entorno, nunca `latest`;
- los datos residirán en un volumen nombrado; eliminarlos requerirá un comando
  explícito y documentado;
- la publicación del puerto será solo para loopback y el servicio expondrá un
  `healthcheck` basado en `pg_isready`;
- Compose tendrá su contrato de variables local propio, con `.env` ignorado y
  `.env.example` versionado conforme a ADR-0017;
- `goose` se ejecutará como comando separado contra PostgreSQL, nunca como efecto
  del arranque normal de la API;
- no habrá datos semilla funcionales hasta que se cierre el primer vertical slice
  de producto.

## Consecuencias

### Positivas

- PostgreSQL real es reproducible y no depende de una instalación manual.
- El equipo aprende ciclo de vida, volúmenes, puertos, salud y recuperación sin
  añadir servicios ajenos a la necesidad actual.
- Go y Expo conservan el feedback rápido y sus herramientas nativas.
- La separación entre dependencia, aplicación y migración hace visibles sus
  responsabilidades operativas.

### Negativas y deuda aceptada

- La API futura necesita diagnosticar una base no disponible sin depender de
  `depends_on`.
- El entorno de desarrollo no valida todavía la imagen final de la API.
- El cliente web no ejercita todavía su estrategia de entrega final.
- El volumen local puede conservar un esquema o datos obsoletos; resetearlo será
  una acción destructiva y explícita.

## Validación

La implementación de Fase 1 deberá demostrar que:

- un clon limpio puede crear el contrato local, levantar PostgreSQL y comprobar
  su estado saludable mediante comandos documentados;
- PostgreSQL conserva datos tras detener y volver a levantar el servicio;
- el procedimiento explícito de reset elimina los datos y deja una base limpia;
- las migraciones `goose` alcanzan la última versión desde una base vacía y se
  repiten sin cambios cuando ya están aplicadas;
- un proceso Go en el host puede conectarse con configuración externa y reporta
  un error accionable cuando PostgreSQL no está disponible;
- no hay secreto en Git ni puertos publicados fuera de loopback;
- no se añade Redis/Valkey, MinIO, observabilidad, Kubernetes ni contenedores de
  cliente sin ADR posterior;
- la documentación contiene arranque, parada, inspección, logs, migración,
  limpieza y diagnóstico básico.

## Disparadores de revisión

- un requisito reproducible demuestra que el bucle local de la API debe probar
  su imagen en cada cambio;
- hay varios servicios locales con dependencias coordinadas que justifican
  ampliar Compose;
- Expo o sus herramientas exigen una estrategia de desarrollo distinta;
- la base de datos local deja de representar contratos relevantes de producción;
- se adopta Kubernetes o un servicio gestionado que cambie de forma material el
  ciclo local;
- un incidente de pérdida de datos o exposición de puerto muestra que el ciclo
  de recuperación o la configuración son insuficientes.

## Documentación afectada

- [TECHNICAL_BASELINE.md](../governance/TECHNICAL_BASELINE.md)
- [DEVELOPMENT.md](../engineering/DEVELOPMENT.md)
- [DEPLOYMENT.md](../operations/DEPLOYMENT.md)
- [DECISIONS.md](../governance/DECISIONS.md)
- [DECISIONS_TO_REVISIT.md](../governance/DECISIONS_TO_REVISIT.md)
- [LEARNING.md](../project/LEARNING.md)
- [CHANGELOG.md](../../CHANGELOG.md)

## Fuentes técnicas

- [Docker Compose: desarrollo y dependencias](https://docs.docker.com/compose/intro/features-uses/)
- [Docker Compose: orden de arranque y salud](https://docs.docker.com/compose/how-tos/startup-order/)
- [Docker Compose: referencia de servicios](https://docs.docker.com/reference/compose-file/services/)
- [Imagen oficial de PostgreSQL](https://hub.docker.com/_/postgres)
