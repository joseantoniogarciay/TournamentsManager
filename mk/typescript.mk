# Toolchain TypeScript compartido del monorepo.

PNPM ?= pnpm

.PHONY: format-ts format-check-ts lint-ts typecheck

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
