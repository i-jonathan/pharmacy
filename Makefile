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

tailwind:
	npx tailwindcss -i backend/template/static/css/input.css -o ./backend/template/static/css/tailwind.css --minify

build-frontend:
	cd $(FRONTEND_DIR) && npm run build

run:
	cd backend && go run .

dev: build-frontend run

.PHONY: create-migrations migrate-up
