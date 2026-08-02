# ADR-0069: Mantener los borradores de torneo solo en local

- **Estado:** Aceptado
- **Fecha:** 2026-08-01
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** ADR-0031 y ADR-0052, exclusivamente en la transferencia y persistencia servidor de borradores
- **Superado por:** Ninguno

## Problema

Una persona debe poder preparar una liga antes de comprometerse a crear una
cuenta, pero el borrador no debe convertirse en un recurso remoto ni en un
estado de una liga.

## Contexto y restricciones

- Una liga creada nace visible y editable en `published` hasta que se inicia
  (ADR-0040).
- Solo una cuenta verificada puede crearla y publicar datos en el servidor.
- El cliente ya dispone de almacenamiento local multiplataforma.
- El producto no necesita recuperar un borrador en otro dispositivo antes de
  crear la liga.

## Criterios de decisión

1. Distinguir preparación local de una liga persistida.
2. Reducir datos temporales, purgas y superficie HTTP del backend.
3. Mantener el acceso como frontera previa a la publicación.
4. Permitir retomar un borrador al reiniciar la misma instalación.

## Alternativas

### A — Transferir el borrador a una cuenta pendiente

- **Ventajas:** permite continuar tras verificar desde otro dispositivo.
- **Inconvenientes:** añade tablas, caducidad, purga, contratos y casos de
  recuperación para datos que todavía no son producto persistido.
- **Coste de mantenimiento:** medio.

### B — Borrador local hasta publicar

- **Ventajas:** separa con claridad preparación y liga, evita estado temporal
  en PostgreSQL y sirve igual a acceso local o federado.
- **Inconvenientes:** no se sincroniza entre dispositivos y puede perderse al
  limpiar el almacenamiento de la aplicación.
- **Coste de mantenimiento:** bajo.

### No cambiar

- **Consecuencias:** se conserva la complejidad remota de un dato que el
  producto no necesita recuperar fuera del dispositivo.

## Comparación

La sincronización entre dispositivos no es un requisito del primer corte. B
preserva la reducción de fricción sin introducir un recurso remoto efímero.

## Recomendación

**Opinión/recomendación:** alternativa B, por ser la solución mínima suficiente.

## Decisión del usuario

**Aceptada el 2026-08-01:** el borrador de torneo existe solo en el cliente.
Se persiste localmente para la misma instalación, se conserva al navegar a
registro o acceso y se elimina tras publicar con éxito. No se envía con el alta,
no se asocia a una cuenta pendiente y no tiene endpoints ni tablas de servidor.

## Consecuencias

### Positivas

- La cuenta verificada es la única frontera para crear una liga.
- El backend solo conserva ligas reales y sus equipos.
- Google y credenciales locales comparten el mismo flujo de creación.

### Negativas y deuda aceptada

- El borrador no sobrevive a un cambio de dispositivo ni a la limpieza de datos
  locales.

## Validación

- Un invitado puede cerrar y reabrir la aplicación y retomar su borrador local.
- Registro y verificación no escriben filas de borrador en PostgreSQL.
- Publicar elimina el borrador local solo después de recibir la liga creada.

## Disparadores de revisión

- Una necesidad demostrada de continuar la preparación desde otro dispositivo.
- Colaboración o edición simultánea antes de publicar.

## Documentación afectada

- [Producto](../project/PRODUCT.md)
- [API](../engineering/API.md)
- [Modelo inicial de datos](../engineering/INITIAL_DATA_MODEL.md)
- [Decisiones](../governance/DECISIONS.md)
- [Aprendizaje](../project/LEARNING.md)
