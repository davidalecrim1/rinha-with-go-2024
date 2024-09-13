lint:
	golangci-lint run

run-db:
	docker compose -f docker-compose.local.yml down 
	docker compose -f docker-compose.local.yml up --build -d postgres-db pgadmin-ui

run:
	docker compose -f docker-compose.local.yml  up --build -d

restart:
	docker compose -f docker-compose.local.yml down
	docker compose -f docker-compose.local.yml up --build -d

stop:
	docker compose -f docker-compose.local.yml down