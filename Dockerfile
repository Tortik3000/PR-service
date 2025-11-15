FROM golang:1.25
WORKDIR /app
COPY . .

RUN make build
CMD ["./bin/pr-service"]