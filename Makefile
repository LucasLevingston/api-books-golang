.PHONY: test cover cover-html docs

test:
	go test ./...

cover:
	go test ./internal/book/... -coverprofile=coverage.out
	@go tool cover -func=coverage.out | grep "handler.go"

cover-html:
	go test ./internal/book/... -coverprofile=coverage.out
	go tool cover -html=coverage.out

docs:
	go install github.com/swaggo/swag/cmd/swag@v1.16.6
	swag init -g cmd/api/main.go -o docs
