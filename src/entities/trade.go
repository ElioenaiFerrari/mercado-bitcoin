package entities

type Trade struct {
	Amount float64 `json:"amount"`
	Date   int32   `json:"date"`
	Price  float64 `json:"price"`
	Tid    int64   `json:"tid"`
	Type   string  `json:"type"`
}
