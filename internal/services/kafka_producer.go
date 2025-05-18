package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/WebChads/SmsService/internal/pkg/smsgen"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type KafkaProducer interface {
	SendSmsCode(phoneNumber string) error
}

type confluentKafkaProducer struct {
	kafkaProducer *kafka.Producer
}

type smsCodeMessage struct {
	PhoneNumber string `json:"phone_number"`
	SmsCode     string `json:"sms_code"`
}

var producerTopicName string = "smstoauth"
var singletoneKafkaProducer *confluentKafkaProducer = &confluentKafkaProducer{}

func (kafkaProducer *confluentKafkaProducer) SendSmsCode(phoneNumber string) error {
	smsCode, err := smsgen.GenerateMockSmsCode()
	if err != nil {
		errorMessage := fmt.Sprintf("happened error while generating sms code: %s", err.Error())
		slog.Error(errorMessage)
		return errors.New(errorMessage)
	}

	dto := smsCodeMessage{PhoneNumber: phoneNumber, SmsCode: smsCode}
	encodedMessage, err := json.Marshal(dto)

	if err != nil {
		errorMessage := fmt.Sprintf("while encoding dto to send happened error: %s", err.Error())
		slog.Error(errorMessage)
		return errors.New(errorMessage)
	}

	message := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &producerTopicName, Partition: kafka.PartitionAny},
		Value:          encodedMessage,
	}

	err = kafkaProducer.kafkaProducer.Produce(message, nil)
	if err != nil {
		return errors.New("while producing message in kafka happened error: " + err.Error())
	}

	return nil
}

func NewKafkaProducer(kafkaAddress string) (KafkaProducer, error) {
	if singletoneKafkaProducer.kafkaProducer == nil {
		producer, err := kafka.NewProducer(&kafka.ConfigMap{
			"bootstrap.servers": kafkaAddress,
		})

		if err != nil {
			return nil, err
		}

		singletoneKafkaProducer.kafkaProducer = producer
	}

	return singletoneKafkaProducer, nil
}
