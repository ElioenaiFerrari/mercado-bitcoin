package gateways

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/ElioenaiFerrari/mercado-bitcoin/src/entities"
)

type MercadoBitcoinApi struct {
	Url string `json:"url"`
}

func NewMercadoBitcoinApi() *MercadoBitcoinApi {

	return &MercadoBitcoinApi{
		Url: os.Getenv("API_URL"),
	}
}

func (mba *MercadoBitcoinApi) GetOrderBook(coin string) (entities.OrderBook, error) {
	var orderBook entities.OrderBook
	res, err := http.Get(fmt.Sprintf("%s/%s/orderbook", mba.Url, coin))

	if err != nil {
		return orderBook, err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return orderBook, err
	}

	if err := json.Unmarshal(body, &orderBook); err != nil {
		return orderBook, err
	}

	return orderBook, nil

}

func (mba *MercadoBitcoinApi) GetTrades(coin string) ([]entities.Trade, error) {
	var trades []entities.Trade
	res, err := http.Get(fmt.Sprintf("%s/%s/trades", mba.Url, coin))

	if err != nil {
		return trades, err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return trades, err
	}

	if err := json.Unmarshal(body, &trades); err != nil {
		return trades, err
	}

	return trades, nil

}

func (mba *MercadoBitcoinApi) GetTicker(coin string) (entities.Ticker, error) {
	type Response struct {
		Ticker entities.Ticker `json:"ticker"`
	}

	var response Response

	res, err := http.Get(fmt.Sprintf("%s/%s/ticker", mba.Url, coin))

	if err != nil {
		return response.Ticker, err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return response.Ticker, err
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return response.Ticker, err
	}

	return response.Ticker, nil

}
