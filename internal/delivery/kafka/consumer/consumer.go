package consumer

import (
	"context"
	"log/slog"
	"sync"

	"github.com/IBM/sarama"
	"github.com/WebChads/SmsService/internal/config"
	"github.com/WebChads/SmsService/internal/pkg/slogerr"
)

// TODO: use Sarama's metrics and monitoring system (e.g. Prometheus).

// This structure implements "ConsumerGroupHandler" interface.
type ConsumerHandler struct{}

// When a new consumer is added to a consumer group or an existing consumer leaves
// (due to a crash or being shutdown), Kafka rebalances the consumers in the group.

// Rebalancing is the process by which Kafka ensures that all the consumers in a
// consumer group are consuming from unique partitions.
// In Sarama, handling rebalancing is done by implementing the Setup and Cleanup
// functions of the ConsumerGroupHandler interface.

func (c *ConsumerHandler) Setup(s sarama.ConsumerGroupSession) error {
	slog.Info("Consumer joins cluster and starts a blocking" +
		"ConsumerGroupSession with ConsumerGroupHandler")
	slog.Info("Consumer group is being rebalanced")

	return nil
}

func (c *ConsumerHandler) Cleanup(s sarama.ConsumerGroupSession) error {
	slog.Info("ConsumeClaim was finished and cleanup procedure was started")
	slog.Info("Rebalancing will happen soon, current session will end")

	return nil
}

// NOTE
// The messages channel will be closed when a new rebalance cycle is due.
func (c *ConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		slog.Info("claimed message info",
			slog.String("topic", message.Topic),
			slog.Int("partition", int(message.Partition)),
			slog.Int64("offset", message.Offset),
			slog.String("value", string(message.Value)))

		session.MarkMessage(message, "message was marked as handled")
	}

	return nil
}

// Start consuming in the separate goroutine.
func subscribe(ctx context.Context, topics []string, consumerGroup sarama.ConsumerGroup) {
	consumerHandler := ConsumerHandler{}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		// consumerGroup.Consume method should be called inside an infinite loop.
		// when a server-side rebalance happens, the consumer session will need to be
		// recreated to get the new claims.
		for {
			if err := consumerGroup.Consume(ctx, topics, &consumerHandler); err != nil {
				return
			}

			if ctx.Err() != nil {
				return
			}
		}
	}()

	// Wait until all goroutines have finished executing.
	wg.Wait()
}

var topicNames = []string{"authtosms"}

// Create, config and start new consumer group.
func StartConsumingPhoneNumber(ctx context.Context, config *config.ServerConfig) {
	const funcPath = "service.kafka.consumer.StartConsumingPhoneNumber"

	logger := slog.With(
		slog.String("path", funcPath),
	)

	consumerConfig := sarama.NewConfig()

	consumerConfig.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	consumerConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	// Consumer group id is "sms".
	consumerGroup, err := sarama.NewConsumerGroup(config.Brokers, "sms", consumerConfig)
	if err != nil {
		logger.Error("create new consumer group error", slogerr.Error(err))
		return
	}

	// Subscribe consumer group on smsfromauth topic.
	subscribe(ctx, topicNames, consumerGroup)
}
