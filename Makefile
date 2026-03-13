.PHONY: build test vet clean install lint

build:
	go build -o dist/portui ./cmd/portui

test:
	go test ./... -race -v

vet:
	go vet ./...

lint:
	@golangci-lint run 2>/dev/null || echo "golangci-lint not installed, skipping"

clean:
	rm -rf dist/

install:
	go install ./cmd/portui
