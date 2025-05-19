package main

import (
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/WebChads/SmsService/internal/config"
	"github.com/WebChads/SmsService/internal/services"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	slog.Info("Resolving dependency injection")

	config, err := config.NewServiceConfig()
	if err != nil {
		errorMessage := fmt.Sprintf("failed to get config: %s", err.Error())
		slog.Error(errorMessage)
		panic(errorMessage)
	}

	kafkaProducer, err := services.NewKafkaProducer(config.KafkaAddress)
	if err != nil {
		errorMessage := fmt.Sprintf("failed to init kafka producer: %s", err.Error())
		slog.Error(errorMessage)
		panic(errorMessage)
	}

	kafkaConsumer, err := services.InitKafkaConsumer(config.Brokers, kafkaProducer)
	if err != nil {
		errorMessage := fmt.Sprintf("failed to init kafka consumer: %s", err.Error())
		slog.Error(errorMessage)
		panic(errorMessage)
	}

	wg := sync.WaitGroup{}

	wg.Add(1)
	go kafkaConsumer.Start()

	defer wg.Done()

	slog.Info("Started SmsService")
	wg.Wait()
}
