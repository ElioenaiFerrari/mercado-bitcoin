package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ElioenaiFerrari/mercado-bitcoin/src/dtos"
	"github.com/ElioenaiFerrari/mercado-bitcoin/src/entities"
	"github.com/ElioenaiFerrari/mercado-bitcoin/src/gateways"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

func hasEvent(events []string, event string) bool {
	for _, e := range events {
		return e == event
	}

	return false
}

func main() {

	mercadoBitcoinApi := gateways.NewMercadoBitcoinApi()
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	router := mux.NewRouter()

	router.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Upgrade", "websocket")
		r.Header.Set("Connection", "Upgrade")
		r.Header.Set("Cache-Control", "no-cache")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		ws, err := upgrader.Upgrade(w, r, nil)

		if err != nil {
			log.Println(err)
			return
		}

		defer ws.Close()

		var subscribeDto dtos.SubscribeDto
		_, p, err := ws.ReadMessage()

		if err != nil {
			log.Println(err)
			return
		}

		if err := json.Unmarshal(p, &subscribeDto); err != nil {
			log.Println(err)
			return
		}

		if subscribeDto.UpdateMs < 200 {
			subscribeDto.UpdateMs = 200
		}

		fmt.Printf("Coin: %s\n", subscribeDto.Coin)
		fmt.Printf("Events: %s\n", subscribeDto.Events)
		fmt.Printf("UpdateMs: %d\n", subscribeDto.UpdateMs)

		channel := make(chan entities.Event)

		for range time.Tick(time.Millisecond * time.Duration(subscribeDto.UpdateMs)) {
			select {
			case event := <-channel:
				if err := ws.WriteJSON(event); err != nil {
					ws.Close()
					return
				}
			default:
				event := entities.Event{
					Type: "",
					Data: "no events",
				}

				jason, err := json.Marshal(event)

				if err != nil {
					log.Println(err)
					return
				}

				ws.WriteMessage(websocket.TextMessage, jason)
			}

			if subscribeDto.Coin != "" {
				if hasEvent(subscribeDto.Events, "orderbook") {
					go func(channel chan entities.Event) {
						orderBook, err := mercadoBitcoinApi.GetOrderBook(subscribeDto.Coin)

						if err != nil {
							return
						}

						event := entities.Event{
							Type: "orderbook",
							Data: orderBook,
						}

						channel <- event
					}(channel)
				}

				if hasEvent(subscribeDto.Events, "trades") {
					go func(channel chan entities.Event) {
						trades, err := mercadoBitcoinApi.GetTrades(subscribeDto.Coin)

						if err != nil {
							return
						}

						event := entities.Event{
							Type: "trades",
							Data: trades,
						}

						channel <- event
					}(channel)
				}

				if hasEvent(subscribeDto.Events, "ticker") {
					go func(channel chan entities.Event) {
						ticker, err := mercadoBitcoinApi.GetTicker(subscribeDto.Coin)

						if err != nil {
							return
						}

						event := entities.Event{
							Type: "ticker",
							Data: ticker,
						}

						channel <- event
					}(channel)
				}
			}
		}

	})

	port := 80

	log.Println(fmt.Sprintf("Listening on port %d", port))

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), router))
}
