package repository

import (
	"database/sql"
	"fmt"
	"time"

	"manisora/database"
	"manisora/model"
)

func CreateTransfer(tx *sql.Tx, t model.Transfer) (int64, error) {
	tempNo := fmt.Sprintf("TMP-%d", time.Now().UnixNano())
	result, err := tx.Exec(
		"INSERT INTO transfers (doc_no, doc_date, from_location_id, to_location_id, status) VALUES (?, ?, ?, ?, ?)",
		tempNo, t.DocDate, t.FromLocationID, t.ToLocationID, t.Status,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func SetTransferDocNo(tx *sql.Tx, id int64, docNo string) error {
	_, err := tx.Exec("UPDATE transfers SET doc_no = ? WHERE id = ?", docNo, id)
	return err
}

func CreateTransferItem(tx *sql.Tx, transferID int64, it model.TransferItem) (int64, error) {
	result, err := tx.Exec(
		"INSERT INTO transfer_items (transfer_id, item_id, qty_sent) VALUES (?, ?, ?)",
		transferID, it.ItemID, it.QtySent,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func LockTransfer(tx *sql.Tx, id int64) (status string, fromID, toID int64, err error) {
	err = tx.QueryRow(
		"SELECT status, from_location_id, to_location_id FROM transfers WHERE id = ? FOR UPDATE",
		id,
	).Scan(&status, &fromID, &toID)
	return
}

func GetTransferItemsTx(tx *sql.Tx, transferID int64) ([]model.TransferItem, error) {
	rows, err := tx.Query(
		"SELECT id, transfer_id, item_id, qty_sent, qty_received, qty_diff FROM transfer_items WHERE transfer_id = ? ORDER BY id",
		transferID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []model.TransferItem{}
	for rows.Next() {
		var it model.TransferItem
		if err := rows.Scan(&it.ID, &it.TransferID, &it.ItemID, &it.QtySent, &it.QtyReceived, &it.QtyDiff); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

func SetTransferStatus(tx *sql.Tx, id int64, status string) error {
	_, err := tx.Exec("UPDATE transfers SET status = ? WHERE id = ?", status, id)
	return err
}

func SetTransferItemReceived(tx *sql.Tx, rowID int64, received, diff float64) error {
	_, err := tx.Exec(
		"UPDATE transfer_items SET qty_received = ?, qty_diff = ? WHERE id = ?",
		received, diff, rowID,
	)
	return err
}

func GetTransferByID(id int64) (model.Transfer, error) {
	var t model.Transfer
	err := database.DB.QueryRow(
		"SELECT id, doc_no, DATE_FORMAT(doc_date, '%Y-%m-%d'), from_location_id, to_location_id, status FROM transfers WHERE id = ?",
		id,
	).Scan(&t.ID, &t.DocNo, &t.DocDate, &t.FromLocationID, &t.ToLocationID, &t.Status)
	if err != nil {
		return t, err
	}

	rows, err := database.DB.Query(
		"SELECT id, transfer_id, item_id, qty_sent, qty_received, qty_diff FROM transfer_items WHERE transfer_id = ? ORDER BY id",
		id,
	)
	if err != nil {
		return t, err
	}
	defer rows.Close()

	t.Items = []model.TransferItem{}
	for rows.Next() {
		var it model.TransferItem
		if err := rows.Scan(&it.ID, &it.TransferID, &it.ItemID, &it.QtySent, &it.QtyReceived, &it.QtyDiff); err != nil {
			return t, err
		}
		t.Items = append(t.Items, it)
	}
	return t, rows.Err()
}

func GetAllTransfers() ([]model.Transfer, error) {
	rows, err := database.DB.Query(
		"SELECT id, doc_no, DATE_FORMAT(doc_date, '%Y-%m-%d'), from_location_id, to_location_id, status FROM transfers ORDER BY id DESC LIMIT 100",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []model.Transfer{}
	for rows.Next() {
		var t model.Transfer
		if err := rows.Scan(&t.ID, &t.DocNo, &t.DocDate, &t.FromLocationID, &t.ToLocationID, &t.Status); err != nil {
			return nil, err
		}
		t.Items = []model.TransferItem{}
		list = append(list, t)
	}
	return list, rows.Err()
}
