# ADR-0084: Publicar la búsqueda parcial de usernames para perfiles futuros

- **Estado:** Aceptado
- **Fecha:** 2026-08-10
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

La asignación de administradores debe permitir encontrar a una persona por una
parte de su username. El producto prevé perfiles y seguimiento entre cuentas,
por lo que esa búsqueda no puede depender de una sesión ni de una liga concreta.

## Contexto y restricciones

El username ya es público, único e inmutable en este corte (ADR-0034 y
ADR-0048), pero el contrato solo permitía comprobar su disponibilidad o asignar
uno exacto. La búsqueda descubre usernames existentes: es una exposición
deliberada distinta de enumerar email, contraseñas o métodos de acceso.

## Alternativas

### A — Búsqueda pública parcial de usernames

- **Ventajas:** resuelve la asignación y reutiliza el mismo recurso para perfiles
  y seguimiento futuros; funciona sin sesión.
- **Inconvenientes:** permite descubrir usernames públicos.
- **Coste de mantenimiento:** bajo: endpoint de solo lectura, resultados y
  límite acotados.

### B — Búsqueda solo para organizadores autenticados

- **Ventajas:** reduce la visibilidad inmediata.
- **Inconvenientes:** duplica una capacidad que perfiles públicos necesitarán y
  no permite descubrir personas fuera de una liga.
- **Coste de mantenimiento:** medio.

### C — Username exacto manual

- **Ventajas:** no añade endpoint.
- **Inconvenientes:** no satisface el flujo solicitado ni el producto social.
- **Coste de mantenimiento:** bajo, con experiencia insuficiente.

## Recomendación

**Recomendación:** A, con resultados de usernames verificados únicamente,
búsqueda que contenga el texto, mínimo de tres caracteres, máximo de veinte
resultados y límite de sesenta peticiones por IP y minuto. No se exponen email,
IDs ni métodos de acceso.

## Decisión del usuario

**Aceptada el 2026-08-10:** alternativa A. Cualquier persona, incluso sin
sesión, puede buscar usernames públicos para perfiles, seguimiento y selección
de administradores futuros.

## Consecuencias

- El contrato incorpora una colección pública de coincidencias.
- La asignación mantiene su autorización de organizador en el servidor.
- La búsqueda no autoriza ninguna acción sobre el perfil encontrado.
- El cliente espera antes de buscar, cancela la petición previa cuando cambia el
  texto y no deja que una respuesta antigua sustituya resultados recientes.
- Un límite de abuso (`429`) y cualquier fallo no recuperable se comunican con
  el banner seguro de la aplicación; no se muestran detalles del backend.

## Validación

- `rau` devuelve usernames verificados que contienen `rau`, con veinte o menos
  resultados.
- Menos de tres caracteres devuelve validación, y no se expone ningún dato que
  no sea el username.
- La pantalla modal cancela búsquedas obsoletas y solo asigna tras pulsar una
  fila.

## Disparadores de revisión

- Necesidad de bloquear perfiles, privacidad por cuenta o nombres de muestra.
- Volumen que requiera un índice específico de búsqueda parcial.

## Documentación afectada

- [Producto](../project/PRODUCT.md)
- [API](../engineering/API.md)
- [Decisiones](../governance/DECISIONS.md)
