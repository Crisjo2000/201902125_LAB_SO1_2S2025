
FROM golang:latest AS builder

WORKDIR /app

COPY . .

RUN go mod tidy

RUN go build -o api1 .

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/api1 .

EXPOSE 8080

CMD ["./api1"]