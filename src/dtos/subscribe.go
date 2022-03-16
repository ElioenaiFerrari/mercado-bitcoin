package dtos

type SubscribeDto struct {
	Coin     string   `json:"coin"`
	Events   []string `json:"events"`
	UpdateMs int64    `json:"updateMs"`
}
