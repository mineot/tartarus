BINARY  := tartarus
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -X tartarus/helpers.version=$(VERSION)

.PHONY: build run test vet fmt clean

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) .

run:
	go run .

test:
	go test -race ./...

vet:
	go vet ./...

fmt:
	gofmt -l .

clean:
	rm -f $(BINARY)