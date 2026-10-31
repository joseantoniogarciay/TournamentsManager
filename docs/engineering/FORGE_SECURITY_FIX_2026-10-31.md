# Corrección de Forge — 2026-10-31

## Decisión y resultado

El usuario autoriza «Si soluciona lo metemos» tras revisar los seis altos.
[ADR-0173](../adr/0173-apply-forge-security-fix-with-exact-age-exception.md)
concreta la excepción puntual en **node-forge 1.4.2**. Se sustituye 1.4.0 en
Expo CLI y code-signing-certificates, sin cambiar otra versión del grafo.

**Auditoría npm: 15 → 3 registros, todos moderados. Cero altos, bajos o críticos.**
Desaparecen los doce registros de Forge, incluidos los seis altos originales;
no aparecen IDs nuevos. Quedan @hono/node-server 1.19.17, glob 13.0.6 y
minimatch 3.1.5/10.2.6. Conservar los motivos y fechas de revisión del
[informe anterior](DEPENDENCY_SECURITY_REVIEW_2026-10-31.md).
El resultado se refiere al checkout corregido, no a producción.

## Por qué 1.4.2

La ficha antigua GHSA-86w9-cpqp-85rv sigue sin declarar parche. El mantenedor
publica para el mismo CVE-2026-85393 el aviso GHSA-2g46-22mc-6gwp: el problema
se corrige en 1.4.1 comparando el DigestInfo con su codificación canónica.
El changelog de 1.4.1 incluye las seis correcciones altas del inventario.

La versión 1.4.1 introduce además un TypeError para entradas PEM de tipo Buffer.
El único cambio documentado en 1.4.2 restaura esa compatibilidad. Se elige el
hotfix exacto para conservar los parches y evitar una regresión conocida.
Los módulos RSA y PEM instalados desde npm coinciden byte a byte con el tag
v1.4.2 oficial; sus hashes quedan en la evidencia.

## Cadena de suministro y excepción

Se resuelve primero un candidato aislado con los manifiestos y parches actuales.
La comparación de paquetes del lockfile confirma una única sustitución:
node-forge@1.4.0 → node-forge@1.4.2. No se añaden dependencias ni se cambia
Expo, módulos nativos, contrato HTTP, UI o telemetría.

1.4.2 se publica el 2026-10-27T14:09:15.471Z. El usuario acepta instalar la
corrección ahora: se añade minimumReleaseAgeExclude únicamente para esa versión
y el selector node-forge@1.4.0. La espera global de siete días y las demás
excepciones exactas permanecen. **Retirar esta excepción el 3 de noviembre a
las 15:09:15.471 de Madrid**, conservando la versión corregida. No es una
excepción genérica para vulnerabilidades altas ni para todo Forge.

## Verificación de seguridad y compatibilidad

- Instalación congelada correcta. Las dos rutas Expo resuelven 1.4.2 y el
  lockfile ya no contiene 1.4.0.
- 254 pruebas oficiales de RSA, PEM, TLS, PKCS#12 y X.509 correctas contra el
  paquete instalado. Se usa el código de pruebas del tag v1.4.1 con un adaptador
  local de describe/it, sin instalar Mocha. Cuatro comparaciones de generación
  determinista con PRNG se omiten porque Node usa generación nativa no
  determinista; no son pruebas omitidas de los seis avisos altos.
- Control negativo del aviso RSA anidado: el vector oficial falla la aserción
  de rechazo sobre 1.4.0 y se rechaza correctamente sobre 1.4.2. Esto verifica
  el parche concreto más allá del conteo del auditor.
- Nueva regresión permanente del consumidor Expo: genera y relee certificado y
  claves PEM, conserva pem.decode(Buffer), valida el certificado, firma un
  payload y verifica que una modificación del payload invalida la firma.
- expo install --check acepta la matriz local bundledNativeModules.json en
  modo offline; no se pudo contrastar el endpoint remoto de versiones. El
  lockfile confirma que no cambia ningún paquete nativo.
- `make verify` completo correcto: formato, lint, tipos, regresiones,
  exportación web, generación comprobada OpenAPI/sqlc, build y govulncheck.
- 126/126 regresiones Node correctas. Integración PostgreSQL omitida localmente
  por ausencia de base aislada; CI del bloque previo terminó correctamente
  con su PostgreSQL. No se atribuye ese resultado al nuevo commit.

La auditoría reutiliza la autorización explícita de enviar el grafo al servicio
público npm concedida en esta sesión. Exit 1 corresponde a los tres moderados
restantes. Evidencia compacta versionada en
[forge-result.json](evidence/dependency-security-2026-10-31/forge-result.json);
logs, fuentes oficiales y detalles de pruebas en `.tmp/forge-security-fix-20261031/`.
No se arrancan servicios persistentes local/dev ni observabilidad; no se
publica o despliega. Este bloque se sube a develop conforme a la instrucción
vigente del usuario.

## Retrospectiva

Una ficha del agregador desactualizada no sustituye al aviso del mantenedor.
Relacionar CVE, parche y regresión permite resolver la discrepancia de RSA.
A la vez, revisar el siguiente parche evita adoptar una corrección que rompe
un formato de entrada existente. El control negativo demuestra que la prueba
puede detectar el fallo anterior; la prueba del consumidor protege el uso
real de Expo. La excepción de edad es una decisión puntual, no una nueva
política de actualización.

## Fuentes

- [Mantenedor: CVE-2026-85393, corregido en 1.4.1](https://github.com/digitalbazaar/forge/security/advisories/GHSA-2g46-22mc-6gwp).
- [Changelog de Forge 1.4.2](https://github.com/digitalbazaar/forge/blob/v1.4.2/CHANGELOG.md).
- [Pruebas RSA oficiales](https://github.com/digitalbazaar/forge/blob/v1.4.1/tests/unit/rsa.js).
- [Pruebas PEM oficiales](https://github.com/digitalbazaar/forge/blob/v1.4.1/tests/unit/pem.js).
- [Pruebas TLS oficiales](https://github.com/digitalbazaar/forge/blob/v1.4.1/tests/unit/tls.js).
- [Pruebas PKCS#12 oficiales](https://github.com/digitalbazaar/forge/blob/v1.4.1/tests/unit/pkcs12.js).
- [Pruebas X.509 oficiales](https://github.com/digitalbazaar/forge/blob/v1.4.1/tests/unit/x509.js).
- [Registro npm](https://registry.npmjs.org/node-forge).
