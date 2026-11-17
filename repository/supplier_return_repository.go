package repository

import (
	"database/sql"
	"fmt"
	"time"

	"manisora/database"
	"manisora/model"
)

func CreateSupplierReturn(tx *sql.Tx, r model.SupplierReturn) (int64, error) {
	tempNo := fmt.Sprintf("TMP-%d", time.Now().UnixNano())

	var receipt interface{} // nil akan tersimpan sebagai NULL
	if r.ReceiptID > 0 {
		receipt = r.ReceiptID
	}

	result, err := tx.Exec(
		"INSERT INTO supplier_returns (doc_no, doc_date, supplier_id, location_id, receipt_id, status) VALUES (?, ?, ?, ?, ?, ?)",
		tempNo, r.DocDate, r.SupplierID, r.LocationID, receipt, r.Status,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func SetSupplierReturnDocNo(tx *sql.Tx, id int64, docNo string) error {
	_, err := tx.Exec("UPDATE supplier_returns SET doc_no = ? WHERE id = ?", docNo, id)
	return err
}

func CreateSupplierReturnItem(tx *sql.Tx, returnID int64, it model.SupplierReturnItem) (int64, error) {
	result, err := tx.Exec(
		"INSERT INTO supplier_return_items (return_id, item_id, qty, reason) VALUES (?, ?, ?, ?)",
		returnID, it.ItemID, it.Qty, it.Reason,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func LockSupplierReturn(tx *sql.Tx, id int64) (status string, supplierID, locationID, receiptID int64, err error) {
	var receipt sql.NullInt64
	err = tx.QueryRow(
		"SELECT status, supplier_id, location_id, receipt_id FROM supplier_returns WHERE id = ? FOR UPDATE",
		id,
	).Scan(&status, &supplierID, &locationID, &receipt)
	if receipt.Valid {
		receiptID = receipt.Int64
	}
	return
}

func GetSupplierReturnItemsTx(tx *sql.Tx, returnID int64) ([]model.SupplierReturnItem, error) {
	rows, err := tx.Query(
		"SELECT id, return_id, item_id, qty, reason FROM supplier_return_items WHERE return_id = ? ORDER BY id",
		returnID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []model.SupplierReturnItem{}
	for rows.Next() {
		var it model.SupplierReturnItem
		if err := rows.Scan(&it.ID, &it.ReturnID, &it.ItemID, &it.Qty, &it.Reason); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

func SetSupplierReturnStatus(tx *sql.Tx, id int64, status string) error {
	_, err := tx.Exec("UPDATE supplier_returns SET status = ? WHERE id = ?", status, id)
	return err
}

// --- Fungsi untuk aturan "retur tidak boleh melebihi penerimaan" ---

// Mengunci baris penerimaan acuan, supaya dua retur yang diposting
// bersamaan terhadap penerimaan yang sama tidak bisa saling melampaui batas
func LockReceiptForReturn(tx *sql.Tx, id int64) (status string, supplierID int64, err error) {
	err = tx.QueryRow(
		"SELECT status, supplier_id FROM goods_receipts WHERE id = ? FOR UPDATE",
		id,
	).Scan(&status, &supplierID)
	return
}

func GetReceivedQty(tx *sql.Tx, receiptID, itemID int64) (float64, error) {
	var qty float64
	err := tx.QueryRow(
		"SELECT COALESCE(SUM(qty), 0) FROM goods_receipt_items WHERE receipt_id = ? AND item_id = ?",
		receiptID, itemID,
	).Scan(&qty)
	return qty, err
}

func GetReturnedQty(tx *sql.Tx, receiptID, itemID int64) (float64, error) {
	var qty float64
	err := tx.QueryRow(
		`SELECT COALESCE(SUM(ri.qty), 0)
		 FROM supplier_return_items ri
		 JOIN supplier_returns r ON r.id = ri.return_id
		 WHERE r.receipt_id = ? AND ri.item_id = ? AND r.status = 'POSTED'`,
		receiptID, itemID,
	).Scan(&qty)
	return qty, err
}

func GetSupplierReturnByID(id int64) (model.SupplierReturn, error) {
	var r model.SupplierReturn
	var receipt sql.NullInt64
	err := database.DB.QueryRow(
		"SELECT id, doc_no, DATE_FORMAT(doc_date, '%Y-%m-%d'), supplier_id, location_id, receipt_id, status FROM supplier_returns WHERE id = ?",
		id,
	).Scan(&r.ID, &r.DocNo, &r.DocDate, &r.SupplierID, &r.LocationID, &receipt, &r.Status)
	if err != nil {
		return r, err
	}
	if receipt.Valid {
		r.ReceiptID = receipt.Int64
	}

	rows, err := database.DB.Query(
		"SELECT id, return_id, item_id, qty, reason FROM supplier_return_items WHERE return_id = ? ORDER BY id",
		id,
	)
	if err != nil {
		return r, err
	}
	defer rows.Close()

	r.Items = []model.SupplierReturnItem{}
	for rows.Next() {
		var it model.SupplierReturnItem
		if err := rows.Scan(&it.ID, &it.ReturnID, &it.ItemID, &it.Qty, &it.Reason); err != nil {
			return r, err
		}
		r.Items = append(r.Items, it)
	}
	return r, rows.Err()
}

func GetAllSupplierReturns() ([]model.SupplierReturn, error) {
	rows, err := database.DB.Query(
		"SELECT id, doc_no, DATE_FORMAT(doc_date, '%Y-%m-%d'), supplier_id, location_id, receipt_id, status FROM supplier_returns ORDER BY id DESC LIMIT 100",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []model.SupplierReturn{}
	for rows.Next() {
		var r model.SupplierReturn
		var receipt sql.NullInt64
		if err := rows.Scan(&r.ID, &r.DocNo, &r.DocDate, &r.SupplierID, &r.LocationID, &receipt, &r.Status); err != nil {
			return nil, err
		}
		if receipt.Valid {
			r.ReceiptID = receipt.Int64
		}
		r.Items = []model.SupplierReturnItem{}
		list = append(list, r)
	}
	return list, rows.Err()
}
