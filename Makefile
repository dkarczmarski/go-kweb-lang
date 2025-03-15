.PHONY: test

go-mockgen-install:
	go install go.uber.org/mock/mockgen@latest

go-generate:
	go generate ./...

test:
	go test -v ./...

dev-install-lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

dev-lint:
	golangci-lint run

dev-lint-all:
	golangci-lint run --enable-all --config=/dev/null

dev-goimports:
	goimports -w .

dev-test-docker:
	docker build -t kweb-test .

dev-test-run:
	docker run --rm -it -v ./cache:/app/cache \
		-v kweb-repo:/app/kubernetes-website \
		-e CACHE_DIR=/app/cache \
		-e REPO_DIR=/app/kubernetes-website \
		-e ALLOWED_LANGS=pl \
		-p 127.0.0.1:8080:8080 \
		kweb-test
