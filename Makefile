.PHONY: all test

all:
	CGO_ENABLED=0 go build

test:
	go test -v -race ./...
