package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/gorilla/websocket"

	"github.com/erastusk/gpscords/types"
)

func retOBUdata() (int, float64, float64) {
	return rand.Int(), rand.Float64()*99 + 1, rand.Float64()*99 + 1
}

var wsEndpoint = "ws://localhost:30000/ws"

func main() {
	conn, _, err := websocket.DefaultDialer.Dial(wsEndpoint, nil)
	if err != nil {
		log.Println("Couldn't dial", err)
	}
	for {
		a, b, c := MiddlewareReceiver(retOBUdata)
		t := types.SourceCoords{
			OBUID: a,
			Lat:   b,
			Lon:   c,
		}
		time.Sleep(time.Second)
		fmt.Printf("Producer: %+v\n", t)
		err = conn.WriteJSON(t)
		if err != nil {
			log.Println("Unable to write message")
		}
	}
}
