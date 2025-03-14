.PHONY: test

go-mockgen-install:
	go install go.uber.org/mock/mockgen@latest

go-generate:
	go generate ./...

test:
	go test -v ./...
