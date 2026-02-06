.PHONY: lint
lint:
	golangci-lint run

.PHONY: fmt
fmt:
	golangci-lint fmt

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: test
test:
	go test -coverpkg=./... -coverprofile=coverage.out -race -timeout 5m ./...

.PHONY: check
check: generate lint test

.PHONY: generate
generate:
	go generate ./...

.PHONY: clean
clean:
	rm -f coverage.out
