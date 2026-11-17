package repository

import (
	"database/sql"
	"fmt"
	"time"

	"manisora/database"
	"manisora/model"
)

func CreateSale(tx *sql.Tx, s model.Sale) (int64, error) {
	tempNo := fmt.Sprintf("TMP-%d", time.Now().UnixNano())
	result, err := tx.Exec(
		"INSERT INTO sales (doc_no, doc_date, location_id, status) VALUES (?, ?, ?, ?)",
		tempNo, s.DocDate, s.LocationID, s.Status,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func SetSaleDocNo(tx *sql.Tx, id int64, docNo string) error {
	_, err := tx.Exec("UPDATE sales SET doc_no = ? WHERE id = ?", docNo, id)
	return err
}

func CreateSaleItem(tx *sql.Tx, saleID int64, it model.SaleItem) (int64, error) {
	result, err := tx.Exec(
		"INSERT INTO sale_items (sale_id, item_id, qty) VALUES (?, ?, ?)",
		saleID, it.ItemID, it.Qty,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func LockSale(tx *sql.Tx, id int64) (status string, locationID int64, err error) {
	err = tx.QueryRow(
		"SELECT status, location_id FROM sales WHERE id = ? FOR UPDATE",
		id,
	).Scan(&status, &locationID)
	return
}

func GetSaleItemsTx(tx *sql.Tx, saleID int64) ([]model.SaleItem, error) {
	rows, err := tx.Query(
		"SELECT id, sale_id, item_id, qty FROM sale_items WHERE sale_id = ? ORDER BY id",
		saleID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []model.SaleItem{}
	for rows.Next() {
		var it model.SaleItem
		if err := rows.Scan(&it.ID, &it.SaleID, &it.ItemID, &it.Qty); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

func SetSaleStatus(tx *sql.Tx, id int64, status string) error {
	_, err := tx.Exec("UPDATE sales SET status = ? WHERE id = ?", status, id)
	return err
}

func GetSaleByID(id int64) (model.Sale, error) {
	var s model.Sale
	err := database.DB.QueryRow(
		"SELECT id, doc_no, DATE_FORMAT(doc_date, '%Y-%m-%d'), location_id, status FROM sales WHERE id = ?",
		id,
	).Scan(&s.ID, &s.DocNo, &s.DocDate, &s.LocationID, &s.Status)
	if err != nil {
		return s, err
	}

	rows, err := database.DB.Query(
		"SELECT id, sale_id, item_id, qty FROM sale_items WHERE sale_id = ? ORDER BY id",
		id,
	)
	if err != nil {
		return s, err
	}
	defer rows.Close()

	s.Items = []model.SaleItem{}
	for rows.Next() {
		var it model.SaleItem
		if err := rows.Scan(&it.ID, &it.SaleID, &it.ItemID, &it.Qty); err != nil {
			return s, err
		}
		s.Items = append(s.Items, it)
	}
	return s, rows.Err()
}

func GetAllSales() ([]model.Sale, error) {
	rows, err := database.DB.Query(
		"SELECT id, doc_no, DATE_FORMAT(doc_date, '%Y-%m-%d'), location_id, status FROM sales ORDER BY id DESC LIMIT 100",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []model.Sale{}
	for rows.Next() {
		var s model.Sale
		if err := rows.Scan(&s.ID, &s.DocNo, &s.DocDate, &s.LocationID, &s.Status); err != nil {
			return nil, err
		}
		s.Items = []model.SaleItem{}
		list = append(list, s)
	}
	return list, rows.Err()
}
