MIGRATIONS_DIR=./db/migrations
MIGRATION_NAME?=new_migration

create-migration:
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(MIGRATION_NAME)

.PHONY: create-migrations