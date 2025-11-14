FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o pr-service ./cmd/pr-service

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/pr-service .
CMD ["./pr-service"]