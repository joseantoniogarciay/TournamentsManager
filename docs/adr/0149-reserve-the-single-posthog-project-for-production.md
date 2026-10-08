# ADR-0149: Reservar el único proyecto PostHog para producción

- **Estado:** Aceptado
- **Fecha:** 2026-10-04
- **Decisor:** Usuario: «Reservarlo para producción y apagar PostHog en beta»
- **Supera a:** ADR-0148 en los dos proyectos; ADR-0105 en entornos activos y analítica opcional; conserva el gate de ADR-0109

## Problema, evidencia y alternativas

El usuario comunica que su cuenta Free solo permite un proyecto. En Safari se
verifica que el proyecto **255144**, «Default project», está en EU Cloud y su
token público coincide con el configurado en el cliente. No se infiere el ID de
una clave ni se afirma que la cuota sea universal para todos los planes.

Compartir proyecto beta/prod y filtrar por entorno ahorra configuración, pero
mezcla datos y cuota. Contratar otro proyecto añade coste. Reservar el disponible
para producción mantiene aislamiento del tráfico futuro sin pagar; se pierde
la señal remota de beta, que puede diagnosticarse localmente durante pruebas.

## Recomendación y decisión explícita

Se recomienda reservar 255144 para producción. El usuario acepta esa opción en
la pregunta estructurada, tras comprobar la limitación de su cuenta.

Solo `production` con `EXPO_PUBLIC_POSTHOG_PRODUCTION_API_KEY` válida inicializa
el SDK. `local` y `development` quedan apagados incluso con claves válidas o
consentimiento previo. La clave existente se migra en los archivos privados de
configuración a producción; no se resetea el token, crea otro proyecto ni despliega.

Producción recoge únicamente fiabilidad mínima: excepciones JS, rechazos de
promesa y crashes nativos. No activa analítica de producto, vistas, correlación
API, replay, flags, push ni identificación. Se retiran los controles de analítica
de Inicio/Ajustes para no ofrecer una capacidad apagada; las preferencias antiguas
no reactivan el SDK. El histórico existente no se borra.

## Implementación, coste y validación

Se conserva el SDK fijado. Metro y Expo preparan debug IDs, maps Hermes, dSYM y
R8 solo para producción. Web vincula mapas al SHA, usa el proyecto CLI explícito
y elimina los mapas públicos antes de publicar. Credenciales administrativas
solo en build, nunca en el bundle ni en Git; el archivo preparado fija ID 255144
y host UE, pero requiere aún la clave personal CLI.

La prueba distribuible, reconstrucción nativa y entrega/simbolización se difieren
por petición del usuario. Antes de distribuir se validan JS/crash nativo en las
dos plataformas, datos mínimos, filtrado nativo, retención y gasto máximo 0 €.
Preparar configuración no acredita recepción real ni modifica servicios activos.

Retrospectiva: la separación de tráfico puede resolverse apagando el entorno de
pruebas, sin añadir un segundo proveedor ni confundir etiquetas con aislamiento.
