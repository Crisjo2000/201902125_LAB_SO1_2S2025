package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	pb "go-grpc-client/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type WeatherTweet struct {
	Municipality string `json:"municipality"`
	Temperature  int32  `json:"temperature"`
	Humidity     int32  `json:"humidity"`
	Weather      string `json:"weather"`
}

var (
	kafkaClient    pb.WeatherTweetServiceClient
	rabbitmqClient pb.WeatherTweetServiceClient
)

func initGRPCClients() error {
	kafkaAddr := os.Getenv("KAFKA_WRITER_ADDR")
	if kafkaAddr == "" {
		kafkaAddr = "go-kafka-writer:50051"
	}

	rabbitmqAddr := os.Getenv("RABBITMQ_WRITER_ADDR")
	if rabbitmqAddr == "" {
		rabbitmqAddr = "go-rabbitmq-writer:50052"
	}

	kafkaConn, err := grpc.Dial(kafkaAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to Kafka writer: %v", err)
	}
	kafkaClient = pb.NewWeatherTweetServiceClient(kafkaConn)

	rabbitmqConn, err := grpc.Dial(rabbitmqAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ writer: %v", err)
	}
	rabbitmqClient = pb.NewWeatherTweetServiceClient(rabbitmqConn)

	log.Println("gRPC clients initialized successfully")
	return nil
}

func handleTweet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var tweet WeatherTweet
	if err := json.NewDecoder(r.Body).Decode(&tweet); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("Received tweet: %+v", tweet)

	municipality := municipalityFromString(tweet.Municipality)
	weather := weatherFromString(tweet.Weather)

	request := &pb.WeatherTweetRequest{
		Municipality: municipality,
		Temperature:  tweet.Temperature,
		Humidity:     tweet.Humidity,
		Weather:      weather,
	}

	errChan := make(chan error, 2)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := kafkaClient.SendTweet(ctx, request)
		if err != nil {
			log.Printf("Error sending to Kafka: %v", err)
		}
		errChan <- err
	}()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := rabbitmqClient.SendTweet(ctx, request)
		if err != nil {
			log.Printf("Error sending to RabbitMQ: %v", err)
		}
		errChan <- err
	}()

	kafkaErr := <-errChan
	rabbitmqErr := <-errChan

	if kafkaErr != nil || rabbitmqErr != nil {
		http.Error(w, "Error forwarding tweet", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "OK"})
}

func municipalityFromString(s string) pb.Municipalities {
	switch s {
	case "mixco":
		return pb.Municipalities_mixco
	case "guatemala":
		return pb.Municipalities_guatemala
	case "amatitlan":
		return pb.Municipalities_amatitlan
	case "chinautla":
		return pb.Municipalities_chinautla
	default:
		return pb.Municipalities_municipalities_unknown
	}
}

func weatherFromString(s string) pb.Weathers {
	switch s {
	case "sunny":
		return pb.Weathers_sunny
	case "cloudy":
		return pb.Weathers_cloudy
	case "rainy":
		return pb.Weathers_rainy
	case "foggy":
		return pb.Weathers_foggy
	default:
		return pb.Weathers_weathers_unknown
	}
}

func main() {
	if err := initGRPCClients(); err != nil {
		log.Fatalf("Failed to initialize gRPC clients: %v", err)
	}

	http.HandleFunc("/tweet", handleTweet)
	http.HandleFunc("/health", healthCheck)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting gRPC client server on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
