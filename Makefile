.PHONY: test run tidy quality swagger

COVERAGE_MIN ?= 25

test:
	go test -v ./...

quality:
	go test -race -covermode=atomic -coverpkg=./... -coverprofile=coverage.out ./...
	@total=$$(go tool cover -func=coverage.out | awk '/^total:/ {gsub(/%/,"",$$3); print $$3}'); \
		echo "Total coverage: $$total%"; \
		awk "BEGIN {exit !($$total >= $(COVERAGE_MIN))}" || (echo "Coverage threshold $(COVERAGE_MIN)% not met" && exit 1)

run:
	go run cmd/api/main.go

tidy:
	go mod tidy

swagger:
	go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -q -g main.go -d cmd/api,internal -o docs
