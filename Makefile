fmt:
	@echo "Running go fmt..."
	@go fmt ./...

test:
	@echo "Running go test..."
	@go test ./... -v

lint:
	@echo "Running go lint..."
	@go golangci-lint run

build:
	@echo "Building the project..."
	@go build -o mirakurun_exporter ./main.go