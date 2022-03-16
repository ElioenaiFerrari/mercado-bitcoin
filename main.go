package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ElioenaiFerrari/mercado-bitcoin/src/dtos"
	"github.com/ElioenaiFerrari/mercado-bitcoin/src/entities"
	"github.com/ElioenaiFerrari/mercado-bitcoin/src/gateways"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

func main() {
	listenEvents := os.Getenv("LISTEN_EVENTS")

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

		var coinDto dtos.CoinDto
		_, p, err := ws.ReadMessage()

		if err != nil {
			log.Println(err)
			return
		}

		if err := json.Unmarshal(p, &coinDto); err != nil {
			log.Println(err)
			return
		}

		fmt.Printf("Coin: %s\n", coinDto.Coin)

		channel := make(chan entities.Event)
		for range time.Tick(time.Millisecond * 200) {
			select {
			case event := <-channel:
				if err := ws.WriteJSON(event); err != nil {
					ws.Close()
					return
				}
			default:
				fmt.Println("no event")
			}

			if coinDto.Coin != "" {
				if strings.Contains(listenEvents, "orderbook") {
					go func(channel chan entities.Event) {
						orderBook, err := mercadoBitcoinApi.GetOrderBook(coinDto.Coin)

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

				if strings.Contains(listenEvents, "trades") {
					go func(channel chan entities.Event) {
						trades, err := mercadoBitcoinApi.GetTrades(coinDto.Coin)

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

				if strings.Contains(listenEvents, "ticker") {
					go func(channel chan entities.Event) {
						ticker, err := mercadoBitcoinApi.GetTicker(coinDto.Coin)

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

	fmt.Printf("Listen events: %s\n", listenEvents)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), router))
}
