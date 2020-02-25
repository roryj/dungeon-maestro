.PHONY: deps clean lint build

lint:
	golangci-lint run ./...

deps:
	go get -u ./...

clean: 
	rm -rf ./maestro
	
build:
	GOOS=linux GOARCH=amd64 go build -o maestro ./
