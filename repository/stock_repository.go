package repository

import "database/sql"

func InsertLedger(tx *sql.Tx, itemID, locationID int64, qtyIn, qtyOut float64,
	movementType, refType string, refID int64) error {
	_, err := tx.Exec(
		`INSERT INTO stock_ledger
		 (trx_date, item_id, location_id, qty_in, qty_out, movement_type, ref_type, ref_id)
		 VALUES (NOW(), ?, ?, ?, ?, ?, ?, ?)`,
		itemID, locationID, qtyIn, qtyOut, movementType, refType, refID,
	)
	return err
}

// delta positif = stok bertambah, negatif = stok berkurang
func AddStockBalance(tx *sql.Tx, itemID, locationID int64, delta float64) error {
	_, err := tx.Exec(
		`INSERT INTO stock_balances (item_id, location_id, qty_on_hand)
		 VALUES (?, ?, ?)
		 ON DUPLICATE KEY UPDATE qty_on_hand = qty_on_hand + ?`,
		itemID, locationID, delta, delta,
	)
	return err
}
