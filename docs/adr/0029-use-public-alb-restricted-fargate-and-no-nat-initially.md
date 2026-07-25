# ADR-0029: Usar ALB público, Fargate restringido y sin NAT inicialmente

- **Estado:** Aceptado
- **Fecha:** 2026-07-25
- **Decisor:** Usuario, mediante aceptación explícita
- **Propietario del análisis:** Asistente como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

La API deberá ser accesible públicamente al publicar el producto, pero la red
debe limitar su exposición sin introducir el coste fijo de una NAT Gateway antes
de que aporte valor demostrado.

## Contexto y restricciones

- ADR-0023 fija ECS con Fargate como runtime cloud futuro y ADR-0026 separa
  `nonprod` y `prod` en cuentas AWS distintas.
- El proyecto es didáctico y no autoriza aún recursos AWS, gasto ni despliegue.
- El usuario quiere una publicación pública real cuando llegue el momento, no
  una arquitectura temporal que deba sustituirse sin motivo.
- Una NAT Gateway cobra por hora y por datos procesados; no encaja como coste
  fijo inicial sin una necesidad de aislamiento de egress demostrada.
- La región, los CIDR, las subredes concretas y el número de AZ siguen
  pendientes y no se deciden aquí.

## Criterios de decisión

1. publicar una entrada HTTPS estable para clientes web y mobile;
2. impedir que Internet llegue directamente a la API o a la base de datos;
3. evitar el coste fijo de NAT sin sacrificar controles de entrada;
4. mantener una evolución explícita a tareas privadas y NAT si aparece una
   necesidad real;
5. conservar una topología comprensible para aprender y operar.

## Alternativas

### Alternativa A — ALB público, tareas Fargate con entrada restringida y sin NAT

Un Application Load Balancer (ALB) público recibe HTTPS y reenvía tráfico a las
tareas Fargate. Las tareas pueden disponer de IP pública para su salida inicial,
pero su security group solo acepta en el puerto de la API al security group del
ALB. PostgreSQL permanece en subredes privadas y no acepta tráfico de Internet.

- **Ventajas:** dominio y TLS centralizados; health checks y enrutamiento de
  tráfico; la API no acepta entrada directa; elimina el coste fijo de NAT.
- **Inconvenientes:** las tareas tienen dirección pública para egress, aunque
  sus reglas bloquean toda entrada salvo la del ALB; ALB y las IPv4 públicas
  siguen teniendo coste.
- **Coste de adopción:** medio al abrir Fase 5: ALB, target group, certificados,
  security groups y reglas explícitas.
- **Coste de mantenimiento:** bajo para una API; exige revisar puertos,
  health checks y reglas cuando cambien los servicios.
- **Riesgos:** abrir por error el security group de la API a `0.0.0.0/0` o
  asumir que IP pública equivale a acceso público. Se mitiga con reglas que
  referencian al security group del ALB y validación automatizada.

### Alternativa B — Tareas privadas tras NAT Gateway

La API no tiene IP pública y usa una NAT Gateway para descargar imágenes,
publicar logs y acceder a Internet.

- **Ventajas:** frontera de red de egress más estricta; patrón habitual en
  entornos de producción.
- **Inconvenientes:** añade un coste fijo por hora y por GB, incluso sin tráfico,
  y su alta disponibilidad puede requerir una NAT por AZ.
- **Coste de adopción y mantenimiento:** medio o alto para el alcance actual.
- **Riesgos:** pagar por complejidad sin una necesidad de egress o disponibilidad
  demostrada.

### Alternativa C — API expuesta directamente

Asignar una IP pública a la tarea y permitir acceso de Internet directamente.

- **Ventajas:** menos recursos y coste inicial menor.
- **Inconvenientes:** no ofrece punto de entrada estable, TLS centralizado ni
  health checks del ALB; expone directamente la API.
- **Coste de adopción:** bajo.
- **Coste de mantenimiento:** aumenta al introducir dominio, HTTPS o varias
  tareas.
- **Riesgos:** superficie de ataque mayor y cambios de dirección al sustituir
  una tarea.

## Comparación

La alternativa A conserva una frontera clara de entrada: AWS recomienda que los
targets acepten tráfico exclusivamente desde el security group del ALB. El ALB
es además el punto único de entrada, reparte tráfico y monitoriza health checks.
La B aporta aislamiento adicional de egress, pero su coste fijo no está
justificado para el aprendizaje actual. La C evita el ALB, pero no satisface una
publicación estable y protegida.

## Recomendación

**Opinión/recomendación:** alternativa A. Aporta la protección que importa para
la API pública —solo el ALB puede entrar— y deja la NAT para cuando haya una
necesidad de egress privado, disponibilidad o cumplimiento que la justifique.

## Decisión del usuario

**Aceptada:** al publicar, la entrada pública será un ALB. Las tareas Fargate
aceptarán tráfico entrante únicamente desde el security group del ALB, aunque
inicialmente dispongan de IP pública para egress. PostgreSQL no tendrá IP pública
y solo aceptará la conexión necesaria desde la API. No se creará NAT Gateway en
la topología inicial.

No se autoriza crear recursos AWS ni gasto. Antes del primer despliegue se
estimará el coste real del ALB, Fargate, base de datos, IPv4, logs y transferencia
para que el usuario autorice el presupuesto. Región y subredes siguen pendientes.

## Consecuencias

### Positivas

- La API se publica tras una entrada HTTPS estable, con health checks y reglas
  de entrada explícitas.
- La base de datos permanece fuera de Internet.
- Se elimina inicialmente el coste fijo de NAT.

### Negativas y deuda aceptada

- Se pagan ALB, capacidad Fargate, IPv4 pública, datos y logs cuando exista
  despliegue.
- La salida de las tareas no queda aislada mediante NAT; se reabrirá la decisión
  si el riesgo o la disponibilidad lo requieren.

## Validación

Antes de publicar se demostrará que:

- Internet solo alcanza el ALB por HTTPS;
- una conexión directa a la IP de una tarea Fargate es rechazada;
- el ALB solo enruta hacia tareas sanas;
- PostgreSQL no tiene IP pública y rechaza conexiones salvo desde la API;
- no se ha creado una NAT Gateway;
- la estimación de coste está revisada y autorizada antes de `apply`.

## Disparadores de revisión

- requisito de egress privado, allowlists salientes o cumplimiento;
- indisponibilidad de egress que exija NAT por AZ;
- coste del ALB o IPv4 incompatible con la carga real;
- varios servicios, rutas o equipos que requieran otra frontera de red;
- incidente de acceso o configuración que demuestre reglas insuficientes.

## Documentación afectada

- [TECHNICAL_BASELINE.md](../governance/TECHNICAL_BASELINE.md)
- [DECISIONS.md](../governance/DECISIONS.md)
- [DECISIONS_TO_REVISIT.md](../governance/DECISIONS_TO_REVISIT.md)
- [DEPLOYMENT.md](../operations/DEPLOYMENT.md)
- [SECURITY.md](../engineering/SECURITY.md)
- [LEARNING.md](../project/LEARNING.md)
- [README.md](../../README.md)
- [CHANGELOG.md](../../CHANGELOG.md)

## Fuentes técnicas

- [Application Load Balancer](https://docs.aws.amazon.com/elasticloadbalancing/latest/application/introduction.html)
- [Security groups de un ALB](https://docs.aws.amazon.com/elasticloadbalancing/latest/application/load-balancer-update-security-groups.html)
- [Fargate y red](https://docs.aws.amazon.com/pdfs/decision-guides/latest/fargate-or-lambda/fargate-or-lambda.pdf)
- [Precios de NAT Gateway](https://docs.aws.amazon.com/vpc/latest/userguide/nat-gateway-pricing.html)
- [Precios de VPC e IPv4 pública](https://aws.amazon.com/vpc/pricing/)
