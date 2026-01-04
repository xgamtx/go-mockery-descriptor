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
	go test -coverpkg=./... -coverprofile=coverage.out -race -timeout 10s ./...

.PHONY: check
check: generate lint test

.PHONY: generate
generate:
	go generate ./pkg/...

FILES_TO_DELETE = 'mock_*.go' '*.gen.go'
.PHONY: clean
clean:
	rm -f coverage.out deadcode.out
	$(foreach file, $(FILES_TO_DELETE), find pkg -type f -name $(file) -delete;)
