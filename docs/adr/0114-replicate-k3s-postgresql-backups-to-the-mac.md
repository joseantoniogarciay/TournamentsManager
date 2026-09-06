# ADR-0114: Replicar los backups PostgreSQL de K3s al Mac

- **Estado:** Aceptado
- **Fecha:** 2026-08-30
- **Decisor:** Usuario, mediante decisión explícita
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

El PostgreSQL de `prod` vivirá en un PVC local de la VM K3s. pgBackRest necesita
un repositorio distinto del volumen de datos para conservar copias físicas y WAL,
pero la VM no comparte carpetas con el Mac por la frontera aceptada en ADR-0110.
Hace falta llevar una copia cifrada de ese repositorio a la ubicación doméstica
sincronizada sin poner una clave de acceso al Mac dentro de Kubernetes.

## Contexto y restricciones

- ADR-0111 exige para `prod` pgBackRest cifrado, WAL archivado, copia completa
  semanal, incrementales diarios y una restauración aislada probada antes de
  publicar los hosts de producción.
- ADR-0108 ya valida ese patrón en `dev`, pero su ruta de iCloud Drive es
  accesible directamente desde Docker Desktop y no desde la VM Ubuntu.
- ADR-0110 prohíbe configurar una carpeta compartida entre UTM y la VM.
- El acceso SSH desde el Mac a la VM usa una clave y el operador introduce
  interactivamente cualquier contraseña de `sudo`; ni claves privadas ni
  contraseñas entran en Git, ConfigMaps o logs.
- Un PVC de `local-path` protege frente al reinicio de un Pod, no frente a la
  pérdida del host ni a una eliminación accidental del volumen.

## Criterios de decisión

1. mantener el repositorio de pgBackRest separado del PVC de datos y cifrado;
2. copiarlo a la ubicación doméstica sincronizada sin romper la frontera de la
   VM ni exponer una credencial del Mac a los Pods;
3. programar copias y transferencias recuperables, con salida observable;
4. evitar un operador, un servidor de backup o un proveedor cloud nuevos para
   una carga doméstica inicial;
5. no presentar la copia como independencia real de Mac, cuenta o proveedor.

## Alternativas

### Alternativa A — PVC de repositorio y réplica iniciada desde el Mac por SSH

PostgreSQL conserva un PVC de datos y otro PVC para el repositorio pgBackRest.
El Mac programa la copia completa o incremental y, una vez correcta, obtiene
una exportación del repositorio por SSH desde la VM. El destino se prepara fuera
de la ruta publicada y se sustituye de forma atómica solo después de una
transferencia completa. La VM permite únicamente wrappers `sudo`, propiedad de
`root`, con subcomandos y rutas fijos: iniciar pgBackRest dentro del Pod y
exportar el repositorio. No aceptan comandos, rutas ni argumentos arbitrarios.

- **Ventajas:** conserva los secretos del backup y las claves de operador fuera
  de Kubernetes; no añade una carpeta UTM, un servicio de red ni credenciales
  persistentes en los Pods; el Mac puede reintentar la transferencia al volver
  a estar disponible.
- **Inconvenientes:** requiere mantener dos scripts acotados en la VM y un job
  de `launchd` en el Mac; una copia completa del repositorio puede ser lenta al
  crecer y la ubicación sincronizada sigue compartiendo Mac, cuenta y proveedor.
- **Coste de adopción:** medio: manifiestos, wrappers de host, job local y una
  restauración aislada que compruebe la réplica recibida.
- **Coste de mantenimiento:** bajo mientras el repositorio sea pequeño; se
  revisará si el volumen o la transferencia crecen.
- **Riesgos:** un wrapper demasiado genérico convertiría `sudo` en acceso de
  administración; una transferencia parcial podría parecer recuperable. Se
  mitigan fijando comandos/rutas, publicando el destino solo al final y
  verificando una restauración desde la réplica del Mac.

### Alternativa B — Carpeta UTM compartida con la ubicación sincronizada

Montar directamente en la VM una carpeta del Mac que iCloud Drive sincroniza.

- **Ventajas:** configuración y transferencia muy simples.
- **Inconvenientes:** contradice la decisión de ADR-0110 de no compartir datos
  entre host y VM; mezcla permisos y semántica de un filesystem sincronizado
  con el recorrido crítico de archivado WAL.
- **Coste de adopción y mantenimiento:** bajo al inicio, con diagnósticos más
  difíciles ante desconexiones o conflictos de sincronización.
- **Riesgos:** degradar la frontera operativa aceptada y dejar WAL a merced de
  un montaje del host no diseñado como almacenamiento de base de datos.

### Alternativa C — Repositorio remoto directo de pgBackRest

Usar desde la VM un proveedor de almacenamiento remoto compatible como destino
primario del repositorio.

- **Ventajas:** elimina la réplica intermedia y puede mejorar la separación
  física.
- **Inconvenientes:** exige elegir proveedor, coste, credenciales, cifrado,
  retención y operación de red antes de que la carga lo justifique.
- **Coste de adopción y mantenimiento:** medio o alto.
- **Riesgos:** adelantar una decisión de proveedor y de gasto que la Fase 4
  doméstica no necesita aún.

### No cambiar

Mantener solo los PVC de la VM no cumple el gate de ADR-0111: una pérdida de
host o de volumen no tendría copia recuperable fuera de ese almacenamiento.

## Comparación

La B reduce trabajo inmediato a costa de romper una frontera explícita y de
acoplar el archivado WAL a un montaje compartido. La C podría aportar más
independencia, pero abre un proveedor y gasto nuevos antes de medir la carga.
La A mantiene pgBackRest cerca del `PGDATA`, deja el control de la copia fuera
del clúster y limita el nuevo privilegio del host a operaciones cerradas.

## Recomendación

**Recomendación:** A. Es la solución mínima suficiente para trasladar la copia
cifrada sin diluir la frontera VM/Mac ni instalar otro sistema de backup. Su
principal deuda aceptada es que no ofrece independencia frente al Mac, la cuenta
ni el proveedor de sincronización.

## Decisión del usuario

**Aceptada el 2026-08-30:** alternativa A. `prod` tendrá PVC propios para datos
y repositorio pgBackRest. El Mac iniciará por SSH el backup y la réplica del
repositorio cifrado hacia su ubicación doméstica sincronizada. Los Pods no
recibirán una clave privada del Mac y la VM expondrá solo wrappers `sudo` con
operaciones y rutas fijas.

**Aclaración aplicada el 2026-08-30:** la jerarquía doméstica será
`FastTourney/postgresql-backups/dev` y
`FastTourney/postgresql-backups/prod`. El directorio `prod` pertenece solo a la
réplica recibida desde K3s; no reutiliza ni la clave ni el repositorio de `dev`.

## Consecuencias

### Positivas

- La API, el PostgreSQL y la réplica no comparten secretos operativos.
- La transferencia se puede reintentar sin cambiar datos activos ni publicar un
  repositorio parcial como recuperable.
- La operación mantiene visibles StatefulSet, PVC, pgBackRest, SSH y el límite
  entre el host y Kubernetes, coherente con la Fase 4.

### Negativas y deuda aceptada

- La VM conserva dos PVC locales, ambos dependientes del mismo host hasta que
  la réplica haya terminado y sincronizado.
- El calendario vive en el Mac, por lo que una suspensión prolongada retrasa el
  RPO; el job debe registrar y alertar del último éxito antes de publicar.
- La primera versión exportará el repositorio completo por simplicidad. Si su
  tamaño o ventana de copia deja de ser aceptable, se evaluará una réplica
  incremental o un repositorio remoto.

## Validación

1. `pgbackrest check` confirma que el WAL llega al PVC de repositorio cifrado.
2. Una copia completa y un incremental aparecen en `pgbackrest info` dentro de
   la VM.
3. El job del Mac recibe la exportación en un destino temporal y solo sustituye
   el destino recuperable tras verificar la transferencia.
4. Una restauración aislada desde la réplica recibida arranca PostgreSQL sin
   montar ni modificar el PVC activo de `prod`.
5. Un usuario SSH no puede convertir el wrapper en ejecución arbitraria como
   `root`, ni leer rutas fuera del repositorio fijado.

## Disparadores de revisión

- El repositorio o su transferencia diaria supera la ventana disponible.
- El RPO/RTO, varios operadores, una pérdida de Mac/cuenta o un requisito de
  inmutabilidad exigen una segunda ubicación o proveedor independiente.
- Un incidente muestra que los wrappers de host son difíciles de auditar o
  mantener.
- La observabilidad detecta fallos de backup o réplica no accionables.

## Documentación afectada

- [Decisiones](../governance/DECISIONS.md)
- [Decisiones a revisar](../governance/DECISIONS_TO_REVISIT.md)
- [Despliegue](../operations/DEPLOYMENT.md)
- [Roadmap](../project/ROADMAP.md)
- [Aprendizaje](../project/LEARNING.md)
- [Changelog](../../CHANGELOG.md)
