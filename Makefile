.PHONY: deps clean format lint build

all: clean format lint build

lint:
	golangci-lint run ./...

format:
	go fmt

deps:
	go get -u ./...

clean: 
	rm -rf ./maestro
	
build:
	GOOS=linux GOARCH=amd64 go build -o maestro ./

build-cli:
	go build -o maestro ./cli/cli.go
