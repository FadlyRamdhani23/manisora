package repository

import (
	"database/sql"
	"fmt"
	"time"

	"manisora/model"
)

func CreateGoodsReceipt(tx *sql.Tx, gr model.GoodsReceipt) (int64, error) {
	tempNo := fmt.Sprintf("TMP-%d", time.Now().UnixNano())
	result, err := tx.Exec(
		"INSERT INTO goods_receipts (doc_no, doc_date, supplier_id, location_id, status) VALUES (?, ?, ?, ?, ?)",
		tempNo, gr.DocDate, gr.SupplierID, gr.LocationID, gr.Status,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func SetGoodsReceiptDocNo(tx *sql.Tx, id int64, docNo string) error {
	_, err := tx.Exec("UPDATE goods_receipts SET doc_no = ? WHERE id = ?", docNo, id)
	return err
}

func CreateGoodsReceiptItem(tx *sql.Tx, receiptID int64, it model.GoodsReceiptItem) (int64, error) {
	result, err := tx.Exec(
		"INSERT INTO goods_receipt_items (receipt_id, item_id, qty) VALUES (?, ?, ?)",
		receiptID, it.ItemID, it.Qty,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}
