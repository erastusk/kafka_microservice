package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/erastusk/gpscords/data_receiver_kafka_producer/kafka"
	"github.com/erastusk/gpscords/types"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1028,
	WriteBufferSize: 1028,
}

func ReceiveWs(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	ReadMessageLoop(c)
}

func ReadMessageLoop(c *websocket.Conn) {
	defer c.Close()
	k, err := kafka.NewKafkaProducer()
	if err != nil {
		fmt.Println(err)
	}
	var recv types.SourceCoords
	for {
		err := c.ReadJSON(&recv)
		if err != nil {
			websocket.IsUnexpectedCloseError(err,
				websocket.CloseAbnormalClosure,
				websocket.CloseGoingAway,
			)
			log.Println("Unexpected closure")
			break

		}
		log.Printf("kafka receiver: %v", recv)
		resp, err := json.Marshal(recv)
		log.Println(string(resp))
		MiddlewareRead(resp, k)
	}
}
