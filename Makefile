
-include backend/.env

SHELL 		    := /bin/bash
MIGRATE         := migrate
MIGRATIONS_DIR  := backend/migrations
DB_URL          := postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(POSTGRESQL_PORT)/$(POSTGRES_DB)?sslmode=disable
MIGRATE_CMD     := $(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DB_URL)"
NAME	?= feature_default_name
FILES 			:= dto.go model.go handler.go service.go repository.go repository_gorm.go

init:
	@echo "Initializing the project..."
	@echo "Loading environment variables from backend/.env"
	@echo "Environment variables loaded."
	bash scripts/bootstrap.sh

.PHONY: gen-feature
gen-feature:
	@echo "Generating feature resources: $(NAME)"
	@mkdir -p backend/internal/features/$(NAME)
	@for file in $(FILES); do \
		echo "package $(NAME)" > backend/internal/features/$(NAME)/$$file; \
	done
	@echo "Created package $(NAME) files successfully."

dc-up:
	docker compose -f docker-compose.yaml up -d

dc-down:
	docker compose -f docker-compose.yaml down

dc-restart:
	docker compose -f docker-compose.yaml down
	docker compose -f docker-compose.yaml up -d

dc-logs:
	docker compose -f docker-compose.yaml logs -f

dc-ps:
	docker compose -f docker-compose.yaml ps

dc-build:
	docker-compose -f docker-compose.yaml build

run-be:
	cd backend && go run ./cmd/server/main.go

install-migrate:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

migrate-up:
	$(MIGRATE_CMD) up

migrate-down:
	$(MIGRATE_CMD) down 1

migrate-down-all:
	$(MIGRATE_CMD) down -all

migrate-status:
	$(MIGRATE_CMD) version

migrate-new:
	@if [ -z "$(name)" ]; then \
		read -p "Enter migration name (snake_case): " name; \
	else \
		name=$(name); \
	fi; \
	$(MIGRATE) create -ext sql -dir $(MIGRATIONS_DIR) -seq $$name