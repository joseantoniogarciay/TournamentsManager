# Cierre acotado del QA de producto — 2026-10-10

> Estado: fase cerrada parcialmente, con límites explícitos.
> Decisión del usuario: aceptar el 2026-10-10 un cierre parcial por riesgo.
> Fuente revisada: develop 0481a91. No autoriza publicación ni distribución.

## Alcance y criterio de salida

Este documento es la lista vigente de cierre. El [informe histórico](PRODUCT_QA_REVIEW_2026-10-04.md)
conserva la evidencia y sus límites; sus menciones cronológicas a «pendiente»
no constituyen tareas adicionales. No se convierte falta de cobertura total
en un defecto demostrado ni revisión de código en ejecución manual.

Se aplica [ADR-0019](../adr/0019-use-risk-based-layered-testing.md): pruebas
por riesgo y capas. Frente a multiplicar deporte × formato × plataforma ×
idioma × lector, se revisa el patrón compartido y se reutilizan muestras ya
ejecutadas. Reduce coste y repetición; conserva riesgo residual de plataforma.
Solo un defecto reproducible o un cambio del patrón justifica ampliar la tanda.

## Consolidación de las 17 áreas funcionales

| Área | Evidencia suficiente para esta fase | Límite conservado |
| --- | --- | --- |
| Registro/verificación | Automatización, API y UI en los tres targets | No repetir idiomas y lectores |
| Login pendiente/reenvío | API y UI web/iOS/Android | No repetir cada combinación |
| Login/logout/refresh | Suites de sesión y recorridos reales, incluidos fallos | No audita todos los tiempos posibles |
| Recuperación/reautenticación | Contrato, persistencia, solicitudes/enlaces y formularios | Cambio de credencial por intervención humana |
| Google/Apple | Código y pruebas de integración del adaptador | OAuth real diferido expresamente |
| Biblioteca/seguimiento | Paginación, vacíos y seguimiento en tres targets | No repetir cada card con lector |
| Administradores | Permisos, menú, acceso directo rechazado y búsqueda | No otorgar más permisos para ampliar QA |
| Invitación | Borrador, entrada nativa, inscripción y conflictos ya registrados | No repetir cada enlace/estado con lector |
| Composición de equipos | Altas, duplicados, bajas y nombres largos existentes | No multiplicar muestras por deporte |
| Deportes/formatos | Contrato, dominio, API y variantes UI ya ejecutadas | Combinaciones nativas restantes fuera de esta fase |
| Liga/mixto | Ciclos, corte, grupos y transición existentes | No repetir por idioma o lector |
| Eliminatoria | Byes, finales, corrección y campeón existentes | No recorrer todos los cuadros nativos |
| Resultados/incidencias | Pruebas de parsers/formulario y fixtures reales existentes | No completar producto cartesiano de incidencias |
| Retirada/cancelación/finalización | Persistencia, permisos y recorridos existentes | No añadir acciones destructivas de cierre |
| Notificaciones/sugerencias | Lista final, destinos, validación/envío y tasa existentes | Borrado efectivo fuera del cierre; scroll iOS ya acreditado |
| Cuenta/transferencia/baja/purga | Backend/API, formularios y cancelaciones existentes | Transferencia efectiva, baja/purga y credenciales en fase específica |
| Transporte/fallback | Suites, corte real y fallos inyectados en tres targets | No acredita timeout propio ni todas las redes |

No se asigna porcentaje: estas áreas contienen riesgos distintos y evidencia
de fechas diferentes. Los contadores del histórico son tests/subtests, no una
medida de cobertura exhaustiva.

## Revisión de código por patrones

| Patrón y archivos revisados | Hecho observado | Conclusión acotada |
| --- | --- | --- |
| `shared/ui/button.tsx`, tokens | Role button, busy/disabled, loading bloquea pulsación, mínimo 44, texto centrado | Reutilizar las muestras de botones; no garantiza locución |
| `shared/ui/text-field.tsx` | Etiqueta desde label/placeholder, error incorporado a AX, interacción por campo, blur/change explícitos, control de contraseña 44 | No recorrer cada campo con VoiceOver |
| `shared/ui/confirmation-dialog.tsx`, `screen.tsx` | Host por pantalla enfocada, modal, cierre accesible y back nativo, botones comunes | Reutilizar diálogos probados; foco/restauración no se certifican solo por código |
| `keyboard-aware-scroll-view.tsx` | Mide campo y teclado, considera ventana Android reducida, insets iOS y limpia listeners | Necesita una muestra Android; no prometer botón visible con teclado abierto |
| `screen.tsx`, `card.tsx`, `use-tab-content-bottom-padding.ts` | Margen de Card 20, reserva inferior en contenido, top inset separado | Reutilizar muestras de cada layout |
| `text.tsx`, `standings.tsx` | Escalado nativo por defecto; false solo en celdas de clasificación; columnas fijas y estadísticas desplazables | Excepción de ADR-0153 conservada; no nueva alternativa de tabla |
| Cards de torneo, biblioteca, notificaciones y búsquedas de administrador/transferencia | Roles, nombres completos/localizados o texto hijo, controles de estado y bloqueo existentes | Revisión de conexiones representativas, no auditoría exhaustiva |
| `api/fetch.ts`, `shared/feedback/request-failure.ts` y suites de sesión/reautenticación/incidencias | Cancelación de transporte separada; conexión y fallback seguro; pruebas de duplicados, respuestas y parsers | Conservar evidencia automatizada/API; no repetir cada estado por pantalla |

La búsqueda de atributos AX literales no encontró labels/hints escritos como
strings JSX en `src`. Es una comprobación estática acotada, no prueba de todos
los textos posibles ni de traducción correcta. No se modificó el cliente.

## Tres muestras manuales finales

| Muestra | Resultado | Cierre |
| --- | --- | --- |
| Formulario Android con teclado/texto ampliado | Android Studio seleccionable, pero AX/captura agotan tiempo de espera | Bloqueada por instrumento; se acepta el límite y no se insiste |
| TalkBack representativo | Mismo bloqueo de Android; no se activa ni se acredita | Fuera de ejecución de esta fase por instrumento |
| Zoom web autenticado | Safari carga la versión actual en Inicio anónimo; no hay sesión QA conservada en ese navegador | No ejecutada en esta tanda; se conserva la evidencia web previa y el límite de zoom autenticado |

VoiceOver iOS reutiliza muestras existentes de formulario, listas y controles.
La tabla tiene evidencia parcial de foco; locución y asociación contextual no
se certifican. No se exige revisar cada componente o escuchar toda la app.

## Defectos y límites que sobreviven al cierre

- Caída de DevLauncher Android al abrir enlace: evidencia histórica abierta;
  arranques release correctos no prueban su corrección.
- Android oscurece diálogos sin el mismo blur iOS: diferencia conocida, sin
  ensayo adicional ni cambio de arquitectura autorizado en este cierre.
- Gate de lint Android y deuda del analizador: seguimiento en ADR-0152;
  separado del QA funcional, sin presentarlo como aprobado.
- Tamaño fijo de clasificación: limitación visual aceptada en ADR-0153.
- Builds/dispositivos físicos, otros SO, orientación completa, OAuth externo,
  entrega real de correo y validación distribuible de telemetría: fases
  específicas; este cierre no las certifica ni autoriza publicar.

## Retrospectiva

El informe cronológico acumuló límites sin retirar pendientes superados y
el QA manual terminó repitiendo patrones. Una lista vigente finita, evidencia
por capa y un presupuesto de tres muestras permiten cerrar con honestidad.
Los defectos abiertos se conservan; las variantes no probadas dejan de crecer
automáticamente como trabajo de esta fase.

## Verificación y cierre operativo

No se modificó código ni se repitieron suites que ya pasaron sin cambios
posteriores relevantes. Se revisaron fuentes y pruebas existentes; los
resultados de ejecución conservan sus fechas en el histórico. La apertura
web confirma carga, no zoom ni autenticación. El primer arranque de Metro
sin permisos de red no llegó a servir; se detuvo y el segundo cargó web.
Al terminar se detienen Metro y dev, se conserva la pestaña Google previa
y se cierra la pestaña QA. No se activa observabilidad ni se modifica Android,
que continúa con el límite instrumental registrado. No hay acciones sobre
datos, credenciales o permisos. La validación documental comprueba enlaces
locales nuevos y git diff --check.
