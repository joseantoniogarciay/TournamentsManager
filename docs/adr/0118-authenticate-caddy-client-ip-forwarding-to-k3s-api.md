# ADR-0118: Autenticar el reenvío de IP de cliente de Caddy hacia la API K3s

- **Estado:** Aceptado
- **Fecha:** 2026-09-05
- **Decisor:** Usuario, mediante decisión explícita
- **Propietario del análisis:** Codex como mentor técnico
- **Supera a:** Ninguno
- **Superado por:** Ninguno

## Problema

La API necesita distinguir visitantes para sus límites de abuso cuando recibe
tráfico a través de Cloudflare Tunnel, Caddy y Traefik. En K3s, la IP inmediata
de Traefik pertenece a una red de Pods dinámica. Confiar en todo ese CIDR
permitiría a cualquier Pod que alcanzase la API falsificar `X-Client-IP`.

## Contexto y restricciones

- Caddy es el único borde público y alcanza el LoadBalancer privado de Traefik
  conforme a ADR-0090 y ADR-0116.
- No hay una `NetworkPolicy` ejecutable ni una identidad de red estable para
  Traefik en el K3s actual.
- La VM tiene un nodo y recursos limitados; no se introduce un CNI, mTLS ni un
  proxy adicional solo para esta frontera.
- El repositorio es público: el secreto no entra en Git, manifiestos ni logs.

## Criterios

1. aceptar una IP reenviada solo desde el borde autorizado;
2. conservar el comportamiento seguro por defecto si falta o no coincide la
   credencial;
3. no confiar en una red completa de Pods;
4. mantener una rotación y un rollback operativos simples.

## Alternativas

### A — Credencial interna Caddy→API

Caddy sobrescribe `X-Client-IP` y `X-FastTourney-Edge-Token`; la API compara el
token con el Secret propio usando tiempo constante antes de aceptar la IP.

- **Ventajas:** frontera explícita, sin confiar en el CIDR de Pods ni instalar
  componentes; un token ausente falla hacia la IP inmediata.
- **Inconvenientes:** requiere distribuir y rotar un secreto en el Mac y K3s.
- **Coste de mantenimiento:** bajo; una rotación coordinada y validada.

### B — Confiar el CIDR de Pods

- **Ventajas:** solo configura `TRUSTED_PROXY_CIDRS`.
- **Inconvenientes:** cualquier Pod de la red confiada puede falsificar la
  cabecera; las IP dinámicas agravan ese alcance.
- **Coste de mantenimiento:** bajo, con una frontera de seguridad inaceptable.

### C — CNI con NetworkPolicy, mTLS o proxy dedicado

- **Ventajas:** identidad o aislamiento de red más fuertes.
- **Inconvenientes:** añade operación, recursos y más puntos de fallo para una
  VM de un nodo.
- **Coste de mantenimiento:** medio o alto.

## Comparación

B es la solución más corta pero contradice mínimo privilegio. C ofrece defensa
en profundidad, pero no responde proporcionalmente a la capacidad ni al riesgo
actual. A autentica el dato que influye en límites de abuso, conserva el
despliegue simple y deja C como disparador de evolución.

## Recomendación

**Recomendación:** A.

## Decisión del usuario

**Aceptada el 2026-09-05:** usar una credencial interna aleatoria entre Caddy y
la API K3s. La API solo acepta `X-Client-IP` cuando `X-FastTourney-Edge-Token`
coincide con el Secret configurado. `TRUSTED_PROXY_CIDRS` permanece vacío en
producción.

## Consecuencias

- La API no registra ni exporta la credencial ni la cabecera.
- El Secret `api-edge-proxy` queda separado de `api-runtime`.
- Caddy siempre sobrescribe ambas cabeceras hacia el upstream.
- Una discrepancia de rotación no permite suplantar IP, pero degrada los límites
  a la IP inmediata hasta corregirla.

## Validación

1. una petición con token correcto usa la IP reenviada;
2. token ausente, incorrecto o una IP inválida conservan la IP inmediata;
3. el Secret no aparece en Git, logs ni manifiestos renderizados;
4. la rotación crea el Secret y recarga Caddy sin publicar el valor;
5. el gate público verifica que los límites distinguen dos IP de prueba.

## Disparadores de revisión

- más de un borde o una API accesible por otros workloads;
- CNI con políticas de red exigibles o identidad de workload disponible;
- necesidad de mTLS, varios operadores o requisitos de cumplimiento superiores.

## Documentos afectados

- [Seguridad](../engineering/SECURITY.md)
- [Despliegue](../operations/DEPLOYMENT.md)
- [Ingress privado](../runbooks/k3s-private-ingress.md)
- [Administración remota](../runbooks/k3s-remote-administration.md)
