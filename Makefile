.PHONY: build run test clean 
 
build: 
	@echo "Building server..." 
	go build -o bin/server ./cmd/server 
	@echo "Building cli..." 
	go build -o bin/cli ./cmd/cli 
 
run: 
	docker-compose up --build 

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.64.8

lint: install-lint-deps
	golangci-lint run ./...
 
test: lint
	go test -v ./... 
 
clean: 
	rm -rf bin/ 
	@echo "Cleaned" 
