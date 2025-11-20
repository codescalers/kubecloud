

.PHONY: run listen cleanup fix-rentables example migrate-db update-swagger update-postman-collection update-docs check-swag-deps check-postman-deps build test test-unit test-verbose test-coverage start-db

run:
	@CGO_ENABLED=1 go run . --config config.json

listen: 
	@CGO_ENABLED=1 go test -v ./tests/... -tags=example -count=1 -run ^TestClient_ListenToSSE$

cleanup:
	@CGO_ENABLED=1 go run cmd/cleanup/main.go --config config.json

fix-rentables:
	@CGO_ENABLED=1 go run cmd/fix_rentables/main.go --dsn ${dsn} --tfchain-url ${tfchain-url}

example: 
	@CGO_ENABLED=1 go test -v ./tests/... -tags=example -count=1 -run ^TestClient_${t}$

migrate-db:
	@CGO_ENABLED=1 go run cmd/migrate_db/main.go --source-db ${source-db} --destination-db ${destination-db}
	
update-swagger: check-swag-deps
	@swag init -g main.go -o docs/swagger --parseDependency --parseInternal

update-postman-collection: docs/swagger/swagger.yaml check-postman-deps
	@npx openapi-to-postmanv2 -s docs/swagger/swagger.yaml -o docs/postman/myceliumcloud_collection.json -p

update-docs: update-swagger update-postman-collection

check-swag-deps:
	@which swag > /dev/null || (echo "Error: swag not found. Please install with: go install github.com/swaggo/swag/cmd/swag@latest" && exit 1)

check-postman-deps:
	@which npx > /dev/null || (echo "Error: npx not found. Please install Node.js and npm" && exit 1)
	@npx openapi-to-postmanv2 --version > /dev/null 2>&1 || npm install -g openapi-to-postmanv2

build:
	@CGO_ENABLED=1 go build -o bin/app .

# Test targets
test:
	@go test ./...

test-unit:
	@go test ./internal/core/services -v

test-verbose:
	@go test ./... -v

test-coverage:
	@go test ./... -cover -coverprofile=coverage.out && go tool cover -html=coverage.out -o coverage.html && echo "Coverage report generated: coverage.html"

test-services:
	@go test ./internal/core/services -v

test-handlers:
	@go test ./internal/api/handlers -v 2>&1 | grep -E "(PASS|FAIL|RUN)" || true

test-quick:
	@go test ./internal/core/services ./internal/core/models ./internal/auth -v

start-db:
	@docker run -d \
		--name kcdb \
		-v pgdata:/var/lib/postgresql/data \
		-e POSTGRES_PASSWORD=password \
		-p 5432:5432 \
		postgres

help:
	@echo "Available targets:"
	@echo "  run                      - Run the application"
	@echo "  build                    - Build the application"
	@echo "  test                     - Run all tests"
	@echo "  test-unit                - Run unit tests (services)"
	@echo "  test-verbose             - Run all tests with verbose output"
	@echo "  test-coverage            - Run tests and generate coverage report"
	@echo "  test-services            - Run service tests"
	@echo "  test-handlers            - Run handler tests (may require TEST_MNEMONIC)"
	@echo "  test-quick               - Run quick tests (services, models, auth)"
	@echo "  cleanup                  - Run cleanup command"
	@echo "  fix-rentables            - Run fix-rentables command"
	@echo "  migrate-db               - Run database migration"
	@echo "  update-docs              - Update Swagger and Postman documentation"
	@echo "  listen                   - Listen to SSE events test"
	@echo "  example                  - Run example test"
	@echo "  start-db                 - Start PostgreSQL database in Docker"
	@echo ""
