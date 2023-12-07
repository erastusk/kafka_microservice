package handlers

import (
	"log"
	"time"

	"github.com/erastusk/gpscords/data_receiver_kafka_producer/kafka"
)

func MiddlewareRead(w []byte, t *kafka.KafkaProducer) {
	start := time.Now()
	defer func() {
		log.Println("Writing to Kafka took: ", time.Since(start))
	}()
	t.KafkaWrite(w)
}
