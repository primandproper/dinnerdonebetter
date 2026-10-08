.PHONY: check test build lint lint-fix format format-check install deploy_prod

# check, build, test and lint each start with `npm run deps:check` (npm ls), which fails when
# what is installed in node_modules is not what package.json asks for: an install left over
# from before a version bump type-checks and builds against the wrong library. `npm ci` fixes it.

# Type-check both apps (svelte-check) and the session and recipes packages (tsc)
check:
	npm run check

# Build both apps for production
build:
	npm run build

# Run unit tests in both apps and the workspace packages
test:
	npm run test

# Lint both apps
lint:
	npm run lint

# Lint with auto-fix
lint-fix:
	npm run lint:fix

# Format all code
format:
	npm run format

# Check formatting only
format-check:
	npm run format:check

# Install dependencies (from frontend root)
install:
	npm install

# Deploy frontend (consumer + admin) to prod. Run from repo root: make deploy_prod_frontend
deploy_prod:
	$(MAKE) -C .. deploy_prod_frontend
