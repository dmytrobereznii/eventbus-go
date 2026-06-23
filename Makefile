.DEFAULT_GOAL := run

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

run: vet
	go run ./cmd

test:
	go test -race -v ./... -timeout 10s