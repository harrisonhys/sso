.PHONY: migrate-up migrate-down migrate-create migrate-force migrate-version

# Database connection
DB_URL=mysql://root:root@tcp(localhost:3306)/sso-main

# Migrate up
migrate-up:
	migrate -path db/migrations -database "$(DB_URL)" -verbose up

# Migrate down
migrate-down:
	migrate -path db/migrations -database "$(DB_URL)" -verbose down

# Create new migration
migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir db/migrations -seq $$name

# Force migration version
migrate-force:
	@read -p "Enter version: " version; \
	migrate -path db/migrations -database "$(DB_URL)" force $$version

# Check current migration version
migrate-version:
	migrate -path db/migrations -database "$(DB_URL)" version

# Run server
run:
	go run cmd/server/main.go

# Build server
build:
	go build -o bin/sso-server cmd/server/main.go

# Test
test:
	go test -v ./...

# Test with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

