# ADR-0031: Conservar borradores previos al acceso hasta verificar la cuenta

- **Estado:** Aceptado
- **Fecha:** 2026-07-26
- **Decisor:** Usuario
- **Propietario del análisis:** Codex
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

El primer flujo de creación debe permitir que una persona prepare un torneo de
fútbol de liga y sus equipos antes de registrarse. Exigir acceso al inicio añade
fricción; conservar el borrador exclusivamente en el dispositivo puede perderlo
si la verificación de correo continúa en otro navegador o dispositivo.

## Contexto y restricciones

- El primer producto gestiona torneos internos e informales; el organizador crea
  los equipos.
- Un torneo persistido solo puede ser creado y publicado por una cuenta
  verificada.
- La identidad propia, credenciales locales y acceso con Apple o Google ya están
  aceptados en [ADR-0010](0010-own-identity-with-federated-login.md).
- No se introducirá persistencia de borradores anónimos en el servidor.
- La política exacta de contraseñas, sesión, OAuth/OIDC, limpieza y contratos
  HTTP se decidirá antes de implementarla.

## Criterios de decisión

1. reducir la fricción antes de comprometerse a registrarse;
2. no perder el borrador al completar el alta desde otro dispositivo;
3. no conceder acciones de negocio a una cuenta no verificada;
4. limitar datos anónimos, abuso y trabajo operativo;
5. mantener una implementación gradual y verificable.

## Alternativas

### Alternativa A — Exigir acceso antes de preparar el torneo

- **Ventajas:** implementación, seguridad operativa y limpieza de datos más
  simples.
- **Inconvenientes:** obliga al registro antes de que la persona compruebe el
  valor del flujo.
- **Coste de adopción:** bajo.
- **Coste de mantenimiento:** bajo.
- **Riesgos:** abandono temprano del flujo de creación.

### Alternativa B — Borrador solo local hasta publicar

- **Ventajas:** no persiste datos de personas sin cuenta en el servidor;
  experiencia inicial sin registro.
- **Inconvenientes:** un enlace de verificación abierto en otro navegador o
  dispositivo no puede recuperar el borrador.
- **Coste de adopción:** bajo.
- **Coste de mantenimiento:** bajo.
- **Riesgos:** pérdida de trabajo y soporte difícil de explicar.

### Alternativa C — Cuenta pendiente y borrador asociado en servidor

Tras enviar el alta local se crea una cuenta pendiente de verificar. El borrador
se asocia a ella, caduca y solo pasa a ser publicable al verificar el correo.

- **Ventajas:** conserva el borrador entre dispositivos; mantiene el registro
  opcional hasta consolidar; no trata el borrador como anónimo.
- **Inconvenientes:** añade estados de cuenta y borrador, expiración y limpieza.
- **Coste de adopción:** moderado.
- **Coste de mantenimiento:** moderado; requiere límites contra abuso y una tarea
  de purga.
- **Riesgos:** una implementación que autorice erróneamente a una cuenta
  pendiente o no borre datos expirados.

### Alternativa D — Borrador anónimo persistido en servidor

- **Ventajas:** el borrador sobrevive sin iniciar el alta.
- **Inconvenientes:** exige identificar, recuperar, expirar y proteger datos
  anónimos, además de controles de abuso antes de conocer al usuario.
- **Coste de adopción:** alto.
- **Coste de mantenimiento:** alto.
- **Riesgos:** spam, enumeración, fuga o retención innecesaria de datos.

### No cambiar

Se mantendría el registro y la creación del perfil solo después de verificar el
correo. El borrador no podría asociarse de forma fiable a la persona que abrió
el enlace de verificación en otro dispositivo.

## Comparación

La alternativa A minimiza el sistema, pero contradice el objetivo de poder
probar la creación antes del registro. B ofrece esa prueba con una pérdida de
datos previsible. D resuelve la recuperación, pero convierte el primer corte en
un sistema de datos anónimos. C conserva la experiencia buscada con una única
complejidad adicional acotada: datos temporales ligados a una cuenta pendiente.

## Recomendación

**Recomendación:** alternativa C. El borrador local permite empezar sin sesión;
la creación de la cuenta pendiente permite conservarlo de forma autenticable; la
verificación sigue siendo la puerta para publicar y gestionar el torneo.

## Decisión del usuario

**Aceptada el 2026-07-26:** adoptar la alternativa C.

- Un invitado puede preparar localmente un borrador de torneo.
- Al registrarse con email y contraseña se crea una cuenta pendiente y el
  borrador se asocia a ella en el servidor.
- Una cuenta pendiente no recibe sesión de producto ni permisos de negocio.
- La verificación del email activa la cuenta; solo entonces puede publicar el
  torneo y operar como organizador.
- Los borradores y cuentas pendientes caducan y se eliminan con una política que
  se fijará antes de la implementación.
- Apple y Google autentican mediante su flujo federado; un `username` no es una
  credencial social. Si hace falta, el nombre de usuario se elige después de que
  el proveedor haya acreditado la identidad.

## Consecuencias

### Positivas

- El primer contacto con el producto no exige cuenta.
- La verificación de correo puede completarse desde otro dispositivo sin perder
  el borrador.
- La publicación conserva una frontera clara: cuenta verificada.

### Negativas y deuda aceptada

- Existen estados `pending_verification` y `verified` que deben ser explícitos.
- Hay que diseñar retención, purga, límites de frecuencia y mensajes que no
  enumeren cuentas.
- El modelo de sesión debe distinguir correctamente registro pendiente,
  verificación y acceso autenticado.

## Validación

- Un invitado puede crear un borrador local sin enviar datos al servidor.
- Un alta local mueve el borrador a una cuenta pendiente sin publicarlo.
- La misma cuenta recupera el borrador tras verificar desde otro dispositivo.
- Una cuenta pendiente no puede publicar, leer datos protegidos ni obtener una
  sesión de producto.
- Una cuenta o borrador caducado deja de poder recuperarse y se elimina conforme
  a la política aprobada.

## Disparadores de revisión

- Abuso o coste excesivo de cuentas pendientes o borradores.
- Necesidad de recuperar borradores sin iniciar un alta.
- Requisitos regulatorios de retención o borrado.
- Cambio de la política de verificación o de proveedores de identidad.

## Documentación afectada

- [PRODUCT.md](../project/PRODUCT.md)
- [IDENTITY.md](../engineering/IDENTITY.md)
- [ADR-0010](0010-own-identity-with-federated-login.md)
- [DECISIONS.md](../governance/DECISIONS.md)
