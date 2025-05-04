package shared

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/WebChads/SmsService/internal/config"
	"github.com/WebChads/SmsService/internal/pkg/slogerr"
)

// MessageProcessor handles the communication between consumer and producer
type MessageProcessor struct {
	Producer    sarama.AsyncProducer
	Consumer    sarama.ConsumerGroup
	ProcessChan chan *sarama.ConsumerMessage // Channel for passing messages
	ErrorChan   chan error                   // Channel for error handling
	Wg          sync.WaitGroup
	Ctx         context.Context
	Cancel      context.CancelFunc
}

func NewMessageProcessor(config *config.ServerConfig) (*MessageProcessor, error) {
	ctx, cancel := context.WithCancel(context.Background())

	producer, err := NewAsyncProducer(config.KafkaAddress)
	if err != nil {
		cancel()
		return nil, err
	}

	groupId := "smsgroup"
	consumer, err := NewConsumerGroup(config, groupId)
	if err != nil {
		cancel()
		producer.Close()
		return nil, err
	}

	return &MessageProcessor{
		Producer:    producer,
		Consumer:    consumer,
		ProcessChan: make(chan *sarama.ConsumerMessage, 100), // Buffered channel
		ErrorChan:   make(chan error, 10),
		Ctx:         ctx,
		Cancel:      cancel,
	}, nil
}

func NewAsyncProducer(conn string) (sarama.AsyncProducer, error) {
	config := sarama.NewConfig()

	// Partitioner which chooses a random partition each time.
	config.Producer.Partitioner = sarama.NewRandomPartitioner

	// Wait acks from all replicants of kafka.
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 10
	config.Producer.Idempotent = true
	config.Net.MaxOpenRequests = 1

	// 16KB batch size
	config.Producer.Flush.Bytes = 16 * 1024
	// Enable message compression
	config.Producer.Compression = sarama.CompressionSnappy

	// For implementation reasons, these fields have to be set to true.
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true

	producer, err := sarama.NewAsyncProducer([]string{conn}, config)

	return producer, err
}

func NewConsumerGroup(config *config.ServerConfig, groupId string) (sarama.ConsumerGroup, error) {
	consumerConfig := sarama.NewConfig()

	// Reliability settings
	consumerConfig.Consumer.Offsets.Initial = sarama.OffsetNewest // Or OffsetOldest
	consumerConfig.Consumer.Offsets.AutoCommit.Enable = false     // Manual commit
	consumerConfig.Consumer.Offsets.Retry.Max = 3
	consumerConfig.Consumer.IsolationLevel = sarama.ReadCommitted // Don't read aborted transactions

	// Performance settings
	consumerConfig.Consumer.Fetch.Default = 1 * 1024 * 1024 // 1MB fetch size
	consumerConfig.Consumer.Fetch.Max = 5 * 1024 * 1024     // 5MB max fetch size
	consumerConfig.Consumer.MaxWaitTime = 500 * time.Millisecond

	// Heartbeat settings
	consumerConfig.Consumer.Group.Heartbeat.Interval = 3 * time.Second
	consumerConfig.Consumer.Group.Session.Timeout = 10 * time.Second
	consumerConfig.Consumer.Group.Rebalance.Timeout = 60 * time.Second

	consumerGroup, err := sarama.NewConsumerGroup(config.Brokers, groupId, consumerConfig)

	return consumerGroup, err
}

func (mp *MessageProcessor) Close() {
	mp.Cancel()
	mp.Wg.Wait()

	if err := mp.Producer.Close(); err != nil {
		slog.Error("Error closing producer", slogerr.Error(err))
	}

	if err := mp.Consumer.Close(); err != nil {
		slog.Error("Error closing consumer", slogerr.Error(err))
	}

	close(mp.ProcessChan)
	close(mp.ErrorChan)
}
