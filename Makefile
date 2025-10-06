run:
	@cd backend; \
	echo "Building backend..."; \
	CGO_ENABLED=1 go build -o myceliumcloud .; \
	./myceliumcloud --config config.json & \
	pid=$$!; \
	echo "Backend started at $$pid"; \
	sleep 1; \
	if ps -p $$pid > /dev/null; then \
		cd ../frontend/kubecloud; \
		[ -d node_modules ] || npm install; \
		trap "echo \"Frontend failed to start, Stopping backend...\"; kill $$pid; rm -f ../../backend/myceliumcloud" EXIT; \
		npm run dev; \
	else \
		echo "Backend failed to start"; \
	fi



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