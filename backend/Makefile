
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

start-db: 
	@docker run -d \
		--name kcdb \
		-v pgdata:/var/lib/postgresql/data \
		-e POSTGRES_PASSWORD=password \
		-p 5432:5432 \
		postgres
