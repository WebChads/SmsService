package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type KafkaConsumer interface {
	Start()
}

type confluentKafkaConsumer struct {
	kafkaConsumer *kafka.Consumer
	kafkaProducer KafkaProducer
	isStarted     bool
}

type phoneNumberRequestDto struct {
	PhoneNumber string `json:"phone_number"`
}

var consumerTopicName = "auth-to-sms"
var compiledPhoneNumberRegex *regexp.Regexp

func (kafkaConsumer *confluentKafkaConsumer) Start() {
	if kafkaConsumer.isStarted {
		panic("kafka consumer was already started")
	}

	err := kafkaConsumer.kafkaConsumer.Subscribe(consumerTopicName, nil)
	if err != nil {
		panic("while subscribing to topic happened error: " + err.Error())
	}

	kafkaConsumer.isStarted = true

	for {
		message, err := kafkaConsumer.kafkaConsumer.ReadMessage(-1)
		if err != nil {
			slog.Error("while listening for messages happened error: " + err.Error())
			continue
		}

		var phoneNumberDto phoneNumberRequestDto
		err = json.Unmarshal(message.Value, &phoneNumberDto)
		if err != nil || phoneNumberDto.PhoneNumber == "" {
			if err != nil {
				slog.Error("while unmarshalling message in listener happened error: " + err.Error())
			} else {
				slog.Error("while unmarshalling message in listener happened error: fields of dto remained empty somehow")
			}

			continue
		}

		slog.Info(fmt.Sprintf("received request to send sms code on phone number %s", phoneNumberDto.PhoneNumber))

		isPhoneNumberCorrect := compiledPhoneNumberRegex.Match([]byte(phoneNumberDto.PhoneNumber))
		if !isPhoneNumberCorrect {
			slog.Error(fmt.Errorf("user sent invalid phone number: %s", phoneNumberDto.PhoneNumber).Error())
		}

		kafkaConsumer.kafkaProducer.SendSmsCode(phoneNumberDto.PhoneNumber)
	}
}

var singletoneKafkaConsumer = &confluentKafkaConsumer{}

func InitKafkaConsumer(brokers []string, kafkaProducer KafkaProducer) (KafkaConsumer, error) {
	if singletoneKafkaConsumer.kafkaConsumer == nil {
		consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
			"bootstrap.servers": strings.Join(brokers, ","),
			"group.id":          "auth-to-sms-listener",
			"auto.offset.reset": "earliest"})

		if err != nil {
			return nil, errors.New("while initing kafka consumer happened error: " + err.Error())
		}

		singletoneKafkaConsumer.kafkaConsumer = consumer
		singletoneKafkaConsumer.kafkaProducer = kafkaProducer

		compiledPhoneNumberRegex, _ = regexp.Compile(`^(8|\+7)(\s|\(|-)?(\d{3})(\s|\)|-)?(\d{3})(\s|-)?(\d{2})(\s|-)?(\d{2})$`)
	}

	return singletoneKafkaConsumer, nil
}
