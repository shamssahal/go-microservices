package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/shamssahal/toll-calculator/types"
)

type DataReceiver struct {
	msgch chan types.OBUData
	conn  *websocket.Conn
}

func (dr *DataReceiver) handleWS(w http.ResponseWriter, r *http.Request) {
	u := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	conn, err := u.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	dr.conn = conn
	go dr.wsReceiveLoop()

}

func (dr *DataReceiver) wsReceiveLoop() {
	fmt.Println("New OBU client connected")
	for {
		var data types.OBUData
		if err := dr.conn.ReadJSON(&data); err != nil {
			log.Println("read error : ", err)
			continue
		}
		fmt.Printf("received obu data from [%d] :: <lat: %.2f long: %.2f> \n", data.OBUID, data.Lat, data.Long)
		dr.msgch <- data
	}
}

func NewDataReceiver() *DataReceiver {
	return &DataReceiver{
		msgch: make(chan types.OBUData, 128),
	}
}

func main() {
	receiver := NewDataReceiver()
	http.HandleFunc("/ws", receiver.handleWS)
	fmt.Println("data receiver running at port :30000")
	http.ListenAndServe(":30000", nil)
}
