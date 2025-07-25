.PHONY: build test test-go test-crystal integration-test clean

# Build the plugin
build:
	go build -o bin/sqlc-gen-crystal

# Run Go unit tests
test-go:
	go test ./internal/crystal/... -v

# Generate Crystal code for integration tests
generate-integration: build
	cd test/integration && sqlc generate

# Run Crystal integration tests
test-crystal: generate-integration
	cd test/integration && shards install && crystal spec

# Generate examples
generate-examples: build
	cd examples/booktest && sqlc generate
	cd examples/authors_books && sqlc generate

# Run all tests
test: test-go test-crystal

# Full integration test (includes building, generating, and testing)
integration-test: build
	@echo "=== Building Plugin ==="
	go build -o bin/sqlc-gen-crystal
	@echo "\n=== Generating Crystal Code ==="
	cd test/integration && sqlc generate
	@echo "\n=== Installing Crystal Dependencies ==="
	cd test/integration && shards install
	@echo "\n=== Running Crystal Integration Tests ==="
	cd test/integration && crystal spec --verbose

# Clean generated files and binaries
clean:
	rm -f bin/sqlc-gen-crystal
	rm -rf test/integration/src/db
	rm -rf test/integration/lib
	rm -f test/integration/shard.lock
	rm -rf examples/booktest/src/db

# Install sqlc (if needed)
install-sqlc:
	go install github.com/sqlc-dev/sqlc/v2/cmd/sqlc@latest

# Format Go code
fmt:
	go fmt ./...

# Run Go linter
lint:
	golangci-lint run

# Build for all platforms
release: build
	GOOS=darwin GOARCH=amd64 go build -o bin/sqlc-gen-crystal-darwin-amd64
	GOOS=darwin GOARCH=arm64 go build -o bin/sqlc-gen-crystal-darwin-arm64
	GOOS=linux GOARCH=amd64 go build -o bin/sqlc-gen-crystal-linux-amd64
	GOOS=linux GOARCH=arm64 go build -o bin/sqlc-gen-crystal-linux-arm64
	GOOS=windows GOARCH=amd64 go build -o bin/sqlc-gen-crystal-windows-amd64.exe