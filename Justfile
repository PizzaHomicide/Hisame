# Set default recipe to list available commands
default:
    @just --list

# Build the Go project
build:
    go build -o bin/hisame ./cmd/hisame

# Run the application
run:
    HISAME_CONFIG_PATH=./test-config.yml HISAME_CONFIG_LOGGING_FILE_PATH=./hisame.log go run ./cmd/hisame

# Run all tests
test:
    go test -v ./...

# Format code
fmt:
    gofmt -s -w .

# Tidy go.mod file
tidy:
    go mod tidy

# Show project dependencies
deps:
    go list -m all

# Update dependencies
update-deps:
    go get -u ./...
    go mod tidy

# Clean build artifacts
clean:
    rm -rf bin/
    go clean
