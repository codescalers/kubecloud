
run:
	@CGO_ENABLED=1 go run . --config config.json

listen: 
	@CGO_ENABLED=1 go test -v ./tests/... -tags=example -count=1 -run ^TestClient_ListenToSSE$

cleanup:
	@CGO_ENABLED=1 go run cmd/cleanup/main.go --config config.json

fix-rentables:
	@CGO_ENABLED=1 go run cmd/fix_rentables/main.go --db ${db} --tfchain-url ${tfchain-url}

example: 
	@CGO_ENABLED=1 go test -v ./tests/... -tags=example -count=1 -run ^TestClient_${t}$

