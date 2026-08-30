# Importar una imagen de API en la VM K3s

> Estado: procedimiento de laboratorio para el runtime doméstico. No sustituye
> el registry ECR ni el despliegue futuro por digest de ADR-0024.

## Propósito y límite

K3s ejecuta sus contenedores con `containerd` dentro de la VM Ubuntu. Por eso
no puede ver automáticamente las imágenes que Docker construye en el Mac. Este
procedimiento transfiere de forma explícita una imagen OCI de API al runtime de
la VM para practicar el despliegue de un artefacto real sin crear aún un
registry, credenciales ni recursos AWS.

La identidad temporal de la imagen será una etiqueta trazable
`tournaments-manager-api:git-<SHA-completo>`. Solo se construye desde un commit
limpio: la etiqueta permite relacionar el artefacto local con su código fuente,
pero no reemplaza la identidad inmutable por digest que se adoptará con ECR.

## Prerrequisitos

- La VM K3s está sana y la sesión SSH del operador permanece disponible.
- El árbol Git está limpio y el commit que se va a probar ya existe localmente.
- Docker Desktop está disponible en el Mac y puede construir imágenes ARM64.
- El operador puede escribir su contraseña de `sudo` en la terminal SSH; no se
  guarda en el repositorio ni en archivos de secretos.

## Flujo previsto

1. En el Mac se construye el target `runtime` del Dockerfile de la API usando un
   tag que contiene el SHA completo del commit.
2. Se exporta la imagen como archivo OCI temporal fuera del repositorio.
3. Se copia el archivo a `/tmp` de la VM y se importa con `k3s ctr`.
4. El manifiesto `Deployment` referencia exactamente ese tag y declara que no
   debe intentar descargarlo de un registry.
5. Tras verificar el Pod, el archivo temporal de `/tmp` se elimina; la imagen
   importada permanece en `containerd` mientras el Deployment la necesite.

## Evolución prevista

Cuando se abra el laboratorio AWS, ECR privado almacenará la imagen y el
Deployment se identificará mediante `@sha256:...`, conforme a ADR-0024. En ese
momento se reemplazarán los pasos de exportación, copia e importación por la
descarga autenticada desde ECR, sin cambiar el contrato de configuración de la
API ni la definición de sus probes.
