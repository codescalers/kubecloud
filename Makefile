run:
	@make backend-run &
	@make frontend-run

backend-run: backend/config.json 
	@cd backend && CGO_ENABLED=1 go run . --config config.json

backend/config.json: config.json
	@cp config.json backend/config.json
	@touch backend/config.json

frontend-run:frontend/kubecloud/.env
	@cd frontend/kubecloud && [ -d node_modules ] || npm install 
	@cd frontend/kubecloud && npm run dev

frontend/kubecloud/.env:
	@cp frontendconfig.env frontend/kubecloud/.env
	@touch frontend/kubecloud/.env