run:
	@cd backend; \
	echo "Building backend..."; \
	CGO_ENABLED=1 go build -o myceliumcloud .; \
	./myceliumcloud --config config.json & \
	pid=$$!; \
	echo "Backend started at $$pid"; \
	if ps -p $$pid > /dev/null; then \
		trap 'if [ -n "$$pid" ]; then echo "Stopping backend..."; kill -9 $$pid 2>/dev/null; rm -f ../../backend/myceliumcloud; pid=""; fi' INT EXIT;\
		cd ../frontend/kubecloud; \
		[ -d node_modules ] || npm install; \
		npm run dev; \
	else \
		echo "Backend failed to start"; \
		rm -f ../../backend/myceliumcloud; \
	fi

run-compose:
	@docker compose build
	@docker compose up -d


backend-run: backend/config.json 
	@cd backend && CGO_ENABLED=1 go run . --config config.json

backend/config.json: config.json
	@cp config.json backend/config.json

frontend-run:frontend/kubecloud/.env
	@cd frontend/kubecloud && [ -d node_modules ] || npm install 
	@cd frontend/kubecloud && npm run dev

frontend/kubecloud/.env:
	@cp frontendconfig.env frontend/kubecloud/.env