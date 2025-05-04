package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	prettylogger "github.com/WebChads/AccountService/pkg/pretty_logger"
	"github.com/WebChads/SmsService/internal/config"
	shared "github.com/WebChads/SmsService/internal/delivery/kafka"
	"github.com/WebChads/SmsService/internal/delivery/kafka/consumer"
	"github.com/WebChads/SmsService/internal/delivery/kafka/producer"
	"github.com/WebChads/SmsService/internal/pkg/slogerr"
)

func main() {
	// Init server configuration
	config := config.NewServerConfig()
	if config == nil {
		return
	}

	// Init logger
	logger := setupLogger(config.LogLevel)

	processor, err := shared.NewMessageProcessor(config)
	if err != nil {
		slog.Error("Failed to create processor", slogerr.Error(err))
		return
	}
	defer processor.Close()

	// Consuming runs in another goroutine, so we use wait group.
	consumer.StartConsumingPhoneNumber(processor, config, logger)

	// Producing runs in another goroutine, so we use wait group.
	producer.StartProducingSmsCode(processor, config, logger)

	// Start error handler
	processor.Wg.Add(1)
	go func() {
		defer processor.Wg.Done()
		for {
			select {
			// Read from this channel or the Producer will deadlock
			// when the channel is full.
			case err := <-processor.Producer.Errors():
				logger.Info("Producer error", slogerr.Error(err))
			case err := <-processor.ErrorChan:
				logger.Info("System error", slogerr.Error(err))
			case <-processor.Ctx.Done():
				logger.Info("error handler context done:53")
				return
			}
		}
	}()

	// Wait for termination signal
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	<-sigterm
	logger.Info("Received termination signal, shutting down...")
}

const (
	envLocal = "local"
	envStage = "stage"
	envProd  = "prod"
)

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		handler := prettylogger.NewPrettyHandler(os.Stdout)
		log = slog.New(handler)
	case envStage:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	default: // If env config is invalid, set prod settings by default due to security
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
