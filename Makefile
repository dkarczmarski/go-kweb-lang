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

dev-test-build:
	go build -o go-kweb-lang .

dev-test-run:
	ALLOWED_LANGS=pl ./go-kweb-lang --once

dev-test-run-once:
	./go-kweb-lang --once

dev-test-run-interval:
	./go-kweb-lang --interval 1

dev-test-docker-build:
	docker build -t kweb-test .

dev-test-docker-vol-create:
	docker volume create kweb-repo
	docker volume create kweb-cache

dev-test-docker-vol-rm-repo:
	docker volume rm kweb-repo

dev-test-docker-vol-rm-cache:
	docker volume rm kweb-cache

dev-test-docker-run:
	docker run --rm -it \
		-v kweb-cache:/app/cache \
		-v kweb-repo:/app/kubernetes-website \
		-e CACHE_DIR=/app/cache \
		-e REPO_DIR=/app/kubernetes-website \
		-e ALLOWED_LANGS=pl \
		-p 127.0.0.1:8080:8080 \
		kweb-test \
		/app/go-kweb-lang --once
