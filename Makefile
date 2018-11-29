.PHONY: deps clean build

deps:
	go get -u ./...

clean: 
	rm -rf ./maestro
	
build:
	GOOS=linux GOARCH=amd64 go build -o maestro ./src/go.roryj.dnd
