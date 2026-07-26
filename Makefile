BACKEND_DIR := apps/backend
MAKEFILE_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

# Cada fragmento agrupa una tecnología; este fichero conserva los comandos que
# combinan ecosistemas y es la entrada pública del monorepo.
include $(MAKEFILE_DIR)/mk/go.mk
include $(MAKEFILE_DIR)/mk/typescript.mk
include $(MAKEFILE_DIR)/mk/postgres.mk

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

# Comprueba todo el formato sin modificar archivos.
format-check: format-check-go format-check-ts

lint: lint-go lint-ts

# Feedback local rápido para todos los ecosistemas activos.
check: format-check lint typecheck test

# Verificación completa local y de CI.
verify: check tidy-check tidy-tools-check build vuln
