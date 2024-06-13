package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/opospisil/grpc-microservices-excercise/model"
)

const kafkaTopic = "obu-data"

func main() {
	receiver, err := NewDataReceiver()
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/ws", receiver.handleWS)
	http.ListenAndServe(":30000", nil)
}

type DataReceiver struct {
	msgChan  chan model.OBUData
	conn     *websocket.Conn
	producer DataProducer
}

func NewDataReceiver() (*DataReceiver, error) {
	var (
		p   DataProducer
		err error
	)
	p, err = NewKafkaProducer(kafkaTopic)
	if err != nil {
		return nil, err
	}

	p = NewLogMiddleware(p)

	return &DataReceiver{
		msgChan:  make(chan model.OBUData, 128),
		producer: p,
	}, nil
}

func (dr *DataReceiver) produceData(data model.OBUData) error {
	return dr.producer.Produce(data)
}

func (dr *DataReceiver) handleWS(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	dr.conn = conn

	go dr.wsReceiveLoop()
}

func (dr *DataReceiver) wsReceiveLoop() {
	fmt.Println("Received OBU connection")
	for {
		var data model.OBUData
		if err := dr.conn.ReadJSON(&data); err != nil {
			log.Println(err)
			continue
		}
		dr.produceData(data)
	}
}
