docker run --name postgres-test \
-e POSTGRES_PASSWORD=postgres \
-p 5433:5432 \
-d postgres

go run cmd/server/main.go