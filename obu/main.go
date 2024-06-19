package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/gorilla/websocket"
	"github.com/opospisil/grpc-microservices-excercise/model"
)

// generator of on-board-unit data to be sent to receiver
const (
	sendInterval = time.Second * 5
	wsEndpoint   = "ws://localhost:30000/ws"
)

func genCoord() float64 {
	n := float64(rand.Intn(100) + 1)
	f := rand.Float64()
	return n + f
}

func genLocation() (float64, float64) {
	return genCoord(), genCoord()
}

func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
}

func generateOBUIDS(n int) []int64 {
	var obuids []int64
	for i := 0; i < n; i++ {
		obuids = append(obuids, int64(rand.Intn(math.MaxInt)))
	}

	return obuids
}

func main() {
	obuIDS := generateOBUIDS(3)
	conn, _, err := websocket.DefaultDialer.Dial(wsEndpoint, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}

	for {
		for i := 0; i < len(obuIDS); i++ {
			lat, lon := genLocation()
			data := model.OBUData{
				OBUID: obuIDS[i],
				Lat:   lat,
				Lon:   lon,
			}
			fmt.Printf("OBU Data: %+v\n", data)
			if err := conn.WriteJSON(data); err != nil {
				log.Fatal(err)
			}
		}
		time.Sleep(sendInterval)
	}
}
