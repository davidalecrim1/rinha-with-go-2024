lint:
	golangci-lint run

run-db:
	docker compose down 
	docker compose up --build -d postgres-db pgadmin-ui

make local-run:
	go run ./cmd/api/server.go