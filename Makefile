
-include backend/.env


init:
	@echo "Initializing the project..."
	@echo "Loading environment variables from backend/.env"
	@echo "Environment variables loaded."
	bash scripts/bootstrap.sh

run:
	cd backend && go run ./cmd/server/main.go