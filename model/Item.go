package model

type Item struct {
	ID       int64  `json:"id"`
	Code     string `json:"code"`
	Name     string `json:"name"`
	BaseUnit string `json:"base_unit"`
}
