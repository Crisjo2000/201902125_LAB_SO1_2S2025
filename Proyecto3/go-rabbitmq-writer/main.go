package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	pb "go-rabbitmq-writer/proto"

	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedWeatherTweetServiceServer
	channel *amqp.Channel
}

type WeatherData struct {
	Municipality string `json:"municipality"`
	Temperature  int32  `json:"temperature"`
	Humidity     int32  `json:"humidity"`
	Weather      string `json:"weather"`
}

func (s *server) SendTweet(ctx context.Context, req *pb.WeatherTweetRequest) (*pb.WeatherTweetResponse, error) {
	log.Printf("RabbitMQ Writer: Received tweet - Municipality: %v, Temp: %d, Humidity: %d, Weather: %v",
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

	err = s.channel.PublishWithContext(
		ctx,
		"",               // exchange
		"weather-tweets", // routing key (queue name)
		false,            // mandatory
		false,            // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        jsonData,
		})

	if err != nil {
		log.Printf("Failed to publish message to RabbitMQ: %v", err)
		return &pb.WeatherTweetResponse{Status: "error"}, err
	}

	log.Println("Message published to RabbitMQ")
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

func initRabbitMQ() (*amqp.Channel, error) {
	rabbitmqURL := os.Getenv("RABBITMQ_URL")
	if rabbitmqURL == "" {
		rabbitmqURL = "amqp://guest:guest@rabbitmq-service:5672/"
	}

	var conn *amqp.Connection
	var err error

	// Retry connection
	for i := 0; i < 10; i++ {
		conn, err = amqp.Dial(rabbitmqURL)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to RabbitMQ (attempt %d/10): %v", i+1, err)
		time.Sleep(5 * time.Second)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ after retries: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %v", err)
	}

	// Declare queue
	_, err = ch.QueueDeclare(
		"weather-tweets", // name
		true,             // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare queue: %v", err)
	}

	log.Println("RabbitMQ channel initialized successfully")
	return ch, nil
}

func main() {
	channel, err := initRabbitMQ()
	if err != nil {
		log.Fatalf("Failed to initialize RabbitMQ: %v", err)
	}
	defer channel.Close()

	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50052"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterWeatherTweetServiceServer(s, &server{channel: channel})

	log.Printf("RabbitMQ Writer gRPC server listening on port %s", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
