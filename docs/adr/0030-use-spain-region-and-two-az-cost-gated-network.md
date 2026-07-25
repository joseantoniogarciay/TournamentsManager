# ADR-0030: Usar la región España y una red en dos AZ con gasto autorizado

- **Estado:** Aceptado
- **Fecha:** 2026-07-25
- **Decisor:** Usuario, mediante aceptación explícita de la propuesta
- **Propietario del análisis:** Asistente como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

La futura red AWS necesita una región, espacio de direcciones y subredes que
permitan publicar la API mediante ALB y mantener PostgreSQL privado. La elección
debe ser comprensible, no producir gasto por sí misma y exigir autorización de
coste antes de crear recursos.

## Contexto y restricciones

- ADR-0029 fija ALB público, API Fargate restringida al ALB, PostgreSQL privado
  y ausencia inicial de NAT Gateway.
- AWS España (`eu-south-2`) dispone de tres AZ y de los servicios base acordados,
  incluido ECS con Fargate, ECR, RDS, VPC y ELB; requiere habilitación explícita
  para una cuenta nueva.
- Un ALB público necesita subredes en al menos dos AZ. RDS requiere que su DB
  subnet group cubra al menos dos AZ aunque la instancia inicial no sea Multi-AZ.
- `CIDR` es únicamente el rango de direcciones internas de una red; no es una IP
  pública, no se anuncia en Internet y no tiene coste.
- No se autoriza crear organización HCP, cuentas AWS, VPC, ALB, tareas, base de
  datos, direcciones IPv4 ni ningún recurso facturable mediante este ADR.

## Criterios de decisión

1. proximidad a usuarios europeos y disponibilidad del stack acordado;
2. separación clara entre entrada pública, API y datos privados;
3. dos AZ para que el ALB y la futura base de datos tengan una topología válida;
4. rangos privados simples, no superpuestos entre `nonprod` y `prod`;
5. ningún gasto recurrente sin estimación y autorización explícita del usuario.

## Alternativas

### Alternativa A — España, dos AZ y VPC separada por cuenta

Usar `eu-south-2`; una VPC `/16` en cada cuenta de workload, con dos subredes
públicas y dos privadas repartidas entre dos AZ.

- **Ventajas:** cercanía geográfica; tres AZ disponibles; cubre ALB, Fargate y
  RDS; un mapa simple y compatible con la publicación futura.
- **Inconvenientes:** la región debe habilitarse; ejecutar dos tareas, ALB o una
  base Multi-AZ costará cuando se autorice el despliegue.
- **Coste de adopción:** nulo hasta crear recursos; el coste se estima antes de
  cualquier `apply`.
- **Coste de mantenimiento:** bajo para una VPC por cuenta y cuatro subredes.
- **Riesgos:** asumir alta disponibilidad sin pagar las copias necesarias. Se
  mitiga distinguiendo el mapa de red de los recursos que se ejecutarán en él.

### Alternativa B — Irlanda, dos AZ y la misma topología

Usar `eu-west-1` con un diseño equivalente.

- **Ventajas:** región europea habitual en ejemplos y con un catálogo amplio.
- **Inconvenientes:** está más lejos del usuario inicial y no aporta una ventaja
  demostrada para los servicios ya elegidos.
- **Coste de adopción y mantenimiento:** equivalente; se debe comparar una
  estimación real antes de usarla como argumento de precio.
- **Riesgos:** elegir por popularidad en vez de por usuarios y requisitos.

### No cambiar — No fijar región ni red

Mantiene bloqueado Terraform cloud, porque no permite definir los recursos ni
estimar el coste de una publicación real.

## Comparación

España satisface los servicios previstos, tiene tres AZ y reduce la distancia
respecto al usuario inicial. Irlanda es viable, pero no ofrece un beneficio
actual que compense esa distancia. Dos AZ no cobran por existir: solo permiten
subredes separadas; los costes aparecen al desplegar copias, ALB, base de datos,
IPv4, logs o transferencia.

## Recomendación

**Opinión/recomendación:** alternativa A. Es el mapa mínimo suficiente para una
publicación pública seria sin NAT: cercano, simple y con dos ubicaciones
disponibles. La regla de autorización de coste evita convertir la decisión de
arquitectura en gasto automático.

## Decisión del usuario

**Aceptada:** usar Europa (España), `eu-south-2`, para `nonprod` y `prod` al
abrir la Fase 5. Cada cuenta tendrá una VPC independiente:

| Cuenta | CIDR VPC | Subredes |
| --- | --- | --- |
| `nonprod` | `10.42.0.0/16` | dos públicas y dos privadas, distribuidas entre dos AZ |
| `prod` | `10.43.0.0/16` | dos públicas y dos privadas, distribuidas entre dos AZ |

Cada subred inicial tendrá tamaño `/20`. Las públicas alojarán el ALB y las
tareas Fargate con egress público restringido por security groups; las privadas
se reservarán para PostgreSQL. Las AZ se seleccionarán por sus AZ IDs al
implementar, no por los sufijos de nombre que pueden variar entre cuentas.

No habrá gasto recurrente ni `terraform apply` hasta que exista una estimación
completa de ALB, tareas Fargate, RDS, IPv4, logs, almacenamiento y transferencia,
y el usuario autorice explícitamente el importe y la duración del despliegue.

## Consecuencias

### Positivas

- El stack futuro tiene una región, VPC y subredes coherentes para implementar
  Terraform sin inventar direcciones durante el apply.
- La base de datos cuenta con dos subredes privadas aptas para RDS.
- La arquitectura conserva dos AZ sin introducir NAT ni copias facturables ahora.
- El coste queda como gate explícito previo a la creación de recursos.

### Negativas y deuda aceptada

- España debe habilitarse en las cuentas y se dependerá inicialmente de una sola
  región.
- No existe recuperación ante la caída completa de la región; multi-región sería
  sobreingeniería para el alcance actual.
- El mapa permite alta disponibilidad, pero no la proporciona sin desplegar y
  pagar recursos redundantes.

## Validación

Antes de un apply cloud se demostrará que:

- `eu-south-2` está habilitada en las cuentas necesarias;
- las VPC y sus CIDR no se solapan;
- hay dos subredes públicas y dos privadas en AZ distintas por cuenta;
- ALB y RDS pueden usar sus subredes requeridas;
- no existe NAT Gateway;
- la estimación de coste y el periodo de ejecución tienen aprobación explícita.

## Disparadores de revisión

- usuarios, requisitos legales o latencia que justifiquen otra región;
- servicios AWS necesarios no disponibles en España;
- conectividad entre cuentas o redes que haga conflictivos los CIDR;
- requisito de recuperación regional o disponibilidad superior;
- coste real incompatible con el presupuesto autorizado.

## Documentación afectada

- [TECHNICAL_BASELINE.md](../governance/TECHNICAL_BASELINE.md)
- [DECISIONS.md](../governance/DECISIONS.md)
- [DECISIONS_TO_REVISIT.md](../governance/DECISIONS_TO_REVISIT.md)
- [DEPLOYMENT.md](../operations/DEPLOYMENT.md)
- [LEARNING.md](../project/LEARNING.md)
- [ROADMAP.md](../project/ROADMAP.md)
- [README.md](../../README.md)
- [CHANGELOG.md](../../CHANGELOG.md)

## Fuentes técnicas

- [Regiones y AZ de AWS](https://docs.aws.amazon.com/global-infrastructure/latest/regions/aws-regions.html)
- [Regiones compatibles con ECS en Fargate](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/AWS_Fargate-Regions.html)
- [Direcciones y CIDR de VPC](https://docs.aws.amazon.com/vpc/latest/userguide/vpc-ip-addressing.html)
- [Subredes de VPC](https://docs.aws.amazon.com/vpc/latest/userguide/configure-subnets.html)
- [Crear una instancia RDS](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_CreateDBInstance.html)
