.PHONY: docs build run clean

# Generate Swagger documentation
docs:
	swag init -g cmd/server/main.go

# Build the application
build:
	go build -o bin/s29-api cmd/server/main.go

# Run the application
run: docs
	go run cmd/server/main.go

# Run with live reload (requires air)
dev: docs
	air

# Clean generated files
clean:
	rm -rf docs/
	rm -f bin/s29-api

# Install development dependencies
install-deps:
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/cosmtrek/air@latest

# Generate docs and build
all: docs build
