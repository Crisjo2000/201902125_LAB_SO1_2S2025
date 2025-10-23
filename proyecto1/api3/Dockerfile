FROM golang:latest AS builder

WORKDIR /app

COPY . .

RUN go mod tidy

RUN go build -o api3 .

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/api3 .

EXPOSE 8080

CMD ["./api3"]