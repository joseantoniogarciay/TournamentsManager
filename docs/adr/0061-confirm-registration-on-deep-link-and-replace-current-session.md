# ADR-0061: Confirmar el registro al abrir el enlace y sustituir la sesión actual

- **Estado:** Aceptado
- **Fecha:** 2026-07-30
- **Decisor:** Usuario
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** ADR-0044 y la regla de confirmación explícita de `SECURITY.md`, exclusivamente para la verificación de un registro local mediante deep link
- **Superado por:** Ninguno

## Problema

La persona ya acreditó que controla el correo al abrir el enlace único. Pedir un segundo toque añade una pantalla intermedia sin recuperación distinta. Al terminar, el cliente debe representar de inmediato la identidad nueva y no conservar una sesión anterior de otra cuenta.

## Contexto y restricciones

- El token sigue siendo aleatorio, opaco, de un solo uso y con vencimiento de 24 horas según ADR-0044.
- Abrir la URL con `GET` no muta el backend: la aplicación extrae el token, elimina la URL del historial y ejecuta el `POST` automáticamente.
- La sesión anterior puede venir por cookie web o Bearer móvil. Debe revocarse en el backend y sustituirse en el cliente antes de mostrar el estado nuevo.
- El login local completo y el contenido autenticado definitivo siguen fuera de este corte. Cuenta muestra un estado autenticado temporal verificable.

## Criterios de decisión

1. reducir pasos sin convertir el `GET` en una mutación;
2. impedir que la sesión anterior siga siendo utilizable;
3. restaurar Inicio, Torneos y Cuenta desde su raíz con los datos de la nueva identidad;
4. conservar una URL web predecible: tras confirmar, `/`.

## Alternativas

### A — Confirmación automática mediante POST y reemplazo de sesión

- **Ventajas:** elimina el toque redundante; preserva el límite HTTP seguro; garantiza una sola sesión actual en el dispositivo.
- **Inconvenientes:** abrir el enlace finaliza el registro de inmediato.
- **Coste de adopción:** bajo; amplía la transacción de verificación con la revocación de la credencial presentada.
- **Coste de mantenimiento:** bajo; reutiliza el endpoint y el transporte ya definidos.
- **Riesgos:** un enlace abierto involuntariamente confirma la cuenta; se acepta porque el buzón es el canal de prueba y el token es de un solo uso.

### B — Botón de confirmación explícita

- **Ventajas:** ofrece una pausa antes de cambiar la sesión.
- **Inconvenientes:** añade un paso sin recuperación diferente y contradice la experiencia solicitada.
- **Coste de mantenimiento:** bajo.

### No cambiar

- **Consecuencias:** se conserva una pantalla intermedia y la sesión anterior puede persistir hasta que el cliente la reemplace por su cuenta.

## Comparación

La alternativa A mantiene la mutación protegida por `POST` y la atomicidad del token, pero elimina interacción redundante. B es más conservadora en UX, no en seguridad del protocolo. No cambiar no satisface el recorrido solicitado.

## Recomendación

**Opinión/recomendación:** alternativa A, limitada al enlace de verificación de registro local y con revocación de la credencial actualmente presentada.

## Decisión del usuario

**Aceptada el 2026-07-30:** alternativa A. Tras alta local, el cliente vuelve a Login con un aviso breve de correo enviado. Al abrir el deep link, confirma de forma automática, descarta la sesión previa, restaura las tres áreas desde su raíz y, en web, termina en Inicio (`/`). Cuenta muestra temporalmente un mock de sesión autenticada hasta definir su contenido final.

### Aclaración de implementación — 2026-07-31

Durante el `POST` y hasta que la nueva raíz esté montada, el cliente muestra una
capa global semitransparente que bloquea interacción y centra un indicador de
progreso. Entra y sale con la duración de movimiento semántica, salvo que la
plataforma solicite movimiento reducido.

El restablecimiento nativo no se implementa cerrando una modal concreta:
reconstruye el árbol raíz de navegación para descartar todas las modales y las
pilas de Inicio, Torneos y Cuenta, incluso si el enlace llegó con la aplicación
ya abierta. En web, se reemplaza la URL por `/`. El reinicio de navegación no
autoriza prefetchar colecciones: cada raíz solicita sus datos solo al entrar o
recibir foco.

## Consecuencias

### Positivas

- El recorrido tiene un único paso humano: abrir el correo.
- Una cookie o Bearer previo deja de ser válido antes de usar la nueva sesión.
- Las pantallas no conservan estado de invitado de una identidad anterior.

### Negativas y deuda aceptada

- No hay deshacer después de abrir el enlace; la cuenta queda verificada.
- El estado autenticado de Cuenta es un mock hasta que se implemente lectura de sesión y sus colecciones.

## Validación

- Un deep link válido activa la cuenta sin botón y llega a `/` en web.
- La sesión entregada en la petición de verificación queda revocada; la nueva credencial autentica a la persona registrada.
- Inicio, Torneos y Cuenta observan el reinicio de sesión; Cuenta muestra el mock de la nueva identidad.
- En iOS y Android, un enlace abierto sobre una o más modales deja las tres tabs
  en sus raíces y no emite acciones de navegación sin manejador.
- La transición impide acciones duplicadas, cubre toda la interfaz y respeta la
  preferencia de movimiento reducido.

## Disparadores de revisión

- Soporte o fraude demuestra que es necesario un segundo gesto de confirmación.
- Se incorporan cambios de email, vinculación de identidades o requisitos de reautenticación que exijan una política distinta de reemplazo.

## Documentación afectada

- `docs/engineering/IDENTITY.md`
- `docs/engineering/SECURITY.md`
- `docs/engineering/ARCHITECTURE.md`
- `docs/project/LEARNING.md`
- `CHANGELOG.md`
