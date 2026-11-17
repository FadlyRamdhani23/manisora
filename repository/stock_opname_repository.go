package repository

import (
	"database/sql"
	"fmt"
	"time"

	"manisora/database"
	"manisora/model"
)

func CreateStockOpname(tx *sql.Tx, o model.StockOpname) (int64, error) {
	tempNo := fmt.Sprintf("TMP-%d", time.Now().UnixNano())
	result, err := tx.Exec(
		"INSERT INTO stock_opnames (doc_no, doc_date, location_id, status) VALUES (?, ?, ?, ?)",
		tempNo, o.DocDate, o.LocationID, o.Status,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func SetStockOpnameDocNo(tx *sql.Tx, id int64, docNo string) error {
	_, err := tx.Exec("UPDATE stock_opnames SET doc_no = ? WHERE id = ?", docNo, id)
	return err
}

// Mengambil "foto" stok di lokasi, mengembalikan jumlah barang yang masuk foto
func SnapshotOpnameItems(tx *sql.Tx, opnameID, locationID int64) (int64, error) {
	result, err := tx.Exec(
		`INSERT INTO stock_opname_items (opname_id, item_id, qty_system)
		 SELECT ?, item_id, qty_on_hand FROM stock_balances WHERE location_id = ?`,
		opnameID, locationID,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func LockStockOpname(tx *sql.Tx, id int64) (status string, locationID int64, err error) {
	err = tx.QueryRow(
		"SELECT status, location_id FROM stock_opnames WHERE id = ? FOR UPDATE",
		id,
	).Scan(&status, &locationID)
	return
}

func scanOpnameItems(rows *sql.Rows) ([]model.StockOpnameItem, error) {
	defer rows.Close()

	items := []model.StockOpnameItem{}
	for rows.Next() {
		var it model.StockOpnameItem
		var physical sql.NullFloat64
		if err := rows.Scan(&it.ID, &it.OpnameID, &it.ItemID, &it.QtySystem, &physical, &it.QtyDiff); err != nil {
			return nil, err
		}
		if physical.Valid {
			v := physical.Float64
			it.QtyPhysical = &v
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

func GetStockOpnameItemsTx(tx *sql.Tx, opnameID int64) ([]model.StockOpnameItem, error) {
	rows, err := tx.Query(
		"SELECT id, opname_id, item_id, qty_system, qty_physical, qty_diff FROM stock_opname_items WHERE opname_id = ? ORDER BY id",
		opnameID,
	)
	if err != nil {
		return nil, err
	}
	return scanOpnameItems(rows)
}

func SetOpnamePhysical(tx *sql.Tx, opnameID, itemID int64, qty float64) error {
	_, err := tx.Exec(
		"UPDATE stock_opname_items SET qty_physical = ? WHERE opname_id = ? AND item_id = ?",
		qty, opnameID, itemID,
	)
	return err
}

func SetOpnameItemDiff(tx *sql.Tx, rowID int64, diff float64) error {
	_, err := tx.Exec("UPDATE stock_opname_items SET qty_diff = ? WHERE id = ?", diff, rowID)
	return err
}

func SetStockOpnameStatus(tx *sql.Tx, id int64, status string) error {
	_, err := tx.Exec("UPDATE stock_opnames SET status = ? WHERE id = ?", status, id)
	return err
}

func GetStockOpnameByID(id int64) (model.StockOpname, error) {
	var o model.StockOpname
	err := database.DB.QueryRow(
		"SELECT id, doc_no, DATE_FORMAT(doc_date, '%Y-%m-%d'), location_id, status FROM stock_opnames WHERE id = ?",
		id,
	).Scan(&o.ID, &o.DocNo, &o.DocDate, &o.LocationID, &o.Status)
	if err != nil {
		return o, err
	}

	rows, err := database.DB.Query(
		"SELECT id, opname_id, item_id, qty_system, qty_physical, qty_diff FROM stock_opname_items WHERE opname_id = ? ORDER BY id",
		id,
	)
	if err != nil {
		return o, err
	}
	o.Items, err = scanOpnameItems(rows)
	return o, err
}

func GetAllStockOpnames() ([]model.StockOpname, error) {
	rows, err := database.DB.Query(
		"SELECT id, doc_no, DATE_FORMAT(doc_date, '%Y-%m-%d'), location_id, status FROM stock_opnames ORDER BY id DESC LIMIT 100",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []model.StockOpname{}
	for rows.Next() {
		var o model.StockOpname
		if err := rows.Scan(&o.ID, &o.DocNo, &o.DocDate, &o.LocationID, &o.Status); err != nil {
			return nil, err
		}
		o.Items = []model.StockOpnameItem{}
		list = append(list, o)
	}
	return list, rows.Err()
}
