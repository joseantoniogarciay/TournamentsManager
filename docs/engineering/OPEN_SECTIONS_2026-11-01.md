# Secciones abiertas — implementación de ADR-0178

Fecha: 2026-11-01. Implementado en el cliente universal; QA visual pendiente.

## Alcance

| Área | Presentación implementada |
| --- | --- |
| Equipos y participantes | Cabecera con contador localizado, filas con separador, añadir junto a la lista, invitación separada. Lista primero en móvil, invitación a la derecha en web ancha. |
| Administradores | Filas abiertas; conserva búsqueda, alta y retirada con confirmación. |
| Sesiones | Explicación y acciones abiertas; sesiones separadas por líneas; paginación sin caja. |
| Notificaciones | Mensaje y fecha sobre canvas con separador; conserva abrir y eliminar. |
| Ediciones | Lista con separadores; conserva edición actual, estado, campeones y paginación. |
| Ficha de torneo | Resumen, sede, preparación del formato y ayudas abiertas. Los encuentros conservan Card. |
| Inicio y pantallas de invitado | Hero, explicación y pasos abiertos; torneos recientes y sugerencias conservan sus fichas. |
| Formularios | Creación, inscripción por enlace, repetición, programación/sede, acceso, registro, contraseña y recuperación usan secciones abiertas. |
| Ajustes y lectura | Apariencia y aviso de notificaciones abiertos; privacidad y condiciones como documento continuo con títulos y separadores. |
| Clasificación | El estado informativo sin tabla queda abierto. Tabla, scroll y excepción de escalado se conservan. |

`ContentSection` y `ListRow` viven en `shared/ui` y consumen exclusivamente tokens
semánticos. Ambas reservan 20 px de margen horizontal y 12 px verticales de
contenido. La sección tiene gap interno y un separador superior opcional; la fila
solo borde inferior. Sin fondos, radios ni elevaciones. No se anidan entre sí.

La variante `ResponsiveColumns.sidebar-end` conserva el orden React/DOM y sitúa
el segundo hijo en 360 px laterales a partir de 960 px CSS, usando la misma
compensación de márgenes existente. En nativo y web estrecha se apilan. No se
añaden anchos mínimos ni remounts al redimensionar.

## Checklist de cliente

- Revisión contra todos los archivos modificados: tokens, primitivas compartidas,
  márgenes de 20 px y ausencia de padding duplicado en scroll/Screen.
- Contador de equipos/participantes en es/en/it/fr; no se importan textos ficticios
  de la maqueta. Los nombres proceden de los datos actuales.
- Se conservan escalado accesible, controles existentes de 44 px, estados de
  carga, validación, bloqueos con motivo y diálogos compartidos.
- Navegación, URLs, cabeceras y NativeTabs conservan su implementación nativa.
- Ninguna operación OpenAPI ni adaptador HTTP modificado. Se conservan funciones
  generadas, apiFetch, feedback seguro, cancelaciones y permisos anteriores.
- No se alteran endpoints, spans, inventario de observabilidad ni lógica backend.
- La búsqueda de Card restante conserva fichas de torneo, encuentros, cuadros,
  sugerencias/apoyo, el control independiente de consentimiento sin consumidores
  actuales y errores recuperables. No hay sustitución global de Card.

## Verificación

- Typecheck y Oxlint correctos.
- Exportación web Expo correcta: 44 rutas.
- Suite Node completa: 153 pruebas correctas, cero fallos, tras permitir los
  sockets localhost de Miniflare y Scalar fuera del sandbox.
- El primer intento detectó un fixture previo de SessionNavigator sin el hook
  usePreferences que ya consume la aplicación. Se completa solo ese mock; las
  aserciones de sustitución de sesión y cancelación se conservan.
- Se comprueban formato de los archivos de cliente y diff sin errores de espacios.

No se acredita revisión visual web o nativa. El intento de CUA informa que el Mac
está bloqueado y no se puede desbloquear automáticamente, y que el navegador local
está deshabilitado en Work. Quedan pendientes equipos con nombres largos, teclado,
último control, claro/oscuro, texto ampliado, cuatro locales, rotación y Split View.
La compilación y la revisión del código no sustituyen esas comprobaciones.

No se arrancan servicios local/dev ni tareas programadas, no se activa
observabilidad y no se publica ni despliega. No se crea una rama temporal.

## Retrospectiva técnica

Las listas simples necesitaban perder su contenedor, no añadir sombras a cada
tarjeta. Dos primitivas pequeñas permiten hacerlo conservando la identidad y el
comportamiento. Revisar cada consumidor evita degradar unidades que sí tienen
sentido como ficha. El orden estable de la lista también evita que una mejora de
composición introduzca pérdida de estado al cambiar de ancho.

## Gate de publicación solicitado

El usuario autoriza «Despliega» tras revisar la comparación antes/después.
`make verify` completo termina correctamente en develop: formato, lint, tipos,
regresiones, seguridad operacional, exportación web, generación determinista,
compilación Go y govulncheck. La integración PostgreSQL se omite localmente por
no configurar TM_INTEGRATION_DATABASE_URL y se comprobará en CI con su servicio
desechable. No se levantan local/dev ni observabilidad. La publicación web se
versiona como v1.15.3 conforme a ADR-0119; conserva el límite de QA visual.

## Publicación web

v1.15.3 activada en Cloudflare, SHA `ab32cea8a01e96f965cd08084aa907f26488bdc6`.
Ambas CI pasan con PostgreSQL. Seis checks de preview y 22 productivos verifican
el artefacto; conserva el límite de QA visual. La autorización específica de
bundles/mapas a PostHog se recibe antes de ejecutar su subida. API, datos y
binarios móviles no se despliegan. [Cierre y rollback](../operations/DEPLOYMENT.md#publicación-v1153--2026-11-01).
