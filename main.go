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
	mercadoBitcoinApi := gateways.NewMercadoBitcoinApi()
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	router := mux.NewRouter()

	router.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Upgrade", "websocket")
		r.Header.Set("Connection", "Upgrade")

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

			if coinDto.Coin != "" {
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

				select {
				case event := <-channel:
					if err := ws.WriteJSON(event); err != nil {
						ws.Close()
						return
					}
				default:
					fmt.Println("no event")
				}
			}

		}

	})

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding")
		availableCoinsString := os.Getenv("AVAILABLE_COINS")
		availableCoins := strings.Split(availableCoinsString, ",")

		jason, err := json.Marshal(availableCoins)

		if err != nil {
			log.Println(err)
			return
		}

		w.Write(jason)
	})
	port := os.Getenv("PORT")

	log.Println(fmt.Sprintf("Listening on port %s", port))

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), router))
}
