package main

import (
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shamssahal/toll-calculator/types"
)

const wsEndpoint = "ws://127.0.0.1:30000/ws"

var sendInterval = time.Millisecond * 100

func genCoord() float64 {
	n := float64(rand.Intn(100) + 1)
	f := rand.Float64()
	return n + f
}

func genLatLong() (float64, float64) {
	return genCoord(), genCoord()
}

func generateOBUIDs(n int) []int {
	ids := make([]int, n)
	for i := range n {
		ids[i] = rand.Intn(math.MaxInt)
	}
	return ids
}

func main() {
	OBUIDs := generateOBUIDs(20)
	conn, _, err := websocket.DefaultDialer.Dial(wsEndpoint, nil)
	if err != nil {
		log.Fatal(err)
	}
	for {
		for i := range OBUIDs {
			currLat, currLong := genLatLong()
			prevLat, prevLong := genLatLong()
			data := types.OBUData{
				OBUID:    OBUIDs[i],
				CurrLat:  currLat,
				CurrLong: currLong,
				PrevLat:  prevLat,
				PrevLong: prevLong,
			}
			if err := conn.WriteJSON(data); err != nil {
				log.Printf("Failed to send data: %v", err)
				return // Exit gracefully instead of log.Fatal
			}
		}
		time.Sleep(sendInterval)
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
