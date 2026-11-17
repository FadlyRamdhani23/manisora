package model

type Sale struct {
	ID         int64      `json:"id"`
	DocNo      string     `json:"doc_no"`
	DocDate    string     `json:"doc_date"`
	LocationID int64      `json:"location_id"`
	Status     string     `json:"status"`
	Items      []SaleItem `json:"items"`
}

type SaleItem struct {
	ID     int64   `json:"id"`
	SaleID int64   `json:"sale_id"`
	ItemID int64   `json:"item_id"`
	Qty    float64 `json:"qty"`
}
