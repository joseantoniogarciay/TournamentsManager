# ADR-0140: Acotar la retención técnica de producción

- **Estado:** Aceptado
- **Fecha:** 2026-10-03
- **Decisor:** Usuario: «vamos con tu recomendación» tras comparar plazos y espacio
- **Supera parcialmente a:** ADR-0106, solo plazos de diagnóstico y seguridad de producción

## Problema y contexto

La VM conserva logs en Loki sin purga, pese a declarar 24 h. El PVC local-path
no impone una cuota; el disco raíz tiene 16 GB libres. Los plazos anteriores
(30 días de diagnóstico y doce meses de seguridad) son decisiones del proyecto,
no una obligación legal acreditada para estos eventos técnicos.

## Alternativas y mantenimiento

1. Un Loki con categorías y retención por stream: menor operación y sin otro
   servicio. Comparte capacidad y permisos de consulta; la clasificación debe
   mantenerse y probarse al añadir operaciones de identidad.
2. Un almacén adicional de seguridad: permite aislar capacidad y accesos, pero
   duplica configuración, recuperación y monitorización sin necesidad actual.
3. Reducir indiscriminadamente todos los logs: sencillo, pero pierde la ventana
   de investigación de seguridad. Descartado.

## Decisión aceptada

- Diagnóstico de producción: siete días. Seguridad mínima: noventa días.
- Mantener un Loki y añadir `retention_category` con valores cerrados
  `diagnostic` y `security`. No son etiquetas de usuario ni incluyen PII.
- Seguridad comprende operaciones de alta, verificación, autenticación,
  sesiones, recuperación de contraseña, credenciales, identidades y RISC;
  también rechazos seguros de autenticación, credenciales, CSRF y límite de tasa
  en cualquier operación. No se infiere una causa de negocio del estado HTTP.
- La aplicación ya emite JSON seguro con plantillas de ruta y causas cerradas.
  Alloy clasifica esos eventos existentes, sin desplegar cambios funcionales de
  la API. Los mensajes RISC explícitos se conservan como seguridad.
- Los streams históricos sin categoría y mensajes de API no reconocidos
  mantienen como máximo noventa días. Es una transición conservadora para evitar
  borrar seguridad mezclada; los nuevos eventos HTTP reconocidos de diagnóstico
  tienen siete días. No se declara reclasificado el histórico.
- Las líneas sintéticas del canary de Loki tienen siete días mediante su
  etiqueta upstream `pod`; no contienen eventos de seguridad.
- Habilitar compactor persistente y purga asíncrona. Descartar solo chequeos
  HTTP correctos de salud y métricas; conservar errores.
- Tempo mantiene siete días. Prometheus conserva un día o 128 MB de TSDB.
  WAL, head y compactación añaden espacio; no es una cuota absoluta.
- Journal técnico del host: siete días, máximo 256 MB y reserva de 1 GB libre.
  Kubelet mantiene sus cinco archivos de 10 MiB por contenedor.
- Medir espacio libre, tamaño de los tres volúmenes técnicos y frescura de la
  medición cada cinco minutos. Exportar únicamente agregados mediante un
  directorio de métricas montado en Alloy como solo lectura. No montar el disco
  raíz ni conceder lectura de Secrets o nuevos permisos Kubernetes.
- Alertar antes de agotar espacio: raíz por debajo del 25 % o 5 GiB (aviso),
  por debajo del 15 % o 3 GiB (crítico); volúmenes técnicos por encima de 512 MiB (aviso)
  o 1 GiB (crítico); alertar también si la medición deja de funcionar o Loki
  deja de completar ciclos de retención. La antigüedad de bloques de Prometheus
  se comprueba continuamente, con margen para compactación y arranque.
- Los umbrales son alertas, no cuotas físicas ni autorización para borrar
  seguridad antes de plazo. Si el crecimiento exige aislamiento físico, se
  revisará el almacenamiento; no se remonta ni migra la raíz durante este cambio.
- Evidencia legal, su bloqueo tras la baja, copias y PostHog no cambian.

## Validación y retrospectiva

Validar los charts fijados, configuración activa y categorías con fixtures sin
PII. Observar compactor, purga de Tempo, retirada de bloques viejos de Prometheus,
series de disco y reglas de alerta. Un rollback de Helm no restaura datos ya
purgados. El aprendizaje es distinguir retención, rotación, umbral y cuota.
