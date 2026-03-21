.PHONY: dev dev-up dev-down dev-backend dev-frontend test build lint swagger cli

dev-up:
	docker compose -f deployments/docker-compose.yml up -d

dev-down:
	docker compose -f deployments/docker-compose.yml down

dev-backend:
	cd backend && air

dev-frontend:
	cd frontend && npm run dev

dev: dev-up
	@echo "Services started. Run 'make dev-backend' and 'make dev-frontend' in separate terminals."

test:
	cd backend && go test -race -count=1 ./...
	cd frontend && npm test -- --run

build:
	cd backend && go build -o bin/server ./cmd/server
	cd frontend && npm run build

lint:
	cd backend && golangci-lint run ./...
	cd frontend && npx biome check ./src

cli:
	cd cli && go build -o bin/hub-cli .

swagger:
	cd backend && swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal
