#!/bin/bash

ZOT_REGISTRY="34.70.157.232:5000"

echo "==========================================================="
echo "Construyendo y pusheando todas las imágenes de Docker a Zot"
echo "==========================================================="

# Rust API
echo "Building Rust API..."
cd rust-api
docker build -t ${ZOT_REGISTRY}/rust-api:latest .
docker push ${ZOT_REGISTRY}/rust-api:latest
cd ..

# Go gRPC Client
echo "Building Go gRPC Client..."
cd go-grpc-client
docker build -t ${ZOT_REGISTRY}/go-grpc-client:latest .
docker push ${ZOT_REGISTRY}/go-grpc-client:latest
cd ..

# Go Kafka Writer
echo "Building Go Kafka Writer..."
cd go-kafka-writer
docker build -t ${ZOT_REGISTRY}/go-kafka-writer:latest .
docker push ${ZOT_REGISTRY}/go-kafka-writer:latest
cd ..

# Go RabbitMQ Writer
echo "Building Go RabbitMQ Writer..."
cd go-rabbitmq-writer
docker build -t ${ZOT_REGISTRY}/go-rabbitmq-writer:latest .
docker push ${ZOT_REGISTRY}/go-rabbitmq-writer:latest
cd ..

# Kafka Consumer
echo "Building Kafka Consumer..."
cd go-kafka-consumer
docker build -t ${ZOT_REGISTRY}/kafka-consumer:latest .
docker push ${ZOT_REGISTRY}/kafka-consumer:latest
cd ..

# RabbitMQ Consumer
echo "Building RabbitMQ Consumer..."
cd go-rabbitmq-consumer
docker build -t ${ZOT_REGISTRY}/rabbitmq-consumer:latest .
docker push ${ZOT_REGISTRY}/rabbitmq-consumer:latest
cd ..

echo "=============================================================="
echo "Todas las imágenes se subieron y pushearon satisfactoriamente!"
echo "=============================================================="