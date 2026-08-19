BINARY := forth
SAMPLES := $(wildcard samples/*.fth)

.PHONY: all build test bench fmt vet clean bundles run

all: build

build:
	go build -o $(BINARY) .

test:
	go vet ./...
	gofmt -l .
	go test ./...

bench:
	go test -run XXX -bench . ./...

fmt:
	gofmt -w .

# Build a standalone binary for every sample into ./dist
bundles: build
	@mkdir -p dist
	@for f in $(SAMPLES); do \
		name=dist/$$(basename $$f .fth); \
		./$(BINARY) -bundle $$f -o $$name; \
	done

run: build
	./$(BINARY)

clean:
	rm -rf $(BINARY) dist
