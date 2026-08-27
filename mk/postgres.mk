# PostgreSQL local: Docker Compose, ciclo de vida y esquema inicial explícito.
# API Go y cliente Expo se ejecutan en el host durante desarrollo.

POSTGRES_LOCAL_DIR := infra/local
POSTGRES_LOCAL_ENV := $(POSTGRES_LOCAL_DIR)/.env
BACKEND_ENV := $(BACKEND_DIR)/.env
DEV_API_ENV := $(POSTGRES_LOCAL_DIR)/api.docker.env
PUBLIC_DEV_DIR := infra/dev
PUBLIC_DEV_ENV := $(PUBLIC_DEV_DIR)/.env
PUBLIC_DEV_API_ENV := $(PUBLIC_DEV_DIR)/api.docker.env
POSTGRES_COMPOSE := docker compose --env-file $(POSTGRES_LOCAL_ENV) -f $(POSTGRES_LOCAL_DIR)/compose.yaml
DEV_COMPOSE := docker compose --env-file $(POSTGRES_LOCAL_ENV) -f $(POSTGRES_LOCAL_DIR)/compose.dev.yaml
PUBLIC_DEV_COMPOSE := docker compose --env-file $(PUBLIC_DEV_ENV) -f $(PUBLIC_DEV_DIR)/compose.yaml

.PHONY: \
	db-init dev-init db-env-check db-backend-env-check dev-api-env-check local-config-check dev-config-check \
	dev-public-init dev-public-config-check dev-public-up dev-public-deploy dev-public-rollback dev-public-down dev-public-reset dev-public-status dev-public-logs dev-public-bootstrap dev-public-migrate dev-public-runtime-verify dev-public-purge dev-public-backup-init dev-public-backup-full dev-public-backup-incremental dev-public-backup-status dev-public-backup-restore-verify \
	db-up db-wait db-down db-status db-logs db-reset db-schema-apply

# Crea los contratos locales sin sobrescribir una configuración ya existente.
db-init:
	@test ! -e $(POSTGRES_LOCAL_ENV) || { echo "db-init: $(POSTGRES_LOCAL_ENV) ya existe; no se sobrescribe"; exit 1; }
	@test ! -e $(BACKEND_ENV) || { echo "db-init: $(BACKEND_ENV) ya existe; no se sobrescribe"; exit 1; }
	cp $(POSTGRES_LOCAL_DIR)/.env.example $(POSTGRES_LOCAL_ENV)
	cp $(BACKEND_DIR)/.env.example $(BACKEND_ENV)
	@echo "db-init: contratos creados; revisa las contraseñas locales antes de arrancar"

# Prepara los contratos usados por Compose sin sobrescribir una configuración ya
# existente. apps/backend/.env solo hace falta para el arranque host opcional.
dev-init:
	@if [ ! -e $(POSTGRES_LOCAL_ENV) ]; then cp $(POSTGRES_LOCAL_DIR)/.env.example $(POSTGRES_LOCAL_ENV); else echo "dev-init: $(POSTGRES_LOCAL_ENV) ya existe; se conserva"; fi
	@if [ ! -e $(DEV_API_ENV) ]; then cp $(POSTGRES_LOCAL_DIR)/api.docker.env.example $(DEV_API_ENV); else echo "dev-init: $(DEV_API_ENV) ya existe; se conserva"; fi
	@echo "dev-init: contratos Compose preparados; revisa las contraseñas locales antes de arrancar"

dev-public-init:
	@if [ ! -e $(PUBLIC_DEV_ENV) ]; then cp $(PUBLIC_DEV_DIR)/.env.example $(PUBLIC_DEV_ENV); else echo "dev-public-init: $(PUBLIC_DEV_ENV) ya existe; se conserva"; fi
	@if [ ! -e $(PUBLIC_DEV_API_ENV) ]; then cp $(PUBLIC_DEV_DIR)/api.docker.env.example $(PUBLIC_DEV_API_ENV); else echo "dev-public-init: $(PUBLIC_DEV_API_ENV) ya existe; se conserva"; fi
	@echo "dev-public-init: contratos creados; configura credenciales distintas para administrador, migrador y API"

db-env-check:
	@test -f $(POSTGRES_LOCAL_ENV) || { echo "Falta $(POSTGRES_LOCAL_ENV). Ejecuta: make db-init"; exit 1; }
	@set -a; . $(POSTGRES_LOCAL_ENV); set +a; \
	for name in POSTGRES_DB POSTGRES_USER POSTGRES_PASSWORD; do \
		eval "value=\$${$$name}"; \
		[ -n "$$value" ] || { echo "Falta $$name en $(POSTGRES_LOCAL_ENV). Revisa infra/local/.env.example"; exit 1; }; \
	done

db-backend-env-check:
	@test -f $(BACKEND_ENV) || { echo "Falta $(BACKEND_ENV). Ejecuta: make db-init"; exit 1; }

dev-api-env-check:
	@test -f $(DEV_API_ENV) || { echo "Falta $(DEV_API_ENV). Ejecuta: make dev-init"; exit 1; }

# Valida los contratos locales sin mostrar valores ni sobrescribir archivos.
local-config-check: db-env-check db-backend-env-check
	@set -a; . $(BACKEND_ENV); set +a; \
	for name in DATABASE_URL HTTP_ADDR SMTP_ADDR SMTP_FROM PUBLIC_BASE_URL CORS_ALLOWED_ORIGINS; do \
		eval "value=\$${$$name}"; \
		[ -n "$$value" ] || { echo "Falta $$name en $(BACKEND_ENV). Revisa apps/backend/.env.example"; exit 1; }; \
	done

dev-config-check: db-env-check dev-api-env-check
	@set -a; . $(DEV_API_ENV); set +a; \
	for name in DATABASE_URL HTTP_ADDR SMTP_ADDR SMTP_FROM PUBLIC_BASE_URL CORS_ALLOWED_ORIGINS; do \
		eval "value=\$${$$name}"; \
		[ -n "$$value" ] || { echo "Falta $$name en $(DEV_API_ENV). Revisa infra/local/api.docker.env.example"; exit 1; }; \
	done

dev-public-config-check:
	@test -f $(PUBLIC_DEV_ENV) || { echo "Falta $(PUBLIC_DEV_ENV). Ejecuta: make dev-public-init"; exit 1; }
	@test -f $(PUBLIC_DEV_API_ENV) || { echo "Falta $(PUBLIC_DEV_API_ENV). Ejecuta: make dev-public-init"; exit 1; }
	@set -a; . $(PUBLIC_DEV_ENV); set +a; \
	for name in POSTGRES_DB POSTGRES_USER POSTGRES_PASSWORD POSTGRES_OWNER_ROLE POSTGRES_MIGRATOR_USER POSTGRES_MIGRATOR_PASSWORD POSTGRES_APP_USER POSTGRES_APP_PASSWORD POSTGRES_BACKUP_DESTINATION PGBACKREST_REPO1_CIPHER_PASS; do \
		eval "value=\$${$$name}"; \
		[ -n "$$value" ] || { echo "Falta $$name en $(PUBLIC_DEV_ENV)"; exit 1; }; \
	done
	@set -a; . $(PUBLIC_DEV_API_ENV); set +a; \
	for name in DATABASE_URL HTTP_ADDR SMTP_ADDR SMTP_FROM SMTP_USERNAME SMTP_PASSWORD PUBLIC_BASE_URL CORS_ALLOWED_ORIGINS; do \
		eval "value=\$${$$name}"; \
		[ -n "$$value" ] || { echo "Falta $$name en $(PUBLIC_DEV_API_ENV)"; exit 1; }; \
	done

dev-public-up: dev-public-config-check
	$(PUBLIC_DEV_COMPOSE) up --build --detach --wait --remove-orphans

# Despliegue manual y recuperable de origin/develop: conserva dos artefactos.
dev-public-deploy: dev-public-config-check
	./infra/home/deploy-dev.sh

dev-public-rollback: dev-public-config-check
	@test -n "$(SHA)" || { echo "Uso: make dev-public-rollback SHA=<SHA-completo>"; exit 1; }
	./infra/home/rollback-dev.sh "$(SHA)"

dev-public-down: dev-public-config-check
	$(PUBLIC_DEV_COMPOSE) down --remove-orphans

# Acción destructiva acotada al volumen PostgreSQL de tournaments-manager-dev.
# No afecta a tournaments-manager-local ni a sus datos.
dev-public-reset: dev-public-config-check
	@printf "Esto eliminará todos los datos de PostgreSQL en tournaments-manager-dev. Escribe DEV_RESET para continuar: "; read answer; \
	[ "$$answer" = "DEV_RESET" ] || { echo "dev-public-reset: cancelado"; exit 1; }
	$(PUBLIC_DEV_COMPOSE) down --volumes --remove-orphans

dev-public-status: dev-public-config-check
	$(PUBLIC_DEV_COMPOSE) ps

dev-public-logs: dev-public-config-check
	$(PUBLIC_DEV_COMPOSE) logs --tail=200

# Inicializa una base vacía siguiendo ADR-0097: roles primero, esquema con el
# migrador como owner, permisos de API al final. No se ejecuta en despliegues.
dev-public-bootstrap: dev-public-config-check
	$(PUBLIC_DEV_COMPOSE) up --detach --wait postgres
	$(PUBLIC_DEV_COMPOSE) exec -T postgres sh -ec 'psql -U "$$POSTGRES_USER" -d "$$POSTGRES_DB" -v ON_ERROR_STOP=1 -v database_name="$$POSTGRES_DB" -v owner_role="$$POSTGRES_OWNER_ROLE" -v migrator_role="$$POSTGRES_MIGRATOR_USER" -v migrator_password="$$POSTGRES_MIGRATOR_PASSWORD" -v app_role="$$POSTGRES_APP_USER" -v app_password="$$POSTGRES_APP_PASSWORD"' < $(PUBLIC_DEV_DIR)/postgresql/bootstrap-roles.sql
	@set -a; . $(PUBLIC_DEV_ENV); set +a; \
	sed '/^--/d' $(BACKEND_DIR)/db/schema/initial_schema.sql | \
		$(PUBLIC_DEV_COMPOSE) exec -T postgres sh -ec 'PGPASSWORD="$$POSTGRES_MIGRATOR_PASSWORD" psql -h postgres -U "$$POSTGRES_MIGRATOR_USER" -d "$$POSTGRES_DB" -v ON_ERROR_STOP=1 -c "SET ROLE \"$$POSTGRES_OWNER_ROLE\"" -f -'
	$(PUBLIC_DEV_COMPOSE) exec -T postgres sh -ec 'PGPASSWORD="$$POSTGRES_MIGRATOR_PASSWORD" psql -h postgres -U "$$POSTGRES_MIGRATOR_USER" -d "$$POSTGRES_DB" -v ON_ERROR_STOP=1 -v app_role="$$POSTGRES_APP_USER" -c "SET ROLE \"$$POSTGRES_OWNER_ROLE\"" -f -' < $(PUBLIC_DEV_DIR)/postgresql/grant-runtime.sql
	$(MAKE) dev-public-migrate
	$(MAKE) dev-public-runtime-verify

# Ejecuta las migraciones inmutables fuera del arranque de la API. La contraseña
# de migración existe solo en el contenedor efímero `migrator`.
dev-public-migrate: dev-public-config-check
	$(PUBLIC_DEV_COMPOSE) --profile migration up --detach --wait postgres
	$(PUBLIC_DEV_COMPOSE) --profile migration build migrator
	$(PUBLIC_DEV_COMPOSE) --profile migration run --rm --no-deps migrator

# Comprueba que la identidad de la API conecta pero no puede cambiar el esquema.
dev-public-runtime-verify: dev-public-config-check
	$(PUBLIC_DEV_COMPOSE) exec -T postgres sh -ec 'PGPASSWORD="$$POSTGRES_APP_PASSWORD" psql -h postgres -U "$$POSTGRES_APP_USER" -d "$$POSTGRES_DB" -v ON_ERROR_STOP=1 -c "SELECT current_user"'
	@! $(PUBLIC_DEV_COMPOSE) exec -T postgres sh -ec 'PGPASSWORD="$$POSTGRES_APP_PASSWORD" psql -h postgres -U "$$POSTGRES_APP_USER" -d "$$POSTGRES_DB" -v ON_ERROR_STOP=1 -c "CREATE TABLE permission_probe (id integer)"' || { echo "La identidad runtime no debe poder crear tablas"; exit 1; }

# Purga exclusivamente cuentas de desarrollo cuya baja ya venció. El comando
# interno no expone una ruta HTTP y reutiliza la API runtime ya en ejecución.
dev-public-purge: dev-public-config-check
	$(PUBLIC_DEV_COMPOSE) exec -T api /api purge-expired-accounts

# Inicializa una vez el repositorio cifrado, prueba el archivado WAL y toma la
# base completa. No se ejecuta durante el arranque ni durante un despliegue.
dev-public-backup-init: dev-public-config-check
	$(PUBLIC_DEV_COMPOSE) up --build --detach --wait postgres
	$(PUBLIC_DEV_COMPOSE) exec -T --user postgres postgres pgbackrest --stanza=fasttourney-dev stanza-create
	$(PUBLIC_DEV_COMPOSE) exec -T --user postgres postgres pgbackrest --stanza=fasttourney-dev check
	$(PUBLIC_DEV_COMPOSE) exec -T --user postgres postgres pgbackrest --stanza=fasttourney-dev --type=full backup

# Copia completa semanal: domingo a las 03:45 mediante launchd.
dev-public-backup-full: dev-public-config-check
	$(PUBLIC_DEV_COMPOSE) exec -T --user postgres postgres pgbackrest --stanza=fasttourney-dev --type=full backup

# Incremental diario: lunes a sábado a las 03:45 mediante launchd.
dev-public-backup-incremental: dev-public-config-check
	$(PUBLIC_DEV_COMPOSE) exec -T --user postgres postgres pgbackrest --stanza=fasttourney-dev --type=incr backup

dev-public-backup-status: dev-public-config-check
	$(PUBLIC_DEV_COMPOSE) exec -T --user postgres postgres pgbackrest --stanza=fasttourney-dev info

# BACKUP es una etiqueta devuelta por `dev-public-backup-status`. La operación
# restaura solo en postgres-restore-data y no toca postgres-data.
dev-public-backup-restore-verify: dev-public-config-check
	@test -n "$(BACKUP)" || { echo "Uso: make dev-public-backup-restore-verify BACKUP=<etiqueta>"; exit 1; }
	$(PUBLIC_DEV_COMPOSE) run --rm --no-deps --entrypoint sh postgres -ec 'mkdir -p /restore; chown postgres:postgres /restore; exec gosu postgres pgbackrest --stanza=fasttourney-dev --pg1-path=/restore --set="$(BACKUP)" --delta restore'
	$(PUBLIC_DEV_COMPOSE) --profile backup-restore up --detach --wait postgres-restore
	$(PUBLIC_DEV_COMPOSE) exec -T postgres-restore sh -ec 'psql -U "$$POSTGRES_USER" -d "$$POSTGRES_DB" -v ON_ERROR_STOP=1 -c "SELECT current_database(), pg_is_in_recovery()"'
	$(PUBLIC_DEV_COMPOSE) --profile backup-restore rm --force --stop postgres-restore

# Arranca PostgreSQL y espera a que el health check confirme que acepta conexiones.
db-up: db-env-check
	$(POSTGRES_COMPOSE) up --detach --wait

db-wait: db-env-check
	$(POSTGRES_COMPOSE) up --detach --wait

# Detiene PostgreSQL conservando el volumen de datos local.
db-down: db-env-check
	$(POSTGRES_COMPOSE) down --remove-orphans

db-status: db-env-check
	$(POSTGRES_COMPOSE) ps

db-logs: db-env-check
	$(POSTGRES_COMPOSE) logs --tail=200 postgres

# Acción destructiva: borra el volumen de datos únicamente tras confirmación explícita.
db-reset: db-env-check
	@printf "Esto eliminará todos los datos PostgreSQL locales. Escribe RESET para continuar: "; read answer; \
	[ "$$answer" = "RESET" ] || { echo "db-reset: cancelado"; exit 1; }
	$(POSTGRES_COMPOSE) down --volumes --remove-orphans

# Aplica el esquema inicial explícitamente; no forma parte del arranque de la API.
db-schema-apply: db-env-check
	@sed '/^--/d' $(BACKEND_DIR)/db/schema/initial_schema.sql | \
		$(POSTGRES_COMPOSE) exec -T postgres sh -c 'psql -U "$$POSTGRES_USER" -d "$$POSTGRES_DB" -v ON_ERROR_STOP=1'
