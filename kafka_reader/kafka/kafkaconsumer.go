package kafka

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"

	"github.com/erastusk/gpscords/types"
)

var (
	server       = "gpscords_app-kafka-1:9092"
	topic        = "gpscoords"
	offset_reset = "earliest"
	group_id     = "gps"
)

type KafkaConsumer struct {
	Consumer *kafka.Consumer
	topic    string
	msgChan  chan types.SourceCoords
}

func NewKafkaConsumer() (*KafkaConsumer, error) {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": server,
		"auto.offset.reset": offset_reset,
		"group.id":          group_id,
	})
	if err != nil {
		log.Fatal("Couldn't create a consumer", err)
	}
	return &KafkaConsumer{
		Consumer: c,
		topic:    topic,
		msgChan:  make(chan types.SourceCoords),
	}, nil
}

func (c *KafkaConsumer) KafkaConsume() error {
	err := c.Consumer.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		log.Fatal("subscribe topics failed", err)
		return err
	}
	go kafkaconsumeLoop(c)
	for a := range c.msgChan {
		fmt.Printf("Kafka consumer : %+v\n", a)
	}
	return nil
}

//	func kafkaconsumeLoop(c *KafkaConsumer) {
//		t := types.SourceCoords{}
//		defer c.Consumer.Close()
//		for {
//			msg, err := c.Consumer.ReadMessage(-1)
//			if err != nil {
//				fmt.Println("Could not read messages", err)
//				log.Fatal(err)
//			}
//			err = json.Unmarshal(msg.Value, &t)
//
//			if err != nil {
//				log.Println("Couldn't unmarshal message", err)
//			}
//			c.msgChan <- t
//		}
//	}
func kafkaconsumeLoop(c *KafkaConsumer) {
	defer c.Consumer.Close()
	t := types.SourceCoords{}
	run := true
	for run == true {
		ev := c.Consumer.Poll(100)
		switch e := ev.(type) {
		case *kafka.Message:
			// application-specific processing
			log.Printf("%v", e.Headers)
			err := json.Unmarshal(e.Value, &t)
			//
			if err != nil {
				log.Println("Couldn't unmarshal message", err)
			}
			c.msgChan <- t
		case kafka.Error:
			fmt.Fprintf(os.Stderr, "%% Error: %v\n", e)
			run = false
		}
	}
}
