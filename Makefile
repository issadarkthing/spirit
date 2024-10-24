VERSION="`git describe --abbrev=0 --tags`"
COMMIT="`git rev-list -1 --abbrev-commit HEAD`"

all: clean fmt test build

fmt:
	@echo "Formatting..."
	@goimports -l -w ./

install:
	mkdir -p ~/.local/lib/spirit
	mkdir -p ~/.local/bin
	cp ./bin/spirit ~/.local/bin/spirit && cp ./lib/core.st ~/.local/lib/spirit

clean:
	@echo "Cleaning up..."
	@rm -rf ./bin
	@go mod tidy -v

test: build-only
	@echo "Running tests..."
	@go test -cover ./...
	@bin/spirit -u -p ./lib/core.st lib/core_test.st

test-verbose:
	@echo "Running tests..."
	@go test -v -cover ./...

benchmark:
	@echo "Running benchmarks..."
	@go test -benchmem -run="none" -benchmem -bench="Benchmark.*" -v ./...

build-only:
	@go build -ldflags="-X main.version=${VERSION} -X main.commit=${COMMIT}" -o ./bin/spirit ./cmd/spirit/

build-small:
	@go build -ldflags="-w -X main.version=${VERSION} -X main.commit=${COMMIT}" -o ./bin/spirit ./cmd/spirit/
	@upx bin/spirit

build: test
	@mkdir -p ./bin
	@go build -ldflags="-X main.version=${VERSION} -X main.commit=${COMMIT}" -o ./bin/spirit ./cmd/spirit/

run: 
	@./bin/spirit -u -p ./lib/core.st ./sample/sample.st

repl:
	@rlwrap bin/spirit -u -p ./lib/core.st
