.PHONY: fmt test vet run

fmt:
	go fmt ./...

test:
	go test ./...

vet:
	go vet ./...

run:
	go run ./cmd/server
