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

1. En el Mac se construyen los targets `runtime` y `migrator` del Dockerfile
   usando tags que contienen el SHA completo del commit. El segundo queda
   preparado para migraciones explícitas, pero no se ejecuta ni recibe el
   Secret de runtime de la API.
2. Se exportan las imágenes como archivos OCI temporales fuera del repositorio.
3. Se copian a `/tmp` de la VM y se importan con `k3s ctr`.
4. El manifiesto `Deployment` referencia exactamente el tag `runtime` y declara que no
   debe intentar descargarlo de un registry.
5. El Secret `api-runtime` se crea desde un fichero privado con una única clave
   `DATABASE_URL`, usando exclusivamente `tournaments_manager_prod_app` y el
   host `postgresql.prod.svc.cluster.local`.
6. Tras verificar el Pod, los archivos temporales de `/tmp` se eliminan; las
   imágenes importadas permanecen en `containerd` mientras sean necesarias.

## Secret y despliegue de la API

La configuración actual mantiene la autenticación SMTP desactivada: el
`ConfigMap` deja `SMTP_USERNAME` vacío, por lo que `api-runtime` contiene solo
`DATABASE_URL`. Esto evita entregar a la API la credencial de migración o una
segunda credencial todavía no configurada para producción.

En una terminal SSH interactiva de la VM, el operador crea el fichero privado
sin mostrar su contenido y aplica el Secret y los manifiestos:

```sh
umask 077
editor /tmp/api-runtime.env
# DATABASE_URL=postgres://tournaments_manager_prod_app:<password>@postgresql.prod.svc.cluster.local:5432/fasttourney_prod?sslmode=disable
sudo /usr/local/bin/k3s kubectl create secret generic api-runtime \\
  --namespace prod --from-env-file=/tmp/api-runtime.env \\
  --dry-run=client -o yaml | sudo /usr/local/bin/k3s kubectl apply -f -
sudo /usr/local/bin/k3s kubectl apply -f infra/k3s/core/api-config.yaml -f infra/k3s/core/api.yaml
sudo /usr/local/bin/k3s kubectl -n prod rollout status deployment/api --timeout=120s
sudo /usr/local/bin/k3s kubectl -n prod get pods -l app.kubernetes.io/name=api
sudo /usr/local/bin/k3s kubectl -n prod get service api
rm -f /tmp/api-runtime.env
```

No se usa el Secret `postgresql-runtime` en el `Deployment` de API: contiene
otros valores y pertenece únicamente a PostgreSQL y al paso explícito de
migración.

Desde el Mac, antes de esa sesión, se construyen, exportan e importan ambos
artefactos. El directorio temporal no pertenece al repositorio:

```sh
SHA=$(git rev-parse HEAD)
docker build --platform linux/arm64 --target runtime \
  -t "tournaments-manager-api:git-$SHA" apps/backend
docker build --platform linux/arm64 --target migrator \
  -t "tournaments-manager-migrator:git-$SHA" apps/backend
docker save "tournaments-manager-api:git-$SHA" -o /tmp/tournaments-manager-api.tar
docker save "tournaments-manager-migrator:git-$SHA" -o /tmp/tournaments-manager-migrator.tar
scp /tmp/tournaments-manager-api.tar /tmp/tournaments-manager-migrator.tar <operator>@<k3s-host>:/tmp/
ssh -t <operator>@<k3s-host> \
  'sudo /usr/local/bin/k3s ctr images import /tmp/tournaments-manager-api.tar &&
   sudo /usr/local/bin/k3s ctr images import /tmp/tournaments-manager-migrator.tar &&
   sudo /usr/local/bin/k3s ctr images ls | grep tournaments-manager'
rm -f /tmp/tournaments-manager-api.tar /tmp/tournaments-manager-migrator.tar
```

Actualiza `infra/k3s/core/api.yaml` con el tag del `runtime` construido antes de
aplicar el Deployment. Las migraciones se ejecutarán posteriormente con el
target `migrator`, la identidad migradora y un Job explícito; importarlo no las
ejecuta.

Si las imágenes y los manifests ya están copiados a
`/tmp/tournaments-manager-k3s` de la VM, el operador puede ejecutar desde la
raíz del repositorio en el Mac el wrapper
`bash infra/k3s/scripts/deploy-api-from-staged-images.sh`. El wrapper abre SSH
por sí mismo, reserva el TTY para la contraseña de `sudo` y ejecuta un fichero
Bash no interactivo sin el perfil de la VM. No debe ejecutarse desde una sesión
SSH de la VM.

## Evolución prevista

Cuando se abra el laboratorio AWS, ECR privado almacenará la imagen y el
Deployment se identificará mediante `@sha256:...`, conforme a ADR-0024. En ese
momento se reemplazarán los pasos de exportación, copia e importación por la
descarga autenticada desde ECR, sin cambiar el contrato de configuración de la
API ni la definición de sus probes.
