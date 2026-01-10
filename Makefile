.PHONY: test run tidy

test:
	go test -v ./...

run:
	go run cmd/api/main.go

tidy:
	go mod tidy
