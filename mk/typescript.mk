# Toolchain TypeScript compartido del monorepo.

PNPM ?= pnpm

.PHONY: format-ts format-check-ts lint-ts typecheck openapi-lint openapi-generate openapi-generate-check openapi-ui

# Modifica los archivos TypeScript y de configuración con Prettier.
format-ts:
	$(PNPM) run format

# Comprueba el formato TypeScript sin modificar archivos.
format-check-ts:
	$(PNPM) run format:check

lint-ts:
	$(PNPM) run lint

typecheck:
	$(PNPM) run typecheck

openapi-lint:
	$(PNPM) run openapi:lint

openapi-generate:
	$(PNPM) run openapi:generate

openapi-generate-check:
	$(PNPM) run openapi:generate:check

# Sirve Swagger UI exclusivamente en loopback para explorar el contrato local.
openapi-ui:
	$(PNPM) run openapi:ui
