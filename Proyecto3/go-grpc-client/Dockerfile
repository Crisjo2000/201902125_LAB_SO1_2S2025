FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

# Forzar versión de Go a 1.22 y permitir toolchain
RUN sed -i 's/go 1.23/go 1.22/g' go.mod && \
    sed -i 's/go 1.24/go 1.22/g' go.mod

# Permitir que Go use una versión diferente si es necesario
ENV GOTOOLCHAIN=auto

RUN go mod download

COPY . .

RUN go build -o grpc-client main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/grpc-client .
ENV GO_KAFKA_URL=go-kafka-writer:50051
ENV GO_RABBITMQ_URL=go-rabbitmq-writer:50052
EXPOSE 8080
CMD ["./grpc-client"]