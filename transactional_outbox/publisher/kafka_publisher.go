package publisher

import (
	"fmt"
	"log"

	"github.com/IBM/sarama"
)

type KafkaPublisher struct {
	producer sarama.SyncProducer
	topic    string
}

func NewKafkaPublisher(brokers []string, topic string) (*KafkaPublisher, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	log.Printf("Kafka producer created successfully, topic: %s", topic)

	return &KafkaPublisher{
		producer: producer,
		topic:    topic,
	}, nil
}

func (p *KafkaPublisher) Publish(key string, message []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(message),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to publish message to Kafka: %w", err)
	}

	log.Printf("Message published to Kafka: topic=%s, partition=%d, offset=%d, key=%s",
		p.topic, partition, offset, key)

	return nil
}

func (p *KafkaPublisher) Close() error {
	if err := p.producer.Close(); err != nil {
		return fmt.Errorf("failed to close Kafka producer: %w", err)
	}
	log.Println("Kafka producer closed")
	return nil
}
