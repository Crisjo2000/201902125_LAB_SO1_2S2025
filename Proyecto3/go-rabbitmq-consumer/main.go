package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

type WeatherData struct {
	Municipality string `json:"municipality"`
	Temperature  int32  `json:"temperature"`
	Humidity     int32  `json:"humidity"`
	Weather      string `json:"weather"`
}

func saveToValkey(ctx context.Context, rdb *redis.Client, data WeatherData) error {
	// Incrementar contador por municipio
	municipalityKey := "municipality:" + data.Municipality
	if err := rdb.Incr(ctx, municipalityKey).Err(); err != nil {
		return err
	}

	// Incrementar contador por clima
	weatherKey := "weather:" + data.Weather
	if err := rdb.Incr(ctx, weatherKey).Err(); err != nil {
		return err
	}

	// Guardar temperatura promedio
	tempKey := "temp:" + data.Municipality
	tempListKey := "temp_list:" + data.Municipality

	if err := rdb.RPush(ctx, tempListKey, data.Temperature).Err(); err != nil {
		return err
	}

	temps, err := rdb.LRange(ctx, tempListKey, 0, -1).Result()
	if err != nil {
		return err
	}

	var sum int32
	for _, temp := range temps {
		var t int32
		json.Unmarshal([]byte(temp), &t)
		sum += t
	}
	avg := float64(sum) / float64(len(temps))

	if err := rdb.Set(ctx, tempKey, avg, 0).Err(); err != nil {
		return err
	}

	// Guardar humedad promedio
	humKey := "humidity:" + data.Municipality
	humListKey := "humidity_list:" + data.Municipality

	if err := rdb.RPush(ctx, humListKey, data.Humidity).Err(); err != nil {
		return err
	}

	hums, err := rdb.LRange(ctx, humListKey, 0, -1).Result()
	if err != nil {
		return err
	}

	var humSum int32
	for _, hum := range hums {
		var h int32
		json.Unmarshal([]byte(hum), &h)
		humSum += h
	}
	humAvg := float64(humSum) / float64(len(hums))

	if err := rdb.Set(ctx, humKey, humAvg, 0).Err(); err != nil {
		return err
	}

	return nil
}

func main() {
	ctx := context.Background()

	// Configurar Redis/Valkey
	valkeyAddr := os.Getenv("VALKEY_ADDR")
	if valkeyAddr == "" {
		valkeyAddr = "valkey-service:6379"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: valkeyAddr,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Valkey: %v", err)
	}
	log.Println("Connected to Valkey successfully")

	// Configurar RabbitMQ
	rabbitmqURL := os.Getenv("RABBITMQ_URL")
	if rabbitmqURL == "" {
		rabbitmqURL = "amqp://guest:guest@rabbitmq-service:5672/"
	}

	var conn *amqp.Connection
	var err error

	for i := 0; i < 10; i++ {
		conn, err = amqp.Dial(rabbitmqURL)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to RabbitMQ (attempt %d/10): %v", i+1, err)
		time.Sleep(5 * time.Second)
	}

	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ after retries: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open channel: %v", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"weather-tweets",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}

	msgs, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to register consumer: %v", err)
	}

	log.Println("RabbitMQ consumer up and running")

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for msg := range msgs {
			log.Printf("RabbitMQ Consumer: Message received: %s", string(msg.Body))

			var data WeatherData
			if err := json.Unmarshal(msg.Body, &data); err != nil {
				log.Printf("Error unmarshaling message: %v", err)
				continue
			}

			if err := saveToValkey(ctx, rdb, data); err != nil {
				log.Printf("Error saving to Valkey: %v", err)
			} else {
				log.Printf("Data saved to Valkey: %+v", data)
			}
		}
	}()

	<-sigterm
	log.Println("Shutting down consumer...")
}
