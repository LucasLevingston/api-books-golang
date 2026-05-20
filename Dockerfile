FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy sources before tidy so it can resolve all imports and generate go.sum.
COPY go.mod ./
COPY . .
RUN go mod tidy && \
    go install github.com/swaggo/swag/cmd/swag@v1.16.6 && \
    swag init -g cmd/api/main.go -o docs && \
    CGO_ENABLED=0 GOOS=linux go build -a -trimpath -o /api ./cmd/api


FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /api .

EXPOSE 8080

CMD ["./api"]
