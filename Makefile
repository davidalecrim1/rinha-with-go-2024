lint:
	golangci-lint run

unit-test:
	go test ./... -coverprofile=./test/results/unit-test-coverage.out -race

integration-test:
	go test -tags=integration ./... -coverprofile=./test/results/integration-test-coverage.out -race 

view-unit-test-coverage:
	go tool cover -html=./test/results/unit-test-coverage.out

view-integration-test-coverage:
	go tool cover -html=./test/results/integration-test-coverage.out

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