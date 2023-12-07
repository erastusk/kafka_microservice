package main

import (
	"log"

	"github.com/erastusk/gpscords/kafka_reader/kafka"
)

func main() {
	c, err := kafka.NewKafkaConsumer()
	if err != nil {
		log.Println(err)
	}
	err = c.KafkaConsume()
	if err != nil {
		log.Println(err)
	}
}
