FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

# Forzar versión de Go a 1.22
RUN sed -i 's/go 1.23/go 1.22/g' go.mod && \
    sed -i 's/go 1.24/go 1.22/g' go.mod

ENV GOTOOLCHAIN=auto

RUN go mod download

COPY . .

RUN go build -o rabbitmq-writer main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/rabbitmq-writer .

ENV GRPC_PORT=50052
ENV RABBITMQ_URL=amqp://guest:guest@rabbitmq-service:5672/

EXPOSE 50052

CMD ["./rabbitmq-writer"]