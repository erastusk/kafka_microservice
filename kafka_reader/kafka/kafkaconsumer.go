// Package kafka provides Kafka consumer functionality for reading GPS coordinates.
// Uses the Confluent Kafka Go client for consuming messages from Kafka topics.
package kafka

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"

	"github.com/erastusk/gpscords/types"
)

// Kafka consumer configuration variables.
var (
	// server is the Kafka broker address.
	// Uses Docker Compose service name for container-to-container communication.
	server = "gpscords_app-kafka-1:9092"
	
	// topic is the Kafka topic to consume messages from.
	topic = "gpscoords"
	
	// offset_reset determines where to start reading if no offset exists.
	// "earliest" = read from beginning of topic
	// "latest" = read only new messages
	offset_reset = "earliest"
	
	// group_id identifies the consumer group for coordinated consumption.
	// Consumers with same group_id share topic partitions.
	group_id = "gps"
)

// KafkaConsumer wraps a Kafka consumer with message processing capabilities.
//
// Fields:
//   - Consumer: Confluent Kafka consumer instance for reading messages
//   - topic: Kafka topic name to subscribe to
//   - msgChan: Channel for delivering parsed GPS coordinates to application
//
// Architecture:
//   Kafka Topic → Consumer (goroutine) → msgChan → Application Logic
//
// The consumer uses a goroutine to poll Kafka and sends parsed messages
// to msgChan for processing by the main application.
type KafkaConsumer struct {
	Consumer *kafka.Consumer            // Kafka consumer instance
	topic    string                     // Topic to consume from
	msgChan  chan types.SourceCoords    // Channel for delivering messages
}

// NewKafkaConsumer creates and initializes a new Kafka consumer.
//
// Consumer Configuration:
//   - bootstrap.servers: Kafka broker address(es)
//   - auto.offset.reset: Where to start reading (earliest/latest)
//   - group.id: Consumer group for coordinated consumption
//
// Consumer Group behavior:
//   - Multiple consumers with same group_id share partition assignment
//   - Each partition consumed by only one consumer in the group
//   - Enables horizontal scaling and fault tolerance
//
// Returns:
//   - *KafkaConsumer: Configured consumer ready to subscribe
//   - error: Connection or configuration errors
//
// Note: Consumer must call SubscribeTopics() before reading messages.
func NewKafkaConsumer() (*KafkaConsumer, error) {
	// Create Kafka consumer with configuration
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": server,       // Broker address
		"auto.offset.reset": offset_reset, // Where to start reading
		"group.id":          group_id,     // Consumer group ID
	})
	if err != nil {
		log.Printf("ERROR: Couldn't create Kafka consumer: %v", err)
		return nil, err
	}
	
	return &KafkaConsumer{
		Consumer: c,
		topic:    topic,
		msgChan:  make(chan types.SourceCoords), // Unbuffered channel
	}, nil
}

// KafkaConsume subscribes to the topic and starts consuming messages.
//
// Flow:
//   1. Subscribe to GPS coordinates topic
//   2. Start background goroutine to poll Kafka
//   3. Block on msgChan, printing each received coordinate
//
// The background goroutine:
//   - Polls Kafka every 100ms for new messages
//   - Deserializes JSON messages into SourceCoords structs
//   - Sends parsed coordinates to msgChan
//   - Handles errors and closes consumer on fatal errors
//
// Returns:
//   - error: Subscription or consumption errors
//
// This function blocks indefinitely until consumer is closed or error occurs.
// GPS coordinates are printed to stdout as they arrive.
func (c *KafkaConsumer) KafkaConsume() error {
	// Subscribe to GPS coordinates topic
	err := c.Consumer.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		log.Printf("ERROR: Failed to subscribe to topics: %v", err)
		return err
	}
	
	log.Printf("Subscribed to topic '%s' in consumer group '%s'", topic, group_id)
	log.Println("Consuming messages...")
	
	// Start background goroutine to poll Kafka
	go kafkaconsumeLoop(c)
	
	// Process messages from channel (blocks until channel closed)
	for a := range c.msgChan {
		fmt.Printf("Kafka consumer received: OBUID=%d, Lat=%.2f, Lon=%.2f\n", a.OBUID, a.Lat, a.Lon)
		// In production, process coordinates here:
		// - Store in database
		// - Calculate distance/speed
		// - Trigger location-based events
		// - Forward to other services
	}
	
	return nil
}

// kafkaconsumeLoop continuously polls Kafka for messages in a background goroutine.
//
// Polling implementation (event-based):
//   - Poll() checks for new events every 100ms
//   - Returns *kafka.Message for successful message delivery
//   - Returns kafka.Error for errors (broker unavailable, etc.)
//   - Returns nil on timeout (no new messages)
//
// Message processing:
//   1. Receive kafka.Message event
//   2. Log message headers for debugging
//   3. Unmarshal JSON payload into SourceCoords
//   4. Send parsed struct to msgChan
//   5. Repeat until error occurs
//
// Error handling:
//   - JSON unmarshal errors are logged but don't stop consumption
//   - Kafka errors terminate the loop and close consumer
//   - Consumer is always closed via defer
//
// Alternative implementation (commented out):
//   - Uses ReadMessage() for simpler blocking reads
//   - Less control over polling timeout
//   - Current polling approach preferred for more control
//
// Parameters:
//   - c: KafkaConsumer instance with active subscription
func kafkaconsumeLoop(c *KafkaConsumer) {
	// Ensure consumer is closed when goroutine exits
	defer c.Consumer.Close()
	
	var t types.SourceCoords
	run := true
	
	// Poll Kafka until error or termination
	for run {
		// Poll for events with 100ms timeout
		ev := c.Consumer.Poll(100)
		
		// Handle different event types
		switch e := ev.(type) {
		case *kafka.Message:
			// New message received
			log.Printf("Message headers: %v", e.Headers)
			
			// Deserialize JSON message to SourceCoords
			err := json.Unmarshal(e.Value, &t)
			if err != nil {
				log.Printf("ERROR: Couldn't unmarshal message: %v", err)
				continue // Skip this message but continue consuming
			}
			
			// Send parsed coordinate to channel
			c.msgChan <- t
			
		case kafka.Error:
			// Fatal error occurred (broker down, authentication failed, etc.)
			fmt.Fprintf(os.Stderr, "ERROR: Kafka consumer error: %v\n", e)
			run = false // Terminate loop
			
		case nil:
			// Poll timeout - no new messages
			// This is normal; continue polling
		}
	}
	
	log.Println("Consumer loop terminated")
}
