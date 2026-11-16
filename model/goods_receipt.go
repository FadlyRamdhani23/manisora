package model

type GoodsReceipt struct {
	ID         int64              `json:"id"`
	DocNo      string             `json:"doc_no"`
	DocDate    string             `json:"doc_date"`
	SupplierID int64              `json:"supplier_id"`
	LocationID int64              `json:"location_id"`
	Status     string             `json:"status"`
	Items      []GoodsReceiptItem `json:"items"`
}

type GoodsReceiptItem struct {
	ID        int64   `json:"id"`
	ReceiptID int64   `json:"receipt_id"`
	ItemID    int64   `json:"item_id"`
	Qty       float64 `json:"qty"`
}
