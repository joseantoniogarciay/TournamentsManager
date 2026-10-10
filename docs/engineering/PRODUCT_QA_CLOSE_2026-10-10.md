# Cierre acotado del QA de producto — 2026-10-10

> Estado: fase cerrada parcialmente, con límites explícitos.
> Decisión del usuario: aceptar el 2026-10-10 un cierre parcial por riesgo.
> Fuente revisada: develop 0481a91; última muestra Android sobre fbc06c1.
> No quedan muestras pendientes de intento en esta tanda; TalkBack conserva validación parcial.
> No autoriza publicación ni distribución.

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
| `keyboard-aware-scroll-view.tsx` | Mide campo y teclado, considera ventana Android reducida, insets iOS y limpia listeners | Muestra Android completada abajo; no generalizar a todos los formularios |
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
| Formulario Android con teclado/texto ampliado | Pixel API 34, bundle actual, texto al 200 %: edición, scroll y foco del segundo campo con teclado visible; botón completo sobre el teclado | Muestra completada sin enviar; no certifica todos los formularios |
| TalkBack representativo | Servicio y exploración táctil activos; foco inicial visible y nombres/roles Android inspeccionados en un detalle de torneo | Ejecutada parcialmente; avance, activación mediante lector y locución sin acreditar |
| Zoom web autenticado | Safari, sesión ficticia: clasificación al 200 % y 300 %, reglas desplazables y cierre comprobados | Muestra completada; no certifica todas las rutas ni 400 % |

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
- TalkBack: recorrido, activación y locución no acreditados en esta muestra.
- Metro registra una advertencia de actualización de estado antes de montar
  un componente; sin traza localizada ni fallo funcional asociado acreditado.
  Se conserva para diagnóstico, sin atribuir causa ni declarar corregida.
- Builds/dispositivos físicos, otros SO, orientación completa, OAuth externo,
  entrega real de correo y validación distribuible de telemetría: fases
  específicas; este cierre no las certifica ni autoriza publicar.

## Retrospectiva

El informe cronológico acumuló límites sin retirar pendientes superados y
el QA manual terminó repitiendo patrones. Una lista vigente finita, evidencia
por capa y un presupuesto de tres muestras permiten cerrar con honestidad.
Los defectos abiertos se conservan; las variantes no probadas dejan de crecer
automáticamente como trabajo de esta fase.

## Ampliación acotada: muestra web completada

El usuario solicita continuar con lo restante. Sobre develop 56183e9 se
arrancan Metro y dev sin observabilidad. Android Studio vuelve a presentar
timeoutReached al pedir captura; se solicita recuperación manual. No se
activan TalkBack ni ajustes de texto Android, y no se insiste sobre el IDE.
Los únicos pendientes de ejecución de las tres muestras son teclado Android
y TalkBack; ambos dependen de recuperar el instrumento.

En Safari se inicia sesión normal con qa_visual_player, cuenta ficticia
existente. Se descarta guardar contraseña. La asignación AX inicial no
actualiza ambos campos React; la entrada nativa permite completar el login.
No se interpreta ese intento instrumental como defecto del cliente.

Se abre desde la biblioteca la liga de equipos con acentos y nombres largos.
El menú de página permite seleccionar y verificar 200 % y 300 % reales.
La clasificación conserva tres filas, posiciones 1/1/3, puntos 3/3/0 y última
fila 2/0/0/2/0/6/-6. Nombres visualmente truncados; AX conserva los completos.
El título y los controles quedan separados. El zoom del navegador amplía
también las celdas: allowFontScaling=false no desactiva zoom web.

Al 300 % se abre Reglas de clasificación, se desplaza dentro del texto hasta
el párrafo final de retirada y se cierra con Ok, completo y visible. El primer
scroll usó coordenadas de una captura anterior y no acredita alcance; el
segundo, sobre la captura actual, muestra el último párrafo. Se restaura 100 %,
se confirma en el menú y se cierra únicamente la pestaña QA, conservando Google.
Capturas y AX en la conversación. Esta es una muestra de clasificación y su
diálogo en sesión autenticada, no una auditoría de todas las rutas ni 400 %.

Metro y dev se apagan conservando volúmenes. No se cambian datos ni permisos.
Solo se actualiza documentación; el salto de línea de expo-env.d.ts generado
por Metro se restaura. Retrospectiva: una sesión ausente se puede preparar con
el login autorizado; no debe confundirse con un bloqueo técnico. Verificar
porcentaje explícito y renovar coordenadas evita acreditar gestos sin efecto.

## Recuperación Android autorizada: límite confirmado

Sobre develop 4c9bfe4, el usuario autoriza cerrar la sesión del IDE, incluso
forzar si no responde, con el riesgo de cambios sin guardar informado. La
segunda salida normal muestra Confirm Exit por tareas de fondo y se completa.
Se reabre Android Studio y se cierra el proyecto restaurado desde Close Project.
La pantalla de bienvenida confirma apps/client/android: el Help que mostraba
node_modules/.pnpm no acreditaba otro proyecto. Se corrige esa hipótesis.

Virtual Device Manager permite arrancar Pixel API 34 sin abrir el proyecto;
ADB lo detecta conectado, pero Running Devices no ofrece la pantalla. Se
reabre el proyecto verificado y una parada/arranque desde Device Manager
recupera imagen de Home y respuesta del botón Home. Esa recuperación parcial
no demuestra que el instrumento sea estable ni que la app actual esté cargada.

Al intentar abrir Fast Tourney aparece «UI del sistema no responde». Se pulsa
Esperar; desaparece el aviso, pero los intentos siguientes muestran el menú
contextual del icono y no una apertura fiable. Android Studio vuelve a indicar
Low memory; una captura anterior tardó unos 62 segundos. No se acredita bundle
Android actual, teclado ampliado ni TalkBack. No se cambian esos ajustes ni
se amplía la matriz. Son bloqueos del entorno, no defectos del producto demostrados.

Se detienen Metro/dev sin observabilidad y se retiran los túneles 8082/8080
añadidos durante la tanda. Se para Pixel desde Device Manager y ADB confirma
que no quedan dispositivos, preservando sus datos. Solo cambia documentación
y aprendizaje. Retrospectiva:
recuperar Home no basta para declarar recuperado el instrumento; comprobar
la apertura de la app antes de cambiar accesibilidad y limitar los reintentos.

## Última muestra Android: teclado completado y TalkBack parcial

Sobre develop fbc06c1, el usuario autoriza forzar el cierre del IDE, usar el
emulador separado temporalmente y controlar exclusivamente Pixel QA con ADB.
Android Studio consume aproximadamente 814–822 % de CPU con el emulador apagado
y muestra memoria baja/error interno. Cerrarlo y mantener el proyecto cerrado
permite arrancar Pixel API 34 desde Device Manager. La preferencia Launch in
Running Devices tool window se desactiva solo para esta tanda.

Metro confirma bundle Android actual (1707 módulos); la app abre la sesión
ficticia existente. Aparece la advertencia de React recogida arriba. Esta
apertura correcta no prueba corregida la caída histórica de DevLauncher por
otra ruta. Dev se inicia sin observabilidad y se añaden túneles 8082/8080.

Con font_scale=2.0, Crear torneo muestra opciones de deporte y campos. Se
introduce únicamente el borrador QA_teclado_sin_crear. El primer campo y su
etiqueta quedan visibles con teclado abierto. Al desplazar el formulario,
Nombre del equipo y Crear torneo quedan completos sobre el teclado; se enfoca
el segundo campo y el teclado sigue visible. No se envía el formulario ni se
crea un torneo. Se oculta el teclado y se restaura font_scale=1.0; la recreación
de la actividad vuelve a Home. No hay captura independiente del teclado oculto
al 200 %, ni evidencia de todas las rutas o de una build distribuible.

Se activa TalkBack y dumpsys confirma servicio enlazado y exploración táctil.
Se rechaza la petición de notificaciones de Suite de Accesibilidad Android.
Durante esa navegación se abre el detalle ficticio QA formulario football;
no se edita ni finaliza. La captura muestra foco verde inicial en Volver.
UI Automator confirma controles Android Button, enabled, clickable y focusable
con nombres Volver, Acciones del torneo, Final, Editar resultado y Finalizar
torneo, además de la imagen Ganador. El XML no acredita locución ni orden de
lectura. Los intentos de gesto y atajo no muestran avance verificable de foco:
no se contabilizan como recorrido ni activación mediante lector. Se detienen
los reintentos dentro de la muestra acotada, conservando ese límite.

## Verificación y cierre operativo

Se restaura y verifica font_scale=1.0, accessibility_enabled=0 y ausencia de
servicios accesibles habilitados, sus valores previos. Se retira el XML temporal
del dispositivo, ambos túneles y se apaga Pixel; ADB confirma lista vacía.
La preferencia Launch in Running Devices tool window vuelve a estar marcada
y se guarda; el IDE queda en bienvenida, sin proyecto abierto. Metro y dev se
apagan conservando volúmenes; no queda escucha en 8082 ni contenedores activos.
No se activa observabilidad ni se actúa sobre producción.

Solo se modifica documentación y aprendizaje; se restaura el salto final de
expo-env.d.ts generado por Metro. No se repiten suites de código sin cambios.
La verificación documental usa git diff --check y comprueba enlaces locales.
La fase sigue cerrada parcialmente: dos muestras completadas y TalkBack
parcial. Los límites registrados no amplían automáticamente esta tanda.

Retrospectiva: el emulador separado permitió aislar la saturación del IDE y
completar la muestra útil de teclado. Servicio activo, atributos accesibles y
foco inicial son evidencias distintas de recorrido y audio; mantener esa
separación evita tanto una aprobación falsa como otra matriz interminable.
