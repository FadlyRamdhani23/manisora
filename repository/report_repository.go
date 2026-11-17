package repository

import (
	"manisora/database"
	"manisora/model"
)

func GetStockMutation(locationID int64, from, to string, itemID int64) ([]model.StockMutation, error) {
	query := `
	SELECT i.id, i.code, i.name, i.base_unit,
	  COALESCE(SUM(CASE WHEN l.trx_date < p.d_from THEN l.qty_in - l.qty_out END), 0),
	  COALESCE(SUM(CASE WHEN l.trx_date >= p.d_from THEN l.qty_in END), 0),
	  COALESCE(SUM(CASE WHEN l.trx_date >= p.d_from THEN l.qty_out END), 0),
	  COALESCE(SUM(CASE WHEN l.trx_date >= p.d_from AND l.movement_type = 'PURCHASE_RECEIPT' THEN l.qty_in END), 0),
	  COALESCE(SUM(CASE WHEN l.trx_date >= p.d_from AND l.movement_type = 'TRANSFER_IN' THEN l.qty_in END), 0),
	  COALESCE(SUM(CASE WHEN l.trx_date >= p.d_from AND l.movement_type = 'SALE' THEN l.qty_out END), 0),
	  COALESCE(SUM(CASE WHEN l.trx_date >= p.d_from AND l.movement_type = 'TRANSFER_OUT' THEN l.qty_out END), 0),
	  COALESCE(SUM(CASE WHEN l.trx_date >= p.d_from AND l.movement_type = 'SUPPLIER_RETURN' THEN l.qty_out END), 0),
	  COALESCE(SUM(CASE WHEN l.trx_date >= p.d_from AND l.movement_type = 'OPNAME_ADJUSTMENT' THEN l.qty_in - l.qty_out END), 0)
	FROM (SELECT CAST(? AS DATE) AS d_from, DATE_ADD(CAST(? AS DATE), INTERVAL 1 DAY) AS d_to) p
	JOIN stock_ledger l ON l.trx_date < p.d_to
	JOIN items i ON i.id = l.item_id
	WHERE l.location_id = ?`

	args := []interface{}{from, to, locationID}
	if itemID > 0 {
		query += " AND l.item_id = ?"
		args = append(args, itemID)
	}
	query += " GROUP BY i.id, i.code, i.name, i.base_unit ORDER BY i.code"

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []model.StockMutation{}
	for rows.Next() {
		var m model.StockMutation
		if err := rows.Scan(&m.ItemID, &m.Code, &m.Name, &m.BaseUnit,
			&m.Opening, &m.QtyIn, &m.QtyOut,
			&m.Breakdown.PurchaseReceipt, &m.Breakdown.TransferIn,
			&m.Breakdown.Sale, &m.Breakdown.TransferOut,
			&m.Breakdown.SupplierReturn, &m.Breakdown.Opname); err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	return list, rows.Err()
}
