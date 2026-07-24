GO ?= go
BACKEND_DIR := apps/backend
TOOL_MODFILE := go.tool.mod
GO_PACKAGES := ./...
GO_BACKEND := $(GO) -C $(BACKEND_DIR)
GO_TOOL := $(GO_BACKEND) tool -modfile=$(TOOL_MODFILE)
GO_SOURCE := $(shell find $(BACKEND_DIR) -type f -name '*.go' -print -quit)

.PHONY: \
	format format-check \
	tidy tidy-check \
	tidy-tools tidy-tools-check tidy-all \
	lint test test-race build vuln \
	check verify

# Modifica los archivos Go mediante el goimports pineado.
format:
	$(GO_TOOL) goimports -w .

# Comprueba el formato sin modificar archivos.
format-check:
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

lint:
ifeq ($(strip $(GO_SOURCE)),)
	@echo "lint: omitido; todavía no existen paquetes Go"
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

# Feedback local rápido.
check: format-check lint test

# Verificación completa local y de CI.
verify: check tidy-check tidy-tools-check build vuln
