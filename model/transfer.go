package model

type Transfer struct {
	ID             int64          `json:"id"`
	DocNo          string         `json:"doc_no"`
	DocDate        string         `json:"doc_date"`
	FromLocationID int64          `json:"from_location_id"`
	ToLocationID   int64          `json:"to_location_id"`
	Status         string         `json:"status"`
	Items          []TransferItem `json:"items"`
}

type TransferItem struct {
	ID          int64   `json:"id"`
	TransferID  int64   `json:"transfer_id"`
	ItemID      int64   `json:"item_id"`
	QtySent     float64 `json:"qty_sent"`
	QtyReceived float64 `json:"qty_received"`
	QtyDiff     float64 `json:"qty_diff"`
}

// Body untuk request "terima"
type ReceiveInput struct {
	Items []ReceiveItem `json:"items"`
}

type ReceiveItem struct {
	ItemID      int64   `json:"item_id"`
	QtyReceived float64 `json:"qty_received"`
}
