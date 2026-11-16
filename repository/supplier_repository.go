package repository

import (
	"manisora/database"
	"manisora/model"
)

func GetAllSuppliers() ([]model.Supplier, error) {
	rows, err := database.DB.Query("SELECT id, code, name FROM suppliers")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	suppliers := []model.Supplier{}
	for rows.Next() {
		var s model.Supplier
		if err := rows.Scan(&s.ID, &s.Code, &s.Name); err != nil {
			return nil, err
		}
		suppliers = append(suppliers, s)
	}
	return suppliers, rows.Err()
}

func CreateSupplier(s model.Supplier) (int64, error) {
	result, err := database.DB.Exec(
		"INSERT INTO suppliers (code, name) VALUES (?, ?)",
		s.Code, s.Name,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}
