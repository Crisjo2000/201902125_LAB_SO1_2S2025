package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"

	pb "go-kafka-writer/proto"

	"github.com/IBM/sarama"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedWeatherTweetServiceServer
	producer sarama.SyncProducer
}

type WeatherData struct {
	Municipality string `json:"municipality"`
	Temperature  int32  `json:"temperature"`
	Humidity     int32  `json:"humidity"`
	Weather      string `json:"weather"`
}

func (s *server) SendTweet(ctx context.Context, req *pb.WeatherTweetRequest) (*pb.WeatherTweetResponse, error) {
	log.Printf("Kafka Writer: Received tweet - Municipality: %v, Temp: %d, Humidity: %d, Weather: %v",
		req.Municipality, req.Temperature, req.Humidity, req.Weather)

	data := WeatherData{
		Municipality: municipalityToString(req.Municipality),
		Temperature:  req.Temperature,
		Humidity:     req.Humidity,
		Weather:      weatherToString(req.Weather),
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling data: %v", err)
		return &pb.WeatherTweetResponse{Status: "error"}, err
	}

	msg := &sarama.ProducerMessage{
		Topic: "weather-tweets",
		Value: sarama.StringEncoder(jsonData),
	}

	partition, offset, err := s.producer.SendMessage(msg)
	if err != nil {
		log.Printf("Failed to send message to Kafka: %v", err)
		return &pb.WeatherTweetResponse{Status: "error"}, err
	}

	log.Printf("Message sent to Kafka partition %d at offset %d", partition, offset)
	return &pb.WeatherTweetResponse{Status: "success"}, nil
}

func municipalityToString(m pb.Municipalities) string {
	switch m {
	case pb.Municipalities_mixco:
		return "mixco"
	case pb.Municipalities_guatemala:
		return "guatemala"
	case pb.Municipalities_amatitlan:
		return "amatitlan"
	case pb.Municipalities_chinautla:
		return "chinautla"
	default:
		return "unknown"
	}
}

func weatherToString(w pb.Weathers) string {
	switch w {
	case pb.Weathers_sunny:
		return "sunny"
	case pb.Weathers_cloudy:
		return "cloudy"
	case pb.Weathers_rainy:
		return "rainy"
	case pb.Weathers_foggy:
		return "foggy"
	default:
		return "unknown"
	}
}

func initKafkaProducer() (sarama.SyncProducer, error) {
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	if kafkaBroker == "" {
		kafkaBroker = "kafka-service:9092"
	}

	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5

	producer, err := sarama.NewSyncProducer([]string{kafkaBroker}, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %v", err)
	}

	log.Println("Kafka producer initialized successfully")
	return producer, nil
}

func main() {
	producer, err := initKafkaProducer()
	if err != nil {
		log.Fatalf("Failed to initialize Kafka producer: %v", err)
	}
	defer producer.Close()

	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50051"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterWeatherTweetServiceServer(s, &server{producer: producer})

	log.Printf("Kafka Writer gRPC server listening on port %s", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
