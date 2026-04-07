# all is the default target
all: test

# help prints a help screen
help:
	@echo "build     - go build"
	@echo "install   - go install"
	@echo "test      - go test"
	@echo "gofmt     - go fmt"
	@echo "linux     - go build linux/amd64"
	@echo "clean     - remove temp files"
	@echo "clean-all - execute all clean commands"

# build compiles fabio and the test dependencies
.PHONY: build
build: gofmt
	go build

# test builds and runs the tests
.PHONY: test
test: build
	go test -v -test.timeout 15s ./...

# mod performs go module maintenance
.PHONY: mod
mod:
	go mod tidy

# gofmt runs gofmt on the code
.PHONY: gofmt
gofmt:
	gofmt -s -w `find . -type f -name '*.go'`

.PHONY: lint
lint:
	golangci-lint run ./...
