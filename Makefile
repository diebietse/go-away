test:
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

vendor:
	go mod tidy
	go mod vendor

build:
	CGO=0 go build -mod=vendor ./cmd/pushserver

.PHONY: test vendor build