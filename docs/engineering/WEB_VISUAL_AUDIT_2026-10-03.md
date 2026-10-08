# Revisión visual web — 2026-10-03

## Resultado

La web no puede darse por validada sin reservas: se han confirmado siete grupos
de problemas. La auditoría inicial documentó siete grupos. Tras la autorización del usuario,
los problemas 1–6 se corrigieron; el 7 se retiró por decisión explícita de
idioma automático (ADR-0143). Las evidencias iniciales se conservan abajo.

## Método y límites

- Navegador integrado, interacción real, capturas y lectura de dimensiones DOM.
- Viewports muestreados: 320×480, 390×844, 768×1024, 1024×768, 1280×720 y
  1440×900. No se ha ejecutado el producto cartesiano de todas las pantallas,
  tamaños, temas, deportes y estados.
- Tema inicial Sistema/claro; muestreo adicional en oscuro. Se restauró Sistema.
- Primera pasada sin sesión y con API inaccesible; segunda pasada con API
  ficticia en localhost:8099 y sesión de organizador. El usuario autorizó los
  mocks. No se crearon cuentas ni se alteraron torneos reales.
- Datos ficticios: nombres largos, ocho equipos, liga de fútbol y voleibol,
  cuadro de tenis, torneo de bádminton publicado, administradores y notificación.
  Son proyecciones para revisar presentación, no evidencia de reglas de negocio.
- Se revisaron ADR-0054/0055/0056/0057, el manifiesto y las reglas del cliente.
  El gate técnico está cerrado. No se introducen dependencias ni un rediseño.
- Idioma observado: español. No fue posible recorrer idiomas desde la UI porque
  falta el selector web acordado. Safari, Firefox, dispositivos nativos, zoom,
  lectores de pantalla reales y contraste cuantitativo quedan fuera de esta pasada.

## Hallazgos confirmados

### 1. Recuperación: el título invade Volver

**Prioridad alta.** En `/account/forgot-password` a 320×480, el texto
«¿Has olvidado tu contraseña?» se pinta sobre el objetivo circular de volver.
Ocurre tanto en claro como en oscuro. La cabecera de Cuenta usa el título
predeterminado sin una restricción que preserve el espacio de los controles.

Fuente: [`_layout.web.tsx`](../../apps/client/src/app/(tabs)/account/_layout.web.tsx).
Evidencia: [cabecera superpuesta](visual-audit-2026-10-03/forgot-password-header.jpg).

**Recomendación:** limitar el título al espacio disponible y permitir el ajuste
necesario, conservando el objetivo de 44 px y los 20 px de separación acordados.
Acortar solo el texto español sería más barato inicialmente, pero deja el
problema abierto para traducciones y otros títulos.

### 2. Cuenta: un username válido queda debajo de los controles

**Prioridad alta.** A 390×844, `organizador_de_prueba_largo` invade el espacio de
notificaciones y ajustes. Su longitud está dentro del contrato. El `headerLeft`
solo añade margen; no reserva espacio para las acciones derechas.

Fuente: [`_layout.web.tsx`](../../apps/client/src/app/(tabs)/account/_layout.web.tsx).
Evidencia: [usuario bajo los botones](visual-audit-2026-10-03/account-long-username.jpg).

**Recomendación:** reservar el ancho de los controles y ajustar/truncar el nombre
en el ancho restante, conservando su contenido accesible completo.

### 3. Transferir torneo carece de cabecera y cierre

**Prioridad alta.** Tanto al cargar la URL como al entrar por Acciones del torneo
→ Transferir torneo, se ven ayuda y búsqueda, pero ningún título ni control de
cierre. La ruta configura sus opciones locales, pero hereda `headerShown: false`
del stack raíz y no está registrada junto a las demás rutas de gestión.

Fuentes: [`_layout.tsx`](../../apps/client/src/app/_layout.tsx) y
[`transfer.tsx`](../../apps/client/src/app/tournament/[id]/transfer.tsx).
Evidencia: [transferencia sin cabecera](visual-audit-2026-10-03/transfer-missing-header.jpg).

**Recomendación:** declarar la presentación de esta ruta en el mismo patrón que
las rutas hermanas. No crear un control de cierre local diferente.

### 4. Enlace de recuperación inválido sin margen ni siguiente acción

**Prioridad media.** `/link/password-reset`, sin token, presenta únicamente un
texto a x=0 y no ofrece un control de recuperación. Se confirmó antes de usar
un token ficticio válido. El formulario válido sí usa Card y mantiene márgenes.

Fuente: [`password-reset.tsx`](../../apps/client/src/app/link/password-reset.tsx).
Evidencia: [estado inválido](visual-audit-2026-10-03/password-reset-invalid.jpg).

**Recomendación:** dar al estado terminal una Card o layout compartido con 20 px
exteriores y una salida útil. La elección concreta del destino requiere acordar
el comportamiento; la falta de margen ya contradice la regla vigente.

### 5. Objetivos pulsables inferiores a los 44 px acordados

**Prioridad media.** Medidas DOM comprobadas:

| Control | Medida | Contexto |
| --- | --- | --- |
| ¿Has olvidado tu contraseña? | alto 21 px | Cuenta sin sesión, 320×480 |
| Permitir analítica de producto | 40×20 px | Inicio y Ajustes web |

El enlace de recuperación sí navega al pulsarlo, pero su área es demasiado baja.
El Switch no tiene un objetivo ampliado en su composición compartida. No se
activó analítica para probarlo.

Fuentes: [`account/index.tsx`](../../apps/client/src/app/(tabs)/account/index.tsx) y
[`product-analytics-preference-card.tsx`](../../apps/client/src/shared/preferences/product-analytics-preference-card.tsx).

**Recomendación:** ampliar los objetivos a 44 px sin agrandar necesariamente su
representación visual; hacerlo en la pieza compartida evita divergencias.

### 6. La selección de tema no se expone como radio seleccionado

**Prioridad media, accesibilidad.** El tema cambia visualmente y persiste al
navegar, pero los tres nodos `role="radio"` carecen de `aria-checked` en el DOM.
Se comprobó después de seleccionar Oscuro. El código declara
`accessibilityState={{ checked: selected }}`; eso no acredita el resultado web.

Fuente: [`settings.tsx`](../../apps/client/src/app/(account-modals)/account/settings.tsx).

**Recomendación:** revisar la adaptación web de esta semántica y comprobar que
exactamente una opción exponga su estado. No se ha ejecutado una prueba con
lector de pantalla real.

### 7. Falta el selector persistente de idioma web

**Prioridad media, cumplimiento del diseño aceptado.** Ajustes solo contiene
apariencia y analítica en web. No hay selector en la navegación ni un estado
compartido de locale elegido; `getCurrentLanguage()` resuelve el idioma del
navegador. ADR-0056 y AGENTS del cliente exigen elección persistente en web.

Fuentes: [`settings.tsx`](../../apps/client/src/app/(account-modals)/account/settings.tsx)
y [`locale.ts`](../../apps/client/src/shared/i18n/locale.ts).
Evidencia: [Ajustes](visual-audit-2026-10-03/settings-no-language-selector.jpg).

**Recomendación:** completar la decisión ya aceptada mediante estado compartido,
sin introducir una selección local en cada pantalla.

## Recorridos realizados

«Sin defecto observado» se limita a los estados y tamaños muestreados; no equivale
a validar cada combinación posible.

| Área | Evidencia de revisión |
| --- | --- |
| Inicio | Sin sesión y con actividad ficticia; márgenes, cards y botón Crear. El Switch tiene el problema 5. |
| Torneos | Sin sesión, Administro con cinco torneos y Sigo vacío; cambio de tab operable y botón flotante visible. |
| Crear torneo | Escritorio y 320×480; opciones con wrap, controles de 44 px, scroll hasta envío y errores inline al intentar continuar vacío. |
| Cuenta | Acceso sin sesión, registro desplazado hasta términos y botón final, acceso autenticado con username largo. Problemas 1, 2 y 5. |
| Datos de acceso | Correo ficticio, filas de métodos y acción de eliminación; se midió esta última a 44 px. Sin envíos. |
| Contraseña | Formulario autenticado con contraseña actual/nueva y botón Guardar; estado de enlace válido con correo ficticio y visibilidad de contraseña. Sin cambios de credenciales. |
| Google | Diálogo de vinculación con confirmación de identidad; cierre por fondo. Sin OAuth externo. |
| Ajustes | Claro y oscuro, persistencia al navegar y restauración del tema. Problemas 5, 6 y 7. |
| Notificaciones | Error de conexión con reintento y una notificación ficticia con nombre largo, fecha y acciones. |
| Legales | Términos y privacidad; entrada por enlace del registro, ajuste de texto y contenido desplazable. |
| Confirmación de email | Enlace sin token: Card terminal centrada con vuelta a Inicio. No se validó una verificación efectiva. |
| Invitación | Enlace ausente y token ficticio en fragmento: título largo, ayuda, campo y acción final accesible por scroll. Sin inscripción. |
| Torneo de liga | Resumen, nombres largos en partidos, Equipos/Clasificación, menú de acciones; lectura en móvil, tablet y escritorio. |
| Resultados | Fútbol: dos columnas y Guardar. Incidencia: opciones largas y scroll hasta Guardar. Tenis: cabeceras, tres sets, columnas flexibles y Guardar dentro del scroll a 320×480. |
| Clasificación | Fútbol y voleibol; columnas fijas, estadísticas con scroll horizontal sincronizado y diálogo de reglas. |
| Cuadro | Tenis con ocho participantes; modo compacto y columnas desktop, barra horizontal a 1024 px, navegación hacia Final y resaltado del destino. |
| Equipos y participantes | Lista en curso y torneo publicado, nombres largos, controles de retirada/eliminación y diálogo Añadir participante. Sin retirar ni guardar. |
| Administradores | Lista con username largo, cierre, acciones y formulario de búsqueda; resultados de búsqueda en transferencia. |
| Transferencia | Búsqueda con resultados; problema 3 confirmado desde menú y URL directa. Sin transferencia. |

No se han certificado: todas las variantes deportivas, desempates de formato
mixto, cuadros de 64 participantes, celebración de campeón, cada permiso/estado
de error, cada traducción ni navegación mediante lector de pantalla. Los mocks
no prueban funcionamiento ni seguridad de la API.

## Verificación y retrospectiva

La revisión visual detecta fallos que typecheck y exportación no detectan: un
texto puede existir en accesibilidad y pintarse sobre otro control, y una
declaración React de accesibilidad puede no producir el atributo DOM esperado.
Conviene conservar casos de 320×480, nombres largos dentro del contrato y
estados terminales en las revisiones futuras.

La alternativa mínima recomendada es corregir estos componentes y rutas sobre
los tokens existentes. Un rediseño general o una nueva librería elevarían el
mantenimiento sin resolver mejor estos defectos concretos. No se ejecutó esa
implementación dentro de la auditoría.

Validación de compilación: `pnpm run typecheck` y `make client-web-export`
pasaron; Expo exportó 35 rutas. No son pruebas de geometría ni sustituyen las
evidencias anteriores. Se detuvieron Metro y la API ficticia al terminar.


## Correcciones autorizadas y cierre

El usuario autorizó corregir los hallazgos 1–6 y mantener idioma automático.

- Cabecera de Cuenta web: título limitado por el viewport y los controles,
  dos líneas, 20 px libres; username de una línea con texto accesible completo.
- Transferencia registrada en el stack raíz como las rutas hermanas, con
  cabecera y cierre compartido.
- Recuperación inválida en Card desplazable, con acción localizada existente
  de vuelta a Inicio; no cambia las respuestas HTTP ni su interpretación.
- Enlace de recuperación con alto mínimo de 44 px. Analítica con objetivo
  compartido de 44×44, una única semántica switch y representación interna
  oculta a accesibilidad y sin foco.
- Radios de tema exponen aria-checked además del estado accesible nativo.
- Idioma automático documentado en ADR-0143; el hallazgo 7 ya no es un defecto.

Verificación manual web: a 320×480 el título ocupa x=84..236 y el botón termina
en x=64: conserva 20 px. Username largo termina en x=160 y las acciones quedan
fuera de ese espacio. Transferencia muestra título/Cerrar y el cierre vuelve al
torneo. El enlace inválido muestra Card a 20 px y su acción vuelve a Inicio.
Sistema/Oscuro exponen exactamente un radio seleccionado y se restauró Sistema.
El switch accesible mide 44×44 y no duplica controles en el árbol accesible.
La cabecera también se revisó a 1280×720.
La preferencia de analítica no se activó; no se alteraron datos de negocio.
No se certifican dispositivos nativos ni lectores de pantalla reales.

Evidencia posterior:
[cabecera corregida](visual-audit-2026-10-03/forgot-password-header-fixed.jpg).

Retrospectiva: reservar espacio antes de truncar mantiene la salida disponible;
el estado accesible debe comprobarse en DOM, no solo en props de React.


Checklist de cierre del cliente: tokens y primitivas existentes; textos
localizados reutilizados; margen Card de 20 px; objetivos de 44 px; cierre
circular compartido; estado accesible sin duplicación. No se alteraron padding
de tabs, transporte HTTP, adaptadores OpenAPI, permisos ni reglas de negocio.
No hay operación OpenAPI tocada que requiera revisar nuevos estados.
El enlace de recuperación sin sesión se midió en 184,38×44 px y navegó a
recuperación; Volver devolvió a Cuenta. Typecheck, ESLint de los archivos
modificados y exportación web (35 rutas) pasaron. La validación nativa sigue
pendiente para los cambios compartidos; no se presenta como realizada.
