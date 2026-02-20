// Package kafka provides Kafka producer functionality for publishing GPS coordinates.
// Uses the Confluent Kafka Go client (librdkafka wrapper) for high-performance message production.
package kafka

import (
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// Kafka broker and topic configuration.
// ConfigMap documentation: https://github.com/confluentinc/librdkafka/blob/master/CONFIGURATION.md
var (
	// server is the Kafka broker address.
	// Uses Docker Compose service name for container-to-container communication.
	server = "gpscords_app-kafka-1:9092"
	
	// topic is the Kafka topic name where GPS coordinates are published.
	topic  = "gpscoords"
)

// KafkaProducer wraps a Kafka producer with event handling capabilities.
//
// Fields:
//   - Producer: Confluent Kafka producer instance for message publication
//   - topic: Target Kafka topic name
//   - chan_event: Event channel for delivery reports and errors (currently unused)
//
// The producer operates asynchronously:
//   - Messages are queued for delivery
//   - Delivery reports arrive via event channel
//   - Flush() waits for all queued messages to be delivered
type KafkaProducer struct {
	Producer   *kafka.Producer     // Kafka producer instance
	topic      string              // Topic name to publish to
	chan_event chan kafka.Event    // Event channel (for future use)
}

// NewKafkaProducer creates and initializes a new asynchronous Kafka producer.
//
// Producer Configuration:
//   - Asynchronous delivery: Messages are queued and sent in background
//   - No manual acknowledgments required
//   - Delivery reports handled via background goroutine
//
// The background goroutine:
//   - Monitors producer events (delivery reports, errors)
//   - Logs successful message delivery with topic, partition, and offset
//   - Logs delivery failures for debugging
//   - Runs for the lifetime of the producer
//
// Returns:
//   - *KafkaProducer: Configured producer ready to send messages
//   - error: Connection or configuration errors
//
// Key features:
//   - Fire-and-forget: Send messages without waiting for acknowledgment
//   - Automatic retries: Kafka client handles transient failures
//   - Buffering: Messages buffered for batch delivery (better throughput)
//
// For production considerations, see:
// https://github.com/confluentinc/librdkafka/blob/master/CONFIGURATION.md
func NewKafkaProducer() (*KafkaProducer, error) {
	// Create Kafka producer with minimal configuration
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": server, // Kafka broker address
	})
	if err != nil {
		fmt.Printf("ERROR: Failed to create Kafka producer: %v\n", err)
		return nil, err
	}
	
	// Start background goroutine to handle delivery reports
	go func() {
		// Process events until producer is closed
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				// Message delivery report
				if ev.TopicPartition.Error != nil {
					// Delivery failed
					fmt.Printf("ERROR: Failed to deliver message: %v\n", ev.TopicPartition.Error)
				} else {
					// Delivery succeeded
					fmt.Printf("✓ Successfully produced record to topic %s partition [%d] @ offset %v\n",
						*ev.TopicPartition.Topic, ev.TopicPartition.Partition, ev.TopicPartition.Offset)
				}
			default:
				// Other events (errors, stats, etc.)
				// Can be extended to handle kafka.Error, kafka.Stats, etc.
			}
		}
	}()
	
	return &KafkaProducer{
		Producer:   p,
		topic:      topic,
		chan_event: make(chan kafka.Event, 1000), // Buffered channel for events
	}, nil
}

// KafkaWrite publishes a GPS coordinate message to the Kafka topic.
//
// Message publication:
//   1. Message is queued in producer's internal buffer
//   2. Kafka client sends message asynchronously in background
//   3. Flush() blocks until all buffered messages are delivered
//   4. Delivery confirmation arrives via Events() channel (handled in goroutine)
//
// Parameters:
//   - word: JSON-encoded GPS coordinate data as byte array
//
// Key behaviors:
//   - Asynchronous: Returns before message is actually sent to broker
//   - Partition: Uses automatic partition assignment (kafka.PartitionAny)
//   - Flush timeout: 15 seconds to wait for delivery completion
//   - No explicit error return: Errors handled via Events() channel
//
// Performance considerations:
//   - Flush() on every message reduces throughput (waits for delivery)
//   - For higher throughput, batch multiple Produce() calls before Flush()
//   - Current implementation trades throughput for immediate delivery confirmation
//
// Production improvements:
//   - Remove Flush() for higher throughput (rely on automatic batching)
//   - Add configurable flush interval
//   - Return errors from Flush() to caller
func (p *KafkaProducer) KafkaWrite(word []byte) {
	// Queue message for asynchronous delivery
	p.Producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,                  // Target topic
			Partition: kafka.PartitionAny,      // Auto-assign partition
		},
		Value: word, // Message payload (GPS coordinate JSON)
	}, nil) // nil = don't send on custom delivery channel

	// Wait for all queued messages to be delivered (blocking)
	// Timeout: 15 seconds (15000 milliseconds)
	// Returns number of messages still in queue (0 = all delivered)
	remaining := p.Producer.Flush(15 * 1000)
	if remaining > 0 {
		fmt.Printf("WARNING: %d messages still in queue after flush timeout\n", remaining)
	}
}
