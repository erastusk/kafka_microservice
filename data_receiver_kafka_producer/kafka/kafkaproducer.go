package kafka

import (
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// ConfigMap documentation
// https://github.com/confluentinc/librdkafka/blob/master/CONFIGURATION.md
var (
	server = "gpscords_app-kafka-1:9092"
	topic  = "gpscoords"
)

type KafkaProducer struct {
	Producer   *kafka.Producer
	topic      string
	chan_event chan kafka.Event
}

// To produce asynchronously, you can use a Goroutine to handle message delivery reports and possibly other event types (errors, stats, etc) concurrently:
func NewKafkaProducer() (*KafkaProducer, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": server,
	})
	if err != nil {
		fmt.Println("Failed to create Kafka producer", err)
		return nil, err
	}
	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("Failed to deliver message: %v\n", ev.TopicPartition)
				} else {
					fmt.Printf("************\nSuccessfully produced record to topic %s partition [%d] @ offset %v\n*****************\n",
						*ev.TopicPartition.Topic, ev.TopicPartition.Partition, ev.TopicPartition.Offset)
				}
			}
		}
	}()
	return &KafkaProducer{
		Producer:   p,
		topic:      topic,
		chan_event: make(chan kafka.Event, 1000),
	}, nil
}

func (p *KafkaProducer) KafkaWrite(word []byte) {
	// Produce messages to topic (asynchrjonously)
	p.Producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          word,
	}, nil)

	// Wait for message deliveries before shutting down
	p.Producer.Flush(15 * 1000)
}
