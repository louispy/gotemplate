APP_NAME=template

run-http:
	@echo "Running HTTP Server..."
	@go run ./cmd/httpserver

run-migration:
	@echo "Running migration..."
	@go run ./cmd/migrate

test:
	@go test -v ./...
