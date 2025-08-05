FROM golang:latest AS builder
WORKDIR /app
COPY go.mod main.go ./
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o sni-finder main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/sni-finder .
ENTRYPOINT ["./sni-finder"]
