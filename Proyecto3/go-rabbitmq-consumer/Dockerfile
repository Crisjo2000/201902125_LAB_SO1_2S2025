FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o rabbitmq-consumer main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/rabbitmq-consumer .

ENV RABBITMQ_URL=amqp://guest:guest@rabbitmq-service:5672/
ENV VALKEY_ADDR=valkey-service:6379

CMD ["./rabbitmq-consumer"]