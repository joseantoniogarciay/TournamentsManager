# PostgreSQL local: Docker Compose, ciclo de vida y esquema inicial explícito.
# API Go y cliente Expo se ejecutan en el host durante desarrollo.

POSTGRES_LOCAL_DIR := infra/local
POSTGRES_LOCAL_ENV := $(POSTGRES_LOCAL_DIR)/.env
BACKEND_ENV := $(BACKEND_DIR)/.env
DEV_API_ENV := $(POSTGRES_LOCAL_DIR)/api.docker.env
POSTGRES_COMPOSE := docker compose --env-file $(POSTGRES_LOCAL_ENV) -f $(POSTGRES_LOCAL_DIR)/compose.yaml
DEV_COMPOSE := docker compose --env-file $(POSTGRES_LOCAL_ENV) -f $(POSTGRES_LOCAL_DIR)/compose.dev.yaml

.PHONY: \
	db-init dev-init db-env-check db-backend-env-check dev-api-env-check local-config-check dev-config-check \
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
