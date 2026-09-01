# ADR-0115: Usar bookmarks de seguridad para publicar backups en iCloud

- **Estado:** Aceptado
- **Fecha:** 2026-09-01
- **Decisor:** Usuario, mediante decisión explícita
- **Propietario del análisis:** Codex como mentor técnico
- **Supera parcialmente a:** ADR-0114, solo en el mecanismo local de publicación
- **Superado por:** Ninguno

## Problema

El LaunchAgent de `prod` puede iniciar el backup por SSH, pero macOS rechaza que
un proceso de fondo altere directamente la carpeta protegida de iCloud Drive.
Dar Full Disk Access a Bash o a `launchd` resolvería el síntoma con un permiso
global, incompatible con el mínimo privilegio.

## Contexto y restricciones

- ADR-0114 conserva iCloud como ubicación doméstica sincronizada y exige que la
  publicación de la réplica no deje un destino parcial recuperable.
- La ejecución interactiva del operador sí puede acceder a iCloud Drive; el
  LaunchAgent no recibe ese permiso de TCC por herencia.
- Apple documenta que una aplicación sandboxed puede conservar acceso a una
  carpeta elegida por la persona con un security-scoped bookmark, que debe
  activarse y liberarse en cada uso.
- No se introducen credenciales cloud, carpeta compartida UTM, Full Disk Access
  ni permisos para rutas ajenas al staging y al directorio seleccionado.

## Criterios de decisión

1. conservar iCloud Drive y la frontera Mac/VM de ADR-0114;
2. limitar el permiso persistente a dos directorios elegidos explícitamente;
3. mantener staging, validación y publicación final separados;
4. permitir que `launchd` ejecute la copia sin depender de una terminal;
5. no añadir proveedor, cuenta ni coste recurrente.

## Alternativas

### A — Helper sandboxed con security-scoped bookmarks

Una utilidad macOS mínima muestra una vez dos selectores: staging local e iCloud
`postgresql-backups`. Conserva bookmarks privados, copia la réplica validada al
staging de iCloud y solo entonces sustituye `prod`.

- **Ventajas:** mantiene iCloud y mínimo privilegio; el permiso es revocable y
  no alcanza a Desktop, otras carpetas ni secretos.
- **Inconvenientes:** incorpora código nativo, firma local y una configuración
  interactiva inicial.
- **Coste de adopción y mantenimiento:** medio y acotado a un helper pequeño.
- **Riesgos:** un bookmark puede quedar obsoleto; el helper debe fallar seguro y
  pedir reconfiguración, sin reemplazar la réplica actual.

### B — Full Disk Access para el proceso de backup

- **Ventajas:** implementación rápida.
- **Inconvenientes:** autoriza acceso amplio a datos protegidos no relacionados.
- **Coste de adopción:** bajo; **mantenimiento y riesgo:** altos por la
  excepción de seguridad difícil de auditar.

### C — Sustituir iCloud por un proveedor con CLI

- **Ventajas:** automatización directa y potencial independencia adicional.
- **Inconvenientes:** contradice la preferencia actual y abre cuenta, coste,
  credenciales, cifrado y soporte de otro proveedor.
- **Coste de adopción y mantenimiento:** medio o alto.

### No cambiar

Mantener el LaunchAgent actual deja fallos recurrentes y no satisface el backup
automatizado de ADR-0111.

## Comparación

La B reduce trabajo a cambio de un permiso desproporcionado. La C aporta otras
capacidades, pero adelanta una decisión de proveedor. La A conserva la decisión
de iCloud y usa la capacidad de acceso persistente y acotado que documenta
macOS.

## Recomendación

**Recomendación:** A, como solución mínima suficiente para automatizar la
publicación sin ampliar el radio de acceso del proceso de fondo.

## Decisión del usuario

**Aceptada el 2026-09-01:** mantener iCloud y publicar mediante un helper
sandboxed con security-scoped bookmarks para el staging y la carpeta de backups.

## Consecuencias

- La primera instalación requiere que el operador seleccione las dos carpetas.
- Los bookmarks viven solo en el contenedor local del helper, fuera de Git.
- La réplica activa permanece intacta ante error de copia, bookmark inválido o
  falta de conectividad de iCloud.

## Validación

1. El helper no puede publicar antes de la selección explícita.
2. Un LaunchAgent publica una incremental sin Full Disk Access.
3. La réplica contiene `backup.info` y `archive.info` tras el intercambio.
4. Una restauración aislada desde la réplica publicada sigue dando
   `fasttourney_prod|f`.

## Disparadores de revisión

- El bookmark se vuelve inestable, iCloud no ofrece un RPO útil o la copia
  excede su ventana.
- Se necesita independencia de Mac, cuenta o proveedor.
- La aplicación auxiliar requiere firma distribuible, varios operadores o una
  superficie de permisos mayor.

## Documentación afectada

- [Decisiones](../governance/DECISIONS.md)
- [Decisiones a revisar](../governance/DECISIONS_TO_REVISIT.md)
- [Runbook PostgreSQL K3s](../runbooks/k3s-postgresql.md)
- [Despliegue](../operations/DEPLOYMENT.md)
- [Aprendizaje](../project/LEARNING.md)
- [Changelog](../../CHANGELOG.md)
