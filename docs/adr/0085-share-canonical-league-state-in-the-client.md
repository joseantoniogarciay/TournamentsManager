# ADR-0085: Compartir el estado canónico de cada liga en el cliente

- **Estado:** Aceptado
- **Fecha:** 2026-08-10
- **Decisor:** Usuario
- **Propietario del análisis:** Codex
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

Una misma liga puede mostrarse en detalle, equipos, clasificación y biblioteca.
Tras una mutación que devuelve la liga actualizada, esas pantallas conservaban
copias locales y podían mostrar estados distintos o volver a consultar la red.

## Contexto y restricciones

- El contrato devuelve una proyección completa de `PublicLeague` en las
  mutaciones de inicio, cancelación, cierre y resultados.
- El estado sigue siendo local por defecto según ADR-0055; no se necesita
  sincronización entre dispositivos en este corte.
- La solución no introduce dependencias de estado ni cambia el contrato HTTP.

## Criterios de decisión

1. actualizar todas las vistas abiertas de la misma liga sin refetch;
2. conservar una única entidad canónica por ID;
3. mantener bajo el coste de adopción y mantenimiento;
4. dejar explícita la frontera con futuras actualizaciones remotas.

## Alternativas

### Alternativa A — Copias locales y recargas por pantalla

- **Ventajas:** no crea infraestructura compartida.
- **Inconvenientes:** estados divergentes y llamadas repetidas tras mutaciones.
- **Coste de mantenimiento:** crece con cada vista de la liga.

### Alternativa B — Almacén reactivo de ligas por ID

- **Ventajas:** la respuesta de una mutación actualiza todas las suscripciones
  de esa liga en memoria, sin otra llamada de red.
- **Inconvenientes:** requiere que cada mutación publique su proyección en el
  almacén.
- **Coste de mantenimiento:** bajo; usa React y `useSyncExternalStore` ya
  disponibles.

### Alternativa C — Librería global de caché o bus de eventos

- **Ventajas:** ofrece más políticas de caché y eventos.
- **Inconvenientes:** añade una dependencia y una abstracción general antes de
  necesitar revalidación compleja u offline.
- **Coste de mantenimiento:** medio.

### No cambiar

Las pantallas seguirían recargando o manteniendo representaciones obsoletas.

## Comparación

La alternativa B satisface la actualización local inmediata y evita las
peticiones duplicadas. La C no aporta valor proporcional mientras no haya
sincronización remota, paginación de caché compleja u offline-first.

## Recomendación

**Recomendación:** alternativa B.

## Decisión del usuario

**Aceptada el 2026-08-10:** usar un almacén reactivo con una entidad canónica
por `leagueId`. Las vistas se suscriben por ID y las mutaciones publican la
respuesta actualizada. No se incorpora un bus de eventos ni una librería de
caché externa en este corte.

## Consecuencias

### Positivas

- Detalle, cajas y rutas auxiliares reflejan la misma mutación inmediatamente.
- La respuesta ya obtenida de la API evita una recarga para propagar el cambio.

### Negativas y deuda aceptada

- Un cambio producido en otro dispositivo no llega automáticamente.
- La persona usuaria conserva refresh explícito como mecanismo de actualización
  remota hasta decidir revalidación o tiempo real.

## Validación

- Cancelar una liga actualiza el detalle y su caja sin otra petición.
- Iniciar, cerrar y registrar resultados sustituyen la misma entidad canónica.
- Abrir equipos o clasificación tras cargar detalle reutiliza la entidad local.

## Disparadores de revisión

- Actualizaciones desde otros dispositivos que requieran inmediatez.
- Offline-first, conflictos de edición o caché paginada compleja.
- Evidencia de que una librería de datos reduce complejidad neta.

## Documentación afectada

- [ADR-0055](0055-use-feature-first-client-architecture-and-platform-adaptive-navigation.md)
- [PRODUCT.md](../project/PRODUCT.md)
- [LEARNING.md](../project/LEARNING.md)
- [DECISIONS.md](../governance/DECISIONS.md)
