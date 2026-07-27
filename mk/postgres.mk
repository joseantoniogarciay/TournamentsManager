# PostgreSQL local: Docker Compose, ciclo de vida y migraciones explícitas.
# API Go y cliente Expo se ejecutan en el host durante desarrollo.

POSTGRES_LOCAL_DIR := infra/local
POSTGRES_LOCAL_ENV := $(POSTGRES_LOCAL_DIR)/.env
BACKEND_ENV := $(BACKEND_DIR)/.env
POSTGRES_COMPOSE := docker compose --env-file $(POSTGRES_LOCAL_ENV) -f $(POSTGRES_LOCAL_DIR)/compose.yaml

.PHONY: \
	db-init db-env-check db-backend-env-check \
	db-up db-wait db-down db-status db-logs db-reset db-migrate

# Crea los contratos locales sin sobrescribir una configuración ya existente.
db-init:
	@test ! -e $(POSTGRES_LOCAL_ENV) || { echo "db-init: $(POSTGRES_LOCAL_ENV) ya existe; no se sobrescribe"; exit 1; }
	@test ! -e $(BACKEND_ENV) || { echo "db-init: $(BACKEND_ENV) ya existe; no se sobrescribe"; exit 1; }
	cp $(POSTGRES_LOCAL_DIR)/.env.example $(POSTGRES_LOCAL_ENV)
	cp $(BACKEND_DIR)/.env.example $(BACKEND_ENV)
	@echo "db-init: contratos creados; revisa las contraseñas locales antes de arrancar"

db-env-check:
	@test -f $(POSTGRES_LOCAL_ENV) || { echo "Falta $(POSTGRES_LOCAL_ENV). Ejecuta: make db-init"; exit 1; }

db-backend-env-check:
	@test -f $(BACKEND_ENV) || { echo "Falta $(BACKEND_ENV). Ejecuta: make db-init"; exit 1; }

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

# Aplica migraciones explícitamente; no forma parte del arranque de la API.
db-migrate: db-env-check db-backend-env-check
	@if [ ! -d $(BACKEND_DIR)/db/migrations ] || \
		! find $(BACKEND_DIR)/db/migrations -maxdepth 1 -type f -name '*.sql' -print -quit | grep -q .; then \
		echo "db-migrate: omitido; todavía no existen migraciones"; \
	else \
		set -a; . $(BACKEND_ENV); set +a; \
		[ -n "$$DATABASE_URL" ] || { echo "DATABASE_URL no está definido en $(BACKEND_ENV)"; exit 1; }; \
		$(GO_TOOL) goose -dir db/migrations postgres "$$DATABASE_URL" up; \
	fi
