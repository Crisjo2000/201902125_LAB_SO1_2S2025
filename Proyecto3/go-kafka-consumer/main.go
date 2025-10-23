package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/IBM/sarama"
	"github.com/redis/go-redis/v9"
)

type WeatherData struct {
	Municipality string `json:"municipality"`
	Temperature  int32  `json:"temperature"`
	Humidity     int32  `json:"humidity"`
	Weather      string `json:"weather"`
}

type Consumer struct {
	ready chan bool
	rdb   *redis.Client
	ctx   context.Context
}

func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	close(consumer.ready)
	return nil
}

func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		log.Printf("Kafka Consumer: Message received: %s", string(message.Value))

		var data WeatherData
		if err := json.Unmarshal(message.Value, &data); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			session.MarkMessage(message, "")
			continue
		}

		// Guardar en Valkey (Redis)
		if err := consumer.saveToValkey(data); err != nil {
			log.Printf("Error saving to Valkey: %v", err)
		} else {
			log.Printf("Data saved to Valkey: %+v", data)
		}

		session.MarkMessage(message, "")
	}

	return nil
}

func (consumer *Consumer) saveToValkey(data WeatherData) error {
	// Incrementar contador por municipio
	municipalityKey := "municipality:" + data.Municipality
	if err := consumer.rdb.Incr(consumer.ctx, municipalityKey).Err(); err != nil {
		return err
	}

	// Incrementar contador por clima
	weatherKey := "weather:" + data.Weather
	if err := consumer.rdb.Incr(consumer.ctx, weatherKey).Err(); err != nil {
		return err
	}

	// Guardar temperatura promedio
	tempKey := "temp:" + data.Municipality
	tempListKey := "temp_list:" + data.Municipality

	// Agregar temperatura a la lista
	if err := consumer.rdb.RPush(consumer.ctx, tempListKey, data.Temperature).Err(); err != nil {
		return err
	}

	// Calcular promedio
	temps, err := consumer.rdb.LRange(consumer.ctx, tempListKey, 0, -1).Result()
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

	if err := consumer.rdb.Set(consumer.ctx, tempKey, avg, 0).Err(); err != nil {
		return err
	}

	// Guardar humedad promedio
	humKey := "humidity:" + data.Municipality
	humListKey := "humidity_list:" + data.Municipality

	if err := consumer.rdb.RPush(consumer.ctx, humListKey, data.Humidity).Err(); err != nil {
		return err
	}

	hums, err := consumer.rdb.LRange(consumer.ctx, humListKey, 0, -1).Result()
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

	if err := consumer.rdb.Set(consumer.ctx, humKey, humAvg, 0).Err(); err != nil {
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

	// Verificar conexión
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Valkey: %v", err)
	}
	log.Println("Connected to Valkey successfully")

	// Configurar Kafka
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	if kafkaBroker == "" {
		kafkaBroker = "kafka-service:9092"
	}

	config := sarama.NewConfig()
	config.Version = sarama.V2_6_0_0
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumer := Consumer{
		ready: make(chan bool),
		rdb:   rdb,
		ctx:   ctx,
	}

	consumerGroup, err := sarama.NewConsumerGroup([]string{kafkaBroker}, "weather-consumer-group", config)
	if err != nil {
		log.Fatalf("Error creating consumer group: %v", err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			if err := consumerGroup.Consume(ctx, []string{"weather-tweets"}, &consumer); err != nil {
				log.Printf("Error from consumer: %v", err)
			}
			consumer.ready = make(chan bool)
		}
	}()

	<-consumer.ready
	log.Println("Kafka consumer up and running")

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	<-sigterm

	log.Println("Shutting down consumer...")
	consumerGroup.Close()
	wg.Wait()
}
