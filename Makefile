.PHONY: build test clean install coverage coverage-html

BINARY_NAME=Gonana
BUILD_DIR=build

build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/gonana

test:
	go test -v ./...

coverage:
	go test -coverprofile=coverage.out ./...
	@echo "\n=== Coverage Summary ==="
	@go tool cover -func=coverage.out | grep total
	@echo "\nGenerate HTML report with: make coverage-html"

coverage-html: coverage
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

install: build
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/

lint:
	golangci-lint run

format:
	go fmt ./...