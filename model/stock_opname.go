package model

type StockOpname struct {
	ID         int64             `json:"id"`
	DocNo      string            `json:"doc_no"`
	DocDate    string            `json:"doc_date"`
	LocationID int64             `json:"location_id"`
	Status     string            `json:"status"`
	Items      []StockOpnameItem `json:"items"`
}

type StockOpnameItem struct {
	ID          int64    `json:"id"`
	OpnameID    int64    `json:"opname_id"`
	ItemID      int64    `json:"item_id"`
	QtySystem   float64  `json:"qty_system"`
	QtyPhysical *float64 `json:"qty_physical"`
	QtyDiff     float64  `json:"qty_diff"`
}

// Body untuk request "isi hasil hitung"
type CountInput struct {
	Items []CountItem `json:"items"`
}

type CountItem struct {
	ItemID      int64   `json:"item_id"`
	QtyPhysical float64 `json:"qty_physical"`
}
