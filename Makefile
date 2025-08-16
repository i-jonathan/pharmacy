MIGRATIONS_DIR=./db/migrations
MIGRATION_NAME?=new_migration

create-migration:
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(MIGRATION_NAME)
	
migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up

rollback-migration:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down 1
	
clean-dirty-migration:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" force $(VERSION)

tailwind:
	npx tailwindcss -i template/static/css/input.css -o ./template/static/css/tailwind.css --minify
	
.PHONY: create-migrations migrate-up