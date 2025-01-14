BINARY_NAME=command-server
build:
	go build -o $(BINARY_NAME) *.go
run:
	go run main.go
clean:
	go clean
	rm -f $(BINARY_NAME)
test:
	go test ./...
.PHONY: build run clean test
