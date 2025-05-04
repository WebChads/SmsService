package producer

import (
	"encoding/json"
	"log/slog"

	"github.com/IBM/sarama"
	"github.com/WebChads/SmsService/internal/config"
	shared "github.com/WebChads/SmsService/internal/delivery/kafka"
	"github.com/WebChads/SmsService/internal/pkg/slogerr"
	"github.com/WebChads/SmsService/internal/pkg/smsgen"
)

type producedData struct {
	PhoneNumber string `json:"phone_number"`
	SmsCode     string `json:"sms_code"`
}

// "msg" contains phone number claimed by consumer
func prepareMessage(msg []byte, topicName string) *sarama.ProducerMessage {
	// Get generated sms code.
	smsCode, err := smsgen.GenerateMockSmsCode()
	if err != nil {
		slog.Error("generate sms code", slogerr.Error(err))
		return nil
	}

	// TODO: implement SMS Gateway API.
	// At this moment we need to send message with the
	// sms code to the user phone using SMS providers.

	message := producedData{
		PhoneNumber: string(msg),
		SmsCode:     smsCode,
	}

	bytes, err := json.Marshal(&message)
	if err != nil {
		slog.Error("cannot marshal to json", slogerr.Error(err))
		return nil
	}

	res := &sarama.ProducerMessage{
		Topic:     topicName,
		Partition: -1,
		Value:     sarama.StringEncoder(bytes),
	}

	return res
}

func StartProducingSmsCode(mp *shared.MessageProcessor, config *config.ServerConfig, logger *slog.Logger) {
	// Start processor
	mp.Wg.Add(1)
	go func() {
		defer mp.Wg.Done()
		for {
			select {
			case msg := <-mp.ProcessChan:
				producerMsg := prepareMessage(msg.Value, config.ProducerTopic)

				val, _ := sarama.StringEncoder.Encode(producerMsg.Value.(sarama.StringEncoder))

				logger.Info("produced message info",
					slog.String("topic", msg.Topic),
					slog.Int("partition", int(msg.Partition)),
					slog.Int64("offset", msg.Offset),
					slog.String("value", string(val)))

				select {
				case mp.Producer.Input() <- producerMsg:
					// Message sent to producer
				case <-mp.Ctx.Done():
					logger.Info("producer context done:141")
					return
				}

			case <-mp.Producer.Errors():
				logger.Info("producer error")
				return
			case <-mp.Ctx.Done():
				return
			}
		}
	}()
}
