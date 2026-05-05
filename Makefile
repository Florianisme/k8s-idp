.PHONY: dev-backend dev-frontend build build-frontend docker-build docker-run test

# Run Go backend (requires frontend/dist to exist; run build-frontend first)
dev-backend:
	MODE=$${MODE:-kubernetes} PORT=8080 go run .

# Run Vite dev server (proxies /api to :8080)
dev-frontend:
	cd frontend && npm run dev

# Build everything: frontend first, then Go binary
build: build-frontend
	CGO_ENABLED=0 go build -o k8s-idp .

build-frontend:
	cd frontend && npm ci && npm run build

# Build Docker image
docker-build:
	docker build -t k8s-idp:latest .

# Run Docker image in docker mode (mounts socket)
docker-run:
	docker run --rm -p 8080:8080 \
		-e MODE=$${MODE:-docker} \
		-v /var/run/docker.sock:/var/run/docker.sock \
		k8s-idp:latest

test:
	go test ./...
