# ADR-0062: Usar access y refresh tokens opacos rotatorios

- **Estado:** Aceptado
- **Fecha:** 2026-07-31
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** las reglas de duración y renovación de ADR-0044
- **Superado por:** Ninguno

## Problema

La aplicación móvil debe recuperar una sesión local tras reiniciarse sin una
consulta de sesión al arrancar, conservar la revocación en el backend y renovar
la credencial de acceso sin carreras entre peticiones concurrentes.

## Contexto y restricciones

- Las credenciales siguen siendo secretos opacos; el cliente no puede deducir
  su validez ni su expiración desde el valor del token.
- El backend conserva la autoridad: persiste únicamente huellas de los secretos
  y devuelve los instantes de expiración con cada emisión o renovación.
- La identidad de interfaz (`user`) y los instantes devueltos se guardan junto
  al secreto móvil en Keychain/Keystore. Son un estado local de arranque, no una
  autorización.
- La web conserva cookies `HttpOnly`; el secreto de refresh nunca se expone a
  JavaScript.

## Alternativas

### A — Access opaco de 7 días y refresh opaco de 30 días rotatorio

- **Ventajas:** el acceso frecuente no fuerza reautenticación; la aplicación
  puede restaurar su identidad sin red; la API mantiene revocación inmediata;
  no se incorporan JWT.
- **Inconvenientes:** dos secretos y una rotación atómica; una sesión usada de
  forma continuada no tiene límite absoluto.
- **Coste de mantenimiento:** medio; el coordinador cliente debe ser
  single-flight y el backend debe detectar el reuso de refresh.

### B — Refresh con límite absoluto de 30 días

- **Ventajas:** acota la vida máxima de una credencial renovada.
- **Inconvenientes:** obliga a reautenticarse aunque la persona use la app de
  forma continuada.
- **Coste de mantenimiento:** similar a A.

### C — Conservar la sesión única de ADR-0044

- **Ventajas:** menos estado y ningún endpoint de refresh.
- **Inconvenientes:** no satisface la restauración local sin consulta ni la
  renovación controlada solicitada.
- **Coste de mantenimiento:** bajo, pero producto insuficiente.

## Recomendación

**Recomendación:** alternativa A. Un límite absoluto mejora la exposición ante
robo de credenciales, pero el producto prioriza no interrumpir a quien usa la
aplicación de forma continuada; el logout explícito, la revocación y la detección
de reuso conservan controles recuperables.

## Decisión del usuario

**Aceptada el 2026-07-31:**

- access token opaco con expiración de 7 días;
- refresh token opaco con expiración de 30 días, rotado al renovar y sin límite
  absoluto mientras la rotación continúe;
- el cliente intenta renovar cuando falte una hora o el access token ya haya
  vencido, pero solo con un refresh todavía válido;
- una sola barrera concurrente comparte la renovación entre todas las
  peticiones protegidas;
- solo el coordinador de sesión puede borrar credenciales y reiniciar la
  navegación a la raíz anónima; un fallo de red no provoca logout.

## Consecuencias

- El contrato incorporará una operación de refresh y distinguirá expiración de
  access y refresh.
- El backend rota los secretos de forma atómica y registra el reuso de un
  refresh como señal de compromiso, revocando su familia de sesión.
- Al reiniciar, móvil puede mostrar la identidad persistida; toda operación
  protegida sigue necesitando autorización válida del backend.
- Logout, refresh inválido, revocado, vencido o reutilizado eliminan ambos
  secretos y el estado local, y resetean la navegación igual que un reemplazo de
  sesión por verificación.

## Validación

- Varias peticiones cercanas a la expiración desencadenan un único refresh.
- Una renovación correcta rota ambos secretos y actualiza las fechas locales.
- Un fallo de red no borra la sesión local.
- Un refresh inválido o reutilizado revoca la familia y deja el cliente en la
  raíz anónima.
- La sesión permanece utilizable indefinidamente mientras sus refreshes se
  roten antes de vencer.

## Disparadores de revisión

- Una incidencia de robo de dispositivo o token exige reintroducir un límite
  absoluto o administración de dispositivos.
- MFA, SSO o cierre global de sesiones exige una política de familias más rica.
