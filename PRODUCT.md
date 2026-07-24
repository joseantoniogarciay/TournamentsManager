# Producto

> Estado: alcance inicial aceptado; diseño funcional pausado durante el gate
> técnico.
>
> Última actualización: 2026-07-24

## Visión

TournamentsManager permitirá descubrir, crear y gestionar torneos. El fútbol será
el primer deporte para aprender y validar el modelo, pero el producto aspira a
incorporar otros deportes sin reescribir su núcleo.

## Hechos aceptados

- Existirán aplicaciones web y mobile.
- Web, iOS y Android ofrecerán el mismo producto con paridad funcional.
- La experiencia será responsive en navegadores y aplicaciones de móvil, tablet y
  escritorio, adaptando la presentación cuando corresponda.
- Una persona invitada podrá consultar torneos públicos ya creados.
- Crear un torneo o unirse a uno exigirá una cuenta.
- La cuenta incluirá registro, inicio de sesión y recuperación de contraseña.
- El fútbol es el deporte inicial.
- Las acciones detalladas de creación y gestión de un torneo se definirán de forma
  incremental.
- Se ha aceptado un cliente universal con React Native; Expo y las decisiones de
  framework, routing y rendering siguen pendientes.

## Actores y capacidades iniciales

### Invitado

- ver el listado de torneos públicos;
- abrir el detalle público de un torneo;
- iniciar registro o login cuando intente una acción protegida.

Un invitado no es una cuenta con rol especial: es una persona sin sesión
autenticada.

### Usuario autenticado

- gestionar su sesión y perfil básico;
- crear un torneo;
- solicitar, aceptar o ejecutar la incorporación a un torneo según el mecanismo
  que se decida;
- consultar los torneos con los que tiene relación.

### Organizador

Es un usuario autenticado con permisos sobre un torneo concreto. La matriz exacta
de permisos se decidirá junto al ciclo de vida del torneo.

### Participante

Es un usuario o entidad vinculada a un torneo. Todavía debe decidirse si una
persona se une directamente, se une mediante un equipo, representa a un equipo o
puede usar varios modelos según el deporte.

## Flujos de identidad

### Registro

1. La persona proporciona el identificador acordado, inicialmente candidato:
   correo electrónico.
2. Acepta las condiciones necesarias.
3. Verifica la propiedad del canal si se exige.
4. Se crea o vincula el perfil interno.

### Login

1. La persona demuestra su identidad mediante el mecanismo elegido.
2. El cliente obtiene una sesión apropiada para web o mobile.
3. El backend autoriza cada acción sobre recursos concretos.

### Recuperación de contraseña

“Recordar contraseña” se interpreta como recuperación segura: nunca se recupera
la contraseña anterior. Se emitirá un token o código temporal mediante un canal
verificado y se permitirá establecer una nueva.

Mantener una sesión abierta (“recuérdame”) es una decisión diferente sobre
duración y renovación de sesiones.

## Primer vertical slice candidato

El primer corte debe ser pequeño y atravesar producto, seguridad, datos, API y
operación:

1. un invitado lista y consulta torneos públicos;
2. una persona se registra, verifica su cuenta e inicia sesión;
3. un usuario autenticado crea un torneo de fútbol con los datos mínimos;
4. otro usuario se une mediante el mecanismo elegido;
5. ambos observan el estado actualizado.

Antes de implementarlo deben decidirse formato, modelo de participante,
visibilidad, incorporación e identidad.

## Fuera del primer alcance

Salvo decisión posterior:

- pagos y premios;
- streaming o contenido multimedia;
- chat;
- marketplace;
- microservicios;
- motor genérico para todos los deportes;
- rankings globales;
- administración avanzada;
- calendario completo y arbitraje;
- notificaciones push.

## Preguntas de producto prioritarias

1. ¿Quién crea el torneo y qué puede delegar?
2. ¿Se unen usuarios, equipos o ambos?
3. ¿Cuál es el primer formato: liga, eliminatoria o grupos más eliminatoria?
4. ¿Un torneo nace borrador y después se publica?
5. ¿Puede ser público, privado o no listado?
6. ¿Cómo se entra: código, invitación, solicitud o enlace?
7. ¿Qué datos públicos se muestran sin login?
8. ¿Quién registra y confirma resultados?
9. ¿Qué ocurre si alguien abandona o es expulsado?
10. ¿Qué estados cierran o cancelan un torneo?

Estas preguntas preceden al esquema de datos y a los endpoints.
Se retomarán cuando termine el
[Technical Baseline](TECHNICAL_BASELINE.md), conforme a
[ADR-0004](docs/adr/0004-technical-baseline-before-product-design.md).
