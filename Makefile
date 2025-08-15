.PHONY: build test
CMD=cmd/ngxstat/main.go
build:
	@mkdir -p bin
	go build -o bin/$(BIN) $(CMD)

test:
	go test -v -race ./...
