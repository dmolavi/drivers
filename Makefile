.PHONY: test
test:
	go test -coverprofile=coverage.txt -race ./...

.PHONY:build
build:
	go build ./...

.PHONY: imports
imports:
	goimports -w -local "github.com/dmolavi" ./

.PHONY: fmt
fmt:
	gofmt -w -s ./
