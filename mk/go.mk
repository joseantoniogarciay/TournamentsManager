# Toolchain Go del backend: herramientas, análisis y construcción.

GO ?= go
TOOL_MODFILE := go.tool.mod
GO_PACKAGES := ./...
GO_BACKEND := $(GO) -C $(BACKEND_DIR)
GO_TOOL := $(GO_BACKEND) tool -modfile=$(TOOL_MODFILE)
GO_SOURCE := $(shell find $(BACKEND_DIR) -type f -name '*.go' -print -quit)

.PHONY: \
	api-up local-api-up dev-up dev-down dev-logs dev-public-up dev-public-down dev-public-logs api-image-build \
	format-go format-check-go \
	tidy tidy-check tidy-tools tidy-tools-check tidy-all \
	lint-go test test-integration test-race build vuln sqlc-generate sqlc-generate-check

# Modifica todos los archivos Go con el formateador pineado.
format-go:
	$(GO_TOOL) goimports -w .

# Comprueba el formato Go sin modificar archivos.
format-check-go:
	@unformatted="$$($(GO_TOOL) goimports -l .)" || exit $$?; \
	if [ -n "$$unformatted" ]; then \
		echo "Los siguientes archivos necesitan formato:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

# Modifica go.mod y go.sum.
tidy:
	$(GO_BACKEND) mod tidy

# Comprueba go.mod y go.sum sin modificarlos.
tidy-check:
	$(GO_BACKEND) mod tidy -diff

# Modifica go.tool.mod y go.tool.sum.
tidy-tools:
	$(GO_BACKEND) mod tidy -modfile=$(TOOL_MODFILE)

# Comprueba go.tool.mod y go.tool.sum sin modificarlos.
tidy-tools-check:
	$(GO_BACKEND) mod tidy -modfile=$(TOOL_MODFILE) -diff

# Modifica y limpia ambos grafos de módulos.
tidy-all: tidy tidy-tools

lint-go:
ifeq ($(strip $(GO_SOURCE)),)
	@echo "lint-go: omitido; todavía no existen paquetes Go"
else
	$(GO_TOOL) golangci-lint run $(GO_PACKAGES)
endif

test:
ifeq ($(strip $(GO_SOURCE)),)
	@echo "test: omitido; todavía no existen paquetes Go"
else
	$(GO_BACKEND) test $(GO_PACKAGES)
endif

# Ejecuta persistencia real solo contra una base aislada proporcionada de forma explícita.
# La CI usa un rol PostgreSQL efímero, no los roles de despliegue. Conservamos
# el DDL de cada Up y omitimos solo su configuración de privilegios.
test-integration:
	@if [ -z "$$TM_INTEGRATION_DATABASE_URL" ]; then \
		echo "test-integration: omitido; TM_INTEGRATION_DATABASE_URL no está definido"; \
	else \
		psql "$$TM_INTEGRATION_DATABASE_URL" -v ON_ERROR_STOP=1 -f $(BACKEND_DIR)/db/schema/initial_schema.sql && \
		for migration in $(BACKEND_DIR)/db/migrations/*.sql; do \
			sed -n '/^-- +goose Up$$/,/^-- +goose Down$$/p' "$$migration" | \
				sed '/^-- +goose /d; /^SET ROLE tournaments_manager_dev_schema_owner;$$/d; /^GRANT .* TO tournaments_manager_dev_app;$$/d' | \
				psql "$$TM_INTEGRATION_DATABASE_URL" -v ON_ERROR_STOP=1 || exit $$?; \
		done && \
		TM_INTEGRATION_DATABASE_URL="$$TM_INTEGRATION_DATABASE_URL" TM_RUN_INTEGRATION=1 $(GO_BACKEND) test ./internal/adapters/postgres -run Integration -count=1; \
	fi

test-race:
ifeq ($(strip $(GO_SOURCE)),)
	@echo "test-race: omitido; todavía no existen paquetes Go"
else
	$(GO_BACKEND) test -race $(GO_PACKAGES)
endif

build:
ifeq ($(strip $(GO_SOURCE)),)
	@echo "build: omitido; todavía no existen paquetes Go"
else
	$(GO_BACKEND) build $(GO_PACKAGES)
endif

vuln:
ifeq ($(strip $(GO_SOURCE)),)
	@echo "vuln: omitido; todavía no existen paquetes Go"
else
	$(GO_TOOL) govulncheck $(GO_PACKAGES)
endif

# Arranca las dependencias locales y la API Go en el host. El esquema inicial se
# aplica explícitamente con `make db-schema-apply`.
api-up: local-api-up

local-api-up: local-config-check db-up
	@set -a; . $(BACKEND_ENV); set +a; \
	$(GO_BACKEND) run ./cmd/api

# Entorno diario: dependencias y API en Compose; Air recompila la API al guardar.
dev-up: dev-config-check
	$(DEV_COMPOSE) up --build --remove-orphans

dev-down: dev-config-check
	$(DEV_COMPOSE) down --remove-orphans

dev-logs: dev-config-check
	$(DEV_COMPOSE) logs --tail=200

# Construye solo la etapa sin compilador ni Air que ejecutará un runtime futuro.
api-image-build:
	docker build --target runtime --tag tournaments-manager-api:local $(BACKEND_DIR)

# sqlc solo se ejecuta al existir al menos una consulta del producto. Así se evita
# generar código de relleno antes de que un caso de uso lo necesite.
sqlc-generate:
	@if ! find $(BACKEND_DIR)/db/queries -type f -name '*.sql' -print -quit | grep -q .; then \
		echo "sqlc-generate: omitido; todavía no existen consultas"; \
	else \
		$(GO_TOOL) sqlc generate; \
	fi

sqlc-generate-check: sqlc-generate
	@git diff --exit-code -- $(BACKEND_DIR)/internal/adapters/postgres/sqlc
