# ADR-0146: Ejecutar desarrollo y observabilidad solo bajo petición

- **Estado:** Aceptado
- **Fecha:** 2026-10-04
- **Decisor:** Usuario, mediante instrucción explícita en esta sesión
- **Supera parcialmente:** ADR-0076 y ADR-0100 en activación automática del stack;
  complementa ADR-0139 sin cambiar su retención ni evidencia legal

## Problema y contexto

El usuario quiere reducir escrituras al SSD del Mac. Ocho contenedores dev
permanecen activos fuera de las pruebas. Retención y rotación limitan espacio,
pero no evitan escritura continua. El Mac tiene 16 GB RAM y swap ocupado;
una muestra breve no acredita desgaste físico ni una tasa sostenida de escritura.
Producción funciona en K3s y conserva sus propios controles y disponibilidad.

## Criterios y alternativas

- Perfiles Compose y comandos explícitos: coste bajo, reutiliza el stack y sus
  volúmenes; requiere activar diagnóstico cuando se necesite.
- Mantener todo encendido reduciendo retención: coste bajo, pero continúa
  escribiendo y no cumple la decisión.
- Cambiar runtime o mover datos a RAM: mantenimiento mayor y riesgos de
  recuperación sin evidencia que justifique ese cambio.

## Recomendación y decisión del usuario

Recomendación: perfiles, apagado explícito y exportador desactivado sin receptor.
El usuario acepta: «la observabilidad solo se activará a petición», y mientras
no se use dev para pruebas «la dejamos apagada». Aplica a los carriles local y
público de desarrollo, sin apagar ni modificar producción.

## Implementación y consecuencias

- La API y dependencias arrancan solo para una sesión de pruebas. Los seis
  servicios técnicos tienen perfil `observability`, sin reinicio automático.
- Los servicios de desarrollo no se recuperan automáticamente al reiniciar Docker.
- El arranque, despliegue y rollback ordinarios no habilitan observabilidad.
  La API usa un endpoint de trazas vacío por defecto; el comando de diagnóstico
  habilita Tempo interno y conserva la imagen activa al recrear la API pública.
- Al apagar dev se detienen todos sus servicios, sin eliminar volúmenes ni
  backups. Los LaunchAgents de dev se suspenden, incluido el renderer; se
  reactivan desde sus plist existentes al abrir una sesión pública de pruebas.
- La retención de ADR-0139 y durabilidad PostgreSQL se conservan. Los logs seguros
  de consola y métricas internas de la API siguen disponibles durante pruebas;
  sin el stack técnico no hay histórico central ni alertas dev.
- Las tareas de backup/purga quedan suspendidas mientras dev está apagado; al
  volver a usarlo deben comprobarse sus ejecuciones y el calendario existente.

## Validación y disparadores de revisión

Probar selección de servicios, endpoint vacío/Tempo, políticas de reinicio,
conservación de imagen, comandos de apagado y alcance de LaunchAgents sin tocar
prod. Confirmar en Docker que ambos proyectos dev están detenidos y que sus
volúmenes permanecen. No arrancar dev solo para demostrar el apagado.
Revisar si dev debe atender usuarios continuamente o se requiere diagnóstico
permanente; medir escrituras reales antes de cambiar almacenamiento o memoria.

## Documentación y retrospectiva

Actualizar comandos Make, guías local/dev, observabilidad y aprendizaje.
La solución mínima reduce actividad continua sin introducir otro runtime ni
perder datos. Un perfil controla selección de servicios; no sustituye la política
de reinicio del contenedor ni el control de tareas programadas.

Evidencia de cierre: once tests de seguridad operacional pasan, Compose resuelve
ambos modos sin imprimir secretos y los scripts pasan comprobación sintáctica.
Docker confirma cero contenedores activos en local/dev; PostgreSQL y los
volúmenes técnicos siguen presentes. Los seis nombres de LaunchAgent dev
conocidos no están cargados. K3s y los pods persistentes de producción siguen
Running/Ready. No se levantó dev para probar su apagado ni se midió desgaste físico.
