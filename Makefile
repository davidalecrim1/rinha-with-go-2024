lint:
	golangci-lint run

testing:
	go test ./... -coverprofile=coverage.out -race

testing-view:
	go tool cover -html=coverage.out


local-run-db:
	docker compose -f docker-compose.local.yml down 
	docker compose -f docker-compose.local.yml up --build -d postgres-db pgadmin-ui

local-run:
	docker compose -f docker-compose.local.yml  up --build -d

local-restart:
	docker compose -f docker-compose.local.yml down
	docker compose -f docker-compose.local.yml up --build -d

local-stop:
	docker compose -f docker-compose.local.yml down


run:
	docker compose -f docker-compose.yml  up --build -d

restart:
	docker compose -f docker-compose.yml down
	docker compose -f docker-compose.yml up --build -d

stop:
	docker compose -f docker-compose.yml down