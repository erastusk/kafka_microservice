// Package handlers provides middleware functions for request processing.
package handlers

import (
	"log"
	"time"

	"github.com/erastusk/gpscords/data_receiver_kafka_producer/kafka"
)

// MiddlewareRead wraps Kafka write operations with timing instrumentation.
//
// This middleware:
//   - Records start time before Kafka write
//   - Executes Kafka write operation
//   - Calculates and logs write latency
//
// Use case:
//   - Performance monitoring: track how long Kafka writes take
//   - Identifying bottlenecks in the data pipeline
//   - Setting up alerts for slow Kafka operations
//
// Parameters:
//   - w: Byte array containing JSON-encoded GPS coordinate data
//   - t: KafkaProducer instance to write data to
//
// The timing measurement includes:
//   - Kafka message serialization
//   - Network transmission to Kafka broker
//   - Waiting for producer flush (delivery confirmation)
func MiddlewareRead(w []byte, t *kafka.KafkaProducer) {
	// Record start time for latency measurement
	start := time.Now()
	
	// Defer logging to ensure it runs after Kafka write completes
	defer func() {
		log.Printf("Writing to Kafka took: %v", time.Since(start))
	}()
	
	// Write GPS data to Kafka topic
	t.KafkaWrite(w)
}
