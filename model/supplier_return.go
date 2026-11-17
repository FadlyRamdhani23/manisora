package model

type SupplierReturn struct {
	ID         int64                `json:"id"`
	DocNo      string               `json:"doc_no"`
	DocDate    string               `json:"doc_date"`
	SupplierID int64                `json:"supplier_id"`
	LocationID int64                `json:"location_id"`
	ReceiptID  int64                `json:"receipt_id,omitempty"`
	Status     string               `json:"status"`
	Items      []SupplierReturnItem `json:"items"`
}

type SupplierReturnItem struct {
	ID       int64   `json:"id"`
	ReturnID int64   `json:"return_id"`
	ItemID   int64   `json:"item_id"`
	Qty      float64 `json:"qty"`
	Reason   string  `json:"reason"`
}
