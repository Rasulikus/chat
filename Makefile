include .env
export

DB_DSN = postgres://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable
MIGRATE = migrate -path migrations -database "$(DB_DSN)"

.PHONY: migrateup migratedown migrateforce

migrateup:
	$(MIGRATE) up

migratedown:
	$(MIGRATE) down
