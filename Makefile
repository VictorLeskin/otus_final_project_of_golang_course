.PHONY: build run test clean 
 
build: 
	@echo "Building server..." 
	go build -o bin/server ./cmd/server 
	@echo "Building cli..." 
	go build -o bin/cli ./cmd/cli 
 
run: 
	docker-compose up --build 
 
test: 
	go test -v ./... 
 
clean: 
	rm -rf bin/ 
	@echo "Cleaned" 
