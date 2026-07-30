# Toolchain Go del backend: herramientas, análisis y construcción.

GO ?= go
TOOL_MODFILE := go.tool.mod
GO_PACKAGES := ./...
GO_BACKEND := $(GO) -C $(BACKEND_DIR)
GO_TOOL := $(GO_BACKEND) tool -modfile=$(TOOL_MODFILE)
GO_SOURCE := $(shell find $(BACKEND_DIR) -type f -name '*.go' -print -quit)

.PHONY: \
	api-up local-api-up \
	format-go format-check-go \
	tidy tidy-check tidy-tools tidy-tools-check tidy-all \
	lint-go test test-race build vuln sqlc-generate sqlc-generate-check

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

# Arranca las dependencias locales y la API Go en el host. Las migraciones se
# mantienen como un paso explícito mediante `make db-migrate`.
api-up: local-api-up

local-api-up: local-config-check db-up
	@set -a; . $(BACKEND_ENV); set +a; \
	$(GO_BACKEND) run ./cmd/api

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
