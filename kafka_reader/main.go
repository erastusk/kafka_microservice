// Package main implements a Kafka consumer that reads GPS coordinate messages
// from a Kafka topic and processes them.
//
// Architecture:
//
//	Kafka Broker → Kafka Consumer (this) → Business Logic/Display
//
// This consumer:
//   - Connects to Kafka broker
//   - Subscribes to 'gpscoords' topic
//   - Continuously polls for new messages
//   - Deserializes GPS coordinates from JSON
//   - Prints coordinates to stdout (can be extended for further processing)
package main

import (
	"log"

	"github.com/erastusk/gpscords/kafka_reader/kafka"
)

// main initializes the Kafka consumer and starts consuming messages.
//
// Error handling:
//   - Consumer creation errors are logged but don't stop execution
//   - Consumption errors are logged and execution continues
//   - Consider adding retry logic and graceful error handling for production
func main() {
	log.Println("Starting Kafka GPS consumer...")

	// Create new Kafka consumer instance
	c, err := kafka.NewKafkaConsumer()
	if err != nil {
		log.Printf("ERROR: Failed to create Kafka consumer: %v", err)
		return
	}

	log.Println("Consumer created successfully, starting to consume messages...")

	// Start consuming messages (blocks indefinitely)
	err = c.KafkaConsume()
	if err != nil {
		log.Printf("ERROR: Consumer error: %v", err)
	}
}
