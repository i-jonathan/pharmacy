MIGRATIONS_DIR=./backend/db/migrations
MIGRATION_NAME?=new_migration

FRONTEND_DIR=frontend
BACKEND_DIR=backend

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

.PHONY: \
	create-migration \
	migrate-up \
	rollback-migration \
	clean-dirty-migration \
	tailwind \
	build-frontend \
	assets \
	run \
	dev
