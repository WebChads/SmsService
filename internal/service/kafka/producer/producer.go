package producer

import (
	"encoding/json"
	"log/slog"
	v2 "math/rand/v2"
	"sync"

	"github.com/IBM/sarama"
	"github.com/WebChads/SmsService/internal/lib/slogerr"
	"github.com/WebChads/SmsService/internal/lib/smsgen"
	"github.com/WebChads/SmsService/internal/service"
)

type producedData struct {
	UUID    int    `json:"uuid"`
	SmsCode string `json:"sms_code"`
}

func newSyncProducer(conn string) (sarama.SyncProducer, error) {
	config := sarama.NewConfig()

	// Partitioner which chooses a random partition each time.
	config.Producer.Partitioner = sarama.NewRandomPartitioner

	// Wait acks from all replicants of kafka.
	config.Producer.RequiredAcks = sarama.WaitForAll

	// For implementation reasons, these fields have to be set to true.
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true

	producerSync, err := sarama.NewSyncProducer([]string{conn}, config)

	return producerSync, err
}

const topicName = "smstoauth"

func prepareMessage(msg []byte) *sarama.ProducerMessage {
	res := &sarama.ProducerMessage{
		Topic:     topicName,
		Partition: -1,
		Value:     sarama.StringEncoder(msg),
	}

	return res
}

func produce(producerSync sarama.SyncProducer, msg producedData) {
	const funcPath = "server.kafka.producer.produce"

	logger := slog.With(
		slog.String("path", funcPath),
	)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			bytes, err := json.Marshal(&msg)
			if err != nil {
				logger.Error("cannot marshal to json", slogerr.Err(err))
			}

			msg := prepareMessage(bytes)

			// Send json structure to the topic.
			partition, offset, err := producerSync.SendMessage(msg)
			if err != nil {
				logger.Error("message sync error", slogerr.Err(err))
				logger.Error("message sync error",
					slog.Int("partition", int(partition)),
					slog.Int64("offset", offset))
			}
		}
	}()

	// Wait until all goroutines have finished executing.
	wg.Wait()
}

func StartProducingSmsCode(config *service.ServiceConfig) {
	const funcPath = "server.kafka.producer.StartProducingSmsCode"

	logger := slog.With(
		slog.String("path", funcPath),
	)

	producerSync, err := newSyncProducer(config.KafkaAddress)
	if err != nil {
		logger.Error("newSyncProducer", slogerr.Err(err))
	}
	defer func() {
		if err := producerSync.Close(); err != nil {
			logger.Error("producer close", slogerr.Err(err))
			return
		}
	}()

	// Get generated sms code.
	smsCode, err := smsgen.GenerateMockSmsCode()
	if err != nil {
		logger.Error("generate sms code", slogerr.Err(err))
		return
	}

	// TODO: implement SMS Gateway API.
	// At this moment we need to send message with the
	// sms code to the user phone number.

	message := producedData{
		UUID:    v2.IntN(100),
		SmsCode: smsCode,
	}

	produce(producerSync, message)
}
