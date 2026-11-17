package model

type StockMutation struct {
	ItemID    int64             `json:"item_id"`
	Code      string            `json:"code"`
	Name      string            `json:"name"`
	BaseUnit  string            `json:"base_unit"`
	Opening   float64           `json:"opening"`
	QtyIn     float64           `json:"qty_in"`
	QtyOut    float64           `json:"qty_out"`
	Closing   float64           `json:"closing"`
	Breakdown MutationBreakdown `json:"breakdown"`
}

type MutationBreakdown struct {
	PurchaseReceipt float64 `json:"purchase_receipt"`
	TransferIn      float64 `json:"transfer_in"`
	Sale            float64 `json:"sale"`
	TransferOut     float64 `json:"transfer_out"`
	SupplierReturn  float64 `json:"supplier_return"`
	Opname          float64 `json:"opname"` // angka bersih: plus = lebih, minus = kurang
}

type StockMutationReport struct {
	LocationID int64           `json:"location_id"`
	From       string          `json:"from"`
	To         string          `json:"to"`
	Items      []StockMutation `json:"items"`
}
