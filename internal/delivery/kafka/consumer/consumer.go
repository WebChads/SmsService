package consumer

import (
	"errors"
	"log/slog"

	"github.com/IBM/sarama"
	"github.com/WebChads/SmsService/internal/config"
	shared "github.com/WebChads/SmsService/internal/delivery/kafka"
)

// TODO: use Sarama's metrics and monitoring system (e.g. Prometheus).

// This structure implements "ConsumerGroupHandler" interface.
type ConsumerHandler struct {
	mp     *shared.MessageProcessor
	logger *slog.Logger
}

// When a new consumer is added to a consumer group or an existing consumer leaves
// (due to a crash or being shutdown), Kafka rebalances the consumers in the group.

// Rebalancing is the process by which Kafka ensures that all the consumers in a
// consumer group are consuming from unique partitions.
// In Sarama, handling rebalancing is done by implementing the Setup and Cleanup
// functions of the ConsumerGroupHandler interface.

func (c *ConsumerHandler) Setup(session sarama.ConsumerGroupSession) error {
	c.logger.Info("Consumer joins cluster and starts a blocking session",
		slog.String("ConsumerHandler", "Setup"))

	return nil
}

func (c *ConsumerHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	c.logger.Info("Stop claiming, run cleanup procedure",
		slog.String("ConsumerHandler", "Cleanup"))

	return nil
}

// NOTE
// The messages channel will be closed when a new rebalance cycle is due.
func (c *ConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim) error {
	c.logger.Info("Start claiming messages",
		slog.String("ConsumerHandler", "ConsumeClaim"))

	for message := range claim.Messages() {
		select {
		case c.mp.ProcessChan <- message:
			c.logger.Info("claimed message info",
				slog.String("topic", message.Topic),
				slog.Int("partition", int(message.Partition)),
				slog.Int64("offset", message.Offset),
				slog.String("value", string(message.Value)))

			session.MarkMessage(message, "message was marked as handled")
		case <-c.mp.Ctx.Done():
			c.logger.Info("consumer context done:61")
			return nil
		}
	}

	return nil
}

// Start consuming in the separate goroutine.
func StartConsumingPhoneNumber(mp *shared.MessageProcessor, config *config.ServerConfig, l *slog.Logger) {
	handler := &ConsumerHandler{
		mp:     mp,
		logger: l,
	}

	mp.Wg.Add(1)
	go func() {
		defer mp.Wg.Done()

		// Consume method should be called inside an infinite loop.
		// When a server-side rebalance happens, the consumer session will need to be
		// recreated to get the new claims.
		for {
			select {
			case <-mp.Ctx.Done():
				l.Info("consumer context done:85")
				return
			default:
				err := mp.Consumer.Consume(mp.Ctx, []string{config.ConsumerTopic}, handler)
				if err != nil {
					mp.ErrorChan <- errors.New("consumer error: " + err.Error())
					return
				}
			}
		}
	}()
}
