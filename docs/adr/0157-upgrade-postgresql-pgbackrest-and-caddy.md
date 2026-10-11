# ADR-0157: Actualizar PostgreSQL, pgBackRest y Caddy

- **Estado:** Aceptado
- **Fecha:** 2026-10-11
- **Decisor:** Usuario, mediante autorización explícita

## Problema, alternativas y decisión

Producción usa PostgreSQL 18.4 y pgBackRest 2.59.1; Caddy del Mac usa 2.11.4.
La revisión del 11 de octubre identifica correcciones oficiales en PostgreSQL
18.6, pgBackRest 2.59.3 y Caddy 2.11.7. Mantener versiones evita cambios pero
conserva defectos corregidos. Actualizar esos tres componentes mantiene las líneas
actuales, con coste de prueba medio (restauración y transición de la base).
Actualizar toda la infraestructura ampliaría riesgo y mantenimiento sin necesidad.

Recomendación: las tres versiones exactas, con imagen previa y binario Caddy
conservados, restauración aislada antes de cambiar PostgreSQL y verificaciones
de salud después. No modificar datos funcionales, esquema, durabilidad ni retención.

El usuario decide «Hacemos los 3 porque el que tarda a las 11 se puede hacer ya
también no hace falta que sean 7 dias exactos». Autoriza estas actualizaciones
operativas, incluida producción, y una excepción puntual para pgBackRest 2.59.3
por las horas restantes. ADR-0138 conserva su regla para otras versiones;
esta excepción no habilita otros paquetes jóvenes.

## Validación y condiciones

Comprobar versiones reales, backup reciente y subclaves cifradas sin imprimirlas.
Verificar checksum oficial de pgBackRest y paquete Homebrew Caddy. Construir
imagen PostgreSQL 18.6 + pgBackRest 2.59.3, importarla en containerd, restaurar
una copia anterior en almacenamiento efímero sin montar el PVC de datos. Solo
tras ese gate cambiar la imagen de PostgreSQL conservando ambos PVC. Comprobar
versión, disponibilidad, archivado WAL, backup completo nuevo y restauración.
Actualizar referencias dev/CI sin iniciar esos servicios ni observabilidad.

Si las subclaves existentes muestran debilidad, actualizar el binario no basta:
conservar repositorio anterior y preparar la remediación oficial sin destruirlo.
No afirmar ausencia absoluta de debilidad por una comprobación heurística.
Caddy conserva configuración, certificados y acceso de borde; validar su
configuración antes de reiniciar el servicio y comprobar hosts públicos después.

## Documentación, revisión y fuentes

Registrar versiones y evidencia en Deployment, runbooks y aprendizaje.
Revisar ante fallo de restauración, archivado o disponibilidad; revertir binario
o imagen dentro de la misma línea sin borrar volúmenes ni backups.

- [Revisión](../operations/DEPENDENCY_REVIEW_2026-10-11.md).
- [pgBackRest: condición y remediación](https://pgbackrest.org/news.html#weak-encryption).
- [PostgreSQL 18.6](https://www.postgresql.org/docs/release/18.6/).
- [Caddy 2.11.7](https://github.com/caddyserver/caddy/releases/tag/v2.11.7).
