package main

import (
	"context"
	"log/slog"

	"github.com/WebChads/SmsService/internal/lib/slogerr"
	"github.com/WebChads/SmsService/internal/service"
	"github.com/WebChads/SmsService/internal/service/kafka/consumer"
	"github.com/WebChads/SmsService/internal/service/kafka/producer"
)

func main() {
	config, err := service.NewServiceConfig()
	if err != nil {
		slog.Error("", slogerr.Err(err))
		return
	}

	ctx := context.Background()

	// create kafka consumer group and start consuming.
	// consuming runs in another goroutine, so we use
	// wait group.
	consumer.StartConsumingPhoneNumber(ctx, config)

	// create kafka sync producer and start producing.
	// producing runs in another goroutine, so we use
	// wait group.
	producer.StartProducingSmsCode(config)
}
