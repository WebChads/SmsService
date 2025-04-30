package main

import (
	"context"

	"github.com/WebChads/SmsService/internal/config"
	"github.com/WebChads/SmsService/internal/delivery/kafka/consumer"
	"github.com/WebChads/SmsService/internal/delivery/kafka/producer"
)

func main() {
	config := config.NewServiceConfig()
	if config == nil {
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
