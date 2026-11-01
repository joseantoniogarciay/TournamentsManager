# Identidad Grada — implementación de ADR-0176

Fecha: 2026-11-01. Implementado; builds verificadas y QA iOS parcial, con límites pendientes.

## Alcance implementado

- Tokens semánticos cálidos y adaptación oscura en `packages/design-tokens`.
  `themeColors` alimenta preferencias, HTML inicial, React Navigation y el fondo
  configurado de splash. El subpath `design-tokens/tokens` permite a Expo config
  consumirlos con Node sin los reexports sin extensión usados por Metro.
- DM Sans 400/500/600/700 para lectura, botones, campos, cabeceras y tabs;
  Barlow Condensed 600/700 para variantes title/display de contenido;
  IBM Plex Mono 400/600 para posición, estadísticas y puntos de clasificación.
  Se retira Figtree del grafo. No se añaden módulos nativos ni una biblioteca UI.
- Se conservan componentes, escala 12–32 px, radios, padding, rutas, márgenes,
  reflow compartido, insets y excepciones de escalado de clasificación.
- NativeTabs conserva su host, cuatro secciones y materiales por plataforma;
  el tintado cambia según tema. Cabeceras de navegación mantienen DM Sans.
- Foco, selección de tema y borde del partido resaltado consumen el color
  adaptado. El loader filled usa blanco también en oscuro.
- El degradado existente mantiene su geometría, con naranja quemado y un tono
  más oscuro. No se traslada la composición HTML de la exploración.

## Verificación ejecutada

| Comprobación | Resultado y límite |
| --- | --- |
| Prettier, Oxlint y typecheck | Correctos |
| Suite Node completa | 136 pruebas correctas, incluidas 5 de contraste |
| Contraste AA | Texto normal, placeholders y feedback >=4,5:1 sobre canvas/default/subtle de ambos temas; indicadores, bordes y foco >=3:1 |
| Acciones filled | Texto blanco >=4,5:1 en 101 muestras del degradado |
| Fuentes locales | Ocho TTF, 705.740 bytes sin comprimir; cmap cubre las 78 letras latinas/dígitos extraídos de los cuatro catálogos |
| Exportación web Expo | Correcta, 44 rutas; mantiene carga local de fuentes |
| Exportación iOS/Android Expo | Correcta, bundles Hermes y assets; no equivale a compilar/ejecutar una development build |
| `expo install --check` | Correcto con matriz local; el modo offline advierte que no contrasta el endpoint remoto |
| Checklist de cliente | Revisada contra los archivos modificados; sin nuevos textos, operaciones OpenAPI ni cambios funcionales |

La regresión de bordes detectó 2,98:1 en oscuro sobre `surface.subtle`; el token
se ajustó antes del cierre. Las pruebas que escuchan en localhost necesitaron
salir del sandbox (el primer intento dio EPERM). La pasada completa posterior
terminó sin fallos. En la primera fase no se ejecutó `make verify` completo;
el cierre adicional sí lo ejecuta tras adaptar el correo, con la omisión de
integración indicada.

La revisión de consumidores no encuentra los antiguos colores Pulse ni Figtree
en el código TS del cliente/tokens. Tras la petición «Haz lo que falte» se
adaptan también el icono, splash y preview social con imagegen, conservando el
símbolo y la composición. Se verifica icono opaco 1024×1024, splash con alpha
1254×1254 y preview opaca 1200×630. El email de verificación conserva estructura,
copy, enlaces, MIME y localización y usa la paleta Grada con fallback sólido.
Las fuentes de sistema permanecen en correo por compatibilidad; no carga CDN.
Las licencias OFL de las tres familias se incluyen en `public/font-licenses`.

## Cierre adicional solicitado

- `make verify` completo termina correctamente. Integración PostgreSQL omitida
  por no configurar `TM_INTEGRATION_DATABASE_URL`; no se usa el volumen local
  para pruebas que truncan datos. Incluye test Go del adaptador SMTP, regresiones,
  exportación web, generadores deterministas, compilación backend y govulncheck.
- Android: CNG incremental, assembleDebug y lintDebug correctos con JDK 21;
  cero errores y 46 avisos heredados. APK instalado en Pixel_API_31, Metro y API
  conectados mediante reverse. La compilación tardó 4m15s. La revisión visual no
  queda acreditada: Android Studio agota el tiempo de control de interfaz y la
  ventana Qt del emulador no está disponible como app en CUA.
- iOS: la primera build detecta una ruta pnpm antigua en un recurso de EXConstants;
  `pod install` regenera los enlaces sin editar Pods manualmente. La segunda
  compilación termina correctamente con firma de simulador y se instala en
  iPhone Duo iOS 27.1. Log: `/tmp/tm-grada-ios-build.log`.
- QA iOS mediante Device Hub: Inicio invitado en claro/oscuro, Duo cerrado y
  abierto, cambio a Cuenta, acceso con cuenta local, biblioteca, ficha de liga
  con nombre largo y resultados, clasificación con IBM Plex Mono y diálogo de
  reglas. Se comprueban giros y postura parcialmente abierta conservando el
  torneo y los datos. Dynamic Type de 3 a 11 y vuelta: titulares envuelven,
  clasificación mantiene la excepción compacta y el diálogo conserva Ok visible.
  El gesto de scroll del simulador no se pudo ejercer de forma fiable por CUA;
  no se acredita todavía el último control en todos los recorridos.
- La QA de ficha detecta que el borde secundario quemado es tenue en oscuro.
  Se corrige en `Button`: mantiene la forma y usa `border.focus` claro en oscuro,
  con contraste >=3:1; las acciones filled conservan su degradado y tinta blanca.
  Se cierra la secuencia antes de editar JS y se comprueba la corrección después
  en una nueva apertura de ficha. No se modifica código durante una secuencia.
- Para QA se inicia local sin observabilidad y se crean dos torneos (liga y cuadro)
  con partidos/resultados reales mediante la API local. Evidencia y credenciales
  efímeras quedan en `/tmp/tm-grada-qa-state.json` (0600), sin versionarlas.
- QA de localización iOS: Inicio invitado en es y autenticado en fr/it/en, con
  datos reales y nombres largos. Francés, italiano e inglés usan argumentos temporales
  `-AppleLanguages`/`-AppleLocale`, sin modificar el idioma persistente del SO.
- iPad Pro 13 M5/iPadOS 27: misma build instalada, Inicio invitado claro/oscuro,
  vertical y horizontal, Dynamic Type 3/11/3, CTA y navegación nativa superior.
  El gesto de redimensionado no acredita Split View: no se afirma esa cobertura.
- Android arranca desde Metro y completa el bundle. La consulta de errores
  ReactNativeJS/AndroidRuntime no registra fallos en la muestra; no sustituye
  inspección visual de la jerarquía ni la matriz de Android.
- Tras la corrección final del borde secundario se repiten typecheck, Oxlint,
  contraste y exportación web. La ejecución completa anterior de `make verify`
  se conserva como evidencia, no se presenta como una segunda pasada posterior.

## Límites pendientes para cerrar la validación

1. En web, revisar visualmente Inicio, ficha y clasificación en claro/oscuro y los cuatro
   locales, con nombres largos, texto ampliado, reflow y controles/feedback.
   El intento de abrir la exportación en el navegador de Work respondió
   «Local browser use is disabled in Work settings»; no hay evidencia visual
   web nueva. El servidor de revisión se apagó al terminar.
2. Completar la matriz de fuentes, cabeceras, NativeTabs/materiales,
   insets, rotación, tablet y Split View. Las builds ya están regeneradas,
   compiladas e instaladas; falta completar la matriz visual Android,
   ventanas/Split View y el acceso al último control.
   La QA iOS acreditada se detalla arriba.
3. No quedan assets de la identidad anterior en los recursos adaptados. El
   nombre de producto continúa FastTourney; Grada identifica la dirección visual.
   La apariencia exacta de splash de distribución requiere una build release;
   una development build no acredita por sí sola ese arranque.

En el cierre adicional se arrancaron API, PostgreSQL y Mailpit solo para QA.
Metro, API, PostgreSQL, Mailpit y los simuladores iniciados se apagan al cerrar.
No se activó observabilidad ni se desplegó/publicó. Se conservan volúmenes,
backups y los cambios previos de Cuenta y OAuth Google.

## Retrospectiva técnica

Separar identidad y anatomía limita el coste de mantenimiento: un cambio de
fundaciones alcanza los consumidores existentes sin rehacer pantallas. Oscuro
necesita distinguir fondo de marca, tinta de acción e inverso del tema: reutilizar
un solo naranja o `text.inverse` para todo rompe contraste. Los ocho pesos
aportan personalidad y lectura de cifras a cambio de más bytes de arranque;
no se añaden más familias o pesos sin una necesidad observada.

Las pruebas de contraste permanecen en `make verify` mediante
`test-design-tokens`. No sustituyen la matriz visual pendiente ni evalúan la
percepción semántica entre éxito, error y marca, que también debe revisarse.

## Autorización de publicación

El usuario solicita «Pues despliega y sube release». Se prepara v1.15.0 con
merge explícito a main, tag anotado y GitHub Release según ADR-0119. La release
incluye también el mantenimiento aceptado pendiente desde v1.14.1 y el ajuste de
feedback de Google presente en el workspace. Se ejecuta el gate final con
PostgreSQL 18.6 desechable; los límites visuales anteriores se conservan en las
notas de release, sin atribuir publicación de apps en las tiendas.

Gate de publicación local: `make verify` termina con PostgreSQL 18.6 desechable
(migraciones y pruebas reales, sin omisión). El cliente psql se ejecuta dentro
del contenedor de pruebas mediante un wrapper temporal, porque el host no lo
incluye en PATH. No se instala software ni se accede al volumen de desarrollo.
Antes del despliegue la API devolvía 502 porque la VM estaba detenida; se arranca
y se verifica K3s Ready con ambas réplicas API y PostgreSQL sanos. Backup cifrado
`20261011-080242F_20261101-090314I`, estado 0, restaurado en emptyDir con repo de
solo lectura y set exacto: Job completo y `fasttourney_prod|f`. No se exportan
los datos de producción. La evidencia de activación se cerrará en DEPLOYMENT.md.
