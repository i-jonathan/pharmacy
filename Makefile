MIGRATIONS_DIR=./backend/db/migrations
MIGRATION_NAME?=new_migration

FRONTEND_DIR=frontend
BACKEND_DIR=backend
GO_LINE_MAX_LEN?=100

create-migration:
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(MIGRATION_NAME)

migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up

rollback-migration:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down 1

clean-dirty-migration:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" force $(VERSION)

# Tailwind for Go HTML templates
tailwind:
	npx tailwindcss -i backend/template/static/css/input.css -o ./backend/template/static/css/tailwind.css --minify

# Vue build
build-frontend:
	cd $(FRONTEND_DIR) && npm run build

# All UI assets
assets: tailwind build-frontend

run:
	cd $(BACKEND_DIR) && go run .

dev: assets run

# Install with: go install github.com/golangci/golines@latest
# Show proposed line wrapping without changing files.
check-go-lines:
	golines --dry-run --max-len=$(GO_LINE_MAX_LEN) $(BACKEND_DIR)

# Apply golines formatting to Go sources.
format-go-lines:
	golines -w --max-len=$(GO_LINE_MAX_LEN) $(BACKEND_DIR)

# Check Go's standard formatting without modifying files.
check-go-format:
	@unformatted="$$(gofmt -l $$(find $(BACKEND_DIR) -type f -name '*.go'))"; \
	if [ -n "$$unformatted" ]; then \
		printf 'Go files need gofmt:\n%s\n' "$$unformatted"; \
		exit 1; \
	fi; \
	printf 'All Go files are gofmt-clean.\n'

# Apply Go's standard formatting to all backend Go files.
format-go:
	gofmt -w $$(find $(BACKEND_DIR) -type f -name '*.go')

.PHONY: \
	create-migration \
	migrate-up \
	rollback-migration \
	clean-dirty-migration \
	tailwind \
	build-frontend \
	assets \
	run \
	dev \
	check-go-lines \
	format-go-lines \
	check-go-format \
	format-go
