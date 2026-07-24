GO ?= go
BACKEND_DIR := apps/backend
TOOL_MODFILE := go.tool.mod
GO_PACKAGES := ./...
GO_BACKEND := $(GO) -C $(BACKEND_DIR)
GO_TOOL := $(GO_BACKEND) tool -modfile=$(TOOL_MODFILE)
GO_SOURCE := $(shell find $(BACKEND_DIR) -type f -name '*.go' -print -quit)
PNPM ?= pnpm

.PHONY: \
	format format-go format-ts \
	format-check format-check-go format-check-ts \
	tidy tidy-check \
	tidy-tools tidy-tools-check tidy-all \
	lint lint-go lint-ts typecheck \
	test test-race build vuln \
	check verify

# Modifica todos los archivos soportados por los formateadores pineados.
format: format-go format-ts

format-go:
	$(GO_TOOL) goimports -w .

format-ts:
	$(PNPM) run format

# Comprueba todo el formato sin modificar archivos.
format-check: format-check-go format-check-ts

format-check-go:
	@unformatted="$$($(GO_TOOL) goimports -l .)" || exit $$?; \
	if [ -n "$$unformatted" ]; then \
		echo "Los siguientes archivos necesitan formato:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

format-check-ts:
	$(PNPM) run format:check

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

lint: lint-go lint-ts

lint-go:
ifeq ($(strip $(GO_SOURCE)),)
	@echo "lint-go: omitido; todavía no existen paquetes Go"
else
	$(GO_TOOL) golangci-lint run $(GO_PACKAGES)
endif

lint-ts:
	$(PNPM) run lint

typecheck:
	$(PNPM) run typecheck

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

# Feedback local rápido para todos los ecosistemas activos.
check: format-check lint typecheck test

# Verificación completa local y de CI.
verify: check tidy-check tidy-tools-check build vuln
