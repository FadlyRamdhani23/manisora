package repository

import (
	"database/sql"
	"fmt"
	"time"

	"manisora/database"
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

func GetAllGoodsReceipts() ([]model.GoodsReceipt, error) {
	rows, err := database.DB.Query(
		"SELECT id, doc_no, DATE_FORMAT(doc_date, '%Y-%m-%d'), supplier_id, location_id, status FROM goods_receipts ORDER BY id DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []model.GoodsReceipt{}
	for rows.Next() {
		var gr model.GoodsReceipt
		if err := rows.Scan(&gr.ID, &gr.DocNo, &gr.DocDate, &gr.SupplierID, &gr.LocationID, &gr.Status); err != nil {
			return nil, err
		}
		gr.Items = []model.GoodsReceiptItem{}
		list = append(list, gr)
	}
	return list, rows.Err()
}

func GetGoodsReceiptByID(id int64) (model.GoodsReceipt, error) {
	var gr model.GoodsReceipt
	err := database.DB.QueryRow(
		"SELECT id, doc_no, DATE_FORMAT(doc_date, '%Y-%m-%d'), supplier_id, location_id, status FROM goods_receipts WHERE id = ?",
		id,
	).Scan(&gr.ID, &gr.DocNo, &gr.DocDate, &gr.SupplierID, &gr.LocationID, &gr.Status)
	if err != nil {
		return gr, err
	}

	rows, err := database.DB.Query(
		"SELECT id, receipt_id, item_id, qty FROM goods_receipt_items WHERE receipt_id = ? ORDER BY id",
		id,
	)
	if err != nil {
		return gr, err
	}
	defer rows.Close()

	gr.Items = []model.GoodsReceiptItem{}
	for rows.Next() {
		var it model.GoodsReceiptItem
		if err := rows.Scan(&it.ID, &it.ReceiptID, &it.ItemID, &it.Qty); err != nil {
			return gr, err
		}
		gr.Items = append(gr.Items, it)
	}
	return gr, rows.Err()
}

// Mengunci baris dokumen sampai transaksi selesai
func LockGoodsReceipt(tx *sql.Tx, id int64) (status string, locationID int64, err error) {
	err = tx.QueryRow(
		"SELECT status, location_id FROM goods_receipts WHERE id = ? FOR UPDATE",
		id,
	).Scan(&status, &locationID)
	return
}

func GetGoodsReceiptItemsTx(tx *sql.Tx, receiptID int64) ([]model.GoodsReceiptItem, error) {
	rows, err := tx.Query(
		"SELECT id, receipt_id, item_id, qty FROM goods_receipt_items WHERE receipt_id = ? ORDER BY id",
		receiptID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []model.GoodsReceiptItem{}
	for rows.Next() {
		var it model.GoodsReceiptItem
		if err := rows.Scan(&it.ID, &it.ReceiptID, &it.ItemID, &it.Qty); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

func SetGoodsReceiptStatus(tx *sql.Tx, id int64, status string) error {
	_, err := tx.Exec("UPDATE goods_receipts SET status = ? WHERE id = ?", status, id)
	return err
}
