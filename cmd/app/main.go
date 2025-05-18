package main

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/WebChads/SmsService/internal/config"
	"github.com/WebChads/SmsService/internal/services"
)

func main() {
	config, err := config.NewServiceConfig()
	if err != nil {
		errorMessage := fmt.Sprintf("failed to get config: %w", err.Error())
		slog.Error(errorMessage)
		panic(errorMessage)
	}

	kafkaConsumer, err := services.InitKafkaConsumer(config.Brokers)
	if err != nil {
		errorMessage := fmt.Sprintf("failed to init kafka consumer: %w", err.Error())
		slog.Error(errorMessage)
		panic(errorMessage)
	}

	wg := sync.WaitGroup{}

	wg.Add(1)
	go kafkaConsumer.Start()

	defer wg.Done()
	wg.Wait()
}
